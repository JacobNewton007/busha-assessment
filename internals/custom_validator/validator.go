package validator

import (
	"fmt"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)	

func Validator() ut.Translator{
	validate := validator.New()
	english := en.New()
	uni := ut.New(english, english)
	trans, _ := uni.GetTranslator("en")
	_ = en_translations.RegisterDefaultTranslations(validate, trans)
	return trans
}



func TranslateError(err error, trans ut.Translator) (errs []error) {
  if err == nil {
    return nil
  }
  validatorErrs := err.(validator.ValidationErrors)
  for _, e := range validatorErrs {
    translatedErr := fmt.Errorf(e.Translate(trans))
    errs = append(errs, translatedErr)
  }
  return errs
}