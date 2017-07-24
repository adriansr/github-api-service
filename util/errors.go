package util

import (
	"errors"
	"fmt"
)

// WrapError decorates an existing error (possibly returned by a 3rd party
// library) with the given message, for context
func WrapError(message string, cause error) error {
	return fmt.Errorf("%s [caused by: %s]", message, cause.Error())
}

// NewError creates an error that prints the given string
func NewError(message string) error {
	return errors.New(message)
}
