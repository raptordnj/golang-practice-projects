# Laravel-Style Database Migrations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform the WorkPulse migration system into a full Laravel Artisan-style migration suite featuring timestamped migrations, a batch-tracking `migrations` table, and commands (`migrate`, `migrate:rollback`, `migrate:status`, `migrate:reset`, `migrate:refresh`, `migrate:fresh`, `make:migration`).

**Architecture:** 
- Embedded timestamped `.up.sql` and `.down.sql` files under `migrations/`.
- Reusable Go migration engine in `cmd/migrate/migrator/` managing the `migrations` table (`id`, `migration`, `batch`) and implementing batch execution and rollback.
- CLI argument dispatcher in `cmd/migrate/main.go` providing intuitive Artisan-style commands (`migrate:*`, `make:migration`, and shorthand aliases).

**Tech Stack:** Go 1.24, MySQL 8, `embed.FS`, standard Go `database/sql`.

## Global Constraints
- Target database: MySQL with `utf8mb4` encoding.
- Migration tracking table name: `migrations` with schema `id INT AUTO_INCREMENT PRIMARY KEY, migration VARCHAR(255) NOT NULL UNIQUE, batch INT NOT NULL`.
- All migration names recorded in the database without file extensions (e.g. `2026_09_03_000001_create_employees_table`).
- Standalone Go execution: works via `go run cmd/migrate/main.go <command>` without external dependencies.
- Zero breaking changes to `Employee` and `User` schema or REST APIs.

---

### Task 1: Standardize Migration Files to Timestamp Format

**Files:**
- Rename: `migrations/000001_create_employees_table.up.sql` -> `migrations/2026_09_03_000001_create_employees_table.up.sql`
- Rename: `migrations/000001_create_employees_table.down.sql` -> `migrations/2026_09_03_000001_create_employees_table.down.sql`
- Rename: `migrations/000002_create_users_table.up.sql` -> `migrations/2026_09_03_000002_create_users_table.up.sql`
- Rename: `migrations/000002_create_users_table.down.sql` -> `migrations/2026_09_03_000002_create_users_table.down.sql`

**Interfaces:**
- Consumes: `migrations/fs.go` (`//go:embed *.sql`)
- Produces: Timestamped SQL migration assets.

- [ ] **Step 1: Rename migration files using git mv**

```bash
git mv migrations/000001_create_employees_table.up.sql migrations/2026_09_03_000001_create_employees_table.up.sql
git mv migrations/000001_create_employees_table.down.sql migrations/2026_09_03_000001_create_employees_table.down.sql
git mv migrations/000002_create_users_table.up.sql migrations/2026_09_03_000002_create_users_table.up.sql
git mv migrations/000002_create_users_table.down.sql migrations/2026_09_03_000002_create_users_table.down.sql
```

- [ ] **Step 2: Verify package compilation**

```bash
go build ./migrations/...
```
Expected: Compiles cleanly with embedded SQL files.

- [ ] **Step 3: Commit**

```bash
git add migrations/
git commit -m "feat(migrations): standardize migration files to timestamp convention"
```

---

### Task 2: Implement Migrator Engine & Output Formatter

**Files:**
- Create: `cmd/migrate/migrator/migrator.go`
- Create: `cmd/migrate/migrator/output.go`
- Create: `cmd/migrate/migrator/migrator_test.go`

**Interfaces:**
- Consumes: `database/sql`, `io/fs`, MySQL driver
- Produces: `Migrator` struct, `Migrate`, `Rollback`, `Reset`, `Refresh`, `Fresh`, `Status`, `MakeMigration` methods, and tabular output renderer.

- [ ] **Step 1: Create `cmd/migrate/migrator/output.go`**

```go
package migrator

import (
	"fmt"
	"strings"
)

// MigrationStatus represents status of a single migration
type MigrationStatus struct {
	Ran       bool
	Migration string
	Batch     *int
}

// RenderStatusTable formats migration statuses into an aligned ASCII table like Laravel
func RenderStatusTable(statuses []MigrationStatus) string {
	if len(statuses) == 0 {
		return "No migrations found."
	}

	maxNameLen := len("Migration")
	for _, s := range statuses {
		if len(s.Migration) > maxNameLen {
			maxNameLen = len(s.Migration)
		}
	}

	border := fmt.Sprintf("+------+-%s-+-------+", strings.Repeat("-", maxNameLen))
	header := fmt.Sprintf("| Ran? | %-*s | Batch |", maxNameLen, "Migration")

	var sb strings.Builder
	sb.WriteString(border + "\n")
	sb.WriteString(header + "\n")
	sb.WriteString(border + "\n")

	for _, s := range statuses {
		ranStr := "No"
		batchStr := "Pending"
		if s.Ran {
			ranStr = "Yes"
			if s.Batch != nil {
				batchStr = fmt.Sprintf("%d", *s.Batch)
			}
		}
		sb.WriteString(fmt.Sprintf("| %-4s | %-*s | %-5s |\n", ranStr, maxNameLen, s.Migration, batchStr))
	}

	sb.WriteString(border)
	return sb.String()
}
```

