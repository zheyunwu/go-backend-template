package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/di"
	"github.com/go-backend-template/internal/middlewares"
)

func InitRoutes(r *gin.Engine, container *di.Container) {
	// Use default CORS middleware, allows all origins, configure as needed.
	r.Use(cors.Default())

	// Global middlewares
	// RequestID should be one of the first middlewares.
	r.Use(middlewares.RequestID())
	// ContextLogger should run after RequestID to include request_id in logs.
	r.Use(middlewares.ContextLogger())
	// RequestLogger will now use the logger from context (which includes request_id).
	r.Use(middlewares.RequestLogger())
	r.Use(middlewares.ErrorHandler())
	r.Use(middlewares.QueryParamParser())

	// Initialize public API route group.
	api := r.Group("/api/v1")
	initRoutes(api, container)

	// Initialize Admin API route group.
	adminApi := r.Group("/admin-api/v1")
	initAdminRoutes(adminApi, container)
}

// Public API routes
func initRoutes(api *gin.RouterGroup, container *di.Container) {
	// Middlewares
	requiredAuthMiddleware := middlewares.RequiredAuthenticate(container.Config, container.UserService) // Must be logged in
	optionalAuthMiddleware := middlewares.OptionalAuthenticate(container.Config, container.UserService) // Optional login

	rateLimiter := middlewares.NewRateLimiter(container.Redis) // Rate limiter
	emailVerificationRateLimit := rateLimiter.EmailVerificationRateLimit()
	passwordResetRateLimit := rateLimiter.PasswordResetRateLimit()

	// Auth related routes
	authRoutes := api.Group("/auth")
	{
		authRoutes.GET("/profile", requiredAuthMiddleware, container.AuthHandler.GetProfile)        // Get user profile
		authRoutes.PATCH("/profile", requiredAuthMiddleware, container.AuthHandler.UpdateProfile)   // Update user profile
		authRoutes.PATCH("/password", requiredAuthMiddleware, container.AuthHandler.UpdatePassword) // Update password

		authRoutes.POST("/register", container.AuthHandler.RegisterWithPassword) // Register with password
		authRoutes.POST("/login", container.AuthHandler.LoginWithPassword)       // Login with password
		authRoutes.POST("/refresh", container.AuthHandler.RefreshToken)          // Refresh access token

		// Email verification related (with rate limiting)
		authRoutes.POST("/email/send-verification", emailVerificationRateLimit, container.AuthHandler.SendEmailVerification) // Send email verification code
		authRoutes.POST("/email/verify", container.AuthHandler.VerifyEmail)                                                  // Verify email

		// Password reset related (with rate limiting)
		authRoutes.POST("/password/reset-request", passwordResetRateLimit, container.AuthHandler.SendPasswordReset) // Send password reset email
		authRoutes.POST("/password/reset", container.AuthHandler.ResetPassword)                                     // Reset password

		// WeChat Mini Program
		authRoutes.POST("/wxmini/register", container.AuthHandler.RegisterFromWechatMiniProgram) // WeChat Mini Program registration
		authRoutes.POST("/wxmini/login", container.AuthHandler.LoginFromWechatMiniProgram)       // WeChat Mini Program login

		// OAuth2
		authRoutes.POST("/wechat/token", container.AuthHandler.ExchangeWechatOAuth)                            // WeChat login (Authorization Code Flow)
		authRoutes.POST("/wechat/bind", requiredAuthMiddleware, container.AuthHandler.BindWechatAccount)       // Bind WeChat account
		authRoutes.DELETE("/wechat/unbind", requiredAuthMiddleware, container.AuthHandler.UnbindWechatAccount) // Unbind WeChat account

		authRoutes.POST("/google/token", container.AuthHandler.ExchangeGoogleOAuth)                            // Google login (Authorization Code Flow with PKCE)
		authRoutes.POST("/google/bind", requiredAuthMiddleware, container.AuthHandler.BindGoogleAccount)       // Bind Google account
		authRoutes.DELETE("/google/unbind", requiredAuthMiddleware, container.AuthHandler.UnbindGoogleAccount) // Unbind Google account
	}

	// Product related routes
	productRoutes := api.Group("/products")
	{
		// List products - supports ?is_liked=true and ?is_favorited=true to filter liked/favorited products
		productRoutes.GET("", optionalAuthMiddleware, container.ProductHandler.ListProducts)
		productRoutes.GET("/:id", optionalAuthMiddleware, container.ProductHandler.GetProduct)

		// User interaction (like, favorite) related routes
		productRoutes.GET("/:id/stats", container.UserInteractionHandler.GetProductStats)                           // Get product statistics (like count, favorite count)
		productRoutes.PUT("/:id/like", requiredAuthMiddleware, container.UserInteractionHandler.ToggleLike)         // Like/unlike product
		productRoutes.PUT("/:id/favorite", requiredAuthMiddleware, container.UserInteractionHandler.ToggleFavorite) // Favorite/unfavorite product
	}

	// Category related routes
	categoryRoutes := api.Group("/categories")
	{
		categoryRoutes.GET("", container.CategoryHandler.ListCategories)       // Get category list
		categoryRoutes.GET("/tree", container.CategoryHandler.GetCategoryTree) // Get category tree structure
	}
}

// Admin API routes
func initAdminRoutes(admin *gin.RouterGroup, container *di.Container) {
	// Middlewares
	admin.Use(middlewares.AdminAuthMiddleware(container.Config, container.UserService)) // All admin routes require admin privileges

	// User management routes
	userRoutes := admin.Group("/users")
	{
		userRoutes.GET("", container.UserHandlerForAdmin.ListUsers)
		userRoutes.GET("/:id", container.UserHandlerForAdmin.GetUser)
		userRoutes.POST("", container.UserHandlerForAdmin.CreateUser)
		userRoutes.PATCH("/:id", container.UserHandlerForAdmin.UpdateUser)
		userRoutes.DELETE("/:id", container.UserHandlerForAdmin.DeleteUser)
		userRoutes.PATCH("/:id/restore", container.UserHandlerForAdmin.RestoreUser) // Restore soft-deleted user
		userRoutes.PATCH("/:id/ban", container.UserHandlerForAdmin.BanUser)
	}

	// Product management routes
	productRoutes := admin.Group("/products")
	{
		productRoutes.GET("", container.ProductHandlerForAdmin.ListProducts)
		productRoutes.GET("/:id", container.ProductHandlerForAdmin.GetProduct)
		productRoutes.POST("", container.ProductHandlerForAdmin.CreateProduct)
		productRoutes.PATCH("/:id", container.ProductHandlerForAdmin.UpdateProduct)
		productRoutes.DELETE("/:id", container.ProductHandlerForAdmin.DeleteProduct)
	}
}
