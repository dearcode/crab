package meta

import (
	"github.com/juju/errors"
)

var (
	//ErrNotFound db not found.
	ErrNotFound = errors.New("not found")
	//ErrArgNotPtr argument is not ptr.
	ErrArgNotPtr = errors.New("argument not ptr")
	//ErrFieldNotFound struct's field not found.
	ErrFieldNotFound = errors.New("struc's field not found")
	//ErrArgIsNil argument is nil.
	ErrArgIsNil = errors.New("argument is nil")
)
