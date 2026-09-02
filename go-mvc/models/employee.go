package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// Employee represents the employee model in the database
type Employee struct {
	Id         int       `orm:"column(id);auto;pk" json:"id"`
	FirstName  string    `orm:"column(first_name);size(100)" json:"first_name" valid:"Required"`
	LastName   string    `orm:"column(last_name);size(100)" json:"last_name" valid:"Required"`
	Email      string    `orm:"column(email);size(150);unique" json:"email" valid:"Required;Email"`
	Phone      string    `orm:"column(phone);size(20);null" json:"phone"`
	Department string    `orm:"column(department);size(100)" json:"department" valid:"Required"`
	Position   string    `orm:"column(position);size(100)" json:"position" valid:"Required"`
	Salary     float64   `orm:"column(salary);digits(12);decimals(2)" json:"salary" valid:"Required"`
	HireDate   time.Time `orm:"column(hire_date);type(date)" json:"hire_date"`
	IsActive   bool      `orm:"column(is_active);default(true)" json:"is_active"`
	CreatedAt  time.Time `orm:"column(created_at);auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt  time.Time `orm:"column(updated_at);auto_now;type(datetime)" json:"updated_at"`
}

func init() {
	orm.RegisterModel(new(Employee))
}

// TableName returns the database table name for Employee model
func (e *Employee) TableName() string {
	return "employees"
}
