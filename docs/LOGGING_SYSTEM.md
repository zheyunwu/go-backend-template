# 日志系统

## 🚀 功能特性

### 相关文件

- `pkg/logger/logger.go`
- `internal/middlewares/logging_middleware.go`

### 自动化功能
1. **请求ID生成**: 自动为每个请求生成唯一ID
2. **上下文传递**: request_id自动注入到所有日志中
3. **用户关联**: 自动检测认证用户并添加到日志
4. **性能监控**: 自动记录请求时间和状态码
5. **错误分级**: 根据HTTP状态码自动调整日志级别

### 结构化日志输出
```json
{
  "time": "2025-06-15T14:23:37Z",
  "level": "INFO",
  "msg": "Request completed",
  "service": "go-backend-template",
  "version": "1.0.0",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123",
  "method": "GET",
  "path": "/api/v1/users/profile",
  "status_code": 200,
  "duration": "45ms",
  "client_ip": "192.168.1.1"
}
```

## 📋 使用指南

### 1. 基本设置
```go
// main.go
logger.Init(&logger.Config{
    Level:      slog.LevelInfo,
    JSONFormat: true,
    Output:     os.Stdout,
    Service:    "my-service",
    Version:    "1.0.0",
})
```

### 2. 中间件配置
```go
// routes.go
r.Use(middlewares.LoggingMiddleware()) // 必须在其他中间件之前
r.Use(middlewares.ErrorHandler())
```

### 3. 在代码中使用
```go
// Handler中
func UserHandler(c *gin.Context) {
    ctx := c.Request.Context()

    logger.Info(ctx, "Creating new user", "email", user.Email)

    if err := userService.Create(ctx, user); err != nil {
        logger.Error(ctx, "Failed to create user", "error", err)
        return
    }

    logger.Debug(ctx, "User created successfully", "user_id", user.ID)
}

// Service中
func (s *UserService) Create(ctx context.Context, user User) error {
    logger.Debug(ctx, "Validating user data")

    // 业务逻辑...

    logger.Info(ctx, "User validation completed")
    return nil
}
```
