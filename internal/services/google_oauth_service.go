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

// GoogleUserInfo Google用户信息结构体
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

// GoogleOAuthService Google OAuth2服务接口
type GoogleOAuthService interface {
	// 验证重定向URI是否在配置的允许列表中（支持客户端类型）
	ValidateRedirectURI(redirectURI, clientType string) bool
	// 验证PKCE code verifier
	ValidateCodeVerifier(codeVerifier, codeChallenge string) bool
	// 交换授权码获取访问令牌并获取用户信息（支持客户端类型）
	ExchangeCodeForUserInfo(ctx context.Context, code, codeVerifier, redirectURI, clientType string) (*GoogleUserInfo, error)
}

type googleOAuthService struct {
	config *config.Config
}

// NewGoogleOAuthService 创建Google OAuth2服务实例
func NewGoogleOAuthService(config *config.Config) GoogleOAuthService {
	return &googleOAuthService{
		config: config,
	}
}

// ValidateRedirectURI 验证重定向URI是否在配置的允许列表中
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

// ValidateCodeVerifier 验证PKCE code verifier
// 根据RFC 7636规范，code_challenge = base64url(sha256(code_verifier))
func (s *googleOAuthService) ValidateCodeVerifier(codeVerifier, codeChallenge string) bool {
	if codeVerifier == "" || codeChallenge == "" {
		return false
	}

	// 计算code_verifier的SHA256哈希
	hash := sha256.Sum256([]byte(codeVerifier))

	// 使用base64 URL编码
	computed := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	return computed == codeChallenge
}

// ExchangeCodeForUserInfo 交换授权码获取访问令牌并获取用户信息
func (s *googleOAuthService) ExchangeCodeForUserInfo(ctx context.Context, code, codeVerifier, redirectURI, clientType string) (*GoogleUserInfo, error) {
	// 验证重定向URI
	if !s.ValidateRedirectURI(redirectURI, clientType) {
		slog.Warn("Invalid redirect URI", "redirectURI", redirectURI, "clientType", clientType)
		return nil, errors.ErrInvalidRedirectURL
	}

	// 根据客户端类型获取对应的配置
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

	// 配置OAuth2
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

	// 使用PKCE交换授权码获取访问令牌
	token, err := oauthConfig.Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		slog.Error("Failed to exchange OAuth code for token", "error", err)
		return nil, errors.ErrOAuthTokenExchange
	}

	// 使用访问令牌获取用户信息
	userInfo, err := s.fetchUserInfo(ctx, token.AccessToken)
	if err != nil {
		slog.Error("Failed to fetch user info", "error", err)
		return nil, errors.ErrOAuthUserInfoFetch
	}

	// 验证必要的用户信息字段
	if userInfo.Email == "" || userInfo.ID == "" {
		slog.Warn("Google user info incomplete", "userInfo", userInfo)
		return nil, errors.ErrGoogleUserInfoIncomplete
	}

	return userInfo, nil
}

// fetchUserInfo 使用访问令牌获取Google用户信息
func (s *googleOAuthService) fetchUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	// 方法1：使用Google API客户端库（推荐）
	oauth2Service, err := oauth2v2.NewService(ctx, option.WithTokenSource(
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to create oauth2 service: %w", err)
	}

	userInfoCall := oauth2Service.Userinfo.Get()
	userInfo, err := userInfoCall.Do()
	if err != nil {
		// 如果API客户端失败，尝试直接HTTP请求
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

// fetchUserInfoByHTTP 使用HTTP请求获取用户信息（备用方法）
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
