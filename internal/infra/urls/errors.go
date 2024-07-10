package urls

import "errors"

var (
	ErrNotFoundURL    = errors.New("the URL with the specified ID was not found")
	ErrURLGone        = errors.New("the URL with the specified ID was removed")
	ErrCorIDNotUnique = errors.New("the specified correlation ID is not unique")
	ErrNotOwner       = errors.New("the URL with the specified ID belongs to other user")
)

type ErrURINotUnique struct {
	ExistingKey string
	Msg         string
}

func (e *ErrURINotUnique) Error() string {
	return e.Msg
}

func newErrURINotUnique(key string) error {
	return &ErrURINotUnique{
		ExistingKey: key,
		Msg:         "the specified full URI is not unique",
	}
}
