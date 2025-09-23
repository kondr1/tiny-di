package container

import "context"

type IHostedService interface {
	Start(context.Context) error
	Stop(context.Context) error
}
