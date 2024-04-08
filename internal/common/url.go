package common

import (
	"errors"
	"strings"
)

var ErrInvalidURL = errors.New("invalid URL")

type URL struct {
	value string
}

func ParseURL(url string) (URL, error) {
	if !strings.HasPrefix(url, "https://") {
		return URL{}, ErrInvalidURL
	}
	return URL{value: url}, nil
}

func (url URL) String() string {
	return url.value
}
