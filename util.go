package ttask

import (
	"context"
)

type ContextKey int

const (
	KeyLoad ContextKey = iota + 1
)

type Sink interface {
	InputChan() chan<- Task
}

type Source interface {
	OutputChan() <-chan Task
}

func Connect(ctx context.Context, sink Sink, source Source) {
	for {
		select {
		case <-ctx.Done():
			return
		case sink.InputChan() <- <-source.OutputChan():
		}
	}
}

type ChanSink struct {
	ch chan Task
}

func MakeSink(ch chan Task) *ChanSink {
	return &ChanSink{
		ch: ch,
	}
}

func (sink *ChanSink) InputChan() chan<- Task {
	return sink.ch
}

type ChanSource struct {
	ch chan Task
}

func MakeSource(ch chan Task) *ChanSource {
	return &ChanSource{
		ch: ch,
	}
}

func (source *ChanSource) OutputChan() <-chan Task {
	return source.ch
}

type NoOp struct{}

func (_ NoOp) Run(_ context.Context) {}
