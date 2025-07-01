# Go Worker (`asyncjob`)

A robust, generic, and type-safe Go package for managing asynchronous jobs using a worker pool pattern.

This package provides a clean and idiomatic Go solution for running background tasks, leveraging modern Go features like generics to ensure type safety for job payloads. It's designed to be easy to integrate, configure, and extend in any Go application.

[![Go Tests](https://github.com/kreddevils18/go-worker/actions/workflows/test.yml/badge.svg)](https://github.com/kreddevils18/go-worker/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/kreddevils18/go-worker)](https://goreportcard.com/report/github.com/kreddevils18/go-worker)

## Features

- **Type-Safe Payloads**: Uses Go generics (`[T any]`) to ensure that each job receives the exact data type it expects, preventing runtime type assertion errors.
- **Concurrent Worker Pool**: Manages a pool of goroutines to process jobs concurrently.
- **Configurable**: Easily configure the number of workers, queue size, and error handling logic using the Functional Options Pattern.
- **Graceful Shutdown**: A clean `Shutdown()` method ensures that all currently running jobs are completed before the application exits.
- **Dependency Injection Friendly**: Designed to work seamlessly with dependency injection, making your job logic clean and testable.
- **Custom Error Handling**: Provides an option to set a custom handler for errors returned by jobs.

## Installation

```sh
go get github.com/kreddevils18/go-worker/asyncjob
```

## Quick Start

Here's a simple example of how to use the `asyncjob` package.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kreddevils18/go-worker/asyncjob"
)

// 1. Define the data structure for your job's payload.
type EmailPayload struct {
	To      string
	Subject string
}

// 2. Implement the asyncjob.Job interface.
type EmailSenderJob struct{}

func (j *EmailSenderJob) Execute(ctx context.Context, payload EmailPayload) error {
	fmt.Printf("-> Sending email to %s...\n", payload.To)
	time.Sleep(1 * time.Second) // Simulate work
	fmt.Printf("-> Email sent to %s\n", payload.To)
	return nil
}

func main() {
	// 1. Create an instance of your job handler.
	emailJob := &EmailSenderJob{}

	// 2. Create a new dispatcher for your payload type.
	dispatcher, err := asyncjob.NewDispatcher[EmailPayload](
		emailJob,
		asyncjob.WithMaxWorkers[EmailPayload](3), // Use 3 concurrent workers
	)
	if err != nil {
		panic(err)
	}

	// 3. Submit jobs to the dispatcher.
	for i := 0; i < 10; i++ {
		payload := EmailPayload{To: fmt.Sprintf("user%d@example.com", i)}
		dispatcher.Submit(payload)
	}

	// 4. Shut down the dispatcher gracefully.
	// This will wait for all 10 jobs to complete.
	fmt.Println("Waiting for jobs to complete...")
	dispatcher.Shutdown()
	fmt.Println("Application finished.")
}
```

## Advanced Usage: Dependency Injection & Error Handling

For more complex applications, you'll want to inject dependencies into your job handlers and set up custom error handling.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kreddevils18/go-worker/asyncjob"
)

// --- Dependency Injection Setup ---

// 1. Define a contract for our dependency (e.g., a mail service).
type MailServiceClient interface {
	Send(ctx context.Context, to, subject, body string) error
}

// 2. Create a concrete implementation of the service.
type concreteMailServiceClient struct{}

func (c *concreteMailServiceClient) Send(ctx context.Context, to, subject, body string) error {
	fmt.Printf("-> Sending email via MailService to %s\n", to)
	time.Sleep(1 * time.Second)
	if to == "fail@example.com" {
		return fmt.Errorf("service failed for %s", to)
	}
	return nil
}

func NewMailServiceClient() MailServiceClient {
	return &concreteMailServiceClient{}
}

// --- Job Definition ---

// 3. Define the job payload.
type EmailPayload struct {
	To      string
	Subject string
	Body    string
}

// 4. Implement the job, now with the dependency.
type EmailSenderJob struct {
	client MailServiceClient
}

// NewEmailSenderJob is a constructor for our job, injecting the dependency.
func NewEmailSenderJob(client MailServiceClient) *EmailSenderJob {
	return &EmailSenderJob{client: client}
}

// The Execute method now delegates the work to the injected client.
func (j *EmailSenderJob) Execute(ctx context.Context, payload EmailPayload) error {
	return j.client.Send(ctx, payload.To, payload.Subject, payload.Body)
}

func main() {
	// 1. Initialize dependencies.
	mailService := NewMailServiceClient()

	// 2. Initialize the job handler, injecting the dependency.
	emailJob := NewEmailSenderJob(mailService)

	// 3. Define a custom error handler.
	customErrorHandler := func(payload EmailPayload, err error) {
		fmt.Printf("[!!!] FAILED JOB: Could not send email to '%s'. Reason: %v\n", payload.To, err)
	}

	// 4. Create the dispatcher with custom options.
	dispatcher, err := asyncjob.NewDispatcher[EmailPayload](
		emailJob,
		asyncjob.WithMaxWorkers[EmailPayload](3),
		asyncjob.WithErrorHandler[EmailPayload](customErrorHandler),
	)
	if err != nil {
		panic(err)
	}

	// 5. Submit jobs, including one that will fail.
	dispatcher.Submit(EmailPayload{To: "ok@example.com", Subject: "Hello"})
	dispatcher.Submit(EmailPayload{To: "fail@example.com", Subject: "This will fail"})
	dispatcher.Submit(EmailPayload{To: "another@example.com", Subject: "Hello Again"})

	// 6. Shutdown gracefully.
	dispatcher.Shutdown()
	fmt.Println("Application finished.")
}
```

## API

### `Job[T any]`

An interface that defines an executable job.

```go
type Job[T any] interface {
    Execute(ctx context.Context, payload T) error
}
```

### `NewDispatcher[T any]`

Creates and starts a new dispatcher.

```go
func NewDispatcher[T any](jobHandler Job[T], opts ...Option[T]) (*Dispatcher[T], error)
```

### Options

- `WithMaxWorkers[T any](count int)`: Sets the number of concurrent worker goroutines. Must be a positive integer. (Default: 10)
- `WithQueueSize[T any](size int)`: Sets the buffer size of the job queue. Must be a positive integer. (Default: 100)
- `WithErrorHandler[T any](handler func(payload T, err error))`: Provides a custom handler for errors returned by `job.Execute`.
- `WithLogger[T any](l gologger.Logger)`: Provides a custom logger that conforms to the `go-logger` interface.

### Methods

- `(d *Dispatcher[T]) Submit(payload T)`: Adds a new job payload to the queue for processing.
- `(d *Dispatcher[T]) Shutdown()`: Stops accepting new jobs and waits for all currently running jobs to complete.

## Contributing

Contributions are welcome! Please feel free to submit a pull request.

## License

This project is licensed under the MIT License.
