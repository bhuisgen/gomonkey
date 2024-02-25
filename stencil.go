package gomonkey

// #include "gomonkey.h"
// #include <stdlib.h>
import "C"

// Stencil implements a JS stencil.
type Stencil struct {
	ptr C.StencilPtr
}

// Release releases the stencil.
func (s *Stencil) Release() {
	C.ReleaseStencil(s.ptr)
}
