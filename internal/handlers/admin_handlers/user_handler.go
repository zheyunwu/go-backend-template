package admin_handlers

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/query_params"
	"github.com/go-backend-template/pkg/response"
)

type UserHandler struct {
	UserService services.UserService
}

func NewUserHandler(UserService services.UserService) *UserHandler {
	return &UserHandler{
		UserService: UserService,
	}
}

/*
5个通用CRUD接口
*/

// GetUsers 获取用户列表
func (h *UserHandler) ListUsers(ctx *gin.Context) {
	// 从上下文中获取已解析的查询参数
	params, _ := ctx.Get("queryParams")
	queryParams, ok := params.(*query_params.QueryParams)
	if !ok {
		slog.Warn("Invalid query parameters type", "params", params)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
		return
	}

	// 获取用户列表
	users, pagination, err := h.UserService.ListUsers(queryParams)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(users, "", *pagination))
}

// GetUser 获取单个用户详情
func (h *UserHandler) GetUser(ctx *gin.Context) {
	// 获取Path参数：user ID
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 调用 Service层 获取 User
	user, err := h.UserService.GetUser(uint(id))
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回200 OK
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(user, ""))
}

// CreateUser 创建新用户
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	// 解析请求体
	var payload dto.RegisterWithPasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid user creation request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 创建 User
	createdUserID, err := h.UserService.CreateUser(&payload)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回201 Created
	slog.Info("User created", "userId", createdUserID)
	ctx.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{"id": createdUserID}, ""))
}

// UpdateUser 更新用户信息
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	// 解析用户ID
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 解析请求体
	var payload dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		slog.Warn("Invalid user update request", "userId", id, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// 调用 Service层 更新 User
	err = h.UserService.UpdateUser(uint(id), &payload)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回204 No Content
	slog.Info("User updated", "userId", id)
	ctx.JSON(http.StatusNoContent, nil)
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	// 解析用户ID
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 调用 Service层 删除 User
	err = h.UserService.DeleteUser(uint(id))
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// 返回204 No Content
	slog.Info("User deleted", "userId", id)
	ctx.JSON(http.StatusNoContent, nil)
}

/*
定制接口
*/

// BanUser 封禁或解除封禁用户
func (h *UserHandler) BanUser(ctx *gin.Context) {
	// 解析用户ID
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// 解析请求体中是否封禁或解除封禁，默认封禁
	var payload struct {
		IsBanned bool `json:"is_banned"`
	}
	// 如果未传值，可默认封禁
	ctx.ShouldBindJSON(&payload)

	// 调用 Service 层封禁/解除封禁 User
	err = h.UserService.BanUser(uint(id), payload.IsBanned)
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	slog.Info("User ban status updated", "userId", id, "banned", payload.IsBanned)
	ctx.JSON(http.StatusNoContent, nil)
}
