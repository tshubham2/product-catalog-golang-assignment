package contracts

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/spanner"

	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
)

type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*domain.Product, error)
	InsertMut(p *domain.Product) *spanner.Mutation
	UpdateMut(p *domain.Product) *spanner.Mutation
}

type OutboxRepository interface {
	InsertMut(event OutboxEvent) *spanner.Mutation
}

// OutboxEvent is the enriched form of a domain event, ready for persistence.
type OutboxEvent struct {
	ID          string
	EventType   string
	AggregateID string
	Payload     json.RawMessage
	CreatedAt   time.Time
}
