package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// SetObject implements a JS set object.
type SetObject struct {
	v *Value
}

// NewSetObject creates a new set.
func NewSetObject(ctx *Context) (*SetObject, error) {
	result := C.NewSetObject(ctx.ptr)
	if !result.ok {
		err := errors.New(C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &SetObject{&Value{result.ptr, ctx}}, nil
}

// Release releases the set.
func (o *SetObject) Release() {
	C.ReleaseValue(o.v.ptr)
}

// Size returns the set size.
func (o *SetObject) Size() uint {
	result := C.SetObjectSize(o.v.ptr)
	if !result.ok {
		C.free(unsafe.Pointer(result.err.message))
		return 0
	}
	return uint(result.value)
}

// Has checks if the set has the given key.
func (o *SetObject) Has(key *Value) bool {
	result := C.SetObjectHas(o.v.ptr, key.ptr)
	if !result.ok {
		C.free(unsafe.Pointer(result.err.message))
		return false
	}
	return bool(result.value)
}

// Add adds a value to the set.
func (o *SetObject) Add(value *Value) error {
	result := C.SetObjectAdd(o.v.ptr, value.ptr)
	if !result.ok {
		err := fmt.Errorf("add set: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// Delete deletes a key from the set.
func (o *SetObject) Delete(key *Value) error {
	result := C.SetObjectDelete(o.v.ptr, key.ptr)
	if !result.ok {
		err := fmt.Errorf("delete set: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// Clear clears the set.
func (o *SetObject) Clear() error {
	result := C.SetObjectClear(o.v.ptr)
	if !result.ok {
		err := fmt.Errorf("clear set: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// Keys returns the map keys.
func (o *SetObject) Keys() (*Value, error) {
	result := C.SetObjectKeys(o.v.ptr)
	if !result.ok {
		err := fmt.Errorf("set keys: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// Values returns the map values.
func (o *SetObject) Values() (*Value, error) {
	result := C.SetObjectValues(o.v.ptr)
	if !result.ok {
		err := fmt.Errorf("set values: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// Entries returns the map entries.
func (o *SetObject) Entries() (*Value, error) {
	result := C.SetObjectEntries(o.v.ptr)
	if !result.ok {
		err := fmt.Errorf("set entries: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// AsValue casts as a JS value.
func (o *SetObject) AsValue() *Value {
	return o.v
}

var _ Valuer = (*MapObject)(nil)
