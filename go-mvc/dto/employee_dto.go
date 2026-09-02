package dto

import "time"

// CreateEmployeeRequest is the DTO for creating a new employee
type CreateEmployeeRequest struct {
	FirstName  string  `json:"first_name" valid:"Required"`
	LastName   string  `json:"last_name" valid:"Required"`
	Email      string  `json:"email" valid:"Required;Email"`
	Phone      string  `json:"phone"`
	Department string  `json:"department" valid:"Required"`
	Position   string  `json:"position" valid:"Required"`
	Salary     float64 `json:"salary" valid:"Required"`
	HireDate   string  `json:"hire_date"` // format: 2006-01-02
}

// UpdateEmployeeRequest is the DTO for updating an employee
type UpdateEmployeeRequest struct {
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Email      string  `json:"email" valid:"Email"`
	Phone      string  `json:"phone"`
	Department string  `json:"department"`
	Position   string  `json:"position"`
	Salary     float64 `json:"salary"`
	HireDate   string  `json:"hire_date"`
	IsActive   *bool   `json:"is_active"`
}

// EmployeeResponse is the DTO for employee API responses
type EmployeeResponse struct {
	Id         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Department string    `json:"department"`
	Position   string    `json:"position"`
	Salary     float64   `json:"salary"`
	HireDate   string    `json:"hire_date"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// APIResponse is a generic API response wrapper
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse wraps paginated list responses
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int64       `json:"total_count"`
	TotalPages int         `json:"total_pages"`
}
