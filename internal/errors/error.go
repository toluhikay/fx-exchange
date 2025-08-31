package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/jackc/pgconn"
)

const (
	ForeignKeyViolation = "23503"
	UniqueViolation     = "23505"
)

var (
	ErrRecordNotFound        = errors.New("record not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUnauthorized          = errors.New("unauthorized access")
	ErrExpiredToken          = errors.New("token has expired")
	ErrInvalidOtp            = errors.New("invalid otp")
	ErrInvalidPayload        = errors.New("invalid payload")
	ErrInternalServer        = errors.New("internal server error")
	ErrDuplicateEmail        = errors.New("this email is already registered, please sign in or use a different one")
	ErrDuplicateKey          = errors.New("record already exists")
	ErrDuplicateVerification = errors.New("record already exists")
	ErrPasswordTooShort      = errors.New("password must be at least 8 characters long")
	ErrPinLength             = errors.New("pin must be 4 numbers")
	ErrPasswordTooLong       = errors.New("password must not exceed 32 characters")
	ErrPasswordMismatch      = errors.New("passwords must match")
	ErrSendingOtp            = errors.New("error sending otp")
	ErrEmptyRequest          = errors.New("request cant be null or empty")
	ErrDoesNtExists          = errors.New("doesn't exists")
	ErrExists                = errors.New("already exists")
)

func ErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
func ConcatinateErrorMessage(value string, err error) error {
	return fmt.Errorf("%s %w", value, err)
}
func ResolveHTTPStatus(err error) int {

	switch {
	case errors.Is(err, ErrRecordNotFound):
		return http.StatusBadRequest
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusBadRequest
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrExpiredToken):
		return http.StatusUnauthorized
	case errors.Is(err, ErrExpiredToken):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidOtp):
		return http.StatusBadRequest
	case errors.Is(err, ErrInvalidPayload):
		return http.StatusBadRequest
	case errors.Is(err, ErrPasswordTooShort), errors.Is(err, ErrPasswordTooLong):
		return http.StatusBadRequest
	case errors.Is(err, ErrInternalServer):
		return http.StatusInternalServerError
	case errors.Is(err, ErrDuplicateEmail):
		return http.StatusBadRequest
	case errors.Is(err, ErrDuplicateEmail):
		return http.StatusBadRequest
	case errors.Is(err, ErrPasswordMismatch):
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

// func GetSQLErrorCode(err error) string {
// 	if err == nil {
// 		return ""
// 	}

// 	if sqlErr, ok := err.Error(); ok {
// 		return sqlErr.Number
// 	}
// 	return ""
// }
