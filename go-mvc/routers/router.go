package routers

import (
	"go-mvc/controllers"
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
	repo := repositories.NewEmployeeRepository()
	service := services.NewEmployeeService(repo)

	employeeController := &controllers.EmployeeController{
		Service: service,
	}

	// API v1 namespace
	ns := web.NewNamespace("/api/v1",
		web.NSNamespace("/employees",
			web.NSRouter("/", employeeController, "post:Create;get:GetAll"),
			web.NSRouter("/:id", employeeController, "get:GetOne;put:Update;delete:Delete"),
		),
	)

	web.AddNamespace(ns)
}
