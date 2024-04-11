package gomonkey_test_array

import (
	"os"
	"runtime"
	"testing"

	"github.com/bhuisgen/gomonkey"
)

func TestMain(m *testing.M) {
	gomonkey.Init()
	code := m.Run()
	gomonkey.ShutDown()
	os.Exit(code)
}

func TestNewArrayObject(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	array, err := gomonkey.NewArrayObject(ctx)
	if err != nil {
		t.Errorf("NewArrayObject error = %v, want %v", err, nil)
	}
	array.Release()
}

func TestArrayObjectRelease(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	array, err := gomonkey.NewArrayObject(ctx)
	if err != nil {
		t.Fatal()
	}

	array.Release()
}

func TestArrayObjectLength(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	array, err := gomonkey.NewArrayObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer array.Release()

	if array.Length() != 0 {
		t.Errorf("a.Length() got %v, want %v", array.Length(), 0)
	}
}

func TestArrayObjectAsValue(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	array, err := gomonkey.NewArrayObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer array.Release()

	val := array.AsValue()
	if val == nil {
		t.Errorf("a.AsValue() = %v", val)
	}
}

func TestArrayObjectAsObject(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	array, err := gomonkey.NewArrayObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer array.Release()

	object, err := array.AsObject()
	if err != nil {
		t.Errorf("a.AsObject() err = %v, want %v", err, nil)
	}
	if object == nil {
		t.Errorf("a.AsObject() = %v", object)
	}
}
