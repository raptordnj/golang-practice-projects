package migrator

import (
	"context"
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
	if err := rows.Err(); err != nil {
		return nil, err
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
		if err != nil {
			return fmt.Errorf("failed to query max batch: %w", err)
		}
		if !lastBatch.Valid {
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
	if err := rows.Err(); err != nil {
		return err
	}

	return m.rollbackList(toRollback)
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
	if err := rows.Err(); err != nil {
		return err
	}

	return m.rollbackList(toRollback)
}

func (m *Migrator) rollbackList(toRollback []string) error {
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

// Refresh rolls back all migrations (or step migrations if step > 0) and re-runs migrate.
func (m *Migrator) Refresh(step int) error {
	var err error
	if step > 0 {
		err = m.Rollback(step)
	} else {
		err = m.Reset()
	}
	if err != nil {
		return err
	}
	return m.Migrate()
}

// Fresh drops all tables from the database and runs all migrations
func (m *Migrator) Fresh(dbName string) error {
	fmt.Println("Dropping all tables...")

	if err := m.dropAllTables(dbName); err != nil {
		return err
	}

	fmt.Println("Dropped all tables successfully.")
	return m.Migrate()
}

func (m *Migrator) dropAllTables(dbName string) error {
	ctx := context.Background()
	conn, err := m.db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get dedicated database connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0;"); err != nil {
		return fmt.Errorf("failed to disable foreign key checks: %w", err)
	}
	defer conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1;")

	rows, err := conn.QueryContext(ctx, `
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
	if err := rows.Err(); err != nil {
		return err
	}

	for _, table := range tables {
		escapedTable := strings.ReplaceAll(table, "`", "``")
		dropQuery := fmt.Sprintf("DROP TABLE IF EXISTS `%s`;", escapedTable)
		if _, err := conn.ExecContext(ctx, dropQuery); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	return nil
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
	trimmed := strings.TrimSpace(sqlContent)
	if trimmed == "" {
		return nil
	}
	if _, err := m.db.Exec(trimmed); err != nil {
		return fmt.Errorf("executing SQL statement failed: %w (SQL: %s)", err, trimmed)
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
