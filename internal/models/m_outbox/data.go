package m_outbox

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/spanner"
)

type Data struct {
	EventID     string
	EventType   string
	AggregateID string
	Payload     json.RawMessage
	Status      string
	CreatedAt   time.Time
	ProcessedAt spanner.NullTime
}

type Model struct{}

func New() *Model { return &Model{} }

func (m *Model) InsertMap(values map[string]interface{}) *spanner.Mutation {
	return spanner.InsertMap(Table, values)
}

func (m *Model) ToRow(d *Data) map[string]interface{} {
	return map[string]interface{}{
		EventID:     d.EventID,
		EventType:   d.EventType,
		AggregateID: d.AggregateID,
		Payload:     string(d.Payload),
		Status:      d.Status,
		CreatedAt:   d.CreatedAt,
	}
}
