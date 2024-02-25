package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"

// Function represents a JS function.
type Function struct {
	v *Value
}

// FunctionCallback implements a JS function callback.
type FunctionCallback func(args []*Value) (*Value, error)

// NewFunction creates a new JS function.
func NewFunction(ctx *Context, name string, callback FunctionCallback) (*Function, error) {
	value, err := ctx.newFunction(name, callback)
	if err != nil {
		return nil, err
	}
	return &Function{value}, nil
}

// Release releases the function.
func (f *Function) Release() {
	C.ReleaseValue(f.v.ptr)
}

// Call calls the function.
func (f *Function) Call(recv Valuer, args ...*Value) (*Value, error) {
	return f.v.ctx.CallFunctionValue(f, recv, args...)
}

// AsValue casts as a JS value.
func (f *Function) AsValue() *Value {
	return f.v
}

var _ Valuer = (*Function)(nil)
