package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	return &Validator{validate: v}
}

func (v *Validator) Validate(i any) error {
	return v.validate.Struct(i)
}

type ValidationError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

func FormatErrors(err error) []ValidationError {
	var errors []ValidationError
	if verr, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range verr {
			var elem ValidationError
			elem.Field = fe.Field()
			elem.Rule = fe.Tag()
			elem.Param = fe.Param()
			elem.Message = formatErrorMessage(fe)
			errors = append(errors, elem)
		}
	}
	return errors
}

func formatErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("The %s field is required", fe.Field())
	case "gt":
		return fmt.Sprintf("The %s field must be greater than %s", fe.Field(), fe.Param())
	case "lt":
		return fmt.Sprintf("The %s field must be less than %s", fe.Field(), fe.Param())
	case "oneof":
		return fmt.Sprintf("The %s filed must be on of [%s]", fe.Field(), fe.Field())
	default:
		return fmt.Sprintf("Field %s failed validation on rule %s", fe.Field(), fe.Tag())
	}
}
