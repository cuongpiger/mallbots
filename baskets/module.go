package baskets

import (
	"context"

	"github.com/cuongpiger/mallbots/baskets/basketspb"
	"github.com/cuongpiger/mallbots/baskets/internal/application"
	"github.com/cuongpiger/mallbots/baskets/internal/domain"
	"github.com/cuongpiger/mallbots/baskets/internal/grpc"
	"github.com/cuongpiger/mallbots/baskets/internal/handlers"
	"github.com/cuongpiger/mallbots/baskets/internal/logging"
	"github.com/cuongpiger/mallbots/baskets/internal/postgres"
	"github.com/cuongpiger/mallbots/baskets/internal/rest"
	"github.com/cuongpiger/mallbots/internal/am"
	"github.com/cuongpiger/mallbots/internal/ddd"
	"github.com/cuongpiger/mallbots/internal/es"
	"github.com/cuongpiger/mallbots/internal/jetstream"
	"github.com/cuongpiger/mallbots/internal/monolith"
	pg "github.com/cuongpiger/mallbots/internal/postgres"
	"github.com/cuongpiger/mallbots/internal/registry"
	"github.com/cuongpiger/mallbots/internal/registry/serdes"
	"github.com/cuongpiger/mallbots/stores/storespb"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = registrations(reg); err != nil {
		return err
	}
	if err = basketspb.Registrations(reg); err != nil {
		return err
	}
	if err = storespb.Registrations(reg); err != nil {
		return err
	}
	eventStream := am.NewEventStream(reg, jetstream.NewStream(mono.Config().Nats.Stream, mono.JS(), mono.Logger()))
	domainDispatcher := ddd.NewEventDispatcher[ddd.Event]()
	aggregateStore := es.AggregateStoreWithMiddleware(
		pg.NewEventStore("baskets.events", mono.DB(), reg),
		pg.NewSnapshotStore("baskets.snapshots", mono.DB(), reg),
	)
	baskets := es.NewAggregateRepository[*domain.Basket](domain.BasketAggregate, reg, aggregateStore)
	conn, err := grpc.Dial(ctx, mono.Config().Rpc.Address())
	if err != nil {
		return err
	}
	stores := postgres.NewStoreCacheRepository("baskets.stores_cache", mono.DB(), grpc.NewStoreRepository(conn))
	products := postgres.NewProductCacheRepository("baskets.products_cache", mono.DB(), grpc.NewProductRepository(conn))

	// setup application
	app := logging.LogApplicationAccess(
		application.New(baskets, stores, products, domainDispatcher),
		mono.Logger(),
	)
	domainEventHandlers := logging.LogEventHandlerAccess[ddd.Event](
		handlers.NewDomainEventHandlers(eventStream),
		"DomainEvents", mono.Logger(),
	)
	integrationEventHandlers := logging.LogEventHandlerAccess[ddd.Event](
		handlers.NewIntegrationEventHandlers(stores, products),
		"IntegrationEvents", mono.Logger(),
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

	return
}

func registrations(reg registry.Registry) error {
	serde := serdes.NewJsonSerde(reg)

	// Basket
	if err := serde.Register(domain.Basket{}, func(v interface{}) error {
		basket := v.(*domain.Basket)
		basket.Items = make(map[string]domain.Item)
		return nil
	}); err != nil {
		return err
	}
	// basket events
	if err := serde.Register(domain.BasketStarted{}); err != nil {
		return err
	}
	if err := serde.Register(domain.BasketCanceled{}); err != nil {
		return err
	}
	if err := serde.Register(domain.BasketCheckedOut{}); err != nil {
		return err
	}
	if err := serde.Register(domain.BasketItemAdded{}); err != nil {
		return err
	}
	if err := serde.Register(domain.BasketItemRemoved{}); err != nil {
		return err
	}
	// basket snapshots
	if err := serde.RegisterKey(domain.BasketV1{}.SnapshotName(), domain.BasketV1{}); err != nil {
		return err
	}

	return nil
}
