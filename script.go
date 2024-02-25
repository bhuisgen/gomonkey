package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"

// Script represents a JS script.
type Script struct {
	ptr C.ScriptPtr
}

// Release releases the script.
func (s *Script) Release() {
	C.ReleaseScript(s.ptr)
}
