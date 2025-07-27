package notifications

import (
	"context"

	"github.com/cuongpiger/mallbots/customers/customerspb"
	"github.com/cuongpiger/mallbots/internal/am"
	"github.com/cuongpiger/mallbots/internal/amotel"
	"github.com/cuongpiger/mallbots/internal/amprom"
	"github.com/cuongpiger/mallbots/internal/jetstream"
	pg "github.com/cuongpiger/mallbots/internal/postgres"
	"github.com/cuongpiger/mallbots/internal/postgresotel"
	"github.com/cuongpiger/mallbots/internal/registry"
	"github.com/cuongpiger/mallbots/internal/system"
	"github.com/cuongpiger/mallbots/internal/tm"
	"github.com/cuongpiger/mallbots/notifications/internal/application"
	"github.com/cuongpiger/mallbots/notifications/internal/constants"
	"github.com/cuongpiger/mallbots/notifications/internal/grpc"
	"github.com/cuongpiger/mallbots/notifications/internal/handlers"
	"github.com/cuongpiger/mallbots/notifications/internal/postgres"
	"github.com/cuongpiger/mallbots/ordering/orderingpb"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono system.Service) (err error) {
	return Root(ctx, mono)
}

func Root(ctx context.Context, svc system.Service) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = customerspb.Registrations(reg); err != nil {
		return err
	}
	if err = orderingpb.Registrations(reg); err != nil {
		return err
	}
	inboxStore := pg.NewInboxStore(constants.InboxTableName, svc.DB())
	messageSubscriber := am.NewMessageSubscriber(
		jetstream.NewStream(svc.Config().Nats.Stream, svc.JS(), svc.Logger()),
		amotel.OtelMessageContextExtractor(),
		amprom.ReceivedMessagesCounter(constants.ServiceName),
	)
	customers := postgres.NewCustomerCacheRepository(
		constants.CustomersCacheTableName,
		postgresotel.Trace(svc.DB()),
		grpc.NewCustomerRepository(svc.Config().Rpc.Service(constants.CustomersServiceName)),
	)

	// setup application
	app := application.New(customers)
	integrationEventHandlers := handlers.NewIntegrationEventHandlers(
		reg, app, customers,
		tm.InboxHandler(inboxStore),
	)

	// setup Driver adapters
	if err := grpc.RegisterServer(ctx, app, svc.RPC()); err != nil {
		return err
	}
	if err = handlers.RegisterIntegrationEventHandlers(messageSubscriber, integrationEventHandlers); err != nil {
		return err
	}

	return nil
}
