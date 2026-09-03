package main

import (
	"database/sql"
	"fmt"
	"testing"
	"testing/fstest"

	"go-mvc/cmd/migrate/migrator"
	"go-mvc/migrations"
)

func TestLoadConfig(t *testing.T) {
	cfg := loadConfig()
	if cfg.Host == "" {
		t.Errorf("expected non-empty Host, got empty")
	}
	if cfg.Port == "" {
		t.Errorf("expected non-empty Port, got empty")
	}
	if cfg.Name == "" {
		t.Errorf("expected non-empty Name, got empty")
	}
}

func TestLaravelMigrationLifecycle(t *testing.T) {
	cfg := loadConfig()
	testDBName := cfg.Name + "_test"
	cfg.Name = testDBName

	if err := ensureDatabaseExists(cfg); err != nil {
		t.Skipf("Skipping database integration test (MySQL server unavailable): %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, testDBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Skipping database integration test (MySQL server unavailable): %v", err)
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

	// 5. Refresh (reset and re-run)
	if err := mgr.Refresh(0); err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// 6. Reset (rollback all)
	if err := mgr.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}
}

func TestExecuteStatementsRobustness(t *testing.T) {
	cfg := loadConfig()
	testDBName := cfg.Name + "_test_robust"
	cfg.Name = testDBName

	if err := ensureDatabaseExists(cfg); err != nil {
		t.Skipf("Skipping database integration test (MySQL server unavailable): %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, testDBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skipf("Skipping database integration test (MySQL server unavailable): %v", err)
	}
	defer func() {
		_ = db.Close()
		serverDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/", cfg.User, cfg.Password, cfg.Host, cfg.Port)
		if rootDB, err := sql.Open("mysql", serverDSN); err == nil {
			_, _ = rootDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`;", testDBName))
			_ = rootDB.Close()
		}
	}()

	mockFS := fstest.MapFS{
		"2026_09_03_000001_test_robustness.up.sql": &fstest.MapFile{
			Data: []byte(`
				-- Migration comment with semicolon; and details;
				CREATE TABLE IF NOT EXISTS robust_items (
					id INT AUTO_INCREMENT PRIMARY KEY,
					notes VARCHAR(255) NOT NULL DEFAULT 'hello; world;'
				);
				-- Second statement with semicolon in string and comment;
				INSERT INTO robust_items (notes) VALUES ('semicolon; in; string;');
			`),
		},
		"2026_09_03_000001_test_robustness.down.sql": &fstest.MapFile{
			Data: []byte(`
				-- Down migration with semicolon; in comment;
				DROP TABLE IF EXISTS robust_items;
			`),
		},
	}

	mgr := migrator.New(db, mockFS)
	if err := mgr.Fresh(testDBName); err != nil {
		t.Fatalf("Fresh failed: %v", err)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM robust_items WHERE notes = 'semicolon; in; string;'").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count = 1, got %d", count)
	}

	if err := mgr.Rollback(0); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
}
