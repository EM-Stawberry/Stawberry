package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"encoding/json"
	"strconv"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"

	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *productRepository {
	return &productRepository{db: db}
}

func (r *productRepository) InsertProduct(
	ctx context.Context,
	product product.Product,
) (uint, error) {
	productModel := model.ConvertProductFromSvc(product)
	return productModel.ID, nil
}

func (r *productRepository) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {
	var productModel model.Product
	query := "SELECT id, name, description, category_id from products WHERE id = $1"
	if err := r.db.GetContext(ctx, &productModel, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Product{}, apperror.ErrProductNotFound
		}
		return entity.Product{}, apperror.New(apperror.DatabaseError, "failed to fetch product", err)
	}
	idAttr, _ := strconv.Atoi(id)
	idUint := uint(idAttr)
	attributes, err := r.GetProductAttributesByID(ctx, idUint)
	if err != nil {
		return entity.Product{}, err
	}
	product := model.ConvertProductToEntity(productModel)
	product.Attributes = attributes

	return product, nil
}

func (r *productRepository) SelectProducts(
	ctx context.Context,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	var total int64
	countQuery := "SELECT COUNT(*) FROM products"

	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}
	query := "SELECT * FROM products LIMIT $1 OFFSET $2"
	var productModels []model.Product
	if err := r.db.SelectContext(ctx, &productModels, query, limit, offset); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	products := make([]entity.Product, len(productModels))
	for i, pm := range productModels {
		products[i] = model.ConvertProductToEntity(pm)
		attributes, err := r.GetProductAttributesByID(ctx, pm.ID)
		if err != nil {
			return nil, 0, err
		}
		products[i].Attributes = attributes
	}

	return products, int(total), nil
}

func (r *productRepository) SelectProductsByName(
	ctx context.Context,
	name string,
	offset,
	limit int,
) ([]entity.Product, int, error) {
	var total int64
	countQuery := `
		SELECT COUNT(*) FROM products 
		WHERE name ILIKE '%' || $1 || '%'`
	if err := r.db.GetContext(ctx, &total, countQuery, name); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}

	var models []model.Product
	searchQuery := `
		SELECT * FROM products 
		WHERE name ILIKE '%' || $1 || '%' 
		LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &models, searchQuery, name, limit, offset); err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to fetch products", err)
	}

	products := make([]entity.Product, len(models))
	for i, pm := range models {
		products[i] = model.ConvertProductToEntity(pm)
		attributes, err := r.GetProductAttributesByID(ctx, pm.ID)
		if err != nil {
			return nil, 0, err
		}
		products[i].Attributes = attributes
	}
	return products, int(total), nil
}

func (r *productRepository) SelectProductsByCategoryAndAttributes(
	ctx context.Context,
	categoryID int,
	filters map[string]interface{},
	limit, offset int,
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

	params = append(params, limit, offset)

	err := r.db.SelectContext(ctx, &models, query, params...)
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

	err = r.db.GetContext(ctx, &totalCount, countQuery, params[:paramIdx-1]...)
	if err != nil {
		return nil, 0, apperror.New(apperror.DatabaseError, "failed to count products", err)
	}
	products := make([]entity.Product, len(models))
	for i, pm := range models {
		products[i] = model.ConvertProductToEntity(pm)
		attributes, err := r.GetProductAttributesByID(ctx, pm.ID)
		if err != nil {
			return nil, 0, err
		}
		products[i].Attributes = attributes
	}

	return products, totalCount, nil
}

func (r *productRepository) SelectStoreProducts(
	ctx context.Context,
	id string, offset, limit int,
) ([]entity.Product, int, error) {
	var total int64

	var products []entity.Product

	return products, int(total), nil
}

func (r *productRepository) UpdateProduct(
	ctx context.Context,
	id string,
	update product.UpdateProduct,
) error {

	return nil
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate") ||
		strings.Contains(err.Error(), "unique violation")
}

func (r *productRepository) GetProductAttributesByID(ctx context.Context, productID uint) (map[string]interface{}, error) {
	var attributesJSONb []byte

	query := `SELECT attributes FROM product_attributes WHERE product_id = $1`
	err := r.db.GetContext(ctx, &attributesJSONb, query, productID)
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
