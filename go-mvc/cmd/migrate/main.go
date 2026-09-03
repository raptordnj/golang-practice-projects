package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"

	"go-mvc/cmd/migrate/migrator"
	"go-mvc/migrations"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func loadConfig() DBConfig {
	if err := web.LoadAppConfig("ini", "conf/app.conf"); err != nil {
		_ = web.LoadAppConfig("ini", "../../conf/app.conf")
	}

	host := os.Getenv("DB_HOST")
	if host == "" && web.AppConfig != nil {
		host, _ = web.AppConfig.String("db::host")
	}
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("DB_PORT")
	if port == "" && web.AppConfig != nil {
		port, _ = web.AppConfig.String("db::port")
	}
	if port == "" {
		port = "3306"
	}

	user := os.Getenv("DB_USER")
	if user == "" && web.AppConfig != nil {
		user, _ = web.AppConfig.String("db::user")
	}
	if user == "" {
		user = "root"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" && web.AppConfig != nil {
		password, _ = web.AppConfig.String("db::password")
	}

	name := os.Getenv("DB_NAME")
	if name == "" && web.AppConfig != nil {
		name, _ = web.AppConfig.String("db::name")
	}
	if name == "" {
		name = "employee_db"
	}

	return DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Name:     name,
	}
}

func ensureDatabaseExists(cfg DBConfig) error {
	serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
	db, err := sql.Open("mysql", serverDSN)
	if err != nil {
		return fmt.Errorf("failed to connect to mysql server: %w", err)
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", cfg.Name)
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("failed to create database %s: %w", cfg.Name, err)
	}
	return nil
}

func printUsage() {
	fmt.Println("WorkPulse Laravel-Style Migration CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  go run cmd/migrate/main.go <command> [options]")
	fmt.Println("\nAvailable Commands:")
	fmt.Println("  migrate (up)                  Run pending migrations (creates a new batch)")
	fmt.Println("  migrate:rollback (down)       Rollback the last migration batch")
	fmt.Println("    options: --step=N           Rollback the last N migrations")
	fmt.Println("  migrate:status (status)       Show the status of each migration (Yes/No, Batch)")
	fmt.Println("  migrate:reset (reset)         Rollback all database migrations")
	fmt.Println("  migrate:refresh (refresh)     Reset and re-run all migrations")
	fmt.Println("  migrate:fresh (fresh)         Drop all tables and re-run all migrations")
	fmt.Println("  make:migration <name>         Create a new migration file with timestamp")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(args[0])

	// Handle make:migration without needing DB connection
	if cmd == "make:migration" || cmd == "make" {
		if len(args) < 2 {
			log.Fatal("Error: Migration name required. Usage: go run cmd/migrate/main.go make:migration <name>")
		}
		name := args[1]
		migrationsDir := "migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			migrationsDir = "../../migrations"
		}
		upFile, _, err := migrator.MakeMigration(migrationsDir, name)
		if err != nil {
			log.Fatalf("Failed to make migration: %v", err)
		}
		fmt.Printf("Created Migration: %s\n", filepath.Base(strings.TrimSuffix(upFile, ".up.sql")))
		return
	}

	cfg := loadConfig()
	if err := ensureDatabaseExists(cfg); err != nil {
		log.Fatalf("Failed to ensure database exists: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	mgr := migrator.New(db, migrations.FS)

	switch cmd {
	case "migrate", "up":
		if err := mgr.Migrate(); err != nil {
			log.Fatalf("Migrate failed: %v", err)
		}

	case "migrate:rollback", "rollback", "down":
		step := 0
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "--step=") {
				s, err := strconv.Atoi(strings.TrimPrefix(arg, "--step="))
				if err == nil && s > 0 {
					step = s
				}
			} else if s, err := strconv.Atoi(arg); err == nil && s > 0 {
				step = s
			}
		}
		if err := mgr.Rollback(step); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}

	case "migrate:status", "status":
		statuses, err := mgr.Status()
		if err != nil {
			log.Fatalf("Failed to retrieve migration status: %v", err)
		}
		fmt.Println(migrator.RenderStatusTable(statuses))

	case "migrate:reset", "reset":
		if err := mgr.Reset(); err != nil {
			log.Fatalf("Reset failed: %v", err)
		}

	case "migrate:refresh", "refresh":
		step := 0
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "--step=") {
				s, err := strconv.Atoi(strings.TrimPrefix(arg, "--step="))
				if err == nil && s > 0 {
					step = s
				}
			} else if s, err := strconv.Atoi(arg); err == nil && s > 0 {
				step = s
			}
		}
		if err := mgr.Refresh(step); err != nil {
			log.Fatalf("Refresh failed: %v", err)
		}

	case "migrate:fresh", "fresh":
		if err := mgr.Fresh(cfg.Name); err != nil {
			log.Fatalf("Fresh failed: %v", err)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}
