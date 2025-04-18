package model

import (

	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/product"
)

type Product struct {
	ID          uint   `db:"id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	CategoryID  uint   `db:"category_id"`
}

type UpdateProduct struct {
	StoreID     *uint    `gorm:"column:store_id"`
	Name        *string  `gorm:"column:name"`
	Description *string  `gorm:"column:description"`
	Price       *float64 `gorm:"column:price"`
	Category    *string  `gorm:"column:category"`
	InStock     *bool    `gorm:"column:in_stock"`
}

func ConvertProductFromSvc(p product.Product) Product {
	return Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CategoryID:  p.CategoryID,
	}
}

func ConvertProductToEntity(p Product) entity.Product {
	return entity.Product{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		CategoryID:  p.CategoryID,
	}
}

func ConvertUpdateProductFromSvc(up product.UpdateProduct) UpdateProduct {
	return UpdateProduct{
		StoreID:     up.StoreID,
		Name:        up.Name,
		Description: up.Description,
		Price:       up.Price,
		Category:    up.Category,
		InStock:     up.InStock,
	}
}
