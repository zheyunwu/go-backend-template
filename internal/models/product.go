package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// JSONData defines a type for storing JSON formatted data.
type JSONData map[string]interface{}

// Scan implements the sql.Scanner interface, used to convert database values to JSONData.
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

// Value implements the driver.Valuer interface, used to convert JSONData to a database storable value.
func (j JSONData) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// DescriptionStatus defines an ENUM type for description status.
type DescriptionStatus string

const (
	PENDING  DescriptionStatus = "PENDING"  // Description generation is pending
	LOADING  DescriptionStatus = "LOADING"  // Description is currently being generated/loaded
	LOADED   DescriptionStatus = "LOADED"   // Description is loaded and up-to-date
	OUTDATED DescriptionStatus = "OUTDATED" // Description is outdated and needs regeneration
)

// Product represents the product table.
type Product struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:255;not null"`
	Barcode     string `json:"barcode" gorm:"size:50;index"`      // Barcode
	BarcodeType string `json:"barcode_type" gorm:"size:20;index"` // Barcode type (EAN13, EAN8, UPC, ISBN, ASIN, GTIN, etc.)

	// Description fields
	Description       JSONData          `json:"description" gorm:"type:json"`                                        // Product description, stored as JSON
	DescriptionStatus DescriptionStatus `json:"description_status" gorm:"type:description_status;default:'PENDING'"` // Description status (PostgreSQL specific type)
	// DescriptionStatus    DescriptionStatus `json:"description_status" gorm:"type:enum('PENDING', 'LOADING', 'LOADED', 'OUTDATED');default:'PENDING'"` // Description status (MySQL enum type)
	DescriptionUpdatedAt *time.Time `json:"description_loaded_at"` // Timestamp of when the description was last updated/loaded

	// Timestamp fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	// Associations - managed automatically by GORM
	Images     []ProductImage `json:"images" gorm:"foreignKey:ProductID"`             // Product images
	Categories []Category     `json:"categories" gorm:"many2many:product_categories"` // Product categories (many-to-many)

	// User interaction associations - using join tables with composite primary keys
	LikedByUsers     []User `json:"-" gorm:"many2many:user_product_likes;joinForeignKey:product_id;joinReferences:user_id"`
	FavoritedByUsers []User `json:"-" gorm:"many2many:user_product_favorites;joinForeignKey:product_id;joinReferences:user_id"`
}

// TableName specifies the table name for the Product model.
func (Product) TableName() string {
	return "products"
}

// ProductImage represents the product image table.
type ProductImage struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ProductID uint           `json:"product_id" gorm:"index"`           // Foreign key to Product
	ImageURL  string         `json:"image_url" gorm:"size:255;not null"` // URL of the image
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
}

// TableName specifies the table name for the ProductImage model.
func (ProductImage) TableName() string {
	return "product_images"
}
