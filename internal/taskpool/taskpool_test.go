package taskpool

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/goleak"
)

type out struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func TestPool(t *testing.T) {
	defer goleak.VerifyNone(t)

	out := out{buf: bytes.Buffer{}}
	pool := New(3)

	// Introduce a small delay afer each pool.Add to
	// ensure the order in which tasks are finished.
	pool.Add(func() { mock(1, &out) })
	time.Sleep(5 * time.Millisecond)

	pool.Add(func() { mock(2, &out) })
	time.Sleep(5 * time.Millisecond)

	pool.Add(func() { mock(3, &out) })
	time.Sleep(5 * time.Millisecond)

	pool.Add(func() { mock(4, &out) })
	time.Sleep(5 * time.Millisecond)

	pool.Add(func() { mock(5, &out) })
	time.Sleep(5 * time.Millisecond)

	pool.Wait()

	expectedOrder := []string{
		"Started #1",
		"Started #2",
		"Started #3",
		"Finished #1",
		"Started #4",
		"Finished #2",
		"Started #5",
		"Finished #3",
		"Finished #4",
		"Finished #5",
	}

	want := strings.Join(expectedOrder, "")
	got := out.buf.String()
	if want != got {
		t.Errorf("results do not match, want %v, got %v\n", want, got)
	}
}

func mock(id int, o *out) {
	o.mu.Lock()
	o.buf.WriteString(fmt.Sprintf("Started #%d", id))
	o.mu.Unlock()

	time.Sleep(250 * time.Millisecond)

	o.mu.Lock()
	o.buf.WriteString(fmt.Sprintf("Finished #%d", id))
	o.mu.Unlock()
}
