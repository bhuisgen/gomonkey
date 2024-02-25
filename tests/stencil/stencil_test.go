package gomonkey_test_stencil

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

func TestStencilRelease(t *testing.T) {
	ctx, err := gomonkey.NewFrontendContext()
	if err != nil {
		t.Fatal()
	}
	defer ctx.Destroy()
	stencil, err := ctx.CompileScriptToStencil("script.js", []byte("(() => { return 42; })()"))
	if err != nil {
		t.Fatal()
	}

	stencil.Release()
}
