package token

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewTokenService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	jwtMananager := NewMockJWTManager(ctrl)
	refreshLife := 24 * time.Hour
	accessLife := time.Hour

	service := NewService(repo, jwtMananager, refreshLife, accessLife)

	assert.NotNil(t, service)
	assert.Equal(t, repo, service.tokenRepository)
	assert.Equal(t, jwtMananager, service.jwtManager)
	assert.Equal(t, refreshLife, service.refreshLife)
	assert.Equal(t, accessLife, service.accessLife)
}

func TestTokenService_GenerateTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	jwtManager := NewMockJWTManager(ctrl)

	accessLife := time.Hour
	refreshLife := 24 * time.Hour
	service := NewService(repo, jwtManager, refreshLife, accessLife)

	tests := []struct {
		name        string
		fingerprint string
		userID      uint
		mockJWT     string
		mockJWTErr  error
		wantErr     bool
		wantJWTCall bool
	}{
		{
			name:        "Success",
			fingerprint: "test-fingerprint",
			userID:      1,
			mockJWT:     "mock.jwt.token",
			wantJWTCall: true,
			wantErr:     false,
		},
		{
			name:        "JWT Generation Failed",
			fingerprint: "test-fingerprint",
			userID:      1,
			mockJWTErr:  fmt.Errorf("jwt error"),
			wantJWTCall: true,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantJWTCall {
				jwtManager.EXPECT().
					Generate(tt.userID, accessLife).
					Return(tt.mockJWT, tt.mockJWTErr)
			}

			accessToken, refreshToken, err := service.GenerateTokens(context.Background(), tt.fingerprint, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.mockJWT, accessToken)
			assert.NotEmpty(t, refreshToken.UUID)
			assert.Equal(t, tt.fingerprint, refreshToken.Fingerprint)
			assert.Equal(t, tt.userID, refreshToken.UserID)
		})
	}
}

func TestTokenService_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	jwtManager := NewMockJWTManager(ctrl)
	service := NewService(repo, jwtManager, 24*time.Hour, time.Hour)

	validToken := "valid-token"
	expiredToken := "expired-token"
	invalidToken := "invalid.token.string"

	tests := []struct {
		name    string
		token   string
		setup   func()
		wantErr error
	}{
		{
			name:  "Valid token",
			token: validToken,
			setup: func() {
				jwtManager.EXPECT().Parse(validToken).Return(entity.AccessToken{
					UserID:    1,
					IssuedAt:  time.Now(),
					ExpiresAt: time.Now().Add(time.Hour),
				}, nil)
			},
			wantErr: nil,
		},
		{
			name:  "Expired token",
			token: expiredToken,
			setup: func() {
				jwtManager.EXPECT().Parse(expiredToken).Return(entity.AccessToken{
					UserID:    1,
					IssuedAt:  time.Now().Add(-2 * time.Hour),
					ExpiresAt: time.Now().Add(-1 * time.Hour),
				}, nil)
			},
			wantErr: apperror.ErrInvalidToken,
		},
		{
			name:  "Invalid token",
			token: invalidToken,
			setup: func() {
				jwtManager.EXPECT().Parse(invalidToken).Return(entity.AccessToken{}, apperror.ErrInvalidToken)
			},
			wantErr: apperror.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			accessToken, err := service.ValidateToken(context.Background(), tt.token)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, uint(1), accessToken.UserID)
		})
	}
}

func TestTokenService_Repository_Methods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	jwtManager := NewMockJWTManager(ctrl)
	service := NewService(repo, jwtManager, 24*time.Hour, time.Hour)

	ctx := context.Background()

	refreshToken := entity.RefreshToken{
		UUID:        uuid.New(),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		Fingerprint: "test-fingerprint",
		UserID:      1,
	}

	t.Run("InsertToken", func(t *testing.T) {
		repo.EXPECT().InsertToken(ctx, refreshToken).Return(nil)
		err := service.InsertToken(ctx, refreshToken)
		assert.NoError(t, err)
	})

	t.Run("GetActivesTokenByUserID", func(t *testing.T) {
		expected := []entity.RefreshToken{refreshToken}
		repo.EXPECT().GetActivesTokenByUserID(ctx, uint(1)).Return(expected, nil)

		tokens, err := service.GetActivesTokenByUserID(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, expected, tokens)
	})

	t.Run("RevokeActivesByUserID", func(t *testing.T) {
		repo.EXPECT().RevokeActivesByUserID(ctx, uint(1), uint(5)).Return(nil)
		err := service.RevokeActivesByUserID(ctx, 1)
		assert.NoError(t, err)
	})

	t.Run("GetByUUID", func(t *testing.T) {
		repo.EXPECT().GetByUUID(ctx, refreshToken.UUID.String()).Return(refreshToken, nil)
		token, err := service.GetByUUID(ctx, refreshToken.UUID.String())
		assert.NoError(t, err)
		assert.Equal(t, refreshToken, token)
	})

	t.Run("Update", func(t *testing.T) {
		repo.EXPECT().Update(ctx, refreshToken).Return(refreshToken, nil)
		token, err := service.Update(ctx, refreshToken)
		assert.NoError(t, err)
		assert.Equal(t, refreshToken, token)
	})
}
