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

	// 2. Check version is >= 1, not dirty
	version, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("m.Version() failed: %v", err)
	}
	if version < 1 {
		t.Errorf("expected version >= 1, got %d", version)
	}
	if dirty {
		t.Errorf("expected dirty=false, got true")
	}

	// 3. Rollback 1 step
	err = m.Steps(-1)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("m.Steps(-1) failed: %v", err)
	}

	// 4. Verify version after rollback of 1 step
	prevVersion, prevDirty, err := m.Version()
	if version == 1 {
		if !errors.Is(err, migrate.ErrNilVersion) {
			t.Fatalf("expected ErrNilVersion after rollback, got %v", err)
		}
	} else {
		if err != nil {
			t.Fatalf("m.Version() after rollback failed: %v", err)
		}
		if prevVersion != version-1 || prevDirty {
			t.Errorf("expected version %d, dirty false, got version %d, dirty %v", version-1, prevVersion, prevDirty)
		}
	}

	// 5. Re-apply migration
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("re-applying m.Up() failed: %v", err)
	}

	// 6. Verify version is restored
	restoredVersion, restoredDirty, err := m.Version()
	if err != nil {
		t.Fatalf("m.Version() after re-up failed: %v", err)
	}
	if restoredVersion != version || restoredDirty {
		t.Errorf("expected version %d, dirty false, got version %d, dirty %v", version, restoredVersion, restoredDirty)
	}
}
