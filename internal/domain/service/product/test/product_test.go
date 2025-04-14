package test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/zuzaaa-dev/stawberry/internal/domain/entity"
	"github.com/zuzaaa-dev/stawberry/internal/domain/service/product"
)

func TestGetProductByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := product.NewMockRepository(ctrl)

	expected := entity.Product{
		ID:    1,
		Name:  "Test Product",
		Description: "Test Description",
		CategoryID: 3,
	}

	mockRepo.
		EXPECT().
		GetProductByID(gomock.Any(), "1").
		Return(expected, nil)

	svc := product.NewProductService(mockRepo)

	result, err := svc.GetProductByID(context.Background(), "1")
	require.NoError(t, err)
	require.Equal(t, expected, result)
}

func TestGetProductByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := product.NewMockRepository(ctrl)

	mockRepo.
		EXPECT().
		GetProductByID(gomock.Any(), "999").
		Return(entity.Product{}, errors.New("not found"))

	svc := product.NewProductService(mockRepo)

	_, err := svc.GetProductByID(context.Background(), "999")
	require.Error(t, err)
}
