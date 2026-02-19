package list_products

import "time"

type ProductSummary struct {
	ID             string
	Name           string
	Category       string
	BasePrice      string
	EffectivePrice string
	Status         string
	CreatedAt      time.Time
}

type ListResult struct {
	Products      []ProductSummary
	NextPageToken string
}
