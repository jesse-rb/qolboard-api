// Bad code for a bad idea
package relations_service

import (
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
	SingleRelationLoaderFunc func(tx *sqlx.Tx, model *IHasRelations) (map[any]any, []IHasRelations, error)
	BatchRelationLoaderFunc  func(tx *sqlx.Tx, models []IHasRelations) (map[any]any, []IHasRelations, error)
)

type Relation struct {
	single SingleRelationLoaderFunc
	batch  BatchRelationLoaderFunc
	assign func(model any, related any)
}

func HasOne[TModel IHasRelations, TRelated IHasRelations](
	name string,
	r RelationRegistry,
	single string,
	batch string,
	assign func(*TModel, *TRelated),
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) {
	r.Relations[name] = Relation{
		single: makeSingleLoader(single, assign),
		batch:  makeBatchLoader(batch, assign, getModelForeignKey, getRelatedPrimaryKey),
		assign: makeAssignFunc(assign),
	}
}

func BelongsTo[TModel IHasRelations, TRelated IHasRelations](
	name string,
	r RelationRegistry,
	single string,
	batch string,
	assign func(*TModel, *TRelated),
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) {
	r.Relations[name] = Relation{
		single: makeSingleLoader(single, assign),
		batch:  makeBatchLoader(batch, assign, getModelForeignKey, getRelatedPrimaryKey),
		assign: makeAssignFunc(assign),
	}
}

func HasMany[TModel IHasRelations, TRelated IHasRelations](
	name string,
	r RelationRegistry,
	single string,
	batch string,
	assign func(*TModel, []TRelated),
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) {
	r.Relations[name] = Relation{
		single: makeHasManySingleLoader(single, assign),
		batch:  makeHasManyBatchLoader(batch, assign, getModelForeignKey, getRelatedPrimaryKey),
		assign: makeAssignManyFunc(assign),
	}
}

func makeAssignFunc[TModel IHasRelations, TRelated IHasRelations](
	assign func(*TModel, *TRelated),
) func(any, any) {
	return func(model any, related any) {
		m, ok1 := model.(*TModel)
		r, ok2 := related.(*TRelated)

		if ok1 && ok2 {
			assign(m, r)
		}
	}
}

func makeAssignManyFunc[TModel IHasRelations, TRelated IHasRelations](
	assign func(*TModel, []TRelated),
) func(any, any) {
	return func(model any, related any) {
		m, ok1 := model.(*TModel)
		r, ok2 := related.([]TRelated)

		if ok1 && ok2 {
			assign(m, r)
		}
	}
}

func makeSingleLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(*TModel, *TRelated),
) SingleRelationLoaderFunc {
	return func(tx *sqlx.Tx, model *IHasRelations) (map[any]any, []IHasRelations, error) {
		related := new(TRelated)
		err := tx.Get(related, query, (*model).GetPrimaryKey())
		if err != nil {
			return nil, nil, err
		}
		if tModel, ok := (*model).(TModel); ok {
			assign(&tModel, related)
			var tmp IHasRelations = tModel
			model = &tmp
		}
		toReturn := make(map[any]any)
		key := (*model).GetPrimaryKey()
		toReturn[key] = *related
		return toReturn, []IHasRelations{*related}, nil
	}
}

func makeHasManySingleLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(*TModel, []TRelated),
) SingleRelationLoaderFunc {
	return func(tx *sqlx.Tx, model *IHasRelations) (map[any]any, []IHasRelations, error) {
		related := make([]TRelated, 0)
		err := tx.Select(&related, query, (*model).GetPrimaryKey())
		if err != nil {
			return nil, nil, err
		}
		if tModel, ok := (*model).(TModel); ok {
			assign(&tModel, related)
			*model = tModel
		}

		toReturnMap := make(map[any]any, len(related))
		toReturnMap[(*model).GetPrimaryKey()] = related
		toReturnSlice := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturnSlice[i] = r
		}
		return toReturnMap, toReturnSlice, nil
	}
}

func makeBatchLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(*TModel, *TRelated),
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) BatchRelationLoaderFunc {
	return func(tx *sqlx.Tx, models []IHasRelations) (map[any]any, []IHasRelations, error) {
		keys := make([]any, 0, len(models))
		for i := range models {
			keys = append(keys, getModelForeignKey(models[i].(TModel)))
		}

		queryStr, args, err := sqlx.In(query, keys)
		if err != nil {
			return nil, nil, err
		}
		queryStr = tx.Rebind(queryStr)

		related := make([]TRelated, 0, len(models))
		err = tx.Select(&related, queryStr, args...)
		if err != nil {
			return nil, nil, err
		}

		relMap := make(map[any]*TRelated)
		for i := range related {
			r := &related[i]
			relMap[getRelatedPrimaryKey(*r)] = r
		}

		for i := range models {
			if tModel, ok := models[i].(TModel); ok {
				if found, ok := relMap[getModelForeignKey(tModel)]; ok {
					assign(&tModel, found)
					models[i] = tModel
				}
			}
		}

		toReturnMap := make(map[any]any, len(related))
		for key, r := range relMap {
			toReturnMap[key] = r
		}
		toReturnSlice := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturnSlice[i] = r
		}

		return toReturnMap, toReturnSlice, nil
	}
}

func makeHasManyBatchLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(*TModel, []TRelated),
	getModelForeignKey func(TModel) any,
	getRelatedPrimaryKey func(TRelated) any,
) BatchRelationLoaderFunc {
	return func(tx *sqlx.Tx, models []IHasRelations) (map[any]any, []IHasRelations, error) {
		keys := make([]any, 0, len(models))
		for i := range models {
			keys = append(keys, models[i].GetPrimaryKey())
		}

		queryStr, args, err := sqlx.In(query, keys)
		if err != nil {
			return nil, nil, err
		}
		queryStr = tx.Rebind(queryStr)

		related := make([]TRelated, len(models))
		err = tx.Select(&related, queryStr, args...)
		if err != nil {
			return nil, nil, err
		}

		relMap := make(map[any][]TRelated)
		for _, r := range related {
			k := getRelatedPrimaryKey(r)
			relMap[k] = append(relMap[k], r)
		}

		for i := range models {
			if tModel, ok := models[i].(TModel); ok {
				if found, ok := relMap[getModelForeignKey(tModel)]; ok {
					assign(&tModel, found)
					models[i] = tModel
				}
			}
		}

		toReturnMap := make(map[any]any, len(related))
		for key, r := range relMap {
			toReturnMap[key] = r
		}
		toReturnSlice := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturnSlice[i] = r
		}

		return toReturnMap, toReturnSlice, nil
	}
}

func (r RelationRegistry) LoadRelations(tx *sqlx.Tx, model *IHasRelations, with map[string]any) error {
	for name := range with {
		if relation, ok := r.Relations[name]; ok {
			loaderFn := relation.single
			relatedMap, relatedSlice, err := loaderFn(tx, model)
			if err != nil {
				return err
			}

			if len(relatedSlice) > 0 {
				rRelated := relatedSlice[0].GetRelations()
				if nestedWithMap, ok := with[name].(map[string]any); ok {
					rRelated.LoadBatchRelations(tx, relatedSlice, nestedWithMap)
				}

				key := (*model).GetPrimaryKey()
				if _, ok := relatedMap[key]; ok {
					assignFn := relation.assign
					assignFn(model, relatedMap[key])
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
			relatedMap, relatedSlice, err := loaderFn(tx, models)
			if err != nil {
				return err
			}

			if len(relatedSlice) > 0 {
				rRelated := relatedSlice[0].GetRelations()
				if nestedWithMap, ok := with[name].(map[string]any); ok {
					rRelated.LoadBatchRelations(tx, relatedSlice, nestedWithMap)
				}

				for i := range models {
					key := (models[i]).GetPrimaryKey()
					if _, ok := relatedMap[key]; ok {
						assignFn := relation.assign
						assignFn(models[i], relatedMap[key])
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
