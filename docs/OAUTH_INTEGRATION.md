# OAuth2 Integration Guide

This backend supports OAuth2 authentication with Google and WeChat providers, using secure authorization flows suitable for different client types.

## 1. Google OAuth2 (Authorization Code Flow + PKCE)

[Google Official Documentation (Auth Code + PKCE)](https://developers.google.com/identity/protocols/oauth2/native-app)

### Configuration

Add your Google OAuth2 credentials to the configuration file. You can configure different client credentials for iOS and Web applications:

```yaml
google:
  ios:
    client_id: "your-ios-google-client-id"
    client_secret: "your-ios-google-client-secret"
    redirect_urls:
      - "com.yourapp.scheme://oauth/callback"  # iOS App Deep Link
  web:
    client_id: "your-web-google-client-id"
    client_secret: "your-web-google-client-secret"
    redirect_urls:
      - "http://localhost:3000/auth/callback"  # Local Development Web App
      - "https://yourapp.com/auth/callback"    # Production Web App
```

### API Endpoints

1.  **Google Login (Exchange Auth Code for Token)**
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

    **Parameter Explanation:**
    - `code`: Google authorization code from the OAuth2 flow.
    - `code_verifier`: PKCE code verifier for security.
    - `redirect_uri`: Must match one of the configured redirect URLs.
    - `client_type`: `"ios"` or `"web"`.

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
        "message": "User authenticated successfully" // "用户认证成功" -> "User authenticated successfully"
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
        "message": "User registered and authenticated successfully" // "用户注册并认证成功" -> "User registered and authenticated successfully"
    }
    ```

2.  **Bind Google Account**
    ```http
    POST /api/v1/auth/google/bind
    Content-Type: application/json
    Authorization: Bearer <access_token>

    {
        "code": "google_oauth_authorization_code",
        "code_verifier": "pkce_code_verifier",
        "redirect_uri": "https://yourapp.com/auth/callback",
        "client_type": "web"  // or "ios"
    }
    ```

3.  **Unbind Google Account**
    ```http
    POST /api/v1/auth/google/unbind
    Content-Type: application/json
    Authorization: Bearer <access_token>
    ```

### Google Login Client Implementation Guide (PKCE)

1.  **Generate PKCE Parameters:**
    ```javascript
    // Generate codeVerifier (43-128 characters)
    const codeVerifier = generateRandomString(128);

    // Generate codeChallenge
    const codeChallenge = base64URLEncode(sha256(codeVerifier));
    ```

2.  **Redirect to Google Authorization:**
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

3.  **Handle Callback and Exchange Code:**
    ```javascript
    // Extract authorization code from callback URL
    const urlParams = new URLSearchParams(window.location.search);
    const code = urlParams.get('code');

    // Exchange for JWT with backend
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

## 2. WeChat OAuth2

[WeChat Official Documentation - Website Apps](https://developers.weixin.qq.com/doc/oplatform/Website_App/WeChat_Login/Wechat_Login.html)
[WeChat Official Documentation - Mobile Apps](https://developers.weixin.qq.com/doc/oplatform/Mobile_App/WeChat_Login/Development_Guide.html)

### Configuration

```yaml
wechat:
  web:
    app_id: "your-web-wechat-app-id"
    secret: "your-web-wechat-secret"
  app:
    app_id: "your-app-wechat-app-id"
    secret: "your-app-wechat-secret"
```

### API Endpoints

1.  **WeChat Login (Exchange Auth Code for Token)**
    ```
    POST /api/v1/auth/wechat/exchange
    Content-Type: application/json

    {
        "code": "wechat_authorization_code",
        "client_type": "web" // or "app"
    }
    ```
2.  **Bind WeChat Account**
    ```http
    POST /api/v1/auth/wechat/bind
    Content-Type: application/json
    Authorization: Bearer <access_token>

    {
        "code": "wechat_authorization_code", // Note: This is a WeChat auth code for binding
        "client_type": "web"  // or "app"
    }
    ```

3.  **Unbind WeChat Account**
    ```http
    POST /api/v1/auth/wechat/unbind
    Content-Type: application/json
    Authorization: Bearer <access_token>
    ```

## 3. WeChat Mini Program

For WeChat Mini Program integration, use dedicated endpoints:

1.  **Register via WeChat Mini Program**
  ```
  POST /api/v1/auth/wxmini/register
  Content-Type: application/json
  x-wx-unionid: <user_union_id>
  x-wx-openid: <user_open_id>

  {
    "phone": "",
    "email": "",
    "name": "Mini Program User 😄", // "小程序用户😄" -> "Mini Program User 😄"
    "avatar_url": "",
    "gender": "MALE",
    "birth_date": "2015-07-27"
  }
  ```

2.  **Login via WeChat Mini Program**
  ```
  POST /api/v1/auth/wxmini/login
  Content-Type: application/json
  x-wx-unionid: <user_union_id>
  x-wx-openid: <user_open_id>
  ```
