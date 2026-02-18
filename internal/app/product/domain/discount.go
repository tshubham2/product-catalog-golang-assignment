package domain

import (
	"math/big"
	"time"
)

// Discount is a time-bound percentage discount. The percentage field is a
// whole number (e.g. 20 means 20%), stored as *big.Rat to support fractional
// percentages like 12.5 if needed.
type Discount struct {
	percentage *big.Rat
	startDate  time.Time
	endDate    time.Time
}

func NewDiscount(percentage *big.Rat, startDate, endDate time.Time) (*Discount, error) {
	zero := new(big.Rat)
	hundred := new(big.Rat).SetInt64(100)

	if percentage.Cmp(zero) <= 0 || percentage.Cmp(hundred) > 0 {
		return nil, ErrInvalidDiscountPercent
	}
	if !startDate.Before(endDate) {
		return nil, ErrInvalidDiscountPeriod
	}
	return &Discount{
		percentage: new(big.Rat).Set(percentage),
		startDate:  startDate,
		endDate:    endDate,
	}, nil
}

// IsValidAt checks whether the discount window covers the given instant.
// Start is inclusive, end is exclusive.
func (d *Discount) IsValidAt(t time.Time) bool {
	return !t.Before(d.startDate) && t.Before(d.endDate)
}

func (d *Discount) Percentage() *big.Rat { return new(big.Rat).Set(d.percentage) }
func (d *Discount) StartDate() time.Time { return d.startDate }
func (d *Discount) EndDate() time.Time   { return d.endDate }
