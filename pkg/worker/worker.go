package worker

import (
	"context"
	"errors"
	"log/slog"

	"github.com/gateixeira/gei-migration-helper/pkg/logging"
)

type Processor func(interface{}, context.Context) error

type Error struct {
	Err    error
	Entity interface{}
}

type Worker struct {
	Processor Processor
	jobs      chan interface{}
	results   chan<- Error
}

func New(processor Processor, jobs chan interface{}, results chan<- Error) (*Worker, error) {
	if processor == nil {
		return nil, errors.New("processor is nil")
	}

	return &Worker{
		Processor: processor,
		jobs:      jobs,
		results:   results,
	}, nil
}

func (w *Worker) Start(ctx context.Context) {
	slog.Debug("worker started", "id", ctx.Value(logging.IDKey))
	for entity := range w.jobs {
		slog.Debug("job received")
		err := w.Processor(entity, ctx)
		w.results <- Error{err, entity}
		slog.Debug("job finished")
	}
	slog.Debug("worker finished", "id", ctx.Value(logging.IDKey))
}
