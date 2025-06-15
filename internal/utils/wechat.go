package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	apperrors "github.com/go-backend-template/internal/errors"
	"github.com/google/uuid"
)

// ContentSecurityCheckRequest represents the request for WeChat content security check.
type ContentSecurityCheckRequest struct {
	Content  string `json:"content"`            // Text content to be checked.
	Version  int    `json:"version"`            // API version, fixed at 2.
	Scene    int    `json:"scene"`              // Scene value: 1 for profile, 2 for comments, 3 for forum, 4 for social.
	OpenID   string `json:"openid"`             // User's OpenID.
	Title    string `json:"title,omitempty"`    // Optional, title of the text.
	Nickname string `json:"nickname,omitempty"` // Optional, user's nickname.
}

// ContentSecurityCheckResult represents the result of a content security check.
type ContentSecurityCheckResult struct {
	Suggest string `json:"suggest"` // Suggestion: "risky", "pass", or "review".
	Label   int    `json:"label"`   // Label value, corresponding to different types of risk.
}

// ContentSecurityCheckResponse represents the response from WeChat content security check API.
type ContentSecurityCheckResponse struct {
	ErrCode int                        `json:"errcode"`  // Error code.
	ErrMsg  string                     `json:"errmsg"`   // Error message.
	Result  ContentSecurityCheckResult `json:"result"`   // Check result.
	TraceID string                     `json:"trace_id"` // Unique request identifier.
}

// SecuritySceneType defines the type for security check scenes.
type SecuritySceneType int

const (
	// SecuritySceneProfile for profile information.
	SecuritySceneProfile SecuritySceneType = 1
	// SecuritySceneComment for comments.
	SecuritySceneComment SecuritySceneType = 2
	// SecuritySceneForum for forum posts.
	SecuritySceneForum SecuritySceneType = 3
	// SecuritySceneSocialLog for social logs/posts.
	SecuritySceneSocialLog SecuritySceneType = 4
)

// ContentSecurityChecker is a utility for WeChat content security checks.
type ContentSecurityChecker struct {
	client *http.Client
}

// NewContentSecurityChecker creates a new ContentSecurityChecker.
func NewContentSecurityChecker() *ContentSecurityChecker {
	return &ContentSecurityChecker{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckText checks if the text content contains sensitive words.
// content: The text content to check.
// openID: The user's OpenID.
// scene: The type of scene.
// title: Optional title.
// nickname: Optional user nickname.
// Returns: whether the check passed, detailed result, error information.
func (c *ContentSecurityChecker) CheckText(content string, openID string, scene SecuritySceneType, title string, nickname string) (bool, *ContentSecurityCheckResponse, error) {
	if len(content) == 0 {
		return true, nil, errors.New("content cannot be empty") // "内容不能为空" -> "content cannot be empty"
	}

	if len(content) > 2500 { // WeChat limit is 2500 characters for this API
		return false, nil, errors.New("content exceeds 2500 character limit") // "内容超过2500字符限制" -> "content exceeds 2500 character limit"
	}

	if openID == "" {
		return false, nil, errors.New("openid cannot be empty") // "openid不能为空" -> "openid cannot be empty"
	}

	// Prepare request body.
	reqBody := ContentSecurityCheckRequest{
		Content:  content,
		Version:  2, // Current API version is 2.
		Scene:    int(scene),
		OpenID:   openID,
		Title:    title,
		Nickname: nickname,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, nil, fmt.Errorf("failed to serialize request: %w", err) // "序列化请求失败:" -> "failed to serialize request:"
	}

	// Construct request.
	// Accessing WeChat API in a cloud-hosted environment does not require an access_token.
	req, err := http.NewRequest("POST", "https://api.weixin.qq.com/wxa/msg_sec_check", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, nil, fmt.Errorf("failed to create request: %w", err) // "创建请求失败:" -> "failed to create request:"
	}

	req.Header.Set("Content-Type", "application/json")
	// Set request ID for traceability.
	requestID := uuid.New().String()
	req.Header.Set("X-Request-ID", requestID)

	// Send request.
	resp, err := c.client.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("failed to send request: %w", err) // "发送请求失败:" -> "failed to send request:"
	}
	defer resp.Body.Close()

	// Parse response.
	var result ContentSecurityCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil, fmt.Errorf("failed to parse response: %w", err) // "解析响应失败:" -> "failed to parse response:"
	}

	// Check for API errors.
	if result.ErrCode != 0 {
		return false, &result, fmt.Errorf("WeChat API error: %s (errcode: %d)", result.ErrMsg, result.ErrCode) // "微信API错误: %s (错误码: %d)" -> "WeChat API error: %s (errcode: %d)"
	}

	// Did the content pass the security check?
	isPass := result.Result.Suggest == "pass"

	return isPass, &result, nil
}

// SimpleCheckText is a simplified content check function that only returns pass/fail and error information.
func (c *ContentSecurityChecker) SimpleCheckText(content string, openID string, scene SecuritySceneType) (bool, error) {
	isPass, _, err := c.CheckText(content, openID, scene, "", "")
	return isPass, err
}

// CheckSensitiveContent checks if text contains sensitive content.
// This check is performed only in the production environment.
// This function can be used by any service for content security checks.
func CheckSensitiveContent(content string, openID string, scene SecuritySceneType) error {
	// Only check in production environment.
	if env := os.Getenv("APP_ENV"); env == "prod" {
		// If content is empty, return directly.
		if content == "" {
			return nil
		}

		// Create a content security checker.
		checker := NewContentSecurityChecker()

		// Call WeChat content security check API.
		isPass, err := checker.SimpleCheckText(content, openID, scene)
		if err != nil {
			slog.Error("Failed to check content security", "error", err)
			return apperrors.ErrContentSecurityAPI.WithDetails(err)
		}

		// Content did not pass security check.
		if !isPass {
			slog.Warn("Content contains sensitive material", "content", content)
			return apperrors.ErrContentSecurityCheck
		}
	}

	return nil
}
