package errors

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Domain represents the different error domains.
type Domain string

// Defined domains.
const (
	Default Domain = "API"
)

// ErrorType represents all different types of errors.
type ErrorType int

// All of the defined errors.
const (
	ErrGeneric ErrorType = iota
	ErrInvalid
	ErrRequired
	ErrExists
	ErrNotFound
	ErrPreconditionNotMet
	ErrUnauthorized
	ErrInsufficientFunds
	ErrNoChange
	ErrUnavailable
)

var messages = map[ErrorType]string{
	ErrGeneric:            "generic error: %v",
	ErrInvalid:            "invalid %q expected %s",
	ErrRequired:           "missing required %q",
	ErrExists:             "%s already exists",
	ErrNotFound:           "%s not found",
	ErrPreconditionNotMet: "precondition not met %q",
	ErrUnauthorized:       "unauthorized to perform operation",
	ErrInsufficientFunds:  "insufficient funds",
	ErrNoChange:           "no changes detected for update",
	ErrUnavailable:        "service is currently unavailable",
}

type domainError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	rc      ErrorType
}

func (e domainError) Error() string {
	return e.Message
}

func (e domainError) MarshalBinary() (data []byte, err error) {
	return json.Marshal(e)
}

func (e domainError) Type() ErrorType {
	return e.rc
}

// NewDomainError creates a new error based on a specific ErrorType within a specific Domain and formats the message using the provided args.
func NewDomainError(errType ErrorType, domain Domain, args ...interface{}) error {
	if domain == "" {
		domain = Default
	}

	if _, ok := messages[errType]; !ok {
		errType = ErrGeneric
	}

	return domainError{
		rc:      errType,
		Code:    fmt.Sprintf("%s-%03d", domain, errType),
		Message: fmt.Sprintf(messages[errType], args...),
	}
}

// RestoreDomain converts an error returned by a client into a Domain error.
func RestoreDomain(e error) error {
	// if the error is a pre-defined API error response which means it should have a body
	// that holds the Error struct. If that is true convert to a Domain error.
	val := reflect.Indirect(reflect.ValueOf(e))
	if val.Kind() == reflect.Struct {
		if p := val.FieldByName("Payload"); p.IsValid() && !p.IsNil() {
			pval := reflect.Indirect(p)
			code := pval.FieldByName("Code").Interface().(*string)
			msg := pval.FieldByName("Message").Interface().(*string)

			if code != nil && msg != nil {
				return domainError{
					rc:      codeToErrorType(*code),
					Code:    *code,
					Message: *msg,
				}
			}
		}
	}

	return e
}

func codeToErrorType(code string) ErrorType {
	parts := strings.Split(code, "-")
	if len(parts) == 2 {
		et, _ := strconv.Atoi(parts[1])

		return ErrorType(et)
	}

	return ErrGeneric
}
