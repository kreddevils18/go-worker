package asyncjob

import (
	gologger "github.com/kreddevils18/go-logger"
)

// ErrorHandler is a function that handles errors from a job.
type ErrorHandler[T any] func(payload T, err error)

// Option is a functional option for configuring the Dispatcher.
type Option[T any] func(*Dispatcher[T])

// WithMaxWorkers sets the number of concurrent workers.
// It must be a positive integer.
func WithMaxWorkers[T any](count int) Option[T] {
	return func(d *Dispatcher[T]) {
		d.MaxWorkers = count
	}
}

// WithQueueSize sets the size of the job queue.
// It must be a positive integer.
func WithQueueSize[T any](size int) Option[T] {
	return func(d *Dispatcher[T]) {
		d.QueueSize = size
	}
}

// WithErrorHandler provides a custom error handler function.
// If not set, errors will be logged by the dispatcher's logger.
func WithErrorHandler[T any](handler ErrorHandler[T]) Option[T] {
	return func(d *Dispatcher[T]) {
		d.errorHandler = handler
	}
}

// WithLogger provides a custom logger.
func WithLogger[T any](l gologger.Logger) Option[T] {
	return func(d *Dispatcher[T]) {
		d.logger = l
	}
}
