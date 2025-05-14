package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/EM-Stawberry/Stawberry/pkg/security"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockTokenService := NewMockTokenService(ctrl)
	userService := NewUserService(mockRepo, mockTokenService)

	ctx := context.Background()
	testUser := User{
		Email:    "test@example.com",
		Password: "password123",
	}
	fingerprint := "test-fingerprint"

	t.Run("successful user creation", func(t *testing.T) {
		mockRepo.EXPECT().InsertUser(ctx, gomock.Any()).Return(uint(1), nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, uint(1)).Return("access-token", entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(nil)

		accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("failed user insertion", func(t *testing.T) {
		mockRepo.EXPECT().InsertUser(ctx, gomock.Any()).Return(uint(0), errors.New("db error"))

		accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})

	t.Run("token generation failure", func(t *testing.T) {
		mockRepo.EXPECT().InsertUser(ctx, gomock.Any()).Return(uint(1), nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, uint(1)).Return("", entity.RefreshToken{}, errors.New("token generation error"))

		accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})

	t.Run("token insertion failure", func(t *testing.T) {
		mockRepo.EXPECT().InsertUser(ctx, gomock.Any()).Return(uint(1), nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, uint(1)).Return("access-token", entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(errors.New("token insertion error"))

		accessToken, refreshToken, err := userService.CreateUser(ctx, testUser, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})

}

func TestUserService_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockTokenService := NewMockTokenService(ctrl)
	userService := NewUserService(mockRepo, mockTokenService)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	fingerprint := "test-fingerprint"

	hashedPassword, err := security.HashArgon2id(password)
	require.NoError(t, err)

	testUser := entity.User{
		ID:       1,
		Email:    email,
		Password: hashedPassword,
	}

	t.Run("successful authentication", func(t *testing.T) {
		mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
		mockTokenService.EXPECT().GetActivesTokenByUserID(ctx, testUser.ID).Return([]entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, testUser.ID).Return("access-token", entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(nil)

		accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().GetUser(ctx, email).Return(entity.User{}, apperror.ErrUserNotFound)

		accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Equal(t, apperror.ErrUserNotFound, err)
	})

	t.Run("invalid password", func(t *testing.T) {
		mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)

		accessToken, refreshToken, err := userService.Authenticate(ctx, email, "wrong_password", fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Contains(t, err.Error(), "invalid password")
	})

	t.Run("max active tokens reached", func(t *testing.T) {
		maxTokens := make([]entity.RefreshToken, maxUsers)
		mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
		mockTokenService.EXPECT().GetActivesTokenByUserID(ctx, testUser.ID).Return(maxTokens, nil)
		mockTokenService.EXPECT().RevokeActivesByUserID(ctx, testUser.ID).Return(nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, testUser.ID).Return("access-token", entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(nil)

		accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
	})

	t.Run("error getting active tokens", func(t *testing.T) {
		mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
		mockTokenService.EXPECT().GetActivesTokenByUserID(ctx, testUser.ID).
			Return(nil, errors.New("database error"))

		accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})

	t.Run("error revoking active tokens", func(t *testing.T) {
		maxTokens := make([]entity.RefreshToken, maxUsers)
		mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
		mockTokenService.EXPECT().GetActivesTokenByUserID(ctx, testUser.ID).Return(maxTokens, nil)
		mockTokenService.EXPECT().RevokeActivesByUserID(ctx, testUser.ID).
			Return(errors.New("revoke error"))

		accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})

	t.Run("error generating tokens during authentication", func(t *testing.T) {
		mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
		mockTokenService.EXPECT().GetActivesTokenByUserID(ctx, testUser.ID).Return([]entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, testUser.ID).
			Return("", entity.RefreshToken{}, errors.New("token generation error"))

		accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})

	t.Run("error inserting token during authentication", func(t *testing.T) {
		mockRepo.EXPECT().GetUser(ctx, email).Return(testUser, nil)
		mockTokenService.EXPECT().GetActivesTokenByUserID(ctx, testUser.ID).Return([]entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, testUser.ID).
			Return("access-token", entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).
			Return(errors.New("insert error"))

		accessToken, refreshToken, err := userService.Authenticate(ctx, email, password, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
	})
}

func TestUserService_Refresh(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockTokenService := NewMockTokenService(ctrl)
	userService := NewUserService(mockRepo, mockTokenService)

	ctx := context.Background()
	refreshTokenStr := uuid.New().String()
	fingerprint := "test-fingerprint"
	userID := uint(1)

	validRefreshToken := entity.RefreshToken{
		UUID:        uuid.New(),
		ExpiresAt:   time.Now().Add(time.Hour),
		Fingerprint: fingerprint,
		UserID:      userID,
	}

	t.Run("successful token refresh", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
		mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{ID: userID}, nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, userID).Return("new-access-token", entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).Return(nil)

		accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, newRefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		invalidToken := validRefreshToken
		invalidToken.ExpiresAt = time.Now().Add(-time.Hour)

		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(invalidToken, nil)

		accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, newRefreshToken)
	})

	t.Run("invalid fingerprint", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)

		accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, "wrong-fingerprint")

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, newRefreshToken)
		assert.Contains(t, err.Error(), "fingerprints don't match")
	})

	t.Run("user not found during refresh", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
		mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{}, errors.New("user not found"))

		accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, newRefreshToken)
	})

	t.Run("error getting refresh token by UUID", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).
			Return(entity.RefreshToken{}, errors.New("database error"))

		accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, newRefreshToken)
	})

	t.Run("error updating refresh token", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
		mockTokenService.EXPECT().Update(ctx, gomock.Any()).
			Return(entity.RefreshToken{}, errors.New("update error"))

		accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, newRefreshToken)
	})

	t.Run("error generating new tokens during refresh", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
		mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{ID: userID}, nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, userID).
			Return("", entity.RefreshToken{}, errors.New("token generation error"))

		accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, newRefreshToken)
	})

	t.Run("error inserting new refresh token", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
		mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(validRefreshToken, nil)
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{ID: userID}, nil)
		mockTokenService.EXPECT().GenerateTokens(ctx, fingerprint, userID).
			Return("new-access-token", entity.RefreshToken{}, nil)
		mockTokenService.EXPECT().InsertToken(ctx, gomock.Any()).
			Return(errors.New("insert error"))

		accessToken, newRefreshToken, err := userService.Refresh(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Empty(t, accessToken)
		assert.Empty(t, newRefreshToken)
	})
}

