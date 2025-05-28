package product

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type Repository interface {
	InsertProduct(ctx context.Context, product Product) (uint, error)
	GetProductByID(ctx context.Context, id string) (entity.Product, error)
	SelectProducts(ctx context.Context, offset, limit int) ([]entity.Product, int, error)
	SelectProductsByName(ctx context.Context, name string, offset, limit int) ([]entity.Product, int, error)
	SelectProductsByCategoryAndAttributes(ctx context.Context, categoryID int, filters map[string]interface{}, offset, limit int) ([]entity.Product, int, error)
	SelectShopProducts(ctx context.Context, shopID int, offset, limit int) ([]entity.Product, int, error)
	UpdateProduct(ctx context.Context, id string, update UpdateProduct) error
	GetProductAttributesByID(ctx context.Context, productID string) (map[string]interface{}, error)
	GetPriceRangeByProductID(ctx context.Context, productID int) (float64, float64, error)
	GetAverageRatingByProductID(ctx context.Context, productID int) (float64, int, error)
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
	product, err := ps.productRepository.GetProductByID(ctx, id)
	if err != nil {
		return entity.Product{}, err
	}
	attrs, err := ps.productRepository.GetProductAttributesByID(ctx, id)
	if err != nil {
		return entity.Product{}, err
	}
	product.Attributes = attrs
	minPrice, maxPrice, _ := ps.productRepository.GetPriceRangeByProductID(ctx, product.ID)
	product.MinimalPrice = minPrice
	product.MaximalPrice = maxPrice
	return product, nil
}

func (ps *Service) SelectProducts(
	ctx context.Context,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.productRepository.SelectProducts(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	products, err = ps.EnrichProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (ps *Service) SelectProductsByName(
	ctx context.Context,
	name string,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.productRepository.SelectProductsByName(ctx, name, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	products, err = ps.EnrichProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (ps *Service) SelectProductsByCategoryAndAttributes(
	ctx context.Context,
	categoryID int,
	filters map[string]interface{},
	offset, limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.productRepository.SelectProductsByCategoryAndAttributes(ctx, categoryID, filters, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	products, err = ps.EnrichProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (ps *Service) SelectShopProducts(
	ctx context.Context,
	shopID int,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.productRepository.SelectShopProducts(ctx, shopID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	products, err = ps.EnrichProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (ps *Service) UpdateProduct(
	ctx context.Context,
	id string,
	updateProduct UpdateProduct,
) error {
	return ps.productRepository.UpdateProduct(ctx, id, updateProduct)
}

func (ps *Service) EnrichProducts(
	ctx context.Context,
	products []entity.Product,
) ([]entity.Product, error) {
	for i := range products {
		minPrice, maxPrice, err := ps.productRepository.GetPriceRangeByProductID(ctx, products[i].ID)
		if err != nil {
			return nil, err
		}

		avgRating, countReviews, err := ps.productRepository.GetAverageRatingByProductID(ctx, products[i].ID)
		if err != nil {
			return nil, err
		}

		products[i].MinimalPrice = minPrice
		products[i].MaximalPrice = maxPrice
		products[i].AverageRating = avgRating
		products[i].CountReviews = countReviews
	}
	return products, nil
}
