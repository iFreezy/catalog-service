package http

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/iFreezy/catalog-service/internal/app/entity"
	"github.com/iFreezy/catalog-service/internal/pkg/http/binding"
	"github.com/iFreezy/catalog-service/internal/pkg/http/httph"
	"github.com/iFreezy/catalog-service/internal/pkg/http/respondent"
)

func makeErrorMiddleware() httph.Middleware {
	const (
		_40001 = "Bad request"
		_40002 = "Категория с таким названием уже существует"
		_40003 = "Товар с таким названием уже существует"
		_40401 = "Not found"
		_50001 = "Internal server error"
	)

	const (
		_40401D = "Entity was deleted or never exist"
		_50001D = "Try again later"
	)

	var makeFallbackExtractor = func(status, errorCode int, message, detail string) respondent.ManifestExtractor {
		var genericManifest = respondent.Manifest{
			Status:      status,
			Error:       message,
			ErrorCode:   errorCode,
			ErrorDetail: detail,
		}
		return func(err error) *respondent.Manifest {
			var appErr *entity.AppError
			if errors.As(err, &appErr) {
				return &respondent.Manifest{
					Status:    appErr.HTTPStatus(),
					Error:     appErr.Message,
					ErrorCode: appErr.HTTPStatus(),
				}
			}
			return &genericManifest
		}
	}

	var replacer = respondent.NewSimpleReplacer().
		ReplaceBy(sql.ErrNoRows, entity.ErrNotFound)

	var expander = respondent.NewSimpleExpander().
		ExtractorFor(
			binding.ErrValidationFailed,
			binding.NewRespondentManifestExtractor(http.StatusBadRequest, 40001, _40001)).
		WithoutDetail(entity.ErrCategoryDuplicate, http.StatusConflict, 40002, _40002).
		WithoutDetail(entity.ErrProductDuplicate, http.StatusConflict, 40003, _40003).
		WithoutDetail(entity.ErrIncorrectParameters, http.StatusBadRequest, 40001, _40001).
		WithoutDetail(entity.ErrCategoryHasProducts, http.StatusConflict, 40901, "Category has products").
		WithoutDetail(entity.ErrInvalidReference, http.StatusBadRequest, 40001, _40001).
		WithDetail(entity.ErrNotFound, http.StatusNotFound, 40401, _40401, _40401D).
		FallbackExtractor(makeFallbackExtractor(http.StatusInternalServerError, 50001, _50001, _50001D))

	var applicator = respondent.NewSimpleApplicator()

	return respondent.NewMiddleware(expander, replacer, applicator)
}
