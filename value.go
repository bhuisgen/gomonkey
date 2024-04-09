package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

// Valuer represents any object extending a JS value.
type Valuer interface {
	AsValue() *Value
}

// Value implements a JS Value.
type Value struct {
	ptr C.ValuePtr
	ctx *Context
}

// NewValueNull creates a new JS null value.
func NewValueNull(ctx *Context) (*Value, error) {
	result := C.NewValueNull(ctx.ptr)
	return valueFromResult(ctx, result)
}

// NewValueUndefined creates a new JS undefined value.
func NewValueUndefined(ctx *Context) (*Value, error) {
	result := C.NewValueUndefined(ctx.ptr)
	return valueFromResult(ctx, result)
}

// NewValueString creates a new JS string value.
func NewValueString(ctx *Context, str string) (*Value, error) {
	cStr := C.CString(str)
	result := C.NewValueString(ctx.ptr, cStr, C.int(len(str)))
	C.free(unsafe.Pointer(cStr))
	return valueFromResult(ctx, result)
}

// NewValueBoolean creates a new JS boolean value.
func NewValueBoolean(ctx *Context, b bool) (*Value, error) {
	result := C.NewValueBoolean(ctx.ptr, C.bool(b))
	return valueFromResult(ctx, result)
}

// NewValueNumber creates a new JS number value.
func NewValueNumber(ctx *Context, f float64) (*Value, error) {
	result := C.NewValueNumber(ctx.ptr, C.double(f))
	return valueFromResult(ctx, result)
}

// NewValueInt32 creates a new JS int32 value.
func NewValueInt32(ctx *Context, i int32) (*Value, error) {
	result := C.NewValueInt32(ctx.ptr, C.int(i))
	return valueFromResult(ctx, result)
}

// Release releases the value.
func (v *Value) Release() {
	C.ReleaseValue(v.ptr)
}

// String returns the value string representation.
func (v *Value) String() string {
	cStr := C.ToString(v.ptr)
	str := C.GoStringN(cStr.data, C.int(cStr.len))
	C.free(unsafe.Pointer(cStr.data))
	return str
}

// Format implements fmt.Formatter.
func (v *Value) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		fallthrough
	case 's':
		_, _ = io.WriteString(f, v.String())
	case 'q':
		fmt.Fprintf(f, "%q", v.String())
	}
}

// MarshalJSON implements json.Marshaler.
func (v *Value) MarshalJSON() ([]byte, error) {
	data, err := JSONStringify(v.ctx, v)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

// AsValue casts as a JS value.
func (v *Value) AsValue() *Value {
	return v
}

// AsObject casts as a JS object.
func (v *Value) AsObject() (*Object, error) {
	if !v.IsObject() {
		return nil, errors.New("not an JS::Object")
	}
	return &Object{v}, nil
}

// AsFunction casts as a JS function.
func (v *Value) AsFunction() (*Function, error) {
	if !v.IsFunction() {
		return nil, errors.New("not a JS::Function")
	}
	return &Function{v}, nil
}

// Is checks if the JS value is the same.
func (v *Value) Is(value *Value) bool {
	return bool(C.ValueIs(v.ptr, value.ptr))
}

// IsUndefined checks if the JS value is undefined.
func (v *Value) IsUndefined() bool {
	return bool(C.ValueIsUndefined(v.ptr))
}

// IsNull checks if the JS value is null.
func (v *Value) IsNull() bool {
	return bool(C.ValueIsNull(v.ptr))
}

// IsNullOrUndefined checks if the JS value is null or undefined.
func (v *Value) IsNullOrUndefined() bool {
	return bool(C.ValueIsNullOrUndefined(v.ptr))
}

// IsTrue checks if the JS value is a JS boolean sets to true.
func (v *Value) IsTrue() bool {
	return bool(C.ValueIsTrue(v.ptr))
}

// IsFalse checks if the JS value is a JS boolean sets to false.
func (v *Value) IsFalse() bool {
	return bool(C.ValueIsFalse(v.ptr))
}

// IsObject checks if the JS value is a JS object.
func (v *Value) IsObject() bool {
	return bool(C.ValueIsObject(v.ptr))
}

// IsFunction checks if the JS value is a JS function.
func (v *Value) IsFunction() bool {
	return bool(C.ValueIsFunction(v.ptr))
}

// IsSymbol checks if the JS value is a JS symbol.
func (v *Value) IsSymbol() bool {
	return bool(C.ValueIsSymbol(v.ptr))
}

// IsString checks if the JS value is a JS string.
func (v *Value) IsString() bool {
	return bool(C.ValueIsString(v.ptr))
}

// IsBoolean checks if the JS value is a JS boolean.
func (v *Value) IsBoolean() bool {
	return bool(C.ValueIsBoolean(v.ptr))
}

// IsNumber checks if the JS value is a JS number.
func (v *Value) IsNumber() bool {
	return bool(C.ValueIsNumber(v.ptr))
}

// IsInt32 checks if the JS value is a JS int32.
func (v *Value) IsInt32() bool {
	return bool(C.ValueIsInt32(v.ptr))
}

// ToString returns the JS string value.
func (v *Value) ToString() string {
	cStr := C.ValueToString(v.ptr)
	str := C.GoStringN(cStr.data, C.int(cStr.len))
	C.free(unsafe.Pointer(cStr.data))
	return str
}

// ToBoolean returns the JS boolean value.
func (v *Value) ToBoolean() bool {
	return C.ValueToBoolean(v.ptr) != 0
}

// ToNumber returns the JS number value.
func (v *Value) ToNumber() float64 {
	return float64(C.ValueToNumber(v.ptr))
}

// ToInt32 returns the JS int32 value.
func (v *Value) ToInt32() int32 {
	return int32(C.ValueToInt32(v.ptr))
}

var _ fmt.Formatter = (*Value)(nil)
var _ json.Marshaler = (*Value)(nil)
var _ Valuer = (*Value)(nil)
