package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"unsafe"
)

// ArrayObject implements a JS array object.
type ArrayObject struct {
	v *Value
}

// NewArrayObject creates a new array.
func NewArrayObject(ctx *Context, values ...*Value) (*ArrayObject, error) {
	argc := len(values)
	var argv *C.ValuePtr
	if argc > 0 {
		cArgs := make([]C.ValuePtr, argc)
		for i, arg := range values {
			cArgs[i] = arg.ptr
		}
		argv = (*C.ValuePtr)(unsafe.Pointer(&cArgs[0]))
	}
	result := C.NewArrayObject(ctx.ptr, C.int(argc), argv)
	if !result.ok {
		err := errors.New(C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &ArrayObject{&Value{result.ptr, ctx}}, nil
}

// Release releases the array.
func (o *ArrayObject) Release() {
	C.ReleaseValue(o.v.ptr)
}

// Length returns the array length.
func (o *ArrayObject) Length() uint {
	result := C.GetArrayObjectLength(o.v.ptr)
	if !result.ok {
		C.free(unsafe.Pointer(result.err.message))
		return 0
	}
	return uint(result.value)
}

// AsValue casts as a JS value.
func (o *ArrayObject) AsValue() *Value {
	return o.v
}

// AsObject casts as a JS object.
func (o *ArrayObject) AsObject() (*Object, error) {
	return o.v.AsObject()
}

var _ Valuer = (*ArrayObject)(nil)
