package repo

import (
	"cloud.google.com/go/spanner"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/models/m_outbox"
)

var _ contracts.OutboxRepository = (*OutboxRepo)(nil)

type OutboxRepo struct {
	model *m_outbox.Model
}

func NewOutboxRepo() *OutboxRepo {
	return &OutboxRepo{model: m_outbox.New()}
}

func (r *OutboxRepo) InsertMut(event contracts.OutboxEvent) *spanner.Mutation {
	data := &m_outbox.Data{
		EventID:     event.ID,
		EventType:   event.EventType,
		AggregateID: event.AggregateID,
		Payload:     event.Payload,
		Status:      "pending",
		CreatedAt:   event.CreatedAt,
	}
	return r.model.InsertMap(r.model.ToRow(data))
}
