package handlers

import (
	"github.com/cuongpiger/mallbots/internal/ddd"
	"github.com/cuongpiger/mallbots/payments/internal/models"
)

func RegisterIntegrationEventHandlers(eventHandlers ddd.EventHandler[ddd.Event], domainSubscriber ddd.EventSubscriber[ddd.Event]) {
	domainSubscriber.Subscribe(eventHandlers,
		models.InvoicePaidEvent,
	)
}
