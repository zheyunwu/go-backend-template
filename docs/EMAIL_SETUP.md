# 邮件验证配置指南

本指南介绍如何在 Go 后端应用中快速配置邮件验证和密码重置功能。

## 🚀 快速开始

### 1. 环境要求

- Redis 服务器
- 邮件服务商（SendGrid 或 SMTP）

```bash
# 启动 Redis（Docker）
docker run --name redis -p 6379:6379 -d redis:alpine
```

### 2. 配置邮件服务

在 `config/config.dev.yaml` 中配置：

```yaml
# Redis 配置
redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

# 邮件配置
email:
  provider: "sendgrid"  # 推荐，或使用 "smtp"
  sendgrid:
    api_key: "your-sendgrid-api-key"
    from_email: "noreply@yourdomain.com"
    from_name: "Your App"
  smtp:  # Gmail 示例
    host: "smtp.gmail.com"
    port: 587
    username: "your-email@gmail.com"
    password: "your-app-password"
    from_email: "noreply@yourdomain.com"
    from_name: "Your App"
```

### 3. 测试验证

```bash
# 启动应用
go run cmd/*.go server

# 测试邮件验证
go run test_email_verification.go
```

## ⚙️ 邮件服务配置

### SendGrid（推荐）

1. 注册 [SendGrid](https://sendgrid.com) 免费账户
2. 创建 API Key：`Settings → API Keys → Create API Key`
3. 权限选择：`Mail Send`
4. 验证发件人邮箱或域名

### SMTP（Gmail 示例）

1. 开启两步验证
2. 生成应用专用密码：`Google 账户 → 安全性 → 应用专用密码`
3. 使用应用专用密码作为配置中的 password

## 📡 API 端点

### 邮箱验证

```http
POST /api/v1/auth/email/send-verification
Content-Type: application/json

{"email": "user@example.com"}
```

```http
POST /api/v1/auth/email/verify
Content-Type: application/json

{"email": "user@example.com", "code": "123456"}
```

### 密码重置

```http
POST /api/v1/auth/password/reset-request
Content-Type: application/json

{"email": "user@example.com"}
```

```http
POST /api/v1/auth/password/reset
Content-Type: application/json

{
  "email": "user@example.com",
  "reset_token": "abc12345",
  "new_password": "NewPassword123!"
}
```

## 🛡️ 安全特性

### 限流保护

| 功能 | 限制 | 窗口期 |
|------|------|--------|
| 邮箱验证 | 3次/邮箱 | 10分钟 |
| 密码重置 | 2次/邮箱 | 15分钟 |

### 验证码规则

| 类型 | 格式 | 有效期 | Redis Key |
|------|------|--------|-----------|
| 邮箱验证 | 6位数字 | 10分钟 | `email_verification:{email}` |
| 密码重置 | 8位字母数字 | 30分钟 | `password_reset:{email}` |

## 🔧 故障排除

### 常见问题

**Redis 连接失败**
```bash
# 检查 Redis 状态
redis-cli ping
```

**邮件发送失败**
- 检查 API Key 权限（SendGrid）
- 验证应用专用密码（Gmail）
- 确认发件人邮箱已验证

**验证码问题**
```bash
# 查看 Redis 中的验证码
redis-cli get "email_verification:user@example.com"

# 监控 Redis 操作
redis-cli monitor
```

### 环境变量覆盖

```bash
# 覆盖邮件配置
export EMAIL_PROVIDER="sendgrid"
export EMAIL_SENDGRID_API_KEY="your-api-key"
export EMAIL_SENDGRID_FROM_EMAIL="noreply@yourdomain.com"

# 覆盖 Redis 配置
export REDIS_HOST="localhost"
export REDIS_PORT=6379
```