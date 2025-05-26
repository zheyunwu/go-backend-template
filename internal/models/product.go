package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// 定义JSONData 类型，用于存储 JSON 格式的数据
type JSONData map[string]interface{}

// 实现 sql.Scanner 接口，用于将数据库中的值转换为 JSONData
func (j *JSONData) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &j)
}

// 实现 driver.Valuer 接口，用于将 JSONData 转换为数据库可存储的值
func (j JSONData) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// 定义ENUM类型
type DescriptionStatus string

const (
	PENDING  DescriptionStatus = "PENDING"
	LOADING  DescriptionStatus = "LOADING"
	LOADED   DescriptionStatus = "LOADED"
	OUTDATED DescriptionStatus = "OUTDATED"
)

// 商品表
type Product struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:255;not null"`
	Barcode     string `json:"barcode" gorm:"size:50;index"`      // 条形码
	BarcodeType string `json:"barcode_type" gorm:"size:20;index"` // 条形码类型 (EAN13, EAN8, UPC, ISBN, ASIN, GTIN, etc.)

	// 描述字段
	Description       JSONData          `json:"description" gorm:"type:json"`                                        // 产品描述，存储为JSON格式
	DescriptionStatus DescriptionStatus `json:"description_status" gorm:"type:description_status;default:'PENDING'"` // 描述状态（PostgreSQL）
	// DescriptionStatus    DescriptionStatus `json:"description_status" gorm:"type:enum('PENDING', 'LOADING', 'LOADED', 'OUTDATED');default:'PENDING'"` // 描述状态（MySQL）
	DescriptionUpdatedAt *time.Time `json:"description_loaded_at"` // 描述更新时间

	// 时间戳字段
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	// 关联字段 - 由GORM自动管理
	Images     []ProductImage `json:"images" gorm:"foreignKey:ProductID"`             // 产品图片
	Categories []Category     `json:"categories" gorm:"many2many:product_categories"` // 多个分类

	// 用户交互关联 - 使用复合主键的连接表
	LikedByUsers     []User `json:"-" gorm:"many2many:user_product_likes;joinForeignKey:product_id;joinReferences:user_id"`
	FavoritedByUsers []User `json:"-" gorm:"many2many:user_product_favorites;joinForeignKey:product_id;joinReferences:user_id"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// 商品图片表
type ProductImage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProductID uint           `json:"product_id" gorm:"index"`
	ImageURL  string         `json:"image_url" gorm:"size:255;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

// TableName 指定表名
func (ProductImage) TableName() string {
	return "product_images"
}
