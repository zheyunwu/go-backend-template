package dto

import (
	"time"

	"github.com/go-backend-template/internal/models"
)

/*
DTOs for Admin operations
*/

// CreateProductRequest is the request body for creating a product.
type CreateProductRequest struct {
	Name        string          `json:"name" validate:"required,min=1,max=255"`
	Barcode     string          `json:"barcode" validate:"omitempty,min=8,max=13"`
	BarcodeType string          `json:"barcode_type" validate:"omitempty,oneof=EAN13 EAN8 UPC ISBN ASIN GTIN"`
	Description models.JSONData `json:"description" validate:"omitempty"`
	CategoryIDs []uint          `json:"category_ids" validate:"omitempty,dive,gt=0"`
	ImageURLs   []string        `json:"image_urls" validate:"omitempty,dive,url"`
}

// ToModel converts the request body to a Product model.
func (r *CreateProductRequest) ToModel() *models.Product {
	return &models.Product{
		Name:        r.Name,
		Barcode:     r.Barcode,
		BarcodeType: r.BarcodeType,
		Description: r.Description,
	}
}

// UpdateProductRequest is the request body for updating a product.
type UpdateProductRequest struct {
	Name              *string                   `json:"name" validate:"omitempty,min=1,max=255"`
	Barcode           *string                   `json:"barcode" validate:"omitempty,min=8,max=13"`
	BarcodeType       *string                   `json:"barcode_type" validate:"omitempty,oneof=EAN13 EAN8 UPC ISBN ASIN GTIN"`
	Description       *models.JSONData          `json:"description" validate:"omitempty"`
	DescriptionStatus *models.DescriptionStatus `json:"description_status" validate:"omitempty,oneof=pending approved rejected"`
	ProductType       *string                   `json:"product_type" validate:"omitempty,min=1"`
	CategoryIDs       []uint                    `json:"category_ids" validate:"omitempty,dive,gt=0"`
	ImageURLs         []string                  `json:"image_urls" validate:"omitempty,dive,url"`
}

// ToMap converts the update request to a map of fields to update.
func (r *UpdateProductRequest) ToMap() map[string]interface{} {
	updates := make(map[string]interface{})

	if r.Name != nil {
		updates["name"] = *r.Name
	}
	if r.Barcode != nil {
		updates["barcode"] = *r.Barcode
	}
	if r.BarcodeType != nil {
		updates["barcode_type"] = *r.BarcodeType
	}
	if r.Description != nil {
		updates["description"] = *r.Description
	}
	if r.DescriptionStatus != nil {
		updates["description_status"] = *r.DescriptionStatus
	}

	return updates
}

/*
DTOs for User-facing operations
*/

// UserProductDTO is the product DTO for user-facing APIs, excluding the deleted_at field.
type UserProductDTO struct {
	ID                   uint                     `json:"id"`
	Barcode              string                   `json:"barcode"`
	BarcodeType          string                   `json:"barcode_type"`
	Name                 string                   `json:"name"`
	Description          models.JSONData          `json:"description"`
	DescriptionStaus     models.DescriptionStatus `json:"description_status"`
	DescriptionUpdatedAt *time.Time               `json:"description_updated_at"`
	CreatedAt            time.Time                `json:"created_at"`
	UpdatedAt            time.Time                `json:"updated_at"`
	// Associated fields
	Categories  []CategoryDTO         `json:"categories"`
	Images      []UserProductImageDTO `json:"images"`
	IsLiked     bool                  `json:"is_liked"`
	IsFavorited bool                  `json:"is_favorited"`
}

// CategoryDTO is a simplified category DTO for display within product details.
type CategoryDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// UserProductImageDTO is the product image DTO for user-facing APIs.
type UserProductImageDTO struct {
	ImageURL string `json:"image_url"`
}

// ToUserProductDTO converts a Product model to a UserProductDTO.
func ToUserProductDTO(product *models.Product) *UserProductDTO {
	if product == nil {
		return nil
	}

	dto := &UserProductDTO{
		ID:                   product.ID,
		Barcode:              product.Barcode,
		BarcodeType:          product.BarcodeType,
		Name:                 product.Name,
		Description:          product.Description,
		DescriptionStaus:     product.DescriptionStatus,
		DescriptionUpdatedAt: product.DescriptionUpdatedAt,
		CreatedAt:            product.CreatedAt,
		UpdatedAt:            product.UpdatedAt,
		IsLiked:              false, // Default value, will be set by handler if user is authenticated
		IsFavorited:          false, // Default value, will be set by handler if user is authenticated
	}

	// Convert categories
	for _, category := range product.Categories {
		dto.Categories = append(dto.Categories, CategoryDTO{
			ID:   category.ID,
			Name: category.Name,
		})
	}

	// Convert images
	for _, img := range product.Images {
		dto.Images = append(dto.Images, UserProductImageDTO{
			ImageURL: img.ImageURL,
		})
	}

	return dto
}

// ToUserProductDTOList converts a list of Product models to a list of UserProductDTOs.
func ToUserProductDTOList(products []models.Product) []UserProductDTO {
	dtos := make([]UserProductDTO, 0, len(products))
	for i := range products {
		if dto := ToUserProductDTO(&products[i]); dto != nil {
			dtos = append(dtos, *dto)
		}
	}
	return dtos
}
