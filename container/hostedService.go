package container

import "context"

type IHostedService interface {
	Start(context.Context) error
	Stop(context.Context) error
}

// example of a hosted service
type BackGroundService struct{}

func (b *BackGroundService) Init() error {
	// constructor
	return nil
}

func (b *BackGroundService) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return b.Stop(ctx)
		default:
			continue
		}
	}
}

func (b *BackGroundService) Stop(ctx context.Context) error {
	// dispose resources, log stopping
	return nil
}
