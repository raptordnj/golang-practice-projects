# Fullstack Authentication MVP Design for WorkPulse (go-mvc)

## 1. Overview
This specification details the design for adding an Authentication MVP to the WorkPulse fullstack application (`go-mvc`). It establishes email/password authentication using standard JWT (JSON Web Tokens) with bcrypt password hashing in the Go Beego v2 REST API backend, protects all employee endpoints, and provides seamless user authentication state, dedicated login/register flows, and route protection in the Next.js frontend.

## 2. Goals & Scope
- **User Management**: Support user registration and login with name, email, and password.
- **Security**: Hash passwords securely using bcrypt (`golang.org/x/crypto/bcrypt`). Use JWT tokens signed with a secret (`golang-jwt/jwt/v5`) with expiration.
- **Database Migrations**: Add version-controlled migration `000002_create_users_table.up.sql` and `000002_create_users_table.down.sql` using the existing migration CLI (`cmd/migrate`).
- **Backend Architecture**: Follow the established clean architecture (`models` -> `repositories` -> `services` -> `controllers` -> `routers`).
- **Endpoint Protection**: Protect `/api/v1/employees/*` and `/api/v1/auth/me` using a Beego HTTP filter middleware that validates the `Authorization: Bearer <token>` header.
- **Frontend Experience**:
  - Provide an `AuthContext` / `useAuth` hook persisting authentication state and token in `localStorage`.
  - Modern Next.js `/login` and `/register` pages styled with Tailwind CSS and shadcn/ui components.
  - Automatically inject the Bearer token into all employee API calls via `frontend/src/lib/api.ts`.
  - Display user information and a logout button in the header. Redirect unauthenticated visitors to `/login`.

## 3. Architecture & Data Model

### 3.1. Database Migration
File: `migrations/000002_create_users_table.up.sql`
```sql
CREATE TABLE IF NOT EXISTS `users` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `name` VARCHAR(100) NOT NULL,
    `email` VARCHAR(150) NOT NULL UNIQUE,
    `password_hash` VARCHAR(255) NOT NULL,
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

File: `migrations/000002_create_users_table.down.sql`
```sql
DROP TABLE IF EXISTS `users`;
```

### 3.2. ORM Model (`models/user.go`)
```go
package models

import (
	"time"
	"github.com/beego/beego/v2/client/orm"
)

type User struct {
	Id           int       `orm:"column(id);auto;pk" json:"id"`
	Name         string    `orm:"column(name);size(100)" json:"name" valid:"Required"`
	Email        string    `orm:"column(email);size(150);unique" json:"email" valid:"Required;Email"`
	PasswordHash string    `orm:"column(password_hash);size(255)" json:"-"`
	CreatedAt    time.Time `orm:"column(created_at);auto_now_add;type(datetime)" json:"created_at"`
	UpdatedAt    time.Time `orm:"column(updated_at);auto_now;type(datetime)" json:"updated_at"`
}

func init() {
	orm.RegisterModel(new(User))
}

