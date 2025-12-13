package container

import "context"

// IHostedService defines the interface for long-running background services.
type IHostedService interface {
	// Start is called to start the hosted service.
	// The provided context can be used to handle cancellation and timeouts.
	// If Start returns an error, the service is considered failed to start.
	Start(context.Context) error

	// Stop is called to gracefully stop the hosted service.
	// The provided context typically includes a timeout for the shutdown process.
	Stop(context.Context) error
}
