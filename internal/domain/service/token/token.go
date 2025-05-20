package token

import (
	"context"
	"fmt"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/google/uuid"

	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/golang-jwt/jwt"
)

var signingMethod = jwt.SigningMethodHS256

//go:generate mockgen -source=$GOFILE -destination=token_mock_test.go -package=token Repository

type Repository interface {
	InsertToken(ctx context.Context, token entity.RefreshToken) error
	GetActivesTokenByUserID(ctx context.Context, userID uint) ([]entity.RefreshToken, error)
	RevokeActivesByUserID(ctx context.Context, userID uint) error
	GetByUUID(ctx context.Context, uuid string) (entity.RefreshToken, error)
	Update(ctx context.Context, refresh entity.RefreshToken) (entity.RefreshToken, error)
}

type Service struct {
	tokenRepository Repository
	jwtSecret       string
	refreshLife     time.Duration
	accessLife      time.Duration
}

func NewService(tokenRepo Repository, secret string, refreshLife, accessLife time.Duration) *Service {
	return &Service{
		tokenRepository: tokenRepo,
		jwtSecret:       secret,
		refreshLife:     refreshLife,
		accessLife:      accessLife,
	}
}

// GenerateTokens генерирует новый токен доступа и токен обновления для пользователя.
func (ts *Service) GenerateTokens(
	ctx context.Context,
	fingerprint string,
	userID uint,
) (string, entity.RefreshToken, error) {

	if ctx.Err() != nil {
		return "", entity.RefreshToken{}, ctx.Err()
	}

	accessToken, err := generateJWT(userID, ts.jwtSecret, ts.accessLife)
	if err != nil {
		return "", entity.RefreshToken{}, err
	}

	if ctx.Err() != nil {
		return "", entity.RefreshToken{}, ctx.Err()
	}

	entityRefreshToken, err := generateRefresh(fingerprint, userID, ts.refreshLife)
	if err != nil {
		return "", entity.RefreshToken{}, err
	}

	return accessToken, entityRefreshToken, nil
}

// ValidateToken проверяет access токен и возвращает расшифрованную информацию, если она действительна.
func (ts *Service) ValidateToken(
	ctx context.Context,
	token string,
) (entity.AccessToken, error) {

	if ctx.Err() != nil {
		return entity.AccessToken{}, ctx.Err()
	}

	accessToken, err := ts.parse(token)
	if err != nil {
		return entity.AccessToken{}, err
	}

	if time.Now().After(accessToken.ExpiresAt) {
		return entity.AccessToken{}, apperror.ErrInvalidToken
	}

	return accessToken, nil
}

// parse извлекает токен JWT и извлекает claims.
func (ts *Service) parse(token string) (entity.AccessToken, error) {
	claim := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claim, func(token *jwt.Token) (any, error) {
		if token.Header["alg"] != signingMethod.Alg() {
			return nil, fmt.Errorf("%w: invalid signing method", apperror.ErrInvalidToken)
		}
		return []byte(ts.jwtSecret), nil
	})
	if err != nil {
		return entity.AccessToken{}, apperror.ErrInvalidToken
	}
	userID, ok := claim["sub"].(float64)
	if !ok {
		return entity.AccessToken{}, apperror.ErrInvalidToken
	}

	unixExpiresAt, ok := claim["exp"].(float64)
	if !ok {
		return entity.AccessToken{}, apperror.ErrInvalidToken
	}
	expiresAt := time.Unix(int64(unixExpiresAt), 0)

	unixIssuedAt, ok := claim["iat"].(float64)
	if !ok {
		return entity.AccessToken{}, apperror.ErrInvalidToken
	}

	issuedAt := time.Unix(int64(unixIssuedAt), 0)

	return entity.AccessToken{
		UserID:    uint(userID),
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}, nil
}

func (ts *Service) InsertToken(
	ctx context.Context,
	token entity.RefreshToken,
) error {
	return ts.tokenRepository.InsertToken(ctx, token)
}

// GetActivesTokenByUserID извлекает все активные токены обновления для конкретного пользователя.
func (ts *Service) GetActivesTokenByUserID(
	ctx context.Context,
	userID uint,
) ([]entity.RefreshToken, error) {
	return ts.tokenRepository.GetActivesTokenByUserID(ctx, userID)
}

// RevokeActivesByUserID аннулирует все активные токены обновления для определенного пользователя.
func (ts *Service) RevokeActivesByUserID(
	ctx context.Context,
	userID uint,
) error {
	return ts.tokenRepository.RevokeActivesByUserID(ctx, userID)
}

func (ts *Service) GetByUUID(
	ctx context.Context,
	uuid string,
) (entity.RefreshToken, error) {
	return ts.tokenRepository.GetByUUID(ctx, uuid)
}

func (ts *Service) Update(
	ctx context.Context,
	refresh entity.RefreshToken,
) (entity.RefreshToken, error) {
	return ts.tokenRepository.Update(ctx, refresh)
}

// generateJWT создает новый токен доступа JWT с указанным userID и сроком действия.
func generateJWT(userID uint, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// generateRefresh создает новый refresh токен обновления с указанным
// userID, fingerprint и сроком действия.
func generateRefresh(fingerprint string, userID uint, refreshLife time.Duration) (entity.RefreshToken, error) {
	now := time.Now()

	refreshUUID, err := uuid.NewRandom()
	if err != nil {
		return entity.RefreshToken{}, err
	}

	return entity.RefreshToken{
		UUID:        refreshUUID,
		CreatedAt:   now,
		ExpiresAt:   now.Add(refreshLife),
		RevokedAt:   nil,
		Fingerprint: fingerprint,
		UserID:      userID,
	}, nil
}
