package main

import (
	"database/sql"
	"fmt"

	_ "go-mvc/models"
	_ "go-mvc/routers"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	dbHost, _ := web.AppConfig.String("db::host")
	if dbHost == "" {
		dbHost = "127.0.0.1"
	}
	dbPort, _ := web.AppConfig.String("db::port")
	if dbPort == "" {
		dbPort = "3306"
	}
	dbUser, _ := web.AppConfig.String("db::user")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPassword, _ := web.AppConfig.String("db::password")
	dbName, _ := web.AppConfig.String("db::name")
	if dbName == "" {
		dbName = "employee_db"
	}

	// Auto-create database if it does not exist
	serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbUser, dbPassword, dbHost, dbPort)
	if db, err := sql.Open("mysql", serverDSN); err == nil {
		_, _ = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbName))
		_ = db.Close()
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)

	orm.RegisterDriver("mysql", orm.DRMySQL)
	orm.RegisterDataBase("default", "mysql", dsn)
	orm.RunSyncdb("default", false, true)
}

func main() {
	web.Run()
}
