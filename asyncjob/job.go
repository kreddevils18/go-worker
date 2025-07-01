package asyncjob

import "context"

// Job is an interface that defines an executable unit of work.
// T is the type of payload that the job will process.
type Job[T any] interface {
	// Execute contains the job execution logic.
	// It takes a context for handling timeout or cancellation,
	// and a payload of type T.
	Execute(ctx context.Context, payload T) error
}
