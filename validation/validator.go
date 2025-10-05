package validation

import (
	"fmt"
	"strings"

	"github.com/FateevDev/orders-api/model"
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

	Validate.RegisterValidation("order_status", validateStatus)

	// Register custom translation for the new rule
	Validate.RegisterTranslation("order_status", Trans, func(ut ut.Translator) error {
		return ut.Add("order_status", "{0} must be a valid status", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("order_status", fe.Field())
		t += fmt.Sprintf(" (one of %s)", strings.Join(model.AllStatusesStrings(), ", "))
		return t
	})
}

func validateStatus(fl validator.FieldLevel) bool {
	status := fl.Field().String()
	_, err := model.ParseStatus(status)
	return err == nil
}