- [ ] **Step 2: Create unit tests in `cmd/migrate/migrator/migrator_test.go`**

```go
package migrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderStatusTable(t *testing.T) {
	batchOne := 1
	statuses := []MigrationStatus{
		{Ran: true, Migration: "2026_09_03_000001_create_employees_table", Batch: &batchOne},
		{Ran: false, Migration: "2026_09_03_000002_create_users_table", Batch: nil},
	}

	rendered := RenderStatusTable(statuses)
	if !strings.Contains(rendered, "Yes") || !strings.Contains(rendered, "Pending") {
		t.Errorf("Expected status table to contain Yes and Pending, got:\n%s", rendered)
	}
	if !strings.Contains(rendered, "2026_09_03_000001_create_employees_table") {
		t.Errorf("Expected status table to contain migration name, got:\n%s", rendered)
	}
}

func TestMakeMigration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "migrations-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	upPath, downPath, err := MakeMigration(tempDir, "create_roles_table")
	if err != nil {
		t.Fatalf("MakeMigration failed: %v", err)
	}

	if _, err := os.Stat(upPath); os.IsNotExist(err) {
		t.Errorf("Expected up migration file to exist: %s", upPath)
	}
	if _, err := os.Stat(downPath); os.IsNotExist(err) {
		t.Errorf("Expected down migration file to exist: %s", downPath)
	}

	upContent, _ := os.ReadFile(upPath)
	if !strings.Contains(string(upContent), "CREATE TABLE IF NOT EXISTS `roles`") {
		t.Errorf("Expected table creation template for roles, got:\n%s", string(upContent))
	}

	downContent, _ := os.ReadFile(downPath)
	if !strings.Contains(string(downContent), "DROP TABLE IF EXISTS `roles`") {
		t.Errorf("Expected table drop template for roles, got:\n%s", string(downContent))
	}
}

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"create_users_table", "users"},
		{"create_employee_records_table", "employee_records"},
		{"add_index_to_users", ""},
	}

	for _, tt := range tests {
		got := extractTableName(tt.input)
		if got != tt.expected {
			t.Errorf("extractTableName(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}
```

- [ ] **Step 3: Implement `cmd/migrate/migrator/migrator.go`**

```go
package migrator

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Migrator struct {
	db   *sql.DB
	fsys fs.FS
}

func New(db *sql.DB, fsys fs.FS) *Migrator {
	return &Migrator{
		db:   db,
		fsys: fsys,
	}
}

// EnsureMigrationsTable creates the migrations table if it does not exist
func (m *Migrator) EnsureMigrationsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS ` + "`migrations`" + ` (
		` + "`id`" + ` INT AUTO_INCREMENT PRIMARY KEY,
		` + "`migration`" + ` VARCHAR(255) NOT NULL UNIQUE,
		` + "`batch`" + ` INT NOT NULL
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;`
	_, err := m.db.Exec(query)
	return err
}

// GetAppliedMigrations returns a map of applied migration names to their batch numbers
func (m *Migrator) GetAppliedMigrations() (map[string]int, error) {
	if err := m.EnsureMigrationsTable(); err != nil {
		return nil, err
	}

	rows, err := m.db.Query("SELECT `migration`, `batch` FROM `migrations` ORDER BY `id` ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]int)
	for rows.Next() {
		var name string
		var batch int
		if err := rows.Scan(&name, &batch); err != nil {
			return nil, err
		}
		applied[name] = batch
	}
	return applied, nil
}

// GetAvailableMigrations finds all migration names ending in .up.sql from the embedded or provided filesystem
func (m *Migrator) GetAvailableMigrations() ([]string, error) {
	var migrations []string
	entries, err := fs.ReadDir(m.fsys, ".")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			name := strings.TrimSuffix(entry.Name(), ".up.sql")
			migrations = append(migrations, name)
		}
	}
	sort.Strings(migrations)
	return migrations, nil
}

