package ttask

import (
	"context"
)

type Task interface {
	Run(context.Context)
}

type Order interface {
	Task
	Breakdown(context.Context) []Task
}
