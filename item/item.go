// Package item models the master data of an inventory: SKU catalog, units
// of measure and prices. Stock levels live in package stock.
package item

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Item is the canonical representation of something the company tracks.
type Item struct {
	ID        uuid.UUID
	TenantID  string
	SKU       string
	Name      string
	UnitOfMeasure string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Repository persists items. Implemented by adapters in the consumer.
type Repository interface {
	Create(ctx context.Context, i *Item) error
	Update(ctx context.Context, i *Item) error
	GetBySKU(ctx context.Context, tenantID, sku string) (*Item, error)
	List(ctx context.Context, tenantID string, q ListQuery) ([]Item, error)
}

// ListQuery captures the filters supported when listing items.
type ListQuery struct {
	OnlyActive bool
	Search     string
	Limit      int
	Offset     int
}

// ErrNotFound is returned by repositories when the item does not exist.
var ErrNotFound = errors.New("item: not found")

// ErrDuplicateSKU is returned when an SKU already exists for the tenant.
var ErrDuplicateSKU = errors.New("item: duplicate SKU")

// Service provides the business operations for items.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService wires the dependencies.
func NewService(r Repository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{repo: r, now: now}
}

// Register creates a new active item. SKU uniqueness is enforced by the
// repository; this layer normalizes the timestamps and the active flag.
func (s *Service) Register(ctx context.Context, i *Item) error {
	now := s.now().UTC()
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	i.Active = true
	i.CreatedAt = now
	i.UpdatedAt = now
	return s.repo.Create(ctx, i)
}

// Deactivate marks an item as inactive without deleting it, preserving
// historical movements that still reference the SKU.
func (s *Service) Deactivate(ctx context.Context, tenantID, sku string) error {
	i, err := s.repo.GetBySKU(ctx, tenantID, sku)
	if err != nil {
		return err
	}
	i.Active = false
	i.UpdatedAt = s.now().UTC()
	return s.repo.Update(ctx, i)
}
