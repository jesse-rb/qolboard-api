package model

import (
	"time"

	"github.com/jesse-rb/imissphp-go"
)

// type (
// 	FuncLoadBelongsTo[T any] func(tx *sqlx.Tx, m *T) error
// 	FuncLoadHasOne[T any]    func(tx *sqlx.Tx, m *T) error
// 	FuncLoadHasMany[T any]   func(tx *sqlx.Tx, m *T) error
//
// 	FuncBatchLoadBelongsTo[T any, U any] func(tx *sqlx.Tx, m []T, with map[string]any) ([]U, error)
// 	FuncBatchLoadHasOne[T any]           func(tx *sqlx.Tx, m []T) error
// 	FuncBatchLoadHasMany[T any]          func(tx *sqlx.Tx, m []T) error
// )
//
// type BelongsToLoader[T any, U any] struct {
// 	Loader      FuncLoadBelongsTo[T]
// 	BatchLoader FuncBatchLoadBelongsTo[T, U]
// }
// type HasOneLoader[T any] struct {
// 	Loader      FuncLoadHasOne[T]
// 	BatchLoader FuncBatchLoadHasOne[T]
// }
// type HasManyLoader[T any] struct {
// 	Loader      FuncLoadHasMany[T]
// 	BatchLoader FuncBatchLoadHasMany[T]
// }
//
// type RelationLoaders[T any, U any] struct {
// 	BelongsTo map[string]BelongsToLoader[T, U]
// 	HasMany   map[string]HasManyLoader[T]
// 	HasOne    map[string]HasOneLoader[T]
// }

type Modelable interface {
	Response() map[string]any
}

type Respondable interface {
	Response() map[string]any
}

type Model struct {
	ID        uint64     `json:"id" db:"id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
}

func (m Model) Response() map[string]any {
	r := imissphp.ToMap(m)
	return r
}

func ParseWithParam(withSlice []string) map[string]any {
	withMap := make(map[string]any, 0)
	for _, with := range withSlice {
		withMap[with] = "test"
	}

	withMap = imissphp.UnFlattenMap(withMap)
	return withMap
}

// func GenericRelationsLoader[T any, U any](relationLoaders RelationLoaders[T, U], model *T, tx *sqlx.Tx, with []string) error {
// 	var err error
//
// 	for _, w := range with {
// 		// BelongsTo
// 		if relationLoader, ok := relationLoaders.BelongsTo[w]; ok {
// 			err = relationLoader.Loader(tx, model)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		// HasOne
// 		if relationLoader, ok := relationLoaders.HasMany[w]; ok {
// 			err = relationLoader.Loader(tx, model)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 		// HasMany
// 		if relationLoader, ok := relationLoaders.HasOne[w]; ok {
// 			err = relationLoader.Loader(tx, model)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
//
// 	return nil
// }

// func GenericBatchRelationsLoader[T any, U any](relationLoaders RelationLoaders[T, U], models []T, tx *sqlx.Tx, with map[string]any) error {
// 	var err error
//
// 	for k, v := range with {
// 		// BelongsTo
// 		if relationLoader, ok := relationLoaders.BelongsTo[k]; ok {
// 			err = relationLoader.BatchLoader(tx, models)
// 			if err != nil {
// 				return err
// 			}
// 		} else if relationLoader, ok := relationLoaders.HasOne[k]; ok {
// 			// HasOne
// 			err = relationLoader.BatchLoader(tx, models)
// 			if err != nil {
// 				return err
// 			}
// 		} else if relationLoader, ok := relationLoaders.HasMany[k]; ok {
// 			// HasMany
// 			err = relationLoader.BatchLoader(tx, models)
// 			if err != nil {
// 				return err
// 			}
// 		}
//
// 		if v != nil {
// 			// GenericBatchRelationsLoader(models)
// 		}
// 	}
//
// 	return nil
// }
