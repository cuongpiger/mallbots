package handlers

import (
	"context"

	"github.com/cuongpiger/mallbots/customers/customerspb"
	"github.com/cuongpiger/mallbots/internal/am"
	"github.com/cuongpiger/mallbots/internal/ddd"
)

func RegisterCustomerHandlers(customerHandlers ddd.EventHandler[ddd.Event], stream am.EventSubscriber) error {
	evtMsgHandler := am.MessageHandlerFunc[am.EventMessage](func(ctx context.Context, eventMsg am.EventMessage) error {
		return customerHandlers.HandleEvent(ctx, eventMsg)
	})

	return stream.Subscribe(customerspb.CustomerAggregateChannel, evtMsgHandler, am.MessageFilter{
		customerspb.CustomerRegisteredEvent,
		customerspb.CustomerSmsChangedEvent,
	}, am.GroupName("notification-customers"))
}
