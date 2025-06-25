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
	single SingleRelationLoaderFunc
	batch  BatchRelationLoaderFunc
}

type Single[TRelated IHasRelations] struct {
	query  string
	assign func(*IHasRelations, TRelated)
}

type Batch[TRelated IHasRelations] struct {
	query         string
	assign        func(*IHasRelations, TRelated)
	getForeignKey func(TRelated) any
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
	}
}

func makeSingleLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(*TModel, *TRelated),
) SingleRelationLoaderFunc {
	return func(tx *sqlx.Tx, model *IHasRelations) ([]IHasRelations, error) {
		related := new(TRelated)
		err := tx.Get(related, query, (*model).GetPrimaryKey())
		if err != nil {
			return nil, err
		}
		if tModel, ok := (*model).(TModel); ok {
			assign(&tModel, related)
			var tmp IHasRelations = tModel
			model = &tmp
		}
		return []IHasRelations{*related}, nil
	}
}

func makeHasManySingleLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(*TModel, []TRelated),
) SingleRelationLoaderFunc {
	return func(tx *sqlx.Tx, model *IHasRelations) ([]IHasRelations, error) {
		related := make([]TRelated, 0)
		err := tx.Select(&related, query, (*model).GetPrimaryKey())
		if err != nil {
			return nil, err
		}
		if tModel, ok := (*model).(TModel); ok {
			assign(&tModel, related)
			*model = tModel
		}

		toReturn := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturn[i] = r
		}
		return toReturn, nil
	}
}

func makeBatchLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(*TModel, *TRelated),
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
					assign(&tModel, found)
					models[i] = tModel
				}
			}
		}

		toReturn := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturn[i] = r
		}

		return toReturn, nil
	}
}

func makeHasManyBatchLoader[TModel IHasRelations, TRelated IHasRelations](
	query string,
	assign func(*TModel, []TRelated),
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
					assign(&tModel, found)
					models[i] = tModel
				}
			}
		}

		toReturn := make([]IHasRelations, len(related))
		for i, r := range related {
			toReturn[i] = r
		}

		return toReturn, nil
	}
}

func (r RelationRegistry) LoadRelations(tx *sqlx.Tx, model *IHasRelations, with map[string]any) error {
	for name := range with {
		if relation, ok := r.Relations[name]; ok {
			loaderFn := relation.single
			related, err := loaderFn(tx, model)
			if err != nil {
				return err
			}

			if len(related) > 0 {
				rRelated := related[0].GetRelations()
				if nestedWithMap, ok := with[name].(map[string]any); ok {
					rRelated.LoadBatchRelations(tx, related, nestedWithMap)
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
			related, err := loaderFn(tx, models)
			if err != nil {
				return err
			}

			if len(related) > 0 {
				rRelated := related[0].GetRelations()
				if nestedWithMap, ok := with[name].(map[string]any); ok {
					rRelated.LoadBatchRelations(tx, related, nestedWithMap)
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
	logging.LogDebug("iHassRelations", "iHasRelations", iHasRelations)

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

	// tModels := make([]TModel, len(iHasRelations))
	for i, rel := range iHasRelations {
		if tModel, ok := rel.(TModel); ok {
			models[i] = tModel
		}
	}
	// models = tModels

	return err
}
