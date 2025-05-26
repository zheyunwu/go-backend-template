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

// UserExists 检查用户是否存在
func (h *AuthHandler) CheckUserExists(ctx *gin.Context) {
	// 获取请求参数
	fieldType := ctx.Query("field_type")
	if fieldType == "" {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("field_type is required"))
		return
	}

	// fieldType 如果是 "mini_program_open_id" 或 "unionid" 则需要从请求头中获取
	var value string
	if fieldType == "mini_program_open_id" {
		value = ctx.GetHeader("x-wx-openid")
	} else if fieldType == "union_id" {
		value = ctx.GetHeader("x-wx-unionid")
	} else {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid field_type: "+fieldType))
		return
	}

	// 调用 Service层 检查用户是否存在
	exists, err := h.UserService.CheckUserExists(fieldType, value)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{"exists": exists}, ""))
}

// GetProfileByUserID 根据UserID查询用户信息
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

// UpdateProfile 更新用户信息
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
	err := h.UserService.UpdateProfile(authenticatedUser.ID, &payload, authenticatedUser)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回204 No Content
	slog.Info("User updated", "requesterId", authenticatedUser.ID)
	ctx.JSON(http.StatusNoContent, nil)
}

// RegisterWithPassword 使用密码注册用户
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

// LoginWithPassword 用户密码登录
func (h *AuthHandler) LoginWithPassword(ctx *gin.Context) {
	// 解析请求体
	var payload dto.LoginWithPasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid login request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 进行认证并获取token
	token, err := h.UserService.LoginWithPassword(payload.EmailOrPhone, payload.Password)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回token
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600 * 24 * 7, // Assuming 7 days expiration
	}, ""))
}

// RegisterFromWechatMiniProgram 使用微信小程序注册用户
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

// LoginFromWechatMiniProgram handles login from WeChat Mini Program
func (h *AuthHandler) LoginFromWechatMiniProgram(ctx *gin.Context) {
	// Extract openID and unionID from header
	openID, unionID, ok := handler_utils.GetWechatIDs(ctx)
	if !ok {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Missing WeChat credentials"))
		return
	}

	// Call Service layer to authenticate and get token
	token, err := h.UserService.LoginFromWechatMiniProgram(unionID, openID)

	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return token
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600 * 24 * 7, // Assuming 7 days expiration
	}, ""))
}
