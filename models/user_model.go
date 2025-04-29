package model

import (
	"github.com/jesse-rb/imissphp-go"
)

type User struct {
	Uuid     string   `json:"uuid" gorm:"column:id;primaryKey;type:uuid" db:"id"`
	Email    string   `json:"email" db:"email"`
	Canvases []Canvas `json:"canvases"`
}

func (u User) Response() map[string]any {
	r := imissphp.ToMap(u)
	return r
}
