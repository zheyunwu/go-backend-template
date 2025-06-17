# Email Verification Configuration Guide

This guide explains how to quickly configure email verification and password reset features in the Go backend application.

## 🚀 Quick Start

### 1. Prerequisites

- Redis server
- Email service provider (SendGrid or SMTP)

```bash
# Start Redis (Docker)
docker run --name redis -p 6379:6379 -d redis:alpine
```

### 2. Configure Email Service

Configure in `config/config.dev.yaml`:

```yaml
# Redis Configuration
redis:
  host: "localhost"
  port: 6379
  password: "" # No password by default for local Redis
  db: 0         # Default DB

# Email Configuration
email:
  provider: "sendgrid"  # Recommended, or use "smtp"
  sendgrid:
    api_key: "your-sendgrid-api-key"
    from_email: "noreply@yourdomain.com"
    from_name: "Your App Name" # Name displayed as sender
  smtp:  # Example for Gmail
    host: "smtp.gmail.com"
    port: 587
    username: "your-email@gmail.com"
    password: "your-app-password" # Use an app-specific password for Gmail
    from_email: "noreply@yourdomain.com"
    from_name: "Your App Name"
```

### 3. Test Verification

```bash
# Start the application
go run cmd/*.go server

# Test email verification (Example script or API call)
# (Assuming you have a test script or use a tool like Postman)
# go run test_email_verification.go
```
(Note: `test_email_verification.go` is a placeholder for your testing method.)

## ⚙️ Email Service Configuration

### SendGrid (Recommended)

1. Register for a free [SendGrid](https://sendgrid.com) account.
2. Create an API Key: `Settings → API Keys → Create API Key`.
3. Choose permissions: `Mail Send`.
4. Verify sender email or domain.

### SMTP (Gmail Example)

1. Enable 2-Step Verification for your Google account.
2. Generate an App Password: `Google Account → Security → App passwords`.
3. Use the App Password as the `password` in the SMTP configuration.

## 📡 API Endpoints

### Email Verification

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

### Password Reset

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

## 🛡️ Security Features

### Rate Limiting

| Feature             | Limit         | Window Period |
|---------------------|---------------|---------------|
| Email Verification  | 3 times/email | 10 minutes    |
| Password Reset      | 2 times/email | 15 minutes    |

### Verification Code Rules

| Type                | Format              | Validity Period | Redis Key                   |
|---------------------|---------------------|-----------------|-----------------------------|
| Email Verification  | 6-digit number      | 10 minutes      | `email_verification:{email}` |
| Password Reset      | 8-char alphanumeric | 30 minutes      | `password_reset:{email}`    |

## 🔧 Troubleshooting

### Common Issues

**Redis Connection Failed**
```bash
# Check Redis status
redis-cli ping
```

**Email Sending Failed**
- Check API Key permissions (SendGrid).
- Verify App Password (Gmail).
- Confirm sender email is verified.

**Verification Code Issues**
```bash
# View verification code in Redis
redis-cli get "email_verification:user@example.com"

# Monitor Redis operations
redis-cli monitor
```

### Environment Variable Override

```bash
# Override email configuration
export EMAIL_PROVIDER="sendgrid"
export EMAIL_SENDGRID_API_KEY="your-api-key"
export EMAIL_SENDGRID_FROM_EMAIL="noreply@yourdomain.com"
export EMAIL_SENDGRID_FROM_NAME="Your App Name"


# Override Redis configuration
export REDIS_HOST="localhost"
export REDIS_PORT=6379
# export REDIS_PASSWORD="" # If you set one
# export REDIS_DB=0
```