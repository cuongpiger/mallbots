package customers

import (
	"context"

	"github.com/cuongpiger/mallbots/customers/customerspb"
	"github.com/cuongpiger/mallbots/customers/internal/application"
	"github.com/cuongpiger/mallbots/customers/internal/grpc"
	"github.com/cuongpiger/mallbots/customers/internal/handlers"
	"github.com/cuongpiger/mallbots/customers/internal/logging"
	"github.com/cuongpiger/mallbots/customers/internal/postgres"
	"github.com/cuongpiger/mallbots/customers/internal/rest"
	"github.com/cuongpiger/mallbots/internal/am"
	"github.com/cuongpiger/mallbots/internal/ddd"
	"github.com/cuongpiger/mallbots/internal/jetstream"
	"github.com/cuongpiger/mallbots/internal/monolith"
	"github.com/cuongpiger/mallbots/internal/registry"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = customerspb.Registrations(reg); err != nil {
		return err
	}
	eventStream := am.NewEventStream(reg, jetstream.NewStream(mono.Config().Nats.Stream, mono.JS()))
	domainDispatcher := ddd.NewEventDispatcher[ddd.AggregateEvent]()
	customers := postgres.NewCustomerRepository("customers.customers", mono.DB())

	// setup application
	app := logging.LogApplicationAccess(
		application.New(customers, domainDispatcher),
		mono.Logger(),
	)
	integrationEventHandlers := logging.LogEventHandlerAccess[ddd.AggregateEvent](
		application.NewIntegrationEventHandlers(eventStream),
		"IntegrationEvents", mono.Logger(),
	)
	if err := grpc.RegisterServer(app, mono.RPC()); err != nil {
		return err
	}
	if err := rest.RegisterGateway(ctx, mono.Mux(), mono.Config().Rpc.Address()); err != nil {
		return err
	}
	if err := rest.RegisterSwagger(mono.Mux()); err != nil {
		return err
	}
	handlers.RegisterIntegrationEventHandlers(integrationEventHandlers, domainDispatcher)

	return nil
}
