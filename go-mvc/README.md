# WorkPulse — Fullstack Employee CRUD Application

A modern, full-stack Employee Management system featuring a **Next.js (React 19 + Tailwind CSS v4 + shadcn/ui)** frontend paired with a high-performance **Beego v2 (Go)** REST API, **MySQL**, and **JWT Authentication**.

## Architecture

```
go-mvc/
├── cmd/
│   └── migrate/
│       └── main.go             # Database migration CLI runner
├── conf/
│   └── app.conf                # Application, DB & JWT configuration
├── controllers/
│   ├── auth_controller.go      # Authentication handlers (register, login, me)
│   └── employee_controller.go  # Employee request handlers
├── dto/
│   ├── auth_dto.go             # Auth request/response DTOs
│   └── employee_dto.go         # Employee request/response DTOs
├── middlewares/
│   └── jwt_auth.go             # JWT Bearer token authentication filter
├── migrations/                 # Embedded SQL migration files
│   ├── 000001_create_employees_table.up.sql
│   ├── 000001_create_employees_table.down.sql
│   ├── 000002_create_users_table.up.sql
│   └── 000002_create_users_table.down.sql
├── models/
│   ├── employee.go             # Employee database entity (ORM model)
│   └── user.go                 # User database entity (ORM model)
├── repositories/
│   ├── employee_repository.go  # Employee data access layer (interface + impl)
│   └── user_repository.go      # User data access layer (interface + impl)
├── services/
│   ├── auth_service.go         # Auth business logic (bcrypt + JWT tokens)
│   └── employee_service.go     # Employee business logic layer
├── routers/
│   └── router.go               # Route definitions, CORS, JWT filter & DI wiring
├── frontend/                   # Next.js 16 + React 19 + Tailwind v4 UI
│   ├── src/
│   │   ├── app/                # App Router (page.tsx, login, register, layout.tsx)
│   │   ├── components/auth/    # Auth UI (AuthGuard)
│   │   ├── components/employee/# Employee UI (Cards, Table, Stats, Modals)
│   │   ├── components/ui/      # shadcn/ui primitives (Button, Input, Badge, Dialog, Toast)
│   │   ├── context/            # AuthContext (token storage, login/logout, state)
│   │   ├── lib/api.ts          # Token-aware API client with Next.js rewrites
│   │   └── types/              # TypeScript types (Employee, User, Auth)
│   ├── package.json
│   └── next.config.ts
├── main.go                     # Backend entry point (Auto DB creation + SyncDB)
├── go.mod
└── go.sum
```

## Features

- **Authentication & Authorization**: User registration and login using bcrypt password hashing and HMAC-SHA256 JWT tokens with 24-hour expiration.
- **Route Protection**: Backend JWT filter middleware safeguarding all `/api/v1/employees/*` and `/api/v1/auth/me` endpoints. Unauthenticated requests receive `401 Unauthorized`.
- **Frontend Auth Guard & State**: React context managing token lifecycle in `localStorage`, automatic redirection of unauthenticated users to `/login`, and user header with logout functionality.
- **Full Employee CRUD**: Create, read, update, and delete employees with pagination, search, department filtering, and responsive UI cards/table views.
- **Database Schema Migrations**: Embedded SQL migrations managed via `golang-migrate/migrate/v4`.

## Running the Application

### 1. Start the Beego Go Backend
```bash
go run .
```
Backend runs on **http://localhost:8080** (automatically ensures the `employee_db` database exists and syncs models).

### 2. Start the Next.js Frontend
In a separate terminal:
```bash
cd frontend
npm run dev
```
Frontend runs on **http://localhost:3000**. All API calls are transparently proxied to the Go backend via Next.js rewrites.
- Visit `http://localhost:3000` to access the employee dashboard (unauthenticated requests automatically redirect to `/login`).
- Visit `http://localhost:3000/register` to create a new user account.
- Visit `http://localhost:3000/login` to log into an existing user account.

## Setup

### 1. Create the MySQL Database

```sql
CREATE DATABASE IF NOT EXISTS employee_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. Configure Application & Database Connection

Edit `conf/app.conf` if your MySQL credentials or JWT secret differ:

```ini
appname = go-mvc
httpport = 8080
runmode = dev
copyrequestbody = true
jwt_secret = workpulse_super_secret_jwt_key_2026

