# WorkPulse — Fullstack Employee CRUD Application

A modern, full-stack Employee Management system featuring a **Next.js (React 19 + Tailwind CSS v4 + shadcn/ui)** frontend paired with a high-performance **Beego v2 (Go)** REST API and **MySQL**.

## Architecture

```
go-mvc/
├── conf/
│   └── app.conf               # Application & DB configuration
├── controllers/
│   └── employee_controller.go  # HTTP request handlers
├── dto/
│   └── employee_dto.go         # Request/Response DTOs
├── models/
│   └── employee.go             # Database entity (ORM model)
├── repositories/
│   └── employee_repository.go  # Data access layer (interface + impl)
├── services/
│   └── employee_service.go     # Business logic layer
├── routers/
│   └── router.go               # Route definitions, CORS & DI wiring
├── frontend/                   # Next.js 16 + React 19 + Tailwind v4 UI
│   ├── src/
│   │   ├── app/                # App Router (page.tsx, layout.tsx, globals.css)
│   │   ├── components/ui/      # shadcn/ui primitives (Button, Input, Badge, Dialog, Toast)
│   │   ├── components/employee/# Employee UI (Cards, Table, Stats, Modals)
│   │   ├── lib/api.ts          # API client with Next.js rewrites
│   │   └── types/              # TypeScript types
│   ├── package.json
│   └── next.config.ts
├── main.go                     # Backend entry point (Auto DB creation + SyncDB)
├── go.mod
└── go.sum
```

## Running the Application

### 1. Start the Beego Go Backend
```bash
go run .
```
Backend runs on **http://localhost:8080** (automatically creates the `employee_db` database and `employees` table).

### 2. Start the Next.js Frontend
In a separate terminal:
```bash
cd frontend
npm run dev
```
Frontend runs on **http://localhost:3000**. All API calls are transparently proxied to the Go backend.

## Setup

### 1. Create the MySQL Database

```sql
CREATE DATABASE IF NOT EXISTS employee_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 2. Configure Database Connection

Edit `conf/app.conf` if your MySQL credentials differ:

```ini
[db]
host = 127.0.0.1
port = 3306
user = root
password =
name = employee_db
```

### 3. Install Dependencies & Run

```bash
go mod tidy
go run main.go
```

The server starts on **http://localhost:8080**. Tables are auto-created via `orm.RunSyncdb`.

## API Endpoints

| Method   | Endpoint                  | Description          |
|----------|---------------------------|----------------------|
| `POST`   | `/api/v1/employees`       | Create employee      |
| `GET`    | `/api/v1/employees`       | List all (paginated) |
| `GET`    | `/api/v1/employees/:id`   | Get by ID            |
| `PUT`    | `/api/v1/employees/:id`   | Update employee      |
| `DELETE` | `/api/v1/employees/:id`   | Delete employee      |

## Usage Examples

### Create Employee

```bash
curl -X POST http://localhost:8080/api/v1/employees \
  -H "Content-Type: application/json" \
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

### List Employees (Paginated)

```bash
curl "http://localhost:8080/api/v1/employees?page=1&page_size=10"
```

### Get Employee by ID

```bash
curl http://localhost:8080/api/v1/employees/1
```

### Update Employee

```bash
curl -X PUT http://localhost:8080/api/v1/employees/1 \
  -H "Content-Type: application/json" \
  -d '{
    "position": "Senior Software Engineer",
    "salary": 105000.00
  }'
```

### Delete Employee

```bash
curl -X DELETE http://localhost:8080/api/v1/employees/1
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
