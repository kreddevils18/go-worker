### Progress: `asyncjob` Package

#### Current Status

The project is in the initial design phase. A detailed Product Requirements Document (`prd.md`) has been created, outlining the full architecture, API, and core components.

- **What Works:** This is a new project. No code has been implemented yet.
- **What's Left to Build:** The entire `asyncjob` package, including:
  - `Job[T]` interface definition.
  - `Dispatcher[T]` struct and its methods (`NewDispatcher`, `Submit`, `Shutdown`).
  - Internal worker logic.
  - Functional options for configuration (`WithMaxWorkers`, `WithQueueSize`, etc.).
- **Known Issues:** None at this stage.

#### Future Enhancements (Post-MVP)

The following features are considered for future versions:

- Jobs with return values (Futures).
- Retry and backoff policies.
- Metrics and observability hooks.
- Priority job queues.
- Persistence layer for job queues.
