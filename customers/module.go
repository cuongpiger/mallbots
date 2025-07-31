package customers

import (
	"context"

	"github.com/cuongpiger/mallbots/customers/internal/application"
	"github.com/cuongpiger/mallbots/customers/internal/grpc"
	"github.com/cuongpiger/mallbots/customers/internal/logging"
	"github.com/cuongpiger/mallbots/customers/internal/postgres"
	"github.com/cuongpiger/mallbots/customers/internal/rest"
	"github.com/cuongpiger/mallbots/internal/monolith"
)

type Module struct{}

func (m Module) Startup(ctx context.Context, mono monolith.Monolith) error {
	customers := postgres.NewCustomerRepository("customers.customers", mono.DB())

	var app application.App
	app = application.New(customers)
	app = logging.LogApplicationAccess(app, mono.Logger())

	if err := grpc.RegisterServer(app, mono.RPC()); err != nil {
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
