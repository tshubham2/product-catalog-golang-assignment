package get_product

import "time"

// ProductDTO is the read-side representation returned by the GetProduct query.
type ProductDTO struct {
	ID             string
	Name           string
	Description    string
	Category       string
	BasePrice      string // decimal string, e.g. "19.99"
	EffectivePrice string // after discount, e.g. "15.99"
	DiscountPercent *string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
