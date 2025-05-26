package dto

import (
	"time"

	"github.com/go-backend-template/internal/models"
)

/*
面向Admin的DTO
*/

// CreateProductRequest 创建产品的请求体
type CreateProductRequest struct {
	Name        string          `json:"name" binding:"required"`
	Barcode     string          `json:"barcode" binding:"omitempty,min=8,max=13"`
	BarcodeType string          `json:"barcode_type" binding:"omitempty,oneof=EAN13 EAN8 UPC ISBN ASIN GTIN"`
	Description models.JSONData `json:"description"`
	CategoryIDs []uint          `json:"category_ids"`
	ImageURLs   []string        `json:"image_urls"`
}

// ToModel 将请求体转换为产品模型
func (r *CreateProductRequest) ToModel() *models.Product {
	return &models.Product{
		Name:        r.Name,
		Barcode:     r.Barcode,
		BarcodeType: r.BarcodeType,
		Description: r.Description,
	}
}

// UpdateProductRequest 更新产品的请求体
type UpdateProductRequest struct {
	Name              *string                   `json:"name"`
	Barcode           *string                   `json:"barcode" binding:"omitempty,min=8,max=13"`
	BarcodeType       *string                   `json:"barcode_type" binding:"omitempty,oneof=EAN13 EAN8 UPC ISBN ASIN GTIN"`
	Description       *models.JSONData          `json:"description"`
	DescriptionStatus *models.DescriptionStatus `json:"description_status"`
	ProductType       *string                   `json:"product_type"`
	CategoryIDs       []uint                    `json:"category_ids"`
	ImageURLs         []string                  `json:"image_urls"`
}

// ToMap 将更新请求转换为更新字段映射
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
面向User的DTO
*/

// UserProductDTO 面向用户的产品DTO，不包含deleted_at字段
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
	// 关联字段
	Categories  []CategoryDTO         `json:"categories"`
	Images      []UserProductImageDTO `json:"images"`
	IsLiked     bool                  `json:"is_liked"`
	IsFavorited bool                  `json:"is_favorited"`
}

// CategoryDTO 简化的分类DTO，用于在产品中展示
type CategoryDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// UserProductImageDTO 面向用户的产品图片DTO
type UserProductImageDTO struct {
	ImageURL string `json:"image_url"`
}

// ToUserProductDTO 将产品模型转换为用户产品DTO
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
		IsLiked:              false,
		IsFavorited:          false,
	}

	// 转换分类
	for _, category := range product.Categories {
		dto.Categories = append(dto.Categories, CategoryDTO{
			ID:   category.ID,
			Name: category.Name,
		})
	}

	// 转换图片
	for _, img := range product.Images {
		dto.Images = append(dto.Images, UserProductImageDTO{
			ImageURL: img.ImageURL,
		})
	}

	return dto
}

// ToUserProductDTOList 将产品模型列表转换为用户产品DTO列表
func ToUserProductDTOList(products []models.Product) []UserProductDTO {
	dtos := make([]UserProductDTO, 0, len(products))
	for i := range products {
		if dto := ToUserProductDTO(&products[i]); dto != nil {
			dtos = append(dtos, *dto)
		}
	}
	return dtos
}
