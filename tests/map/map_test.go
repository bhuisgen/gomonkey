package gomonkey_test_map

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

func TestNewMapObject(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Errorf("NewMapObject error = %v, want %v", err, nil)
	}
	object.Release()
}

func TestMapObjectRelease(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}

	object.Release()
}

func TestMapObjectSize(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()

	if got := object.Size(); got != 0 {
		t.Errorf("o.Size() = %v, want %v", got, 0)
	}
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Set(key, value); err != nil {
		t.Fatal()
	}
	if got := object.Size(); got != 1 {
		t.Errorf("o.Size() = %v, want %v", got, 1)
	}
}

func TestMapObjectHas(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()

	if err := object.Set(key, value); err != nil {
		t.Fatal()
	}

	if got := object.Has(key); got != true {
		t.Errorf("o.Has() got %v, want %v", got, true)
	}
}
func TestMapObjectGet(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()

	if err := object.Set(key, value); err != nil {
		t.Fatal()
	}

	got, err := object.Get(key)
	if err != nil {
		t.Errorf("o.Get() error = %v, want %v", err, nil)
	}
	if !got.Is(value) {
		t.Errorf("o.Get() got %v, want %v", got, value)
	}
}

func TestMapObjectSet(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()

	if err := object.Set(key, value); err != nil {
		t.Errorf("o.Set() error = %v, want %v", err, nil)
	}
}
func TestMapObjectDelete(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Set(key, value); err != nil {
		t.Fatal()
	}

	if err := object.Delete(key); err != nil {
		t.Errorf("o.Delete() error = %v, want %v", err, nil)
	}
}

func TestMapObjectClear(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Set(key, value); err != nil {
		t.Fatal()
	}

	if err := object.Clear(); err != nil {
		t.Errorf("o.Clear() error = %v, want %v", err, nil)
	}
}

func TestMapObjectKeys(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Set(key, value); err != nil {
		t.Fatal()
	}

	keys, err := object.Keys()
	if err != nil {
		t.Errorf("o.Keys() error = %v, want %v", err, nil)
	}
	defer keys.Release()
}

func TestMapObjectValues(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if err := object.Set(key, value); err != nil {
		t.Fatal()
	}

	values, err := object.Values()
	if err != nil {
		t.Errorf("o.Values() error = %v, want %v", err, nil)
	}
	defer values.Release()
}

func TestMapObjectEntries(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	object, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer object.Release()
	key, err := gomonkey.NewValueString(ctx, "key")
	if err != nil {
		t.Fatal()
	}
	defer key.Release()
	value, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()

	entries, err := object.Entries()
	if err != nil {
		t.Errorf("o.Entries() error = %v, want %v", err, nil)
	}
	defer entries.Release()
	if err := object.Set(key, value); err != nil {
		t.Fatal()
	}
}

func TestMapObjectAsValue(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewMapObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	val := v.AsValue()
	if val == nil {
		t.Errorf("o.AsValue() = %v", val)
	}
}
