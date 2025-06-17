package di

import (
	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/handlers"
	"github.com/go-backend-template/internal/handlers/admin_handlers"
	"github.com/go-backend-template/internal/infra"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/internal/services"
	"github.com/openai/openai-go" // imported as openai
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Container is the dependency injection container, managing component dependencies.
type Container struct {
	// Configuration
	Config         *config.Config
	DB             *gorm.DB
	Redis          *redis.Client
	OpenAIClient   *openai.Client
	MoonshotClient *openai.Client
	DeepSeekClient *openai.Client

	// Service Layer (Core Services)
	EmailService        services.EmailService
	VerificationService services.VerificationService
	GoogleOAuthService  services.GoogleOAuthService

	// Repository Layer
	UserRepository            repositories.UserRepository
	CategoryRepository        repositories.CategoryRepository
	ProductRepository         repositories.ProductRepository
	UserInteractionRepository repositories.UserInteractionRepository

	// Service Layer (Business Services)
	UserService            services.UserService
	CategoryService        services.CategoryService
	ProductService         services.ProductService
	UserInteractionService services.UserInteractionService

	// Handler Layer
	AuthHandler            *handlers.AuthHandler
	CategoryHandler        *handlers.CategoryHandler
	ProductHandler         *handlers.ProductHandler
	UserInteractionHandler *handlers.UserInteractionHandler

	// Admin Handler Layer
	UserHandlerForAdmin    *admin_handlers.UserHandler
	ProductHandlerForAdmin *admin_handlers.ProductHandler
}

// NewContainer creates a new dependency injection container.
func NewContainer(env string) *Container {
	container := &Container{}

	// Load configuration.
	cfg, err := config.LoadConfig(env)
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}
	db := infra.InitDB(cfg)
	redis := infra.InitRedis(cfg)
	openaiClient := infra.InitOpenAIClient(cfg)
	moonshotClient := infra.InitMoonshotClient(cfg)
	deepSeekClient := infra.InitDeepSeekClient(cfg)

	// Set configuration and database connections.
	container.Config = cfg
	container.DB = db
	container.Redis = redis
	container.OpenAIClient = openaiClient
	container.MoonshotClient = moonshotClient
	container.DeepSeekClient = deepSeekClient

	// Initialize core services.
	container.EmailService = services.NewEmailService(cfg)
	container.VerificationService = services.NewVerificationService(redis)
	container.GoogleOAuthService = services.NewGoogleOAuthService(cfg)

	// Initialize repository layer.
	container.initRepositoryLayer(db)

	// Initialize service layer.
	container.initServiceLayer(cfg)

	// Initialize handler layer.
	container.initHandlerLayer()

	return container
}

// initRepositoryLayer initializes the repository layer.
func (c *Container) initRepositoryLayer(db *gorm.DB) {
	c.UserRepository = repositories.NewUserRepository(db)
	c.CategoryRepository = repositories.NewCategoryRepository(db)
	c.ProductRepository = repositories.NewProductRepository(db, c.CategoryRepository)
	c.UserInteractionRepository = repositories.NewUserInteractionRepository(db, c.CategoryRepository)
}

// initServiceLayer initializes the service layer.
func (c *Container) initServiceLayer(cfg *config.Config) {
	c.UserService = services.NewUserService(cfg, c.UserRepository, c.GoogleOAuthService, c.EmailService, c.VerificationService)
	c.CategoryService = services.NewCategoryService(c.CategoryRepository)
	c.ProductService = services.NewProductService(c.ProductRepository, c.CategoryRepository)
	c.UserInteractionService = services.NewUserInteractionService(c.UserInteractionRepository, c.ProductRepository)
}

// initHandlerLayer initializes the handler layer.
func (c *Container) initHandlerLayer() {
	// Public handlers
	c.AuthHandler = handlers.NewAuthHandler(c.UserService)
	c.CategoryHandler = handlers.NewCategoryHandler(c.CategoryService)
	c.ProductHandler = handlers.NewProductHandler(c.ProductService, c.UserInteractionService)
	c.UserInteractionHandler = handlers.NewUserInteractionHandler(c.UserInteractionService)

	// Admin handlers
	c.UserHandlerForAdmin = admin_handlers.NewUserHandler(c.UserService)
	c.ProductHandlerForAdmin = admin_handlers.NewProductHandler(c.ProductService)
}
