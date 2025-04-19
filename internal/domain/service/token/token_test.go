package token

import (
	"context"
	"testing"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewTokenService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	secret := "test-secret"
	refreshLife := 24 * time.Hour
	accessLife := time.Hour

	service := NewTokenService(repo, secret, refreshLife, accessLife)

	assert.NotNil(t, service)
	assert.Equal(t, repo, service.tokenRepository)
	assert.Equal(t, secret, service.jwtSecret)
	assert.Equal(t, refreshLife, service.refreshLife)
	assert.Equal(t, accessLife, service.accessLife)
}

func TestTokenService_GenerateTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	service := NewTokenService(repo, "test-secret", 24*time.Hour, time.Hour)

	tests := []struct {
		name        string
		fingerprint string
		userID      uint
		wantErr     bool
	}{
		{
			name:        "Success",
			fingerprint: "test-fingerprint",
			userID:      1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accessToken, refreshToken, err := service.GenerateTokens(context.Background(), tt.fingerprint, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, accessToken)
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
	secret := "test-secret"
	service := NewTokenService(repo, secret, 24*time.Hour, time.Hour)

	validToken, err := generateJWT(1, secret, time.Hour)
	require.NoError(t, err)

	expiredToken, err := generateJWT(1, secret, -time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: nil,
		},
		{
			name:    "Expired token",
			token:   expiredToken,
			wantErr: apperror.ErrInvalidToken,
		},
		{
			name:    "Invalid token",
			token:   "invalid.token.string",
			wantErr: apperror.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	service := NewTokenService(repo, "test-secret", 24*time.Hour, time.Hour)
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
		repo.EXPECT().RevokeActivesByUserID(ctx, uint(1)).Return(nil)
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

func TestTokenService_Parse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := NewMockRepository(ctrl)
	secret := "test-secret"
	service := NewTokenService(repo, secret, 24*time.Hour, time.Hour)

	tests := []struct {
		name    string
		setup   func() string
		wantErr error
	}{
		{
			name: "Valid token",
			setup: func() string {
				token, _ := generateJWT(1, secret, time.Hour)
				return token
			},
			wantErr: nil,
		},
		{
			name: "Invalid signing method",
			setup: func() string {
				token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
					"sub": float64(1),
					"exp": float64(time.Now().Add(time.Hour).Unix()),
					"iat": float64(time.Now().Unix()),
				})
				tokenString, _ := token.SignedString([]byte(secret))
				return tokenString
			},
			wantErr: apperror.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setup()
			accessToken, err := service.parse(token)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, uint(1), accessToken.UserID)
		})
	}
}
