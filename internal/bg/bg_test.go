package bg

import (
	"log/slog"
	"os"
	"sync"
	"testing"
)

func TestRun(t *testing.T) {
	Init(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	t.Run("runs task in goroutine successfully", func(t *testing.T) {
		ch := make(chan int)
		ans := 42
		Run(func() { ch <- ans })
		got := <-ch
		if got != ans {
			t.Errorf("got %d, want %d", got, ans)
		}
	})

	t.Run("runs task in goroutine and recovers from panic", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(1)

		// If Run does not recover from panic, the test will fail.
		Run(func() {
			defer func() {
				wg.Done()
			}()
			panic("oops")
		})

		wg.Wait()
	})
}
