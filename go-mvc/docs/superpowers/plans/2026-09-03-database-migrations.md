# Database Migrations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide an explicit, version-controlled database migration system using `golang-migrate/migrate/v4` and a Go CLI runner `cmd/migrate/main.go` for the `go-mvc` application.

**Architecture:** Versioned SQL files in `migrations/` applied to MySQL via `golang-migrate/migrate/v4` database driver and file/iofs source. A CLI tool in `cmd/migrate/main.go` reads configuration from `conf/app.conf` or environment variables, ensures database existence, and handles `up`, `down`, `version`, and `force` commands.

**Tech Stack:** Go 1.23, MySQL 8, Beego v2, `github.com/golang-migrate/migrate/v4`.

## Global Constraints
- Target database: MySQL with `utf8mb4` encoding and `utf8mb4_unicode_ci` collation.
- Database credentials loaded from `conf/app.conf` with fallback to standard environment variables (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`).
- Backward compatible with existing Beego models and routes.

---

### Task 1: Add Migration Dependencies and Initial SQL Files

**Files:**
- Modify: `go.mod`
- Create: `migrations/000001_create_employees_table.up.sql`
- Create: `migrations/000001_create_employees_table.down.sql`

**Interfaces:**
- Consumes: None
- Produces: `migrations/*.sql` schema definitions for `employees` table.

- [ ] **Step 1: Install `golang-migrate/migrate/v4` dependencies**

Run:
```bash
go get github.com/golang-migrate/migrate/v4
go get github.com/golang-migrate/migrate/v4/database/mysql
go get github.com/golang-migrate/migrate/v4/source/file
go get github.com/golang-migrate/migrate/v4/source/iofs
go mod tidy
```

- [ ] **Step 2: Create forward migration `migrations/000001_create_employees_table.up.sql`**

```sql
CREATE TABLE IF NOT EXISTS `employees` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `first_name` VARCHAR(100) NOT NULL,
    `last_name` VARCHAR(100) NOT NULL,
    `email` VARCHAR(150) NOT NULL UNIQUE,
    `phone` VARCHAR(20) NULL,
    `department` VARCHAR(100) NOT NULL,
    `position` VARCHAR(100) NOT NULL,
    `salary` DECIMAL(12, 2) NOT NULL,
    `hire_date` DATE NOT NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

- [ ] **Step 3: Create backward migration `migrations/000001_create_employees_table.down.sql`**

```sql
DROP TABLE IF EXISTS `employees`;
```

- [ ] **Step 4: Commit**

```bash
git add go.mod go.sum migrations/
git commit -m "feat(migrations): add golang-migrate dependencies and initial employee schema"
```

---

### Task 2: Implement Migration CLI Runner

**Files:**
- Create: `cmd/migrate/main.go`

**Interfaces:**
- Consumes: `conf/app.conf`, `migrations/*.sql`
- Produces: CLI commands `up`, `down`, `version`, `force`

- [ ] **Step 1: Write `cmd/migrate/main.go`**

```go
package main

import (
	"database/sql"
	"embed"
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
)

//go:embed all:../../migrations/*.sql
var migrationFiles embed.FS

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

func loadConfig() DBConfig {
	// Try loading from app.conf if available
	_ = web.LoadAppConfig("ini", "conf/app.conf")

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

	d, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		// Fallback for path prefix if nested
		d, err = iofs.New(migrationFiles, "../../migrations")
		if err != nil {
			db.Close()
			return nil, nil, fmt.Errorf("failed to create iofs source: %w", err)
		}
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
```

- [ ] **Step 2: Verify `cmd/migrate/main.go` builds cleanly**

Run: `go build -o /dev/null ./cmd/migrate`
Expected: Exits 0 with no compiler errors.

- [ ] **Step 3: Commit**

```bash
git add cmd/migrate/main.go
git commit -m "feat(migrations): implement CLI migration runner"
```

---

### Task 3: Test and Verify Database Migrations & Rollbacks

**Files:**
- Modify: `conf/app.conf` (if needed)

**Interfaces:**
- Consumes: `cmd/migrate/main.go`, `migrations/`
- Produces: Active MySQL tables and verification

- [ ] **Step 1: Test `go run cmd/migrate/main.go up`**

Run: `go run cmd/migrate/main.go up`
Expected: Output indicates all migrations applied successfully or database is up to date.

- [ ] **Step 2: Test `go run cmd/migrate/main.go version`**

Run: `go run cmd/migrate/main.go version`
Expected: Output indicates `Current version: 1 (dirty: false)`.

- [ ] **Step 3: Test Rollback and Re-up**

Run:
```bash
go run cmd/migrate/main.go down 1
go run cmd/migrate/main.go version
go run cmd/migrate/main.go up
go run cmd/migrate/main.go version
```
Expected: Version transitions from 1 -> None -> 1.

- [ ] **Step 4: Commit**

```bash
git commit --allow-empty -m "test(migrations): verify up, down, and version migration lifecycle"
```

---

### Task 4: Documentation and Main App Integration

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: Migration CLI
- Produces: Updated documentation in `README.md`

- [ ] **Step 1: Update `README.md` with Database Migration instructions**

Add documentation detailing how to run migrations (`go run cmd/migrate/main.go up`, `down`, `version`, `force`).

- [ ] **Step 2: Verify all Go packages compile and tests pass**

Run: `go test ./...` and `go vet ./...`

- [ ] **Step 3: Commit**

```bash
git add README.md
git commit -m "docs: document database migration commands in README"
```
