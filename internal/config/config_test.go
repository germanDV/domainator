package config

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func setup() {
	os.Setenv("STR_VAR", "testing")
	os.Setenv("INT_VAR", "42")
	os.Setenv("INT_VAR_BAD", "nan")
	os.Setenv("BOOL_VAR_LOWER", "false")
	os.Setenv("BOOL_VAR_UPPER", "FALSE")
	os.Setenv("BOOL_VAR_NUM", "1")
	os.Setenv("BOOL_VAR_BAD", "not_bool")
	os.Setenv("DUR_VAR_S", "5s")
	os.Setenv("DUR_VAR_M", "15m")
	os.Setenv("DUR_VAR_H", "12h")
}

func TestGetString(t *testing.T) {
	t.Run("gets_from_env", func(t *testing.T) {
		got := GetString("STR_VAR")
		want := "testing"
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("panics_when_not_set", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		GetString("NOT_SET")
	})
}

func TestGetInt(t *testing.T) {
	t.Run("gets_as_int", func(t *testing.T) {
		got := GetInt("INT_VAR")
		want := 42
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("panics_when_not_castable", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		GetInt("INT_VAR_BAD")
	})
}

func TestGetBool(t *testing.T) {
	t.Run("gets_lower_as_bool", func(t *testing.T) {
		got := GetBool("BOOL_VAR_LOWER")
		want := false
		if got != want {
			t.Errorf("got %t want %t", got, want)
		}
	})

	t.Run("gets_upper_as_bool", func(t *testing.T) {
		got := GetBool("BOOL_VAR_UPPER")
		want := false
		if got != want {
			t.Errorf("got %t want %t", got, want)
		}
	})

	t.Run("gets_num_as_bool", func(t *testing.T) {
		got := GetBool("BOOL_VAR_NUM")
		want := true
		if got != want {
			t.Errorf("got %t want %t", got, want)
		}
	})

	t.Run("panics_when_not_castable", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		GetBool("BOOL_VAR_BAD")
	})
}

func TestGetDuration(t *testing.T) {
	t.Run("gets_duration_secs", func(t *testing.T) {
		got := GetDuration("DUR_VAR_S")
		want := time.Duration(5) * time.Second
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("gets_duration_mins", func(t *testing.T) {
		got := GetDuration("DUR_VAR_M")
		want := time.Duration(15) * time.Minute
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("gets_duration_hours", func(t *testing.T) {
		got := GetDuration("DUR_VAR_H")
		want := time.Duration(12) * time.Hour
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	})

	t.Run("panics_when_not_castable", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic")
			}
		}()
		GetDuration("DUR_VAR_BAD")
	})
}
