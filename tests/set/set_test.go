package gomonkey_test_set

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

func TestNewSetObject(t *testing.T) {
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
}
func TestSetObjectDelete(t *testing.T) {
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
}

func TestSetObjectClear(t *testing.T) {
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
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewSetObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	val := v.AsValue()
	if val == nil {
		t.Errorf("o.AsValue() = %v", val)
	}
}
