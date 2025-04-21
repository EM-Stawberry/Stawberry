package repository

import (
	"context"
	"fmt"

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
		Price:       product.Price,
		Quantity:    product.Quantity,
	}

	//Проверяем существование продукта
	var id uint

	err := sq.Select("id").
		From("products").
		Where(sq.Eq{"name": productModel.Name, "description": productModel.Description, "category_id": productModel.CategoryID}).
		RunWith(pr.db).
		PlaceholderFormat(sq.Dollar).ScanContext(ctx, &id)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return 0, fmt.Errorf("failed to check if product exists: %w", err)
	}

	if id != 0 {
		return 0, fmt.Errorf("product already exists")
	}

	//Если продукта нет, то создаем
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
		Values(productModel.Name, productModel.Description, productModel.CategoryID).PlaceholderFormat(sq.Dollar).
		Suffix("RETURNING id").ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build insert product query: %w", err)
	}

	err = tx.QueryRowxContext(ctx, sql, args...).Scan(&productModel.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert product: %w", err)
	}

	SPInventoryModel.ProductID = productModel.ID

	//Добавляем инфу в таблицу shop_point_inventory
	sql, args, err = sq.Insert("shop_point_inventory").
		Columns("shop_point_id", "product_id", "price", "quantity").
		Values(SPInventoryModel.ShopPointID,
			SPInventoryModel.ProductID,
			SPInventoryModel.Price,
			SPInventoryModel.Quantity).
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

	return productModel.ID, nil
}

func (pr *productRepository) GetProductByID(ctx context.Context, id uint) (*product.Product, error) {

	var productModel model.Product
	productModel.ID = id

	sql, args, err := sq.Select("name", "description", "category_id").
		From("products").
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

	sql, args, err := sq.Select("id", "name", "description", "category_id").
		From("products").
		Offset(uint64(offset)).Limit(uint64(limit)).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build insert product query: %w", err)
	}

	rows, err := pr.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to select products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var productModel model.Product
		err = rows.StructScan(&productModel)
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

	sql, args, err := sq.Select("id", "name", "description", "category_id").
		From("products").
		InnerJoin("shop_inventory ON products.id = shop_inventory.product_id").
		Where(sq.Eq{"shop_inventory.shop_id": id}).
		Offset(uint64(offset)).Limit(uint64(limit)).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build insert product query: %w", err)
	}

	rows, err := pr.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to select products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var productModel model.Product
		err = rows.StructScan(&productModel)
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

func (pr *productRepository) UpdateProduct(ctx context.Context, id uint, update *product.UpdateProduct) error {

	//Проверяем что продукт есть в бд
	var existingId uint

	err := sq.Select("id").
		From("products").
		InnerJoin("shop_point_inventory ON products.id = shop_point_inventory.product_id").
		Where(sq.Eq{"id": id, "shop_point_inventory.shop_point_id": *update.ShopPointID}).
		RunWith(pr.db).
		PlaceholderFormat(sq.Dollar).ScanContext(ctx, &existingId)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return fmt.Errorf("failed to check if product exists: %w", err)
	}

	if existingId == 0 {
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

	//Проверяем что продукт есть в бд
	var existingId uint

	err := sq.Select("id").
		From("products").
		InnerJoin("shop_point_inventory ON products.id = shop_point_inventory.product_id").
		Where(sq.Eq{"id": id}).
		RunWith(pr.db).
		PlaceholderFormat(sq.Dollar).ScanContext(ctx, &existingId)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return fmt.Errorf("failed to check if product exists: %w", err)
	}

	if existingId == 0 {
		return fmt.Errorf("product with id %d not found", id)
	}

	sql, args, err := sq.Delete("shop_point_inventory").
		Where(sq.Eq{"product_id": id}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete product query: %w", err)
	}

	_, err = pr.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	sql, args, err = sq.Delete("products").
		Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete product query: %w", err)
	}

	_, err = pr.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}
