package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// Object implements a JS object.
type Object struct {
	v *Value
}

// NewObject creates a new object.
func NewObject(ctx *Context) (*Object, error) {
	result := C.NewPlainObject(ctx.ptr)
	if !result.ok {
		err := errors.New(C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Object{&Value{result.ptr, ctx}}, nil
}

// Release releases the object.
func (o *Object) Release() {
	C.ReleaseValue(o.v.ptr)
}

// Has checks if the object has the given property.
func (o *Object) Has(key string) bool {
	cKey := C.CString(key)
	result := C.ObjectHasProperty(o.v.ptr, cKey)
	C.free(unsafe.Pointer(cKey))
	if !result.ok {
		C.free(unsafe.Pointer(result.err.message))
		return false
	}
	return bool(result.value)
}

// Get returns an object property.
func (o *Object) Get(key string) (*Value, error) {
	cKey := C.CString(key)
	result := C.ObjectGetProperty(o.v.ptr, cKey)
	C.free(unsafe.Pointer(cKey))
	if !result.ok {
		err := fmt.Errorf("get property: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// Set defines an object property.
func (o *Object) Set(key string, value *Value) error {
	cKey := C.CString(key)
	result := C.ObjectSetProperty(o.v.ptr, cKey, value.ptr)
	C.free(unsafe.Pointer(cKey))
	if !result.ok {
		err := fmt.Errorf("set property: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// Delete deletes an object property.
func (o *Object) Delete(key string) error {
	cKey := C.CString(key)
	result := C.ObjectDeleteProperty(o.v.ptr, cKey)
	C.free(unsafe.Pointer(cKey))
	if !result.ok {
		err := fmt.Errorf("delete property: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// Call calls a method.
func (o *Object) Call(name string, args ...*Value) (*Value, error) {
	propValue, err := o.Get(name)
	if err != nil {
		return nil, err
	}
	defer propValue.Release()
	fn, err := propValue.AsFunction()
	if err != nil {
		err := fmt.Errorf("get function: %s", err)
		return nil, err
	}

	return fn.Call(o, args...)
}

// HasElement checks if the object has the given element.
func (o *Object) HasElement(index int) bool {
	result := C.ObjectHasElement(o.v.ptr, C.uint(index))
	return bool(result.value)
}

// GetElement returns an object element.
func (o *Object) GetElement(index int) (*Value, error) {
	result := C.ObjectGetElement(o.v.ptr, C.uint(index))
	if !result.ok {
		err := fmt.Errorf("get element: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// SetElement defines an object element.
func (o *Object) SetElement(index int, value *Value) error {
	result := C.ObjectSetElement(o.v.ptr, C.uint(index), value.ptr)
	if !result.ok {
		err := fmt.Errorf("set element: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// DeleteElement deletes an object element.
func (o *Object) DeleteElement(index int) error {
	result := C.ObjectDeleteElement(o.v.ptr, C.uint(index))
	if !result.ok {
		err := fmt.Errorf("delete element: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// AsValue casts as a JS value.
func (o *Object) AsValue() *Value {
	return o.v
}

var _ Valuer = (*Object)(nil)
