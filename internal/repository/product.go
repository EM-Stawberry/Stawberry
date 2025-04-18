package repository

import (
	"context"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/product"
	"github.com/zuzaaa-dev/stawberry/internal/repository/model"
)

type productRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *productRepository {
	return &productRepository{db: db}
}

func (pr *productRepository) InsertProduct(ctx context.Context, product *product.Product) (uint, error) {

	productModel := model.Product{
		Name:        product.Name,
		Description: product.Description,
		CategoryID:  product.CategoryID,
	}

	SPInventoryModel := model.ShopPointInventory{
		ShopPointID: product.ShopPointID,
		ProductID:   product.ID,
		Price:       product.Price,
		Quantity:    product.Quantity,
	}

	//Начинаем транзакцию
	tx, err := pr.db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	//Отменяем транзакцию если получили ошибку или панику до завершения коммита
	defer func() {
		if r := recover(); r != nil || err != nil {
			tx.Rollback()
		}
	}()

	//Добавляем инфу в таблицу products
	sql, args, err := sq.Insert("products").
		Columns("name", "description", "category_id").
		Values(productModel.Name, productModel.Description, productModel.CategoryID).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build insert product query: %w", err)
	}

	res, err := tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert product: %w", err)
	}

	//Добавляем инфу в таблицу shopPointInventory
	sql, args, err = sq.Insert("shop_point_inventory").
		Columns("shop_point_id", "product_id", "price", "quantity").
		Values(SPInventoryModel.ShopPointID, SPInventoryModel.ProductID, SPInventoryModel.Price, SPInventoryModel.Quantity).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build insert product query: %w", err)

	}

	_, err = tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to insert product: %w", err)
	}

	//Подтверждаем транзакцию
	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint(id), nil
}

func (pr *productRepository) GetProductByID(ctx context.Context, id uint) (*product.Product, error) {

	var productModel model.Product
	productModel.ID = id

	sql, args, err := sq.Select("products").
		Columns("name", "description", "category_id").
		Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build insert product query: %w", err)
	}

	err = pr.db.QueryRowContext(ctx, sql, args...).Scan(&productModel.Name, &productModel.Description, &productModel.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product by id: %w", err)
	}

	product := &product.Product{
		ID:          productModel.ID,
		Name:        productModel.Name,
		Description: productModel.Description,
		CategoryID:  productModel.CategoryID,
	}

	return product, nil
}

func (pr *productRepository) SelectProducts(ctx context.Context, offset, limit int) ([]*product.Product, int, error) {

	products := make([]*product.Product, 0, limit)

	sql, args, err := sq.Select("products").
		Columns("id", "name", "description", "category_id").
		Offset(uint64(offset)).Limit(uint64(limit)).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build insert product query: %w", err)
	}

	rows, err := pr.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to select products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var productModel model.Product
		err = rows.Scan(&productModel)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		product := &product.Product{
			ID:          productModel.ID,
			Name:        productModel.Name,
			Description: productModel.Description,
			CategoryID:  productModel.CategoryID,
		}

		products = append(products, product)

	}

	return products, len(products), nil
}

func (pr *productRepository) SelectStoreProducts(ctx context.Context, id uint, offset, limit int) ([]*product.Product, int, error) {
	products := make([]*product.Product, 0, limit)

	sql, args, err := sq.Select("products").
		Columns("id", "name", "description", "category_id").
		InnerJoin("shop_inventory ON products.id = shop_inventory.product_id").
		Where(sq.Eq{"shop_inventory.shop_id": id}).
		Offset(uint64(offset)).Limit(uint64(limit)).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build insert product query: %w", err)
	}

	rows, err := pr.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to select products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var productModel model.Product
		err = rows.Scan(&productModel)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		product := &product.Product{
			ID:          productModel.ID,
			Name:        productModel.Name,
			Description: productModel.Description,
			CategoryID:  productModel.CategoryID,
		}

		products = append(products, product)

	}

	return products, len(products), nil
}

// ДОДЕЛАТЬ
func (pr *productRepository) UpdateProduct(ctx context.Context, id uint, update *product.UpdateProduct) error {

	var exists bool
	//Проверяем что продукт есть в бд
	err := sq.Select("products").
		Columns("id").
		InnerJoin("shop_point_inventory ON products.id = shop_point_inventory.product_id").
		Where(sq.Eq{"id": id, "shop_point_inventory.shop_point_id": update.ShopPointID}).PlaceholderFormat(sq.Dollar).ScanContext(ctx, &exists)
	if err != nil {
		return fmt.Errorf("failed to check if product exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("product with id %d not found", id)
	}

	//Если продукт есть, обновляем

	tx, err := pr.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil || err != nil {
			tx.Rollback()
		}
	}()

	if update.Name != nil || update.Description != nil || update.CategoryID != nil {

		updateBuilder := sq.Update("products").Where(sq.Eq{"id": id})
		if update.Name != nil {
			updateBuilder = updateBuilder.Set("name", *update.Name)
		}
		if update.Description != nil {
			updateBuilder = updateBuilder.Set("description", *update.Description)
		}
		if update.CategoryID != nil {
			updateBuilder = updateBuilder.Set("category_id", *update.CategoryID)
		}

		sql, args, err := updateBuilder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return fmt.Errorf("failed to build update product query: %w", err)
		}

		_, err = tx.ExecContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to update product: %w", err)
		}
	}

	if update.Price != nil || update.Quantity != nil {
		updateBuilder := sq.Update("shop_point_inventory").Where(sq.Eq{"shop_point_id": update.ShopPointID, "product_id": id})
		if update.Price != nil {
			updateBuilder = updateBuilder.Set("price", *update.Price)
		}
		if update.Quantity != nil {
			updateBuilder = updateBuilder.Set("quantity", *update.Quantity)
		}

		sql, args, err := updateBuilder.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return fmt.Errorf("failed to build update shop_point_inventory query: %w", err)
		}

		_, err = tx.ExecContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to update shop_point_inventory: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (pr *productRepository) DeleteProduct(ctx context.Context, id uint) error {

	sql, args, err := sq.Delete("products").
		Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update product query: %w", err)
	}

	_, err = pr.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

func isDuplicateError(err error) bool {
	return strings.Contains(err.Error(), "duplicate") ||
		strings.Contains(err.Error(), "unique violation")
}
