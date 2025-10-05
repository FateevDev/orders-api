package validation

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	Validate *validator.Validate
	Trans    ut.Translator
)

func init() {
	Validate = validator.New()

	english := en.New()
	uni := ut.New(english, english)
	Trans, _ = uni.GetTranslator("en")

	// Register default translations for all the common tags.
	en_translations.RegisterDefaultTranslations(Validate, Trans)
}
