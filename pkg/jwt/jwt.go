package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID    uint      `json:"user_id"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"token_type"` // 新增字段用于区分token类型
	jwt.RegisteredClaims
}

// TokenDetails contains decoded token information
type TokenDetails struct {
	UserID    uint
	Role      string
	TokenType TokenType
}

// generateTokenWithDuration creates a JWT token with custom expiration duration
func generateTokenWithDuration(userID uint, role string, tokenType TokenType, jwt_secret string, duration time.Duration) (string, error) {
	// Set token expiration time
	expirationTime := time.Now().Add(duration)

	// Create claims with user ID, role, token type and expiration time
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "go-backend-template",
		},
	}

	// Create token with claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret
	tokenString, err := token.SignedString([]byte(jwt_secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateAccessToken creates a new JWT access token for a user
func GenerateAccessToken(userID uint, role string, jwt_secret string, jwt_expiration_hours int) (string, error) {
	duration := time.Hour * time.Duration(jwt_expiration_hours)
	return generateTokenWithDuration(userID, role, AccessToken, jwt_secret, duration)
}

// GenerateRefreshToken creates a new JWT refresh token for a user
func GenerateRefreshToken(userID uint, role string, jwt_secret string, jwt_expiration_hours int) (string, error) {
	// Refresh token typically expires in 30 days (720 hours) or longer
	refreshExpirationHours := jwt_expiration_hours * 10 // 10x longer than access token
	if refreshExpirationHours < 720 {                   // minimum 30 days
		refreshExpirationHours = 720
	}

	duration := time.Hour * time.Duration(refreshExpirationHours)
	return generateTokenWithDuration(userID, role, RefreshToken, jwt_secret, duration)
}

// ValidateToken validates the JWT token and returns the user details
func ValidateToken(tokenString string, jwt_secret string) (*TokenDetails, error) {
	// Parse token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwt_secret), nil
	})

	// Check parsing errors
	if err != nil {
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Return user details from claims
	return &TokenDetails{
		UserID:    claims.UserID,
		Role:      claims.Role,
		TokenType: claims.TokenType,
	}, nil
}
