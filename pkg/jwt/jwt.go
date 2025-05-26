package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenDetails contains decoded token information
type TokenDetails struct {
	UserID uint
	Role   string
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(userID uint, role string, jwt_secret string, jwt_expiration_hours int) (string, error) {
	// Set token expiration time
	expirationTime := time.Now().Add(time.Hour * time.Duration(jwt_expiration_hours))

	// Create claims with user ID, role and expiration time
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
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
		UserID: claims.UserID,
		Role:   claims.Role,
	}, nil
}
