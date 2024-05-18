package urls

import "errors"

var (
	ErrNotFoundURL    = errors.New("the URL with the specified id was not found")
	ErrCorIDNotUnique = errors.New("the specified correlation ID is not unique")
)
