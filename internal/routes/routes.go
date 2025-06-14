package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-backend-template/internal/di"
	"github.com/go-backend-template/internal/middlewares"
)

func InitRoutes(r *gin.Engine, container *di.Container) {
	// 使用默认的CORS中间件，允许所有来源，按需配置
	r.Use(cors.Default())

	// 全局中间件
	r.Use(middlewares.RequestLogger())
	r.Use(middlewares.ErrorHandler())
	r.Use(middlewares.QueryParamParser())

	// 初始化公共API路由组
	api := r.Group("/api/v1")
	initRoutes(api, container)

	// 初始化Admin API路由组
	adminApi := r.Group("/admin-api/v1")
	initAdminRoutes(adminApi, container)
}

// 公共API路由
func initRoutes(api *gin.RouterGroup, container *di.Container) {
	// 中间件
	requiredAuthMiddleware := middlewares.RequiredAuthenticate(container.Config, container.UserService) // 必须登录
	optionalAuthMiddleware := middlewares.OptionalAuthenticate(container.Config, container.UserService) // 可选登录

	rateLimiter := middlewares.NewRateLimiter(container.Redis) // 速率限制
	emailVerificationRateLimit := rateLimiter.EmailVerificationRateLimit()
	passwordResetRateLimit := rateLimiter.PasswordResetRateLimit()

	// Auth相关路由
	authRoutes := api.Group("/auth")
	{
		authRoutes.GET("/profile", requiredAuthMiddleware, container.AuthHandler.GetProfile)        // 获取用户个人资料
		authRoutes.PATCH("/profile", requiredAuthMiddleware, container.AuthHandler.UpdateProfile)   // 更新用户个人资料
		authRoutes.PATCH("/password", requiredAuthMiddleware, container.AuthHandler.UpdatePassword) // 更新密码

		authRoutes.POST("/register", container.AuthHandler.RegisterWithPassword) // 使用密码注册
		authRoutes.POST("/login", container.AuthHandler.LoginWithPassword)       // 使用密码登录
		authRoutes.POST("/refresh", container.AuthHandler.RefreshToken)          // 刷新访问令牌

		// 邮箱验证相关 (with rate limiting)
		authRoutes.POST("/email/send-verification", emailVerificationRateLimit, container.AuthHandler.SendEmailVerification) // 发送邮箱验证码
		authRoutes.POST("/email/verify", container.AuthHandler.VerifyEmail)                                                  // 验证邮箱

		// 密码重置相关 (with rate limiting)
		authRoutes.POST("/password/reset-request", passwordResetRateLimit, container.AuthHandler.SendPasswordReset) // 发送密码重置邮件
		authRoutes.POST("/password/reset", container.AuthHandler.ResetPassword)                                     // 重置密码

		// 微信小程序端
		authRoutes.POST("/wxmini/register", container.AuthHandler.RegisterFromWechatMiniProgram) // 微信小程序注册
		authRoutes.POST("/wxmini/login", container.AuthHandler.LoginFromWechatMiniProgram)       // 微信小程序登录

		// OAuth2
		authRoutes.POST("/wechat/token", container.AuthHandler.ExchangeWechatOAuth)                            // 微信登录（Authorization Code Flow）
		authRoutes.POST("/wechat/bind", requiredAuthMiddleware, container.AuthHandler.BindWechatAccount)       // 绑定微信账号
		authRoutes.DELETE("/wechat/unbind", requiredAuthMiddleware, container.AuthHandler.UnbindWechatAccount) // 解绑微信账号

		authRoutes.POST("/google/token", container.AuthHandler.ExchangeGoogleOAuth)                            // Google登录（Authorization Code Flow with PKCE）
		authRoutes.POST("/google/bind", requiredAuthMiddleware, container.AuthHandler.BindGoogleAccount)       // 绑定Google账号
		authRoutes.DELETE("/google/unbind", requiredAuthMiddleware, container.AuthHandler.UnbindGoogleAccount) // 解绑Google账号
	}

	// 产品相关路由
	productRoutes := api.Group("/products")
	{
		// 查询产品列表 - 支持?is_liked=true和?is_favorited=true筛选已点赞和已收藏的产品
		productRoutes.GET("", optionalAuthMiddleware, container.ProductHandler.ListProducts)
		productRoutes.GET("/:id", optionalAuthMiddleware, container.ProductHandler.GetProduct)

		// 用户交互（点赞、收藏）相关路由
		productRoutes.GET("/:id/stats", container.UserInteractionHandler.GetProductStats)                           // 获取产品统计信息(点赞数、收藏数)
		productRoutes.PUT("/:id/like", requiredAuthMiddleware, container.UserInteractionHandler.ToggleLike)         // 点赞/取消点赞产品
		productRoutes.PUT("/:id/favorite", requiredAuthMiddleware, container.UserInteractionHandler.ToggleFavorite) // 收藏/取消收藏产品
	}

	// 分类相关路由
	categoryRoutes := api.Group("/categories")
	{
		categoryRoutes.GET("", container.CategoryHandler.ListCategories)       // 获取分类列表
		categoryRoutes.GET("/tree", container.CategoryHandler.GetCategoryTree) // 获取分类树结构
	}
}

// Admin API路由
func initAdminRoutes(admin *gin.RouterGroup, container *di.Container) {
	// 中间件
	admin.Use(middlewares.AdminAuthMiddleware(container.Config, container.UserService)) // 所有后台路由都需要管理员权限

	// 用户管理路由
	userRoutes := admin.Group("/users")
	{
		userRoutes.GET("", container.UserHandlerForAdmin.ListUsers)
		userRoutes.GET("/:id", container.UserHandlerForAdmin.GetUser)
		userRoutes.POST("", container.UserHandlerForAdmin.CreateUser)
		userRoutes.PATCH("/:id", container.UserHandlerForAdmin.UpdateUser)
		userRoutes.DELETE("/:id", container.UserHandlerForAdmin.DeleteUser)
		userRoutes.PATCH("/:id/restore", container.UserHandlerForAdmin.RestoreUser) // 恢复软删除的用户
		userRoutes.PATCH("/:id/ban", container.UserHandlerForAdmin.BanUser)
	}

	// 产品管理路由
	productRoutes := admin.Group("/products")
	{
		productRoutes.GET("", container.ProductHandlerForAdmin.ListProducts)
		productRoutes.GET("/:id", container.ProductHandlerForAdmin.GetProduct)
		productRoutes.POST("", container.ProductHandlerForAdmin.CreateProduct)
		productRoutes.PATCH("/:id", container.ProductHandlerForAdmin.UpdateProduct)
		productRoutes.DELETE("/:id", container.ProductHandlerForAdmin.DeleteProduct)
	}
}
