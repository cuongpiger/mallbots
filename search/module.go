package search

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog"

	"github.com/cuongpiger/mallbots/customers/customerspb"
	"github.com/cuongpiger/mallbots/internal/ddd"
	"github.com/cuongpiger/mallbots/internal/di"
	"github.com/cuongpiger/mallbots/internal/jetstream"
	pg "github.com/cuongpiger/mallbots/internal/postgres"
	"github.com/cuongpiger/mallbots/internal/registry"
	"github.com/cuongpiger/mallbots/internal/system"
	"github.com/cuongpiger/mallbots/internal/tm"
	"github.com/cuongpiger/mallbots/ordering/orderingpb"
	"github.com/cuongpiger/mallbots/search/internal/application"
	"github.com/cuongpiger/mallbots/search/internal/grpc"
	"github.com/cuongpiger/mallbots/search/internal/handlers"
	"github.com/cuongpiger/mallbots/search/internal/logging"
	"github.com/cuongpiger/mallbots/search/internal/postgres"
	"github.com/cuongpiger/mallbots/search/internal/rest"
	"github.com/cuongpiger/mallbots/stores/storespb"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono system.Service) (err error) {
	return Root(ctx, mono)
}

func Root(ctx context.Context, svc system.Service) (err error) {
	container := di.New()
	// setup Driven adapters
	container.AddSingleton("registry", func(c di.Container) (any, error) {
		reg := registry.New()
		if err := orderingpb.Registrations(reg); err != nil {
			return nil, err
		}
		if err := customerspb.Registrations(reg); err != nil {
			return nil, err
		}
		if err := storespb.Registrations(reg); err != nil {
			return nil, err
		}
		return reg, nil
	})
	container.AddSingleton("logger", func(c di.Container) (any, error) {
		return svc.Logger(), nil
	})
	container.AddSingleton("stream", func(c di.Container) (any, error) {
		return jetstream.NewStream(svc.Config().Nats.Stream, svc.JS(), c.Get("logger").(zerolog.Logger)), nil
	})
	container.AddSingleton("db", func(c di.Container) (any, error) {
		return svc.DB(), nil
	})
	container.AddSingleton("conn", func(c di.Container) (any, error) {
		return grpc.Dial(ctx, svc.Config().Rpc.Address())
	})
	container.AddScoped("tx", func(c di.Container) (any, error) {
		db := c.Get("db").(*sql.DB)
		return db.Begin()
	})
	container.AddScoped("inboxMiddleware", func(c di.Container) (any, error) {
		tx := c.Get("tx").(*sql.Tx)
		inboxStore := pg.NewInboxStore("search.inbox", tx)
		return tm.NewInboxHandlerMiddleware(inboxStore), nil
	})
	container.AddScoped("customers", func(c di.Container) (any, error) {
		return postgres.NewCustomerCacheRepository(
			"search.customers_cache",
			c.Get("tx").(*sql.Tx),
			grpc.NewCustomerRepository(c.Get("conn").(*grpc.ClientConn)),
		), nil
	})
	container.AddScoped("stores", func(c di.Container) (any, error) {
		return postgres.NewStoreCacheRepository(
			"search.stores_cache",
			c.Get("tx").(*sql.Tx),
			grpc.NewStoreRepository(c.Get("conn").(*grpc.ClientConn)),
		), nil
	})
	container.AddScoped("products", func(c di.Container) (any, error) {
		return postgres.NewProductCacheRepository(
			"search.products_cache",
			c.Get("tx").(*sql.Tx),
			grpc.NewProductRepository(c.Get("conn").(*grpc.ClientConn)),
		), nil
	})
	container.AddScoped("orders", func(c di.Container) (any, error) {
		return postgres.NewOrderRepository("search.orders", c.Get("tx").(*sql.Tx)), nil
	})

	// setup application
	container.AddScoped("app", func(c di.Container) (any, error) {
		return logging.LogApplicationAccess(
			application.New(
				c.Get("orders").(application.OrderRepository),
			),
			c.Get("logger").(zerolog.Logger),
		), nil
	})
	container.AddScoped("integrationEventHandlers", func(c di.Container) (any, error) {
		return logging.LogEventHandlerAccess[ddd.Event](
			handlers.NewIntegrationEventHandlers(
				c.Get("orders").(application.OrderRepository),
				c.Get("customers").(application.CustomerCacheRepository),
				c.Get("stores").(application.StoreCacheRepository),
				c.Get("products").(application.ProductCacheRepository),
			),
			"IntegrationEvents", c.Get("logger").(zerolog.Logger),
		), nil
	})

	// setup Driver adapters
	if err = grpc.RegisterServerTx(container, svc.RPC()); err != nil {
		return err
	}
	if err = rest.RegisterGateway(ctx, svc.Mux(), svc.Config().Rpc.Address()); err != nil {
		return err
	}
	if err = rest.RegisterSwagger(svc.Mux()); err != nil {
		return err
	}
	if err = handlers.RegisterIntegrationEventHandlersTx(container); err != nil {
		return err
	}

	return nil
}
