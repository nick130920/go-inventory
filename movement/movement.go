// Package movement records the events that change stock levels. Every
// inbound (purchase, return) and outbound (sale, transfer) is persisted as
// an immutable event so the on-hand quantity is auditable.
package movement

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/nick130920/go-inventory/stock"
)

// Direction describes whether the movement increases or decreases stock.
type Direction string

const (
	DirectionInbound  Direction = "inbound"
	DirectionOutbound Direction = "outbound"
)

// Movement is the immutable record of a quantity change.
type Movement struct {
	ID        uuid.UUID
	TenantID  string
	ItemID    uuid.UUID
	Location  string
	Direction Direction
	Quantity  int
	Reason    string
	HappenedAt time.Time
}

// Repository persists movement events.
type Repository interface {
	Append(ctx context.Context, m *Movement) error
	List(ctx context.Context, tenantID string, itemID uuid.UUID, location string, limit int) ([]Movement, error)
}

// ErrInvalidQuantity is returned by Service.Apply when the requested
// quantity is not strictly positive.
var ErrInvalidQuantity = errors.New("movement: quantity must be > 0")

// Service is a write model: it appends a movement and updates the matching
// stock level inside the same transactional boundary the caller provides.
type Service struct {
	stock     stock.Repository
	movements Repository
	now       func() time.Time
}

// NewService wires the dependencies.
func NewService(s stock.Repository, m Repository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{stock: s, movements: m, now: now}
}

// Apply records the movement and updates the affected stock level. Callers
// must wrap the call in a database transaction so that both writes commit
// together.
func (s *Service) Apply(ctx context.Context, m *Movement) error {
	if m.Quantity <= 0 {
		return ErrInvalidQuantity
	}
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	m.HappenedAt = s.now().UTC()

	level, err := s.stock.Get(ctx, m.TenantID, m.ItemID, m.Location)
	if err != nil {
		level = &stock.Level{
			TenantID: m.TenantID,
			ItemID:   m.ItemID,
			Location: m.Location,
		}
	}
	switch m.Direction {
	case DirectionInbound:
		level.Quantity += m.Quantity
	case DirectionOutbound:
		if level.Quantity-m.Quantity < 0 {
			return stock.ErrInsufficient
		}
		level.Quantity -= m.Quantity
	}
	level.UpdatedAt = m.HappenedAt
	if err := s.stock.Upsert(ctx, level); err != nil {
		return err
	}
	return s.movements.Append(ctx, m)
}
