package application

import (
	"context"

	"github.com/cuongpiger/mallbots/search/internal/models"
)

type CustomerRepository interface {
	Find(ctx context.Context, customerID string) (*models.Customer, error)
}
