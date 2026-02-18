package services

import (
	"math/big"
	"time"

	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
)

// CalculateEffectivePrice returns basePrice * (100 - discountPercent) / 100
// when the discount is currently active, or the base price unchanged otherwise.
func CalculateEffectivePrice(basePrice *domain.Money, discount *domain.Discount, now time.Time) *domain.Money {
	if discount == nil || !discount.IsValidAt(now) {
		return basePrice
	}

	hundred := new(big.Rat).SetInt64(100)
	pct := discount.Percentage()
	factor := new(big.Rat).Sub(hundred, pct)
	factor.Quo(factor, hundred)

	return basePrice.Multiply(factor)
}
