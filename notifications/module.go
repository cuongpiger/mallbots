package notifications

import (
	"context"

	"github.com/cuongpiger/mallbots/internal/monolith"
	"github.com/cuongpiger/mallbots/notifications/internal/application"
	"github.com/cuongpiger/mallbots/notifications/internal/grpc"
	"github.com/cuongpiger/mallbots/notifications/internal/logging"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono monolith.Monolith) error {
	// setup Driven adapters
	conn, err := grpc.Dial(ctx, mono.Config().Rpc.Address())
	if err != nil {
		return err
	}
	customers := grpc.NewCustomerRepository(conn)

	// setup application
	var app application.App
	app = application.New(customers)
	app = logging.LogApplicationAccess(app, mono.Logger())

	// setup Driver adapters
	if err := grpc.RegisterServer(ctx, app, mono.RPC()); err != nil {
		return err
	}

	return nil
}
