# Item service

The `item` package owns the SKU catalog: master data that never moves
quantitatively but does change qualitatively (rename, deactivate).

## Aggregate

```go
type Item struct {
    ID            uuid.UUID
    TenantID      string
    SKU           string
    Name          string
    UnitOfMeasure string
    Active        bool
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

## Service

```go
import (
    "time"

    "github.com/nick130920/go-inventory/item"
)

svc := item.NewService(itemRepo, time.Now)

err := svc.Register(ctx, &item.Item{
    TenantID:      "hases",
    SKU:           "EPP-CASCO-N1",
    Name:          "Casco de seguridad clase G",
    UnitOfMeasure: "unit",
})
```

`Register` normalizes the timestamps, generates the ID when missing, and
forces `Active = true`. SKU uniqueness per tenant is delegated to the
repository.

## Soft deletion

Items are never deleted, only deactivated:

```go
if err := svc.Deactivate(ctx, "hases", "EPP-CASCO-N1"); err != nil {
    return err
}
```

This preserves historical movements that still reference the SKU and
keeps reporting consistent.

## Repository contract

```go
type Repository interface {
    Create(ctx context.Context, i *Item) error
    Update(ctx context.Context, i *Item) error
    GetBySKU(ctx context.Context, tenantID, sku string) (*Item, error)
    List(ctx context.Context, tenantID string, q ListQuery) ([]Item, error)
}
```
