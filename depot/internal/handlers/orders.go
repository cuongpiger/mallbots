package handlers

import (
	"github.com/cuongpiger/mallbots/depot/internal/application"
	"github.com/cuongpiger/mallbots/depot/internal/domain"
	"github.com/cuongpiger/mallbots/internal/ddd"
)

func RegisterOrderHandlers(orderHandlers application.DomainEventHandlers, domainSubscriber ddd.EventSubscriber) {
	domainSubscriber.Subscribe(domain.ShoppingListCompleted{}, orderHandlers.OnShoppingListCompleted)
}
