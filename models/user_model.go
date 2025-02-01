package model

type User struct {
	Uuid  string `json:"uuid" gorm:"column:id;primaryKey;type:uuid"`
	Email string `json:"email"`
}

func (u User) TableName() string {
	return "auth.users"
}
