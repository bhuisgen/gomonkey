package gomonkey_test_value

import (
	"bytes"
	"fmt"
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

func TestNewValueNull(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	v, err := gomonkey.NewValueNull(ctx)
	if err != nil {
		t.Errorf("NewValueNull() err = %v, want %v", err, nil)
	}
	defer v.Release()
	if got := v.IsNull(); got != true {
		t.Errorf("v.IsNull() got %v, want %v", got, true)
	}
}

func TestNewValueUndefined(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	v, err := gomonkey.NewValueUndefined(ctx)
	if err != nil {
		t.Errorf("NewValueUndefined() err = %v, want %v", err, nil)
	}
	defer v.Release()
	if got := v.IsUndefined(); got != true {
		t.Errorf("v.IsUndefined() got %v, want %v", got, true)
	}
}

func TestNewValueString(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	v, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Errorf("NewValueString() err = %v, want %v", err, nil)
	}
	defer v.Release()
	if got := v.IsString(); got != true {
		t.Errorf("v.IsString() got %v, want %v", got, true)
	}
	if got := v.String(); got != "test" {
		t.Errorf("v.String() got %v, want %v", got, "test")
	}
}

func TestNewValueBoolean(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	v, err := gomonkey.NewValueBoolean(ctx, true)
	if err != nil {
		t.Errorf("NewValueBoolean() err = %v, want %v", err, nil)
	}
	defer v.Release()
	if got := v.IsBoolean(); got != true {
		t.Errorf("v.IsBoolean() got %v, want %v", got, true)
	}
	if got := v.ToBoolean(); got != true {
		t.Errorf("v.ToBoolean() got %v, want %v", got, true)
	}
}

func TestNewValueNumber(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	v, err := gomonkey.NewValueNumber(ctx, 123.456)
	if err != nil {
		t.Errorf("NewValueNumber() err = %v, want %v", err, nil)
	}
	defer v.Release()
	if got := v.IsNumber(); got != true {
		t.Errorf("v.IsBoolean() got %v, want %v", got, true)
	}
	if got := v.ToNumber(); got != 123.456 {
		t.Errorf("v.ToBoolean() got %v, want %v", got, 123.456)
	}
}

func TestNewValueInt32(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	v, err := gomonkey.NewValueInt32(ctx, 123)
	if err != nil {
		t.Errorf("NewValueInt32() err = %v, want %v", err, nil)
	}
	defer v.Release()
	if got := v.IsInt32(); got != true {
		t.Errorf("v.IsBoolean() got %v, want %v", got, true)
	}
	if got := v.ToInt32(); got != 123 {
		t.Errorf("v.ToBoolean() got %v, want %v", got, 123)
	}
}

func TestValueRelease(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}

	v.Release()
}

func TestValueString(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.String(); got != "test" {
		t.Errorf("v.String() got %v, got %v", got, "test")
	}
}

func TestValueFormat(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := fmt.Sprintf("%s", v); got != "test" {
		t.Errorf("fmt.Sprintf() got %v, want %v", got, "test")
	}
	if got := fmt.Sprintf("%v", v); got != "test" {
		t.Errorf("fmt.Sprintf() got %v, want %v", got, "test")
	}
	if got := fmt.Sprintf("%q", v); got != "\"test\"" {
		t.Errorf("fmt.Sprintf() got %v, want %v", got, "\"test\"")
	}
}

