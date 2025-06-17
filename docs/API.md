# API说明

[点击这里](v1.postman_collection.json)查看对应Postman Collection

## 登录注册相关

- 邮箱密码注册
    ```http
    POST /api/v1/auth/register
    Content-Type: application/json

    {
        "email": "example@domain.com",
        "password": "12345678",
        "name": "起个用户昵称😁",
        "gender": "MALE",
        "birth_date": "2015-07-31",
        "locale": "zh"
    }
    ```

    响应示例：
    ```json
    {
        "status": "success",
        "data": {
            "id": 12
        }
    }
    ```

- 邮箱密码登录
    ```http
    POST /api/v1/auth/login
    Content-Type: application/json

    {
        "email_or_phone": "example@domain.com",
        "password": "12345678"
    }
    ```

    响应示例：
    ```json
    {
        "status": "success",
        "data": {
            "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "expires_in": 604800,
            "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "token_type": "Bearer"
        }
    }
    ```

- 刷新Token
    ```http
    POST /api/v1/auth/refresh
    Content-Type: application/json

    {
        "refresh_token": "{{REFRESH_TOKEN}}"
    }
    ```

    响应示例：
    ```json
    {
        "status": "success",
        "message": "Token refreshed successfully",
        "data": {
            "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "expires_in": 604800,
            "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
            "token_type": "Bearer"
        }
    }
    ```

- 获取个人Profile
    ```http
    GET /api/v1/auth/profile
    Authorization: Bearer <ACCESS_TOKEN>
    ```

    响应示例：
    ```json
    {
        "status": "success",
        "data": {
            "id": 7,
            "name": "奥特曼",
            "avatar_url": "",
            "gender": "MALE",
            "email": "kejosat522@nab4.com",
            "is_email_verified": false,
            "phone": null,
            "birth_date": "2015-07-31T00:00:00Z",
            "locale": "zh",
            "created_at": "2025-06-14T21:10:14.11812Z",
            "updated_at": "2025-06-14T21:12:21.748986Z"
        }
    }
    ```

- 更新个人Profile
    ```http
    PATCH /api/v1/auth/profile
    Content-Type: application/json
    Authorization: Bearer <ACCESS_TOKEN>

    {
        "name": "新昵称",
        "avatar_url": "https://example.com/avatar.jpg",
        "gender": "OTHER",
        "email": "new@example.com",
        "phone": "+86123456789",
        "birth_date": "1995-01-01",
        "locale": "zh"
    }
    ```

- 修改密码
    ```http
    PATCH /api/v1/auth/password
    Content-Type: application/json
    Authorization: Bearer <ACCESS_TOKEN>

    {
        "current_password": "12345678",
        "new_password": "999911119999"
    }
    ```

- 发送邮箱验证码
    ```http
    POST /api/v1/auth/email/send-verification
    Content-Type: application/json

    {
        "email": "kejosat522@nab4.com"
    }
    ```

- 验证邮箱
    ```http
    POST /api/v1/auth/email/verify
    Content-Type: application/json

    {
        "email": "kejosat522@nab4.com",
        "code": "221224"
    }
    ```

- 请求重置密码
    ```http
    POST /api/v1/auth/password/reset-request
    Content-Type: application/json

    {
        "email": "kejosat522@nab4.com"
    }
    ```

- 重置密码
    ```http
    POST /api/v1/auth/password/reset
    Content-Type: application/json

    {
        "email": "kejosat522@nab4.com",
        "new_password": "newPassword123",
        "reset_token": "M1RGYKCS"
    }
    ```

## 微信小程序（云托管）

- 微信小程序注册
    ```http
    POST /api/v1/auth/wxmini/register
    Content-Type: application/json

    {
        "code": "wx_mini_program_code",
        "name": "用户昵称",
        "avatar_url": "https://example.com/avatar.jpg"
    }
    ```

- 微信小程序登录
    ```http
    POST /api/v1/auth/wxmini/login
    Content-Type: application/json

    {
        "code": "wx_mini_program_code"
    }
    ```


## OAuth2登录

- Google登录 - Auth Code换Token
    ```http
    POST /api/v1/auth/google/token
    Content-Type: application/json

    {
        "code": "google_oauth_authorization_code",
        "code_verifier": "pkce_code_verifier",
        "redirect_uri": "https://yourapp.com/auth/callback",
        "client_type": "web"
    }
    ```

- 微信登录 - Auth Code换Token
    ```http
    POST /api/v1/auth/wechat/token
    Content-Type: application/json

    {
        "code": "wechat_authorization_code",
        "client_type": "web"
    }
    ```

## 绑定第三方账号

- 绑定Google账号
    ```http
    POST /api/v1/auth/google/bind
    Content-Type: application/json
    Authorization: Bearer <ACCESS_TOKEN>

    {
        "code": "google_oauth_authorization_code",
        "code_verifier": "pkce_code_verifier",
        "redirect_uri": "https://yourapp.com/auth/callback",
        "client_type": "web"  // 或 "ios"
    }
    ```

- 解绑Google账号
    ```http
    POST /api/v1/auth/google/unbind
    Content-Type: application/json
    Authorization: Bearer <ACCESS_TOKEN>
    ```

- 绑定微信账号
    ```http
    POST /api/v1/auth/wechat/bind
    Content-Type: application/json
    Authorization: Bearer <ACCESS_TOKEN>

    {
        "code": "wechat_authorization_code",
        "client_type": "web"
    }
    ```

