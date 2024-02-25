package gomonkey_test_context

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/bhuisgen/gomonkey"
)

func TestMain(m *testing.M) {
	gomonkey.Init()
	code := m.Run()
	gomonkey.ShutDown()
	os.Exit(code)
}

func TestNewContext(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Errorf("NewContext() err = %v, want %v", err, nil)
	}
	if ctx == nil {
		t.Errorf("NewContext() = %v", ctx)
	}
	ctx.Destroy()
}

func TestNewContext_WithOptions(t *testing.T) {
	ctx, err := gomonkey.NewContext(
		gomonkey.WithHeapMaxBytes(32*1024*1024),
		gomonkey.WithNativeStackSize(1024*1024),
		gomonkey.WithGCMaxBytes(1*1024*1024),
		gomonkey.WithGCIncrementalEnabled(true),
		gomonkey.WithGCSliceTimeBudget(50*time.Millisecond),
	)
	if err != nil {
		t.Errorf("NewContext() err = %v, want %v", err, nil)
	}
	if ctx == nil {
		t.Errorf("NewContext() = %v", ctx)
	}
	ctx.Destroy()
}

func TestContextDestroy(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	ctx.Destroy()
}

func TestContextRequestInterrupt(t *testing.T) {
	ctxCh := make(chan *gomonkey.Context, 1)
	errCh := make(chan error, 1)

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		ctx, err := gomonkey.NewContext()
		if err != nil {
			errCh <- err
			return
		}
		defer ctx.Destroy()
		ctxCh <- ctx

		result, err := ctx.Evaluate([]byte("while(true) {}"))
		if err != nil {
			errCh <- err
			return
		}
		result.Release()
		errCh <- nil
	}()

	ctx := <-ctxCh

	select {
	case <-errCh:
		t.Fail()
	case <-time.After(100 * time.Millisecond):
		ctx.RequestInterrupt()
		err := <-errCh
		if err == nil {
			t.Fail()
		}
	}
}

func TestContextGlobal(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	global, err := ctx.Global()
	if err != nil {
		t.Errorf("ctx.Global() err = %v, want %v", err, nil)
	}
	global.Release()
}

func TestContextDefineObject_Global(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	global, err := ctx.Global()
	if err != nil {
		t.Fatal()
	}
	defer global.Release()

	object, err := ctx.DefineObject(global, "test", 0)
	if err != nil {
		t.Errorf("ctx.DefineObject() err = %v, want %v", err, nil)
	}
	defer object.Release()
	if !global.Has("test") {
		t.Fatal()
	}
}

func TestContextDefineObject_Object(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	parent, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer parent.Release()

	object, err := ctx.DefineObject(parent, "test", 0)
	if err != nil {
		t.Errorf("ctx.DefineObject() err = %v, want %v", err, nil)
	}
	defer object.Release()
	if !parent.Has("test") {
		t.Fatal()
	}
}

func TestContextDefineProperty_Global(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	global, err := ctx.Global()
	if err != nil {
		t.Fatal()
	}
	defer global.Release()
	propValue, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer propValue.Release()

	if err := ctx.DefineProperty(global, "test", propValue, 0); err != nil {
		t.Errorf("ctx.DefineProperty() err = %v, want %v", err, nil)
	}
	if !global.Has("test") {
		t.Fatal()
	}
}

func TestContextDefineProperty_Object(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	propValue, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer propValue.Release()

	if err := ctx.DefineProperty(object, "test", propValue, 0); err != nil {
		t.Errorf("ctx.DefineProperty() err = %v, want %v", err, nil)
	}
	if !object.Has("test") {
		t.Fatal()
	}
}

func TestContextDefineFunction_Global(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	global, err := ctx.Global()
	if err != nil {
		t.Fatal()
	}
	defer global.Release()

	add := func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			t.Fatal()
		}
		if !args[0].IsInt32() || !args[1].IsInt32() {
			t.Fatal()
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			t.Fatal()
		}
		return retValue, nil
	}
	if err := ctx.DefineFunction(global, "add", add, 2, gomonkey.PropertyAttributeDefault); err != nil {
		t.Errorf("ctx.DefineFunction() err = %v, want %v", err, nil)
	}
	if !global.Has("add") {
		t.Fatal()
	}
}

