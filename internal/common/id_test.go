package common

import (
	"errors"
	"testing"
)

func TestParseID(t *testing.T) {
	t.Parallel()
	id := NewID()

	tt := []struct {
		input string
		want  ID
		err   error
	}{
		{id.value, id, nil},
		{"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11", ID{value: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"}, nil},
		{" a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11 ", ID{value: "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"}, nil},
		{"9c0b-4ef8-bb6d-6bb9bd380a11", ID{}, ErrInvalidID},
		{"wtf", ID{}, ErrInvalidID},
		{"", ID{}, ErrInvalidID},
	}

	for _, tc := range tt {
		got, err := ParseID(tc.input)
		if !errors.Is(err, tc.err) {
			t.Errorf("expected error %q but got %q", tc.err, err)
		}
		if got.value != tc.want.value {
			t.Errorf("expected %q but got %q", tc.want.value, got.value)
		}
	}
}