[db]
host = 127.0.0.1
port = 3306
user = root
password = root
name = employee_db
```

Alternatively, environment variables can be used (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `JWT_SECRET`).

### 3. Run Database Migrations

Apply schema migrations using the CLI runner:

```bash
go run cmd/migrate/main.go up
```

To verify the migration version:
```bash
go run cmd/migrate/main.go version
```
Expected output: `Current version: 2 (dirty: false)`.

### 4. Install Dependencies & Run

```bash
go mod tidy
go run main.go
```

The server starts on **http://localhost:8080**.

## Database Migrations

WorkPulse uses [`golang-migrate`](https://github.com/golang-migrate/migrate) for version-controlled schema migrations with embedded SQL scripts. A standalone CLI runner is provided in `cmd/migrate/main.go`.

### CLI Commands

| Command | Description |
|---|---|
| `go run cmd/migrate/main.go up` | Apply all pending migrations |
| `go run cmd/migrate/main.go up <n>` | Apply next `n` migrations (e.g. `up 1`) |
| `go run cmd/migrate/main.go down` | Roll back all applied migrations |
| `go run cmd/migrate/main.go down <n>` | Roll back `n` migrations (e.g. `down 1`) |
| `go run cmd/migrate/main.go version` | Print current schema migration version and dirty status |
| `go run cmd/migrate/main.go force <version>` | Force set migration version (used to recover from a dirty state) |

### Migration History

1. `000001_create_employees_table`: Creates the `employees` table.
2. `000002_create_users_table`: Creates the `users` table with unique `email` index and hashed passwords.

## API Endpoints

### Authentication Endpoints

| Method | Endpoint | Auth Required | Description |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | No | Register a new user account and obtain JWT token |
| `POST` | `/api/v1/auth/login` | No | Log in with email & password and obtain JWT token |
| `GET` | `/api/v1/auth/me` | **Yes** (Bearer Token) | Retrieve authenticated user profile |

### Employee Endpoints (Protected)

> [!IMPORTANT]
> All employee endpoints require authentication. Requests must supply a valid JWT token via the `Authorization: Bearer <token>` header. Unauthenticated requests receive HTTP `401 Unauthorized`.

| Method   | Endpoint                  | Auth Required          | Description          |
|----------|---------------------------|------------------------|----------------------|
| `POST`   | `/api/v1/employees`       | **Yes** (Bearer Token) | Create employee      |
| `GET`    | `/api/v1/employees`       | **Yes** (Bearer Token) | List all (paginated) |
| `GET`    | `/api/v1/employees/:id`   | **Yes** (Bearer Token) | Get by ID            |
| `PUT`    | `/api/v1/employees/:id`   | **Yes** (Bearer Token) | Update employee      |
| `DELETE` | `/api/v1/employees/:id`   | **Yes** (Bearer Token) | Delete employee      |

## Usage Examples

### 1. Register User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Doe",
    "email": "jane@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "token": "<JWT_TOKEN>",
    "user": {
      "id": 1,
      "name": "Jane Doe",
      "email": "jane@example.com",
      "created_at": "2026-09-03T13:41:37Z"
    }
  }
}
```

### 2. Login User

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "jane@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "<JWT_TOKEN>",
    "user": {
      "id": 1,
      "name": "Jane Doe",
      "email": "jane@example.com",
      "created_at": "2026-09-03T13:41:37Z"
    }
  }
}
```

### 3. Get Current User Profile

```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

Response:
```json
{
  "success": true,
  "message": "",
  "data": {
    "id": 1,
    "name": "Jane Doe",
    "email": "jane@example.com",
    "created_at": "2026-09-03T13:41:37Z"
  }
}
```

### 4. Create Employee (Authenticated)

```bash
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "department": "Engineering",
    "position": "Software Engineer",
    "salary": 85000.00,
    "hire_date": "2024-01-15"
  }'
```

### 5. List Employees (Authenticated & Paginated)

```bash
curl -X GET "http://localhost:8080/api/v1/employees?page=1&page_size=10" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 6. Get Employee by ID (Authenticated)

```bash
curl -X GET http://localhost:8080/api/v1/employees/1 \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### 7. Update Employee (Authenticated)

```bash
curl -X PUT http://localhost:8080/api/v1/employees/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -d '{
    "position": "Senior Software Engineer",
    "salary": 105000.00
  }'
```

### 8. Delete Employee (Authenticated)

```bash
curl -X DELETE http://localhost:8080/api/v1/employees/1 \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

## Response Format

### Success Response

```json
{
  "success": true,
  "message": "Employee created successfully",
  "data": {
    "id": 1,
    "first_name": "John",
    "last_name": "Doe",
    "email": "john.doe@example.com",
    "phone": "+1234567890",
    "department": "Engineering",
    "position": "Software Engineer",
    "salary": 85000,
    "hire_date": "2024-01-15",
    "is_active": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

### Paginated Response

```json
{
  "success": true,
  "message": "Employees retrieved successfully",
  "data": [...],
  "page": 1,
  "page_size": 10,
  "total_count": 25,
  "total_pages": 3
}
```

### Error Response

```json
{
  "success": false,
  "message": "employee not found"
}
```

### Unauthorized Error Response (HTTP 401)

```json
{
  "success": false,
  "message": "Authorization token is missing or malformed"
}
```