func TestContextDefineFunction_Object(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()

	add := func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			t.Fatal()
		}
		if !args[0].IsInt32() || !args[1].IsInt32() {
			t.Fatal()
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			t.Fatal()
		}
		return retValue, nil
	}
	if err := ctx.DefineFunction(object, "add", add, 2, gomonkey.PropertyAttributeDefault); err != nil {
		t.Errorf("ctx.DefineFunction() err = %v, want %v", err, nil)
	}
	if !object.Has("add") {
		t.Fatal()
	}
}

func TestContextDefineElement(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	value, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()

	if err := ctx.DefineElement(object, 0, value, gomonkey.PropertyAttributeDefault); err != nil {
		t.Errorf("ctx.DefineFunction() err = %v, want %v", err, nil)
	}
	if !object.HasElement(0) {
		t.Fatal()
	}
}

func TestContextCallFunctionName_DefineFunction(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	add := func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			t.Fatal()
		}
		if !args[0].IsInt32() || !args[1].IsInt32() {
			t.Fatal()
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			t.Fatal()
		}
		return retValue, nil
	}
	if err := ctx.DefineFunction(object, "add", add, 2, gomonkey.PropertyAttributeDefault); err != nil {
		t.Fatal()
	}

	fnVal, err := object.Get("add")
	if err != nil {
		t.Fatal()
	}
	defer fnVal.Release()
	arg0, err := gomonkey.NewValueInt32(ctx, 1)
	if err != nil {
		t.Fatal()
	}
	defer arg0.Release()
	arg1, err := gomonkey.NewValueInt32(ctx, 2)
	if err != nil {
		t.Fatal()
	}
	defer arg1.Release()
	result, err := ctx.CallFunctionName("add", object.AsValue(), arg0, arg1)
	if err != nil {
		t.Errorf("ctx.CallFunctionName() err = %v, want %v", err, nil)
	}
	defer result.Release()
	if !result.IsInt32() || result.ToInt32() != 3 {
		t.Fatal()
	}
}

func TestContextCallFunctionName_NewFunction(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	fn, err := gomonkey.NewFunction(ctx, "add", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			t.Fatal()
		}
		if !args[0].IsInt32() || !args[1].IsInt32() {
			t.Fatal()
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			t.Fatal()
		}
		return retValue, nil
	})
	if err != nil {
		t.Fatal()
	}
	defer fn.Release()
	if err := object.Set("add", fn.AsValue()); err != nil {
		t.Fatal()
	}

	arg0, err := gomonkey.NewValueInt32(ctx, 1)
	if err != nil {
		t.Fatal()
	}
	defer arg0.Release()
	arg1, err := gomonkey.NewValueInt32(ctx, 2)
	if err != nil {
		t.Fatal()
	}
	defer arg1.Release()
	result, err := ctx.CallFunctionName("add", object.AsValue(), arg0, arg1)
	if err != nil {
		t.Errorf("ctx.CallFunctionName() err = %v, want %v", err, nil)
	}
	defer result.Release()
	if !result.IsInt32() || result.ToInt32() != 3 {
		t.Fatal()
	}
}
func TestContextCallFunctionValue_DefineFunction(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	add := func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			t.Fatal()
		}
		if !args[0].IsInt32() || !args[1].IsInt32() {
			t.Fatal()
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			t.Fatal()
		}
		return retValue, nil
	}
	if err := ctx.DefineFunction(object, "add", add, 2, gomonkey.PropertyAttributeDefault); err != nil {
		t.Fatal()
	}

	fnValue, err := object.Get("add")
	if err != nil {
		t.Fatal()
	}
	defer fnValue.Release()
	arg0, err := gomonkey.NewValueInt32(ctx, 1)
	if err != nil {
		t.Fatal()
	}
	defer arg0.Release()
	arg1, err := gomonkey.NewValueInt32(ctx, 2)
	if err != nil {
		t.Fatal()
	}
	defer arg1.Release()
	result, err := ctx.CallFunctionValue(fnValue, object.AsValue(), arg0, arg1)
	if err != nil {
		t.Errorf("ctx.CallFunctionValue() err = %v, want %v", err, nil)
	}
	defer result.Release()
	if !result.IsInt32() || result.ToInt32() != 3 {
		t.Fatal()
	}
}

