package gomonkey_test_json

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

func TestJSONParse(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	object, err := gomonkey.JSONParse(ctx, `{"test":"value"}`)
	if err != nil {
		t.Errorf("ctx.ParseJSON() err = %v, want %v", err, nil)
	}
	defer object.Release()
	if !object.Has("test") {
		t.Fatal()
	}
	value, err := object.Get("test")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	if !value.IsString() || value.ToString() != "value" {
		t.Fatal()
	}
}

func TestJSONStringify_Object(t *testing.T) {
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
	propValue, err := gomonkey.NewValueString(ctx, "value")
	if err != nil {
		t.Fatal()
	}
	defer propValue.Release()
	if err := object.Set("test", propValue); err != nil {
		t.Fatal()
	}
	str, err := gomonkey.JSONStringify(ctx, object)
	if err != nil {
		t.Errorf("ctx.Stringify() err = %v, want %v", err, nil)
	}
	if str != `{"test":"value"}` {
		t.Errorf("ctx.Stringify() %v, want %v", str, `{"test":"value"}`)
	}
}

func TestJSONStringify_Value(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	value, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer value.Release()
	str, err := gomonkey.JSONStringify(ctx, value)
	if err != nil {
		t.Errorf("ctx.Stringify() err = %v, want %v", err, nil)
	}
	if str != `"test"` {
		t.Errorf("ctx.Stringify() %v, want %v", str, `"test"`)
	}
}
