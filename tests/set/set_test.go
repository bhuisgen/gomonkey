package gomonkey_test_set

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

func TestNewSetObject(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Errorf("NewSetObject error = %v, want %v", err, nil)
	}
	object.Release()
}

func TestSetObjectRelease(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}

	object.Release()
}

func TestSetObjectSize(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()

	if got := object.Size(); got != 0 {
		t.Errorf("o.Size() = %v, want %v", got, 0)
	}
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Add(value); err != nil {
		t.Fatal()
	}
	if got := object.Size(); got != 1 {
		t.Errorf("o.Size() = %v, want %v", got, 1)
	}
}

func TestSetObjectHas(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Add(value); err != nil {
		t.Fatal()
	}

	if got := object.Has(value); got != true {
		t.Errorf("o.Has() got %v, want %v", got, true)
	}
}

func TestSetObjectAdd(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()

	if err := object.Add(value); err != nil {
		t.Errorf("o.Add() error = %v, want %v", err, nil)
	}
	if got := object.Has(value); got != true {
		t.Errorf("o.Has() got %v, want %v", got, true)
	}
}

func TestSetObjectDelete(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Add(value); err != nil {
		t.Fatal()
	}

	if err := object.Delete(value); err != nil {
		t.Errorf("o.Delete() error = %v, want %v", err, nil)
	}
	if got := object.Has(value); got != false {
		t.Errorf("o.Has() got %v, want %v", got, false)
	}
}

func TestSetObjectClear(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Add(value); err != nil {
		t.Fatal()
	}

	if err := object.Clear(); err != nil {
		t.Errorf("o.Clear() error = %v, want %v", err, nil)
	}
}

func TestSetObjectKeys(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Add(value); err != nil {
		t.Fatal()
	}

	keys, err := object.Keys()
	if err != nil {
		t.Errorf("o.Keys() error = %v, want %v", err, nil)
	}
	defer keys.Release()
}

func TestSetObjectValues(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Add(value); err != nil {
		t.Fatal()
	}

	values, err := object.Values()
	if err != nil {
		t.Errorf("o.Values() error = %v, want %v", err, nil)
	}
	defer values.Release()
}

func TestSetObjectEntries(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Add(value); err != nil {
		t.Fatal()
	}

	entries, err := object.Entries()
	if err != nil {
		t.Errorf("o.Entries() error = %v, want %v", err, nil)
	}
	defer entries.Release()
}

func TestSetObjectAsValue(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()

	v := object.AsValue()
	if v == nil {
		t.Errorf("o.AsValue() = %v", v)
	}
}
