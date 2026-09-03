# Laravel-Style Database Migrations System Design for WorkPulse (go-mvc)

## 1. Overview
This specification details the design for replacing the linear integer-version migration CLI with a full **Laravel Artisan-style database migration system** in the WorkPulse (`go-mvc`) application.

It introduces:
- Timestamped migration filenames (`YYYY_MM_DD_HHMMSS_<name>.up.sql` and `.down.sql`).
- A Laravel-style `migrations` tracking table recording migration name and **batch numbers**.
- Complete Artisan CLI commands: `migrate`, `migrate:rollback`, `migrate:status`, `migrate:reset`, `migrate:refresh`, `migrate:fresh`, and `make:migration <name>`.
- Formatted console output matching Laravel's Artisan command line experience.

## 2. Goals
- **Batch Tracking**: Group migrations applied together in the same execution run into a shared `batch` integer, allowing atomic rollbacks of the last applied batch.
- **Timestamp Versioning**: Standardize on timestamp prefixes (`YYYY_MM_DD_HHMMSS_...`) to eliminate merge conflict collisions inherent to sequential numbers.
- **Artisan Command Suite**:
  - `migrate`: Runs all outstanding migrations in a new batch.
  - `migrate:rollback [--step=N]`: Rolls back the last batch of migrations (or `N` latest migrations).
  - `migrate:status`: Displays an aligned tabular view of all migrations, ran state, and batch number.
  - `migrate:reset`: Rolls back all migrations ever applied.
  - `migrate:refresh`: Resets all migrations and then re-runs `migrate`.
  - `migrate:fresh`: Drops all database tables and runs all migrations from scratch.
  - `make:migration <name>`: Automatically generates timestamped `.up.sql` and `.down.sql` template files.
- **Embedded & Standalone**: Migrations continue to be embedded via Go `embed.FS` in `migrations/fs.go` for zero-runtime external file dependencies, with files also accessible on disk for `make:migration`.
- **Zero Global Dependencies**: Operates natively via `go run cmd/migrate/main.go <command>` without requiring PHP or third-party global binaries.

## 3. Database Schema

The tracking table replaces `schema_migrations` with Laravel's standard `migrations` table:

