package contracts

import (
	"context"
	"math/big"
	"time"
)

// ProductReadModel provides read-optimised access, bypassing the aggregate.
type ProductReadModel interface {
	GetByID(ctx context.Context, id string) (*ProductView, error)
	ListActive(ctx context.Context, pageSize int, pageToken string, category string) ([]*ProductView, string, error)
}

// ProductView is a flat projection of a product row. The query layer
// runs pricing calculations on top of this to derive effective prices.
type ProductView struct {
	ID                   string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	DiscountPercent      *big.Rat
	DiscountStartDate    *time.Time
	DiscountEndDate      *time.Time
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
