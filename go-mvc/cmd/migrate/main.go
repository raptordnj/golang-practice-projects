package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/server/web"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"

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
	// Try loading from app.conf if available
	if err := web.LoadAppConfig("ini", "conf/app.conf"); err != nil {
		_ = web.LoadAppConfig("ini", "../../conf/app.conf")
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		host, _ = web.AppConfig.String("db::host")
	}
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port, _ = web.AppConfig.String("db::port")
	}
	if port == "" {
		port = "3306"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user, _ = web.AppConfig.String("db::user")
	}
	if user == "" {
		user = "root"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password, _ = web.AppConfig.String("db::password")
	}

	name := os.Getenv("DB_NAME")
	if name == "" {
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

func getMigrator(cfg DBConfig) (*migrate.Migrate, *sql.DB, error) {
	if err := ensureDatabaseExists(cfg); err != nil {
		return nil, nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create mysql migration driver: %w", err)
	}

	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create iofs source: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, cfg.Name, driver)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to initialize migrator: %w", err)
	}

	return m, db, nil
}

func printUsage() {
	fmt.Println("Usage: go run cmd/migrate/main.go <command> [arguments]")
	fmt.Println("\nCommands:")
	fmt.Println("  up              Apply all pending migrations")
	fmt.Println("  up <n>          Apply next n migrations")
	fmt.Println("  down            Rollback all migrations")
	fmt.Println("  down <n>        Rollback n migrations")
	fmt.Println("  version         Print current migration version and dirty status")
	fmt.Println("  force <version> Force set migration version (recovers from dirty state)")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := strings.ToLower(args[0])
	cfg := loadConfig()

	m, db, err := getMigrator(cfg)
	if err != nil {
		log.Fatalf("Migration initialization error: %v", err)
	}
	defer db.Close()

	switch cmd {
	case "up":
		if len(args) > 1 {
			steps, err := strconv.Atoi(args[1])
			if err != nil {
				log.Fatalf("Invalid steps value: %s", args[1])
			}
			err = m.Steps(steps)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("Migration up (%d steps) failed: %v", steps, err)
			}
			fmt.Printf("Successfully applied %d migration(s)\n", steps)
		} else {
			err = m.Up()
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("Migration up failed: %v", err)
			}
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No new migrations to apply (database is up to date).")
			} else {
				fmt.Println("All migrations applied successfully.")
			}
		}

	case "down":
		if len(args) > 1 {
			steps, err := strconv.Atoi(args[1])
			if err != nil {
				log.Fatalf("Invalid steps value: %s", args[1])
			}
			err = m.Steps(-steps)
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("Migration down (%d steps) failed: %v", steps, err)
			}
			fmt.Printf("Successfully rolled back %d migration(s)\n", steps)
		} else {
			err = m.Down()
			if err != nil && !errors.Is(err, migrate.ErrNoChange) {
				log.Fatalf("Migration down failed: %v", err)
			}
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to rollback.")
			} else {
				fmt.Println("All migrations rolled back successfully.")
			}
		}

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("Current version: None (no migrations applied)")
				return
			}
			log.Fatalf("Failed to retrieve version: %v", err)
		}
		fmt.Printf("Current version: %d (dirty: %t)\n", version, dirty)

	case "force":
		if len(args) < 2 {
			log.Fatal("Error: force command requires a version number. Usage: go run cmd/migrate/main.go force <version>")
		}
		version, err := strconv.Atoi(args[1])
		if err != nil {
			log.Fatalf("Invalid version number: %s", args[1])
		}
		if err := m.Force(version); err != nil {
			log.Fatalf("Failed to force version %d: %v", version, err)
		}
		fmt.Printf("Successfully forced version to %d\n", version)

	default:
		printUsage()
		os.Exit(1)
	}
}
