package main

import (
	"context"
	"fmt"
	"time"

	"github.com/kreddevils18/go-worker/asyncjob"
)

// ADVANCED EXAMPLE: Dependency Injection
//
// 1. Define a contract for our dependency (the email service).
type MailServiceClient interface {
	Send(ctx context.Context, to, subject, body string) error
}

// 2. Create a concrete implementation of the service.
// This would be in its own package in a real application.
type concreteMailServiceClient struct {
	// This could have its own config, like an API key.
	apiKey string
}

func (c *concreteMailServiceClient) Send(ctx context.Context, to, subject, body string) error {
	fmt.Printf("-> Sending email to %s... (Subject: %s)\n", to, subject)

	// Simulate time-consuming work (e.g., calling email service API)
	time.Sleep(1 * time.Second)

	// Simulate a possible error
	if to == "fail@example.com" {
		fmt.Printf("-> ERROR sending email to %s!\n", to)
		return fmt.Errorf("email service not responding for recipient: %s", to)
	}

	fmt.Printf("-> Successfully sent email to %s!\n", to)
	return nil
}

// Factory function for our service.
func NewMailServiceClient(apiKey string) MailServiceClient {
	return &concreteMailServiceClient{apiKey: apiKey}
}

// 3. Define Payload (Data for the job)
type EmailPayload struct {
	To      string
	Subject string
	Body    string
}

// 4. Implement asyncjob.Job interface, now with the dependency.
type EmailSenderJob struct {
	// The job now holds a reference to the service dependency.
	client MailServiceClient
}

// NewEmailSenderJob is a constructor for our job, injecting the dependency.
func NewEmailSenderJob(client MailServiceClient) *EmailSenderJob {
	return &EmailSenderJob{client: client}
}

// Execute contains the job execution logic, now delegating the work.
func (j *EmailSenderJob) Execute(ctx context.Context, payload EmailPayload) error {
	// The job's responsibility is now to orchestrate, not to do the work itself.
	return j.client.Send(ctx, payload.To, payload.Subject, payload.Body)
}

func main() {
	fmt.Println("--- Starting application ---")

	// 1. Initialize dependencies
	mailService := NewMailServiceClient("dummy-api-key-123")

	// 2. Initialize job handler, injecting the dependency
	emailJob := NewEmailSenderJob(mailService)

	// 3. Define a custom error handler
	customErrorHandler := func(payload EmailPayload, err error) {
		fmt.Printf("[!!!] BUSINESS ERROR: Cannot process payload for '%s'. Reason: %v\n", payload.To, err)
		// Here you can log, send to another queue for retry, etc.
	}

	// 4. Initialize dispatcher with EmailPayload type
	dispatcher, err := asyncjob.NewDispatcher[EmailPayload](
		emailJob,
		asyncjob.WithMaxWorkers[EmailPayload](3),  // 3 concurrent workers
		asyncjob.WithQueueSize[EmailPayload](100), // Queue holds 100 jobs
		asyncjob.WithErrorHandler[EmailPayload](customErrorHandler), // Assign error handler
	)
	if err != nil {
		panic(fmt.Sprintf("Cannot initialize dispatcher: %v", err))
	}

	fmt.Println("--- Dispatcher started. Beginning to send jobs... ---")

	// 5. Send jobs into the system
	for i := 0; i < 10; i++ {
		recipient := fmt.Sprintf("user%d@example.com", i)
		// This job will fail
		if i == 5 {
			recipient = "fail@example.com"
		}

		payload := EmailPayload{
			To:      recipient,
			Subject: fmt.Sprintf("Welcome Email #%d", i),
			Body:    "Hello!",
		}
		fmt.Printf("Sending job for: %s\n", payload.To)
		dispatcher.Submit(payload)
	}

	fmt.Println("--- All jobs sent. Waiting for processing... ---")

	// Give jobs time to process.
	time.Sleep(5 * time.Second)

	// 6. Shutdown the system safely
	fmt.Println("--- Starting shutdown process... ---")
	dispatcher.Shutdown()
	fmt.Println("--- Application has ended. ---")
}
