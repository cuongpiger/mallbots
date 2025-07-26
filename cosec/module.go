package cosec

import (
	"context"

	"github.com/cuongpiger/mallbots/cosec/internal"
	"github.com/cuongpiger/mallbots/cosec/internal/handlers"
	"github.com/cuongpiger/mallbots/cosec/internal/logging"
	"github.com/cuongpiger/mallbots/cosec/internal/models"
	"github.com/cuongpiger/mallbots/customers/customerspb"
	"github.com/cuongpiger/mallbots/depot/depotpb"
	"github.com/cuongpiger/mallbots/internal/am"
	"github.com/cuongpiger/mallbots/internal/ddd"
	"github.com/cuongpiger/mallbots/internal/jetstream"
	"github.com/cuongpiger/mallbots/internal/monolith"
	pg "github.com/cuongpiger/mallbots/internal/postgres"
	"github.com/cuongpiger/mallbots/internal/registry"
	"github.com/cuongpiger/mallbots/internal/registry/serdes"
	"github.com/cuongpiger/mallbots/internal/sec"
	"github.com/cuongpiger/mallbots/ordering/orderingpb"
	"github.com/cuongpiger/mallbots/payments/paymentspb"
)

type Module struct{}

func (Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	reg := registry.New()
	if err = registrations(reg); err != nil {
		return err
	}
	if err = orderingpb.Registrations(reg); err != nil {
		return err
	}
	if err = customerspb.Registrations(reg); err != nil {
		return err
	}
	if err = depotpb.Registrations(reg); err != nil {
		return err
	}
	if err = paymentspb.Registrations(reg); err != nil {
		return err
	}
	stream := jetstream.NewStream(mono.Config().Nats.Stream, mono.JS(), mono.Logger())
	eventStream := am.NewEventStream(reg, stream)
	commandStream := am.NewCommandStream(reg, stream)
	replyStream := am.NewReplyStream(reg, stream)
	sagaStore := pg.NewSagaStore("cosec.sagas", mono.DB(), reg)
	sagaRepo := sec.NewSagaRepository[*models.CreateOrderData](reg, sagaStore)

	// setup application
	orchestrator := logging.LogReplyHandlerAccess[*models.CreateOrderData](
		sec.NewOrchestrator[*models.CreateOrderData](internal.NewCreateOrderSaga(), sagaRepo, commandStream),
		"CreateOrderSaga", mono.Logger(),
	)
	integrationEventHandlers := logging.LogEventHandlerAccess[ddd.Event](
		handlers.NewIntegrationEventHandlers(orchestrator),
		"IntegrationEvents", mono.Logger(),
	)

	// setup Driver adapters
	if err = handlers.RegisterIntegrationEventHandlers(eventStream, integrationEventHandlers); err != nil {
		return err
	}
	if err = handlers.RegisterReplyHandlers(replyStream, orchestrator); err != nil {
		return err
	}

	return
}

func registrations(reg registry.Registry) (err error) {
	serde := serdes.NewJsonSerde(reg)

	// Saga data
	if err = serde.RegisterKey(internal.CreateOrderSagaName, models.CreateOrderData{}); err != nil {
		return err
	}

	return nil
}
