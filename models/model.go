package model

import (
	"time"

	"github.com/jesse-rb/imissphp-go"
)

type Modelable interface {
	Response() map[string]any
}

type Respondable interface {
	Response() map[string]any
}

type Model struct {
	ID        string     `json:"id" db:"id"`
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
