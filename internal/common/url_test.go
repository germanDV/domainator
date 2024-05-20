package common

import (
	"errors"
	"testing"
)

func TestParseURL(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input string
		want  URL
		err   error
	}{
		{"https://example.io", URL{value: "https://example.io"}, nil},
		{"https://go.dev", URL{value: "https://go.dev"}, nil},
		{"  https://go.dev ", URL{value: "https://go.dev"}, nil},
		{"http://go.dev", URL{}, ErrInvalidURL},
		{"debian.org", URL{}, ErrInvalidURL},
		{"", URL{}, ErrInvalidURL},
		{" ", URL{}, ErrInvalidURL},
		{"https://open.spotify.com", URL{value: "https://open.spotify.com"}, nil},
	}

	for _, tc := range tt {
		got, err := ParseURL(tc.input)
		if !errors.Is(err, tc.err) {
			t.Errorf("expected error %q but got %q", tc.err, err)
		}
		if got.value != tc.want.value {
			t.Errorf("expected %q but got %q", tc.want.value, got.value)
		}
	}
}
