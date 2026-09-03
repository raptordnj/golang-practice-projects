package services

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"go-mvc/dto"
	"go-mvc/models"

	"github.com/golang-jwt/jwt/v5"
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

func TestAuthService_Register_ValidationErrors(t *testing.T) {
	repo := newMockUserRepository()
	authService := NewAuthService(repo, "test-secret-key")

	// Empty name
	_, err := authService.Register(&dto.RegisterRequest{
		Name:     "",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err == nil || err.Error() != "name is required" {
		t.Errorf("Expected 'name is required', got: %v", err)
	}

	// Empty email
	_, err = authService.Register(&dto.RegisterRequest{
		Name:     "Test User",
		Email:    "",
		Password: "password123",
	})
	if err == nil || err.Error() != "email is required" {
		t.Errorf("Expected 'email is required', got: %v", err)
	}

	// Short password
	_, err = authService.Register(&dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "12345",
	})
	if err == nil || err.Error() != "password must be at least 6 characters long" {
		t.Errorf("Expected 'password must be at least 6 characters long', got: %v", err)
	}

	// First successful registration
	_, err = authService.Register(&dto.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	// Duplicate email (case-insensitive)
	_, err = authService.Register(&dto.RegisterRequest{
		Name:     "Another User",
		Email:    "TEST@example.com",
		Password: "password456",
	})
	if err == nil || err.Error() != "email already exists" {
		t.Errorf("Expected 'email already exists', got: %v", err)
	}
}

func TestAuthService_Login_ValidationErrors(t *testing.T) {
	repo := newMockUserRepository()
	authService := NewAuthService(repo, "test-secret-key")

	// Empty credentials
	_, err := authService.Login(&dto.LoginRequest{Email: "", Password: ""})
	if err == nil || err.Error() != "email and password are required" {
		t.Errorf("Expected 'email and password are required', got: %v", err)
	}

	_, err = authService.Login(&dto.LoginRequest{Email: "user@example.com", Password: ""})
	if err == nil || err.Error() != "email and password are required" {
		t.Errorf("Expected 'email and password are required', got: %v", err)
	}

	// User not found
	_, err = authService.Login(&dto.LoginRequest{Email: "notfound@example.com", Password: "password123"})
	if err == nil || err.Error() != "invalid email or password" {
		t.Errorf("Expected 'invalid email or password', got: %v", err)
	}
}

func TestAuthService_GetUserByID(t *testing.T) {
	repo := newMockUserRepository()
	authService := NewAuthService(repo, "test-secret-key")

	regRes, err := authService.Register(&dto.RegisterRequest{
		Name:     "Bob Smith",
		Email:    "bob@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}

	user, err := authService.GetUserByID(regRes.User.Id)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if user.Name != "Bob Smith" || user.Email != "bob@example.com" {
		t.Errorf("Unexpected user: %+v", user)
	}

	_, err = authService.GetUserByID(999)
	if err == nil {
		t.Errorf("Expected error for non-existent ID, got nil")
	}
}

func TestAuthService_ValidateToken_Errors(t *testing.T) {
	repo := newMockUserRepository()
	authService := NewAuthService(repo, "test-secret-key")
	otherService := NewAuthService(repo, "different-secret-key")

	// Invalid token string
	_, err := authService.ValidateToken("invalid.token.string")
	if err == nil {
		t.Errorf("Expected error for invalid token string, got nil")
	}

	// Token signed with different secret
	user := &models.User{Id: 42, Email: "test@example.com"}
	token, err := otherService.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = authService.ValidateToken(token)
	if err == nil {
		t.Errorf("Expected signature error when validating token with different secret, got nil")
	}
}

func TestAuthService_NewAuthService_FallbackSecret(t *testing.T) {
	repo := newMockUserRepository()
	svc := NewAuthService(repo, "")
	if svc == nil {
		t.Fatal("Expected NewAuthService to return non-nil instance")
	}

	user := &models.User{Id: 10, Email: "fallback@example.com"}
	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed with fallback secret: %v", err)
	}
	if claims.Subject != "10" {
		t.Errorf("Expected subject '10', got '%s'", claims.Subject)
	}
}

func TestAuthService_ValidateToken_Expired(t *testing.T) {
	repo := newMockUserRepository()
	authService := NewAuthService(repo, "test-secret-key")

	claims := jwt.RegisteredClaims{
		Subject:   "1",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to sign expired token: %v", err)
	}

	_, err = authService.ValidateToken(tokenString)
	if err == nil {
		t.Errorf("Expected error for expired token, got nil")
	}
}

func TestAuthService_NilInputs_And_PasswordLength(t *testing.T) {
	repo := newMockUserRepository()
	authService := NewAuthService(repo, "test-secret-key")

	// Nil request on Register
	_, err := authService.Register(nil)
	if err == nil || err.Error() != "request cannot be nil" {
		t.Errorf("Expected 'request cannot be nil', got %v", err)
	}

	// Nil request on Login
	_, err = authService.Login(nil)
	if err == nil || err.Error() != "request cannot be nil" {
		t.Errorf("Expected 'request cannot be nil', got %v", err)
	}

	// Nil user on GenerateToken
	_, err = authService.GenerateToken(nil)
	if err == nil || err.Error() != "user cannot be nil" {
		t.Errorf("Expected 'user cannot be nil', got %v", err)
	}

	// Password exceeding 72 bytes on Register
	longPassword := strings.Repeat("a", 73)
	_, err = authService.Register(&dto.RegisterRequest{
		Name:     "Long Pass",
		Email:    "longpass@example.com",
		Password: longPassword,
	})
	if err == nil || err.Error() != "password cannot exceed 72 bytes" {
		t.Errorf("Expected 'password cannot exceed 72 bytes', got %v", err)
	}

	// Password exceeding 72 bytes on Login
	_, err = authService.Login(&dto.LoginRequest{
		Email:    "longpass@example.com",
		Password: longPassword,
	})
	if err == nil || err.Error() != "password cannot exceed 72 bytes" {
		t.Errorf("Expected 'password cannot exceed 72 bytes', got %v", err)
	}
}

func TestAuthService_NewAuthService_EnvSecret(t *testing.T) {
	repo := newMockUserRepository()
	os.Setenv("JWT_SECRET", "custom_env_secret")
	defer os.Unsetenv("JWT_SECRET")

	svc := NewAuthService(repo, "")
	user := &models.User{Id: 99, Email: "env@example.com"}
	token, err := svc.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed with env secret: %v", err)
	}
	if claims.Subject != "99" {
		t.Errorf("Expected subject '99', got '%s'", claims.Subject)
	}
}
