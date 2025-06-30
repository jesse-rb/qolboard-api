// Bad code for a bad idea
package relations_service

import (
	"qolboard-api/services/logging"

	"github.com/jesse-rb/imissphp-go"
	"github.com/jmoiron/sqlx"
)

type RelationRegistry struct {
	Relations map[string]Relation
}

func NewRelationRegistry() RelationRegistry {
	return RelationRegistry{
		Relations: make(map[string]Relation),
	}
}

type IHasRelations interface {
	GetRelations() RelationRegistry
	GetPrimaryKey() any
}

type (
	SingleRelationLoaderFunc func(tx *sqlx.Tx, model *IHasRelations) ([]IHasRelations, error)
	BatchRelationLoaderFunc  func(tx *sqlx.Tx, models []IHasRelations) ([]IHasRelations, error)
)

type Relation struct {
	single       SingleRelationLoaderFunc
	batch        BatchRelationLoaderFunc
	assign       func(model IHasRelations, related any) IHasRelations
	getModelFk   func(model IHasRelations) any
	getRelatedPk func(related IHasRelations) any
	kind         string
}

func HasOne[TModel IHasRelations, TRelated IHasRelations](
	name string,
	r RelationRegistry,
	single string,
	batch string,
	assign func(TModel, TRelated) TModel,
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) {
	r.Relations[name] = Relation{
		single:       makeSingleLoader(single, assign),
		batch:        makeBatchLoader(batch, assign, getModelForeignKey, getRelatedPrimaryKey),
		assign:       makeAssignFunc(assign),
		getModelFk:   makeGetModelPkFunc(getModelForeignKey),
		getRelatedPk: makeGetRelatedPkFunc(getRelatedPrimaryKey),
		kind:         "has_one",
	}
}

func BelongsTo[TModel IHasRelations, TRelated IHasRelations](
	name string,
	r RelationRegistry,
	single string,
	batch string,
	assign func(TModel, TRelated) TModel,
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) {
	r.Relations[name] = Relation{
		single:       makeSingleLoader(single, assign),
		batch:        makeBatchLoader(batch, assign, getModelForeignKey, getRelatedPrimaryKey),
		assign:       makeAssignFunc(assign),
		getModelFk:   makeGetModelPkFunc(getModelForeignKey),
		getRelatedPk: makeGetRelatedPkFunc(getRelatedPrimaryKey),
		kind:         "belongs_to",
	}
}

func HasMany[TModel IHasRelations, TRelated IHasRelations](
	name string,
	r RelationRegistry,
	single string,
	batch string,
	assign func(TModel, []TRelated) TModel,
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) {
	r.Relations[name] = Relation{
		single:       makeHasManySingleLoader(single, assign),
		batch:        makeHasManyBatchLoader(batch, assign, getModelForeignKey, getRelatedPrimaryKey),
		assign:       makeAssignManyFunc(assign),
		getModelFk:   makeGetModelPkFunc(getModelForeignKey),
		getRelatedPk: makeGetRelatedPkFunc(getRelatedPrimaryKey),
		kind:         "has_many",
	}
}

func makeAssignFunc[TModel IHasRelations, TRelated IHasRelations](
	assign func(TModel, TRelated) TModel,
) func(IHasRelations, any) IHasRelations {
	return func(model IHasRelations, related any) IHasRelations {
		m, ok1 := model.(TModel)
		r, ok2 := related.(TRelated)
		if ok1 && ok2 {
			model = assign(m, r)
		}

		return model
	}
}

func makeAssignManyFunc[TModel IHasRelations, TRelated IHasRelations](
	assign func(TModel, []TRelated) TModel,
) func(IHasRelations, any) IHasRelations {
	return func(model IHasRelations, related any) IHasRelations {
		m, ok1 := model.(TModel)
		r, ok2 := related.([]IHasRelations)

		logging.LogDebug("makeAssignManyFunc", "ok1, ok2", map[string]any{
			"ok1": ok1,
			"ok2": ok2,
		})

		if ok1 && ok2 {
			_r := make([]TRelated, 0)
			for i := range r {
				if tRelated, ok := r[i].(TRelated); ok {
					_r = append(_r, tRelated)
				}
			}
			model = assign(m, _r)
		}

		return model
	}
}

