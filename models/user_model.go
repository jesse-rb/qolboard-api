package model

type User struct {
	Uuid  string `gorm:"column:id;primaryKey;type:uuid"`
	Email string
}

func (u User) TableName() string {
	return "auth.users"
}
