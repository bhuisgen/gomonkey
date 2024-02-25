package gomonkey

// #include "gomonkey.h"
// #include <stdint.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// JSONParse decodes a JSON-encoded string to a JS object.
func JSONParse(c *Context, str string) (*Object, error) {
	cData := C.CString(str)
	result := C.JSONParse(c.ptr, cData)
	C.free(unsafe.Pointer(cData))
	if !result.ok {
		err := fmt.Errorf("JSON parse: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Object{&Value{result.ptr, c}}, nil
}

// JSONStringify encodes a JS value to a JSON-encoded string.
func JSONStringify(c *Context, v Valuer) (string, error) {
	result := C.JSONStringify(c.ptr, v.AsValue().ptr)
	if !result.ok {
		err := fmt.Errorf("JSON stringify: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return "", err
	}
	str := C.GoStringN(result.data, result.len)
	C.free(unsafe.Pointer(result.data))
	return str, nil
}
