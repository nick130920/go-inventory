# Stock & movements

The `stock` package stores on-hand quantities indexed by `(tenant, item,
location)`. The `movement` package stores the append-only ledger of
quantity changes and updates the stock projection inside the same
transaction.

## Aggregates

```go
type stock.Level struct {
    TenantID  string
    ItemID    uuid.UUID
    Location  string
    Quantity  int
    UpdatedAt time.Time
}

type movement.Movement struct {
    ID         uuid.UUID
    TenantID   string
    ItemID     uuid.UUID
    Location   string
    Direction  Direction       // inbound | outbound
    Quantity   int
    Reason     string
    HappenedAt time.Time
}
```

## Apply a movement

```go
import (
    "context"
    "time"

    "github.com/nick130920/go-inventory/movement"
    "github.com/nick130920/go-inventory/stock"
)

mvSvc := movement.NewService(stockRepo, movementRepo, time.Now)

err := mvSvc.Apply(ctx, &movement.Movement{
    TenantID:  "hases",
    ItemID:    cascoID,
    Location:  "warehouse-bogota",
    Direction: movement.DirectionInbound,
    Quantity:  10,
    Reason:    "Compra OC-1234",
})
```

`Apply` does three things in order:

1. Validates that `Quantity > 0`.
2. Loads (or initializes) the matching `stock.Level`.
3. Updates the level (`+` for inbound, `-` for outbound) and appends the
   movement record.

If the projected balance after an outbound movement would be negative,
`stock.ErrInsufficient` is returned and **nothing is written**.

## Wrapping in a transaction

The library does not start a transaction; the consumer must:

```go
err := db.WithTx(ctx, func(stockRepo stock.Repository, mvRepo movement.Repository) error {
    svc := movement.NewService(stockRepo, mvRepo, time.Now)
    return svc.Apply(ctx, m)
})
```

Use cases like batch counts or transfers that touch multiple movements in
a single business operation simply call `Apply` repeatedly inside the
same transaction.

## Reading levels

```go
level, err := stockRepo.Get(ctx, "hases", cascoID, "warehouse-bogota")
if err != nil {
    return err
}
fmt.Printf("On hand: %d\n", level.Quantity)
```

Reads bypass the movement service deliberately: they are the only path
that does not need transactional coupling.
