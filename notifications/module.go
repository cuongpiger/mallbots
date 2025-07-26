package notifications

import (
	"context"

	"github.com/cuongpiger/mallbots/customers/customerspb"
	"github.com/cuongpiger/mallbots/internal/am"
	"github.com/cuongpiger/mallbots/internal/ddd"
	"github.com/cuongpiger/mallbots/internal/jetstream"
	"github.com/cuongpiger/mallbots/internal/monolith"
	"github.com/cuongpiger/mallbots/internal/registry"
	"github.com/cuongpiger/mallbots/notifications/internal/application"
	"github.com/cuongpiger/mallbots/notifications/internal/grpc"
	"github.com/cuongpiger/mallbots/notifications/internal/handlers"
	"github.com/cuongpiger/mallbots/notifications/internal/logging"
	"github.com/cuongpiger/mallbots/notifications/internal/postgres"
	"github.com/cuongpiger/mallbots/ordering/orderingpb"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = customerspb.Registrations(reg); err != nil {
		return err
	}
	if err = orderingpb.Registrations(reg); err != nil {
		return err
	}
	eventStream := am.NewEventStream(reg, jetstream.NewStream(mono.Config().Nats.Stream, mono.JS(), mono.Logger()))
	conn, err := grpc.Dial(ctx, mono.Config().Rpc.Address())
	if err != nil {
		return err
	}
	customers := postgres.NewCustomerCacheRepository("notifications.customers_cache", mono.DB(), grpc.NewCustomerRepository(conn))

	// setup application
	app := logging.LogApplicationAccess(
		application.New(customers),
		mono.Logger(),
	)
	integrationEventHandlers := logging.LogEventHandlerAccess[ddd.Event](
		handlers.NewIntegrationEventHandlers(app, customers),
		"IntegrationEvents", mono.Logger(),
	)

	// setup Driver adapters
	if err := grpc.RegisterServer(ctx, app, mono.RPC()); err != nil {
		return err
	}
	if err = handlers.RegisterIntegrationEventHandlers(eventStream, integrationEventHandlers); err != nil {
		return err
	}

	return nil
}
