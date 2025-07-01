package asyncjob_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kreddevils18/go-worker/asyncjob"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock Job để sử dụng trong các bài test.
type mockJob[T any] struct {
	executeFunc func(ctx context.Context, payload T) error
}

func (j *mockJob[T]) Execute(ctx context.Context, payload T) error {
	if j.executeFunc != nil {
		return j.executeFunc(ctx, payload)
	}
	return nil
}

func TestJobIsExecuted(t *testing.T) {
	// Arrange
	var executed bool
	var mu sync.Mutex
	done := make(chan struct{})

	mock := &mockJob[string]{
		executeFunc: func(ctx context.Context, payload string) error {
			mu.Lock()
			executed = true
			mu.Unlock()
			close(done)
			return nil
		},
	}

	dispatcher, err := asyncjob.NewDispatcher[string](mock)
	require.NoError(t, err)
	require.NotNil(t, dispatcher)

	// Action
	dispatcher.Submit("test-payload")

	// Assert
	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for job to be executed")
	}

	mu.Lock()
	assert.True(t, executed, "Job's Execute method should have been called")
	mu.Unlock()
}

func TestShutdown_WaitsForRunningJobs(t *testing.T) {
	// Arrange
	var jobStarted, jobFinished bool
	var mu sync.Mutex
	jobDuration := 100 * time.Millisecond

	mock := &mockJob[string]{
		executeFunc: func(ctx context.Context, payload string) error {
			mu.Lock()
			jobStarted = true
			mu.Unlock()

			time.Sleep(jobDuration)

			mu.Lock()
			jobFinished = true
			mu.Unlock()
			return nil
		},
	}

	dispatcher, err := asyncjob.NewDispatcher[string](mock, asyncjob.WithMaxWorkers[string](1))
	require.NoError(t, err)

	// Action
	dispatcher.Submit("test")
	// Give the job a moment to start
	time.Sleep(jobDuration / 2)

	shutdownStartTime := time.Now()
	dispatcher.Shutdown()
	shutdownDuration := time.Since(shutdownStartTime)

	// Assert
	mu.Lock()
	assert.True(t, jobStarted, "Job should have started before shutdown was called")
	assert.True(t, jobFinished, "Job should have finished because shutdown waits for it")
	mu.Unlock()

	// Assert that shutdown took roughly the remaining time.
	// It should be more than the first half, but not excessively long.
	assert.InDelta(t, (jobDuration / 2).Milliseconds(), shutdownDuration.Milliseconds(), 20, "Shutdown should wait for roughly the remainder of the job")
}

func TestErrorHandler_IsCalledOnError(t *testing.T) {
	// Arrange
	var receivedPayload string
	var receivedError error
	var mu sync.Mutex
	done := make(chan struct{})

	expectedError := errors.New("job failed")
	expectedPayload := "error-payload"

	mock := &mockJob[string]{
		executeFunc: func(ctx context.Context, payload string) error {
			return expectedError
		},
	}

	errorHandler := func(payload string, err error) {
		mu.Lock()
		receivedPayload = payload
		receivedError = err
		mu.Unlock()
		close(done)
	}

	dispatcher, err := asyncjob.NewDispatcher[string](
		mock,
		asyncjob.WithErrorHandler[string](errorHandler),
	)
	require.NoError(t, err)

	// Action
	dispatcher.Submit(expectedPayload)

	// Assert
	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for error handler to be called")
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, expectedPayload, receivedPayload, "Error handler should receive the correct payload")
	assert.Equal(t, expectedError, receivedError, "Error handler should receive the correct error")
}

func TestNewDispatcher_DefaultSettings(t *testing.T) {
	// Action
	dispatcher, err := asyncjob.NewDispatcher[string](&mockJob[string]{})

	// Assert
	require.NoError(t, err, "NewDispatcher should not return an error with default settings")
	require.NotNil(t, dispatcher, "Dispatcher should not be nil")

	assert.Equal(t, asyncjob.DefaultMaxWorkers, dispatcher.MaxWorkers, "Should use default max workers")
	assert.Equal(t, asyncjob.DefaultQueueSize, dispatcher.QueueSize, "Should use default queue size")
}

func TestNewDispatcher_WithOptions(t *testing.T) {
	// Arrange
	customWorkers := 5
	customQueueSize := 50

	// Action
	dispatcher, err := asyncjob.NewDispatcher[string](
		&mockJob[string]{},
		asyncjob.WithMaxWorkers[string](customWorkers),
		asyncjob.WithQueueSize[string](customQueueSize),
	)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, dispatcher)
	assert.Equal(t, customWorkers, dispatcher.MaxWorkers, "Should use custom max workers")
	assert.Equal(t, customQueueSize, dispatcher.QueueSize, "Should use custom queue size")
}

func TestNewDispatcher_InvalidConfiguration(t *testing.T) {
	t.Run("zero workers", func(t *testing.T) {
		// Action
		dispatcher, err := asyncjob.NewDispatcher[string](
			&mockJob[string]{},
			asyncjob.WithMaxWorkers[string](0),
		)

		// Assert
		assert.Error(t, err, "Expected an error for zero workers")
		assert.Nil(t, dispatcher, "Dispatcher should be nil on error")
	})

	t.Run("negative workers", func(t *testing.T) {
		// Action
		dispatcher, err := asyncjob.NewDispatcher[string](
			&mockJob[string]{},
			asyncjob.WithMaxWorkers[string](-1),
		)

		// Assert
		assert.Error(t, err, "Expected an error for negative workers")
		assert.Nil(t, dispatcher, "Dispatcher should be nil on error")
	})
}
