package gomonkey

// #cgo CXXFLAGS: -std=c++17 -isystem ${SRCDIR}/deps/include/mozjs-115 -Wall -Wextra -Wno-mismatched-tags -DCGO
// #cgo LDFLAGS: -lmozjs-115
// #cgo freebsd,amd64 LDFLAGS: -L${SRCDIR}/deps/lib/freebsd_amd64/release/lib
// #cgo linux,amd64 LDFLAGS: -L${SRCDIR}/deps/lib/linux_amd64/release/lib
// #cgo netbsd,amd64 LDFLAGS: -L${SRCDIR}/deps/lib/netbsd_amd64/release/lib
// #cgo openbsd,amd64 LDFLAGS: -L${SRCDIR}/deps/lib/openbsd_amd64/release/lib
// #include "gomonkey.h"
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

// Init initializes SpiderMonkey.
func Init() {
	C.Init()
}

// ShutDown shutdowns SpiderMonkey.
func ShutDown() {
	C.ShutDown()
}

// Version returns the version of SpiderMonkey.
func Version() string {
	cVersion := C.Version()
	version := C.GoString(cVersion)
	C.free(unsafe.Pointer(cVersion))
	return version
}
