package middlewares

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"go-mvc/dto"
	"go-mvc/models"
	"go-mvc/services"

	beegoCtx "github.com/beego/beego/v2/server/web/context"
	"github.com/golang-jwt/jwt/v5"
)

type mockUserRepoForMiddleware struct{}

func (m *mockUserRepoForMiddleware) Create(user *models.User) error { return nil }
func (m *mockUserRepoForMiddleware) FindByEmail(email string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMiddleware) FindByID(id int) (*models.User, error) { return nil, nil }

func TestJWTAuthFilter_BearerSchemes(t *testing.T) {
	authService := services.NewAuthService(&mockUserRepoForMiddleware{}, "test-jwt-secret")
	token, err := authService.GenerateToken(&models.User{Id: 123, Email: "test@example.com"})
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	filter := JWTAuthFilter(authService)

	tests := []struct {
		name           string
		method         string
		authHeader     string
		expectedStatus int
		expectUserID   bool
	}{
		{
			name:           "OPTIONS preflight passes through",
			method:         "OPTIONS",
			authHeader:     "",
			expectedStatus: 200,
			expectUserID:   false,
		},
		{
			name:           "Standard Bearer header",
			method:         "GET",
			authHeader:     "Bearer " + token,
			expectedStatus: 200,
			expectUserID:   true,
		},
		{
			name:           "Lowercase bearer header",
			method:         "GET",
			authHeader:     "bearer " + token,
			expectedStatus: 200,
			expectUserID:   true,
		},
		{
			name:           "Mixed case bEaReR header",
			method:         "GET",
			authHeader:     "bEaReR " + token,
			expectedStatus: 200,
			expectUserID:   true,
		},
		{
			name:           "Missing Authorization header",
			method:         "GET",
			authHeader:     "",
			expectedStatus: 401,
			expectUserID:   false,
		},
		{
			name:           "Malformed Authorization header without prefix",
			method:         "GET",
			authHeader:     token,
			expectedStatus: 401,
			expectUserID:   false,
		},
		{
			name:           "Invalid scheme Basic",
			method:         "GET",
			authHeader:     "Basic " + token,
			expectedStatus: 401,
			expectUserID:   false,
		},
		{
			name:           "Invalid token string",
			method:         "GET",
			authHeader:     "Bearer invalid.token.value",
			expectedStatus: 401,
			expectUserID:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/v1/employees", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()

			ctx := beegoCtx.NewContext()
			ctx.Reset(w, req)

			filter(ctx)

			if tc.expectedStatus == 401 {
				if w.Code != 401 {
					t.Errorf("Expected status 401, got %d", w.Code)
				}
				var resp dto.APIResponse
				if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
					t.Errorf("Expected JSON response, unmarshal failed: %v", err)
				}
				if resp.Success {
					t.Errorf("Expected success=false, got true")
				}
			} else {
				if w.Code != 200 {
					t.Errorf("Expected status 200, got %d", w.Code)
				}
				if tc.expectUserID {
					val := ctx.Input.GetData("userID")
					if val == nil || val.(int) != 123 {
						t.Errorf("Expected userID 123 in context, got %v", val)
					}
				}
			}
		})
	}
}

func TestJWTAuthFilter_ExpiredToken(t *testing.T) {
	authService := services.NewAuthService(&mockUserRepoForMiddleware{}, "test-jwt-secret")

	claims := jwt.RegisteredClaims{
		Subject:   "123",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := expiredToken.SignedString([]byte("test-jwt-secret"))

	filter := JWTAuthFilter(authService)

	req := httptest.NewRequest("GET", "/api/v1/employees", nil)
	req.Header.Set("Authorization", "bearer "+tokenString)
	w := httptest.NewRecorder()

	ctx := beegoCtx.NewContext()
	ctx.Reset(w, req)

	filter(ctx)

	if w.Code != 401 {
		t.Errorf("Expected 401 for expired token, got %d", w.Code)
	}
}
