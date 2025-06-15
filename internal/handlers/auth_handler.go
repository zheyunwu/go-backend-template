package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/dto"
	"github.com/go-backend-template/internal/handlers/handler_utils"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/internal/utils" // Added validator utility
	"github.com/go-backend-template/pkg/logger"
	"github.com/go-backend-template/pkg/response"
)

var customValidator = utils.NewCustomValidator() // Create a validator instance

type AuthHandler struct {
	UserService services.UserService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(UserService services.UserService) *AuthHandler {
	return &AuthHandler{
		UserService: UserService,
	}
}

// GetProfile retrieves the user's profile.
func (h *AuthHandler) GetProfile(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Call service layer to get user.
	user, err := h.UserService.GetUser(ctx.Request.Context(), authenticatedUser.ID) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Convert to DTO.
	userProfile := dto.ToUserProfileDTO(user)

	// Return 200 OK.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(userProfile, ""))
}

// UpdateProfile updates the user's profile.
func (h *AuthHandler) UpdateProfile(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Parse request body.
	var payload dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid user update request", "requesterId", authenticatedUser.ID, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for UpdateProfile", "requesterId", authenticatedUser.ID, "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to update user.
	err := h.UserService.UpdateUser(ctx.Request.Context(), authenticatedUser.ID, &payload) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 204 No Content.
	logger.Info(ctx, "User Profile updated", "requesterId", authenticatedUser.ID)
	ctx.JSON(http.StatusNoContent, nil)
}

// UpdatePassword updates the user's password.
func (h *AuthHandler) UpdatePassword(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Parse request body.
	var payload dto.UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid password update request", "requesterId", authenticatedUser.ID, "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for UpdatePassword", "requesterId", authenticatedUser.ID, "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to update password.
	err := h.UserService.UpdatePassword(ctx.Request.Context(), authenticatedUser.ID, payload.CurrentPassword, payload.NewPassword) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 204 No Content.
	logger.Info(ctx, "Password updated successfully", "requesterId", authenticatedUser.ID)
	ctx.JSON(http.StatusNoContent, nil)
}

// RegisterWithPassword handles registration with email/password.
func (h *AuthHandler) RegisterWithPassword(ctx *gin.Context) {
	// Parse request body.
	var payload dto.RegisterWithPasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid user registration request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for RegisterWithPassword", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to create user.
	createdUserID, err := h.UserService.RegisterWithPassword(ctx.Request.Context(), &payload) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 201 Created.
	logger.Info(ctx, "User created with password", "userId", createdUserID)
	ctx.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{"id": createdUserID}, ""))
}

// LoginWithPassword handles login with email/password.
func (h *AuthHandler) LoginWithPassword(ctx *gin.Context) {
	// Parse request body.
	var payload dto.LoginWithPasswordRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid login request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for LoginWithPassword", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to authenticate and get tokens.
	accessToken, refreshToken, err := h.UserService.LoginWithPassword(ctx.Request.Context(), payload.EmailOrPhone, payload.Password) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return tokens.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7, // Assuming 7 days expiration
	}, ""))
}

// RefreshToken refreshes an access token.
func (h *AuthHandler) RefreshToken(ctx *gin.Context) {
	// Parse request body.
	var payload dto.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid refresh token request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for RefreshToken", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to refresh token.
	newAccessToken, newRefreshToken, err := h.UserService.RefreshAccessToken(ctx.Request.Context(), payload.RefreshToken) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return new tokens.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7, // 7 days expiration
	}, "Token refreshed successfully"))
}

// SendEmailVerification sends an email verification code.
func (h *AuthHandler) SendEmailVerification(ctx *gin.Context) {
	// Parse request body.
	var payload dto.SendEmailVerificationRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid email verification request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for SendEmailVerification", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to send verification code.
	err := h.UserService.SendEmailVerification(ctx.Request.Context(), payload.Email) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return success response.
	logger.Info(ctx, "Email verification sent", "email", payload.Email)
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Verification code sent successfully"))
}

// VerifyEmail verifies an email.
func (h *AuthHandler) VerifyEmail(ctx *gin.Context) {
	// Parse request body.
	var payload dto.VerifyEmailRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid verify email request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for VerifyEmail", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to verify email.
	err := h.UserService.VerifyEmail(ctx.Request.Context(), payload.Email, payload.Code) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return success response.
	logger.Info(ctx, "Email verified successfully", "email", payload.Email)
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Email verified successfully"))
}

// SendPasswordReset sends a password reset email.
func (h *AuthHandler) SendPasswordReset(ctx *gin.Context) {
	// Parse request body.
	var payload dto.PasswordResetRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid password reset request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for SendPasswordReset", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to send password reset email.
	err := h.UserService.SendPasswordReset(ctx.Request.Context(), payload.Email) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return success response.
	logger.Info(ctx, "Password reset email sent", "email", payload.Email)
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Password reset email sent successfully"))
}

// ResetPassword resets the user's password.
func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	// Parse request body.
	var payload dto.PasswordResetConfirmRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid password reset confirm request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for ResetPassword", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to reset password.
	err := h.UserService.ResetPassword(ctx.Request.Context(), payload.Email, payload.ResetToken, payload.NewPassword) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return success response.
	logger.Info(ctx, "Password reset successfully", "email", payload.Email)
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Password reset successfully"))
}

