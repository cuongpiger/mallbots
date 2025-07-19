package handlers

import (
	"github.com/cuongpiger/mallbots/baskets/internal/application"
	"github.com/cuongpiger/mallbots/baskets/internal/domain"
	"github.com/cuongpiger/mallbots/internal/ddd"
)

func RegisterOrderHandlers(orderHandlers application.DomainEventHandlers, domainSubscriber ddd.EventSubscriber) {
	domainSubscriber.Subscribe(domain.BasketCheckedOut{}, orderHandlers.OnBasketCheckedOut)
}
