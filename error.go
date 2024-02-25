package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"io"
	"unsafe"
)

// JSError implements a JS error.
type JSError struct {
	Message     string
	Filename    string
	LineNumber  int
	ErrorNumber int
}

// newJSError creates a new error.
func newJSError(e C.Error) *JSError {
	err := &JSError{
		Message:     C.GoString(e.message),
		Filename:    C.GoString(e.filename),
		LineNumber:  int(e.lineno),
		ErrorNumber: int(e.number),
	}
	C.free(unsafe.Pointer(e.message))
	return err
}

// Error returns the error message.
func (e *JSError) Error() string {
	return e.Message
}

// Format implements fmt.Formatter.
func (e *JSError) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		fallthrough
	case 's':
		_, _ = io.WriteString(f, e.Error())
	case 'q':
		fmt.Fprintf(f, "%q", e.Error())
	}
}

var _ error = (*JSError)(nil)
var _ fmt.Formatter = (*Value)(nil)
