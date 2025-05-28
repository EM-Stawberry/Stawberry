package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"strings"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type ProductRepository struct {
	Db *sqlx.DB
}

func NewProductRepository(Db *sqlx.DB) *ProductRepository {
	return &ProductRepository{Db: Db}
}

func (r *ProductRepository) InsertProduct(
	ctx context.Context,
	product product.Product,
) (uint, error) {
	productModel := model.ConvertProductFromSvc(product)
	return productModel.ID, nil
}

// GetProductByID позволяет получить продукт по его ID
func (r *ProductRepository) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {
	var productModel model.Product
	query := "SELECT id, name, description, category_id from products WHERE id = $1"
	if err := r.Db.GetContext(ctx, &productModel, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Product{}, apperror.ErrProductNotFound
		}
		return entity.Product{}, apperror.New(apperror.DatabaseError, "failed to fetch product", err)
	}

	return model.ConvertProductToEntity(productModel), nil
}

// SelectProducts выводит весь список продуктов
func (r *ProductRepository) SelectProducts(
	ctx context.Context,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	var total int
	countQuery := "SELECT COUNT(*) FROM products"

	if err := r.Db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}
	query := "SELECT * FROM products LIMIT $1 OFFSET $2"
	var productModels []model.Product
	if err := r.Db.SelectContext(ctx, &productModels, query, limit, offset); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	products := make([]entity.Product, len(productModels))
	for i, pm := range productModels {
		products[i] = model.ConvertProductToEntity(pm)
	}

	return products, total, nil
}

// SelectProductsByName выполняет поиск по имени
func (r *ProductRepository) SelectProductsByName(
	ctx context.Context,
	name string,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	var total int64
	countQuery := `
		SELECT COUNT(*) FROM products 
		WHERE name ILIKE '%' || $1 || '%'`
	if err := r.Db.GetContext(ctx, &total, countQuery, name); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}

	var models []model.Product
	searchQuery := `
		SELECT * FROM products 
		WHERE name ILIKE '%' || $1 || '%' 
		LIMIT $2 OFFSET $3`
	if err := r.Db.SelectContext(ctx, &models, searchQuery, name, limit, offset); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	products := make([]entity.Product, len(models))
	for i, pm := range models {
		products[i] = model.ConvertProductToEntity(pm)
	}
	return products, int(total), nil
}

// SelectProductsByCategoryAndAttributes выполняет фильтрацию по ID категории и аттрибутам продукта
func (r *ProductRepository) SelectProductsByCategoryAndAttributes(
	ctx context.Context,
	categoryID int,
	filters map[string]interface{},
	offset, limit int,
) ([]entity.Product, int, error) {
	var models []model.Product

	var params []interface{}

	params = append(params, categoryID)

	paramIdx := 2

	var joinAttributes bool
	var attrConditions []string

	for attr, val := range filters {
		joinAttributes = true
		attrConditions = append(attrConditions, fmt.Sprintf("pa.attributes ->> '%s' = $%d", attr, paramIdx))
		params = append(params, val)
		paramIdx++
	}

	var query string
	if joinAttributes {
		query = fmt.Sprintf(`
		WITH RECURSIVE category_tree AS (
			SELECT id FROM categories WHERE id = $1
			UNION
			SELECT c.id FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
		)
			SELECT p.id, p.name, p.description, p.category_id
			FROM products p
			JOIN product_attributes pa ON p.id = pa.product_id
			WHERE category_id IN (SELECT id FROM category_tree) AND %s
			LIMIT $%d OFFSET $%d
		`,
			strings.Join(attrConditions, " AND "),
			paramIdx, paramIdx+1,
		)
	} else {
		query = fmt.Sprintf(`
		WITH RECURSIVE category_tree AS (
			SELECT id FROM categories WHERE id = $1
			UNION
			SELECT c.id FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
		)
			SELECT p.id, p.name, p.description, p.category_id
			FROM products p
			WHERE category_id IN (SELECT id FROM category_tree)
			LIMIT $%d OFFSET $%d
		`,
			paramIdx, paramIdx+1,
		)
	}

	params = append(params, offset, limit)

	err := r.Db.SelectContext(ctx, &models, query, params...)
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	var totalCount int
	var countQuery string
	if joinAttributes {
		countQuery = fmt.Sprintf(`
		WITH RECURSIVE category_tree AS (
			SELECT id FROM categories WHERE id = $1
			UNION
			SELECT c.id FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
		)
			SELECT COUNT(*)
			FROM products p
			JOIN product_attributes pa ON p.id = pa.product_id
			WHERE category_id IN (SELECT id FROM category_tree) AND %s
		`,
			strings.Join(attrConditions, " AND "),
		)
	} else {
		countQuery = `
		WITH RECURSIVE category_tree AS (
			SELECT id FROM categories WHERE id = $1
			UNION
			SELECT c.id FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
		)
			SELECT COUNT(*)
			FROM products p
			WHERE category_id IN (SELECT id FROM category_tree)
		`
	}

	err = r.Db.GetContext(ctx, &totalCount, countQuery, params[:paramIdx-1]...)
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}
	products := make([]entity.Product, len(models))
	for i, pm := range models {
		products[i] = model.ConvertProductToEntity(pm)
	}
	return products, totalCount, nil
}

