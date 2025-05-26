package di

import (
	"github.com/go-backend-template/config"
	"github.com/go-backend-template/internal/handlers"
	"github.com/go-backend-template/internal/handlers/admin_handlers"
	"github.com/go-backend-template/internal/infra"
	"github.com/go-backend-template/internal/repositories"
	"github.com/go-backend-template/internal/services"
	"github.com/openai/openai-go" // imported as openai
	"gorm.io/gorm"
)

// Container 依赖注入容器，统一管理组件依赖
type Container struct {
	// 配置
	Config         *config.Config
	DB             *gorm.DB
	OpenAIClient   *openai.Client
	MoonshotClient *openai.Client
	DeepSeekClient *openai.Client

	// Repository层
	UserRepository            repositories.UserRepository
	CategoryRepository        repositories.CategoryRepository
	ProductRepository         repositories.ProductRepository
	UserInteractionRepository repositories.UserInteractionRepository

	// Service层
	UserService            services.UserService
	CategoryService        services.CategoryService
	ProductService         services.ProductService
	UserInteractionService services.UserInteractionService

	// Handler层
	AuthHandler            *handlers.AuthHandler
	CategoryHandler        *handlers.CategoryHandler
	ProductHandler         *handlers.ProductHandler
	UserInteractionHandler *handlers.UserInteractionHandler

	// Admin Handler层
	UserHandlerForAdmin    *admin_handlers.UserHandler
	ProductHandlerForAdmin *admin_handlers.ProductHandler
}

// NewContainer 创建一个新的依赖注入容器
func NewContainer(env string) *Container {
	container := &Container{}

	// 加载配置
	cfg, err := config.LoadConfig(env)
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}
	db := infra.InitDB(cfg)
	openaiClient := infra.InitOpenAIClient(cfg)
	moonshotClient := infra.InitMoonshotClient(cfg)
	deepSeekClient := infra.InitDeepSeekClient(cfg)

	// 设置配置和数据库连接
	container.Config = cfg
	container.DB = db
	container.OpenAIClient = openaiClient
	container.MoonshotClient = moonshotClient
	container.DeepSeekClient = deepSeekClient

	// 初始化仓库层
	container.initRepositoryLayer(db)

	// 初始化服务层
	container.initServiceLayer(cfg)

	// 初始化处理器层
	container.initHandlerLayer()

	return container
}

// 初始化Repository层
func (c *Container) initRepositoryLayer(db *gorm.DB) {
	c.UserRepository = repositories.NewUserRepository(db)
	c.CategoryRepository = repositories.NewCategoryRepository(db)
	c.ProductRepository = repositories.NewProductRepository(db, c.CategoryRepository)
	c.UserInteractionRepository = repositories.NewUserInteractionRepository(db, c.CategoryRepository)
}

// 初始化Service层
func (c *Container) initServiceLayer(cfg *config.Config) {
	c.UserService = services.NewUserService(cfg, c.UserRepository)
	c.CategoryService = services.NewCategoryService(c.CategoryRepository)
	c.ProductService = services.NewProductService(c.ProductRepository, c.CategoryRepository)
	c.UserInteractionService = services.NewUserInteractionService(c.UserInteractionRepository, c.ProductRepository)
}

// 初始化Handler层
func (c *Container) initHandlerLayer() {
	// 公共处理器
	c.AuthHandler = handlers.NewAuthHandler(c.UserService)
	c.CategoryHandler = handlers.NewCategoryHandler(c.CategoryService)
	c.ProductHandler = handlers.NewProductHandler(c.ProductService, c.UserInteractionService)
	c.UserInteractionHandler = handlers.NewUserInteractionHandler(c.UserInteractionService)

	// Admin处理器
	c.UserHandlerForAdmin = admin_handlers.NewUserHandler(c.UserService)
	c.ProductHandlerForAdmin = admin_handlers.NewProductHandler(c.ProductService)
}
