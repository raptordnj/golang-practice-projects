# Authentication MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement a fullstack Authentication MVP with email/password registration and login, bcrypt password hashing, JWT Bearer token protection for employee API endpoints, and a Next.js frontend with authentication state, login/register pages, and route guards.

**Architecture:** 
- Backend (Go + Beego v2): `models/user.go` ORM entity and database migration in `migrations/000002_create_users_table.*.sql`. `repositories/user_repository.go` for database operations. `services/auth_service.go` for bcrypt hashing and JWT token handling (`golang-jwt/jwt/v5`). `controllers/auth_controller.go` exposing `/api/v1/auth` endpoints. `middlewares/jwt_auth.go` filter protecting `/api/v1/employees/*` and `/api/v1/auth/me`.
- Frontend (Next.js 16 + React 19): `AuthContext` managing JWT token and user profile in `localStorage`. `frontend/src/lib/api.ts` passing `Authorization: Bearer <token>` to protected endpoints. Dedicated `/login` and `/register` routes with styled forms and validation. `AuthGuard` protecting the employee dashboard.

**Tech Stack:** Go 1.24, Beego v2, MySQL 8, `golang-migrate/migrate/v4`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`, Next.js 16, React 19, Tailwind CSS v4, Lucide React.

## Global Constraints
- Target database: MySQL with `utf8mb4` encoding.
- Passwords must be hashed using bcrypt (cost 10+). Never store plaintext passwords or serialize password hashes in JSON.
- JWT tokens signed using HMAC-SHA256 with secret configurable via `conf/app.conf` (`jwt_secret = ...`).
- Protected endpoints return `401 Unauthorized` with JSON `{"success": false, "message": "Unauthorized: ..."}` on missing or invalid token.
- Frontend routes `/login` and `/register` public; `/` (employee dashboard) protected by `AuthGuard`.

---

### Task 1: Add Users Migration and User ORM Model

**Files:**
- Create: `migrations/000002_create_users_table.up.sql`
- Create: `migrations/000002_create_users_table.down.sql`
- Create: `models/user.go`

**Interfaces:**
- Consumes: `golang-migrate` CLI (`cmd/migrate/main.go`), Beego ORM (`github.com/beego/beego/v2/client/orm`)
- Produces: `models.User` struct, `users` database table

- [ ] **Step 1: Create forward migration `migrations/000002_create_users_table.up.sql`**

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

- [ ] **Step 2: Create backward migration `migrations/000002_create_users_table.down.sql`**

```sql
DROP TABLE IF EXISTS `users`;
```

- [ ] **Step 3: Create ORM Model `models/user.go`**

```go
package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

// User represents the user model in the database
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

// TableName returns the database table name for User model
func (u *User) TableName() string {
	return "users"
}
```

- [ ] **Step 4: Run migration and verify table creation**

Run:
```bash
go run cmd/migrate/main.go up
go run cmd/migrate/main.go version
```
Expected output: Migration applies cleanly and version is 2.

- [ ] **Step 5: Commit**

```bash
git add migrations/000002_create_users_table.* models/user.go
git commit -m "feat(auth): add user migration and ORM model"
```

---

### Task 2: Create Auth DTOs and User Repository

**Files:**
- Create: `dto/auth_dto.go`
- Create: `repositories/user_repository.go`

**Interfaces:**
- Consumes: `models.User`, Beego ORM
- Produces: `dto.RegisterRequest`, `dto.LoginRequest`, `dto.UserResponse`, `dto.AuthResponse`, `repositories.UserRepository` interface and `NewUserRepository()`

- [ ] **Step 1: Create DTOs `dto/auth_dto.go`**

```go
package dto

import "time"

// RegisterRequest represents the user registration payload
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the user login payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse represents safe user data returned in responses
type UserResponse struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse represents the response containing token and user profile
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
```

- [ ] **Step 2: Create User Repository `repositories/user_repository.go`**

```go
package repositories

import (
	"errors"
	"go-mvc/models"

	"github.com/beego/beego/v2/client/orm"
)

// UserRepository defines the database operations for users
type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id int) (*models.User, error)
}