// SelectShopProducts выполняет фильтрацию по ID магазина
func (r *ProductRepository) SelectShopProducts(
	ctx context.Context,
	shopID int, offset, limit int,
) ([]entity.Product, int, error) {
	var total int64
	countQuery := `
		SELECT COUNT(*) FROM products p
		JOIN shop_inventory si ON p.id = si.product_id 
		WHERE si.shop_id = $1 `
	if err := r.Db.GetContext(ctx, &total, countQuery, shopID); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}
	var models []model.Product
	searchQuery := `
		SELECT id, name, category_id, description  FROM products p
		JOIN shop_inventory si ON p.id = si.product_id 
		WHERE si.shop_id = $1
		LIMIT $2 OFFSET $3 `
	if err := r.Db.SelectContext(ctx, &models, searchQuery, shopID, limit, offset); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	products := make([]entity.Product, len(models))
	for i, pm := range models {
		products[i] = model.ConvertProductToEntity(pm)
	}
	return products, int(total), nil
}

func (r *ProductRepository) UpdateProduct(
	ctx context.Context,
	id string,
	update product.UpdateProduct,
) error {

	_ = ctx
	_ = id
	_ = update

	return nil
}

// GetProductAttributesByID получает аттрибуты продукта по его ID
func (r *ProductRepository) GetProductAttributesByID(ctx context.Context, productID string) (map[string]interface{}, error) {
	var attributesJSONb []byte

	query := `SELECT attributes FROM product_attributes WHERE product_id = $1`
	err := r.Db.GetContext(ctx, &attributesJSONb, query, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, apperror.New(apperror.DatabaseError, "failed to fetch product attributes", err)
	}

	var attributes map[string]interface{}
	if err := json.Unmarshal(attributesJSONb, &attributes); err != nil {
		return nil, apperror.New(apperror.DatabaseError, "failed to unmarshal product attributes", err)
	}

	return attributes, nil
}

// GetPriceRangeByProductID получает минимальную и максимальную цену на продукт
func (r *ProductRepository) GetPriceRangeByProductID(ctx context.Context, productID int) (float64, float64, error) {
	var priceRange struct {
		Min sql.NullFloat64 `Db:"min"`
		Max sql.NullFloat64 `Db:"max"`
	}
	query := `SELECT MIN(price) AS min, MAX(price) AS max FROM shop_inventory WHERE product_id = $1`
	err := r.Db.GetContext(ctx, &priceRange, query, productID)
	if err != nil {
		return 0, 0, apperror.New(apperror.DatabaseError, "failed to calculate min/max price", err)
	}
	minPrice := 0.0
	maxPrice := 0.0
	if priceRange.Min.Valid {
		minPrice = priceRange.Min.Float64
	}
	if priceRange.Max.Valid {
		maxPrice = priceRange.Max.Float64
	}

	return minPrice, maxPrice, nil
}

// GetAverageRatingByProductID получает средний рейтинг и количество отзывов на продукт
func (r *ProductRepository) GetAverageRatingByProductID(ctx context.Context, productID int) (float64, int, error) {
	var reviewStats struct {
		Average sql.NullFloat64 `Db:"average"`
		Count   sql.NullInt64   `Db:"count"`
	}
	query := `SELECT AVG(rating) AS average, COUNT(*) AS count FROM product_reviews WHERE product_id = $1`
	err := r.Db.GetContext(ctx, &reviewStats, query, productID)
	if err != nil {
		return 0, 0, apperror.New(apperror.DatabaseError, "failed to calculate average rating/count of reviews", err)
	}
	avg := 0.0
	count := 0
	if reviewStats.Average.Valid {
		avg = reviewStats.Average.Float64
	}
	if reviewStats.Count.Valid {
		count = int(reviewStats.Count.Int64)
	}

	return avg, count, nil
}
