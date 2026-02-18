package domain

import (
	"fmt"
	"math/big"
)

// Money wraps *big.Rat for precise monetary calculations.
// Stored as numerator/denominator in the DB to avoid any floating-point path.
type Money struct {
	amount *big.Rat
}

func NewMoney(numerator, denominator int64) (*Money, error) {
	if denominator == 0 {
		return nil, fmt.Errorf("money: denominator cannot be zero")
	}
	r := new(big.Rat).SetFrac64(numerator, denominator)
	if r.Sign() < 0 {
		return nil, ErrInvalidPrice
	}
	return &Money{amount: r}, nil
}

func NewMoneyFromRat(r *big.Rat) (*Money, error) {
	if r == nil || r.Sign() < 0 {
		return nil, ErrInvalidPrice
	}
	return &Money{amount: new(big.Rat).Set(r)}, nil
}

// Amount returns a defensive copy of the underlying rational.
func (m *Money) Amount() *big.Rat {
	return new(big.Rat).Set(m.amount)
}

func (m *Money) Numerator() int64  { return m.amount.Num().Int64() }
func (m *Money) Denominator() int64 { return m.amount.Denom().Int64() }

func (m *Money) Multiply(factor *big.Rat) *Money {
	return &Money{amount: new(big.Rat).Mul(m.amount, factor)}
}

func (m *Money) Sub(other *Money) *Money {
	return &Money{amount: new(big.Rat).Sub(m.amount, other.amount)}
}

func (m *Money) Equal(other *Money) bool {
	if m == nil || other == nil {
		return m == other
	}
	return m.amount.Cmp(other.amount) == 0
}

func (m *Money) IsZero() bool {
	return m.amount.Sign() == 0
}

// String formats as a two-decimal string, e.g. "19.99".
func (m *Money) String() string {
	return m.amount.FloatString(2)
}
