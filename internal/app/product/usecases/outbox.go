package usecases

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
)

// EnrichEvent turns a domain event into an OutboxEvent with a JSON payload.
func EnrichEvent(aggregateID string, event domain.DomainEvent) contracts.OutboxEvent {
	raw, _ := json.Marshal(buildPayload(event))
	return contracts.OutboxEvent{
		ID:          uuid.NewString(),
		EventType:   event.EventType(),
		AggregateID: aggregateID,
		Payload:     raw,
		CreatedAt:   event.OccurredAt(),
	}
}

func buildPayload(event domain.DomainEvent) interface{} {
	switch e := event.(type) {
	case *domain.ProductCreatedEvent:
		return map[string]interface{}{
			"product_id": e.ProductID,
			"name":       e.Name,
			"category":   e.Category,
		}
	case *domain.ProductUpdatedEvent:
		return map[string]interface{}{"product_id": e.ProductID}
	case *domain.ProductActivatedEvent:
		return map[string]interface{}{"product_id": e.ProductID}
	case *domain.ProductDeactivatedEvent:
		return map[string]interface{}{"product_id": e.ProductID}
	case *domain.DiscountAppliedEvent:
		return map[string]interface{}{
			"product_id": e.ProductID,
			"percentage": e.Percentage.FloatString(2),
			"start_date": e.StartDate,
			"end_date":   e.EndDate,
		}
	case *domain.DiscountRemovedEvent:
		return map[string]interface{}{"product_id": e.ProductID}
	default:
		return map[string]interface{}{}
	}
}