// GetNextBatchNumber returns the next batch number (MAX(batch) + 1)
func (m *Migrator) GetNextBatchNumber() (int, error) {
	if err := m.EnsureMigrationsTable(); err != nil {
		return 1, err
	}

	var maxBatch sql.NullInt64
	err := m.db.QueryRow("SELECT MAX(`batch`) FROM `migrations`").Scan(&maxBatch)
	if err != nil {
		return 1, err
	}
	if !maxBatch.Valid {
		return 1, nil
	}
	return int(maxBatch.Int64) + 1, nil
}

// Migrate runs all pending migrations in a new batch
func (m *Migrator) Migrate() error {
	if err := m.EnsureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	available, err := m.GetAvailableMigrations()
	if err != nil {
		return fmt.Errorf("failed to get available migrations: %w", err)
	}

	var pending []string
	for _, name := range available {
		if _, ok := applied[name]; !ok {
			pending = append(pending, name)
		}
	}

	if len(pending) == 0 {
		fmt.Println("Nothing to migrate.")
		return nil
	}

	batch, err := m.GetNextBatchNumber()
	if err != nil {
		return fmt.Errorf("failed to get next batch number: %w", err)
	}

	for _, name := range pending {
		fmt.Printf("Migrating: %s\n", name)
		start := time.Now()

		upFile := name + ".up.sql"
		content, err := fs.ReadFile(m.fsys, upFile)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", upFile, err)
		}

		if err := m.executeStatements(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", name, err)
		}

		_, err = m.db.Exec("INSERT INTO `migrations` (`migration`, `batch`) VALUES (?, ?)", name, batch)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}

		fmt.Printf("Migrated:  %s (%s)\n", name, time.Since(start).Round(time.Millisecond))
	}

	return nil
}

// Rollback rolls back migrations. If step <= 0, rolls back all migrations in the last batch; otherwise rolls back `step` migrations.
func (m *Migrator) Rollback(step int) error {
	if err := m.EnsureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	var query string
	var args []interface{}

	if step > 0 {
		query = "SELECT `migration` FROM `migrations` ORDER BY `id` DESC LIMIT ?"
		args = append(args, step)
	} else {
		var lastBatch sql.NullInt64
		err := m.db.QueryRow("SELECT MAX(`batch`) FROM `migrations`").Scan(&lastBatch)
		if err != nil || !lastBatch.Valid {
			fmt.Println("Nothing to rollback.")
			return nil
		}
		query = "SELECT `migration` FROM `migrations` WHERE `batch` = ? ORDER BY `id` DESC"
		args = append(args, lastBatch.Int64)
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return fmt.Errorf("failed to query migrations for rollback: %w", err)
	}
	defer rows.Close()

	var toRollback []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		toRollback = append(toRollback, name)
	}

	if len(toRollback) == 0 {
		fmt.Println("Nothing to rollback.")
		return nil
	}

	for _, name := range toRollback {
		fmt.Printf("Rolling back: %s\n", name)
		start := time.Now()

		downFile := name + ".down.sql"
		content, err := fs.ReadFile(m.fsys, downFile)
		if err != nil {
			return fmt.Errorf("failed to read rollback file %s: %w", downFile, err)
		}

		if err := m.executeStatements(string(content)); err != nil {
			return fmt.Errorf("failed to execute rollback %s: %w", name, err)
		}

		_, err = m.db.Exec("DELETE FROM `migrations` WHERE `migration` = ?", name)
		if err != nil {
			return fmt.Errorf("failed to delete migration record %s: %w", name, err)
		}

		fmt.Printf("Rolled back:  %s (%s)\n", name, time.Since(start).Round(time.Millisecond))
	}

	return nil
}

// Reset rolls back all migrations ever applied
func (m *Migrator) Reset() error {
	if err := m.EnsureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to ensure migrations table: %w", err)
	}

	rows, err := m.db.Query("SELECT `migration` FROM `migrations` ORDER BY `id` DESC")
	if err != nil {
		return fmt.Errorf("failed to query migrations for reset: %w", err)
	}
	defer rows.Close()

	var toRollback []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		toRollback = append(toRollback, name)
	}

	if len(toRollback) == 0 {
		fmt.Println("Nothing to rollback.")
		return nil
	}

	for _, name := range toRollback {
		fmt.Printf("Rolling back: %s\n", name)
		start := time.Now()

		downFile := name + ".down.sql"
		content, err := fs.ReadFile(m.fsys, downFile)
		if err != nil {
			return fmt.Errorf("failed to read rollback file %s: %w", downFile, err)
		}

		if err := m.executeStatements(string(content)); err != nil {
			return fmt.Errorf("failed to execute rollback %s: %w", name, err)
		}

		_, err = m.db.Exec("DELETE FROM `migrations` WHERE `migration` = ?", name)
		if err != nil {
			return fmt.Errorf("failed to delete migration record %s: %w", name, err)
		}

		fmt.Printf("Rolled back:  %s (%s)\n", name, time.Since(start).Round(time.Millisecond))
	}

	return nil
}