func TestContextCallFunctionValue_NewFunction(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	fn, err := gomonkey.NewFunction(ctx, "add", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		if len(args) != 2 {
			t.Fatal()
		}
		if !args[0].IsInt32() || !args[1].IsInt32() {
			t.Fatal()
		}
		retValue, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			t.Fatal()
		}
		return retValue, nil
	})
	if err != nil {
		t.Fatal()
	}
	defer fn.Release()

	arg0, err := gomonkey.NewValueInt32(ctx, 1)
	if err != nil {
		t.Fatal()
	}
	defer arg0.Release()
	arg1, err := gomonkey.NewValueInt32(ctx, 2)
	if err != nil {
		t.Fatal()
	}
	defer arg1.Release()
	result, err := ctx.CallFunctionValue(fn.AsValue(), object.AsValue(), arg0, arg1)
	if err != nil {
		t.Errorf("ctx.CallFunctionValue() err = %v, want %v", err, nil)
	}
	defer result.Release()
	if !result.IsInt32() || result.ToInt32() != 3 {
		t.Fatal()
	}
}

func TestContextEvaluate(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	result, err := ctx.Evaluate([]byte(`(() => { return "test"; })()`))
	if err != nil {
		t.Errorf("ctx.Evaluate() err = %v, want %v", err, nil)
	}
	result.Release()
}

func TestContextCompileScript(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	script, err := ctx.CompileScript("script.js", []byte(`(() => { return "test"; })()`))
	if err != nil {
		t.Errorf("ctx.CompileScript() err = %v, want %v", err, nil)
	}
	script.Release()
}

func TestContextExecuteScript(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	script, err := ctx.CompileScript("script.js", []byte(`(() => { return "test"; })()`))
	if err != nil {
		t.Fatal()
	}
	defer script.Release()

	result, err := ctx.ExecuteScript(script)
	if err != nil {
		t.Errorf("ctx.ExecuteScript() err = %v, want %v", err, nil)
	}
	defer result.Release()
	if !result.IsString() || result.ToString() != "test" {
		t.Fatal()
	}
}

func TestContextExecuteScriptFromStencil(t *testing.T) {
	fc, err := gomonkey.NewFrontendContext()
	if err != nil {
		t.Fatal()
	}
	stencil, err := fc.CompileScriptToStencil("script.js", []byte(`(() => { return "test"; })()`))
	if err != nil {
		t.Fatal()
	}
	fc.Destroy()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	result, err := ctx.ExecuteScriptFromStencil(stencil)
	if err != nil {
		t.Errorf("ctx.ExecuteScriptFromStencil() err = %v, want %v", err, nil)
	}
	defer result.Release()
	if !result.IsString() || result.ToString() != "test" {
		t.Fatal()
	}

	stencil.Release()
}

func TestNewFrontendContext(t *testing.T) {
	ctx, err := gomonkey.NewFrontendContext()
	if err != nil {
		t.Errorf("NewFrontendContext() err = %v, want %v", err, nil)
	}
	if ctx == nil {
		t.Errorf("NewFrontendContext() = %v", ctx)
	}
	ctx.Destroy()
}

func TestNewFrontendContext_WithOptions(t *testing.T) {
	ctx, err := gomonkey.NewFrontendContext(
		gomonkey.WithFrontendNativeStackSize(1024 * 1024),
	)
	if err != nil {
		t.Errorf("NewFrontendContext() err = %v, want %v", err, nil)
	}
	if ctx == nil {
		t.Errorf("NewFrontendContext() = %v", ctx)
	}
	ctx.Destroy()
}

func TestFrontendContextDestroy(t *testing.T) {
	ctx, err := gomonkey.NewFrontendContext()
	if err != nil {
		t.Fatal()
	}
	ctx.Destroy()
}

func TestFrontendContextCompileScript(t *testing.T) {
	ctx, err := gomonkey.NewFrontendContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	stencil, err := ctx.CompileScriptToStencil("script.js", []byte(`(() => { return "test"; })()`))
	if err != nil {
		t.Errorf("ctx.CompileScriptToStencil() err = %v, want %v", err, nil)
	}
	stencil.Release()
}
