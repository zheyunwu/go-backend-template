# Go 后端模版

基于 **Golang + Gin + GORM** 的后端项目模版，支持 PostgreSQL 和 MySQL。

## 🚀 快速开始

### 环境要求
- Go 1.19+
- PostgreSQL 或 MySQL

### 本地开发

1. **启动数据库**
   ```bash
   # PostgreSQL
   docker run --name test-pg -e POSTGRES_USER=dbuser -e POSTGRES_PASSWORD=dbpassword -e POSTGRES_DB=database_name -p 5432:5432 -d postgres:14

   # 或 MySQL
   docker run --name test-mysql -e MYSQL_ROOT_PASSWORD=root_password -e MYSQL_USER=dbuser -e MYSQL_PASSWORD=dbpassword -e MYSQL_DATABASE=database_name -p 3306:3306 -d mysql:latest
   ```

2. **运行项目**
   ```bash
   # 数据库迁移
   go run cmd/*.go migrate

   # 启动服务器
   go run cmd/*.go server
   ```

服务器将运行在 http://localhost:8080

### 生产部署

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

# 运行容器
docker run -d \
  --name go-backend \
  -p 8080:8080 \
  -e DATABASE_HOST=your-db-host \
  -e DATABASE_USER=your-db-user \
  -e DATABASE_PASSWORD=your-db-password \
  -e DATABASE_NAME=your-db-name \
  -e JWT_SECRET=your-jwt-secret \
  go-backend-template

# 或使用 docker-compose
# 创建 docker-compose.yml 后运行：
docker-compose up -d
```

#### Docker Compose 示例
```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=prod
      - DATABASE_HOST=db
      - DATABASE_USER=dbuser
      - DATABASE_PASSWORD=dbpassword
      - DATABASE_NAME=database_name
      - JWT_SECRET=your-jwt-secret
    depends_on:
      - db

  db:
    image: postgres:14
    environment:
      - POSTGRES_USER=dbuser
      - POSTGRES_PASSWORD=dbpassword
      - POSTGRES_DB=database_name
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

## ⚙️ 配置

### 配置文件
- `config/config.dev.yaml` - 开发环境
- `config/config.prod.yaml` - 生产环境

### 环境变量覆盖
```bash
export APP_ENV=prod

export SERVER_PORT=8080

export DATABASE_HOST="mysql.db.host"
export DATABASE_PORT=3306
export DATABASE_USER="admin"
export DATABASE_PASSWORD="supersecret"
export DATABASE_NAME="deshop"

export JWT_SECRET="jwt-secret"
export JWT_EXPIRE_HOURS=72

export AI_OPENAI_API_KEY="your-api-key"
export AI_MOONSHOT_API_KEY="your-api-key"
export AI_DEEPSEEK_API_KEY="your-api-key"

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
│   ├── errors/                # 自定义错误
│   ├── routes/                # 路由定义
│   ├── infra/
│   │   ├── db.go              # DB 连接 & 初始化
│   │   ├── llm.go             # LLM Client初始化
│   ├── handlers/              # HTTP请求处理层
│   │   ├── admin_handlers/    # 面向后台管理的API Handlers
│   │   ├── hander_utils/      # Handler层公共逻辑
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
│   │   ├── request_logger.go  # 记录HTTP请求的生命周期
│   ├── utils/                 # 工具函数
├── pkg/                       # 公共库
├── sql/                       # SQL脚本
├── scripts/                   # 放一些脚本
├── docs/                      # API 文档
│   └── api.yaml               # OpenAPI 规范
└── Dockerfile
```

## 📋 API 规范

### List接口Query Parameters
支持以下查询参数：
- `page`, `limit` - 分页
- `search` - 搜索
- `filter` - 过滤 (JSON格式)
- `sort` - 排序 (格式: `field:asc|desc`)

示例：`GET /products?page=1&limit=10&search=laptop&filter={"barcode":"4337256850032","categories":[1]}&sort=updated_at:desc`

