package gomonkey_test_object

import (
	"errors"
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

func TestNewObject(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Errorf("NewObject error = %v, want %v", err, nil)
	}
	object.Release()
}

func TestObjectRelease(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}

	object.Release()
}

func TestObjectHas(t *testing.T) {
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
	if err := object.Set("test", value); err != nil {
		t.Fatal()
	}

	if got := object.Has("test"); got != true {
		t.Errorf("o.Has() got %v, want %v", got, true)
	}
}

func TestObjectSet(t *testing.T) {
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

	if err := object.Set("test", value); err != nil {
		t.Errorf("o.Set() err = %v, want %v", err, nil)
	}
}

func TestObjectGet(t *testing.T) {
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
	if err := object.Set("test", value); err != nil {
		t.Fatal()
	}

	propValue, err := object.Get("test")
	if err != nil {
		t.Errorf("o.Get() err = %v, want %v", err, nil)
	}
	defer propValue.Release()
	if got := propValue.IsString(); got != true {
		t.Errorf("propValue.IsString() got %v, want %v", got, true)
	}
	if got := propValue.String(); got != "test" {
		t.Errorf("propValue.String() got %v, want %v", got, nil)
	}
}

func TestObjectDelete(t *testing.T) {
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
	if err := object.Set("test", value); err != nil {
		t.Fatal()
	}

	if err := object.Delete("test"); err != nil {
		t.Errorf("o.Delete() err = %v, want %v", err, nil)
	}
}

func TestObjectCall(t *testing.T) {
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
		value, err := gomonkey.NewValueInt32(ctx, args[0].ToInt32()+args[1].ToInt32())
		if err != nil {
			t.Fatal()
		}
		return value, nil
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
	value, err := object.Call("add", arg0, arg1)
	if err != nil {
		t.Errorf("o.Call() err = %v, want %v", err, nil)
	}
	defer value.Release()
	if !value.IsInt32() || value.ToInt32() != 3 {
		t.Fatal()
	}
}

func TestObjectCall_Error(t *testing.T) {
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

	errMessage := "test error"

	fn, err := gomonkey.NewFunction(ctx, "test", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		return nil, errors.New(errMessage)
	})
	if err != nil {
		t.Fatal()
	}
	defer fn.Release()
	if err := object.Set("Test", fn.AsValue()); err != nil {
		t.Fatal()
	}

	result, err := object.Call("Test")
	if err == nil {
		result.Release()
		t.Errorf("o.Call() err = %v, want %v", err, nil)
	}
	if e, ok := err.(*gomonkey.JSError); !ok {
		t.Errorf("o.Call() err type = %T, want JSError", e)
	}
	if m := err.Error(); m != errMessage {
		t.Errorf("o.Call() err message = %s, want %s", errMessage, m)
	}
}

func TestObjectHasElement(t *testing.T) {
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
	if err := object.SetElement(0, value); err != nil {
		t.Fatal()
	}

	if got := object.HasElement(0); got != true {
		t.Errorf("o.HasElement() got %v, want %v", got, true)
	}
}

func TestObjectSetElement(t *testing.T) {
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

	if err := object.SetElement(0, value); err != nil {
		t.Errorf("o.SetElement() err = %v, want %v", err, nil)
	}
}

func TestObjectGetElement(t *testing.T) {
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
	if err := object.SetElement(0, value); err != nil {
		t.Fatal()
	}

	propValue, err := object.GetElement(0)
	if err != nil {
		t.Errorf("o.GetElement() err = %v, want %v", err, nil)
	}
	defer propValue.Release()
	if !propValue.IsString() || propValue.String() != "test" {
		t.Fatal()
	}
}

func TestObjectDeleteElement(t *testing.T) {
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
	if err := object.SetElement(0, value); err != nil {
		t.Fatal()
	}

	err = object.DeleteElement(0)
	if err != nil {
		t.Errorf("o.DeleteElement() err = %v, want %v", err, nil)
	}
}

func TestObjectAsValue(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	val := v.AsValue()
	if val == nil {
		t.Errorf("o.AsValue() = %v", val)
	}
}
