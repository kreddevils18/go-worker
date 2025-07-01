### Active Context: `asyncjob` Package

#### Current Focus

The immediate goal is to begin the implementation of the `asyncjob` Go package based on the finalized design specified in `prd.md`.

#### Recent Changes

- The design and architecture for the `asyncjob` package have been documented and approved in `prd.md`.

#### Next Steps

1.  Set up the initial Go module and package structure for `asyncjob`.
2.  Implement the `Job[T any]` interface.
3.  Implement the `Dispatcher[T any]` struct with its core fields.
4.  Implement the `NewDispatcher` constructor, including the default settings and the application of functional options.
5.  Implement the worker logic as an internal function of the dispatcher.
6.  Implement the `Submit` and `Shutdown` methods.
7.  Create a basic example usage to test the implementation.
