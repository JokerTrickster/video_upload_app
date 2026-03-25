package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupResponseTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	return c, w
}

func TestRespondSuccess(t *testing.T) {
	c, w := setupResponseTestContext()

	data := map[string]string{"key": "value"}
	RespondSuccess(c, "operation successful", data)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "operation successful", resp.Message)
	assert.NotNil(t, resp.Data)
}

func TestRespondSuccess_NilData(t *testing.T) {
	c, w := setupResponseTestContext()

	RespondSuccess(c, "done", nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Nil(t, resp.Data)
}

func TestRespondBadRequest(t *testing.T) {
	c, w := setupResponseTestContext()

	RespondBadRequest(c, "invalid input", "field X is required")

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "bad_request", resp.Error)
	assert.Equal(t, "invalid input", resp.Message)
}

func TestRespondUnauthorized(t *testing.T) {
	c, w := setupResponseTestContext()

	RespondUnauthorized(c, "token expired")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", resp.Error)
	assert.Equal(t, "token expired", resp.Message)
}

func TestRespondNotFound(t *testing.T) {
	c, w := setupResponseTestContext()

	RespondNotFound(c, "user not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "not_found", resp.Error)
}

func TestRespondInternalServerError(t *testing.T) {
	c, w := setupResponseTestContext()

	RespondInternalServerError(c, "something went wrong")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "internal_server_error", resp.Error)
}

func TestRespondTooManyRequests(t *testing.T) {
	c, w := setupResponseTestContext()

	RespondTooManyRequests(c, "rate limit exceeded")

	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var resp ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "rate_limit_exceeded", resp.Error)
}

func TestRespondError_DomainErrorMapping(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "user not found",
			err:            domain.ErrUserNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "user_not_found",
		},
		{
			name:           "user already exists",
			err:            domain.ErrUserAlreadyExists,
			expectedStatus: http.StatusConflict,
			expectedCode:   "user_already_exists",
		},
		{
			name:           "invalid user data",
			err:            domain.ErrInvalidUserData,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "invalid_user_data",
		},
		{
			name:           "token not found",
			err:            domain.ErrTokenNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   "token_not_found",
		},
		{
			name:           "token expired",
			err:            domain.ErrTokenExpired,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "token_expired",
		},
		{
			name:           "token invalid",
			err:            domain.ErrTokenInvalid,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "token_invalid",
		},
		{
			name:           "token blacklisted",
			err:            domain.ErrTokenBlacklisted,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "token_blacklisted",
		},
		{
			name:           "unauthorized",
			err:            domain.ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "unauthorized",
		},
		{
			name:           "invalid credentials",
			err:            domain.ErrInvalidCredentials,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "invalid_credentials",
		},
		{
			name:           "invalid state",
			err:            domain.ErrInvalidState,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "invalid_state",
		},
		{
			name:           "oauth failed",
			err:            domain.ErrOAuthFailed,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   "oauth_failed",
		},
		{
			name:           "invalid input",
			err:            domain.ErrInvalidInput,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "invalid_input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, w := setupResponseTestContext()

			RespondError(c, http.StatusInternalServerError, tt.err, nil)

			assert.Equal(t, tt.expectedStatus, w.Code,
				"status code mismatch for %s: got %d, want %d", tt.name, w.Code, tt.expectedStatus)

			var resp ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCode, resp.Error,
				"error code mismatch for %s: got %s, want %s", tt.name, resp.Error, tt.expectedCode)
		})
	}
}
