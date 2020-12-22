package ttask

import "context"

type Queue struct {
	tasks  []Task
	input  chan Task
	output chan Task
}

func NewQueue() *Queue {
	queue := &Queue{
		input:  make(chan Task),
		output: make(chan Task),
	}

	return queue
}

func (queue *Queue) outputChan() chan Task {
	if len(queue.tasks) > 0 {
		return queue.output
	}
	return nil
}

func (queue *Queue) outputTask() Task {
	if len(queue.tasks) > 0 {
		return queue.tasks[0]
	}
	return nil
}

func (queue *Queue) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-queue.input:
			queue.tasks = append(queue.tasks, task)
		case queue.outputChan() <- queue.outputTask():
			queue.tasks = queue.tasks[1:]
		}
	}
}

func (queue *Queue) OutputChan() <-chan Task {
	return queue.output
}

func (queue *Queue) InputChan() chan<- Task {
	return queue.input
}
