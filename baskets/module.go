package baskets

import (
	"context"

	"github.com/cuongpiger/mallbots/baskets/internal/application"
	"github.com/cuongpiger/mallbots/baskets/internal/grpc"
	"github.com/cuongpiger/mallbots/baskets/internal/logging"
	"github.com/cuongpiger/mallbots/baskets/internal/postgres"
	"github.com/cuongpiger/mallbots/baskets/internal/rest"
	"github.com/cuongpiger/mallbots/internal/monolith"
)

type Module struct{}

func (m *Module) Startup(ctx context.Context, mono monolith.Monolith) (err error) {
	// setup Driven adapters
	baskets := postgres.NewBasketRepository("baskets.baskets", mono.DB())
	conn, err := grpc.Dial(ctx, mono.Config().Rpc.Address())
	if err != nil {
		return err
	}
	stores := grpc.NewStoreRepository(conn)
	products := grpc.NewProductRepository(conn)
	orders := grpc.NewOrderRepository(conn)

	// setup application
	var app application.App
	app = application.New(baskets, stores, products, orders)
	app = logging.LogApplicationAccess(app, mono.Logger())

	// setup Driver adapters
	if err := grpc.RegisterServer(app, mono.RPC()); err != nil {
		return err
	}
	if err := rest.RegisterGateway(ctx, mono.Mux(), mono.Config().Rpc.Address()); err != nil {
		return err
	}
	if err := rest.RegisterSwagger(mono.Mux()); err != nil {
		return err
	}

	return
}