package handlers

import (
	"github.com/cuongpiger/mallbots/depot/internal/domain"
	"github.com/cuongpiger/mallbots/internal/ddd"
)

func RegisterOrderHandlers(orderHandlers ddd.EventHandler[ddd.AggregateEvent], domainSubscriber ddd.EventSubscriber[ddd.AggregateEvent]) {
	domainSubscriber.Subscribe(orderHandlers, domain.ShoppingListCompletedEvent)
}
