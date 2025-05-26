package models

import "time"

// UserProductFavorite 用户收藏产品记录
type UserProductFavorite struct {
	UserID    uint      `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	ProductID uint      `json:"product_id" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (UserProductFavorite) TableName() string {
	return "user_product_favorites"
}

// UserProductLike 用户点赞产品记录
type UserProductLike struct {
	UserID    uint      `json:"user_id" gorm:"primaryKey;autoIncrement:false"`
	ProductID uint      `json:"product_id" gorm:"primaryKey;autoIncrement:false"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (UserProductLike) TableName() string {
	return "user_product_likes"
}
