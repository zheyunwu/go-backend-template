package services

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	oauth2v2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// GoogleUserInfo represents the structure of user information obtained from Google.
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// GoogleOAuthService defines the interface for Google OAuth2 service.
type GoogleOAuthService interface {
	// ValidateRedirectURI checks if the redirect URI is in the configured allowed list (supports client types).
	ValidateRedirectURI(redirectURI, clientType string) bool
	// ValidateCodeVerifier validates the PKCE code verifier.
	ValidateCodeVerifier(codeVerifier, codeChallenge string) bool
	// ExchangeCodeForUserInfo exchanges the authorization code for an access token and fetches user info (supports client types).
	ExchangeCodeForUserInfo(ctx context.Context, code, codeVerifier, redirectURI, clientType string) (*GoogleUserInfo, error)
}

type googleOAuthService struct {
	config *config.Config
}

// NewGoogleOAuthService creates a new instance of GoogleOAuthService.
func NewGoogleOAuthService(config *config.Config) GoogleOAuthService {
	return &googleOAuthService{
		config: config,
	}
}

// ValidateRedirectURI checks if the redirect URI is in the configured allowed list.
func (s *googleOAuthService) ValidateRedirectURI(redirectURI, clientType string) bool {
	var allowedURIs []string
	switch clientType {
	case "ios":
		allowedURIs = s.config.Google.IOS.RedirectURLs
	case "web":
		allowedURIs = s.config.Google.Web.RedirectURLs
	default:
		return false
	}

	for _, allowedURI := range allowedURIs {
		if redirectURI == allowedURI {
			return true
		}
	}
	return false
}

// ValidateCodeVerifier validates the PKCE code verifier.
// According to RFC 7636, code_challenge = base64url(sha256(code_verifier)).
func (s *googleOAuthService) ValidateCodeVerifier(codeVerifier, codeChallenge string) bool {
	if codeVerifier == "" || codeChallenge == "" {
		return false
	}

	// Calculate SHA256 hash of the code_verifier.
	hash := sha256.Sum256([]byte(codeVerifier))

	// Use base64 URL encoding.
	computed := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	return computed == codeChallenge
}

// ExchangeCodeForUserInfo exchanges the authorization code for an access token and fetches user info.
func (s *googleOAuthService) ExchangeCodeForUserInfo(ctx context.Context, code, codeVerifier, redirectURI, clientType string) (*GoogleUserInfo, error) {
	// Validate the redirect URI.
	if !s.ValidateRedirectURI(redirectURI, clientType) {
		slog.WarnContext(ctx, "Invalid redirect URI", "redirectURI", redirectURI, "clientType", clientType) // Use slog.WarnContext
		return nil, errors.ErrInvalidRedirectURL
	}

	// Get corresponding configuration based on client type.
	var clientID, clientSecret string
	switch clientType {
	case "ios":
		clientID = s.config.Google.IOS.ClientID
		clientSecret = s.config.Google.IOS.ClientSecret
	case "web":
		clientID = s.config.Google.Web.ClientID
		clientSecret = s.config.Google.Web.ClientSecret
	default:
		return nil, errors.ErrInvalidClientType
	}

	// Configure OAuth2.
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Exchange authorization code for access token using PKCE.
	token, err := oauthConfig.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		slog.ErrorContext(ctx, "Failed to exchange OAuth code for token", "error", err) // Use slog.ErrorContext
		return nil, errors.ErrOAuthTokenExchange
	}

	// Fetch user information using the access token.
	userInfo, err := s.fetchUserInfo(ctx, token.AccessToken)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch user info", "error", err) // Use slog.ErrorContext
		return nil, errors.ErrOAuthUserInfoFetch
	}

	// Validate necessary user information fields.
	if userInfo.Email == "" || userInfo.ID == "" {
		slog.WarnContext(ctx, "Google user info incomplete", "userInfo", userInfo) // Use slog.WarnContext
		return nil, errors.ErrGoogleUserInfoIncomplete
	}

	return userInfo, nil
}

// fetchUserInfo fetches Google user information using an access token.
func (s *googleOAuthService) fetchUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	// Method 1: Use Google API client library (recommended).
	oauth2Service, err := oauth2v2.NewService(ctx, option.WithTokenSource(
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 service: %w", err)
	}

	userInfoCall := oauth2Service.Userinfo.Get()
	userInfo, err := userInfoCall.Do()
	if err != nil {
		// If API client fails, try direct HTTP request.
		return s.fetchUserInfoByHTTP(accessToken)
	}

	return &GoogleUserInfo{
		ID:            userInfo.Id,
		Email:         userInfo.Email,
		VerifiedEmail: userInfo.VerifiedEmail != nil && *userInfo.VerifiedEmail,
		Name:          userInfo.Name,
		GivenName:     userInfo.GivenName,
		FamilyName:    userInfo.FamilyName,
		Picture:       userInfo.Picture,
		Locale:        userInfo.Locale,
	}, nil
}

// fetchUserInfoByHTTP fetches user information using an HTTP request (fallback method).
func (s *googleOAuthService) fetchUserInfoByHTTP(accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user info, status: %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}
