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

// ContentSecurityCheckRequest 表示微信内容安全检查请求
type ContentSecurityCheckRequest struct {
	Content  string `json:"content"`            // 待检测的文本内容
	Version  int    `json:"version"`            // API版本，固定为2
	Scene    int    `json:"scene"`              // 场景值 1：资料；2：评论；3：论坛；4：社交
	OpenID   string `json:"openid"`             // 用户的OpenID
	Title    string `json:"title,omitempty"`    // 可选，文本标题
	Nickname string `json:"nickname,omitempty"` // 可选，用户昵称
}

// ContentSecurityCheckResult 表示内容安全检查结果
type ContentSecurityCheckResult struct {
	Suggest string `json:"suggest"` // 建议，有risky(风险)、pass(通过)、review(人工审核)三种值
	Label   int    `json:"label"`   // 标签值，对应不同类型的风险
}

// ContentSecurityCheckResponse 表示微信内容安全检查API的响应
type ContentSecurityCheckResponse struct {
	ErrCode int                        `json:"errcode"`  // 错误码
	ErrMsg  string                     `json:"errmsg"`   // 错误信息
	Result  ContentSecurityCheckResult `json:"result"`   // 检查结果
	TraceID string                     `json:"trace_id"` // 唯一请求标识
}

// SecuritySceneType 定义安全检查场景类型
type SecuritySceneType int

const (
	// SecuritySceneProfile 资料场景
	SecuritySceneProfile SecuritySceneType = 1
	// SecuritySceneComment 评论场景
	SecuritySceneComment SecuritySceneType = 2
	// SecuritySceneForum 论坛场景
	SecuritySceneForum SecuritySceneType = 3
	// SecuritySceneSocialLog 社交日志场景
	SecuritySceneSocialLog SecuritySceneType = 4
)

// ContentSecurityChecker 微信内容安全检查工具
type ContentSecurityChecker struct {
	client *http.Client
}

// NewContentSecurityChecker 创建一个新的内容安全检查工具
func NewContentSecurityChecker() *ContentSecurityChecker {
	return &ContentSecurityChecker{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckText 检查文本内容是否包含敏感词
// content: 要检查的文本内容
// openID: 用户的OpenID
// scene: 场景类型
// title: 可选的标题
// nickname: 可选的用户昵称
// 返回: 是否通过检查，详细结果，错误信息
func (c *ContentSecurityChecker) CheckText(content string, openID string, scene SecuritySceneType, title string, nickname string) (bool, *ContentSecurityCheckResponse, error) {
	if len(content) == 0 {
		return true, nil, errors.New("内容不能为空")
	}

	if len(content) > 2500 {
		return false, nil, errors.New("内容超过2500字符限制")
	}

	if openID == "" {
		return false, nil, errors.New("openid不能为空")
	}

	// 准备请求体
	reqBody := ContentSecurityCheckRequest{
		Content:  content,
		Version:  2, // 当前API版本为2
		Scene:    int(scene),
		OpenID:   openID,
		Title:    title,
		Nickname: nickname,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	// 构建请求
	// 在云托管环境下访问微信API不需要access_token
	req, err := http.NewRequest("POST", "https://api.weixin.qq.com/wxa/msg_sec_check", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// 设置请求ID，便于追踪
	requestID := uuid.New().String()
	req.Header.Set("X-Request-ID", requestID)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return false, nil, fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 解析响应
	var result ContentSecurityCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, nil, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查API错误
	if result.ErrCode != 0 {
		return false, &result, fmt.Errorf("微信API错误: %s (错误码: %d)", result.ErrMsg, result.ErrCode)
	}

	// 内容是否通过安全检查
	isPass := result.Result.Suggest == "pass"

	return isPass, &result, nil
}

// SimpleCheckText 简化版的内容检查函数，只返回是否通过和错误信息
func (c *ContentSecurityChecker) SimpleCheckText(content string, openID string, scene SecuritySceneType) (bool, error) {
	isPass, _, err := c.CheckText(content, openID, scene, "", "")
	return isPass, err
}

// CheckSensitiveContent 检查文本是否包含敏感内容
// 仅在生产环境下进行检查
// 这个函数可以被任何服务使用，便于进行内容安全检查
func CheckSensitiveContent(content string, openID string, scene SecuritySceneType) error {
	// 仅在生产环境检查
	if env := os.Getenv("APP_ENV"); env == "prod" {
		// 如果内容为空，则直接返回
		if content == "" {
			return nil
		}

		// 创建内容安全检查器
		checker := NewContentSecurityChecker()

		// 调用微信内容安全检查API
		isPass, err := checker.SimpleCheckText(content, openID, scene)
		if err != nil {
			slog.Error("Failed to check content security", "error", err)
			return apperrors.ErrContentSecurityAPI.WithDetails(err)
		}

		// 内容未通过安全检查
		if !isPass {
			slog.Warn("Content contains sensitive material", "content", content)
			return apperrors.ErrContentSecurityCheck
		}
	}

	return nil
}
