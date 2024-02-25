package gomonkey_test_script

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

func TestScriptRelease(t *testing.T) {
	ctx, err := gomonkey.NewContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	script, err := ctx.CompileScript("script.js", []byte("(() => { return 42; })()"))
	if err != nil {
		t.Fatal()
	}

	script.Release()
}