func TestValueMarshalJSON(t *testing.T) {
	type args struct {
		value func(*gomonkey.Context) (*gomonkey.Value, error)
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "string",
			args: args{
				value: func(ctx *gomonkey.Context) (*gomonkey.Value, error) {
					result, err := ctx.Evaluate([]byte(`(() => { return "test"; })()`))
					if err != nil {
						t.Fatal()
					}
					return result, nil
				},
			},
			want: []byte(`"test"`),
		},
		{
			name: "object",
			args: args{
				value: func(ctx *gomonkey.Context) (*gomonkey.Value, error) {
					result, err := ctx.Evaluate([]byte(`(() => { return {"test":"value"}; })()`))
					if err != nil {
						t.Fatal()
					}
					return result, nil
				},
			},
			want: []byte(`{"test":"value"}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runtime.LockOSThread()
			defer runtime.UnlockOSThread()

			ctx, err := gomonkey.NewContext()
			if err != nil {
				t.Fatal()
			}
			defer ctx.Destroy()
			value, err := tt.args.value(ctx)
			if err != nil {
				t.Fatal()
			}
			defer value.Release()
			data, err := value.MarshalJSON()
			if err != nil {
				t.Errorf("v.MarshalJSON() err = %v, want %v", err, nil)
			}
			if !bytes.Equal(data, tt.want) {
				t.Errorf("v.MarshalJSON() got %v, want %v", data, tt.want)
			}
		})
	}
}

func TestValueAsValue(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	val := v.AsValue()
	if val == nil {
		t.Errorf("v.AsValue() = %v", val)
	}
}

func TestValueAsObject(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	o, err := gomonkey.NewObject(ctx)
	if err != nil {
		t.Fatal()
	}
	defer o.Release()

	v := o.AsValue()
	if v == nil {
		t.Fatal()
	}
	object, err := v.AsObject()
	if err != nil {
		t.Errorf("v.AsObject() err = %v, want %v", err, nil)
	}
	if object == nil {
		t.Errorf("v.AsObject() = %v", object)
	}
}

func TestValueAsFunction(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	f, err := gomonkey.NewFunction(ctx, "test", func(args []*gomonkey.Value) (*gomonkey.Value, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal()
	}
	defer f.Release()

	v := f.AsValue()
	if v == nil {
		t.Fatal()
	}
	fn, err := v.AsFunction()
	if err != nil {
		t.Errorf("v.AsFunction() err = %v, want %v", err, nil)
	}
	if fn == nil {
		t.Errorf("v.AsFunction() = %v", fn)
	}
}

func TestValueIs(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.Is(v); got != true {
		t.Errorf("v.Is() got %v, want %v", got, true)
	}
}

func TestValueIsUndefined(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueUndefined(ctx)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.IsUndefined(); got != true {
		t.Errorf("v.IsUndefined() got %v, want %v", got, true)
	}
}

func TestValueIsNull(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueNull(ctx)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.IsNull(); got != true {
		t.Errorf("v.IsNull() got %v, want %v", got, true)
	}
}

func TestValueIsNullOrUndefined(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v1, err := gomonkey.NewValueNull(ctx)
	if err != nil {
		t.Fatal()
	}
	defer v1.Release()
	v2, err := gomonkey.NewValueUndefined(ctx)
	if err != nil {
		t.Fatal()
	}
	defer v2.Release()

	if got := v1.IsNullOrUndefined(); got != true {
		t.Errorf("v.IsNullOrUndefined() got %v, want %v", got, true)
	}
	if got := v2.IsNullOrUndefined(); got != true {
		t.Errorf("v.IsNullOrUndefined() got %v, want %v", got, true)
	}
}

func TestValueIsTrue(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueBoolean(ctx, true)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.IsTrue(); got != true {
		t.Errorf("v.IsTrue() got %v, want %v", got, true)
	}
}

func TestValueIsFalse(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueBoolean(ctx, false)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.IsFalse(); got != true {
		t.Errorf("v.IsFalse() got %v, want %v", got, true)
	}
}

func TestValueIsString(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.IsString(); got != true {
		t.Errorf("v.IsString() got %v, want %v", got, true)
	}
}

func TestValueIsBoolean(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueBoolean(ctx, true)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.IsBoolean(); got != true {
		t.Errorf("v.IsBoolean() got %v, want %v", got, true)
	}
}

func TestValueIsNumber(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueNumber(ctx, 123.456)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.IsNumber(); got != true {
		t.Errorf("v.IsNumber() got %v, want %v", got, true)
	}
}

func TestValueIsInt32(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueInt32(ctx, 123)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.IsInt32(); got != true {
		t.Errorf("v.IsInt32() got %v, want %v", got, true)
	}
}

func TestValueToString(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueString(ctx, "test")
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.ToString(); got != "test" {
		t.Errorf("v.ToString() got %v, want %v", got, "test")
	}
}

func TestValueToBoolean(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v1, err := gomonkey.NewValueBoolean(ctx, true)
	if err != nil {
		t.Fatal()
	}
	defer v1.Release()
	v2, err := gomonkey.NewValueBoolean(ctx, false)
	if err != nil {
		t.Fatal()
	}
	defer v2.Release()

	if got := v1.ToBoolean(); got != true {
		t.Errorf("v.ToBoolean() got %v, want %v", got, true)
	}
	if got := v2.ToBoolean(); got != false {
		t.Errorf("v.ToBoolean() got %v, want %v", got, true)
	}
}

func TestValueToNumber(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueNumber(ctx, 123.456)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.ToNumber(); got != 123.456 {
		t.Errorf("v.ToNumber() got %v, want %v", got, 123.456)
	}
}

func TestValueToInt32(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	v, err := gomonkey.NewValueInt32(ctx, 123)
	if err != nil {
		t.Fatal()
	}
	defer v.Release()

	if got := v.ToInt32(); got != 123 {
		t.Errorf("v.ToInt32() got %v, want %v", got, 123)
	}
}
