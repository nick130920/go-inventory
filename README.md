# go-inventory

Reusable inventory domain library in Go. Models items, stock levels and
movements; leaves persistence and HTTP transport to the consumer.

## Install

```bash
go get github.com/nick130920/go-inventory
```

## Packages

| Path | Responsibility |
|---|---|
| `item` | Item master data, CRUD service, `Repository` interface. |
| `stock` | On-hand quantities per location with `Level` aggregate. |
| `movement` | Append-only movements service that updates stock atomically. |

## Architecture

```
        HTTP / CLI / Worker
               |
               v
       movement.Service       item.Service
        |        |                |
        v        v                v
 stock.Repository  movement.Repository  item.Repository
        \________________|_______________/
                         v
                  Postgres / Mongo
```

## Quick start

```go
import (
    "context"
    "time"

    "github.com/nick130920/go-inventory/item"
    "github.com/nick130920/go-inventory/movement"
    "github.com/nick130920/go-inventory/stock"
)

itemSvc := item.NewService(itemRepo, time.Now)
mvSvc := movement.NewService(stockRepo, movementRepo, time.Now)

err := mvSvc.Apply(ctx, &movement.Movement{
    TenantID:  "hases",
    ItemID:    laptopID,
    Location:  "warehouse-bogota",
    Direction: movement.DirectionInbound,
    Quantity:  10,
    Reason:    "Compra OC-1234",
})
```

## License

Apache-2.0.
