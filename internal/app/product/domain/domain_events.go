package domain

import (
	"math/big"
	"time"
)

// DomainEvent represents an intent captured by the aggregate. Events are
// collected during business operations and later persisted to the outbox.
type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type baseEvent struct {
	occurredAt time.Time
}

func (e baseEvent) OccurredAt() time.Time { return e.occurredAt }

type ProductCreatedEvent struct {
	baseEvent
	ProductID string
	Name      string
	Category  string
}

func (e *ProductCreatedEvent) EventType() string { return "product.created" }

type ProductUpdatedEvent struct {
	baseEvent
	ProductID string
}

func (e *ProductUpdatedEvent) EventType() string { return "product.updated" }

type ProductActivatedEvent struct {
	baseEvent
	ProductID string
}

func (e *ProductActivatedEvent) EventType() string { return "product.activated" }

type ProductDeactivatedEvent struct {
	baseEvent
	ProductID string
}

func (e *ProductDeactivatedEvent) EventType() string { return "product.deactivated" }

type DiscountAppliedEvent struct {
	baseEvent
	ProductID  string
	Percentage *big.Rat
	StartDate  time.Time
	EndDate    time.Time
}

func (e *DiscountAppliedEvent) EventType() string { return "discount.applied" }

type DiscountRemovedEvent struct {
	baseEvent
	ProductID string
}

func (e *DiscountRemovedEvent) EventType() string { return "discount.removed" }
