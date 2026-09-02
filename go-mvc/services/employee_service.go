package services

import (
	"errors"
	"math"
	"time"

	"go-mvc/dto"
	"go-mvc/models"
	"go-mvc/repositories"
)

// EmployeeService defines the interface for employee business logic
type EmployeeService interface {
	CreateEmployee(req *dto.CreateEmployeeRequest) (*dto.EmployeeResponse, error)
	GetEmployee(id int) (*dto.EmployeeResponse, error)
	GetAllEmployees(page, pageSize int) ([]*dto.EmployeeResponse, int64, int, error)
	UpdateEmployee(id int, req *dto.UpdateEmployeeRequest) (*dto.EmployeeResponse, error)
	DeleteEmployee(id int) error
}

type employeeService struct {
	repo repositories.EmployeeRepository
}

func NewEmployeeService(repo repositories.EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

// CreateEmployee validates business rules and creates a new employee
func (s *employeeService) CreateEmployee(req *dto.CreateEmployeeRequest) (*dto.EmployeeResponse, error) {
	exists, err := s.repo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("employee with this email already exists")
	}

	hireDate := time.Now()
	if req.HireDate != "" {
		if parsed, err := time.Parse("2006-01-02", req.HireDate); err == nil {
			hireDate = parsed
		}
	}

	emp := &models.Employee{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Phone:      req.Phone,
		Department: req.Department,
		Position:   req.Position,
		Salary:     req.Salary,
		HireDate:   hireDate,
		IsActive:   true,
	}

	id, err := s.repo.Create(emp)
	if err != nil {
		return nil, err
	}

	emp.Id = int(id)
	return toEmployeeResponse(emp), nil
}

// GetEmployee retrieves an employee by ID
func (s *employeeService) GetEmployee(id int) (*dto.EmployeeResponse, error) {
	emp, err := s.repo.GetById(id)
	if err != nil || emp == nil {
		return nil, errors.New("employee not found")
	}

	return toEmployeeResponse(emp), nil
}

// GetAllEmployees retrieves paginated employees
func (s *employeeService) GetAllEmployees(page, pageSize int) ([]*dto.EmployeeResponse, int64, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	employees, totalCount, err := s.repo.GetAll(page, pageSize)
	if err != nil {
		return nil, 0, 0, err
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(pageSize)))
	}

	responses := make([]*dto.EmployeeResponse, 0, len(employees))
	for _, emp := range employees {
		responses = append(responses, toEmployeeResponse(emp))
	}

	return responses, totalCount, totalPages, nil
}

// UpdateEmployee updates existing employee details
func (s *employeeService) UpdateEmployee(id int, req *dto.UpdateEmployeeRequest) (*dto.EmployeeResponse, error) {
	emp, err := s.repo.GetById(id)
	if err != nil || emp == nil {
		return nil, errors.New("employee not found")
	}

	var fields []string

	if req.FirstName != "" {
		emp.FirstName = req.FirstName
		fields = append(fields, "FirstName")
	}
	if req.LastName != "" {
		emp.LastName = req.LastName
		fields = append(fields, "LastName")
	}
	if req.Email != "" && req.Email != emp.Email {
		exists, err := s.repo.ExistsByEmail(req.Email, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("employee with this email already exists")
		}
		emp.Email = req.Email
		fields = append(fields, "Email")
	}
	if req.Phone != "" {
		emp.Phone = req.Phone
		fields = append(fields, "Phone")
	}
	if req.Department != "" {
		emp.Department = req.Department
		fields = append(fields, "Department")
	}
	if req.Position != "" {
		emp.Position = req.Position
		fields = append(fields, "Position")
	}
	if req.Salary > 0 {
		emp.Salary = req.Salary
		fields = append(fields, "Salary")
	}
	if req.HireDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.HireDate)
		if err != nil {
			return nil, errors.New("invalid hire date format, expected YYYY-MM-DD")
		}
		emp.HireDate = parsedDate
		fields = append(fields, "HireDate")
	}
	if req.IsActive != nil {
		emp.IsActive = *req.IsActive
		fields = append(fields, "IsActive")
	}

	if len(fields) > 0 {
		if err := s.repo.Update(emp, fields...); err != nil {
			return nil, err
		}
	}

	updatedEmp, err := s.repo.GetById(id)
	if err != nil || updatedEmp == nil {
		return nil, errors.New("employee not found")
	}

	return toEmployeeResponse(updatedEmp), nil
}

// DeleteEmployee removes an employee by ID after checking existence
func (s *employeeService) DeleteEmployee(id int) error {
	emp, err := s.repo.GetById(id)
	if err != nil || emp == nil {
		return errors.New("employee not found")
	}

	return s.repo.Delete(id)
}

// toEmployeeResponse maps a models.Employee entity to a dto.EmployeeResponse
func toEmployeeResponse(emp *models.Employee) *dto.EmployeeResponse {
	if emp == nil {
		return nil
	}

	return &dto.EmployeeResponse{
		Id:         emp.Id,
		FirstName:  emp.FirstName,
		LastName:   emp.LastName,
		Email:      emp.Email,
		Phone:      emp.Phone,
		Department: emp.Department,
		Position:   emp.Position,
		Salary:     emp.Salary,
		HireDate:   emp.HireDate.Format("2006-01-02"),
		IsActive:   emp.IsActive,
		CreatedAt:  emp.CreatedAt,
		UpdatedAt:  emp.UpdatedAt,
	}
}
