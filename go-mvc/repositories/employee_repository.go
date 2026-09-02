package repositories

import (
	"go-mvc/models"

	"github.com/beego/beego/v2/client/orm"
)

// EmployeeRepository defines the interface for employee data access
type EmployeeRepository interface {
	Create(employee *models.Employee) (int64, error)
	GetById(id int) (*models.Employee, error)
	GetAll(page, pageSize int) ([]*models.Employee, int64, error)
	Update(employee *models.Employee, fields ...string) error
	Delete(id int) error
	ExistsByEmail(email string, excludeId ...int) (bool, error)
}

// employeeRepository implements EmployeeRepository using Beego ORM
type employeeRepository struct{}

// NewEmployeeRepository creates a new repository instance
func NewEmployeeRepository() EmployeeRepository {
	return &employeeRepository{}
}

func (r *employeeRepository) orm() orm.Ormer {
	return orm.NewOrm()
}

// Create inserts a new employee and returns the generated ID.
func (r *employeeRepository) Create(employee *models.Employee) (int64, error) {
	id, err := r.orm().Insert(employee)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetById retrieves an employee by their primary key ID.
func (r *employeeRepository) GetById(id int) (*models.Employee, error) {
	employee := &models.Employee{Id: id}
	err := r.orm().Read(employee)
	if err != nil {
		return nil, err
	}
	return employee, nil
}

// GetAll retrieves a paginated list of employees ordered by created_at descending and the total count.
func (r *employeeRepository) GetAll(page, pageSize int) ([]*models.Employee, int64, error) {
	var employees []*models.Employee
	qs := r.orm().QueryTable(new(models.Employee))

	count, err := qs.Count()
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	_, err = qs.OrderBy("-created_at").Limit(pageSize, offset).All(&employees)
	if err != nil {
		return nil, 0, err
	}

	return employees, count, nil
}

// Update updates an employee's specified fields or all fields if none are specified.
func (r *employeeRepository) Update(employee *models.Employee, fields ...string) error {
	_, err := r.orm().Update(employee, fields...)
	return err
}

// Delete deletes an employee by their ID.
func (r *employeeRepository) Delete(id int) error {
	employee := &models.Employee{Id: id}
	_, err := r.orm().Delete(employee)
	return err
}

// ExistsByEmail checks if an employee with the specified email exists, optionally excluding a given ID.
func (r *employeeRepository) ExistsByEmail(email string, excludeId ...int) (bool, error) {
	qs := r.orm().QueryTable(new(models.Employee)).Filter("email", email)
	if len(excludeId) > 0 && excludeId[0] > 0 {
		qs = qs.Exclude("id", excludeId[0])
	}
	return qs.Exist(), nil
}
