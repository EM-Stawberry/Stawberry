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

func TestSelectProducts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := product.NewMockRepository(ctrl)

	expected := []entity.Product{
		{ID:    1,
		Name:  "Test Product",
		Description: "Test Description",
		CategoryID: 3,
		},
		{ID:    2,
			Name:  "Another Test Product",
			Description: "Another Test Description",
			CategoryID: 9,
		},
		{ID:    3,
			Name:  "One more Test Product",
			Description: "One more Test Description",
			CategoryID: 42,
		},
	}
	expectedTotal := 3

	mockRepo.
		EXPECT().
		SelectProducts(gomock.Any(), 1, 10).
		Return(expected, expectedTotal, nil)

	svc := product.NewProductService(mockRepo)

	result, total, err := svc.SelectProducts(context.Background(), 1, 10)
	require.NoError(t, err)
	require.Len(t, result, 3)
	require.Equal(t, expected, result)
	require.Equal(t, expectedTotal, total)
}

func TestSelectProductsLimitOffset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := product.NewMockRepository(ctrl)
	
	expected := []entity.Product{
		{ID:    2,
			Name:  "Another Test Product",
			Description: "Another Test Description",
			CategoryID: 9,
		},
		{ID:    3,
			Name:  "One more Test Product",
			Description: "One more Test Description",
			CategoryID: 42,
		},
		{ID:    4,
			Name:  "One One more Test Product",
			Description: "One One more Test Description",
			CategoryID: 2,
		},
	}
	expectedTotal := 3

	mockRepo.
		EXPECT().
		SelectProducts(gomock.Any(),1 , 3).
		Return(expected, expectedTotal, nil)

	svc := product.NewProductService(mockRepo)

	result, total, err := svc.SelectProducts(context.Background(), 1, 3)
	require.NoError(t, err)
	require.Len(t, result, 3)
	require.ElementsMatch(t, expected, result)
	require.Equal(t, expectedTotal, total)
}

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
