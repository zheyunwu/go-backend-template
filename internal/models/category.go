package models

import (
	"time"

	"gorm.io/gorm"
)

// 商品分类表
type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:255;not null"`
	NameZH    string         `json:"name_zh" gorm:"size:255"`
	ParentID  *uint          `json:"parent_id" gorm:"index"`
	Enabled   bool           `json:"enabled" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	// 关联字段
	Products []Product  `json:"products" gorm:"many2many:product_categories"` // 多个产品
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`  // 父分类
	Children []Category `json:"children,omitempty" gorm:"-"`                  // 子分类，不持久化到数据库，仅用于API返回
}

// TableName 指定表名
func (Category) TableName() string {
	return "categories"
}
