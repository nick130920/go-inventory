# Architecture

`go-inventory` is split into three focused packages, each with a single
aggregate and a single responsibility. Stock levels are a *projection*
maintained by the movement service, never written to directly by callers.

## Module layout

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

## Why separate stock from movement?

Stock is **derived state**: the on-hand quantity is implied by the sum of
movements. Treating it as a write target leads to drift between the ledger
and the projection. The library encodes the rule:

- Callers may only `Apply(Movement)`.
- The movement service updates the stock projection inside the same
  transaction, atomically.
- Reads of the level go through `stock.Repository.Get`, but writes go
  exclusively through `movement.Service.Apply`.

## Negative stock guard

Outbound movements are rejected with `stock.ErrInsufficient` when the
projected balance would go below zero. This rule lives in the domain,
not in SQL constraints, so it surfaces as a typed error consumers can
translate into a `409 Conflict` or a domain-specific message.

## Transactional boundaries

The library does not own a database transaction. Consumers wrap
`movement.Service.Apply` in their own transaction so that:

- The `movements` table append commits.
- The `stock_levels` upsert commits.
- Any side-effect (notification, kafka publish) fires only on commit.

A typical pgx adapter signature looks like:

```go
func (r *PgRepo) WithTx(ctx context.Context, fn func(stock.Repository, movement.Repository) error) error
```
