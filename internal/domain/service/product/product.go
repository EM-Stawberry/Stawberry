package product

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type Repository interface {
	InsertProduct(ctx context.Context, product Product) (uint, error)
	GetProductByID(ctx context.Context, id string) (entity.Product, error)
	SelectProducts(ctx context.Context, offset, limit int) ([]entity.Product, int, error)
	SelectStoreProducts(ctx context.Context, id string, offset, limit int) ([]entity.Product, int, error)
	UpdateProduct(ctx context.Context, id string, update UpdateProduct) error
}

type Service struct {
	productRepository Repository
}

func NewService(productRepo Repository) *Service {
	return &Service{productRepository: productRepo}
}

func (ps *Service) CreateProduct(
	ctx context.Context,
	product Product,
) (uint, error) {
	return ps.productRepository.InsertProduct(ctx, product)
}

func (ps *Service) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {
	return ps.productRepository.GetProductByID(ctx, id)
}

func (ps *Service) GetProducts(
	ctx context.Context,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	return ps.productRepository.SelectProducts(ctx, offset, limit)
}

func (ps *Service) GetStoreProducts(
	ctx context.Context,
	id string,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	return ps.productRepository.SelectStoreProducts(ctx, id, offset, limit)
}

func (ps *Service) UpdateProduct(
	ctx context.Context,
	id string,
	updateProduct UpdateProduct,
) error {
	return ps.productRepository.UpdateProduct(ctx, id, updateProduct)
}
