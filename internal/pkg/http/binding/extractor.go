package binding

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entr "github.com/go-playground/validator/v10/translations/en"
	"github.com/iFreezy/catalog-service/internal/pkg/http/respondent"
)

var (
	defaultEnTranslator ut.Translator
)

var (
	ErrMalformedSource  = errors.New("malformed request source")
	ErrValidationFailed = (*validationFailedError)(nil)
)

type validationFailedError struct {
	originalErr validator.ValidationErrors
}

func (e *validationFailedError) Error() string {
	return "Validation failed"
}

func (e *validationFailedError) Is(other error) bool {
	var typed *validationFailedError
	return errors.As(other, &typed)
}

func NewRespondentManifestExtractor(
	status, errorCode int,
	message string,
) respondent.ManifestExtractor {
	return func(err error) *respondent.Manifest {
		manifest := respondent.Manifest{
			Status:    status,
			ErrorCode: errorCode,
			Error:     message,
		}

		var errList validator.ValidationErrors
		if errList1, ok := err.(validator.ValidationErrors); ok {
			errList = errList1
		} else if typedErr, ok := err.(*validationFailedError); ok {
			errList = typedErr.originalErr
		} else {
			return nil
		}

		manifest.ErrorDetails = make([]string, len(errList))
		for i := 0; i < len(errList); i++ {
			if defaultEnTranslator != nil {
				manifest.ErrorDetails[i] = errList[i].Translate(defaultEnTranslator)
			} else {
				manifest.ErrorDetails[i] = errList[i].Error()
			}
		}

		return &manifest
	}
}

func init() {
	v, _ := Validator.Engine().(*validator.Validate)

	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)

	var found bool
	defaultEnTranslator, found = uni.GetTranslator("en")
	if !found {
		panic("EN translator not found")
	}

	if err := entr.RegisterDefaultTranslations(v, defaultEnTranslator); err != nil {
		panic("Failed to register EN translations: " + err.Error())
	}
}
