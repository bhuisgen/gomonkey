package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// MapObject implements a JS map object.
type MapObject struct {
	v *Value
}

// NewMapObject creates a new map.
func NewMapObject(ctx *Context) (*MapObject, error) {
	result := C.NewMapObject(ctx.ptr)
	if !result.ok {
		err := errors.New(C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &MapObject{&Value{result.ptr, ctx}}, nil
}

// Release releases the map.
func (o *MapObject) Release() {
	C.ReleaseValue(o.v.ptr)
}

// Size returns the map size.
func (o *MapObject) Size() uint {
	result := C.MapObjectSize(o.v.ptr)
	if !result.ok {
		C.free(unsafe.Pointer(result.err.message))
		return 0
	}
	return uint(result.value)
}

// Has checks if the map has the given key.
func (o *MapObject) Has(key *Value) bool {
	result := C.MapObjectHas(o.v.ptr, key.ptr)
	if !result.ok {
		C.free(unsafe.Pointer(result.err.message))
		return false
	}
	return bool(result.value)
}

// Get returns a map value.
func (o *MapObject) Get(key *Value) (*Value, error) {
	result := C.MapObjectGet(o.v.ptr, key.ptr)
	if !result.ok {
		err := fmt.Errorf("get map value: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// Set sets a map value.
func (o *MapObject) Set(key *Value, value *Value) error {
	result := C.MapObjectSet(o.v.ptr, key.ptr, value.ptr)
	if !result.ok {
		err := fmt.Errorf("set map value: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// Delete deletes a map value.
func (o *MapObject) Delete(key *Value) error {
	result := C.MapObjectDelete(o.v.ptr, key.ptr)
	if !result.ok {
		err := fmt.Errorf("delete map value: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// Clear clears the map.
func (o *MapObject) Clear() error {
	result := C.MapObjectClear(o.v.ptr)
	if !result.ok {
		err := fmt.Errorf("clear map: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return err
	}
	return nil
}

// Keys returns the map keys.
func (o *MapObject) Keys() (*Value, error) {
	result := C.MapObjectKeys(o.v.ptr)
	if !result.ok {
		err := fmt.Errorf("map keys: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// Values returns the map values.
func (o *MapObject) Values() (*Value, error) {
	result := C.MapObjectValues(o.v.ptr)
	if !result.ok {
		err := fmt.Errorf("map values: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// Entries returns the map entries.
func (o *MapObject) Entries() (*Value, error) {
	result := C.MapObjectEntries(o.v.ptr)
	if !result.ok {
		err := fmt.Errorf("map entries: %s", C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, o.v.ctx}, nil
}

// AsValue casts as a JS value.
func (o *MapObject) AsValue() *Value {
	return o.v
}

var _ Valuer = (*MapObject)(nil)
