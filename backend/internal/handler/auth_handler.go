package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/service"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// GetGoogleAuthURL generates Google OAuth authentication URL
// @Summary Get Google OAuth URL
// @Description Generates OAuth URL for Google authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GetAuthURLRequest false "Optional redirect URL"
// @Success 200 {object} SuccessResponse{data=AuthURLResponse}
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/google/url [get]
func (h *AuthHandler) GetGoogleAuthURL(c *gin.Context) {
	var req GetAuthURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body provided, that's okay - redirect URL is optional
	}

	authURL, state, err := h.authService.GenerateAuthURL(c.Request.Context())
	if err != nil {
		RespondInternalServerError(c, "Failed to generate OAuth URL")
		return
	}

	RespondSuccess(c, "OAuth URL generated successfully", AuthURLResponse{
		AuthURL: authURL,
		State:   state,
	})
}

// HandleGoogleCallback handles OAuth callback from Google
// @Summary Handle Google OAuth callback
// @Description Processes OAuth callback and creates/updates user session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GoogleCallbackRequest true "OAuth callback parameters"
// @Success 200 {object} SuccessResponse{data=AuthResponse}
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid state or OAuth failure"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/google/callback [post]
func (h *AuthHandler) HandleGoogleCallback(c *gin.Context) {
	var req GoogleCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Handle OAuth callback
	user, accessToken, refreshToken, err := h.authService.HandleCallback(
		c.Request.Context(),
		req.Code,
		req.State,
	)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err, nil)
		return
	}

	// Convert domain user to response DTO
	userResponse := UserResponse{
		ID:                 user.ID,
		Email:              user.Email,
		GoogleID:           user.GoogleID,
		YouTubeChannelID:   user.YouTubeChannelID,
		YouTubeChannelName: user.YouTubeChannelName,
		ProfileImageURL:    user.ProfileImageURL,
		CreatedAt:          user.CreatedAt,
	}

	// Access token expires in 15 minutes (900 seconds)
	RespondSuccess(c, "Authentication successful", AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
		TokenType:    "Bearer",
		User:         userResponse,
	})
}

// RefreshToken refreshes the access token using refresh token
// @Summary Refresh access token
// @Description Generates new access token using valid refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} SuccessResponse{data=TokenRefreshResponse}
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized - Invalid or expired refresh token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondBadRequest(c, "Invalid request body", err.Error())
		return
	}

	// Refresh access token
	newAccessToken, err := h.authService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		RespondError(c, http.StatusUnauthorized, err, nil)
		return
	}

	RespondSuccess(c, "Token refreshed successfully", TokenRefreshResponse{
		AccessToken: newAccessToken,
		ExpiresIn:   900, // 15 minutes
		TokenType:   "Bearer",
	})
}

// GetCurrentUser retrieves the current authenticated user information
// @Summary Get current user
// @Description Retrieves authenticated user's information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse{data=UserResponse}
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 404 {object} ErrorResponse "User not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		RespondUnauthorized(c, "User not authenticated")
		return
	}

	// Get user by ID
	user, err := h.authService.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		if err == domain.ErrUserNotFound {
			RespondNotFound(c, "User not found")
			return
		}
		RespondInternalServerError(c, "Failed to retrieve user information")
		return
	}

	// Convert domain user to response DTO
	userResponse := UserResponse{
		ID:                 user.ID,
		Email:              user.Email,
		GoogleID:           user.GoogleID,
		YouTubeChannelID:   user.YouTubeChannelID,
		YouTubeChannelName: user.YouTubeChannelName,
		ProfileImageURL:    user.ProfileImageURL,
		CreatedAt:          user.CreatedAt,
	}

	RespondSuccess(c, "User information retrieved successfully", userResponse)
}

// Logout logs out the current user by blacklisting their JWT token
// @Summary Logout user
// @Description Blacklists the current JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} SuccessResponse
// @Failure 401 {object} ErrorResponse "Unauthorized - Missing or invalid token"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Extract user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		RespondUnauthorized(c, "User not authenticated")
		return
	}

	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		RespondUnauthorized(c, "Authorization header missing")
		return
	}

	// Parse Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		RespondBadRequest(c, "Invalid authorization header format", "Expected format: Bearer <token>")
		return
	}

	token := parts[1]

	// Blacklist the token
	if err := h.authService.Logout(c.Request.Context(), token, userID.(string)); err != nil {
		RespondError(c, http.StatusInternalServerError, err, nil)
		return
	}

	RespondSuccess(c, "Logged out successfully", nil)
}
