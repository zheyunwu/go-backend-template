package middlewares

import (
	stderrors "errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/errors"
	"github.com/go-backend-template/internal/models"
	"github.com/go-backend-template/internal/services"
	"github.com/go-backend-template/pkg/jwt"
	"github.com/go-backend-template/pkg/response"
)

// RequiredAuthenticate middleware requires a valid authentication, otherwise returns 401
// RequiredAuthenticate 中间件：要求用户必须登录，否则返回401
func RequiredAuthenticate(config *config.Config, userService services.UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth, authenticated := authenticateWithJWT(ctx, config)

		if !authenticated {
			ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
			ctx.Abort()
			return
		}

		// Get User
		authenticatedUser, err := userService.GetUser(auth.UserID)
		if err != nil {
			if stderrors.Is(err, errors.ErrUserNotFound) {
				slog.Warn("Authenticated user not found", "userId", auth.UserID)
				ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
				ctx.Abort()
				return
			}
			// Log error and return internal server error
			slog.Error("Failed to get authenticated user", "userId", auth.UserID, "error", err)
			ctx.JSON(http.StatusInternalServerError, response.NewErrorResponse(errors.ErrInternalServer.Message))
			ctx.Abort()
			return
		}

		// Set authentication details in context
		setAuthContext(ctx, authenticatedUser)
		ctx.Next()
	}
}

// OptionalAuthenticate middleware attempts authentication but does not enforce it
// OptionalAuthenticate 中间件：尝试登录，但不强制要求
func OptionalAuthenticate(config *config.Config, userService services.UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if auth, authenticated := authenticateWithJWT(ctx, config); authenticated {
			// Get User
			authenticatedUser, err := userService.GetUser(auth.UserID)
			if err != nil {
				if stderrors.Is(err, errors.ErrUserNotFound) {
					slog.Warn("Authenticated user not found", "userId", auth.UserID)
					ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
					ctx.Abort()
					return
				}
				slog.Error("Failed to get authenticated user", "userId", auth.UserID, "error", err)
				ctx.JSON(http.StatusInternalServerError, response.NewErrorResponse(errors.ErrInternalServer.Message))
				ctx.Abort()
				return
			}
			setAuthContext(ctx, authenticatedUser)
		}
		ctx.Next()
	}
}

// AdminAuthMiddleware middleware requires a valid authentication with admin role
// AdminAuthMiddleware 中间件：要求用户必须登录且是管理员
func AdminAuthMiddleware(config *config.Config, userService services.UserService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// For admin authentication, we only support JWT
		auth, authenticated := authenticateWithJWT(ctx, config)

		if !authenticated {
			ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
			ctx.Abort()
			return
		}

		// Get User
		authenticatedUser, err := userService.GetUser(auth.UserID)
		if err != nil {
			if stderrors.Is(err, errors.ErrUserNotFound) {
				slog.Warn("Authenticated user not found", "userId", auth.UserID)
				ctx.JSON(http.StatusUnauthorized, response.NewErrorResponse(errors.ErrUnauthorized.Message))
				ctx.Abort()
				return
			}
			slog.Error("Failed to get authenticated user", "userId", auth.UserID, "error", err)
			ctx.JSON(http.StatusInternalServerError, response.NewErrorResponse(errors.ErrInternalServer.Message))
			ctx.Abort()
			return
		}

		// Check for admin role
		if authenticatedUser.Role != "admin" {
			slog.Warn("Access denied - not an admin user",
				"userId", authenticatedUser.ID,
				"role", authenticatedUser.Role,
			)
			ctx.JSON(http.StatusForbidden, response.NewErrorResponse(errors.ErrPermissionDenied.Message))
			ctx.Abort()
			return
		}

		// Set authentication details in context
		setAuthContext(ctx, authenticatedUser)
		ctx.Next()
	}
}

// authenticateWithJWT attempts to authenticate using JWT token from the Authorization header
func authenticateWithJWT(ctx *gin.Context, config *config.Config) (*models.UserAuthDetails, bool) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, false
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix
	tokenDetails, err := jwt.ValidateToken(tokenString, config.JWT.Secret)
	if err != nil || tokenDetails.TokenType != jwt.AccessToken {
		slog.Debug("JWT validation failed", "error", err)
		return nil, false
	}

	return &models.UserAuthDetails{
		UserID: tokenDetails.UserID,
		Role:   tokenDetails.Role,
	}, true
}

// setAuthContext sets user authentication details in the request context
func setAuthContext(ctx *gin.Context, authenticatedUser *models.User) {
	ctx.Set("authenticatedUser", authenticatedUser) // Store as pointer
}
