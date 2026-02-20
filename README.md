# Product Catalog Service

Simplified product catalog microservice in Go, backed by Cloud Spanner and exposed over gRPC. Follows DDD / clean architecture with CQRS and a transactional outbox for reliable event publishing.

## Project layout

```
cmd/server/              Service entry point
internal/
  app/product/
    domain/              Pure business logic — no context, no DB imports
    domain/services/     Pricing calculator
    usecases/            Write-side interactors (Golden Mutation Pattern)
    queries/             Read-side handlers
    contracts/           Interfaces for repos & read model
    repo/                Spanner implementations
  models/                DB row types + field constants
  transport/grpc/        Thin gRPC handlers
  services/              DI wiring
  pkg/                   Clock abstraction, typed committer wrapper
commitplan/              Standalone module for atomic mutation plans
proto/product/v1/        Protobuf defs + generated Go code
```

## Running locally

**1. Start the Spanner emulator**

```
docker-compose up -d
```

**2. Run tests**

```
# Unit tests — no emulator needed
go test -v ./internal/app/product/domain/...

# E2E tests — needs the emulator running
SPANNER_EMULATOR_HOST=localhost:9010 go test -v -count=1 ./tests/e2e/...
```

Or just `make test`.

**3. Start the server**

```
make run
```

Listens on `:50051` by default (override with `PORT` env var). The emulator must be running. Schema is applied automatically by the E2E test suite; for manual migration, feed `migrations/001_initial_schema.sql` through the Spanner admin API.

**4. Regenerate proto (optional)**

Generated `.pb.go` files are checked in. To regenerate:

```
make tools   # installs protoc-gen-go + protoc-gen-go-grpc
make proto
```

## Configuration

| Variable | Default | Purpose |
|---|---|---|
| `SPANNER_EMULATOR_HOST` | *(none)* | Set to `localhost:9010` for local dev |
| `SPANNER_PROJECT` | `test-project` | GCP project |
| `SPANNER_INSTANCE` | `test-instance` | Spanner instance |
| `SPANNER_DATABASE` | `test-database` | Spanner database |
| `PORT` | `50051` | gRPC listen port |

## Design notes

**Domain purity.** The `domain` package only imports stdlib (`time`, `math/big`, `errors`, `fmt`). No `context.Context`, no Spanner SDK, no proto types — keeps business rules testable in isolation and enforces that infrastructure stays at the edges.

**Golden Mutation Pattern.** Every write-side flow does the same dance: load the aggregate, call domain methods, ask the repo for `*spanner.Mutation` values (repo never applies them), collect everything into a `commitplan.Plan`, and apply the plan in one shot. The usecase is the only code that calls `committer.Apply`. I considered putting the apply call in the handler layer instead, but keeping it in the usecase means the handler never touches infrastructure directly — which felt cleaner for testing.

**Money with `*big.Rat`.** I store prices as numerator/denominator INT64 columns. `big.Rat` normalises the fraction (so 2000/100 becomes 20/1 internally), but the values are mathematically identical and display correctly with `FloatString(2)`. I thought about storing cents as a single INT64 but the requirements were explicit about `big.Rat`, and rational representation handles arbitrary discount percentages without rounding.

**Outbox.** Domain events are simple intent structs captured during aggregate mutations. The usecase marshals them to JSON and writes them to `outbox_events` in the same commit plan as the business data. No background processor is implemented — that's out of scope — but the events are guaranteed to be written atomically alongside the state change.

**Change tracking.** `ChangeTracker` lets the repo build targeted `UPDATE` mutations with only the dirty fields. Without this, every update would rewrite the entire row, which is wasteful and risky if two concurrent writes touch different columns (though we don't have optimistic locking yet, so that scenario is already lossy).

**Pagination.** Cursor-based using `product_id` as the sort key. The page token is just the last product ID from the previous page. UUIDs give a stable (if meaningless) ordering, which is fine here — a real system would probably sort by `created_at` + `product_id` for deterministic results.

## What I'd do differently with more time

- **Optimistic locking.** Right now concurrent updates can clobber each other. A `version` column with a conditional write (or using Spanner's `ReadWriteTransaction` to do a read-then-write in the same transaction) would fix this.
- **Pagination sort order.** Sorting by UUID is stable but not useful. A `created_at` based cursor would be more practical.
- **Richer outbox payloads.** The event JSON currently has minimal data. In production you'd want the full before/after state, a schema version, and metadata like `user_id` or `correlation_id`.
- **Error wrapping.** I'm using sentinel errors everywhere. In a bigger codebase I'd wrap them with `fmt.Errorf("...: %w", err)` for better stack context.