func makeGetModelPkFunc[TModel IHasRelations](
	getModelPk func(TModel) any,
) func(IHasRelations) any {
	return func(model IHasRelations) any {
		m, ok := model.(TModel)

		if ok {
			return getModelPk(m)
		} else {
			return nil
		}
	}
}

func makeGetRelatedPkFunc[TRelated IHasRelations](
	getRelatedPk func(TRelated) any,
) func(IHasRelations) any {
	return func(related IHasRelations) any {
		r, ok := related.(TRelated)

		if ok {
			return getRelatedPk(r)
		} else {
			return nil
		}
	}
}

func makeSingleLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(TModel, TRelated) TModel,
) SingleRelationLoaderFunc {
	return func(tx *sqlx.Tx, model *IHasRelations) ([]IHasRelations, error) {
		related := new(TRelated)
		err := tx.Get(related, query, (*model).GetPrimaryKey())
		if err != nil {
			return nil, err
		}
		if tModel, ok := (*model).(TModel); ok {
			tModel = assign(tModel, *related)
			var tmp IHasRelations = tModel
			*model = tmp
		}
		return []IHasRelations{*related}, nil
	}
}

func makeHasManySingleLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(TModel, []TRelated) TModel,
) SingleRelationLoaderFunc {
	return func(tx *sqlx.Tx, model *IHasRelations) ([]IHasRelations, error) {
		related := make([]TRelated, 0)
		err := tx.Select(&related, query, (*model).GetPrimaryKey())
		if err != nil {
			return nil, err
		}
		if tModel, ok := (*model).(TModel); ok {
			tModel = assign(tModel, related)
			*model = tModel
		}

		toReturnSlice := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturnSlice[i] = r
		}
		return toReturnSlice, nil
	}
}

func makeBatchLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(TModel, TRelated) TModel,
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) BatchRelationLoaderFunc {
	return func(tx *sqlx.Tx, models []IHasRelations) ([]IHasRelations, error) {
		keys := make([]any, 0, len(models))
		for i := range models {
			keys = append(keys, getModelForeignKey(models[i].(TModel)))
		}

		queryStr, args, err := sqlx.In(query, keys)
		if err != nil {
			return nil, err
		}
		queryStr = tx.Rebind(queryStr)

		related := make([]TRelated, 0, len(models))
		err = tx.Select(&related, queryStr, args...)
		if err != nil {
			return nil, err
		}

		relMap := make(map[any]*TRelated)
		for i := range related {
			r := &related[i]
			relMap[getRelatedPrimaryKey(*r)] = r
		}

		for i := range models {
			if tModel, ok := models[i].(TModel); ok {
				if found, ok := relMap[getModelForeignKey(tModel)]; ok {
					tModel = assign(tModel, *found)
					models[i] = tModel
				}
			}
		}

		toReturnSlice := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturnSlice[i] = r
		}

		return toReturnSlice, nil
	}
}

func makeHasManyBatchLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(TModel, []TRelated) TModel,
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) BatchRelationLoaderFunc {
	return func(tx *sqlx.Tx, models []IHasRelations) ([]IHasRelations, error) {
		keys := make([]any, 0, len(models))
		for i := range models {
			keys = append(keys, models[i].GetPrimaryKey())
		}

		queryStr, args, err := sqlx.In(query, keys)
		if err != nil {
			return nil, err
		}
		queryStr = tx.Rebind(queryStr)

		related := make([]TRelated, len(models))
		err = tx.Select(&related, queryStr, args...)
		if err != nil {
			return nil, err
		}

		relMap := make(map[any][]TRelated)
		for _, r := range related {
			k := getRelatedPrimaryKey(r)
			relMap[k] = append(relMap[k], r)
		}

		for i := range models {
			if tModel, ok := models[i].(TModel); ok {
				if found, ok := relMap[getModelForeignKey(tModel)]; ok {
					models[i] = assign(tModel, found)
				}
			}
		}

		toReturnSlice := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturnSlice[i] = r
		}

		return toReturnSlice, nil
	}
}

