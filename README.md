# Go 后端模版

基于 **Golang + Gin + GORM** 的后端项目模版，支持 PostgreSQL、MySQL、Redis，集成邮件验证和OAuth2登录。

## ✨ 主要功能

- 🎯 分层架构设计
- 🔐 JWT 认证系统
- 📧 邮件验证功能（SendGrid/SMTP）
- 🔑 OAuth2 登录（Google、微信）
- 🔗 账号绑定和解绑（Google、微信）
- 📊 Redis 缓存和限流
- 📝 CRUD操作示例
- 🐳 Docker 支持

## 🚀 快速开始

### 环境要求
- Go 1.19+
- PostgreSQL 或 MySQL
- Redis（邮件验证功能需要）

### 本地开发

1. **启动数据库**
   ```bash
   # PostgreSQL
   docker run --name test-pg -e POSTGRES_USER=dbuser -e POSTGRES_PASSWORD=dbpassword -e POSTGRES_DB=database_name -p 5432:5432 -d postgres:14

   # 或 MySQL
   docker run --name test-mysql -e MYSQL_ROOT_PASSWORD=root_password -e MYSQL_USER=dbuser -e MYSQL_PASSWORD=dbpassword -e MYSQL_DATABASE=database_name -p 3306:3306 -d mysql:latest
   ```

2. **启动Redis**（邮件验证功能需要）
   ```bash
   # Docker方式
   docker run --name test-redis -p 6379:6379 -d redis:alpine
   ```

3. **运行项目**
   ```bash
   # 安装依赖
   go mod tidy

   # 数据库迁移
   go run cmd/*.go migrate

   # 启动服务器
   go run cmd/*.go server
   ```

服务器将运行在 http://localhost:8080

### 生产环境

准备好DB、Redis服务，在[`config/config.prod.yaml`](config/config.prod.yaml)里填入对应配置，然后：

#### 方式一：直接运行
```bash
# 设置环境
export APP_ENV=prod

# 数据库迁移
APP_ENV=prod go run cmd/*.go migrate

# 启动服务器
APP_ENV=prod go run cmd/*.go server
```

#### 方式二：Docker 部署
```bash
# 构建镜像
docker build -t go-backend-template .

# 运行容器（直接用容器内的config.prod.yaml）
docker run -d \
  --name go-backend \
  -p 8080:8080 \
  -e APP_ENV=prod \
  go-backend-template

# 运行容器（用ENV覆盖config文件，具体说明请参考"配置"部分）
docker run -d \
  --name go-backend \
  -p 8080:8080 \
  -e APP_ENV=prod \
  -e DATABASE_HOST=your-db-host \
  -e DATABASE_USER=your-db-user \
  -e DATABASE_PASSWORD=your-db-password \
  -e DATABASE_NAME=your-db-name \
  -e REDIS_HOST=your-redis-host \
  -e REDIS_PASSWORD=your-redis-password \
  -e JWT_SECRET=your-jwt-secret \
  go-backend-template
```

## ⚙️ 配置

### 配置文件

根据环境变量`APP_ENV`控制程序读取的配置文件（如无该变量，则默认读取dev）：

- `export APP_ENV=dev` -> `config/config.dev.yaml`
- `export APP_ENV=prod` -> `config/config.prod.yaml`

### 环境变量覆盖

`config/config.<ENV>.yaml`里的值可以被环境变量覆盖。环境变量名规则：用下划线连接不同层级的名称。例如：

```bash
# server:
#   port: 8080
export SERVER_PORT=8080

# database:
#   driver: postgres
#   host: localhost
#   port: 5432
#   user: dbuser
#   password: dbpassword
#   name: database_name
export DATABASE_HOST="mysql.db.host"
export DATABASE_PORT=3306
export DATABASE_USER="admin"
export DATABASE_PASSWORD="supersecret"
export DATABASE_NAME="deshop"

# jwt:
#   secret: "super-secret-prod-key"
#   expire_hours: 72
export JWT_SECRET="jwt-secret"
export JWT_EXPIRE_HOURS=72

# redis:
#   host: "localhost"
#   port: 6379
#   password: ""
#   db: 0
export REDIS_HOST="localhost"
export REDIS_PORT=6379
export REDIS_PASSWORD="redis-pwd"
export REDIS_DB=0

# 两个例外：微信云托管对象存储服务的相关变量
export COS_BUCKET="bucket_name"
export COS_REGION="ap-shanghai"
```

