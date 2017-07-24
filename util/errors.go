package util

import (
	"errors"
	"fmt"
)

func WrapError(message string, cause error) error {
	return fmt.Errorf("%s [caused by: %s]", message, cause.Error())
}

func NewError(message string) error {
	return errors.New(message)
}
