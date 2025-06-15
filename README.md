# Go Backend Template

A backend project template based on **Golang + Gin + GORM**, supporting PostgreSQL, MySQL, Redis, and integrated with email verification and OAuth2 login.

## ✨ Key Features

- 🎯 Layered Architecture Design
- 🔐 JWT Authentication System
- 📧 Email Verification (SendGrid/SMTP)
- 🔑 OAuth2 Login (Google, WeChat)
- 🔗 Account Binding and Unbinding (Google, WeChat)
- 📊 Redis Cache and Rate Limiting
- 📝 CRUD Operation Examples
- 🐳 Docker Support
- 🧪 Unit and Integration Testing Setup
- 📜 Request ID and Context-aware Logging for Traceability
- ⚡ Input Validation for DTOs
- ⚙️ Context Propagation throughout services and repositories

## 🚀 Quick Start

### Prerequisites
- Go 1.19+
- PostgreSQL or MySQL
- Redis (required for email verification and rate limiting)

### Local Development

1. **Start Database**
   ```bash
   # PostgreSQL
   docker run --name test-pg -e POSTGRES_USER=dbuser -e POSTGRES_PASSWORD=dbpassword -e POSTGRES_DB=database_name -p 5432:5432 -d postgres:14

   # Or MySQL
   docker run --name test-mysql -e MYSQL_ROOT_PASSWORD=root_password -e MYSQL_USER=dbuser -e MYSQL_PASSWORD=dbpassword -e MYSQL_DATABASE=database_name -p 3306:3306 -d mysql:latest
   ```

2. **Start Redis** (required for email verification and rate limiting)
   ```bash
   # Using Docker
   docker run --name test-redis -p 6379:6379 -d redis:alpine
   ```

3. **Run the Project**
   ```bash
   # Install dependencies
   go mod tidy

   # Database migration
   go run cmd/*.go migrate

   # Start server
   go run cmd/*.go server
   ```

The server will run at http://localhost:8080

### Production Environment

Prepare DB and Redis services, fill in the corresponding configurations in [`config/config.prod.yaml`](config/config.prod.yaml), then:

#### Method 1: Direct Run
```bash
# Set environment
export APP_ENV=prod

# Database migration
APP_ENV=prod go run cmd/*.go migrate

# Start server
APP_ENV=prod go run cmd/*.go server
```

#### Method 2: Docker Deployment
```bash
# Build image
docker build -t go-backend-template .

# Run container (using config.prod.yaml from within the container)
docker run -d \
  --name go-backend \
  -p 8080:8080 \
  -e APP_ENV=prod \
  go-backend-template

# Run container (overriding config file with ENV variables, see "Configuration" section for details)
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

## ⚙️ Configuration

### Configuration Files

The configuration file read by the program is controlled by the `APP_ENV` environment variable (defaults to `dev` if not set):

- `export APP_ENV=dev` -> `config/config.dev.yaml`
- `export APP_ENV=prod` -> `config/config.prod.yaml`

### Environment Variable Override

Values in `config/config.<ENV>.yaml` can be overridden by environment variables. The naming convention for environment variables is to connect hierarchical names with underscores. For example:

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

# Two exceptions: variables related to WeChat Cloud Base Object Storage service
export COS_BUCKET="bucket_name"
export COS_REGION="ap-shanghai"
```

## 📁 Project Structure

```
go-backend-template/
├── cmd/                       # Application entry points
│   ├── main.go                # Main entry file (controls server/migrate)
│   ├── migrate.go             # Runs database migrations
│   ├── server.go              # Starts the HTTP server
├── config/                    # Configuration
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   ├── config.go              # Config loading program
├── internal/                  # Internal application logic
│   ├── di/                    # Dependency Injection container
│   ├── dto/                   # Data Transfer Object definitions
│   ├── errors/                # Custom Errors
│   ├── routes/                # Route definitions
│   ├── infra/                 # Infrastructure (DB, Redis, external clients)
│   │   ├── db.go              # DB Client connection & initialization
│   │   ├── redis.go           # Redis Client connection & initialization
│   │   ├── llm.go             # LLM Client initialization
│   ├── handlers/              # HTTP request handling layer
│   │   ├── admin_handlers/    # API Handlers for admin panel
│   │   ├── handler_utils/     # Common logic for Handlers
│   │   ├── xxx_handlers.go    # Public Handlers
│   ├── services/              # Business logic layer
│   ├── repositories/          # Data access layer
│   │   ├── mocks/             # Mock implementations for repositories (for testing)
│   ├── models/                # Database Models
│   │   ├── user.go            # User table
│   │   ├── product.go         # Product table
│   │   ├── ...
│   ├── middlewares/           # Middlewares
│   │   ├── authenticate.go    # Authentication (JWT -> user)
│   │   ├── context_logger.go  # Injects request-scoped logger into context
│   │   ├── error_handler.go   # Global error handling
│   │   ├── query_parser.go    # Parses query parameters
│   │   ├── rate_limiter.go    # Rate limiting
│   │   ├── request_id.go      # Injects X-Request-ID
│   │   ├── request_logger.go  # Logs HTTP request lifecycle
│   ├── utils/                 # Utility functions
│   ├── tests/                 # Integration test setup and helpers
├── pkg/                       # Public libraries/utilities shared across projects
├── sql/                       # SQL scripts
├── scripts/                   # Various scripts
├── docs/                      # Detailed documentation
│   ├── EMAIL_SETUP.md         # Email verification setup
│   └── OAUTH_INTEGRATION.md   # OAuth2 integration guide
└── Dockerfile
```

## 📋 API Specification

### Query Parameters
List APIs uniformly support the following query parameters via the [QueryParamParser Middleware](internal/middlewares/query_parser.go):
- `page`, `limit` - Pagination
- `search` - Search
- `filter` - Filtering (JSON format)
- `sort` - Sorting (Format: `field:asc|desc`)

Example: `GET /products?page=1&limit=10&search=laptop&filter={"barcode":"4337256850032","categories":[1]}&sort=updated_at:desc`

### Response Format

**List API:**
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

**Get API:**
```json
{
  "status": "success",
  "data": {...}
}
```

**Error Response Format:**
```json
{
  "status": "error",
  "message": "Error description"
}
```

## 🔑 OAuth2 Login

Supports Google and WeChat OAuth2 login using the secure Authorization Code Flow with PKCE.

**Main API Endpoints:**
- `POST /api/v1/auth/google/exchange` - Google OAuth2 login/registration
- `POST /api/v1/auth/wechat/exchange` - WeChat OAuth2 login/registration

**Basic Flow:**
1. Frontend guides user through OAuth2 authorization.
2. Frontend calls the backend exchange endpoint with the authorization code.
3. Backend returns JWT tokens to complete the login.

**Detailed Integration Guide:** [OAuth2 Integration Document](docs/OAUTH_INTEGRATION.md)

## 📧 Email System

Supports integration with email systems for verification and other notifications.

**Detailed Integration Guide:** [Email Setup Document](docs/EMAIL_SETUP.md)
