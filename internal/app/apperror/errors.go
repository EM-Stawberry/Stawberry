package apperror

import (
	"fmt"
)

const (
	NotFound           = "NOT_FOUND"
	DatabaseError      = "DATABASE_ERROR"
	InternalError      = "INTERNAL_ERROR"
	DuplicateError     = "DUPLICATE_ERROR"
	BadRequest         = "BAD_REQUEST"
	Unauthorized       = "UNAUTHORIZED"
	InvalidToken       = "INVALID_TOKEN"
	InvalidFingerprint = "INVALID_FINGERPRINT"
	Conflict           = "CONFLICT"
)

type AppError interface {
	error
	Code() string
	Message() string
	Unwrap() error
}

type Error struct {
	ErrCode    string
	ErrMsg     string
	WrappedErr error
}

func (e *Error) Error() string {
	if e.WrappedErr != nil {
		return fmt.Sprintf("%s: %v", e.ErrMsg, e.WrappedErr)
	}
	return e.ErrMsg
}
func (e *Error) Code() string    { return e.ErrCode }
func (e *Error) Message() string { return e.ErrMsg }
func (e *Error) Unwrap() error   { return e.WrappedErr }

func New(code, message string, cause error) *Error {
	return &Error{ErrCode: code, ErrMsg: message, WrappedErr: cause}
}

var (
	ErrProductNotFound = New(NotFound, "product not found", nil)
	ErrStoreNotFound   = New(NotFound, "store not found", nil)

	ErrOfferNotFound = New(NotFound, "offer not found", nil)

	ErrUserNotFound             = New(NotFound, "user not found", nil)
	ErrIncorrectPassword        = New(Unauthorized, "incorrect password", nil)
	ErrFailedToGeneratePassword = New(InternalError, "failed to generate password", nil)
	ErrInvalidFingerprint       = New(InvalidFingerprint, "fingerprints don't match", nil)

	ErrInvalidToken  = New(InvalidToken, "invalid token", nil)
	ErrTokenNotFound = New(NotFound, "token not found", nil)

	ErrNotificationNotFound = New(NotFound, "notification not found", nil)
)
