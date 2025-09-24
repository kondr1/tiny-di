package container

import "context"

// IHostedService defines the interface for long-running background services.
//
// Hosted services are managed by the dependency injection container and provide
// lifecycle hooks for starting and stopping background operations. They are
// typically used for background tasks, scheduled jobs, message queue processors,
// or other services that need to run continuously throughout the application lifecycle.
//
// Services implementing this interface should be registered using AddHostedService
// and will have their Start method called when the application begins and Stop
// method called during application shutdown.
//
// Example implementation:
//
//	type BackgroundProcessor struct {
//		logger ILogger
//		cancel context.CancelFunc
//	}
//
//	func (b *BackgroundProcessor) Init(logger ILogger) error {
//		b.logger = logger
//		return nil
//	}
//
//	func (b *BackgroundProcessor) Start(ctx context.Context) error {
//		b.logger.Log("Starting background processor")
//		processingCtx, cancel := context.WithCancel(ctx)
//		b.cancel = cancel
//
//		go b.processLoop(processingCtx)
//		return nil
//	}
//
//	func (b *BackgroundProcessor) Stop(ctx context.Context) error {
//		b.logger.Log("Stopping background processor")
//		if b.cancel != nil {
//			b.cancel()
//		}
//		return nil
//	}
type IHostedService interface {
	// Start is called to start the hosted service.
	// The provided context can be used to handle cancellation and timeouts.
	// If Start returns an error, the service is considered failed to start.
	Start(context.Context) error

	// Stop is called to gracefully stop the hosted service.
	// The provided context typically includes a timeout for the shutdown process.
	// Services should complete their work and release resources within the timeout.
	Stop(context.Context) error
}
