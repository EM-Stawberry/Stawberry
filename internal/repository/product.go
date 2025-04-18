package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/zuzaaa-dev/stawberry/internal/domain/service/product"
	"github.com/zuzaaa-dev/stawberry/internal/repository/model"

	"github.com/zuzaaa-dev/stawberry/internal/app/apperror"

	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
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
	/* if err := r.db.WithContext(ctx).Create(productModel).Error; err != nil {
		if isDuplicateError(err) {
			return 0, &apperror.ProductError{
				Code:    apperror.DuplicateError,
				Message: "product with this id already exists",
				Err:     err,
			}
		}
		return 0, &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to create product",
			Err:     err,
		}
	}
 */
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
		return entity.Product{}, &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch product",
			Err:     err,
		}
	}

	return model.ConvertProductToEntity(productModel), nil
}

func (r *productRepository) SelectProducts(
	ctx context.Context,
	offset,
	limit int,
	) ([]entity.Product, int, error) {
	var total int64
	countQuery := "SELECT COUNT(*) FROM products"
	
	if err := r.db.GetContext(ctx, &total, countQuery); err != nil {
		return nil, 0, &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to count products",
			Err:     err,
		}
	}
	query := "SELECT * FROM products LIMIT $1 OFFSET $2"
	var productModels []model.Product
	if err := r.db.SelectContext(ctx, &productModels, query, limit, offset); err != nil {
		return nil, 0, &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch products",
			Err:     err,
		}
	}

	products := make([]entity.Product, len(productModels))
	for i, pm := range productModels {
		products[i] = model.ConvertProductToEntity(pm)
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
		return nil, 0, &apperror.ProductError{
			Code: apperror.NotFound,
			Message: "product not found",
			Err: err,
		}
	}

	var models []model.Product
	searchQuery := `
		SELECT * FROM products 
		WHERE name ILIKE '%' || $1 || '%' 
		LIMIT $2 OFFSET $3`
	if err := r.db.SelectContext(ctx, &models, searchQuery, name, limit, offset); err != nil {
		return nil, 0, &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch products",
			Err:     err,
		}
	}

	products := make([]entity.Product, len(models))
	for i, p := range models {
		products[i] = model.ConvertProductToEntity(p)
	}
	return products, int(total), nil
}

func (r *productRepository) SelectProductsByCategoryID(
	ctx context.Context, 
	categoryID string, 
	offset, 
	limit int,
	) ([]entity.Product, int, error) {
	var models []model.Product
	var total int
	countQuery := `
		WITH RECURSIVE category_tree AS (
			SELECT id FROM categories WHERE id = $1
			UNION
			SELECT c.id FROM categories c
			INNER JOIN category_tree ct ON c.parent_id = ct.id
		)
		SELECT COUNT(*) FROM products WHERE category_id IN (SELECT id FROM category_tree);
		`
	if err := r.db.GetContext(ctx, &total, countQuery, categoryID); err != nil {
		return nil, 0, &apperror.ProductError{
			Code: apperror.NotFound,
			Message: "failed to count products",
			Err: err,
		}
	}
    query := `
        WITH RECURSIVE category_tree AS (
            SELECT id FROM categories WHERE id = $1
            UNION
            SELECT c.id FROM categories c
            INNER JOIN category_tree ct ON c.parent_id = ct.id
        )
        SELECT * FROM products WHERE category_id IN (SELECT id FROM category_tree) LIMIT $2 OFFSET $3;
    `
    if err := r.db.SelectContext(ctx, &models, query, categoryID, limit, offset); err != nil {
        return nil, 0, &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch products",
			Err:     err,
		}
    }
	products := make([]entity.Product, len(models))
	for i, p := range  models{
		products[i] = model.ConvertProductToEntity(p)
	}
    return products, total, nil
	}

func (r *productRepository) SelectStoreProducts(
	ctx context.Context,
	id string, offset, limit int,
) ([]entity.Product, int, error) {
	var total int64
	/* if err := r.db.WithContext(ctx).
		Model(&model.Product{}).
		Where("store_id = ?", id).
		Count(&total).Error; err != nil {
		return nil, 0, &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to count store products",
			Err:     err,
		}
	} */

	var products []entity.Product
	/* if err := r.db.WithContext(ctx).
		Where("store_id = ?", id).
		Offset(offset).Limit(limit).
		Find(&products).Error; err != nil {
		return nil, 0, &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to fetch store products",
			Err:     err,
		}
	} */

	return products, int(total), nil
}

func (r *productRepository) UpdateProduct(
	ctx context.Context,
	id string,
	update product.UpdateProduct,
) error {
	/* updateModel := model.ConvertUpdateProductFromSvc(update)
	tx := r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ?", id).Updates(updateModel)
	if tx.Error != nil {
		if isDuplicateError(tx.Error) {
			return &apperror.ProductError{
				Code:    apperror.DuplicateError,
				Message: "product with these details already exists",
				Err:     tx.Error,
			}
		}
		return &apperror.ProductError{
			Code:    apperror.DatabaseError,
			Message: "failed to update product",
			Err:     tx.Error,
		}
	}

	if tx.RowsAffected == 0 {
		return apperror.ErrProductNotFound
	} */

	return nil
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate") ||
		strings.Contains(err.Error(), "unique violation")
}
