package repository

import (
	"context"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/product"
	"github.com/zuzaaa-dev/stawberry/internal/repository/model"

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

	return 0, nil
}

func (r *productRepository) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {

	var produnilctModel model.Product

	return model.ConvertProductToEntity(produnilctModel), nil
}

func (r *productRepository) SelectProducts(
	ctx context.Context,
	offset,
	limit int,
) ([]entity.Product, int, error) {

	return nil, 0, nil
}

func (r *productRepository) SelectStoreProducts(
	ctx context.Context,
	id string, offset, limit int,
) ([]entity.Product, int, error) {

	return nil, 0, nil
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
