package ttask

import (
	"context"
	"sync"
	"testing"
)

type TestTask struct {
	n int
}

func (task *TestTask) Run(ctx context.Context) {}

func (task *TestTask) String() string {
	return "test task"
}

func TestConnect(t *testing.T) {
	sinkCh := make(chan Task)
	sourceCh := make(chan Task)

	sink := MakeSink(sinkCh)
	source := MakeSource(sourceCh)

	ctx, cancel := context.WithCancel(context.Background())

	go Connect(ctx, sink, source)

	go func() {
		sourceCh <- &TestTask{n: 0}
		sourceCh <- &TestTask{n: 1}
	}()

	r := <-sinkCh
	if r.(*TestTask).n != 0 {
		t.Error()
	}

	cancel()
}

func TestQueue(t *testing.T) {
	queue := NewQueue()

	ctx, cancel := context.WithCancel(context.Background())

	go queue.Run(ctx)

	for i := 0; i < 5; i++ {
		queue.input <- &TestTask{}
	}

	for i := 0; i < 4; i++ {
		<-queue.output
	}

	cancel()
}

type TestWaitTask struct {
	ch chan struct{}
}

func (task *TestWaitTask) Run(ctx context.Context) {
	close(task.ch)
}

func (task *TestWaitTask) Wait() {
	<-task.ch
}

func TestWorker(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	sourceCh := make(chan Task)
	source := MakeSource(sourceCh)

	worker := NewWorker()
	go worker.Run(ctx)

	go Connect(ctx, worker, source)

	task1 := &TestWaitTask{make(chan struct{})}
	task2 := &TestWaitTask{make(chan struct{})}

	go func() {
		sourceCh <- task1
		sourceCh <- task2
	}()

	task1.Wait()

	cancel()
}

func TestLoad(t *testing.T) {
	ctx := context.Background()

	load := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, KeyLoad, load)

	sourceCh := make(chan Task)
	source := MakeSource(sourceCh)

	worker := NewWorker()
	go worker.Run(ctx)

	go Connect(ctx, worker, source)

	load.Add(5)

	for i := 0; i < 5; i++ {
		go func() {
			sourceCh <- &TestTask{}
		}()
	}

	load.Wait()
}

func TestPipeline(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	sourceCh := make(chan Task)
	source := MakeSource(sourceCh)

	queue1 := NewQueue()
	go queue1.Run(ctx)

	queue2 := NewQueue()
	go queue2.Run(ctx)

	worker := NewWorker()
	go worker.Run(ctx)

	go Connect(ctx, queue1, source)
	go Connect(ctx, queue2, queue1)
	go Connect(ctx, worker, queue2)

	task1 := &TestWaitTask{make(chan struct{})}
	task2 := &TestWaitTask{make(chan struct{})}

	go func() {
		sourceCh <- task1
		sourceCh <- task2
	}()

	task1.Wait()

	cancel()
}

type TestBreakdownTask struct {
	n  int
	ch chan int
}

func (task *TestBreakdownTask) Run(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case task.ch <- task.n:
	}
}

func (task *TestBreakdownTask) Breakdown(ctx context.Context) []Task {
	if task.n > 0 {
		return []Task{
			&TestBreakdownTask{n: task.n - 1, ch: task.ch},
			&TestBreakdownTask{n: task.n - 1, ch: task.ch},
		}
	}
	return nil
}

func TestBreakdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	load := &sync.WaitGroup{}
	ctx = context.WithValue(ctx, KeyLoad, load)

	sourceCh := make(chan Task)
	source := MakeSource(sourceCh)

	queue := NewQueue()
	go queue.Run(ctx)

	worker := NewWorker()
	go worker.Run(ctx)

	go Connect(ctx, queue, source)
	go Connect(ctx, worker, queue)
	go Connect(ctx, queue, worker)

	var total int = 0
	iCh := make(chan int)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case n := <-iCh:
				total += n
			}
		}
	}()

	load.Add(1)

	sourceCh <- &TestBreakdownTask{n: 5, ch: iCh}

	load.Wait()

	cancel()

	if total != 57 {
		t.Error()
	}
}
