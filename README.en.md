# Go Backend Template

A backend project template based on **Golang + Gin + GORM**, supporting PostgreSQL and MySQL.

## 🚀 Quick Start

### Requirements
- Go 1.19+
- PostgreSQL or MySQL

### Local Development

1. **Start Database**
   ```bash
   # PostgreSQL
   docker run --name test-pg -e POSTGRES_USER=dbuser -e POSTGRES_PASSWORD=dbpassword -e POSTGRES_DB=database_name -p 5432:5432 -d postgres:14

   # Or MySQL
   docker run --name test-mysql -e MYSQL_ROOT_PASSWORD=root_password -e MYSQL_USER=dbuser -e MYSQL_PASSWORD=dbpassword -e MYSQL_DATABASE=database_name -p 3306:3306 -d mysql:latest
   ```

2. **Run the Project**
   ```bash
   # Database migration
   go run cmd/*.go migrate

   # Start server
   go run cmd/*.go server
   ```

Server will run at http://localhost:8080

### Production Deployment

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

# Run container
docker run -d \
  --name go-backend \
  -p 8080:8080 \
  -e DATABASE_HOST=your-db-host \
  -e DATABASE_USER=your-db-user \
  -e DATABASE_PASSWORD=your-db-password \
  -e DATABASE_NAME=your-db-name \
  -e JWT_SECRET=your-jwt-secret \
  go-backend-template

# Or use docker-compose
# After creating docker-compose.yml, run:
docker-compose up -d
```

#### Docker Compose Example
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

## ⚙️ Configuration

### Config Files
- `config/config.dev.yaml` - Development
- `config/config.prod.yaml` - Production

### Environment Variable Override
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

## 📁 Project Structure

```
go-backend-template/
├── cmd/                       # Entrypoint
│   ├── main.go                # Main (controls server/migrate)
│   ├── migrate.go             # DB migration
│   ├── server.go              # Start HTTP server
├── config/                    # Config
│   ├── config.dev.yaml
│   ├── config.prod.yaml
│   ├── config.go              # Config loader
├── internal/                  # Internal logic
│   ├── di/                    # DI container
│   ├── dto/                   # DTOs
│   ├── errors/                # Custom errors
│   ├── routes/                # Routes
│   ├── infra/
│   │   ├── db.go              # DB connection/init
│   │   ├── llm.go             # LLM Client init
│   ├── handlers/              # HTTP handlers
│   │   ├── admin_handlers/    # Admin API Handlers
│   │   ├── hander_utils/      # Handler utils
│   │   ├── xxx_handlers.go    # Common Handlers
│   ├── services/              # Services
│   ├── repositories/          # Data access
│   ├── models/                # Models
│   │   ├── user.go            # User table
│   │   ├── product.go         # Product table
│   │   ├── ...
│   ├── middlewares/           # Middlewares
│   │   ├── authenticate.go    # Auth (JWT -> user)
│   │   ├── error_handler.go   # Global error handler
│   │   ├── query_parser.go    # Query param parser
│   │   ├── request_logger.go  # HTTP request logger
│   ├── utils/                 # Utils
├── pkg/                       # Public libs
├── sql/                       # SQL scripts
├── scripts/                   # Scripts
├── docs/                      # API docs
│   └── api.yaml               # OpenAPI spec
└── Dockerfile
```

## 📋 API Spec

### List API Query Parameters
Supports:
- `page`, `limit` - Pagination
- `search` - Search
- `filter` - Filter (JSON)
- `sort` - Sort (format: `field:asc|desc`)

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

**Single Record:**
```json
{
  "status": "success",
  "data": {...}
}
```

**Error Response:**
```json
{
  "status": "error",
  "message": "Error description"
}
```

## Google OAuth2 Login (Authorization Code Flow with PKCE)

This backend supports Google OAuth2 login using the Authorization Code Flow with PKCE (Proof Key for Code Exchange), which is the recommended approach for native mobile apps and SPAs.

### Configuration

Add your Google OAuth2 credentials to your config file. You can configure different client credentials for iOS and Web applications:

```yaml
google:
  ios:
    client_id: "your-ios-google-client-id"
    client_secret: "your-ios-google-client-secret"
    redirect_urls:
      - "com.yourapp.scheme://oauth/callback"  # iOS app deep link
  web:
    client_id: "your-web-google-client-id"
    client_secret: "your-web-google-client-secret"
    redirect_urls:
      - "http://localhost:3000/auth/callback"  # Local development web app
      - "https://yourapp.com/auth/callback"    # Production web app
```

### API Endpoints

#### Unified Google OAuth2 Exchange
**POST** `/api/v1/auth/google/exchange`

Authenticate using Google OAuth2 with automatic login/registration detection. This is the recommended approach where clients only need one "Sign in with Google" button.

**Request Body:**
```json
{
  "code": "google_authorization_code",
  "code_verifier": "pkce_code_verifier",
  "redirect_uri": "https://yourapp.com/auth/callback",
  "client_type": "web"
}
```

**Parameters:**
- `code`: Google authorization code from OAuth2 flow
- `code_verifier`: PKCE code verifier for security
- `redirect_uri`: Must match one of the configured redirect URLs
- `client_type`: Either `"ios"` or `"web"` to specify which client configuration to use

**Response (200 OK - Existing User Login):**
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

**Response (201 Created - New User Registration):**
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

### Client Implementation Guide

#### Frontend Flow (PKCE)

1. **Generate PKCE Parameters:**
```javascript
// Generate code verifier (43-128 characters)
const codeVerifier = generateRandomString(128);

// Generate code challenge
const codeChallenge = base64URLEncode(sha256(codeVerifier));
```

2. **Redirect to Google Authorization:**
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

3. **Handle Callback and Exchange Code:**
```javascript
// Extract authorization code from callback URL
const urlParams = new URLSearchParams(window.location.search);
const code = urlParams.get('code');

// Exchange for JWT with the backend
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
  // Store JWT token
  localStorage.setItem('access_token', result.data.access_token);
}
```

### Error Responses

Common error responses:

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
