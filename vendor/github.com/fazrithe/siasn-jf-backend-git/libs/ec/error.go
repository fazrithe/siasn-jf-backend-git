package ec

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const UnspecifiedErrorCode = 0

// CodeUserDefault is a variable and can be changed. CodeUserDefault is returned by Wrap when it retrieves an error
// that is not ec.Error instance. It indicates a generic error.
var CodeUserDefault = UnspecifiedErrorCode

// Error represent an error that wraps an another error along with a `Code` to differentiate the kind of errors.
type Error struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Cause   error       `json:"cause,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *Error) UnmarshalJSON(bytes []byte) error {
	type rawError struct {
		Message string      `json:"message"`
		Code    int         `json:"code"`
		Cause   string      `json:"cause"`
		Data    interface{} `json:"data"`
	}

	re := &rawError{}
	err := json.Unmarshal(bytes, re)
	if err != nil {
		return err
	}

	e.Message = re.Message
	e.Code = re.Code
	if re.Cause != "" {
		e.Cause = fmt.Errorf(re.Cause)
	}
	e.Data = re.Data
	return nil
}

func (e *Error) MarshalJSON() ([]byte, error) {
	p := map[string]interface{}{
		"message": e.Message,
		"code":    e.Code,
	}
	if e.Cause != nil && reflect.ValueOf(e.Cause).IsValid() {
		p["cause"] = e.Cause.Error()
	}
	if e.Data != nil && reflect.ValueOf(e.Data).IsValid() {
		p["data"] = e.Data
	}
	return json.Marshal(p)
}

// NewError create a new Error where the `Code` can be specified.
func NewError(code int, message string, cause error) *Error {
	return &Error{Message: message, Code: code, Cause: cause}
}

// NewErrorBasic creates a new Error without a cause.
func NewErrorBasic(code int, message string) *Error {
	return &Error{Message: message, Code: code}
}

func (e *Error) Error() string {
	builder := &strings.Builder{}
	builder.WriteRune('[')
	builder.WriteString(strconv.Itoa(e.Code))
	builder.WriteString("] ")
	builder.WriteString(e.Message)
	if e.Cause != nil {
		builder.WriteString(": ")
		builder.WriteString(e.Cause.Error())
	}
	return builder.String()
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// Wrap wraps an instance of error as a generic ec.Error.
// If the given err is nil, it will return nil. If the given err is already an instance of *ec.Error,
// it will return the value as is. Otherwise, It will be wrapped as a generic error code.
//
// It is preferable to always call Wrap to wrap errors from functions that you do not control, or do not control
// completely. This ensures that all errors that the client/user retrieves always have error codes, making sure that
// all errors can be handled correctly.
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}

	// It's already an instance of errorcode.Error
	var e *Error
	if errors.As(err, &e) {
		return e
	}

	return NewError(CodeUserDefault, "generic error", err)
}
