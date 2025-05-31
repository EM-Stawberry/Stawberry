package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"database/sql"

	"github.com/jmoiron/sqlx"

	sq "github.com/Masterminds/squirrel"


	"github.com/EM-Stawberry/Stawberry/internal/repository/model"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type ProductRepository struct {
	Db *sqlx.DB
}

func NewProductRepository(Db *sqlx.DB) *ProductRepository {
	return &ProductRepository{Db: Db}
}

// GetProductByID позволяет получить продукт по его ID
func (r *ProductRepository) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	queryBuilder := psql.
		Select("id", "name", "description", "category_id").
		From("products").
		Where(sq.Eq{"id": id})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return entity.Product{}, apperror.New(apperror.DatabaseError, "failed to build SQL query", err)
	}

	var productModel model.Product
	if err := r.Db.GetContext(ctx, &productModel, query, args...); err != nil {
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
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	countBuilder := psql.
		Select("COUNT(*)").
		From("products")

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to build count query", err)
	}

	var total int
	if err := r.Db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}

	selectBuilder := psql.
		Select("*").
		From("products").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	selectQuery, selectArgs, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to build select query", err)
	}

	var productModels []model.Product
	if err := r.Db.SelectContext(ctx, &productModels, selectQuery, selectArgs...); err != nil {
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
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	nameLike := fmt.Sprintf("%%%s%%", name)

	countBuilder := psql.
		Select("COUNT(*)").
		From("products").
		Where(sq.ILike{"name": nameLike})

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to build count query", err)
	}

	var total int64
	if err := r.Db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}

	selectBuilder := psql.
		Select("*").
		From("products").
		Where(sq.ILike{"name": nameLike}).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	selectQuery, selectArgs, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to build select query", err)
	}
	var models []model.Product
	if err := r.Db.SelectContext(ctx, &models, selectQuery, selectArgs...); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	products := make([]entity.Product, len(models))
	for i, pm := range models {
		products[i] = model.ConvertProductToEntity(pm)
	}
	return products, int(total), nil
}

// SelectProductsByCategoryAndAttributes выполняет фильтрацию по ID категории и аттрибутам продукта
func (r *ProductRepository) SelectProductsByFilters(
	ctx context.Context,
	categoryID int,
	filters map[string]interface{},
	offset, limit int,
) ([]entity.Product, int, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)


	selectBuilder := sq.StatementBuilder.
	PlaceholderFormat(sq.Dollar).
	Select("p.id", "p.name", "p.description", "p.category_id").
	From("products p").
	Where("category_id IN (SELECT id FROM category_tree)").
	Limit(uint64(limit)).
	Offset(uint64(offset))

	if len(filters) > 0 {
		selectBuilder = selectBuilder.Join("product_attributes pa ON p.id = pa.product_id")
		for attr, val := range filters {
			condition := fmt.Sprintf("pa.attributes ->> '%s' = ?", attr)
			strVal := fmt.Sprintf("%v", val)
			selectBuilder = selectBuilder.Where(sq.Expr(condition, strVal))
		}
	}

	selectSQL, selectArgs, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to build select query", err)
	}

	selectSQL = shiftPlaceholders(selectSQL, 1)

	recursivePart := `
	WITH RECURSIVE category_tree AS (
		SELECT id FROM categories WHERE id = $1
		UNION
		SELECT c.id FROM categories c
		INNER JOIN category_tree ct ON c.parent_id = ct.id
	)
	`

	fullSQL := recursivePart + selectSQL
	args := append([]interface{}{categoryID}, selectArgs...)

	fmt.Println(fullSQL)
	fmt.Printf("%#v\n", args)

	var models []model.Product
	if err := r.Db.SelectContext(ctx, &models, fullSQL, args...); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	countBuilder := psql.
		Select("COUNT(*)").
		From("products p").
		Where("category_id IN (SELECT id FROM category_tree)")

	if len(filters) > 0 {
		countBuilder = countBuilder.Join("product_attributes pa ON p.id = pa.product_id")
		for attr, val := range filters {
			condition := fmt.Sprintf("pa.attributes ->> '%s' = ?", attr)
			strVal := fmt.Sprintf("%v", val)
			countBuilder = countBuilder.Where(sq.Expr(condition, strVal))
		}
	}

	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to build count query", err)
	}

	countSQL = shiftPlaceholders(countSQL, 1)

	fullCountSQL := recursivePart + countSQL

	countArgs = append([]interface{}{categoryID}, countArgs...)

	var total int
	if err := r.Db.GetContext(ctx, &total, fullCountSQL, countArgs...); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}

	products := make([]entity.Product, len(models))
	for i, pm := range models {
		products[i] = model.ConvertProductToEntity(pm)
	}

	return products, total, nil
}

// SelectShopProducts выполняет фильтрацию по ID магазина
func (r *ProductRepository) SelectShopProducts(
	ctx context.Context,
	shopID int, offset, limit int,
) ([]entity.Product, int, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	countBuilder := psql.
		Select("COUNT(*)").
		From("products p").
		Join("shop_inventory si ON p.id = si.product_id").
		Where(sq.Eq{"si.shop_id": shopID})

	countQuery, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to build count query", err)
	}

	var total int
	if err := r.Db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}

	selectBuilder := psql.
		Select("p.id", "p.name", "p.category_id", "p.description").
		From("products p").
		Join("shop_inventory si ON p.id = si.product_id").
		Where(sq.Eq{"si.shop_id": shopID}).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	selectQuery, selectArgs, err := selectBuilder.ToSql()
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to build select query", err)
	}

	var models []model.Product
	if err := r.Db.SelectContext(ctx, &models, selectQuery, selectArgs...); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	products := make([]entity.Product, len(models))
	for i, pm := range models {
		products[i] = model.ConvertProductToEntity(pm)
	}

	return products, total, nil
}

// GetAttributesByID получает аттрибуты продукта по его ID
func (r *ProductRepository) GetAttributesByID(ctx context.Context, productID string) (map[string]interface{}, error) {
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

func shiftPlaceholders(sql string, offset int) string {
	re := regexp.MustCompile(`\$(\d+)`)
	return re.ReplaceAllStringFunc(sql, func(match string) string {
		num, _ := strconv.Atoi(match[1:])
		return "$" + strconv.Itoa(num+offset)
	})
}