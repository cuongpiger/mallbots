package handlers

import (
	"github.com/cuongpiger/mallbots/internal/ddd"
	"github.com/cuongpiger/mallbots/ordering/internal/domain"
)

func RegisterInvoiceHandlers(invoiceHandlers ddd.EventHandler[ddd.AggregateEvent], domainSubscriber ddd.EventSubscriber[ddd.AggregateEvent]) {
	domainSubscriber.Subscribe(domain.OrderReadiedEvent, invoiceHandlers)
}
