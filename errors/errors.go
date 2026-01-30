package errors

import (
	"fmt"
	"strings"
)

type LanguageNotFoundError struct {
	Arg       string
	Message   string
	Supported []string
}

func (e *LanguageNotFoundError) Error() string {
	return fmt.Sprintf(
		"Arg %s could not be Found : %s. Supported Types are : [%s]",
		e.Arg, e.Message, strings.Join(e.Supported, ","),
	)
}

func NewLanguageNotSupportedError(lang string) *LanguageNotFoundError {
	return &LanguageNotFoundError{
		Arg:       lang,
		Message:   "Language not supported",
		Supported: make([]string, 0),
	}
}