// Refresh rolls back the latest batch (or step) and runs migrate
func (m *Migrator) Refresh(step int) error {
	if err := m.Rollback(step); err != nil {
		return err
	}
	return m.Migrate()
}

// Fresh drops all tables from the database and runs all migrations
func (m *Migrator) Fresh(dbName string) error {
	fmt.Println("Dropping all tables...")

	if _, err := m.db.Exec("SET FOREIGN_KEY_CHECKS = 0;"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}
	defer m.db.Exec("SET FOREIGN_KEY_CHECKS = 1;")

	rows, err := m.db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = ? AND table_type = 'BASE TABLE'`, dbName)
	if err != nil {
		return fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return err
		}
		tables = append(tables, tableName)
	}

	for _, table := range tables {
		dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", table)
		if _, err := m.db.Exec(dropQuery); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	fmt.Println("Dropped all tables successfully.")
	return m.Migrate()
}

// Status returns the migration status list
func (m *Migrator) Status() ([]MigrationStatus, error) {
	applied, err := m.GetAppliedMigrations()
	if err != nil {
		return nil, err
	}

	available, err := m.GetAvailableMigrations()
	if err != nil {
		return nil, err
	}

	var result []MigrationStatus
	for _, name := range available {
		batchNum, ran := applied[name]
		var batchPtr *int
		if ran {
			b := batchNum
			batchPtr = &b
		}
		result = append(result, MigrationStatus{
			Ran:       ran,
			Migration: name,
			Batch:     batchPtr,
		})
	}

	return result, nil
}

func (m *Migrator) executeStatements(sqlContent string) error {
	statements := strings.Split(sqlContent, ";")
	for _, stmt := range statements {
		trimmed := strings.TrimSpace(stmt)
		if trimmed == "" {
			continue
		}
		if _, err := m.db.Exec(trimmed); err != nil {
			return fmt.Errorf("executing SQL statement failed: %w (SQL: %s)", err, trimmed)
		}
	}
	return nil
}

// MakeMigration scaffolds new .up.sql and .down.sql files
func MakeMigration(dir string, name string) (string, string, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create migrations dir: %w", err)
	}

	cleanName := strings.TrimSpace(strings.ToLower(name))
	cleanName = strings.ReplaceAll(cleanName, " ", "_")
	cleanName = strings.ReplaceAll(cleanName, "-", "_")

	timestamp := time.Now().Format("2006_01_02_150405")
	baseName := fmt.Sprintf("%s_%s", timestamp, cleanName)

	upFile := filepath.Join(dir, baseName+".up.sql")
	downFile := filepath.Join(dir, baseName+".down.sql")

	tableName := extractTableName(cleanName)

	var upContent, downContent string
	if tableName != "" {
		upContent = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS `+"`%s`"+` (
    `+"`id`"+` INT AUTO_INCREMENT PRIMARY KEY,
    `+"`created_at`"+` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `+"`updated_at`"+` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`, tableName)

		downContent = fmt.Sprintf("DROP TABLE IF EXISTS `%s`;\n", tableName)
	} else {
		upContent = "-- Migration up SQL statement\n"
		downContent = "-- Migration down SQL statement\n"
	}

	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write %s: %w", upFile, err)
	}

	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write %s: %w", downFile, err)
	}

	return upFile, downFile, nil
}