## 📁 项目结构

```
go-backend-template/
├── cmd/                       # 应用程序入口
│   ├── main.go                # 入口文件（控制 server/migrate）
│   ├── migrate.go             # 运行数据库迁移
│   ├── server.go              # 启动 HTTP 服务器
├── config/                    # 配置
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   ├── config.go              # 载入Config程序
├── internal/                  # 应用内部逻辑
│   ├── di/                    # 依赖注入容器
│   ├── dto/                   # DTO定义
│   ├── errors/                # 自定义Errors
│   ├── routes/                # 路由定义
│   ├── infra/
│   │   ├── db.go              # DB Client 连接 & 初始化
│   │   ├── redis.go           # Redis Client 连接 & 初始化
│   │   ├── llm.go             # LLM Client 初始化
│   ├── handlers/              # HTTP请求处理层
│   │   ├── admin_handlers/    # 面向后台管理的API Handlers
│   │   ├── handler_utils/     # Handler层公共逻辑
│   │   ├── xxx_handlers.go    # 公共Handlers
│   ├── services/              # 业务逻辑层
│   ├── repositories/          # 数据访问层
│   ├── models/                # Models
│   │   ├── user.go            # 用户表
│   │   ├── product.go         # 商品表
│   │   ├── ...
│   ├── middlewares/           # 中间件
│   │   ├── authenticate.go    # 鉴权（JWT -> user）
│   │   ├── error_handler.go   # 全局错误处理
│   │   ├── query_parser.go    # 解析查询参数
│   │   ├── rate_limiter.go    # 速率限制
│   │   ├── request_logger.go  # 记录HTTP请求的生命周期
│   ├── utils/                 # 工具函数
├── pkg/                       # 公共库
├── sql/                       # SQL脚本
├── scripts/                   # 放一些脚本
├── docs/                      # 详细文档
│   ├── EMAIL_SETUP.md         # 邮件验证设置
│   └── OAUTH_INTEGRATION.md   # OAuth2 集成指南
└── Dockerfile
```

## 📋 API 规范

### 查询参数 Query Parameters
List接口通过[QueryParamParser 中间件](internal/middlewares/query_parser.go)统一支持以下查询参数：
- `page`, `limit` - 分页
- `search` - 搜索
- `filter` - 过滤 (JSON格式)
- `sort` - 排序 (格式: `field:asc|desc`)

示例：`GET /products?page=1&limit=10&search=laptop&filter={"barcode":"4337256850032","categories":[1]}&sort=updated_at:desc`

### 响应格式

**List接口：**
```json
{
  "status": "success",
  "data": [...],
  "pagination": {
    "total_count": 100,
    "page_size": 10,
    "current_page": 1,
    "total_pages": 10
  }
}
```

**Get接口：**
```json
{
  "status": "success",
  "data": {...}
}
```

**错误响应格式：**
```json
{
  "status": "error",
  "message": "错误描述"
}
```

## 🔑 OAuth2 登录

支持 Google 和微信 OAuth2 登录，采用安全的 Authorization Code Flow with PKCE。

**主要 API 接口：**
- `POST /api/v1/auth/google/exchange` - Google OAuth2 登录/注册
- `POST /api/v1/auth/wechat/exchange` - 微信 OAuth2 登录/注册

**基本流程：**
1. 前端引导用户完成 OAuth2 授权
2. 使用授权码调用后端 exchange 接口
3. 后端返回 JWT token 完成登录

**详细集成指南：** [OAuth2 集成文档](docs/OAUTH_INTEGRATION.md)

## 📧 邮件系统

支持集成邮件系统

**详细集成指南：** [邮件设置文档](docs/EMAIL_SETUP.md)
