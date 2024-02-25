package gomonkey_test_error

import (
	"fmt"
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

func TestError(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()

	result, err := ctx.Evaluate([]byte(`
function add(a, b) { return a + c; }
add(1,2);
`))
	if err == nil {
		result.Release()
		t.Errorf("ctx.Evaluate() error = %s, want %v", err, nil)
	}
	if e, ok := err.(*gomonkey.JSError); !ok {
		t.Errorf("ctx.Evaluate() error type = %T, want *gomonkey.JSError{}", e)
	}
}

func TestErrorError(t *testing.T) {
	message := "test message"

	err := gomonkey.JSError{
		Message: message,
	}

	got := err.Error()
	if got != message {
		t.Errorf("err.Error() = %s, want %s", got, message)
	}
}

func TestErrorFormat(t *testing.T) {
	message := "test message"

	err := &gomonkey.JSError{
		Message: message,
	}

	s := fmt.Sprintf("%s", err)
	if s != message {
		t.Errorf("err.message = %s, want %s", s, message)
	}
	v := fmt.Sprintf("%v", err)
	if v != message {
		t.Errorf("err.message = %s, want %s", v, message)
	}
	q := fmt.Sprintf("%q", err)
	if q != "\"test message\"" {
		t.Errorf("err.message = %s, want %s", q, "test message")
	}
}
