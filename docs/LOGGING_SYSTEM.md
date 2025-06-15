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

#### Handlers和Services中的日志使用

**✅ 推荐做法 - 使用context-aware logger：**
```go
// Handler中
func (h *AuthHandler) UpdateProfile(ctx *gin.Context) {
    // 使用统一的logger，会自动包含request_id和user_id
    logger.Info(ctx, "Processing profile update request")

    if err := validateRequest(); err != nil {
        logger.Warn(ctx, "Validation failed", "error", err)
        return
    }

    logger.Info(ctx, "Profile updated successfully", "userId", userID)
}

// Service中
func (s *UserService) UpdateUser(ctx context.Context, userID uint, data *UpdateData) error {
    logger.Debug(ctx, "Starting user update", "userId", userID)

    if err := s.repository.Update(ctx, userID, data); err != nil {
        logger.Error(ctx, "Failed to update user", "userId", userID, "error", err)
        return err
    }

    logger.Info(ctx, "User updated successfully", "userId", userID)
    return nil
}

// Repository中
func (r *userRepository) Update(ctx context.Context, userID uint, data *UpdateData) error {
    logger.Debug(ctx, "Executing user update query", "userId", userID)

    result := r.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(data)
    if result.Error != nil {
        logger.Error(ctx, "Database update failed", "userId", userID, "error", result.Error)
        return result.Error
    }

    return nil
}
```

**❌ 避免使用 - 直接使用slog：**
```go
// 这样会丢失request_id和user_id等上下文信息
slog.Info("Profile updated", "userId", userID)
slog.Warn("Validation failed", "error", err)
```

#### 不同场景的最佳实践

1. **Handlers**: 使用 `logger.Info(ctx, ...)` 等方法
2. **Services**: 使用 `logger.Info(ctx, ...)` 等方法
3. **Repositories**: 使用 `logger.Debug(ctx, ...)` 进行调试日志
4. **Middlewares**: 使用 `logger.Error(ctx, ...)` 等方法
5. **启动/关闭代码**: 可以直接使用 `slog.Info()` (没有请求上下文)
6. **工具包/解析器**: 可以使用 `slog.Warn()` (低级别工具函数)

#### 日志级别使用指南

```go
// Debug: 详细的调试信息
logger.Debug(ctx, "Processing step completed", "step", "validation")

// Info: 正常的业务流程信息
logger.Info(ctx, "User registered successfully", "userId", user.ID)

// Warn: 警告信息，不影响正常流程
logger.Warn(ctx, "Invalid parameter provided", "param", invalidParam)

// Error: 错误信息，影响正常流程
logger.Error(ctx, "Database operation failed", "error", err)
```
