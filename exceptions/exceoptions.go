package exceptions

import (
	"fmt"
)

var (
	ErrSysNotImplemented = fmt.Errorf("not implemented")
)

var (
	// ErrDatabaseNotOpen is returned when a DB instance is accessed before it
	// is opened or after it is closed.
	ErrDatabaseNotOpen = fmt.Errorf("database not open")
)
