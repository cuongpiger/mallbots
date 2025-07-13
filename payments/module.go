package payments

import (
	"context"

	"github.com/cuongpiger/mallbots/internal/monolith"
	"github.com/cuongpiger/mallbots/payments/internal/application"
	"github.com/cuongpiger/mallbots/payments/internal/grpc"
	"github.com/cuongpiger/mallbots/payments/internal/logging"
	"github.com/cuongpiger/mallbots/payments/internal/postgres"
	"github.com/cuongpiger/mallbots/payments/internal/rest"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono monolith.Monolith) error {
	// setup Driven adapters
	invoices := postgres.NewInvoiceRepository("payments.invoices", mono.DB())
	payments := postgres.NewPaymentRepository("payments.payments", mono.DB())
	conn, err := grpc.Dial(ctx, mono.Config().Rpc.Address())
	if err != nil {
		return err
	}
	orders := grpc.NewOrderRepository(conn)

	// setup application
	var app application.App
	app = application.New(invoices, payments, orders)
	app = logging.LogApplicationAccess(app, mono.Logger())

	// setup Driver adapters
	if err := grpc.RegisterServer(ctx, app, mono.RPC()); err != nil {
		return err
	}
	if err := rest.RegisterGateway(ctx, mono.Mux(), mono.Config().Rpc.Address()); err != nil {
		return err
	}
	if err := rest.RegisterSwagger(mono.Mux()); err != nil {
		return err
	}

	return nil
}
