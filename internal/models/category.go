package models

import (
	"time"

	"gorm.io/gorm"
)

// Category represents the product category table.
type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:255;not null"` // Name in default language (e.g., English)
	NameZH    string         `json:"name_zh" gorm:"size:255"`       // Name in Chinese
	ParentID  *uint          `json:"parent_id" gorm:"index"`        // Pointer to allow null for root categories
	Enabled   bool           `json:"enabled" gorm:"default:false"`  // Whether the category is enabled
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	// Associations
	Products []Product  `json:"products" gorm:"many2many:product_categories"` // Products belonging to this category
	Parent   *Category  `json:"parent,omitempty" gorm:"foreignKey:ParentID"`  // Parent category
	Children []Category `json:"children,omitempty" gorm:"-"`                  // Child categories, not persisted to DB, only for API response
}

// TableName specifies the table name for the Category model.
func (Category) TableName() string {
	return "categories"
}
