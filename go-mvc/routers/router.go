package routers

import (
	"go-mvc/controllers"
	"go-mvc/middlewares"
	"go-mvc/repositories"
	"go-mvc/services"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

func init() {
	// CORS filter
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))

	// Dependency injection
	employeeRepo := repositories.NewEmployeeRepository()
	employeeService := services.NewEmployeeService(employeeRepo)
	employeeController := &controllers.EmployeeController{
		Service: employeeService,
	}

	userRepo := repositories.NewUserRepository()
	authService := services.NewAuthService(userRepo, "")
	authController := &controllers.AuthController{
		Service: authService,
	}

	// JWT Auth filter for protected routes
	jwtFilter := middlewares.JWTAuthFilter(authService)
	web.InsertFilter("/api/v1/employees", web.BeforeRouter, jwtFilter)
	web.InsertFilter("/api/v1/employees/*", web.BeforeRouter, jwtFilter)
	web.InsertFilter("/api/v1/auth/me", web.BeforeRouter, jwtFilter)

	// API v1 namespace
	ns := web.NewNamespace("/api/v1",
		web.NSNamespace("/auth",
			web.NSRouter("/register", authController, "post:Register"),
			web.NSRouter("/login", authController, "post:Login"),
			web.NSRouter("/me", authController, "get:Me"),
		),
		web.NSNamespace("/employees",
			web.NSRouter("/", employeeController, "post:Create;get:GetAll"),
			web.NSRouter("/:id", employeeController, "get:GetOne;put:Update;delete:Delete"),
		),
	)

	web.AddNamespace(ns)
}
