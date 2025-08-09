package logging

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/cuongpiger/mallbots/notifications/internal/application"
)

var _ application.App = (*Application)(nil)

type Application struct {
	application.App
	logger zerolog.Logger
}

func LogApplicationAccess(application application.App, logger zerolog.Logger) Application {
	return Application{
		App:    application,
		logger: logger,
	}
}

func (a Application) NotifyOrderCreated(ctx context.Context, notify application.OrderCreated) (err error) {
	a.logger.Info().Msg("--> Notifications.NotifyOrderCreated")
	defer func() { a.logger.Info().Err(err).Msg("<-- Notifications.NotifyOrderCreated") }()
	return a.App.NotifyOrderCreated(ctx, notify)
}

func (a Application) NotifyOrderCanceled(ctx context.Context, notify application.OrderCanceled) (err error) {
	a.logger.Info().Msg("--> Notifications.NotifyOrderCanceled")
	defer func() { a.logger.Info().Err(err).Msg("<-- Notifications.NotifyOrderCanceled") }()
	return a.App.NotifyOrderCanceled(ctx, notify)
}

func (a Application) NotifyOrderReady(ctx context.Context, notify application.OrderReady) (err error) {
	a.logger.Info().Msg("--> Notifications.NotifyOrderReady")
	defer func() { a.logger.Info().Err(err).Msg("<-- Notifications.NotifyOrderReady") }()
	return a.App.NotifyOrderReady(ctx, notify)
}