// RegisterFromWechatMiniProgram handles registration from WeChat Mini Program.
func (h *AuthHandler) RegisterFromWechatMiniProgram(ctx *gin.Context) {
	// Get and validate OpenID and UnionID.
	openID, unionID, ok := handler_utils.GetWechatIDs(ctx)
	if !ok {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Missing WeChat credentials"))
		return
	}

	// Parse request body.
	var payload dto.RegisterFromWechatMiniProgramRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid user creation request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate UpdateProfileRequest part of the payload as RegisterFromWechatMiniProgramRequest embeds it.
	if validationErrs := customValidator.ValidateStruct(&payload.UpdateProfileRequest); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for RegisterFromWechatMiniProgram (ProfileData)", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to create user.
	createdUserID, err := h.UserService.RegisterFromWechatMiniProgram(ctx.Request.Context(), &payload, unionID, openID) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return 201 Created.
	logger.Info(ctx, "User created", "userId", createdUserID)
	ctx.JSON(http.StatusCreated, response.NewSuccessResponse(gin.H{"id": createdUserID}, ""))
}

// LoginFromWechatMiniProgram handles login from WeChat Mini Program.
func (h *AuthHandler) LoginFromWechatMiniProgram(ctx *gin.Context) {
	// Extract openID and unionID from header.
	openID, unionID, ok := handler_utils.GetWechatIDs(ctx)
	if !ok {
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Missing WeChat credentials"))
		return
	}

	// Call Service layer to authenticate and get token.
	accessToken, refreshToken, err := h.UserService.LoginFromWechatMiniProgram(ctx.Request.Context(), unionID, openID) // Pass context

	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return token.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7, // Assuming 7 days expiration
	}, ""))
}

// ExchangeWechatOAuth handles WeChat OAuth code exchange (auto determines login/registration).
func (h *AuthHandler) ExchangeWechatOAuth(ctx *gin.Context) {
	var payload dto.WechatOAuthRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid Wechat OAuth request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid parameters: "+err.Error())) // "参数错误: " -> "Invalid parameters: "
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for ExchangeWechatOAuth", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	accessToken, refreshToken, isNewUser, err := h.UserService.ExchangeWechatOAuth(ctx.Request.Context(), &payload) // Pass context
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
		ctx.JSON(http.StatusCreated, response.NewSuccessResponse(responseData, "User registered and logged in successfully")) // "用户注册并登录成功" -> "User registered and logged in successfully"
	} else {
		ctx.JSON(http.StatusOK, response.NewSuccessResponse(responseData, "User logged in successfully")) // "用户登录成功" -> "User logged in successfully"
	}
}

// BindWechatAccount binds a WeChat account.
func (h *AuthHandler) BindWechatAccount(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Parse request body.
	var payload dto.BindWechatAccountRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid bind WeChat account request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for BindWechatAccount", "requesterId", authenticatedUser.ID, "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to bind WeChat account.
	err := h.UserService.BindWechatAccount(ctx.Request.Context(), authenticatedUser.ID, &payload, authenticatedUser) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return success response.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "WeChat account bound successfully"))
}

// UnbindWechatAccount unbinds a WeChat account.
func (h *AuthHandler) UnbindWechatAccount(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Call service layer to unbind WeChat account.
	err := h.UserService.UnbindWechatAccount(ctx.Request.Context(), authenticatedUser.ID, authenticatedUser) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return success response.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "WeChat account unbound successfully"))
}

// ExchangeGoogleOAuth handles Google OAuth code exchange (auto determines login/registration).
func (h *AuthHandler) ExchangeGoogleOAuth(ctx *gin.Context) {
	// Parse request body.
	var payload dto.GoogleOAuthRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid Google OAuth request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for ExchangeGoogleOAuth", "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to authenticate (auto determines login/registration).
	accessToken, refreshToken, isNewUser, err := h.UserService.ExchangeGoogleOAuth(ctx.Request.Context(), &payload) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return token and user status.
	responseData := gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600 * 24 * 7, // 7 days expiration
		"is_new_user":   isNewUser,     // Indicates if it's a newly registered user
	}

	// Return different HTTP status codes based on whether it's a new user.
	if isNewUser {
		ctx.JSON(http.StatusCreated, response.NewSuccessResponse(responseData, "User registered and authenticated successfully"))
	} else {
		ctx.JSON(http.StatusOK, response.NewSuccessResponse(responseData, "User authenticated successfully"))
	}
}

// BindGoogleAccount binds a Google account.
func (h *AuthHandler) BindGoogleAccount(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Parse request body.
	var payload dto.BindGoogleAccountRequest
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		logger.Warn(ctx, "Invalid bind Google account request", "error", err)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse("Invalid request body: "+err.Error()))
		return
	}

	// Validate payload.
	if validationErrs := customValidator.ValidateStruct(&payload); validationErrs != nil {
		logger.Warn(ctx, "Validation failed for BindGoogleAccount", "requesterId", authenticatedUser.ID, "errors", validationErrs)
		ctx.JSON(http.StatusBadRequest, response.NewErrorResponse(utils.FormatValidationErrors(validationErrs)))
		return
	}

	// Call service layer to bind Google account.
	err := h.UserService.BindGoogleAccount(ctx.Request.Context(), authenticatedUser.ID, &payload, authenticatedUser) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return success response.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Google account bound successfully"))
}

// UnbindGoogleAccount unbinds a Google account.
func (h *AuthHandler) UnbindGoogleAccount(ctx *gin.Context) {
	// Get current authenticated user.
	authenticatedUser, ok := handler_utils.GetAuthenticatedUser(ctx)
	if !ok {
		return
	}

	// Call service layer to unbind Google account.
	err := h.UserService.UnbindGoogleAccount(ctx.Request.Context(), authenticatedUser.ID, authenticatedUser) // Pass context
	if err != nil {
		handler_utils.HandleError(ctx, err)
		return
	}

	// Return success response.
	ctx.JSON(http.StatusOK, response.NewSuccessResponse(nil, "Google account unbound successfully"))
}
