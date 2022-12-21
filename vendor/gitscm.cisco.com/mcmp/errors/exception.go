package errors

import (
	"encoding"

	wraperrors "github.com/pkg/errors"
)

// Exception defines a common interface for errors.
type Exception interface {
	error
	encoding.BinaryMarshaler
	Type() ErrorType
}

// Convert handls conversion from an error of type Exception to any BinaryUnmarshaler.
func Convert(in error, out encoding.BinaryUnmarshaler) error {
	ex, ok := in.(encoding.BinaryMarshaler)
	if !ok {
		return wraperrors.Wrap(in, "provided error does not implement encoding.BinaryMarshaler interface")
	}

	data, err := ex.MarshalBinary()
	if err != nil {
		return err
	}

	return out.UnmarshalBinary(data)
}

// IsType determines if an error is an Exception of a specific ErrorType.
func IsType(errType ErrorType, err error) bool {
	e, ok := err.(Exception)
	if !ok {
		return false
	}

	return e.Type() == errType
}
