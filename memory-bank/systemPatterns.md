### System Patterns: `asyncjob` Package

The `asyncjob` package is built on several core architectural principles and design patterns.

#### Core Components

1.  **`Job[T any]` Interface:** The central abstraction. Any task must implement this interface. It uses Go generics (`[T any]`) to define a specific, type-safe payload for each job.

    ```go
    type Job[T any] interface {
        Execute(ctx context.Context, payload T) error
    }
    ```

2.  **`Dispatcher[T any]` (Worker Pool):** The coordinator. It manages a pool of workers and distributes jobs to them. It's responsible for the lifecycle of the workers. A single dispatcher instance is tied to a single job type via the `Job[T]` handler, ensuring a clean design.

3.  **`Worker[T any]`:** An internal, non-public component representing a single goroutine that executes jobs. It listens on an internal job channel, executes the job using the handler, and respects the context for graceful shutdown.

#### Design Patterns & Concurrency Model

- **Worker Pool Pattern:** The `Dispatcher` implements this pattern to control concurrent execution. It maintains a fixed number of worker goroutines.
- **Functional Options Pattern:** Used for configuring the `Dispatcher` on initialization. This provides a clean, extensible, and readable API for setting parameters like the number of workers, queue size, and error handlers.
- **Decoupling via Interfaces:** The `Job` interface decouples the job submission logic from the execution logic. The dispatcher doesn't know the details of the job, and the job doesn't know about the worker.
- **Concurrency:**
  - **Channels** are used for communication: a public queue for incoming payloads and an internal queue for jobs ready for workers.
  - **`context.Context`** is used for graceful shutdown. Canceling the context signals all workers to stop processing new jobs and terminate.
  - **`sync.WaitGroup`** is used to ensure the main program waits for all workers to finish their current tasks before exiting.
