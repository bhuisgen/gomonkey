package gomonkey_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/bhuisgen/gomonkey"
)

func TestMain(m *testing.M) {
	gomonkey.Init()
	code := m.Run()
	gomonkey.ShutDown()
	os.Exit(code)
}

func TestVersion(t *testing.T) {
	rgx := regexp.MustCompile(`^\d+\.\d+$`)
	v := gomonkey.Version()
	if !rgx.MatchString(v) {
		t.Errorf("version string is in the incorrect format: %s", v)
	}
}
