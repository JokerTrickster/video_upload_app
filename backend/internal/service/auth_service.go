package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/JokerTrickster/video-upload-backend/internal/config"
	"github.com/JokerTrickster/video-upload-backend/internal/domain"
	"github.com/JokerTrickster/video-upload-backend/internal/repository"
)

// AuthService defines authentication business logic
type AuthService interface {
	// GenerateAuthURL generates Google OAuth authentication URL
	GenerateAuthURL(ctx context.Context) (authURL string, state string, err error)

	// HandleCallback handles OAuth callback and creates/updates user
	HandleCallback(ctx context.Context, code, state string) (*domain.User, string, string, error)

	// GenerateJWT generates JWT access and refresh tokens
	GenerateJWT(ctx context.Context, userID string) (accessToken, refreshToken string, err error)

	// ValidateJWT validates JWT token and returns claims
	ValidateJWT(ctx context.Context, token string) (*JWTClaims, error)

	// RefreshAccessToken refreshes access token using refresh token
	RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken string, err error)

	// GetUserByID retrieves user by ID
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)

	// Logout blacklists JWT token
	Logout(ctx context.Context, token string, userID string) error
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// authService implements AuthService interface
type authService struct {
	userRepo           repository.UserRepository
	tokenRepo          repository.TokenRepository
	tokenService       TokenService
	youtubeService     YouTubeService
	oauthConfig        *oauth2.Config
	jwtSecret          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewAuthService creates new AuthService instance
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	tokenService TokenService,
	youtubeService YouTubeService,
	cfg *config.Config,
) AuthService {
	return &authService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		tokenService:   tokenService,
		youtubeService: youtubeService,
		oauthConfig: &oauth2.Config{
			ClientID:     cfg.Google.ClientID,
			ClientSecret: cfg.Google.ClientSecret,
			RedirectURL:  cfg.Google.RedirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/youtube.upload",
				"https://www.googleapis.com/auth/youtube",
			},
			Endpoint: google.Endpoint,
		},
		jwtSecret:          cfg.JWT.Secret,
		accessTokenExpiry:  cfg.JWT.AccessTokenExpiry,
		refreshTokenExpiry: cfg.JWT.RefreshTokenExpiry,
	}
}

// GenerateAuthURL generates OAuth URL with random state
func (s *authService) GenerateAuthURL(ctx context.Context) (string, string, error) {
	state, err := generateRandomState(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Save state to Redis with 10-minute TTL
	if err := s.tokenService.SaveOAuthState(ctx, state, 600); err != nil {
		return "", "", fmt.Errorf("failed to save state: %w", err)
	}

	authURL := s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	return authURL, state, nil
}

// generateRandomState generates cryptographically secure random state
func generateRandomState(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HandleCallback handles OAuth callback
func (s *authService) HandleCallback(ctx context.Context, code, state string) (*domain.User, string, string, error) {
	// Validate state parameter
	valid, err := s.tokenService.ValidateOAuthState(ctx, state)
	if err != nil || !valid {
		return nil, "", "", domain.ErrInvalidState
	}

	// Exchange code for tokens
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to exchange code: %w", err)
	}

	// Get user profile
	profile, err := s.youtubeService.GetUserProfile(ctx, token.AccessToken)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get user profile: %w", err)
	}

	// Get YouTube channel info (optional)
	channelInfo, _ := s.youtubeService.GetChannelInfo(ctx, token.AccessToken)

	// Find or create user
	user, err := s.userRepo.FindByGoogleID(ctx, profile.GoogleID)
	if err == domain.ErrUserNotFound {
		// Create new user
		user = &domain.User{
			Email:           profile.Email,
			GoogleID:        profile.GoogleID,
			ProfileImageURL: &profile.Picture,
		}

		if channelInfo != nil {
			user.YouTubeChannelID = &channelInfo.ChannelID
			user.YouTubeChannelName = &channelInfo.ChannelName
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, "", "", fmt.Errorf("failed to create user: %w", err)
		}
	} else if err != nil {
		return nil, "", "", err
	} else {
		// Update existing user
		if channelInfo != nil {
			user.YouTubeChannelID = &channelInfo.ChannelID
			user.YouTubeChannelName = &channelInfo.ChannelName
		}
		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, "", "", fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Encrypt and save OAuth tokens
	encryptedAccess, err := s.tokenService.EncryptToken(ctx, token.AccessToken)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encryptedRefresh, err := s.tokenService.EncryptToken(ctx, token.RefreshToken)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	userToken := &domain.Token{
		UserID:                user.ID,
		EncryptedAccessToken:  encryptedAccess,
		EncryptedRefreshToken: encryptedRefresh,
		TokenType:             "Bearer",
		ExpiresAt:             token.Expiry,
	}

	// Save or update token
	existingToken, err := s.tokenRepo.FindByUserID(ctx, user.ID.String())
	if err == domain.ErrTokenNotFound {
		if err := s.tokenRepo.Create(ctx, userToken); err != nil {
			return nil, "", "", fmt.Errorf("failed to save token: %w", err)
		}
	} else if err != nil {
		return nil, "", "", err
	} else {
		userToken.ID = existingToken.ID
		if err := s.tokenRepo.Update(ctx, userToken); err != nil {
			return nil, "", "", fmt.Errorf("failed to update token: %w", err)
		}
	}

	// Generate JWT tokens
	accessToken, refreshToken, err := s.GenerateJWT(ctx, user.ID.String())
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return user, accessToken, refreshToken, nil
}

// GenerateJWT generates JWT access and refresh tokens
func (s *authService) GenerateJWT(ctx context.Context, userID string) (string, string, error) {
	now := time.Now()

	// Get user for email
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", "", err
	}

	// Generate access token
	accessClaims := JWTClaims{
		UserID:    userID,
		Email:     user.Email,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenExpiry)),
		},
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := JWTClaims{
		UserID:    userID,
		Email:     user.Email,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenExpiry)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessTokenString, refreshTokenString, nil
}

// ValidateJWT validates JWT token and returns claims
func (s *authService) ValidateJWT(ctx context.Context, tokenString string) (*JWTClaims, error) {
	// Check if token is blacklisted
	blacklisted, err := s.tokenService.IsBlacklisted(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, domain.ErrTokenBlacklisted
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, domain.ErrTokenInvalid
}

// RefreshAccessToken refreshes access token
func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := s.ValidateJWT(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	if claims.TokenType != "refresh" {
		return "", domain.ErrTokenInvalid
	}

	// Generate new access token only
	now := time.Now()
	accessClaims := JWTClaims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return accessToken, nil
}

// GetUserByID retrieves user by ID
func (s *authService) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// Logout blacklists JWT token
func (s *authService) Logout(ctx context.Context, token string, userID string) error {
	claims, err := s.ValidateJWT(ctx, token)
	if err != nil {
		return err
	}

	// Calculate TTL (time until token expires)
	ttl := claims.ExpiresAt.Time.Unix() - time.Now().Unix()
	if ttl <= 0 {
		return nil // Token already expired
	}

	return s.tokenService.AddToBlacklist(ctx, token, ttl)
}
