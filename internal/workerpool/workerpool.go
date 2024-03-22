package workerpool

import "sync"

type Executer interface {
	Execute()
}

type WorkerPool[T Executer] struct {
	tasks       []T
	concurrency int
	tasksChan   chan Executer
	wg          sync.WaitGroup
}

// New creates a WorkerPool , which offers concurrent execution of tasks with a given number of workers (goroutines).
//
// Usage:
//
//		 ```go
//			type Task struct {URL string}
//
//			func (m *Task) Execute() {
//				fmt.Printf("Executing task https://%s\n", m.URL)
//			}
//
//			func main() {
//	     tasks := []*Task{{URL: "one.io"}, {URL: "two.io"}}
//				wp := New(tasks, 2)
//				wp.Run()
//			}
//		 ```
func New[T Executer](tasks []T, concurrency int) *WorkerPool[T] {
	return &WorkerPool[T]{
		tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan Executer, len(tasks)),
		wg:          sync.WaitGroup{},
	}
}

func (wp *WorkerPool[T]) work() {
	for task := range wp.tasksChan {
		task.Execute()
		wp.wg.Done()
	}
}

// Run starts the WorkerPool and blocks until all tasks have been executed.
func (wp *WorkerPool[T]) Run() {
	for i := 0; i < wp.concurrency; i++ {
		go wp.work()
	}

	wp.wg.Add(len(wp.tasks))
	for _, task := range wp.tasks {
		wp.tasksChan <- task
	}
	close(wp.tasksChan)

	wp.wg.Wait()
}
