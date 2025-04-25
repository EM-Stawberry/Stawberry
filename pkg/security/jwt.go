package security

import (
	"fmt"
	"time"

	"github.com/EM-Stawberry/Stawberry/internal/app/apperror"
	"github.com/EM-Stawberry/Stawberry/internal/domain/entity"
	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret        string
	signingMethod jwt.SigningMethod
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		secret:        secret,
		signingMethod: jwt.SigningMethodHS256,
	}
}

func (j *JWTManager) Generate(userID uint, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(j.signingMethod, claims)
	return token.SignedString([]byte(j.secret))
}

func (j *JWTManager) Parse(token string) (entity.AccessToken, error) {
	claim := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claim, func(token *jwt.Token) (any, error) {
		if token.Header["alg"] != j.signingMethod.Alg() {
			return nil, fmt.Errorf("%w: invalid signing method", apperror.ErrInvalidToken)
		}
		return []byte(j.secret), nil
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
