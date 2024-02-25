package gomonkey

// #include "gomonkey.h"
// #include <stdint.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"sync"
	"time"
	"unsafe"
)

// Context represents a JS context.
type Context struct {
	options     contextOptions
	ref         uint
	functions   map[string]FunctionCallback
	muFunctions sync.RWMutex
	ptr         C.ContextPtr
}

// contextOptions implements the context options.
type contextOptions struct {
	heapMaxBytes         uint
	nativeStackSize      uint
	gcMaxBytes           uint
	gcIncrementalEnabled uint
	gcSliceTimeBudgetMs  uint
}

// ContextOptionFunc represents a context option function.
type ContextOptionFunc func(c *Context) error

var contexts map[uint]*Context = map[uint]*Context{}
var contextsSeq uint
var muContexts sync.RWMutex

// NewContext creates a new context.
func NewContext(options ...ContextOptionFunc) (*Context, error) {
	context := &Context{}

	for _, option := range options {
		if err := option(context); err != nil {
			return nil, err
		}
	}

	context.functions = map[string]FunctionCallback{}

	muContexts.Lock()
	contextsSeq += 1
	context.ref = contextsSeq
	contexts[context.ref] = context
	muContexts.Unlock()

	ptr := C.NewContext(C.uint(context.ref), C.ContextOptions{
		heapMaxBytes:         C.uint(context.options.heapMaxBytes),
		stackSize:            C.uint(context.options.nativeStackSize),
		gcMaxBytes:           C.uint(context.options.gcMaxBytes),
		gcIncrementalEnabled: C.uint(context.options.gcIncrementalEnabled),
		gcSliceTimeBudgetMs:  C.uint(context.options.gcSliceTimeBudgetMs),
	})
	if ptr == nil {
		return nil, errors.New("new context")
	}
	context.ptr = ptr

	return context, nil
}

// WithHeapMaxBytes sets the maximum heap size in bytes.
func WithHeapMaxBytes(max uint) ContextOptionFunc {
	return func(c *Context) error {
		c.options.heapMaxBytes = max
		return nil
	}
}

// WithNativeStackSize sets the native stack size in bytes.
func WithNativeStackSize(size uint) ContextOptionFunc {
	return func(c *Context) error {
		c.options.nativeStackSize = size
		return nil
	}
}

// WithGCMaxBytes sets the maximum heap size in bytes before GC.
func WithGCMaxBytes(max uint) ContextOptionFunc {
	return func(c *Context) error {
		c.options.gcMaxBytes = max
		return nil
	}
}

// WithGCIncrementalEnabled enables the incremental GC.
func WithGCIncrementalEnabled(state bool) ContextOptionFunc {
	return func(c *Context) error {
		if state {
			c.options.gcIncrementalEnabled = 1
		}
		return nil
	}
}

// WithGCSliceTimeBudget sets the maximal time to spend in an incremental GC slice.
func WithGCSliceTimeBudget(d time.Duration) ContextOptionFunc {
	return func(c *Context) error {
		c.options.gcSliceTimeBudgetMs = uint(d.Milliseconds())
		return nil
	}
}

// Destroy destroys the context.
func (c *Context) Destroy() {
	C.DestroyContext(c.ptr)

	muContexts.Lock()
	delete(contexts, c.ref)
	muContexts.Unlock()
}

// RequestInterrupt requests the context interruption.
func (c *Context) RequestInterrupt() {
	C.RequestInterruptContext(c.ptr)
}

// Global returns the global object.
func (c *Context) Global() (*Object, error) {
	result := C.GetGlobalObject(c.ptr)
	if !result.ok {
		return nil, newJSError(result.err)
	}
	return &Object{&Value{result.ptr, c}}, nil
}

//export goFunctionContext
func goFunctionContext(contextRef C.uint) C.ContextPtr {
	muContexts.RLock()
	ctx, ok := contexts[uint(contextRef)]
	muContexts.RUnlock()
	if !ok {
		return nil
	}
	return ctx.ptr
}

// registerCallback registers a function callback.
func (c *Context) registerCallback(name string, cb FunctionCallback) {
	c.muFunctions.Lock()
	c.functions[name] = cb
	c.muFunctions.Unlock()
}

