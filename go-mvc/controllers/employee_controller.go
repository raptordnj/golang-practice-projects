package controllers

import (
	"encoding/json"
	"strconv"

	"go-mvc/dto"
	"go-mvc/services"

	"github.com/beego/beego/v2/server/web"
)

// EmployeeController handles employee CRUD operations
type EmployeeController struct {
	web.Controller
	Service services.EmployeeService
}

// Create handles POST /api/v1/employees to create a new employee
func (c *EmployeeController) Create() {
	var req dto.CreateEmployeeRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	response, err := c.Service.CreateEmployee(&req)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = dto.APIResponse{
		Success: true,
		Message: "Employee created successfully",
		Data:    response,
	}
	c.ServeJSON()
}

// GetOne handles GET /api/v1/employees/:id to retrieve an employee by ID
func (c *EmployeeController) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid employee ID",
		}
		c.ServeJSON()
		return
	}

	response, err := c.Service.GetEmployee(id)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = dto.APIResponse{
		Success: true,
		Data:    response,
	}
	c.ServeJSON()
}

// GetAll handles GET /api/v1/employees with pagination query parameters
func (c *EmployeeController) GetAll() {
	page := 1
	pageSize := 10

	if pageStr := c.GetString("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.GetString("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	employees, totalCount, totalPages, err := c.Service.GetAllEmployees(page, pageSize)
	if err != nil {
		c.Ctx.Output.SetStatus(500)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = dto.PaginatedResponse{
		Success:    true,
		Message:    "Employees retrieved successfully",
		Data:       employees,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
	}
	c.ServeJSON()
}

// Update handles PUT /api/v1/employees/:id to update an existing employee
func (c *EmployeeController) Update() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid employee ID",
		}
		c.ServeJSON()
		return
	}

	var req dto.UpdateEmployeeRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	response, err := c.Service.UpdateEmployee(id, &req)
	if err != nil {
		if err.Error() == "employee not found" {
			c.Ctx.Output.SetStatus(404)
		} else {
			c.Ctx.Output.SetStatus(400)
		}
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = dto.APIResponse{
		Success: true,
		Message: "Employee updated successfully",
		Data:    response,
	}
	c.ServeJSON()
}

// Delete handles DELETE /api/v1/employees/:id to delete an employee
func (c *EmployeeController) Delete() {
	idStr := c.Ctx.Input.Param(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid employee ID",
		}
		c.ServeJSON()
		return
	}

	err = c.Service.DeleteEmployee(id)
	if err != nil {
		if err.Error() == "employee not found" {
			c.Ctx.Output.SetStatus(404)
		} else {
			c.Ctx.Output.SetStatus(500)
		}
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = dto.APIResponse{
		Success: true,
		Message: "Employee deleted successfully",
	}
	c.ServeJSON()
}
