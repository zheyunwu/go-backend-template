package models

import (
	"time"
)

// ProductCategory 产品与分类的多对多关联表
type ProductCategory struct {
	ProductID  uint      `json:"product_id" gorm:"primaryKey;autoIncrement:false"`
	CategoryID uint      `json:"category_id" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt  time.Time `json:"created_at"`

	// 关联对象引用（可选）
	Product  Product  `json:"-" gorm:"foreignKey:ProductID;references:ID"`
	Category Category `json:"-" gorm:"foreignKey:CategoryID;references:ID"`
}

// TableName 指定表名
func (ProductCategory) TableName() string {
	return "product_categories"
}
