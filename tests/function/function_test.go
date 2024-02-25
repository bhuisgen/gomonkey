package gomonkey_test_function

import (
	"os"
	"testing"

	"github.com/bhuisgen/gomonkey"
)

func TestMain(m *testing.M) {
	gomonkey.Init()
	code := m.Run()
	gomonkey.ShutDown()
	os.Exit(code)
}

func TestNewFunction(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	fn, err := gomonkey.NewFunction(ctx, "test", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("NewFunction() err = %v, want %v", err, nil)
	}
	fn.Release()
}

func TestFunctionRelease(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	fn, err := gomonkey.NewFunction(ctx, "test", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal()
	}
	fn.Release()
}

func TestFunctionCall(t *testing.T) {
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
	arg1, err := gomonkey.NewValueInt32(ctx, 2)
	if err != nil {
		t.Fatal()
	}
	defer arg1.Release()
	result, err := fn.Call(fn.AsValue(), arg0, arg1)
	if err != nil {
		t.Errorf("f.Call() error = %v, want %v", err, nil)
	}
	defer result.Release()
	if !result.IsInt32() || result.ToInt32() != 3 {
		t.Fatal()
	}
}

func TestFunctionAsValue(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	fn, err := gomonkey.NewFunction(ctx, "test", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal()
	}
	defer fn.Release()

	val := fn.AsValue()
	if val == nil {
		t.Errorf("f.AsValue() = %v", val)
	}
}
