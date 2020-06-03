package error

import (
	"runtime"
	"strings"
)

//
// Wrap an error.
// Returns `err` when err is `nil` or *LibError.
func Wrap(err error) error {
	if err == nil {
		return err
	}
	if le, cast := err.(*LibError); cast {
		return le
	}
	bfr := make([]byte, 1<<22)
	len := runtime.Stack(bfr, false)
	return &LibError{
		stack:   bfr[:len],
		wrapped: err,
	}
}

//
// Lib error.
type LibError struct {
	// Original error.
	wrapped error
	// Stack.
	stack []byte
}

//
// Error description.
func (e LibError) Error() string {
	return e.wrapped.Error()
}

//
// Error stack trace.
func (e LibError) Stack() string {
	return strings.Join(
		[]string{
			e.wrapped.Error(),
			string(e.stack),
		},
		"\n")
}

//
// Unwrap the error.
func (e LibError) Unwrap() error {
	return e.wrapped
}
