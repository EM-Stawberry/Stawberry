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
	GetPriceRangeByProductID(ctx context.Context, productID uint) (float64, float64, error)
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
	product.Attributes["Minimal Price"] = minPrice
	product.Attributes["Maximal Price"] = maxPrice
	return product, nil
}

func (ps *productService) SelectProducts(
	ctx context.Context,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.productRepository.SelectProducts(ctx, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range products {
		minPrice, maxPrice, _ := ps.productRepository.GetPriceRangeByProductID(ctx, products[i].ID)
		products[i].Attributes = map[string]interface{}{
			"Minimal Price": minPrice,
			"Maximal Price": maxPrice,
		}
	}
	return products, total, nil
}

func (ps *productService) SelectProductsByName(
	ctx context.Context,
	name string,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.productRepository.SelectProductsByName(ctx, name, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range products {
		minPrice, maxPrice, _ := ps.productRepository.GetPriceRangeByProductID(ctx, products[i].ID)
		products[i].Attributes = map[string]interface{}{
			"Minimal Price": minPrice,
			"Maximal Price": maxPrice,
		}
	}
	return products, total, nil
}

func (ps *productService) SelectProductsByCategoryAndAttributes(
	ctx context.Context,
	categoryID int,
	filters map[string]interface{},
	offset, limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.productRepository.SelectProductsByCategoryAndAttributes(ctx, categoryID, filters, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range products {
		minPrice, maxPrice, _ := ps.productRepository.GetPriceRangeByProductID(ctx, products[i].ID)
		products[i].Attributes = map[string]interface{}{
			"Minimal Price": minPrice,
			"Maximal Price": maxPrice,
		}
	}
	return products, total, nil
}

func (ps *productService) SelectShopProducts(
	ctx context.Context,
	shopID int,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	products, total, err := ps.productRepository.SelectShopProducts(ctx, shopID, offset, limit)
	if err != nil {
		return nil, 0, err
	}
	for i := range products {
		minPrice, maxPrice, _ := ps.productRepository.GetPriceRangeByProductID(ctx, products[i].ID)
		products[i].Attributes = map[string]interface{}{
			"Minimal Price": minPrice,
			"Maximal Price": maxPrice,
		}
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