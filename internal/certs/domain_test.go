package certs

import (
	"errors"
	"testing"
)

func TestParseDomain(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input string
		want  Domain
		err   error
	}{
		{"example.io", Domain{value: "example.io"}, nil},
		{"   example.io   ", Domain{value: "example.io"}, nil},
		{"sub.domain.dev", Domain{value: "sub.domain.dev"}, nil},
		{"", Domain{}, ErrInvalidDomain},
		{"  ", Domain{}, ErrInvalidDomain},
		{"incomplete.", Domain{}, ErrInvalidDomain},
		{"*.star", Domain{}, ErrInvalidDomain},
	}

	for _, tc := range tt {
		got, err := ParseDomain(tc.input)
		if !errors.Is(err, tc.err) {
			t.Errorf("expected error %q but got %q", tc.err, err)
		}
		if got.value != tc.want.value {
			t.Errorf("expected %q but got %q", tc.want.value, got.value)
		}
	}
}
