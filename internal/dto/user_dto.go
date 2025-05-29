package dto

import (
	"time"

	"github.com/go-backend-template/internal/models"
)

/* Response DTOs */

// UserProfileDTO 用户信息
type UserProfileDTO struct {
	ID                uint              `json:"id"`
	Name              string            `json:"name"`
	AvatarURL         *string           `json:"avatar_url"`
	Gender            models.GenderType `json:"gender"`
	Email             *string           `json:"email"`
	Phone             *string           `json:"phone"`
	BirthDate         *time.Time        `json:"birth_date"`
	PreferredLanguage string            `json:"preferred_language"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FromUser 从用户模型转换为用户信息DTO
func ToUserProfileDTO(user *models.User) *UserProfileDTO {
	if user == nil {
		return nil
	}

	return &UserProfileDTO{
		ID:                user.ID,
		Name:              user.Name,
		AvatarURL:         user.AvatarURL,
		Gender:            user.Gender,
		Email:             user.Email,
		Phone:             user.Phone,
		BirthDate:         user.BirthDate,
		PreferredLanguage: user.PreferredLanguage,
		CreatedAt:         user.CreatedAt,
		UpdatedAt:         user.UpdatedAt,
	}
}

/* Request DTOs */

// UpdateProfileRequest 用户更新个人资料请求
type UpdateProfileRequest struct {
	Name              string            `json:"name"`
	AvatarURL         *string           `json:"avatar_url"`
	Gender            models.GenderType `json:"gender"`
	Email             *string           `json:"email"`      // Changed from string to *string
	Phone             *string           `json:"phone"`      // Changed from string to *string
	BirthDate         *string           `json:"birth_date"` // Changed from string to *string
	PreferredLanguage string            `json:"preferred_language"`
}

// ToMap 将更新请求转换为更新字段映射
func (r *UpdateProfileRequest) ToUpdatesMap() map[string]interface{} {
	updates := map[string]interface{}{}
	if r.Name != "" {
		updates["Name"] = r.Name
	}
	if r.AvatarURL != nil && *r.AvatarURL != "" {
		updates["AvatarURL"] = r.AvatarURL
	}
	if r.Gender != "" {
		updates["Gender"] = r.Gender
	}
	if r.Email != nil && *r.Email != "" {
		updates["Email"] = r.Email
	}
	if r.Phone != nil && *r.Phone != "" {
		updates["Phone"] = r.Phone
	}
	if r.BirthDate != nil && *r.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *r.BirthDate) // Dereference r.BirthDate
		if err == nil {                                          // 只在解析成功时添加
			// 确保时间部分为零值，只保留日期部分
			birthDate = time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
			updates["BirthDate"] = birthDate
		}
	}
	if r.PreferredLanguage != "" {
		updates["PreferredLanguage"] = r.PreferredLanguage
	}
	return updates
}

// RegisterFromWechatMiniProgramRequest 微信小程序注册请求
type RegisterFromWechatMiniProgramRequest struct {
	UpdateProfileRequest
}

// ToUser 将微信小程序注册请求转换为用户模型
func (r *RegisterFromWechatMiniProgramRequest) ToModel() *models.User {
	user := models.User{
		Name:      r.Name,
		AvatarURL: r.AvatarURL,
		Gender:    r.Gender,
	}

	if r.Email != nil && *r.Email != "" {
		user.Email = r.Email
	}
	if r.Phone != nil && *r.Phone != "" {
		user.Phone = r.Phone
	}
	if r.BirthDate != nil && *r.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *r.BirthDate)
		// 只在解析成功时设置出生日期
		if err == nil {
			// 确保时间部分为零值，只保留日期部分
			birthDate = time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
			user.BirthDate = &birthDate
		}
	}
	if r.PreferredLanguage != "" {
		user.PreferredLanguage = r.PreferredLanguage
	}

	return &user
}

// RegisterWithPasswordRequest 使用密码注册请求
type RegisterWithPasswordRequest struct {
	UpdateProfileRequest
	Password string `json:"password" binding:"required,min=8"`
}

func (r *RegisterWithPasswordRequest) ToModel(hashedPassword string) *models.User {
	user := models.User{
		Name:      r.Name,
		AvatarURL: r.AvatarURL,
		Gender:    r.Gender,
	}

	if r.Email != nil && *r.Email != "" {
		user.Email = r.Email
	}
	if r.Phone != nil && *r.Phone != "" {
		user.Phone = r.Phone
	}
	if r.BirthDate != nil && *r.BirthDate != "" {
		birthDate, err := time.Parse("2006-01-02", *r.BirthDate)
		// 只在解析成功时设置出生日期
		if err == nil {
			// 确保时间部分为零值，只保留日期部分
			birthDate = time.Date(birthDate.Year(), birthDate.Month(), birthDate.Day(), 0, 0, 0, 0, time.UTC)
			user.BirthDate = &birthDate
		}
	}
	if r.PreferredLanguage != "" {
		user.PreferredLanguage = r.PreferredLanguage
	}

	if hashedPassword != "" {
		user.Password = &hashedPassword // 使用哈希后的密码
	}

	return &user
}

// LoginWithPasswordRequest 使用密码登录请求
type LoginWithPasswordRequest struct {
	// 用户名或邮箱
	EmailOrPhone string `json:"email_or_phone" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

// Google OAuth2 相关 DTOs

// GoogleOAuthRequest Google OAuth2统一请求（登录或注册）
type GoogleOAuthRequest struct {
	Code         string `json:"code" binding:"required"`                      // OAuth authorization code
	CodeVerifier string `json:"code_verifier" binding:"required"`             // PKCE code verifier
	RedirectURI  string `json:"redirect_uri" binding:"required"`              // 重定向URI，必须与配置中的匹配
	ClientType   string `json:"client_type" binding:"required,oneof=ios web"` // 客户端类型：ios 或 web
}
