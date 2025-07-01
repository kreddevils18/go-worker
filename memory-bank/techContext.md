### Tech Context: `asyncjob` Package

#### Technology Stack

- **Language:** Go (version 1.18+ is required due to the use of Generics).

#### Core Go Features

- **Go Generics:** Used extensively to provide type safety for job payloads, eliminating the need for unsafe `interface{}` type assertions.
- **Goroutines & Channels:** The fundamental building blocks for concurrency in the worker pool.
- **`context` package:** For managing request-scoped data, cancellation signals, and deadlines, enabling graceful shutdown.
- **`sync` package:** Specifically `sync.WaitGroup` to coordinate the termination of goroutines.

#### Usage Example

This demonstrates how a developer would use the `asyncjob` package.

**1. Define a Payload and Job:**

```go
// Define the data structure for the job
type EmailPayload struct {
    To      string
    Subject string
    Body    string
}

// Implement the asyncjob.Job interface
type EmailSenderJob struct {}

func (j *EmailSenderJob) Execute(ctx context.Context, payload EmailPayload) error {
    fmt.Printf("Sending email to %s...\n", payload.To)
    // Real email sending logic here
    fmt.Printf("Email to %s sent.\n", payload.To)
    return nil
}
```

**2. Initialize and Use the Dispatcher:**

```go
func main() {
    // Create a job handler instance
    emailJob := &EmailSenderJob{}

    // Create and configure a new dispatcher for EmailPayload jobs
    dispatcher := asyncjob.NewDispatcher[EmailPayload](
        emailJob,
        asyncjob.WithMaxWorkers[EmailPayload](5),
        asyncjob.WithQueueSize[EmailPayload](100),
    )

    // Submit jobs
    for i := 0; i < 20; i++ {
        payload := EmailPayload{To: fmt.Sprintf("user%d@example.com", i)}
        dispatcher.Submit(payload)
    }

    // Shutdown the system gracefully
    dispatcher.Shutdown()
}
```
