package admin_handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/logger"
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
5 general CRUD interfaces
*/

// ListUsers retrieves a list of users.
func (h *UserHandler) ListUsers(ctx *gin.Context) {
	// Get parsed query parameters from context.
	params, _ := ctx.Get("queryParams")
	queryParams, ok := params.(*query_params.QueryParams)
	if !ok {
		logger.Warn(ctx, "Invalid query parameters type", "params", params)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid query parameters type"))
		return
	}

	// Check if soft-deleted records should be included.
	includeSoftDeleted := ctx.Query("include_soft_deleted") == "true"

	// Get user list.
	users, pagination, err := h.UserService.ListUsers(ctx.Request.Context(), queryParams, includeSoftDeleted) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(users, "", *pagination))
}

// GetUser retrieves details for a single user.
func (h *UserHandler) GetUser(ctx *gin.Context) {
	// Get user ID from path parameters.
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Check if soft-deleted records should be included.
	includeSoftDeleted := ctx.Query("include_soft_deleted") == "true"

	// Call service layer to get the user.
	user, err := h.UserService.GetUser(ctx.Request.Context(), uint(id), includeSoftDeleted) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(user, ""))
}

// CreateUser creates a new user.
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	// Parse request body.
	var payload dto.RegisterWithPasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid user creation request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Call service layer to create the user.
	createdUserID, err := h.UserService.CreateUser(ctx.Request.Context(), &payload) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 201 Created.
	logger.Info(ctx, "User created", "userId", createdUserID)
	ctx.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{"id": createdUserID}, ""))
}

// UpdateUser updates user information.
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	// Parse user ID.
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Parse request body.
	var payload dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid user update request", "userId", id, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Call service layer to update the user.
	err = h.UserService.UpdateUser(ctx.Request.Context(), uint(id), &payload) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 204 No Content.
	logger.Info(ctx, "User updated", "userId", id)
	ctx.JSON(http.StatusNoContent, nil)
}

// DeleteUser deletes a user.
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	// Parse user ID.
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Call service layer to delete the user.
	err = h.UserService.DeleteUser(ctx.Request.Context(), uint(id)) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 204 No Content.
	logger.Info(ctx, "User deleted", "userId", id)
	ctx.JSON(http.StatusNoContent, nil)
}

// RestoreUser restores a soft-deleted user.
func (h *UserHandler) RestoreUser(ctx *gin.Context) {
	// Parse user ID.
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Call service layer to restore the soft-deleted user.
	err = h.UserService.RestoreUser(ctx.Request.Context(), uint(id)) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	logger.Info(ctx, "User restored", "userId", id)
	ctx.JSON(http.StatusNoContent, nil)
}

/*
Custom interfaces
*/

// BanUser bans or unbans a user.
func (h *UserHandler) BanUser(ctx *gin.Context) {
	// Parse user ID.
	id, err := handler_utils.ParseUintParam(ctx, "id")
	if err != nil {
		return
	}

	// Parse whether to ban or unban from the request body, defaults to ban.
	var payload struct {
		IsBanned bool `json:"is_banned"`
	}
	// If no value is provided, it can default to ban (or handle as needed).
	ctx.ShouldBindJSON(&payload)

	// Call service layer to ban/unban the user.
	err = h.UserService.BanUser(ctx.Request.Context(), uint(id), payload.IsBanned) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	logger.Info(ctx, "User ban status updated", "userId", id, "banned", payload.IsBanned)
	ctx.JSON(http.StatusNoContent, nil)
}
