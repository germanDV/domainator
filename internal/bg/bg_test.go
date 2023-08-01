package bg

import (
	"domainator/internal/logger"
	"io"
	"os"
	"sync"
	"testing"
)

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func setup() {
	logger.Init(io.Discard, io.Discard)
}

func TestRun(t *testing.T) {
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