func extractTableName(name string) string {
	if strings.HasPrefix(name, "create_") && strings.HasSuffix(name, "_table") {
		trimmed := strings.TrimPrefix(name, "create_")
		return strings.TrimSuffix(trimmed, "_table")
	}
	return ""
}
```

- [ ] **Step 4: Run unit tests**

```bash
go test -v ./cmd/migrate/migrator/...
```
Expected: All tests pass.

- [ ] **Step 5: Commit**

```bash
git add cmd/migrate/migrator/
git commit -m "feat(migrations): implement Laravel-style migrator engine and status table formatter"
```

---

### Task 3: Refactor CLI Runner & Wire Artisan Commands

**Files:**
- Modify: `cmd/migrate/main.go`
- Modify: `cmd/migrate/main_test.go`

**Interfaces:**
- Consumes: `cmd/migrate/migrator`, `migrations.FS`
- Produces: CLI commands `migrate`, `migrate:rollback`, `migrate:status`, `migrate:reset`, `migrate:refresh`, `migrate:fresh`, `make:migration`, and shorthand aliases.

- [ ] **Step 1: Update `cmd/migrate/main.go`**

Refactor `cmd/migrate/main.go` to parse commands and dispatch to `migrator`:

```go
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
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
```

- [ ] **Step 2: Update `cmd/migrate/main_test.go`**

Rewrite integration tests in `cmd/migrate/main_test.go` to test Laravel migration lifecycle (`Fresh`, `Migrate`, `Status`, `Rollback`, `Reset`):

```go
package main

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"go-mvc/cmd/migrate/migrator"
	"go-mvc/migrations"
)

func TestLaravelMigrationLifecycle(t *testing.T) {
	cfg := loadConfig()
	testDBName := cfg.Name + "_test"
	cfg.Name = testDBName

	if err := ensureDatabaseExists(cfg); err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, testDBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	defer func() {
		_ = db.Close()
		serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
		if rootDB, err := sql.Open("mysql", serverDSN); err == nil {
			_, _ = rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`;", testDBName))
			_ = rootDB.Close()
		}
	}()

	mgr := migrator.New(db, migrations.FS)

	// 1. Initial Fresh / Migrate
	if err := mgr.Fresh(testDBName); err != nil {
		t.Fatalf("Fresh migration failed: %v", err)
	}

	// 2. Verify Status
	statuses, err := mgr.Status()
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if len(statuses) < 2 {
		t.Fatalf("Expected at least 2 migrations, got %d", len(statuses))
	}
	for _, s := range statuses {
		if !s.Ran {
			t.Errorf("Expected migration %s to have ran", s.Migration)
		}
		if s.Batch == nil || *s.Batch != 1 {
			t.Errorf("Expected migration %s to be in batch 1, got %v", s.Migration, s.Batch)
		}
	}

	// 3. Rollback last batch (batch 1)
	if err := mgr.Rollback(0); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	statusesAfterRollback, err := mgr.Status()
	if err != nil {
		t.Fatalf("Status after rollback failed: %v", err)
	}
	for _, s := range statusesAfterRollback {
		if s.Ran {
			t.Errorf("Expected migration %s to be rolled back", s.Migration)
		}
	}

	// 4. Re-migrate
	if err := mgr.Migrate(); err != nil {
		t.Fatalf("Re-migrate failed: %v", err)
	}
}
```

- [ ] **Step 3: Run integration test**

```bash
go test -v ./cmd/migrate/...
```
Expected: `TestLaravelMigrationLifecycle` passes cleanly.

- [ ] **Step 4: Commit**

```bash
git add cmd/migrate/main.go cmd/migrate/main_test.go
git commit -m "feat(migrations): wire Laravel-style Artisan commands into migration CLI runner"
```

---

### Task 4: End-to-End Verification & Documentation Update

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: `go run cmd/migrate/main.go`, `README.md`
- Produces: Verified database state, updated documentation.

- [ ] **Step 1: Execute `migrate:fresh` and verify status**

```bash
go run cmd/migrate/main.go migrate:fresh
go run cmd/migrate/main.go migrate:status
```
Expected: Both employees and users tables created, status table displays `Yes` with `Batch: 1`.

- [ ] **Step 2: Test `make:migration` command**

```bash
go run cmd/migrate/main.go make:migration create_test_dummies_table
```
Verify generated file, then delete dummy files.

- [ ] **Step 3: Run full repository test suite**

```bash
go test -count=1 ./...
```
Expected: All tests pass.

- [ ] **Step 4: Update `README.md`**

Replace old `golang-migrate` commands (`up`, `down`, `version`, `force`) in `README.md` with Laravel-style commands:
- `migrate` (alias `up`)
- `migrate:rollback` (alias `rollback`, `down`)
- `migrate:status` (alias `status`)
- `migrate:reset` (alias `reset`)
- `migrate:refresh` (alias `refresh`)
- `migrate:fresh` (alias `fresh`)
- `make:migration <name>`

- [ ] **Step 5: Commit documentation update**

```bash
git add README.md
git commit -m "docs: document Laravel-style migration commands in README"
```
