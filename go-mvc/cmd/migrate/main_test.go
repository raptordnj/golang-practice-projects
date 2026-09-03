package main

import (
	"database/sql"
	"fmt"
	"testing"

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