### 响应格式

**列表接口：**
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

**单条记录：**
```json
{
  "status": "success",
  "data": {...}
}
```

**错误响应：**
```json
{
  "status": "error",
  "message": "错误描述"
}
```

## Google OAuth2 登录（Authorization Code Flow with PKCE）

本后端支持 Google OAuth2 登录，采用推荐的 Authorization Code Flow with PKCE（适用于原生移动端和 SPA）。

### 配置

在配置文件中添加 Google OAuth2 凭证。iOS 和 Web 可分别配置不同的 client：

```yaml
google:
  ios:
    client_id: "your-ios-google-client-id"
    client_secret: "your-ios-google-client-secret"
    redirect_urls:
      - "com.yourapp.scheme://oauth/callback"  # iOS应用深链接
  web:
    client_id: "your-web-google-client-id"
    client_secret: "your-web-google-client-secret"
    redirect_urls:
      - "http://localhost:3000/auth/callback"  # 本地开发Web应用
      - "https://yourapp.com/auth/callback"    # 生产环境Web应用
```

### API 接口

#### 统一 Google OAuth2 Exchange
**POST** `/api/v1/auth/google/exchange`

自动判断登录/注册，仅需一个“Google 登录”按钮。

**请求体：**
```json
{
  "code": "google_authorization_code",
  "code_verifier": "pkce_code_verifier",
  "redirect_uri": "https://yourapp.com/auth/callback",
  "client_type": "web"
}
```

**参数说明：**
- `code`：Google OAuth2 授权码
- `code_verifier`：PKCE 校验参数
- `redirect_uri`：需与配置一致
- `client_type`：`"ios"` 或 `"web"`

**响应（200 已有用户登录）：**
```json
{
  "success": true,
  "data": {
    "access_token": "jwt_token_here",
    "token_type": "Bearer",
    "expires_in": 604800,
    "is_new_user": false
  },
  "message": "User authenticated successfully"
}
```

**响应（201 新用户注册）：**
```json
{
  "success": true,
  "data": {
    "access_token": "jwt_token_here",
    "token_type": "Bearer",
    "expires_in": 604800,
    "is_new_user": true
  },
  "message": "User registered and authenticated successfully"
}
```

### 客户端实现指引

#### 前端流程（PKCE）

1. **生成 PKCE 参数：**
```javascript
// 生成 code verifier（43-128位）
const codeVerifier = generateRandomString(128);

// 生成 code challenge
const codeChallenge = base64URLEncode(sha256(codeVerifier));
```

2. **跳转 Google 授权：**
```javascript
const authUrl = `https://accounts.google.com/o/oauth2/v2/auth?` +
  `client_id=${CLIENT_ID}&` +
  `redirect_uri=${encodeURIComponent(REDIRECT_URI)}&` +
  `response_type=code&` +
  `prompt=consent&` +
  `scope=openid email profile&` +
  `code_challenge=${codeChallenge}&` +
  `code_challenge_method=S256&` +
  `state=${generateRandomString(32)}`;

window.location.href = authUrl;
```

3. **回调后换取 JWT：**
```javascript
// 从回调URL提取 code
const urlParams = new URLSearchParams(window.location.search);
const code = urlParams.get('code');

// 向后端换取 JWT
const response = await fetch('/api/v1/auth/google/exchange', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    code: code,
    code_verifier: codeVerifier,
    redirect_uri: REDIRECT_URI,
    client_type: 'web'
  }),
});

const result = await response.json();
if (result.success) {
  // 存储 JWT token
  localStorage.setItem('access_token', result.data.access_token);
}
```

### 错误响应

常见错误：

```json
{
  "success": false,
  "message": "Invalid OAuth authorization code",
  "error": "invalid_oauth_code"
}
```

```json
{
  "success": false,
  "message": "Invalid redirect URL",
  "error": "invalid_redirect_url"
}
```

```json
{
  "success": false,
  "message": "Email already exists",
  "error": "email_already_exists"
}
```
