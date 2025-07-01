package asyncjob

import (
	"context"
	"io"
	"log"
	"sync"

	"github.com/go-playground/validator/v10"
	gologger "github.com/kreddevils18/go-logger"
)

const (
	DefaultMaxWorkers = 10
	DefaultQueueSize  = 100
)

// Dispatcher manages a pool of workers for asynchronous job processing.
// T is the type of payload this dispatcher will process.
type Dispatcher[T any] struct {
	MaxWorkers      int
	QueueSize       int
	jobPayloadQueue chan T
	jobHandler      Job[T]
	errorHandler    ErrorHandler[T]
	logger          gologger.Logger
	testWriter      io.Writer // For testing only
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
}

// dispatcherConfig is used for validation purposes.
type dispatcherConfig struct {
	MaxWorkers int `validate:"gte=1"`
	QueueSize  int `validate:"gte=1"`
}

// NewDispatcher creates and starts a new dispatcher.
// It returns an error if the configuration is not valid.
func NewDispatcher[T any](jobHandler Job[T], opts ...Option[T]) (*Dispatcher[T], error) {
	ctx, cancel := context.WithCancel(context.Background())
	defaultLogger, _ := gologger.NewDefaultLogger()

	d := &Dispatcher[T]{
		MaxWorkers: DefaultMaxWorkers,
		QueueSize:  DefaultQueueSize,
		jobHandler: jobHandler,
		logger:     defaultLogger,
		ctx:        ctx,
		cancel:     cancel,
	}

	for _, opt := range opts {
		opt(d)
	}

	d.jobPayloadQueue = make(chan T, d.QueueSize)

	validate := validator.New()
	config := dispatcherConfig{
		MaxWorkers: d.MaxWorkers,
		QueueSize:  d.QueueSize,
	}

	if err := validate.Struct(config); err != nil {
		cancel()
		return nil, err
	}

	// Start workers
	d.start()

	return d, nil
}

// start launches the worker goroutines.
func (d *Dispatcher[T]) start() {
	d.wg.Add(d.MaxWorkers)
	for i := 0; i < d.MaxWorkers; i++ {
		go d.worker()
	}
}

// worker is a goroutine that processes jobs from the queue.
func (d *Dispatcher[T]) worker() {
	defer d.wg.Done()
	for {
		select {
		case payload, ok := <-d.jobPayloadQueue:
			if !ok {
				// Channel closed
				return
			}
			if err := d.jobHandler.Execute(d.ctx, payload); err != nil {
				if d.errorHandler != nil {
					d.errorHandler(payload, err)
				} else if d.logger != nil {
					d.logger.Errorw("asyncjob: error executing job", "error", err, "payload", payload)
				} else {
					log.Printf("asyncjob: error executing job for payload %#v: %v", payload, err)
				}
			}
		case <-d.ctx.Done():
			// Dispatcher is shutting down
			return
		}
	}
}

// Submit adds a job payload to the queue for processing.
func (d *Dispatcher[T]) Submit(payload T) {
	d.jobPayloadQueue <- payload
}

// Shutdown stops the dispatcher gracefully.
// It stops accepting new jobs and waits for all running jobs to complete.
func (d *Dispatcher[T]) Shutdown() {
	// Stop accepting new jobs
	close(d.jobPayloadQueue)

	// Wait for all workers to finish
	d.wg.Wait()

	// Cancel the context to clean up any other resources if needed.
	d.cancel()
}
