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
| [`item`](usage-item.md) | Item master data, CRUD service, `Repository` interface. |
| [`stock`](usage-movements.md) | On-hand quantities per location with `Level` aggregate. |
| [`movement`](usage-movements.md) | Append-only movements service that updates stock atomically. |

## Why this library exists

Endowment, warehouse and supply scenarios all share the same primitives:
an immutable item catalog, on-hand quantities indexed by location, and a
ledger of inbound/outbound movements. By shipping these as a library:

- Every consuming service inherits the **same audit guarantees** (every
  quantity change is an event with a reason and a timestamp).
- New use cases (returns, cycle counts, EPP delivery to workers) only need
  a thin command/handler on top.
- Companies adopting it can plug their own SKU schema and location codes
  via embedding, without forking the library.
