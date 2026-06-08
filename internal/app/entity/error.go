package entity

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrCategoryDuplicate   = errors.New("category duplicate")
	ErrProductDuplicate    = errors.New("product duplicate")
	ErrInvalidReference    = errors.New("invalid reference")
	ErrCategoryHasProducts = errors.New("category has linked products")
	ErrIncorrectParameters = errors.New("incorrect parameters")
)