- 解绑微信账号
    ```http
    POST /api/v1/auth/wechat/unbind
    Content-Type: application/json
    Authorization: Bearer <ACCESS_TOKEN>
    ```

## 产品相关

- 获取产品列表
    ```http
    GET /api/v1/products?limit=10&page=1&sort=updated_at:desc
    Authorization: Bearer <ACCESS_TOKEN>
    ```

    响应示例：
    ```json
    {
        "status": "success",
        "data": [
            {
                "id": 17,
                "barcode": "8888888822221",
                "barcode_type": "EAN13",
                "name": "产品名称",
                "name_cn": "中文名称",
                "description": "产品描述",
                "description_cn": "中文描述",
                "category_id": 0,
                "created_at": "2025-03-10T14:45:27+01:00",
                "updated_at": "2025-03-10T16:25:43+01:00",
                "images": [
                    {
                        "image_url": "https://example.com/image1.jpg"
                    }
                ],
                "retailers": [
                    {
                        "name": "REWE",
                        "url": "https://shop.rewe.de/product/123"
                    }
                ],
                "is_liked": true,
                "is_favorited": false
            }
        ],
        "pagination": {
            "total_count": 16,
            "page_size": 10,
            "current_page": 1,
            "total_pages": 2
        }
    }
    ```

- 获取产品详情
    ```http
    GET /api/v1/products/{id}
    Authorization: Bearer <ACCESS_TOKEN>
    ```

- 产品点赞/取消点赞
    ```http
    PUT /api/v1/products/{id}/like
    Content-Type: application/json
    Authorization: Bearer <ACCESS_TOKEN>

    {
        "is_liked": true
    }
    ```

- 产品收藏/取消收藏
    ```http
    PUT /api/v1/products/{id}/favorite
    Content-Type: application/json
    Authorization: Bearer <ACCESS_TOKEN>

    {
        "is_favorited": true
    }
    ```

- 获取产品统计信息
    ```http
    GET /api/v1/products/{id}/stats
    Authorization: Bearer <ACCESS_TOKEN>
    ```

    响应示例：
    ```json
    {
        "status": "success",
        "data": {
            "favorite_count": 1,
            "like_count": 1
        }
    }
    ```

## 分类相关

- 获取分类列表
    ```http
    GET /api/v1/categories
    Authorization: Bearer <ACCESS_TOKEN>
    ```

- 获取分类树状结构
    ```http
    GET /api/v1/categories/tree?depth=5
    Authorization: Bearer <ACCESS_TOKEN>
    ```

## 后台管理接口

### 用户管理

> 需要管理员权限

- 获取用户列表
    ```http
    GET /api/v1/admin/users
    Authorization: Bearer <ADMIN_ACCESS_TOKEN>
    ```

- 获取用户详情
    ```http
    GET /api/v1/admin/users/{id}
    Authorization: Bearer <ADMIN_ACCESS_TOKEN>
    ```

- 更新用户
    ```http
    PUT /api/v1/admin/users/{id}
    Content-Type: application/json
    Authorization: Bearer <ADMIN_ACCESS_TOKEN>

    {
        "name": "新用户名",
        "email": "new@example.com",
        "status": "active"
    }
    ```

- 删除用户
    ```http
    DELETE /api/v1/admin/users/{id}
    Authorization: Bearer <ADMIN_ACCESS_TOKEN>
    ```

### 产品管理

> 需要管理员权限

- 创建产品
    ```http
    POST /api/v1/admin/products
    Content-Type: application/json
    Authorization: Bearer <ADMIN_ACCESS_TOKEN>

    {
        "barcode": "1234567890123",
        "barcode_type": "EAN13",
        "name": "产品名称",
        "name_cn": "中文名称",
        "description": "产品描述",
        "description_cn": "中文描述",
        "category_id": 1,
        "images": [
            {
                "image_url": "https://example.com/image.jpg"
            }
        ],
        "retailers": [
            {
                "name": "商店名称",
                "url": "https://example.com/product"
            }
        ]
    }
    ```

- 更新产品
    ```http
    PUT /api/v1/admin/products/{id}
    Content-Type: application/json
    Authorization: Bearer <ADMIN_ACCESS_TOKEN>

    {
        "name": "更新后的产品名称",
        "description": "更新后的描述"
    }
    ```

- 删除产品
    ```http
    DELETE /api/v1/admin/products/{id}
    Authorization: Bearer <ADMIN_ACCESS_TOKEN>
    ```

## 通用响应格式

### 成功响应
```json
{
    "status": "success",
    "data": {
        // 具体数据
    },
    "message": "操作成功" // 可选
}
```

### 错误响应
```json
{
    "status": "error",
    "message": "错误信息",
    "code": "ERROR_CODE" // 可选
}
```

### 常见HTTP状态码
- `200 OK` - 请求成功
- `201 Created` - 资源创建成功
- `204 No Content` - 请求成功，无返回内容
- `400 Bad Request` - 请求参数错误
- `401 Unauthorized` - 未授权
- `403 Forbidden` - 权限不足
- `404 Not Found` - 资源不存在
- `409 Conflict` - 资源冲突（如邮箱已存在）
- `500 Internal Server Error` - 服务器内部错误

## 认证说明

大部分接口需要在请求头中携带JWT Token：

```http
Authorization: Bearer <ACCESS_TOKEN>
```

Token可通过以下方式获取：
1. 邮箱密码登录
2. OAuth2登录（Google、微信）
3. 微信小程序登录
4. 刷新Token

Token有效期为7天，可使用refresh_token刷新。