type userRepository struct {
	o orm.Ormer
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository() UserRepository {
	return &userRepository{
		o: orm.NewOrm(),
	}
}

// Create inserts a new user record
func (r *userRepository) Create(user *models.User) error {
	_, err := r.o.Insert(user)
	return err
}

// FindByEmail retrieves a user by email address
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{Email: email}
	err := r.o.Read(user, "Email")
	if errors.Is(err, orm.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return user, err
}

// FindByID retrieves a user by ID
func (r *userRepository) FindByID(id int) (*models.User, error) {
	user := &models.User{Id: id}
	err := r.o.Read(user)
	if errors.Is(err, orm.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return user, err
}
```

- [ ] **Step 3: Verify compilation**

Run:
```bash
go build ./...
```
Expected: Build succeeds with 0 errors.

- [ ] **Step 4: Commit**

```bash
git add dto/auth_dto.go repositories/user_repository.go
git commit -m "feat(auth): add auth DTOs and user repository"
```

---

### Task 3: Implement Auth Service with Bcrypt & JWT

**Files:**
- Modify: `go.mod`
- Create: `services/auth_service.go`
- Create: `services/auth_service_test.go`
- Modify: `conf/app.conf`

**Interfaces:**
- Consumes: `repositories.UserRepository`, `dto.*`, `golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt`
- Produces: `services.AuthService` interface with `Register`, `Login`, `GetUserByID`, `GenerateToken`, `ValidateToken`

- [ ] **Step 1: Install `github.com/golang-jwt/jwt/v5` and add JWT secret to `conf/app.conf`**

Run:
```bash
go get github.com/golang-jwt/jwt/v5
go mod tidy
```

In `conf/app.conf`, append:
```ini
jwt_secret = workpulse_super_secret_jwt_key_2026
```

- [ ] **Step 2: Write unit test `services/auth_service_test.go`**

```go
package services

import (
	"errors"
	"testing"

	"go-mvc/dto"
	"go-mvc/models"
)

type mockUserRepository struct {
	usersByEmail map[string]*models.User
	usersByID    map[int]*models.User
	nextID       int
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		usersByEmail: make(map[string]*models.User),
		usersByID:    make(map[int]*models.User),
		nextID:       1,
	}
}

func (m *mockUserRepository) Create(user *models.User) error {
	if _, exists := m.usersByEmail[user.Email]; exists {
		return errors.New("duplicate email")
	}
	user.Id = m.nextID
	m.nextID++
	m.usersByEmail[user.Email] = user
	m.usersByID[user.Id] = user
	return nil
}

func (m *mockUserRepository) FindByEmail(email string) (*models.User, error) {
	if u, exists := m.usersByEmail[email]; exists {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) FindByID(id int) (*models.User, error) {
	if u, exists := m.usersByID[id]; exists {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func TestAuthService_Register_And_Login(t *testing.T) {
	repo := newMockUserRepository()
	authService := NewAuthService(repo, "test-secret-key")

	// Register
	regReq := &dto.RegisterRequest{
		Name:     "Alice Doe",
		Email:    "alice@example.com",
		Password: "password123",
	}
	res, err := authService.Register(regReq)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if res.Token == "" {
		t.Errorf("Expected token, got empty string")
	}
	if res.User.Email != "alice@example.com" {
		t.Errorf("Expected email alice@example.com, got %s", res.User.Email)
	}

	// Login with correct credentials
	loginReq := &dto.LoginRequest{
		Email:    "alice@example.com",
		Password: "password123",
	}
	loginRes, err := authService.Login(loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loginRes.Token == "" {
		t.Errorf("Expected token, got empty string")
	}

	// Login with wrong password
	badReq := &dto.LoginRequest{
		Email:    "alice@example.com",
		Password: "wrongpassword",
	}
	_, err = authService.Login(badReq)
	if err == nil {
		t.Fatalf("Expected error for wrong password, got nil")
	}

	// Validate token
	claims, err := authService.ValidateToken(loginRes.Token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.Subject != "1" {
		t.Errorf("Expected subject '1', got '%s'", claims.Subject)
	}
}
```

- [ ] **Step 3: Run test to verify it fails before implementation**

Run:
```bash
go test ./services/...
```
Expected: Compilation failure because `NewAuthService` is not implemented yet.

- [ ] **Step 4: Implement `services/auth_service.go`**

```go
package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go-mvc/dto"
	"go-mvc/models"
	"go-mvc/repositories"

	"github.com/beego/beego/v2/server/web"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic
type AuthService interface {
	Register(req *dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req *dto.LoginRequest) (*dto.AuthResponse, error)
	GetUserByID(id int) (*dto.UserResponse, error)
	GenerateToken(user *models.User) (string, error)
	ValidateToken(tokenString string) (*jwt.RegisteredClaims, error)
}

type authService struct {
	repo      repositories.UserRepository
	jwtSecret []byte
}

// NewAuthService creates a new AuthService instance
func NewAuthService(repo repositories.UserRepository, secret string) AuthService {
	if secret == "" {
		cfgSecret, _ := web.AppConfig.String("jwt_secret")
		if cfgSecret != "" {
			secret = cfgSecret
		} else {
			secret = "workpulse_super_secret_jwt_key_2026"
		}
	}
	return &authService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

// Register creates a new user and issues a JWT token
func (s *authService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)

	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	if req.Email == "" {
		return nil, errors.New("email is required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters long")
	}

	// Check existing user
	if existing, _ := s.repo.FindByEmail(req.Email); existing != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			Id:        user.Id,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// Login validates user credentials and issues a JWT token
func (s *authService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || req.Password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			Id:        user.Id,
			Name:      user.Name,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// GetUserByID fetches user details by ID
func (s *authService) GetUserByID(id int) (*dto.UserResponse, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		Id:        user.Id,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

// GenerateToken generates a signed JWT token valid for 24 hours
func (s *authService) GenerateToken(user *models.User) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(user.Id),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken parses and validates a JWT token string
func (s *authService) ValidateToken(tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
```

- [ ] **Step 5: Run unit tests to verify they pass**

Run:
```bash
go test -v ./services/...
```
Expected: PASS with 100% success.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum conf/app.conf services/auth_service.go services/auth_service_test.go
git commit -m "feat(auth): implement auth service with bcrypt and JWT"
```

---

### Task 4: Auth Controller, JWT Filter & Router Wiring

**Files:**
- Create: `controllers/auth_controller.go`
- Create: `middlewares/jwt_auth.go`
- Modify: `routers/router.go`

**Interfaces:**
- Consumes: `services.AuthService`, `dto.*`, Beego `web.Controller`, Beego `web.InsertFilter`
- Produces: API endpoints `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/me`, and JWT protection filter.

- [ ] **Step 1: Create `controllers/auth_controller.go`**

```go
package controllers

import (
	"encoding/json"
	"strconv"

	"go-mvc/dto"
	"go-mvc/services"

	"github.com/beego/beego/v2/server/web"
)

// AuthController handles user authentication endpoints
type AuthController struct {
	web.Controller
	Service services.AuthService
}

// Register handles POST /api/v1/auth/register
func (c *AuthController) Register() {
	var req dto.RegisterRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	res, err := c.Service.Register(&req)
	if err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(201)
	c.Data["json"] = dto.APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    res,
	}
	c.ServeJSON()
}

// Login handles POST /api/v1/auth/login
func (c *AuthController) Login() {
	var req dto.LoginRequest
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		c.Ctx.Output.SetStatus(400)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
		}
		c.ServeJSON()
		return
	}

	res, err := c.Service.Login(&req)
	if err != nil {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: err.Error(),
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = dto.APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    res,
	}
	c.ServeJSON()
}

// Me handles GET /api/v1/auth/me
func (c *AuthController) Me() {
	userIDVal := c.Ctx.Input.GetData("userID")
	if userIDVal == nil {
		c.Ctx.Output.SetStatus(401)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "Unauthorized",
		}
		c.ServeJSON()
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		if idStr, okStr := userIDVal.(string); okStr {
			userID, _ = strconv.Atoi(idStr)
		}
	}

	user, err := c.Service.GetUserByID(userID)
	if err != nil {
		c.Ctx.Output.SetStatus(404)
		c.Data["json"] = dto.APIResponse{
			Success: false,
			Message: "User not found",
		}
		c.ServeJSON()
		return
	}

	c.Ctx.Output.SetStatus(200)
	c.Data["json"] = dto.APIResponse{
		Success: true,
		Data:    user,
	}
	c.ServeJSON()
}
```

- [ ] **Step 2: Create `middlewares/jwt_auth.go`**

```go
package middlewares

import (
	"encoding/json"
	"strconv"
	"strings"

	"go-mvc/dto"
	"go-mvc/services"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
)

// JWTAuthFilter returns a Beego filter handler that checks JWT Bearer tokens
func JWTAuthFilter(authService services.AuthService) func(ctx *beegoCtx.Context) {
	return func(ctx *beegoCtx.Context) {
		// Allow OPTIONS requests for CORS preflight
		if ctx.Input.Method() == "OPTIONS" {
			return
		}

		authHeader := ctx.Input.Header("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			respondUnauthorized(ctx, "Authorization token is missing or malformed")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			respondUnauthorized(ctx, "Invalid or expired token")
			return
		}

		userID, err := strconv.Atoi(claims.Subject)
		if err != nil {
			respondUnauthorized(ctx, "Invalid token subject")
			return
		}

		ctx.Input.SetData("userID", userID)
	}
}

func respondUnauthorized(ctx *beegoCtx.Context, message string) {
	ctx.Output.SetStatus(401)
	ctx.Output.Header("Content-Type", "application/json")
	resp, _ := json.Marshal(dto.APIResponse{
		Success: false,
		Message: message,
	})
	_ = ctx.Output.Body(resp)
}
```

- [ ] **Step 3: Update `routers/router.go`**

Update `routers/router.go` to wire `UserRepository`, `AuthService`, `AuthController`, and register the routes and JWT filter:

```go
package routers

import (
	"go-mvc/controllers"
	"go-mvc/middlewares"
	"go-mvc/repositories"
	"go-mvc/services"

	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/filter/cors"
)

func init() {
	// CORS filter
	web.InsertFilter("*", web.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin"},
		AllowCredentials: true,
	}))

	// Dependency injection
	employeeRepo := repositories.NewEmployeeRepository()
	employeeService := services.NewEmployeeService(employeeRepo)
	employeeController := &controllers.EmployeeController{
		Service: employeeService,
	}

	userRepo := repositories.NewUserRepository()
	authService := services.NewAuthService(userRepo, "")
	authController := &controllers.AuthController{
		Service: authService,
	}

	// JWT Auth filter for protected routes
	jwtFilter := middlewares.JWTAuthFilter(authService)
	web.InsertFilter("/api/v1/employees", web.BeforeRouter, jwtFilter)
	web.InsertFilter("/api/v1/employees/*", web.BeforeRouter, jwtFilter)
	web.InsertFilter("/api/v1/auth/me", web.BeforeRouter, jwtFilter)

	// API v1 namespace
	ns := web.NewNamespace("/api/v1",
		web.NSNamespace("/auth",
			web.NSRouter("/register", authController, "post:Register"),
			web.NSRouter("/login", authController, "post:Login"),
			web.NSRouter("/me", authController, "get:Me"),
		),
		web.NSNamespace("/employees",
			web.NSRouter("/", employeeController, "post:Create;get:GetAll"),
			web.NSRouter("/:id", employeeController, "get:GetOne;put:Update;delete:Delete"),
		),
	)

	web.AddNamespace(ns)
}
```

- [ ] **Step 4: Verify build and test package compilation**

Run:
```bash
go build ./...
```
Expected: Compiles cleanly.

- [ ] **Step 5: Commit**

```bash
git add controllers/auth_controller.go middlewares/jwt_auth.go routers/router.go
git commit -m "feat(auth): add auth controller, JWT filter, and route registration"
```

---

### Task 5: Frontend Auth Types, API Client & AuthContext

**Files:**
- Create: `frontend/src/types/auth.ts`
- Modify: `frontend/src/lib/api.ts`
- Create: `frontend/src/context/AuthContext.tsx`
- Modify: `frontend/src/app/layout.tsx`

**Interfaces:**
- Consumes: Next.js App Router, React 19 Context
- Produces: `AuthContext`, `useAuth` hook, `loginApi`, `registerApi`, `fetchCurrentUser`

- [ ] **Step 1: Create `frontend/src/types/auth.ts`**

```typescript
export interface User {
  id: number;
  name: string;
  email: string;
  created_at: string;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface LoginPayload {
  email: string;
  password: string;
}

export interface RegisterPayload {
  name: string;
  email: string;
  password: string;
}
```

- [ ] **Step 2: Update `frontend/src/lib/api.ts`**

Update `frontend/src/lib/api.ts` to include token helper functions and inject `Authorization: Bearer <token>` into API calls, plus auth functions:

```typescript
import {
  Employee,
  CreateEmployeePayload,
  UpdateEmployeePayload,
  PaginatedResponse,
  SingleResponse,
} from "@/types/employee";
import { AuthResponse, LoginPayload, RegisterPayload, User } from "@/types/auth";

const EMPLOYEES_URL = "/api/v1/employees";
const AUTH_URL = "/api/v1/auth";

const TOKEN_KEY = "workpulse_auth_token";

export function getStoredToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function setStoredToken(token: string): void {
  if (typeof window !== "undefined") {
    localStorage.setItem(TOKEN_KEY, token);
  }
}

export function clearStoredToken(): void {
  if (typeof window !== "undefined") {
    localStorage.removeItem(TOKEN_KEY);
  }
}

function getAuthHeaders(): HeadersInit {
  const token = getStoredToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  return headers;
}

export async function loginApi(payload: LoginPayload): Promise<AuthResponse> {
  const res = await fetch(`${AUTH_URL}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to log in");
  }
  return json.data;
}

export async function registerApi(payload: RegisterPayload): Promise<AuthResponse> {
  const res = await fetch(`${AUTH_URL}/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to register");
  }
  return json.data;
}

export async function fetchCurrentUser(): Promise<User> {
  const res = await fetch(`${AUTH_URL}/me`, {
    headers: getAuthHeaders(),
    cache: "no-store",
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to fetch current user");
  }
  return json.data;
}

export async function fetchEmployees(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<Employee>> {
  const res = await fetch(`${EMPLOYEES_URL}?page=${page}&page_size=${pageSize}`, {
    headers: getAuthHeaders(),
    cache: "no-store",
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to fetch employees: ${res.statusText}`);
  }
  return res.json();
}

export async function fetchEmployeeById(id: number): Promise<Employee> {
  const res = await fetch(`${EMPLOYEES_URL}/${id}`, {
    headers: getAuthHeaders(),
    cache: "no-store",
  });
  if (!res.ok) {
    const errorData = await res.json().catch(() => ({}));
    throw new Error(errorData.message || `Failed to fetch employee: ${res.statusText}`);
  }
  const json: SingleResponse<Employee> = await res.json();
  return json.data;
}

export async function createEmployee(
  payload: CreateEmployeePayload
): Promise<Employee> {
  const res = await fetch(EMPLOYEES_URL, {
    method: "POST",
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to create employee");
  }
  return json.data;
}

export async function updateEmployee(
  id: number,
  payload: UpdateEmployeePayload
): Promise<Employee> {
  const res = await fetch(`${EMPLOYEES_URL}/${id}`, {
    method: "PUT",
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to update employee");
  }
  return json.data;
}

export async function deleteEmployee(id: number): Promise<void> {
  const res = await fetch(`${EMPLOYEES_URL}/${id}`, {
    method: "DELETE",
    headers: getAuthHeaders(),
  });
  const json = await res.json();
  if (!res.ok || !json.success) {
    throw new Error(json.message || "Failed to delete employee");
  }
}
```

- [ ] **Step 3: Create `frontend/src/context/AuthContext.tsx`**

```tsx
"use client";

import React, { createContext, useContext, useEffect, useState } from "react";
import { User, LoginPayload, RegisterPayload } from "@/types/auth";
import {
  fetchCurrentUser,
  getStoredToken,
  setStoredToken,
  clearStoredToken,
  loginApi,
  registerApi,
} from "@/lib/api";

interface AuthContextType {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  login: (payload: LoginPayload) => Promise<void>;
  register: (payload: RegisterPayload) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const existingToken = getStoredToken();
    if (existingToken) {
      setToken(existingToken);
      fetchCurrentUser()
        .then((userData) => setUser(userData))
        .catch(() => {
          clearStoredToken();
          setToken(null);
          setUser(null);
        })
        .finally(() => setIsLoading(false));
    } else {
      setIsLoading(false);
    }
  }, []);

  const login = async (payload: LoginPayload) => {
    const authData = await loginApi(payload);
    setStoredToken(authData.token);
    setToken(authData.token);
    setUser(authData.user);
  };

  const register = async (payload: RegisterPayload) => {
    const authData = await registerApi(payload);
    setStoredToken(authData.token);
    setToken(authData.token);
    setUser(authData.user);
  };

  const logout = () => {
    clearStoredToken();
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, token, isLoading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
```

- [ ] **Step 4: Update `frontend/src/app/layout.tsx`**

Wrap `children` in `AuthProvider`:

```tsx
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/context/AuthContext";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "WorkPulse — Employee Management",
  description: "Modern Fullstack Employee Management System",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <AuthProvider>{children}</AuthProvider>
      </body>
    </html>
  );
}
```

- [ ] **Step 5: Verify build**

Run:
```bash
cd frontend && npm run build && cd ..
```
Expected: Next.js builds successfully.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/types/auth.ts frontend/src/lib/api.ts frontend/src/context/AuthContext.tsx frontend/src/app/layout.tsx
git commit -m "feat(auth): add auth context, types, and token-aware API client"
```

---

### Task 6: Frontend Login, Register Pages & AuthGuard

**Files:**
- Create: `frontend/src/components/auth/AuthGuard.tsx`
- Create: `frontend/src/app/login/page.tsx`
- Create: `frontend/src/app/register/page.tsx`
- Modify: `frontend/src/app/page.tsx`

**Interfaces:**
- Consumes: `useAuth`, Next.js `useRouter`, shadcn components
- Produces: `/login` page, `/register` page, `AuthGuard` wrapper, and user profile header with logout in `/`.

- [ ] **Step 1: Create `frontend/src/components/auth/AuthGuard.tsx`**

```tsx
"use client";

import React, { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && !user) {
      router.push("/login");
    }
  }, [isLoading, user, router]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950">
        <div className="flex flex-col items-center gap-3">
          <div className="w-8 h-8 border-4 border-indigo-600 border-t-transparent rounded-full animate-spin" />
          <p className="text-sm text-zinc-500 font-medium">Checking authentication...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  return <>{children}</>;
}
```

- [ ] **Step 2: Create `frontend/src/app/login/page.tsx`**

```tsx
"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";
import { Users, Lock, Mail, ArrowRight } from "lucide-react";

export default function LoginPage() {
  const { login } = useAuth();
  const router = useRouter();

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      await login({ email, password });
      router.push("/");
    } catch (err: any) {
      setError(err.message || "Failed to sign in");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950 px-4">
      <div className="w-full max-w-md bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-2xl p-8 shadow-xl">
        <div className="flex flex-col items-center text-center mb-8">
          <div className="w-12 h-12 bg-indigo-600 rounded-xl flex items-center justify-center text-white mb-3 shadow-md shadow-indigo-200 dark:shadow-none">
            <Users className="w-6 h-6" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
            WorkPulse
          </h1>
          <p className="text-sm text-zinc-500 mt-1">Sign in to your account</p>
        </div>

        {error && (
          <div className="mb-6 p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/40 dark:text-red-400 border border-red-200 dark:border-red-900 rounded-lg">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 uppercase tracking-wider mb-1">
              Email Address
            </label>
            <div className="relative">
              <Mail className="w-4 h-4 text-zinc-400 absolute left-3 top-3" />
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="name@company.com"
                className="w-full pl-9 pr-3 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white dark:focus:bg-zinc-900 transition"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 uppercase tracking-wider mb-1">
              Password
            </label>
            <div className="relative">
              <Lock className="w-4 h-4 text-zinc-400 absolute left-3 top-3" />
              <input
                type="password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full pl-9 pr-3 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white dark:focus:bg-zinc-900 transition"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full mt-2 flex items-center justify-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2.5 px-4 rounded-lg transition shadow-md shadow-indigo-200 dark:shadow-none disabled:opacity-50"
          >
            {isSubmitting ? "Signing in..." : "Sign In"}
            {!isSubmitting && <ArrowRight className="w-4 h-4" />}
          </button>
        </form>

        <p className="text-center text-xs text-zinc-500 mt-6">
          Don&apos;t have an account?{" "}
          <Link
            href="/register"
            className="text-indigo-600 dark:text-indigo-400 font-semibold hover:underline"
          >
            Create account
          </Link>
        </p>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Create `frontend/src/app/register/page.tsx`**

```tsx
"use client";

import React, { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { useAuth } from "@/context/AuthContext";
import { Users, Lock, Mail, User as UserIcon, ArrowRight } from "lucide-react";

export default function RegisterPage() {
  const { register } = useAuth();
  const router = useRouter();

  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      await register({ name, email, password });
      router.push("/");
    } catch (err: any) {
      setError(err.message || "Failed to create account");
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-zinc-50 dark:bg-zinc-950 px-4">
      <div className="w-full max-w-md bg-white dark:bg-zinc-900 border border-zinc-200 dark:border-zinc-800 rounded-2xl p-8 shadow-xl">
        <div className="flex flex-col items-center text-center mb-8">
          <div className="w-12 h-12 bg-indigo-600 rounded-xl flex items-center justify-center text-white mb-3 shadow-md shadow-indigo-200 dark:shadow-none">
            <Users className="w-6 h-6" />
          </div>
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-100">
            WorkPulse
          </h1>
          <p className="text-sm text-zinc-500 mt-1">Create a new account</p>
        </div>

        {error && (
          <div className="mb-6 p-3 text-sm text-red-600 bg-red-50 dark:bg-red-950/40 dark:text-red-400 border border-red-200 dark:border-red-900 rounded-lg">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 uppercase tracking-wider mb-1">
              Full Name
            </label>
            <div className="relative">
              <UserIcon className="w-4 h-4 text-zinc-400 absolute left-3 top-3" />
              <input
                type="text"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="John Doe"
                className="w-full pl-9 pr-3 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white dark:focus:bg-zinc-900 transition"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 uppercase tracking-wider mb-1">
              Email Address
            </label>
            <div className="relative">
              <Mail className="w-4 h-4 text-zinc-400 absolute left-3 top-3" />
              <input
                type="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="name@company.com"
                className="w-full pl-9 pr-3 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white dark:focus:bg-zinc-900 transition"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-zinc-700 dark:text-zinc-300 uppercase tracking-wider mb-1">
              Password (min 6 characters)
            </label>
            <div className="relative">
              <Lock className="w-4 h-4 text-zinc-400 absolute left-3 top-3" />
              <input
                type="password"
                required
                minLength={6}
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                className="w-full pl-9 pr-3 py-2 text-sm border border-zinc-300 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:bg-white dark:focus:bg-zinc-900 transition"
              />
            </div>
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="w-full mt-2 flex items-center justify-center gap-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium py-2.5 px-4 rounded-lg transition shadow-md shadow-indigo-200 dark:shadow-none disabled:opacity-50"
          >
            {isSubmitting ? "Creating account..." : "Sign Up"}
            {!isSubmitting && <ArrowRight className="w-4 h-4" />}
          </button>
        </form>

        <p className="text-center text-xs text-zinc-500 mt-6">
          Already have an account?{" "}
          <Link
            href="/login"
            className="text-indigo-600 dark:text-indigo-400 font-semibold hover:underline"
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Update `frontend/src/app/page.tsx` with `AuthGuard`, user avatar, and logout**

Wrap the content of `frontend/src/app/page.tsx` in `<AuthGuard>` and add user info and a logout button to the navbar.

- [ ] **Step 5: Verify build**

Run:
```bash
cd frontend && npm run build && cd ..
```
Expected: Build passes with 0 errors.

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/auth/AuthGuard.tsx frontend/src/app/login/page.tsx frontend/src/app/register/page.tsx frontend/src/app/page.tsx
git commit -m "feat(auth): add login and register pages and dashboard AuthGuard"
```

---

### Task 7: Fullstack End-to-End Verification & Documentation

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: Go backend, Next.js frontend, MySQL database
- Produces: Verified fullstack authentication flow, updated API documentation

- [ ] **Step 1: Apply database migrations**

Run:
```bash
go run cmd/migrate/main.go up
go run cmd/migrate/main.go version
```
Expected: Version 2, dirty false.

- [ ] **Step 2: Run Go backend tests**

Run:
```bash
go test -v ./...
```
Expected: All tests pass.

- [ ] **Step 3: Test API endpoints via curl**

1. Register user:
```bash
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"password123"}'
```
Expected: 201 Created with JSON containing JWT token.

2. Login user:
```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```
Expected: 200 OK with token.

3. Unauthenticated access to `/api/v1/employees`:
```bash
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/api/v1/employees
```
Expected: 401.

4. Authenticated access to `/api/v1/employees`:
```bash
curl -s -o /dev/null -w "%{http_code}\n" \
  -H "Authorization: Bearer <TOKEN>" \
  http://localhost:8080/api/v1/employees
```
Expected: 200.

- [ ] **Step 4: Update `README.md` with authentication endpoints & usage**

Add authentication documentation to `README.md`:
- Document `/api/v1/auth/register`, `/api/v1/auth/login`, and `/api/v1/auth/me`.
- Document how to obtain and pass the `Authorization: Bearer <token>` header.

- [ ] **Step 5: Commit documentation and complete verification**

```bash
git add README.md
git commit -m "docs: update README with authentication endpoints and usage"
```