func TestUserService_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockTokenService := NewMockTokenService(ctrl)
	userService := NewUserService(mockRepo, mockTokenService)

	ctx := context.Background()
	refreshTokenStr := uuid.New().String()
	fingerprint := "test-fingerprint"

	validRefreshToken := entity.RefreshToken{
		UUID:        uuid.New(),
		ExpiresAt:   time.Now().Add(time.Hour),
		Fingerprint: fingerprint,
	}

	t.Run("successful logout", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
		mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(entity.RefreshToken{}, nil)

		err := userService.Logout(ctx, refreshTokenStr, fingerprint)

		assert.NoError(t, err)
	})

	t.Run("invalid token", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(entity.RefreshToken{}, apperror.ErrInvalidToken)

		err := userService.Logout(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Equal(t, apperror.ErrInvalidToken, err)
	})

	t.Run("invalid fingerprint during logout", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)

		err := userService.Logout(ctx, refreshTokenStr, "wrong-fingerprint")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fingerprints don't match")
	})

	t.Run("token update failure during logout", func(t *testing.T) {
		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(validRefreshToken, nil)
		mockTokenService.EXPECT().Update(ctx, gomock.Any()).Return(entity.RefreshToken{}, errors.New("update error"))

		err := userService.Logout(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to revoke refresh token")
	})

	t.Run("expired token during logout", func(t *testing.T) {
		expiredToken := entity.RefreshToken{
			UUID:        uuid.New(),
			ExpiresAt:   time.Now().Add(-time.Hour), // Token expired an hour ago
			Fingerprint: fingerprint,
		}

		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(expiredToken, nil)

		err := userService.Logout(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Equal(t, apperror.ErrInvalidToken, err)
	})

	t.Run("revoked token during logout", func(t *testing.T) {
		revokedTime := time.Now().Add(-time.Hour)
		revokedToken := entity.RefreshToken{
			UUID:        uuid.New(),
			ExpiresAt:   time.Now().Add(time.Hour), // Token not expired
			RevokedAt:   &revokedTime,              // But token is revoked
			Fingerprint: fingerprint,
		}

		mockTokenService.EXPECT().GetByUUID(ctx, refreshTokenStr).Return(revokedToken, nil)

		err := userService.Logout(ctx, refreshTokenStr, fingerprint)

		assert.Error(t, err)
		assert.Equal(t, apperror.ErrInvalidToken, err)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepository(ctrl)
	mockTokenService := NewMockTokenService(ctrl)
	userService := NewUserService(mockRepo, mockTokenService)

	ctx := context.Background()
	userID := uint(1)
	expectedUser := entity.User{
		ID:    userID,
		Email: "test@example.com",
	}

	t.Run("successful get user", func(t *testing.T) {
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(expectedUser, nil)

		user, err := userService.GetUserByID(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().GetUserByID(ctx, userID).Return(entity.User{}, apperror.ErrUserNotFound)

		user, err := userService.GetUserByID(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, entity.User{}, user)
		assert.Equal(t, apperror.ErrUserNotFound, err)
	})
}
