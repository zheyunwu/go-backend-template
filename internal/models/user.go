package models

import (
	"time"

	"gorm.io/gorm"
)

// 定义ENUM类型
type GenderType string

const (
	PREFER_NOT_TO_SAY GenderType = "PREFER_NOT_TO_SAY"
	MALE              GenderType = "MALE"
	FEMALE            GenderType = "FEMALE"
	OTHER             GenderType = "OTHER"
)

type UserProvider struct {
	UserID        uint       `json:"user_id" gorm:"index;not null"`                    // 外键，指向users表
	Provider      string     `json:"provider" gorm:"primaryKey;type:varchar(50)"`      // 联合主键的一部分
	ProviderUID   string     `json:"provider_uid" gorm:"primaryKey;type:varchar(100)"` // 联合主键的一部分
	WechatUnionID *string    `json:"wechat_union_id" gorm:"type:varchar(100);index"`   // 微信的 UnionID，适用于小程序 & App 端用户
	AccessToken   *string    `json:"access_token" gorm:"type:text"`                    // 可选
	RefreshToken  *string    `json:"refresh_token" gorm:"type:text"`                   // 可选
	ExpiresAt     *time.Time `json:"expires_at"`                                       // 可选，OAuth2 的过期时间
	CreatedAt     time.Time  `json:"created_at"`                                       // 创建时间
	UpdatedAt     time.Time  `json:"updated_at"`                                       // 更新时间
}

// 用户表
type User struct {
	ID              uint    `json:"id" gorm:"primaryKey;<-:create"` // 内部系统的 UserID
	Email           *string `json:"email" gorm:"size:100;uniqueIndex:idx_email"`
	IsEmailVerified bool    `json:"is_email_verified" gorm:"default:false"`     // 邮箱是否已验证
	Phone           *string `json:"phone" gorm:"size:20;uniqueIndex:idx_phone"` // 手机号遵循 E.164 格式
	Password        *string `json:"-" gorm:"size:255"`                          // 密码字段，存储哈希值，不直接暴露

	Name      string     `json:"name" gorm:"size:100"`
	AvatarURL *string    `json:"avatar_url" gorm:"size:255"`
	Gender    GenderType `json:"gender" gorm:"type:gender_enum;default:'PREFER_NOT_TO_SAY'"` // 性别（PostgreSQL）
	// Gender            GenderType `json:"gender" gorm:"type:enum('PREFER_NOT_TO_SAY', 'MALE', 'FEMALE', 'OTHER');default:'PREFER_NOT_TO_SAY'"` // 性别（MySQL）
	BirthDate *time.Time `json:"birth_date" gorm:"type:date"`                  // 显式指定为DATE类型而非默认的DATETIME
	Locale    string     `gorm:"size:35;not null;default:en-US" json:"locale"` // 用户语言地区，遵循IETF BCP 47标准

	Role      string     `json:"role" gorm:"size:20;default:'user'"` // user, admin
	IsBanned  bool       `json:"is_banned" gorm:"default:false"`     // 新增字段：用户封禁状态，默认为false
	LastLogin *time.Time `json:"last_login"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	// 关联字段 - 由GORM自动管理
	UserProviders []UserProvider `json:"user_providers" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"` // 用户关联的第三方登录提供商

	// 用户交互关联 - 使用复合主键的连接表
	Favorites    []Product `json:"-" gorm:"many2many:user_product_favorites;joinForeignKey:user_id;joinReferences:product_id"`
	ProductLikes []Product `json:"-" gorm:"many2many:user_product_likes;joinForeignKey:user_id;joinReferences:product_id"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// UserAuthDetails 包含用户认证信息
type UserAuthDetails struct {
	UserID uint
	Role   string
}
