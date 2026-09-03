package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// User represents the user model in the database
type User struct {
	Id           int       `orm:"column(id);auto;pk" json:"id"`
	Name         string    `orm:"column(name);size(100)" json:"name" valid:"Required"`
	Email        string    `orm:"column(email);size(150);unique" json:"email" valid:"Required;Email"`
	PasswordHash string    `orm:"column(password_hash);size(255)" json:"-"`
	CreatedAt    time.Time `orm:"column(created_at);auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt    time.Time `orm:"column(updated_at);auto_now;type(datetime)" json:"updated_at"`
}

func init() {
	orm.RegisterModel(new(User))
}

// TableName returns the database table name for User model
func (u *User) TableName() string {
	return "users"
}
