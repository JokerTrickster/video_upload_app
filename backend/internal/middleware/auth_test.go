package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/service"
)

// MockAuthService mocks service.AuthService for middleware tests
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateAuthURL(ctx context.Context) (string, string, error) {
	args := m.Called(ctx)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) HandleCallback(ctx context.Context, code, state string) (*domain.User, string, string, error) {
	args := m.Called(ctx, code, state)
	if args.Get(0) == nil {
		return nil, "", "", args.Error(3)
	}
	return args.Get(0).(*domain.User), args.String(1), args.String(2), args.Error(3)
}

func (m *MockAuthService) GenerateJWT(ctx context.Context, userID string) (string, string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) ValidateJWT(ctx context.Context, token string) (*service.JWTClaims, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.JWTClaims), args.Error(1)
}

func (m *MockAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	args := m.Called(ctx, refreshToken)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, token string, userID string) error {
	args := m.Called(ctx, token, userID)
	return args.Error(0)
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	mockAuth := new(MockAuthService)

	router := gin.New()
	router.Use(AuthMiddleware(mockAuth))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	mockAuth := new(MockAuthService)

	router := gin.New()
	router.Use(AuthMiddleware(mockAuth))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name       string
		authHeader string
	}{
		{name: "no Bearer prefix", authHeader: "token-value"},
		{name: "wrong prefix", authHeader: "Basic token-value"},
		{name: "only Bearer", authHeader: "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusOK, w.Code,
				"should not return 200 for auth header: %s", tt.authHeader)
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockAuth.On("ValidateJWT", mock.Anything, "invalid-token").
		Return(nil, domain.ErrTokenInvalid)

	router := gin.New()
	router.Use(AuthMiddleware(mockAuth))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuth.AssertExpectations(t)
}

func TestAuthMiddleware_RefreshTokenRejected(t *testing.T) {
	mockAuth := new(MockAuthService)
	claims := &service.JWTClaims{
		UserID:    "user-123",
		Email:     "test@example.com",
		TokenType: "refresh", // refresh tokens should be rejected
	}
	mockAuth.On("ValidateJWT", mock.Anything, "refresh-token").
		Return(claims, nil)

	router := gin.New()
	router.Use(AuthMiddleware(mockAuth))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer refresh-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidAccessToken(t *testing.T) {
	mockAuth := new(MockAuthService)
	claims := &service.JWTClaims{
		UserID:    "user-123",
		Email:     "test@example.com",
		TokenType: "access",
	}
	mockAuth.On("ValidateJWT", mock.Anything, "valid-access-token").
		Return(claims, nil)

	var capturedUserID string
	var capturedEmail string

	router := gin.New()
	router.Use(AuthMiddleware(mockAuth))
	router.GET("/protected", func(c *gin.Context) {
		if v, ok := c.Get("user_id"); ok {
			capturedUserID = v.(string)
		}
		if v, ok := c.Get("user_email"); ok {
			capturedEmail = v.(string)
		}
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-access-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "user-123", capturedUserID)
	assert.Equal(t, "test@example.com", capturedEmail)
	mockAuth.AssertExpectations(t)
}
