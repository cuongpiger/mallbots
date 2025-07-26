package ordering

import (
	"context"

	"github.com/cuongpiger/mallbots/baskets/basketspb"
	"github.com/cuongpiger/mallbots/depot/depotpb"
	"github.com/cuongpiger/mallbots/internal/am"
	"github.com/cuongpiger/mallbots/internal/ddd"
	"github.com/cuongpiger/mallbots/internal/es"
	"github.com/cuongpiger/mallbots/internal/jetstream"
	"github.com/cuongpiger/mallbots/internal/monolith"
	pg "github.com/cuongpiger/mallbots/internal/postgres"
	"github.com/cuongpiger/mallbots/internal/registry"
	"github.com/cuongpiger/mallbots/internal/registry/serdes"
	"github.com/cuongpiger/mallbots/ordering/internal/application"
	"github.com/cuongpiger/mallbots/ordering/internal/domain"
	"github.com/cuongpiger/mallbots/ordering/internal/grpc"
	"github.com/cuongpiger/mallbots/ordering/internal/handlers"
	"github.com/cuongpiger/mallbots/ordering/internal/logging"
	"github.com/cuongpiger/mallbots/ordering/internal/rest"
	"github.com/cuongpiger/mallbots/ordering/orderingpb"
)

type Module struct{}

func (Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = registrations(reg); err != nil {
		return err
	}
	if err = basketspb.Registrations(reg); err != nil {
		return err
	}
	if err = orderingpb.Registrations(reg); err != nil {
		return err
	}
	if err = depotpb.Registrations(reg); err != nil {
		return err
	}
	domainDispatcher := ddd.NewEventDispatcher[ddd.Event]()
	stream := jetstream.NewStream(mono.Config().Nats.Stream, mono.JS(), mono.Logger())
	eventStream := am.NewEventStream(reg, stream)
	commandStream := am.NewCommandStream(reg, stream)
	aggregateStore := es.AggregateStoreWithMiddleware(
		pg.NewEventStore("ordering.events", mono.DB(), reg),
		pg.NewSnapshotStore("ordering.snapshots", mono.DB(), reg),
	)
	orders := es.NewAggregateRepository[*domain.Order](domain.OrderAggregate, reg, aggregateStore)
	conn, err := grpc.Dial(ctx, mono.Config().Rpc.Address())
	if err != nil {
		return err
	}
	shopping := grpc.NewShoppingListRepository(conn)

	// setup application
	app := logging.LogApplicationAccess(
		application.New(orders, shopping, domainDispatcher),
		mono.Logger(),
	)
	domainEventHandlers := logging.LogEventHandlerAccess[ddd.Event](
		handlers.NewDomainEventHandlers(eventStream),
		"DomainEvents", mono.Logger(),
	)
	integrationEventHandlers := logging.LogEventHandlerAccess[ddd.Event](
		handlers.NewIntegrationEventHandlers(app),
		"IntegrationEvents", mono.Logger(),
	)
	commandHandlers := logging.LogCommandHandlerAccess[ddd.Command](
		handlers.NewCommandHandlers(app),
		"Commands", mono.Logger(),
	)

	// setup Driver adapters
	if err = grpc.RegisterServer(app, mono.RPC()); err != nil {
		return err
	}
	if err = rest.RegisterGateway(ctx, mono.Mux(), mono.Config().Rpc.Address()); err != nil {
		return err
	}
	if err = rest.RegisterSwagger(mono.Mux()); err != nil {
		return err
	}
	handlers.RegisterDomainEventHandlers(domainDispatcher, domainEventHandlers)
	if err = handlers.RegisterIntegrationEventHandlers(eventStream, integrationEventHandlers); err != nil {
		return err
	}
	if err = handlers.RegisterCommandHandlers(commandStream, commandHandlers); err != nil {
		return err
	}

	return nil
}

func registrations(reg registry.Registry) (err error) {
	serde := serdes.NewJsonSerde(reg)

	// Order
	if err = serde.Register(domain.Order{}, func(v any) error {
		order := v.(*domain.Order)
		order.Aggregate = es.NewAggregate("", domain.OrderAggregate)
		return nil
	}); err != nil {
		return err
	}
	// order events
	if err = serde.Register(domain.OrderCreated{}); err != nil {
		return err
	}
	if err = serde.Register(domain.OrderRejected{}); err != nil {
		return err
	}
	if err = serde.Register(domain.OrderApproved{}); err != nil {
		return err
	}
	if err = serde.Register(domain.OrderCanceled{}); err != nil {
		return err
	}
	if err = serde.Register(domain.OrderReadied{}); err != nil {
		return err
	}
	if err = serde.Register(domain.OrderCompleted{}); err != nil {
		return err
	}
	// order snapshots
	if err = serde.RegisterKey(domain.OrderV1{}.SnapshotName(), domain.OrderV1{}); err != nil {
		return err
	}

	return nil
}
