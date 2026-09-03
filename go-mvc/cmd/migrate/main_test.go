package main

import (
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
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

func TestMigrationLifecycle(t *testing.T) {
	cfg := loadConfig()
	m, db, err := getMigrator(cfg)
	if err != nil {
		t.Skipf("Skipping migration integration test (DB not available): %v", err)
		return
	}
	defer db.Close()

	// 1. Apply all migrations
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("m.Up() failed: %v", err)
	}

	// 2. Check version is 1, not dirty
	version, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("m.Version() failed: %v", err)
	}
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
	if dirty {
		t.Errorf("expected dirty=false, got true")
	}

	// 3. Rollback 1 step
	err = m.Steps(-1)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("m.Steps(-1) failed: %v", err)
	}

	// 4. Verify version is now nil
	_, _, err = m.Version()
	if !errors.Is(err, migrate.ErrNilVersion) {
		t.Fatalf("expected ErrNilVersion after rollback, got %v", err)
	}

	// 5. Re-apply migration
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("re-applying m.Up() failed: %v", err)
	}

	// 6. Verify version is 1 again
	version, dirty, err = m.Version()
	if err != nil {
		t.Fatalf("m.Version() after re-up failed: %v", err)
	}
	if version != 1 || dirty {
		t.Errorf("expected version 1, dirty false, got version %d, dirty %v", version, dirty)
	}
}
