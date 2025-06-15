package middlewares

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/jwt"
	"github.com/go-backend-template/pkg/logger"
	"github.com/go-backend-template/pkg/response"
)

// RequiredAuthenticate middleware requires a valid authentication, otherwise returns 401.
// It ensures that the user must be logged in.
func RequiredAuthenticate(config *config.Config, userService services.UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth, authenticated := authenticateWithJWT(ctx, config)

		if !authenticated {
			ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
			ctx.Abort()
			return
		}

		// Get User details.
		authenticatedUser, err := userService.GetUser(ctx.Request.Context(), auth.UserID) // Pass context
		if err != nil {
			if stderrors.Is(err, errors.ErrUserNotFound) {
				logger.Warn(ctx.Request.Context(), "Authenticated user not found", "userId", auth.UserID) // Pass context
				ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
				ctx.Abort()
				return
			}
			// Log error and return internal server error.
			logger.Error(ctx.Request.Context(), "Failed to get authenticated user", "userId", auth.UserID, "error", err) // Pass context
			ctx.JSON(http.StatusInternalServerError, response.NewErrorResponse(errors.ErrInternalServer.Message))
			ctx.Abort()
			return
		}

		// Set authentication details in context.
		setAuthContext(ctx, authenticatedUser)
		ctx.Next()
	}
}

// OptionalAuthenticate middleware attempts authentication but does not enforce it.
// It tries to log in the user but proceeds even if authentication fails.
func OptionalAuthenticate(config *config.Config, userService services.UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if auth, authenticated := authenticateWithJWT(ctx, config); authenticated {
			// Get User details.
			authenticatedUser, err := userService.GetUser(ctx.Request.Context(), auth.UserID) // Pass context
			if err != nil {
				if stderrors.Is(err, errors.ErrUserNotFound) {
					logger.Warn(ctx.Request.Context(), "Authenticated user not found for optional auth", "userId", auth.UserID) // Pass context
					// Do not abort here, just don't set the user in context.
				} else {
					logger.Error(ctx.Request.Context(), "Failed to get authenticated user for optional auth", "userId", auth.UserID, "error", err) // Pass context
					// Do not abort here, proceed without authenticated user.
				}
			} else {
				setAuthContext(ctx, authenticatedUser)
			}
		}
		ctx.Next()
	}
}

// AdminAuthMiddleware middleware requires a valid authentication with admin role.
// It ensures that the user is logged in and is an administrator.
func AdminAuthMiddleware(config *config.Config, userService services.UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// For admin authentication, we only support JWT.
		auth, authenticated := authenticateWithJWT(ctx, config)

		if !authenticated {
			ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
			ctx.Abort()
			return
		}

		// Get User details.
		authenticatedUser, err := userService.GetUser(ctx.Request.Context(), auth.UserID) // Pass context
		if err != nil {
			if stderrors.Is(err, errors.ErrUserNotFound) {
				logger.Warn(ctx.Request.Context(), "Authenticated admin user not found", "userId", auth.UserID) // Pass context
				ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
				ctx.Abort()
				return
			}
			logger.Error(ctx.Request.Context(), "Failed to get authenticated admin user", "userId", auth.UserID, "error", err) // Pass context
			ctx.JSON(http.StatusInternalServerError, response.NewErrorResponse(errors.ErrInternalServer.Message))
			ctx.Abort()
			return
		}

		// Check for admin role.
		if authenticatedUser.Role != "admin" {
			logger.Warn(ctx.Request.Context(), "Access denied - not an admin user", // Pass context
				"userId", authenticatedUser.ID,
				"role", authenticatedUser.Role,
			)
			ctx.JSON(http.StatusForbidden, response.NewErrorResponse(errors.ErrPermissionDenied.Message))
			ctx.Abort()
			return
		}

		// Set authentication details in context.
		setAuthContext(ctx, authenticatedUser)
		ctx.Next()
	}
}

// authenticateWithJWT attempts to authenticate using JWT token from the Authorization header.
func authenticateWithJWT(ctx *gin.Context, config *config.Config) (*models.UserAuthDetails, bool) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, false
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix.
	tokenDetails, err := jwt.ValidateToken(tokenString, config.JWT.Secret)
	if err != nil || tokenDetails.TokenType != jwt.AccessToken {
		logger.Debug(ctx.Request.Context(), "JWT validation failed", "error", err) // Pass context
		return nil, false
	}

	return &models.UserAuthDetails{
		UserID: tokenDetails.UserID,
		Role:   tokenDetails.Role,
	}, true
}

// setAuthContext sets user authentication details in the request context.
func setAuthContext(ctx *gin.Context, authenticatedUser *models.User) {
	ctx.Set("authenticatedUser", authenticatedUser) // Store as pointer.

	// 同时将user_id添加到logger context中
	updatedCtx := logger.WithUserID(ctx.Request.Context(), fmt.Sprintf("%d", authenticatedUser.ID))
	ctx.Request = ctx.Request.WithContext(updatedCtx)
}
