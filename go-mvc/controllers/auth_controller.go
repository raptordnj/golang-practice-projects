package controllers

import (
	"encoding/json"
	"strconv"

	"go-mvc/dto"
	"go-mvc/services"

	"github.com/beego/beego/v2/server/web"
)

// AuthController handles user authentication endpoints
type AuthController struct {
	web.Controller
	Service services.AuthService
}

// Register handles POST /api/v1/auth/register
func (c *AuthController) Register() {
	var req dto.RegisterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	res, err := c.Service.Register(&req)
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
		Message: "User registered successfully",
		Data:    res,
	}
	c.ServeJSON()
}

// Login handles POST /api/v1/auth/login
func (c *AuthController) Login() {
	var req dto.LoginRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	res, err := c.Service.Login(&req)
	if err != nil {
		c.Ctx.Output.SetStatus(401)
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
		Message: "Login successful",
		Data:    res,
	}
	c.ServeJSON()
}

// Me handles GET /api/v1/auth/me
func (c *AuthController) Me() {
	userIDVal := c.Ctx.Input.GetData("userID")
	if userIDVal == nil {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Unauthorized",
		}
		c.ServeJSON()
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		if idStr, okStr := userIDVal.(string); okStr {
			userID, _ = strconv.Atoi(idStr)
		}
	}

	user, err := c.Service.GetUserByID(userID)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "User not found",
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = dto.APIResponse{
		Success: true,
		Data:    user,
	}
	c.ServeJSON()
}
