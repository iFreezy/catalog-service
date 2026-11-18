package entity

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

type ErrorKind string

const (
	KindUnknown            ErrorKind = "unknown"
	KindNotFound           ErrorKind = "not_found"
	KindAlreadyExists      ErrorKind = "already_exists"
	KindInvalidArgument    ErrorKind = "invalid_argument"
	KindFailedPrecondition ErrorKind = "failed_precondition"
)

type AppError struct {
	Kind    ErrorKind
	Message string
}

func (e *AppError) Error() string { return e.Message }

func (e *AppError) HTTPStatus() int {
	switch e.Kind {
	case KindNotFound:
		return http.StatusNotFound
	case KindAlreadyExists, KindFailedPrecondition:
		return http.StatusConflict
	case KindInvalidArgument:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func (e *AppError) GRPCCode() codes.Code {
	switch e.Kind {
	case KindNotFound:
		return codes.NotFound
	case KindAlreadyExists:
		return codes.AlreadyExists
	case KindFailedPrecondition:
		return codes.FailedPrecondition
	case KindInvalidArgument:
		return codes.InvalidArgument
	default:
		return codes.Internal
	}
}

func NewAppError(kind ErrorKind, message string) *AppError {
	return &AppError{Kind: kind, Message: message}
}

var (
	ErrNotFound            = NewAppError(KindNotFound, "not found")
	ErrCategoryDuplicate   = NewAppError(KindAlreadyExists, "category duplicate")
	ErrProductDuplicate    = NewAppError(KindAlreadyExists, "product duplicate")
	ErrInvalidReference    = NewAppError(KindInvalidArgument, "invalid reference")
	ErrCategoryHasProducts = NewAppError(KindFailedPrecondition, "category has linked products")
	ErrIncorrectParameters = NewAppError(KindInvalidArgument, "incorrect parameters")
)