func (r RelationRegistry) LoadRelations(tx *sqlx.Tx, model *IHasRelations, with map[string]any) error {
	for name := range with {
		if relation, ok := r.Relations[name]; ok {
			loaderFn := relation.single
			relatedSlice, err := loaderFn(tx, model)
			if err != nil {
				return err
			}

			if len(relatedSlice) > 0 {
				rRelated := relatedSlice[0].GetRelations()
				if nestedWithMap, ok := with[name].(map[string]any); ok {
					rRelated.LoadBatchRelations(tx, relatedSlice, nestedWithMap)
				}

				// IF kind is has_one or belongs_to, we simply assign related to the model (there is only one)
				// ELSE IF kind is has_many, we simply assign the entire related collection to the model
				if relation.kind == "has_one" || relation.kind == "belongs_to" {
					m := relation.assign(*model, relatedSlice[0])
					*model = m
				} else if relation.kind == "has_many" {
					m := relation.assign(*model, relatedSlice)
					*model = m
				}
			}
		}
	}
	return nil
}

func (r RelationRegistry) LoadBatchRelations(tx *sqlx.Tx, models []IHasRelations, with map[string]any) error {
	for name := range with {
		if relation, ok := r.Relations[name]; ok {
			loaderFn := relation.batch
			relatedSlice, err := loaderFn(tx, models)
			if err != nil {
				return err
			}

			if len(relatedSlice) > 0 {
				rRelated := relatedSlice[0].GetRelations()
				if nestedWithMap, ok := with[name].(map[string]any); ok {
					rRelated.LoadBatchRelations(tx, relatedSlice, nestedWithMap)
				}
				// logging.LogDebug("loadBatchRelations -- after loading related", name, relation.kind)

				// IF kind is has_one or belongs_to, we assign each related to the corresponding model (key by model pk)
				// ELSE IF kind is has_many, we assign the entire related collection to the corresponding model model (collect/group by model pk)
				if kind := relation.kind; kind == "has_one" || kind == "belongs_to" {
					// logging.LogDebug("loadBatchRelations:has_one,belongs_to -- relatedSlice", name, relatedSlice)
					m := make(map[any]IHasRelations, len(relatedSlice))
					for _, related := range relatedSlice {
						key := relation.getRelatedPk(related)
						m[key] = related
					}
					logging.LogDebug("loadBatchRelations:has_one,belongs_to -- m", name, m)

					for i := range models {
						key := relation.getModelFk(models[i])
						if _, ok := m[key]; ok {
							// logging.LogDebug("loadBatchRelations:has_one,belongs_to -- found related, m[key]", name, m[key])
							models[i] = relation.assign(models[i], m[key])
						}
					}
				} else if kind == "has_many" {
					m := make(map[any][]IHasRelations)
					for _, related := range relatedSlice {
						key := relation.getRelatedPk(related)
						m[key] = append(m[key], related)
					}

					for i := range models {
						key := relation.getModelFk(models[i])
						if _, ok := m[key]; ok {
							models[i] = relation.assign(models[i], m[key])
						}
					}
				}
			}
		}
	}
	return nil
}

func HandleWithParams(with []string) map[string]any {
	withMap := make(map[string]any)
	for _, w := range with {
		withMap[w] = nil // or some default value if needed
	}
	withMap = imissphp.UnFlattenMap(withMap)
	return withMap
}

func Load[TModel IHasRelations](tx *sqlx.Tx, r RelationRegistry, model *TModel, with []string) error {
	var iHasRelations IHasRelations = *model
	err := r.LoadRelations(tx, &iHasRelations, HandleWithParams(with))

	if tModel, ok := iHasRelations.(TModel); ok {
		*model = tModel
	}
	return err
}

func LoadBatch[TModel IHasRelations](tx *sqlx.Tx, r RelationRegistry, models []TModel, with []string) error {
	iHasRelations := make([]IHasRelations, len(models))
	for i, t := range models {
		iHasRelations[i] = t
	}
	err := r.LoadBatchRelations(tx, iHasRelations, HandleWithParams(with))

	for i, rel := range iHasRelations {
		if tModel, ok := rel.(TModel); ok {
			models[i] = tModel
		}
	}

	return err
}
