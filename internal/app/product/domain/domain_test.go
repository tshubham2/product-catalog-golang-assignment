package domain_test

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain/services"
)

// --- Money ---

func TestNewMoney(t *testing.T) {
	m, err := domain.NewMoney(1999, 100)
	require.NoError(t, err)
	assert.Equal(t, "19.99", m.String())
	assert.Equal(t, int64(1999), m.Numerator())
	assert.Equal(t, int64(100), m.Denominator())
}

func TestNewMoney_NegativePrice(t *testing.T) {
	_, err := domain.NewMoney(-100, 1)
	assert.ErrorIs(t, err, domain.ErrInvalidPrice)
}

func TestNewMoney_ZeroDenominator(t *testing.T) {
	_, err := domain.NewMoney(100, 0)
	assert.Error(t, err)
}

func TestMoney_Arithmetic(t *testing.T) {
	a, _ := domain.NewMoney(2000, 100) // 20.00
	b, _ := domain.NewMoney(500, 100)  // 5.00

	diff := a.Sub(b)
	assert.Equal(t, "15.00", diff.String())

	doubled := a.Multiply(big.NewRat(2, 1))
	assert.Equal(t, "40.00", doubled.String())
}

func TestMoney_Equal(t *testing.T) {
	a, _ := domain.NewMoney(1000, 100)
	b, _ := domain.NewMoney(10, 1) // same value, different fraction
	assert.True(t, a.Equal(b))
}

// --- Discount ---

func TestNewDiscount_Valid(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)

	d, err := domain.NewDiscount(big.NewRat(20, 1), start, end)
	require.NoError(t, err)
	assert.True(t, d.IsValidAt(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)))
	assert.False(t, d.IsValidAt(time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)))
	assert.False(t, d.IsValidAt(end)) // end is exclusive
}

func TestNewDiscount_InvalidPercentage(t *testing.T) {
	start := time.Now()
	end := start.Add(24 * time.Hour)

	_, err := domain.NewDiscount(big.NewRat(0, 1), start, end)
	assert.ErrorIs(t, err, domain.ErrInvalidDiscountPercent)

	_, err = domain.NewDiscount(big.NewRat(101, 1), start, end)
	assert.ErrorIs(t, err, domain.ErrInvalidDiscountPercent)
}

func TestNewDiscount_InvalidPeriod(t *testing.T) {
	now := time.Now()
	_, err := domain.NewDiscount(big.NewRat(10, 1), now, now)
	assert.ErrorIs(t, err, domain.ErrInvalidDiscountPeriod)

	_, err = domain.NewDiscount(big.NewRat(10, 1), now, now.Add(-time.Hour))
	assert.ErrorIs(t, err, domain.ErrInvalidDiscountPeriod)
}

// --- Product lifecycle ---

func TestNewProduct(t *testing.T) {
	now := time.Now().UTC()
	price, _ := domain.NewMoney(1999, 100)

	p, err := domain.NewProduct("id-1", "Widget", "A widget", "gadgets", price, now)
	require.NoError(t, err)

	assert.Equal(t, "id-1", p.ID())
	assert.Equal(t, "Widget", p.Name())
	assert.Equal(t, domain.ProductStatusActive, p.Status())
	assert.Len(t, p.DomainEvents(), 1)
	assert.Equal(t, "product.created", p.DomainEvents()[0].EventType())
}

func TestNewProduct_Validation(t *testing.T) {
	price, _ := domain.NewMoney(100, 1)
	now := time.Now()

	_, err := domain.NewProduct("id", "", "desc", "cat", price, now)
	assert.ErrorIs(t, err, domain.ErrProductNameRequired)

	_, err = domain.NewProduct("id", "name", "desc", "", price, now)
	assert.ErrorIs(t, err, domain.ErrCategoryRequired)
}

func TestProduct_UpdateDetails(t *testing.T) {
	p := activeProduct(t)
	p.ClearEvents()
	now := time.Now().UTC()

	err := p.UpdateDetails("New Name", p.Description(), p.Category(), now)
	require.NoError(t, err)

	assert.Equal(t, "New Name", p.Name())
	assert.True(t, p.Changes().Dirty(domain.FieldName))
	assert.False(t, p.Changes().Dirty(domain.FieldCategory))
	assert.Len(t, p.DomainEvents(), 1)
	assert.Equal(t, "product.updated", p.DomainEvents()[0].EventType())
}

func TestProduct_UpdateDetails_NoChanges(t *testing.T) {
	p := activeProduct(t)
	p.ClearEvents()

	err := p.UpdateDetails(p.Name(), p.Description(), p.Category(), time.Now())
	require.NoError(t, err)
	assert.Empty(t, p.DomainEvents())
}

