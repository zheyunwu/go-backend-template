package handler_utils

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/pkg/response"
	"github.com/google/uuid"
)

// ParseUintParam 解析路径中的无符号整数参数
func ParseUintParam(ctx *gin.Context, paramName string) (uint64, error) {
	idStr := ctx.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		slog.Warn("Invalid parameter format", "param", paramName, "value", idStr, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid "+paramName+" parameter"))
		return 0, err
	}
	return id, nil
}

// ParseUUIDParam 解析路径中的UUID参数
func ParseUUIDParam(ctx *gin.Context, paramName string) (string, error) {
	uuidStr := ctx.Param(paramName)
	uid, err := uuid.Parse(uuidStr)
	if err != nil {
		slog.Warn("Invalid uuid parameter format", "param", paramName, "value", uuidStr, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Ivalid"+paramName+" parameter"))
		return "", err
	}
	return uid.String(), nil
}

// GetAuthenticatedUser 获取经过身份验证的用户信息
func GetAuthenticatedUser(ctx *gin.Context) (*models.User, bool) {
	authenticatedUser, exists := ctx.Get("authenticatedUser")
	if !exists {
		// 不存在没关系，返回空值和false
		return nil, false
	}

	// Handle user as pointer (*models.User) which is how it's stored in the context
	if user, ok := authenticatedUser.(*models.User); ok {
		return user, true
	}

	slog.Error("Invalid type for authenticatedUser")
	ctx.JSON(http.StatusInternalServerError, response.NewErrorResponse("Internal Server Error"))
	return nil, false
}

// GetWechatIDs 获取微信相关ID (OpenID, UnionID)
func GetWechatIDs(ctx *gin.Context) (*string, *string, bool) {
	openIDStr := ctx.GetHeader("x-wx-openid")
	unionIDStr := ctx.GetHeader("x-wx-unionid")

	// 都不存在的话返回 nil, nil, false
	if openIDStr == "" && unionIDStr == "" {
		return nil, nil, false
	}

	// 分别处理 openID 和 unionID
	var openID, unionID *string
	if openIDStr != "" {
		openID = &openIDStr
	}

	if unionIDStr != "" {
		unionID = &unionIDStr
	}

	return openID, unionID, true
}

// HandleError 统一处理错误响应
func HandleError(ctx *gin.Context, err error) {
	// 处理 AppError 类型的错误
	if appError, ok := err.(*errors.AppError); ok {
		ctx.JSON(appError.Status, response.NewErrorResponse(appError.Message))
		return
	}

	// 处理未知错误类型
	ctx.JSON(http.StatusInternalServerError, response.NewErrorResponse(err.Error()))
}