```sql
CREATE TABLE IF NOT EXISTS `migrations` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `migration` VARCHAR(255) NOT NULL UNIQUE,
    `batch` INT NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### Existing Migrations Transition
Existing migrations will be renamed to follow timestamp conventions:
- `migrations/000001_create_employees_table.*` -> `migrations/2026_09_03_000001_create_employees_table.*.sql`
- `migrations/000002_create_users_table.*` -> `migrations/2026_09_03_000002_create_users_table.*.sql`

The migration name stored in the `migration` column will be the base name without extensions:
e.g. `2026_09_03_000001_create_employees_table`.

## 4. CLI Architecture & Commands

### 4.1. Directory Structure
```
go-mvc/
├── cmd/
│   └── migrate/
│       ├── main.go                  # CLI entry point, argument router
│       ├── main_test.go             # End-to-end integration tests
│       └── migrator/
│           ├── migrator.go          # Core Laravel migration engine
│           ├── migrator_test.go     # Unit tests for migrator
│           └── output.go            # Tabular status & console formatting
├── migrations/
│   ├── fs.go                        # go:embed *.sql
│   ├── 2026_09_03_000001_create_employees_table.up.sql
│   ├── 2026_09_03_000001_create_employees_table.down.sql
│   ├── 2026_09_03_000002_create_users_table.up.sql
│   └── 2026_09_03_000002_create_users_table.down.sql
```

### 4.2. Command Specifications

#### 1. `migrate` (alias: `up`)
- Ensures `migrations` table exists.
- Finds all available migration names from `migrations.FS`.
- Queries `SELECT migration FROM migrations` to determine already ran migrations.
- Identifies pending migrations sorted alphanumerically.
- If no pending migrations: prints `Nothing to migrate.`
- Computes `currentBatch = MAX(batch) + 1` (or 1 if no records).
- For each pending migration:
  - Prints: `Migrating: <migration_name>`
  - Executes `.up.sql` SQL statements.
  - Inserts row into `migrations` (`migration`, `batch`).
  - Prints: `Migrated:  <migration_name> (<duration>ms)`

#### 2. `migrate:rollback [--step=N]` (alias: `rollback`, `down`)
- Determines migrations to rollback:
  - If `--step=N` is specified: finds the `N` most recent migrations by `id DESC`.
  - If `--step` is not specified: finds all migrations matching `batch = MAX(batch)` ordered by `id DESC`.
- If no migrations found: prints `Nothing to rollback.`
- For each migration:
  - Prints: `Rolling back: <migration_name>`
  - Executes corresponding `.down.sql` SQL statements.
  - Deletes row from `migrations` table.
  - Prints: `Rolled back:  <migration_name> (<duration>ms)`

#### 3. `migrate:status` (alias: `status`)
- Scans all migration files in `migrations/`.
- Queries `SELECT migration, batch FROM migrations`.
- Prints formatted tabular status:
  ```
  +------+-----------------------------------------------+-------+
  | Ran? | Migration                                     | Batch |
  +------+-----------------------------------------------+-------+
  | Yes  | 2026_09_03_000001_create_employees_table      | 1     |
  | Yes  | 2026_09_03_000002_create_users_table          | 1     |
  +------+-----------------------------------------------+-------+
  ```

#### 4. `migrate:reset` (alias: `reset`)
- Rolls back all applied migrations in reverse order (`ORDER BY id DESC`).
- Reports `Rolled back: <migration_name>` for each.

#### 5. `migrate:refresh [--step=N]` (alias: `refresh`)
- Rolls back migrations (last batch or `--step=N`), then immediately runs `migrate`.

#### 6. `migrate:fresh` (alias: `fresh`)
- Drops all tables in the active database.
  - Temporarily sets `FOREIGN_KEY_CHECKS = 0`.
  - Queries `SELECT table_name FROM information_schema.tables WHERE table_schema = ?`.
  - Drops each table with `DROP TABLE IF EXISTS <table_name>`.
  - Re-enables `FOREIGN_KEY_CHECKS = 1`.
- Prints: `Dropped all tables successfully.`
- Creates `migrations` table and runs `migrate`.

#### 7. `make:migration <name>` (alias: `make <name>`)
- Formats timestamp `time.Now().Format("2006_01_02_150405")`.
- Generates file path `migrations/<timestamp>_<name>.up.sql` and `.down.sql`.
- If name begins with `create_<table_name>_table`:
  - Generates sensible starter schema in `.up.sql`:
    ```sql
    CREATE TABLE IF NOT EXISTS `<table_name>` (
        `id` INT AUTO_INCREMENT PRIMARY KEY,
        `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
    ```
  - Generates `.down.sql`:
    ```sql
    DROP TABLE IF EXISTS `<table_name>`;
    ```
- Prints: `Created Migration: <timestamp>_<name>`

## 5. Error Handling & Edge Cases
- **Multi-Statement Migrations**: Connect with `multiStatements=true` to allow multi-line and composite DDL commands.
- **Transaction Safety**: For DDL that MySQL auto-commits, execution checks statements sequentially and halts on error, reporting the exact failed statement and migration name.
- **Missing Down File**: If a `.down.sql` file is missing during rollback, aborts with a clear error indicating the missing file.
- **Existing `schema_migrations` Table**: If an old `golang-migrate` table exists, `migrate:fresh` cleans it up; migration runner ignores `schema_migrations` and manages `migrations`.

## 6. Verification & Testing
1. **Unit & Integration Tests (`migrator_test.go`, `main_test.go`)**:
   - Verify `migrate` applies migrations and sets batch = 1.
   - Verify subsequent `migrate` applies any new migration with batch = 2.
   - Verify `migrate:rollback` rolls back only batch 2.
   - Verify `migrate:rollback --step=1` rolls back only 1 migration.
   - Verify `migrate:status` reports correct ran status and batch numbers.
   - Verify `migrate:fresh` clears all tables and re-migrates.
   - Verify `make:migration` creates `.up.sql` and `.down.sql` files with correct timestamps.
2. **End-to-End Verification**:
   - Run `go run cmd/migrate/main.go fresh`.
   - Run `go run cmd/migrate/main.go status`.
   - Verify employee and user CRUD APIs work against the fresh database schema.
   - Update `README.md` with Laravel-style migration commands.
