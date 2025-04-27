package model

import (
	"time"

	"github.com/jesse-rb/imissphp-go"
	"github.com/jmoiron/sqlx"
)

type (
	FuncLoadBelongsTo[T any] func(tx *sqlx.Tx, m T) error
	FuncLoadHasOne[T any]    func(tx *sqlx.Tx, m T) error
	FuncLoadHasMany[T any]   func(tx *sqlx.Tx, m T) error

	FuncBatchLoadBelongsTo[T any] func(tx *sqlx.Tx, m []T) error
	FuncBatchLoadHasOne[T any]    func(tx *sqlx.Tx, m []T) error
	FuncBatchLoadHasMany[T any]   func(tx *sqlx.Tx, m []T) error
)

type BelongsToLoader[T any] struct {
	Loader      FuncLoadBelongsTo[T]
	BatchLoader FuncBatchLoadBelongsTo[T]
}
type HasOneLoader[T any] struct {
	Loader      FuncLoadHasOne[T]
	BatchLoader FuncBatchLoadHasOne[T]
}
type HasManyLoader[T any] struct {
	Loader      FuncLoadHasMany[T]
	BatchLoader FuncBatchLoadHasMany[T]
}

type RelationLoaders[T any] struct {
	BelongsTo map[string]BelongsToLoader[T]
	HasMany   map[string]HasManyLoader[T]
	HasOne    map[string]HasOneLoader[T]
}

type Modelable interface {
	LoadRelations(tx *sqlx.Tx, with []string) error
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

func genericRelationsLoader[T any](relationLoaders RelationLoaders[T], model T, tx *sqlx.Tx, with []string) error {
	var err error

	for _, w := range with {
		// BelongsTo
		if relationLoader, ok := relationLoaders.BelongsTo[w]; ok {
			err = relationLoader.Loader(tx, model)
			if err != nil {
				return err
			}
		}
		// HasOne
		if relationLoader, ok := relationLoaders.HasMany[w]; ok {
			err = relationLoader.Loader(tx, model)
			if err != nil {
				return err
			}
		}
		// HasMany
		if relationLoader, ok := relationLoaders.HasOne[w]; ok {
			err = relationLoader.Loader(tx, model)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func genericBatchRelationsLoader[T any](relationLoaders RelationLoaders[T], models []T, tx *sqlx.Tx, with []string) error {
	var err error

	for _, w := range with {
		// BelongsTo
		if relationLoader, ok := relationLoaders.BelongsTo[w]; ok {
			err = relationLoader.BatchLoader(tx, models)
			if err != nil {
				return err
			}
		}
		// HasOne
		if relationLoader, ok := relationLoaders.HasOne[w]; ok {
			err = relationLoader.BatchLoader(tx, models)
			if err != nil {
				return err
			}
		}
		// HasMany
		if relationLoader, ok := relationLoaders.HasMany[w]; ok {
			err = relationLoader.BatchLoader(tx, models)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