func TestProduct_Deactivate_Activate(t *testing.T) {
	p := activeProduct(t)
	now := time.Now().UTC()

	err := p.Deactivate(now)
	require.NoError(t, err)
	assert.Equal(t, domain.ProductStatusInactive, p.Status())

	err = p.Activate(now.Add(time.Second))
	require.NoError(t, err)
	assert.Equal(t, domain.ProductStatusActive, p.Status())
}

func TestProduct_CannotDeactivateInactiveProduct(t *testing.T) {
	p := activeProduct(t)
	require.NoError(t, p.Deactivate(time.Now()))

	err := p.Deactivate(time.Now())
	assert.ErrorIs(t, err, domain.ErrProductAlreadyInactive)
}

func TestProduct_CannotActivateActiveProduct(t *testing.T) {
	p := activeProduct(t)
	err := p.Activate(time.Now())
	assert.ErrorIs(t, err, domain.ErrProductAlreadyActive)
}

func TestProduct_Archive(t *testing.T) {
	p := activeProduct(t)
	now := time.Now().UTC()

	err := p.Archive(now)
	require.NoError(t, err)
	assert.Equal(t, domain.ProductStatusArchived, p.Status())
	assert.NotNil(t, p.ArchivedAt())

	// Cannot perform operations on archived products.
	assert.ErrorIs(t, p.Activate(now), domain.ErrProductArchived)
	assert.ErrorIs(t, p.Deactivate(now), domain.ErrProductArchived)
	assert.ErrorIs(t, p.Archive(now), domain.ErrProductArchived)
	assert.ErrorIs(t, p.UpdateDetails("x", "y", "z", now), domain.ErrProductArchived)
}

// --- Discount application ---

func TestProduct_ApplyDiscount(t *testing.T) {
	p := activeProduct(t)
	now := time.Now().UTC()
	discount := validDiscount(t, now)

	err := p.ApplyDiscount(discount, now)
	require.NoError(t, err)
	assert.NotNil(t, p.Discount())
	assert.True(t, p.Changes().Dirty(domain.FieldDiscount))
}

func TestProduct_ApplyDiscount_InactiveProduct(t *testing.T) {
	p := activeProduct(t)
	require.NoError(t, p.Deactivate(time.Now()))
	now := time.Now().UTC()
	discount := validDiscount(t, now)

	err := p.ApplyDiscount(discount, now)
	assert.ErrorIs(t, err, domain.ErrProductNotActive)
}

func TestProduct_RemoveDiscount(t *testing.T) {
	p := activeProduct(t)
	now := time.Now().UTC()
	discount := validDiscount(t, now)
	require.NoError(t, p.ApplyDiscount(discount, now))
	p.ClearEvents()

	err := p.RemoveDiscount(now)
	require.NoError(t, err)
	assert.Nil(t, p.Discount())
	require.Len(t, p.DomainEvents(), 1)
	assert.Equal(t, "discount.removed", p.DomainEvents()[0].EventType())
}

func TestProduct_RemoveDiscount_NoDiscount(t *testing.T) {
	p := activeProduct(t)
	err := p.RemoveDiscount(time.Now())
	assert.ErrorIs(t, err, domain.ErrNoActiveDiscount)
}

// --- Pricing calculator ---

func TestCalculateEffectivePrice_NoDiscount(t *testing.T) {
	base, _ := domain.NewMoney(10000, 100) // $100.00
	result := services.CalculateEffectivePrice(base, nil, time.Now())
	assert.Equal(t, "100.00", result.String())
}

func TestCalculateEffectivePrice_WithDiscount(t *testing.T) {
	base, _ := domain.NewMoney(10000, 100) // $100.00
	now := time.Now().UTC()
	discount := validDiscount(t, now) // 20%

	result := services.CalculateEffectivePrice(base, discount, now)
	assert.Equal(t, "80.00", result.String())
}

func TestCalculateEffectivePrice_ExpiredDiscount(t *testing.T) {
	base, _ := domain.NewMoney(10000, 100)
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
	discount, _ := domain.NewDiscount(big.NewRat(20, 1), start, end)

	result := services.CalculateEffectivePrice(base, discount, time.Now())
	assert.Equal(t, "100.00", result.String())
}

// --- helpers ---

func activeProduct(t *testing.T) *domain.Product {
	t.Helper()
	price, err := domain.NewMoney(1999, 100)
	require.NoError(t, err)
	p, err := domain.NewProduct("test-id", "Test Product", "A test", "electronics", price, time.Now().UTC())
	require.NoError(t, err)
	return p
}

func validDiscount(t *testing.T, now time.Time) *domain.Discount {
	t.Helper()
	d, err := domain.NewDiscount(
		big.NewRat(20, 1),
		now.Add(-time.Hour),
		now.Add(24*time.Hour),
	)
	require.NoError(t, err)
	return d
}
