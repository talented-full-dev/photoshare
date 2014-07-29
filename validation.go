package photoshare

import (
	"regexp"
)

var emailRegex = regexp.MustCompile(".+@.+\\..+")

type validationFailure struct {
	Errors map[string]string `json:"errors"`
}

func (f validationFailure) Error() string {
	return "Validation failure"
}

type validator interface {
	validate(*context, map[string]string) error
}

func validateEmail(email string) bool {
	return emailRegex.Match([]byte(email))
}