//export goFunctionCallback
func goFunctionCallback(contextRef C.uint, name *C.char, argc C.uint, vp *C.ValuePtr) C.ResultGoFunctionCallback {
	result := C.ResultGoFunctionCallback{}

	muContexts.RLock()
	ctx, ok := contexts[uint(contextRef)]
	muContexts.RUnlock()
	if !ok {
		result.err = C.CString("invalid context ref")
		return result
	}

	ctx.muFunctions.RLock()
	gName := C.GoString(name)
	callback, ok := ctx.functions[gName]
	ctx.muFunctions.RUnlock()
	if !ok {
		result.err = C.CString("invalid function name")
		return result
	}

	vps := unsafe.Slice(vp, argc)
	values := make([]*Value, 0, int(argc))
	for i := 0; i < int(argc); i++ {
		value := &Value{ptr: vps[i], ctx: ctx}
		values = append(values, value)
	}

	val, err := callback(values)
	if err != nil {
		result.err = C.CString(err.Error())
		return result
	}
	if val != nil {
		result.ptr = val.ptr
		return result
	}
	return result
}

// newFunction creates a new JS function.
func (c *Context) newFunction(name string, callback FunctionCallback) (*Value, error) {
	cName := C.CString(name)
	result := C.NewFunction(c.ptr, cName)
	C.free(unsafe.Pointer(cName))
	if !result.ok {
		return nil, newJSError(result.err)
	}
	c.registerCallback(name, callback)
	return valueFromResultWithJSError(c, result)
}

// DefineObject defines a new JS object and sets it as a property of the given JS object.
func (c *Context) DefineObject(object *Object, name string, attrs PropertyAttributes) (*Object, error) {
	cName := C.CString(name)
	result := C.DefineObject(c.ptr, object.AsValue().ptr, cName, C.uint(attrs))
	C.free(unsafe.Pointer(cName))
	if !result.ok {
		return nil, newJSError(result.err)
	}
	return &Object{&Value{result.ptr, c}}, nil
}

// DefineProperty defines a new property on the given JS object.
func (c *Context) DefineProperty(object *Object, name string, value *Value, attrs PropertyAttributes) error {
	cName := C.CString(name)
	result := C.DefineProperty(c.ptr, object.AsValue().ptr, cName, value.ptr, C.uint(attrs))
	C.free(unsafe.Pointer(cName))
	if !result.ok {
		return newJSError(result.err)
	}
	return nil
}

// DefineElement defines a new element on the given JS object.
func (c *Context) DefineElement(object *Object, index uint, value *Value, attrs PropertyAttributes) error {
	result := C.DefineElement(c.ptr, object.AsValue().ptr, C.uint(index), value.ptr, C.uint(attrs))
	if !result.ok {
		return newJSError(result.err)
	}
	return nil
}

// DefineFunction defines a new JS function and sets it as a property of the given JS object.
func (c *Context) DefineFunction(object *Object, name string, callback FunctionCallback, args uint,
	attrs PropertyAttributes) error {
	cName := C.CString(name)
	result := C.DefineFunction(c.ptr, object.AsValue().ptr, cName, C.uint(args), C.uint(attrs))
	C.free(unsafe.Pointer(cName))
	if !result.ok {
		return newJSError(result.err)
	}
	c.registerCallback(name, callback)
	return nil
}

// CallFunctionName executes a JS function by its name.
func (c *Context) CallFunctionName(name string, receiver Valuer, args ...*Value) (*Value, error) {
	argc := len(args)
	var argv *C.ValuePtr
	if argc > 0 {
		cArgs := make([]C.ValuePtr, argc)
		for i, arg := range args {
			cArgs[i] = arg.ptr
		}
		argv = (*C.ValuePtr)(unsafe.Pointer(&cArgs[0]))
	}

	cName := C.CString(name)
	result := C.CallFunctionName(c.ptr, cName, receiver.AsValue().ptr, C.int(argc), argv)
	C.free(unsafe.Pointer(cName))
	return valueFromResultWithJSError(c, result)
}

// CallFunctionName executes a JS function by its value.
func (c *Context) CallFunctionValue(function Valuer, receiver Valuer, args ...*Value) (*Value, error) {
	argc := len(args)
	var argv *C.ValuePtr
	if argc > 0 {
		cArgs := make([]C.ValuePtr, argc)
		for i, arg := range args {
			cArgs[i] = arg.ptr
		}
		argv = (*C.ValuePtr)(unsafe.Pointer(&cArgs[0]))
	}
	result := C.CallFunctionValue(c.ptr, function.AsValue().ptr, receiver.AsValue().ptr, C.int(argc), argv)
	return valueFromResultWithJSError(c, result)
}

