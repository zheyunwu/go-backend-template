package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/response"
)

type AuthHandler struct {
	UserService services.UserService
}

func NewAuthHandler(UserService services.UserService) *AuthHandler {
	return &AuthHandler{
		UserService: UserService,
	}
}

// GetProfile 获取用户个人资料
func (h *AuthHandler) GetProfile(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 调用 Service层 获取 User
	user, err := h.UserService.GetUser(authenticatedUser.ID)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 转换为DTO
	userProfile := dto.ToUserProfileDTO(user)

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProfile, ""))
}

// UpdateProfile 更新用户个人资料
func (h *AuthHandler) UpdateProfile(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 解析请求体
	var payload dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid user update request", "requesterId", authenticatedUser.ID, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 更新 User
	err := h.UserService.UpdateUser(authenticatedUser.ID, &payload)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回204 No Content
	slog.Info("User Profile updated", "requesterId", authenticatedUser.ID)
	ctx.JSON(http.StatusNoContent, nil)
}

// UpdatePassword 更新密码
func (h *AuthHandler) UpdatePassword(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 解析请求体
	var payload dto.UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid password update request", "requesterId", authenticatedUser.ID, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 更新密码
	err := h.UserService.UpdatePassword(authenticatedUser.ID, payload.CurrentPassword, payload.NewPassword)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回204 No Content
	slog.Info("Password updated successfully", "requesterId", authenticatedUser.ID)
	ctx.JSON(http.StatusNoContent, nil)
}

// RegisterWithPassword 邮箱密码注册
func (h *AuthHandler) RegisterWithPassword(ctx *gin.Context) {
	// 解析请求体
	var payload dto.RegisterWithPasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid user registration request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 创建 User
	createdUserID, err := h.UserService.RegisterWithPassword(&payload)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回201 Created
	slog.Info("User created with password", "userId", createdUserID)
	ctx.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{"id": createdUserID}, ""))
}

// LoginWithPassword 邮箱密码登录
func (h *AuthHandler) LoginWithPassword(ctx *gin.Context) {
	// 解析请求体
	var payload dto.LoginWithPasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid login request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 进行认证并获取token
	accessToken, refreshToken, err := h.UserService.LoginWithPassword(payload.EmailOrPhone, payload.Password)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回token
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7, // Assuming 7 days expiration
	}, ""))
}

// RefreshToken 刷新访问令牌
func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	// 解析请求体
	var payload dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid refresh token request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 刷新token
	newAccessToken, newRefreshToken, err := h.UserService.RefreshAccessToken(payload.RefreshToken)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回新的token
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7, // 7天过期
	}, "Token refreshed successfully"))
}

// SendEmailVerification 发送邮箱验证码
func (h *AuthHandler) SendEmailVerification(ctx *gin.Context) {
	// 解析请求体
	var payload dto.SendEmailVerificationRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid email verification request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 发送验证码
	err := h.UserService.SendEmailVerification(payload.Email)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回成功响应
	slog.Info("Email verification sent", "email", payload.Email)
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Verification code sent successfully"))
}

// VerifyEmail 验证邮箱
func (h *AuthHandler) VerifyEmail(ctx *gin.Context) {
	// 解析请求体
	var payload dto.VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid verify email request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 验证邮箱
	err := h.UserService.VerifyEmail(payload.Email, payload.Code)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回成功响应
	slog.Info("Email verified successfully", "email", payload.Email)
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Email verified successfully"))
}

// SendPasswordReset 发送密码重置邮件
func (h *AuthHandler) SendPasswordReset(ctx *gin.Context) {
	// 解析请求体
	var payload dto.PasswordResetRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid password reset request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 发送密码重置邮件
	err := h.UserService.SendPasswordReset(payload.Email)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回成功响应
	slog.Info("Password reset email sent", "email", payload.Email)
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Password reset email sent successfully"))
}

// ResetPassword 重置密码
func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	// 解析请求体
	var payload dto.PasswordResetConfirmRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid password reset confirm request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 重置密码
	err := h.UserService.ResetPassword(payload.Email, payload.ResetToken, payload.NewPassword)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回成功响应
	slog.Info("Password reset successfully", "email", payload.Email)
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Password reset successfully"))
}

