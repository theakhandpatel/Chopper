package validator

import "regexp"

var (
	ShortCodeRX = regexp.MustCompile("^[a-zA-Z0-9]+$")
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key string, message string) {
	_, exists := v.Errors[key]
	if !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Check(ok bool, key string, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func (v *Validator) Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