// Evaluates executes a JS code.
func (c *Context) Evaluate(code []byte) (*Value, error) {
	cCode := C.CString(string(code))
	result := C.Evaluate(c.ptr, cCode)
	C.free(unsafe.Pointer(cCode))
	return valueFromResultWithJSError(c, result)
}

// CompileScript compiles a JS code into a script.
func (c *Context) CompileScript(name string, code []byte) (*Script, error) {
	cName := C.CString(name)
	cCode := C.CString(string(code))
	result := C.CompileScript(c.ptr, cName, cCode)
	C.free(unsafe.Pointer(cName))
	C.free(unsafe.Pointer(cCode))
	if !result.ok {
		return nil, newJSError(result.err)
	}
	return &Script{result.ptr}, nil
}

// Execute executes a script.
func (c *Context) ExecuteScript(script *Script) (*Value, error) {
	result := C.ExecuteScript(c.ptr, script.ptr)
	return valueFromResultWithJSError(c, result)
}

// Execute executes a script from a stencil.
func (c *Context) ExecuteScriptFromStencil(stencil *Stencil) (*Value, error) {
	result := C.ExecuteScriptFromStencil(c.ptr, stencil.ptr)
	return valueFromResultWithJSError(c, result)
}

// FrontendContext represents a JS frontend context.
type FrontendContext struct {
	options frontendContextOptions
	ptr     C.FrontendContextPtr
}

// frontendContextOptions implements the frontend context options.
type frontendContextOptions struct {
	nativeStackSize uint
}

// FrontendContextOptionFunc represents a frontend context option function.
type FrontendContextOptionFunc func(c *FrontendContext) error

// NewFrontendContext creates a new context.
func NewFrontendContext(options ...FrontendContextOptionFunc) (*FrontendContext, error) {
	context := &FrontendContext{}

	for _, option := range options {
		if err := option(context); err != nil {
			return nil, err
		}
	}

	ptr := C.NewFrontendContext(C.FrontendContextOptions{
		stackSize: C.uint(context.options.nativeStackSize),
	})
	if ptr == nil {
		return nil, errors.New("new frontend context")
	}
	context.ptr = ptr

	return context, nil
}

// WithNativeStackSize sets the native stack size in bytes.
func WithFrontendNativeStackSize(size uint) FrontendContextOptionFunc {
	return func(c *FrontendContext) error {
		c.options.nativeStackSize = size
		return nil
	}
}

// Destroy destroys the context.
func (c *FrontendContext) Destroy() {
	C.DestroyFrontendContext(c.ptr)
}

// CompileScriptToStencil compiles a script to a stencil.
func (c *FrontendContext) CompileScriptToStencil(name string, code []byte) (*Stencil, error) {
	cName := C.CString(name)
	cCode := C.CString(string(code))
	result := C.CompileScriptToStencil(c.ptr, cName, cCode)
	C.free(unsafe.Pointer(cName))
	C.free(unsafe.Pointer(cCode))
	if !result.ok {
		return nil, errors.New("compile script")
	}
	return &Stencil{ptr: result.ptr}, nil
}

// PropertyAttributes represents the attributes of a property.
type PropertyAttributes uint8

const (
	PropertyAttributeDefault   PropertyAttributes = 0
	PropertyAttributeEnumerate                    = 1 << iota
	PropertyAttributeReadOnly
	PropertyAttributePermanent
)

// Has checks if an attribute is set.
func (a PropertyAttributes) Has(attr PropertyAttributes) bool {
	return a&attr != 0
}

// valueFromResult returns a value from a result or a generic error.
func valueFromResult(ctx *Context, result C.ResultValue) (*Value, error) {
	if !result.ok {
		err := errors.New(C.GoString(result.err.message))
		C.free(unsafe.Pointer(result.err.message))
		return nil, err
	}
	return &Value{result.ptr, ctx}, nil
}

// valueFromResultWithJSError returns a value from a result or a JSError error.
func valueFromResultWithJSError(ctx *Context, result C.ResultValue) (*Value, error) {
	if !result.ok {
		return nil, newJSError(result.err)
	}
	return &Value{result.ptr, ctx}, nil
}
