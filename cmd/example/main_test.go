package main

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

func TestContexts(t *testing.T) {
	if err := contexts(); err != nil {
		t.Errorf("invalid code sample, got err: %s", err)
	}
}

func TestObjects(t *testing.T) {
	if err := objects(); err != nil {
		t.Errorf("invalid code sample, got error: %s", err)
	}
}

func TestEvaluate(t *testing.T) {
	if err := evaluate(); err != nil {
		t.Errorf("invalid code, got error: %s", err)
	}
}

func TestExecuteScript(t *testing.T) {
	if err := executeScript(); err != nil {
		t.Errorf("invalid code, got error: %s", err)
	}
}

func TestExecuteScriptFromStencil(t *testing.T) {
	if err := executeScriptFromStencil(); err != nil {
		t.Errorf("invalid code, got error: %s", err)
	}
}

func TestCreateFunctionGlobal(t *testing.T) {
	if err := createGlobalFunction(); err != nil {
		t.Errorf("invalid code, got error: %s", err)
	}
}

func TestCreateFunctionObject(t *testing.T) {
	if err := createFunctionObject(); err != nil {
		t.Errorf("invalid code, got error: %s", err)
	}
}

func TestCreateObjectMethod(t *testing.T) {
	if err := createObjectMethod(); err != nil {
		t.Errorf("invalid code, got error: %s", err)
	}
}
