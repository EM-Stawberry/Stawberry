package product

import (
	"context"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type Repository interface {
	GetProductByID(ctx context.Context, id string) (entity.Product, error)
	SelectProducts(ctx context.Context, offset, limit int) ([]entity.Product, int, error)
	SelectProductsByName(ctx context.Context, name string, offset, limit int) ([]entity.Product, int, error)
	SelectProductsByFilters(ctx context.Context, categoryID int, filters map[string]interface{},
		offset, limit int) ([]entity.Product, int, error)
	SelectShopProducts(ctx context.Context, shopID int, offset, limit int) ([]entity.Product, int, error)
	GetAttributesByID(ctx context.Context, productID string) (map[string]interface{}, error)
	GetPriceRangeByProductID(ctx context.Context, productID int) (int, int, error)
	GetAverageRatingByProductID(ctx context.Context, productID int) (float64, int, error)
}

type Service struct {
	ProductRepository Repository
}

func NewService(productRepo Repository) *Service {
	return &Service{ProductRepository: productRepo}
}

// GetProductByID получает продукт по его ID
func (ps *Service) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {
	product, err := ps.ProductRepository.GetProductByID(ctx, id)
	if err != nil {
		return entity.Product{}, err
	}
	attrs, err := ps.ProductRepository.GetAttributesByID(ctx, id)
	if err != nil {
		return entity.Product{}, err
	}
	product.Attributes = attrs
	minPrice, maxPrice, _ := ps.ProductRepository.GetPriceRangeByProductID(ctx, product.ID)
	product.MinimalPrice = minPrice
	product.MaximalPrice = maxPrice
	return product, nil
}

// SelectProducts получает весь список продуктов
func (ps *Service) SelectProducts(
	ctx context.Context,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.ProductRepository.SelectProducts(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	products, err = ps.enrichProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// SelectProductsByName выполняет поиск продукта по его имени
func (ps *Service) SelectProductsByName(
	ctx context.Context,
	name string,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.ProductRepository.SelectProductsByName(ctx, name, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	products, err = ps.enrichProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// SelectProductsByFilters выполняет фильтрацию по ID категории и аттрибутам продукта
func (ps *Service) SelectProductsByFilters(
	ctx context.Context,
	categoryID int,
	filters map[string]interface{},
	offset, limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.ProductRepository.SelectProductsByFilters(ctx, categoryID, filters, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	products, err = ps.enrichProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// SelectShopProducts выполняет фильтрацию по ID магазина
func (ps *Service) SelectShopProducts(
	ctx context.Context,
	shopID int,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.ProductRepository.SelectShopProducts(ctx, shopID, offset, limit)
	if err != nil {
		return nil, 0, err
	}

	products, err = ps.enrichProducts(ctx, products)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// EnrichProducts выполняет обогащение продукта информацией о диапазоне цены, средней оценке и количестве отзывов
func (ps *Service) enrichProducts(
	ctx context.Context,
	products []entity.Product,
) ([]entity.Product, error) {
	for i := range products {
		minPrice, maxPrice, err := ps.ProductRepository.GetPriceRangeByProductID(ctx, products[i].ID)
		if err != nil {
			return nil, err
		}

		avgRating, countReviews, err := ps.ProductRepository.GetAverageRatingByProductID(ctx, products[i].ID)
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
