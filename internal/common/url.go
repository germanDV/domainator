package common

import (
	"errors"
	"fmt"
	"strings"
)

var ErrInvalidURL = errors.New("invalid URL")

type URL struct {
	value string
}

func ParseURL(url string) (URL, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "https://") {
		return URL{}, fmt.Errorf("error parsing url %s: %w", url, ErrInvalidURL)
	}
	return URL{value: url}, nil
}

func (url URL) String() string {
	return url.value
}
