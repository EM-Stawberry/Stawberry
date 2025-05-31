package product

import (
	"context"
)

type ProductRepository interface {
	InsertProduct(ctx context.Context, product *Product) (uint, error)
	GetProductByID(ctx context.Context, id uint) (*Product, error)
	SelectProducts(ctx context.Context, offset, limit int) ([]*Product, int, error)
	SelectStoreProducts(ctx context.Context, id uint, offset, limit int) ([]*Product, int, error)
	UpdateProduct(ctx context.Context, id uint, update *UpdateProduct) error
	DeleteProduct(ctx context.Context, id uint) error
}

type productService struct {
	productRepository ProductRepository
}

func NewService(productRepo ProductRepository) *productService {
	return &productService{productRepository: productRepo}
}

func (ps *productService) CreateProduct(ctx context.Context, product *Product) (uint, error) {
	return ps.productRepository.InsertProduct(ctx, product)
}

func (ps *productService) GetProductByID(ctx context.Context, id uint) (*Product, error) {
	return ps.productRepository.GetProductByID(ctx, id)
}

func (ps *productService) GetProducts(ctx context.Context, offset, limit int) ([]*Product, int, error) {
	return ps.productRepository.SelectProducts(ctx, offset, limit)
}

func (ps *productService) GetStoreProducts(ctx context.Context, id uint, offset, limit int) ([]*Product, int, error) {
	return ps.productRepository.SelectStoreProducts(ctx, id, offset, limit)
}

func (ps *productService) UpdateProduct(ctx context.Context, id uint, updateProduct *UpdateProduct) error {
	return ps.productRepository.UpdateProduct(ctx, id, updateProduct)
}

func (ps *productService) DeleteProduct(ctx context.Context, id uint) error {
	return ps.productRepository.DeleteProduct(ctx, id)
}
