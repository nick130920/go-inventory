// Package stock keeps per-location quantities for items. Persistence
// adapters and read models stay outside the domain package.
package stock

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Level is the on-hand quantity of an item at a specific location.
type Level struct {
	TenantID  string
	ItemID    uuid.UUID
	Location  string
	Quantity  int
	UpdatedAt time.Time
}

// Repository persists stock levels. Implementations must guarantee that the
// composite key (TenantID, ItemID, Location) is unique.
type Repository interface {
	Get(ctx context.Context, tenantID string, itemID uuid.UUID, location string) (*Level, error)
	Upsert(ctx context.Context, l *Level) error
}

// ErrInsufficient is returned when an outbound movement would drive the
// quantity below zero.
var ErrInsufficient = errors.New("stock: insufficient quantity")
