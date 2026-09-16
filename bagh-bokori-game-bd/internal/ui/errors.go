package ui

import "errors"

// errorsIs is a tiny indirection so friendlyError reads cleanly.
func errorsIs(err, target error) bool { return errors.Is(err, target) }