// RegisterFromWechatMiniProgram 微信小程序注册
func (h *AuthHandler) RegisterFromWechatMiniProgram(ctx *gin.Context) {
	// 获取并验证 OpenID 和 UnionID
	openID, unionID, ok := handler_utils.GetWechatIDs(ctx)
	if !ok {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Missing WeChat credentials"))
		return
	}

	// 解析请求体
	var payload dto.RegisterFromWechatMiniProgramRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid user creation request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 创建 User
	createdUserID, err := h.UserService.RegisterFromWechatMiniProgram(&payload, unionID, openID)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回201 Created
	slog.Info("User created", "userId", createdUserID)
	ctx.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{"id": createdUserID}, ""))
}

// LoginFromWechatMiniProgram 微信小程序端登录
func (h *AuthHandler) LoginFromWechatMiniProgram(ctx *gin.Context) {
	// Extract openID and unionID from header
	openID, unionID, ok := handler_utils.GetWechatIDs(ctx)
	if !ok {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Missing WeChat credentials"))
		return
	}

	// Call Service layer to authenticate and get token
	accessToken, refreshToken, err := h.UserService.LoginFromWechatMiniProgram(unionID, openID)

	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return token
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7, // Assuming 7 days expiration
	}, ""))
}

// ExchangeWechatOAuth 微信OAuth授权码交换（自动判断登录/注册）
func (h *AuthHandler) ExchangeWechatOAuth(ctx *gin.Context) {
	var payload dto.WechatOAuthRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid Google OAuth request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("参数错误: "+err.Error()))
		return
	}

	accessToken, refreshToken, isNewUser, err := h.UserService.ExchangeWechatOAuth(&payload)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	responseData := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7,
		"is_new_user":   isNewUser,
	}
	if isNewUser {
		ctx.JSON(http.StatusCreated, response.NewSuccessResponse(responseData, "用户注册并登录成功"))
	} else {
		ctx.JSON(http.StatusOK, response.NewSuccessResponse(responseData, "用户登录成功"))
	}
}

// BindWechatAccount 绑定微信账号
func (h *AuthHandler) BindWechatAccount(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 解析请求体
	var payload dto.BindWechatAccountRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid bind WeChat account request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 绑定微信账号
	err := h.UserService.BindWechatAccount(authenticatedUser.ID, &payload, authenticatedUser)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "WeChat account bound successfully"))
}

// UnbindWechatAccount 解绑微信账号
func (h *AuthHandler) UnbindWechatAccount(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 调用 Service层 解绑微信账号
	err := h.UserService.UnbindWechatAccount(authenticatedUser.ID, authenticatedUser)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "WeChat account unbound successfully"))
}

// ExchangeGoogleOAuth Google OAuth授权码交换（自动判断登录/注册）
func (h *AuthHandler) ExchangeGoogleOAuth(ctx *gin.Context) {
	// 解析请求体
	var payload dto.GoogleOAuthRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid Google OAuth request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 进行认证（自动判断登录/注册）
	accessToken, refreshToken, isNewUser, err := h.UserService.ExchangeGoogleOAuth(&payload)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回token和用户状态
	responseData := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7, // 7天过期
		"is_new_user":   isNewUser,     // 标识是否为新注册用户
	}

	// 根据是否为新用户返回不同的HTTP状态码
	if isNewUser {
		ctx.JSON(http.StatusCreated, response.NewSuccessResponse(responseData, "User registered and authenticated successfully"))
	} else {
		ctx.JSON(http.StatusOK, response.NewSuccessResponse(responseData, "User authenticated successfully"))
	}
}

// BindGoogleAccount 绑定Google账号
func (h *AuthHandler) BindGoogleAccount(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 解析请求体
	var payload dto.BindGoogleAccountRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid bind Google account request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 绑定Google账号
	err := h.UserService.BindGoogleAccount(authenticatedUser.ID, &payload, authenticatedUser)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Google account bound successfully"))
}

// UnbindGoogleAccount 解绑Google账号
func (h *AuthHandler) UnbindGoogleAccount(ctx *gin.Context) {
	// 获取当前authenticatedUser
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// 调用 Service层 解绑Google账号
	err := h.UserService.UnbindGoogleAccount(authenticatedUser.ID, authenticatedUser)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回成功响应
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Google account unbound successfully"))
}
