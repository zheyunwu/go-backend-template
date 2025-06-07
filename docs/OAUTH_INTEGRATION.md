# OAuth2 集成指南

此后端支持 Google 和微信提供商的 OAuth2 身份验证，使用适合不同客户端类型的安全授权流程。

## Google OAuth2 (授权码流程 + PKCE)

### 配置

将您的 Google OAuth2 凭据添加到配置文件中。您可以为 iOS 和 Web 应用程序配置不同的客户端凭据：

```yaml
google:
  ios:
    client_id: "your-ios-google-client-id"
    client_secret: "your-ios-google-client-secret"
    redirect_urls:
      - "com.yourapp.scheme://oauth/callback"  # iOS 应用深度链接
  web:
    client_id: "your-web-google-client-id"
    client_secret: "your-web-google-client-secret"
    redirect_urls:
      - "http://localhost:3000/auth/callback"  # 本地开发 Web 应用
      - "https://yourapp.com/auth/callback"    # 生产环境 Web 应用
```

### API 端点

- Auth Code换JWT
  ```
  POST /api/v1/auth/google/exchange
  Content-Type: application/json

  {
    "code": "google_authorization_code",
    "code_verifier": "pkce_code_verifier",
    "redirect_uri": "https://yourapp.com/auth/callback",
    "client_type": "web"
  }
  ```

  **参数解释：**
  - `code`: 来自 OAuth2 流程的 Google 授权码
  - `code_verifier`: 用于安全性的 PKCE 代码验证器
  - `redirect_uri`: 必须匹配配置的重定向 URL 之一
  - `client_type`: `"ios"` 或 `"web"`

  **响应 (200 OK - 现有用户登录)：**
  ```json
  {
    "success": true,
    "data": {
      "access_token": "jwt_token_here",
      "token_type": "Bearer",
      "expires_in": 604800,
      "is_new_user": false
    },
    "message": "用户认证成功"
  }
  ```

  **响应 (201 Created - 新用户注册)：**
  ```json
  {
    "success": true,
    "data": {
      "access_token": "jwt_token_here",
      "token_type": "Bearer",
      "expires_in": 604800,
      "is_new_user": true
    },
    "message": "用户注册并认证成功"
  }
  ```

### Google Login 客户端实现指南 (PKCE)

1. **生成 PKCE 参数：**
    ```javascript
    // 生成codeVerifier (43-128 字符)
    const codeVerifier = generateRandomString(128);

    // 生成codeChallenge
    const codeChallenge = base64URLEncode(sha256(codeVerifier));
    ```

2. **重定向到 Google 授权：**
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

3. **处理回调并交换代码：**
    ```javascript
    // 从回调 URL 中提取授权码
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');

    // 与后端交换 JWT
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
      // 存储 JWT 令牌
      localStorage.setItem('access_token', result.data.access_token);
    }
    ```

## 微信 OAuth2

### 配置

```yaml
wechat:
  web:
    app_id: "your-web-wechat-app-id"
    secret: "your-web-wechat-secret"
  app:
    app_id: "your-app-wechat-app-id"
    secret: "your-app-wechat-secret"
```

### API 端点

- Auth Code换JWT
  ```
  POST /api/v1/auth/wechat/exchange
  Content-Type: application/json

  {
    "code": "wechat_authorization_code",
    "client_type": "web"
  }
  ```

## 微信小程序

对于微信小程序集成，使用专用端点：

- 通过微信小程序注册
  ```
  POST /api/v1/auth/wxmini/register
  Content-Type: application/json
  x-wx-unionid: <用户union_id>
  x-wx-openid: <用户open_id>

  {
    "phone": "",
    "email": "",
    "name": "小程序用户😄",
    "avatar_url": "",
    "gender": "MALE",
    "birth_date": "2015-07-27"
  }
  ```

- 通过微信小程序登录
  ```
  POST /api/v1/auth/wxmini/login
  Content-Type: application/json
  x-wx-unionid: <用户union_id>
  x-wx-openid: <用户open_id>
  ```

## 错误响应

### 常见错误

**无效授权码：**
```json
{
  "success": false,
  "message": "无效的 OAuth 授权码",
  "error": "invalid_oauth_code"
}
```

**无效重定向 URL：**
```json
{
  "success": false,
  "message": "无效的重定向 URL",
  "error": "invalid_redirect_url"
}
```

**邮箱已存在：**
```json
{
  "success": false,
  "message": "邮箱已存在",
  "error": "email_already_exists"
}
```
