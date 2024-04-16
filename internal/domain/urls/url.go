package urls

import (
	"errors"
	"math/rand"
	"net/url"
	"time"
)

type URL struct {
	id      string
	longURL string
}

var (
	ErrInvalidURL = errors.New("please enter a valid URL")
)

// Creates a new URL object with the URL provided
func NewURLWithoutID(longURL string) (*URL, error) {
	id := generateRandomID(8)
	return NewURL(longURL, id)
}

// Create a new URL object when you know which ID to use
func NewURL(longURL string, id string) (*URL, error) {
	if err := validateURL(longURL); err != nil {
		return nil, err
	}

	return &URL{
		id:      id,
		longURL: longURL,
	}, nil
}

func (u *URL) ID() string {
	return u.id
}

func (u *URL) LongURL() string {
	return u.longURL
}

func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || string(u.Host[0]) == "." || string(u.Host[len(u.Host)-1]) == "." {
		return ErrInvalidURL
	} else {
		return nil
	}
}

func generateRandomID(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, n)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
