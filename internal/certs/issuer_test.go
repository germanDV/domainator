package certs

import (
	"errors"
	"testing"
)

func TestParseIssuer(t *testing.T) {
	t.Parallel()
	tt := []struct {
		input string
		want  Issuer
		err   error
	}{
		{"Let's Encrypt", Issuer{value: "Let's Encrypt"}, nil},
		{"  Let's Encrypt  ", Issuer{value: "Let's Encrypt"}, nil},
		{"", Issuer{}, ErrInvalidIssuer},
		{" ", Issuer{}, ErrInvalidIssuer},
	}

	for _, tc := range tt {
		got, err := ParseIssuer(tc.input)
		if !errors.Is(err, tc.err) {
			t.Errorf("expected error %q but got %q", tc.err, err)
		}
		if got.value != tc.want.value {
			t.Errorf("expected %q but got %q", tc.want.value, got.value)
		}
	}
}
