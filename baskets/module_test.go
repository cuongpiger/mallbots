package baskets

import (
	"testing"
	
	"github.com/cuongpiger/mallbots/internal/registry"
)

func TestRegistration(t *testing.T) {
	registry := registry.New()
	if err := registrations(registry); err != nil {
		t.Fatalf("failed to register: %v", err)
	}
}