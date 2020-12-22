package ttask

import (
	"context"
	"sync"
)

type Worker struct {
	input  chan Task
	output chan Task
}

func NewWorker() *Worker {
	worker := &Worker{
		input:  make(chan Task),
		output: make(chan Task),
	}

	return worker
}

func (worker *Worker) Run(ctx context.Context) {
	load := ctx.Value(ContextKey(KeyLoad))

	for {
		var tasks []Task

		select {
		case <-ctx.Done():
			return
		case task := <-worker.input:
			task.Run(ctx)

			if order, is := task.(Order); is {
				tasks = order.Breakdown(ctx)
			} else {
				tasks = nil
			}

			if load, is := load.(*sync.WaitGroup); is && load != nil {
				load.Add(len(tasks) - 1)
			}
		}

		for _, task := range tasks {
			select {
			case <-ctx.Done():
				return
			case worker.output <- task:
			}
		}
	}
}

func (worker *Worker) InputChan() chan<- Task {
	return worker.input
}

func (worker *Worker) OutputChan() <-chan Task {
	return worker.output
}