func (u *User) TableName() string {
	return "users"
}
```

## 4. Backend Components

### 4.1. DTO Layer (`dto/auth_dto.go`)
- `RegisterRequest`: `Name`, `Email`, `Password` (validated: min length 6)
- `LoginRequest`: `Email`, `Password`
- `UserResponse`: `Id`, `Name`, `Email`, `CreatedAt`
- `AuthResponse`: `Token`, `User` (`UserResponse`)

### 4.2. Repository Layer (`repositories/user_repository.go`)
- Interface `UserRepository`:
  - `Create(user *models.User) error`
  - `FindByEmail(email string) (*models.User, error)`
  - `FindByID(id int) (*models.User, error)`
- Implementation `userRepository` using Beego ORM.

### 4.3. Service Layer (`services/auth_service.go`)
- Interface `AuthService`:
  - `Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)`
  - `Login(req *dto.LoginRequest) (*dto.AuthResponse, error)`
  - `GetUserByID(id int) (*dto.UserResponse, error)`
  - `GenerateToken(user *models.User) (string, error)`
  - `ValidateToken(tokenString string) (*jwt.RegisteredClaims, error)`
- Features:
  - Passwords hashed using `bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)`.
  - Comparison using `bcrypt.CompareHashAndPassword`.
  - JWT token signed with HMAC-SHA256 using `jwt_secret` from `conf/app.conf` (fallback default for dev).
  - Claims contain `Subject: strconv.Itoa(user.Id)` and expiration time (24 hours).

### 4.4. Controller Layer (`controllers/auth_controller.go`)
- Endpoints:
  - `POST /api/v1/auth/register`: Unmarshals `RegisterRequest`, returns 201 with `AuthResponse`.
  - `POST /api/v1/auth/login`: Unmarshals `LoginRequest`, returns 200 with `AuthResponse`.
  - `GET /api/v1/auth/me`: Reads authenticated user ID from context, returns 200 with `UserResponse`.

### 4.5. Middleware / Filter (`middlewares/jwt_auth.go` or `routers/auth_filter.go`)
- Beego filter registered with `web.InsertFilter`:
  - Intercepts requests to `/api/v1/employees/*` and `/api/v1/auth/me`.
  - Extracts `Authorization` header with format `Bearer <token>`.
  - Validates JWT signature and expiration.
  - On invalid/missing token: responds immediately with HTTP 401 Unauthorized (`{"success": false, "message": "Unauthorized"}`).
  - On valid token: stores parsed `userID` into `ctx.Input.SetData("userID", id)` and continues.

### 4.6. Router Configuration (`routers/router.go`)
- Dependency injection: instantiate `UserRepository`, `AuthService`, and `AuthController`.
- Register routes:
  - `/api/v1/auth/register` -> `AuthController.Register`
  - `/api/v1/auth/login` -> `AuthController.Login`
  - `/api/v1/auth/me` -> `AuthController.Me`
- Apply JWT filter to `/api/v1/employees/*` and `/api/v1/auth/me`.

## 5. Frontend Architecture (Next.js)

### 5.1. Types & API Client (`frontend/src/types/auth.ts` & `frontend/src/lib/api.ts`)
- TypeScript interfaces: `User`, `AuthResponse`, `LoginPayload`, `RegisterPayload`.
- Auth API methods: `loginApi(payload)`, `registerApi(payload)`, `fetchCurrentUser(token)`.
- Token helper: `getAuthToken()`, `setAuthToken(token)`, `clearAuthToken()`.
- Update `frontend/src/lib/api.ts` employee fetch calls to include `Authorization: Bearer ${token}` header whenever present.

### 5.2. Auth State (`frontend/src/context/AuthContext.tsx`)
- Provides:
  - `user: User | null`
  - `token: string | null`
  - `isLoading: boolean`
  - `login(payload: LoginPayload): Promise<void>`
  - `register(payload: RegisterPayload): Promise<void>`
  - `logout(): void`
- On initial mount: reads token from `localStorage`, validates via `/api/v1/auth/me`, sets `user`. If invalid, clears storage.

### 5.3. Pages & UI Components
- `frontend/src/app/login/page.tsx`: Clean login form with email & password, error alerts, submit button, and link to `/register`.
- `frontend/src/app/register/page.tsx`: Clean registration form with name, email, password, error alerts, submit button, and link to `/login`.
- `frontend/src/components/auth/AuthGuard.tsx`: Wraps protected pages. If `isLoading`, shows loading skeleton/spinner. If not authenticated, redirects to `/login`.
- Header update in `frontend/src/app/page.tsx`: Adds authenticated user badge (name & email avatar) and "Sign Out" button.

## 6. Error Handling & Edge Cases
- **Duplicate Email Registration**: Returns 400 Bad Request with "email already exists".
- **Invalid Credentials on Login**: Returns 401 Unauthorized with "invalid email or password" (constant-time check).
- **Missing / Malformed Authorization Header**: Returns 401 Unauthorized.
- **Expired / Tampered Token**: Returns 401 Unauthorized.
- **CORS Preflight**: OPTIONS requests pass through unauthenticated to allow browser preflight.
- **Frontend Session Expiry**: When receiving a 401 response in API calls, trigger logout and redirect to `/login`.

## 7. Verification & Testing
1. **Migrations**:
   - Run `go run cmd/migrate/main.go up` -> verify `users` table created.
   - Run `go run cmd/migrate/main.go version` -> verify version 2.
2. **Go Backend Tests**:
   - Add unit tests for `services.AuthService` (password hashing, JWT creation & validation, duplicate handling).
   - Test endpoints via curl / HTTP tests:
     - Register user -> 201 Created with JWT token.
     - Login with correct credentials -> 200 OK with JWT token.
     - Login with wrong password -> 401 Unauthorized.
     - Access `/api/v1/employees` without header -> 401 Unauthorized.
     - Access `/api/v1/employees` with `Authorization: Bearer <token>` -> 200 OK.
3. **Frontend Build & Test**:
   - Verify `npm run build` passes with zero TypeScript or Lint errors.
   - Test UI flow in browser: Register -> auto-login/redirect -> manage employees -> logout -> login.
