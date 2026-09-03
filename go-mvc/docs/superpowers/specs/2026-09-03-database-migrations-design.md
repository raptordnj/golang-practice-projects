# Database Migrations System Design for WorkPulse (go-mvc)

## 1. Overview
This specification details the database migration system for the `go-mvc` WorkPulse application. It introduces version-controlled SQL migrations using `golang-migrate/migrate/v4` and a standalone Go CLI runner at `cmd/migrate/main.go`.

## 2. Goals
- Provide explicit, reproducible, and reversible (`up` / `down`) schema migrations for MySQL.
- Provide a dedicated CLI tool (`cmd/migrate/main.go`) that executes migrations without needing external global binary installations.
- Automatically parse database connection settings from `conf/app.conf` or environment variable overrides.
- Ensure automated database creation if the database does not already exist.
- Update project documentation (`README.md`) to guide users on running migrations.

## 3. Directory Structure
```
go-mvc/
├── cmd/
│   └── migrate/
│       └── main.go                               # CLI tool to run migrations
├── migrations/
│   ├── 000001_create_employees_table.up.sql      # Forward migration for employees table
│   └── 000001_create_employees_table.down.sql    # Backward migration for employees table
├── conf/
│   └── app.conf                                  # Database configuration
├── models/
│   └── employee.go                               # Beego ORM model
└── README.md                                     # Project docs with migration instructions
```

## 4. Migration Files

### 4.1. `migrations/000001_create_employees_table.up.sql`
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

### 4.2. `migrations/000001_create_employees_table.down.sql`
```sql
DROP TABLE IF EXISTS `employees`;
```

## 5. Migration CLI Implementation (`cmd/migrate/main.go`)

### 5.1. Configuration Resolution
The CLI will resolve connection parameters in order of priority:
1. Environment variables (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`).
2. Values from `conf/app.conf` under the `[db]` section or default keys.
3. Fallback defaults:
   - Host: `127.0.0.1`
   - Port: `3306`
   - User: `root`
   - Password: `""` (or as configured)
   - Database: `employee_db`

### 5.2. Database Bootstrap
Before running migrations, `cmd/migrate/main.go` will connect to MySQL server DSN (`root:pass@tcp(host:port)/`) and run:
```sql
CREATE DATABASE IF NOT EXISTS `employee_db` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 5.3. Migration Execution Engine
Using `github.com/golang-migrate/migrate/v4`:
- Drivers: `github.com/golang-migrate/migrate/v4/database/mysql` and `github.com/golang-migrate/migrate/v4/source/file` (or `iofs`/`embed`).
- Commands supported:
  - `up`: Applies all pending migrations or `up <n>` steps.
  - `down`: Reverts all migrations or `down <n>` steps.
  - `version`: Outputs current active migration version and dirty flag.
  - `force <v>`: Forces database migration version state (clearing dirty state).

## 6. Verification Plan
1. Add dependencies to `go.mod` via `go get github.com/golang-migrate/migrate/v4`.
2. Run `go run cmd/migrate/main.go up` and verify migration applies cleanly.
3. Check version via `go run cmd/migrate/main.go version`.
4. Verify table creation and structure in MySQL.
5. Verify `go test ./...` and that `main.go` starts and interacts properly with the migrated database.
