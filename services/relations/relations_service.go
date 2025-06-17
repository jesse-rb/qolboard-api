package relations_service

import (
	"qolboard-api/services/logging"

	"github.com/jmoiron/sqlx"
)

type (
	SingleRelationLoaderFunc[TModel any] func(tx *sqlx.Tx, model *TModel) error
	BatchRelationLoaderFunc[TModel any]  func(tx *sqlx.Tx, models []TModel) error
)

type RelationRegistry[TModel any] struct {
	singleLoaders map[string]SingleRelationLoaderFunc[TModel]
	batchLoaders  map[string]BatchRelationLoaderFunc[TModel]
}

func NewRelationRegistry[TModel any]() *RelationRegistry[TModel] {
	return &RelationRegistry[TModel]{
		singleLoaders: make(map[string]SingleRelationLoaderFunc[TModel]),
		batchLoaders:  make(map[string]BatchRelationLoaderFunc[TModel]),
	}
}

func (r *RelationRegistry[TModel]) RegisterSingle(name string, fn SingleRelationLoaderFunc[TModel]) {
	r.singleLoaders[name] = fn
}

func (r *RelationRegistry[TModel]) RegisterBatch(name string, fn BatchRelationLoaderFunc[TModel]) {
	r.batchLoaders[name] = fn
}

func MakeSingleLoader[TModel any, TRelated any](
	query string,
	getKey func(*TModel) any,
	assign func(*TModel, *TRelated),
) SingleRelationLoaderFunc[TModel] {
	return func(tx *sqlx.Tx, m *TModel) error {
		obj := new(TRelated)
		err := tx.Get(obj, query, getKey(m))
		if err != nil {
			return err
		}
		assign(m, obj)
		return nil
	}
}

func MakeHasManySingleLoader[TModel any, TRelated any](
	query string,
	getModelKey func(*TModel) any,
	assign func(*TModel, []TRelated),
) SingleRelationLoaderFunc[TModel] {
	return func(tx *sqlx.Tx, model *TModel) error {
		var related []TRelated
		err := tx.Select(&related, query, getModelKey(model))
		if err != nil {
			return err
		}
		assign(model, related)
		return nil
	}
}

func MakeBatchLoader[TModel any, TRelated any, TKey comparable](
	query string,
	getModelKey func(*TModel) TKey,
	assign func(*TModel, *TRelated),
	getRelatedKey func(*TRelated) TKey,
) BatchRelationLoaderFunc[TModel] {
	return func(tx *sqlx.Tx, models []TModel) error {
		keys := make([]TKey, 0, len(models))
		for i := range models {
			keys = append(keys, getModelKey(&models[i]))
		}

		queryStr, args, err := sqlx.In(query, keys)
		if err != nil {
			return err
		}
		queryStr = tx.Rebind(queryStr)

		var related []TRelated
		err = tx.Select(&related, queryStr, args...)
		if err != nil {
			return err
		}
		logging.LogDebug("relations_service::MakeBatchLoader", "related", related)

		relMap := make(map[TKey]*TRelated)
		for i := range related {
			r := related[i]
			relMap[getRelatedKey(&r)] = &r
		}

		for i := range models {
			if r, ok := relMap[getModelKey(&models[i])]; ok {
				assign(&models[i], r)
			}
		}

		return nil
	}
}

func MakeHasManyBatchLoader[TModel any, TRelated any, TKey comparable](
	query string,
	getModelKey func(*TModel) TKey,
	assign func(*TModel, []TRelated),
	getRelatedKey func(*TRelated) TKey,
) BatchRelationLoaderFunc[TModel] {
	return func(tx *sqlx.Tx, models []TModel) error {
		keys := make([]TKey, 0, len(models))
		for i := range models {
			keys = append(keys, getModelKey(&models[i]))
		}

		queryStr, args, err := sqlx.In(query, keys)
		if err != nil {
			return err
		}
		queryStr = tx.Rebind(queryStr)

		var related []TRelated
		err = tx.Select(&related, queryStr, args...)
		if err != nil {
			return err
		}

		relMap := make(map[TKey][]TRelated)
		for _, r := range related {
			k := getRelatedKey(&r)
			relMap[k] = append(relMap[k], r)
		}

		for i := range models {
			assign(&models[i], relMap[getModelKey(&models[i])])
		}

		return nil
	}
}

func LoadRelations[T any](r *RelationRegistry[T], m *T, tx *sqlx.Tx, with []string) error {
	for _, name := range with {
		if fn, ok := r.singleLoaders[name]; ok {
			if err := fn(tx, m); err != nil {
				return err
			}
		}
	}
	return nil
}

func LoadBatchRelations[T any](r *RelationRegistry[T], models []T, tx *sqlx.Tx, with []string) error {
	for _, name := range with {
		if fn, ok := r.batchLoaders[name]; ok {
			if err := fn(tx, models); err != nil {
				return err
			}
		}
	}
	return nil
}
