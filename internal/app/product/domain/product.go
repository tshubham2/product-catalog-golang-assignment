package domain

import (
	"time"
)

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

// Product is the aggregate root for the catalog bounded context.
type Product struct {
	id          string
	name        string
	description string
	category    string
	basePrice   *Money
	discount    *Discount
	status      ProductStatus
	createdAt   time.Time
	updatedAt   time.Time
	archivedAt  *time.Time

	changes *ChangeTracker
	events  []DomainEvent
}

// NewProduct creates a product in active status and records a created event.
func NewProduct(id, name, description, category string, basePrice *Money, now time.Time) (*Product, error) {
	if name == "" {
		return nil, ErrProductNameRequired
	}
	if category == "" {
		return nil, ErrCategoryRequired
	}

	p := &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		status:      ProductStatusActive,
		createdAt:   now,
		updatedAt:   now,
		changes:     NewChangeTracker(),
	}

	p.events = append(p.events, &ProductCreatedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: id,
		Name:      name,
		Category:  category,
	})

	return p, nil
}

// Reconstitute rebuilds a product from persisted data without firing events
// or marking anything dirty. Used only by the repository.
func Reconstitute(
	id, name, description, category string,
	basePrice *Money,
	discount *Discount,
	status ProductStatus,
	createdAt, updatedAt time.Time,
	archivedAt *time.Time,
) *Product {
	return &Product{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		basePrice:   basePrice,
		discount:    discount,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		archivedAt:  archivedAt,
		changes:     NewChangeTracker(),
	}
}

func (p *Product) ID() string              { return p.id }
func (p *Product) Name() string            { return p.name }
func (p *Product) Description() string     { return p.description }
func (p *Product) Category() string        { return p.category }
func (p *Product) BasePrice() *Money       { return p.basePrice }
func (p *Product) Discount() *Discount     { return p.discount }
func (p *Product) Status() ProductStatus   { return p.status }
func (p *Product) CreatedAt() time.Time    { return p.createdAt }
func (p *Product) UpdatedAt() time.Time    { return p.updatedAt }
func (p *Product) ArchivedAt() *time.Time  { return p.archivedAt }
func (p *Product) Changes() *ChangeTracker { return p.changes }

func (p *Product) DomainEvents() []DomainEvent { return p.events }
func (p *Product) ClearEvents()                { p.events = nil }

func (p *Product) UpdateDetails(name, description, category string, now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}

	changed := false

	if name != p.name {
		if name == "" {
			return ErrProductNameRequired
		}
		p.name = name
		p.changes.MarkDirty(FieldName)
		changed = true
	}
	if description != p.description {
		p.description = description
		p.changes.MarkDirty(FieldDescription)
		changed = true
	}
	if category != p.category {
		if category == "" {
			return ErrCategoryRequired
		}
		p.category = category
		p.changes.MarkDirty(FieldCategory)
		changed = true
	}

	if changed {
		p.updatedAt = now
		p.events = append(p.events, &ProductUpdatedEvent{
			baseEvent: baseEvent{occurredAt: now},
			ProductID: p.id,
		})
	}
	return nil
}

func (p *Product) Activate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.status == ProductStatusActive {
		return ErrProductAlreadyActive
	}

	p.status = ProductStatusActive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)

	p.events = append(p.events, &ProductActivatedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: p.id,
	})
	return nil
}

func (p *Product) Deactivate(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}
	if p.status != ProductStatusActive {
		return ErrProductAlreadyInactive
	}

	p.status = ProductStatusInactive
	p.updatedAt = now
	p.changes.MarkDirty(FieldStatus)

	p.events = append(p.events, &ProductDeactivatedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: p.id,
	})
	return nil
}

func (p *Product) Archive(now time.Time) error {
	if p.status == ProductStatusArchived {
		return ErrProductArchived
	}

	p.status = ProductStatusArchived
	p.updatedAt = now
	archived := now
	p.archivedAt = &archived
	p.changes.MarkDirty(FieldStatus)
	return nil
}

func (p *Product) ApplyDiscount(discount *Discount, now time.Time) error {
	if p.status != ProductStatusActive {
		return ErrProductNotActive
	}
	if !discount.IsValidAt(now) {
		return ErrDiscountNotActive
	}

	p.discount = discount
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)

	p.events = append(p.events, &DiscountAppliedEvent{
		baseEvent:  baseEvent{occurredAt: now},
		ProductID:  p.id,
		Percentage: discount.Percentage(),
		StartDate:  discount.StartDate(),
		EndDate:    discount.EndDate(),
	})
	return nil
}

func (p *Product) RemoveDiscount(now time.Time) error {
	if p.discount == nil {
		return ErrNoActiveDiscount
	}

	p.discount = nil
	p.updatedAt = now
	p.changes.MarkDirty(FieldDiscount)

	p.events = append(p.events, &DiscountRemovedEvent{
		baseEvent: baseEvent{occurredAt: now},
		ProductID: p.id,
	})
	return nil
}
