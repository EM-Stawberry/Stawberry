package repository

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/EM-Stawberry/Stawberry/internal/domain/service/product"
	"github.com/EM-Stawberry/Stawberry/internal/repository/model"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
)

type ProductRepository struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) InsertProduct(
	ctx context.Context,
	product product.Product,
) (uint, error) {

	_ = ctx
	_ = product

	return 0, nil
}

func (r *ProductRepository) GetProductByID(
	ctx context.Context,
	id string,
) (entity.Product, error) {

	var produnilctModel model.Product

	_ = ctx
	_ = id

	return model.ConvertProductToEntity(produnilctModel), nil
}

func (r *ProductRepository) SelectProducts(
	ctx context.Context,
	offset,
	limit int,
) ([]entity.Product, int, error) {

	_ = ctx
	_ = offset
	_ = limit

	return nil, 0, nil
}

func (r *ProductRepository) SelectStoreProducts(
	ctx context.Context,
	id string, offset, limit int,
) ([]entity.Product, int, error) {

	_ = ctx
	_ = id
	_ = limit
	_ = offset

	return nil, 0, nil
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
