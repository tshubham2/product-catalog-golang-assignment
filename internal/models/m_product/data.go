package m_product

import (
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
)

type Data struct {
	ProductID            string
	Name                 string
	Description          string
	Category             string
	BasePriceNumerator   int64
	BasePriceDenominator int64
	DiscountPercent      spanner.NullNumeric
	DiscountStartDate    spanner.NullTime
	DiscountEndDate      spanner.NullTime
	Status               string
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ArchivedAt           spanner.NullTime
}

type Model struct{}

func New() *Model { return &Model{} }

func (m *Model) InsertMap(values map[string]interface{}) *spanner.Mutation {
	return spanner.InsertMap(Table, values)
}

func (m *Model) UpdateMap(id string, values map[string]interface{}) *spanner.Mutation {
	values[ProductID] = id
	return spanner.UpdateMap(Table, values)
}

func (m *Model) ToRow(d *Data) map[string]interface{} {
	row := map[string]interface{}{
		ProductID:            d.ProductID,
		Name:                 d.Name,
		Description:          d.Description,
		Category:             d.Category,
		BasePriceNumerator:   d.BasePriceNumerator,
		BasePriceDenominator: d.BasePriceDenominator,
		Status:               d.Status,
		CreatedAt:            d.CreatedAt,
		UpdatedAt:            d.UpdatedAt,
	}

	if d.DiscountPercent.Valid {
		row[DiscountPercent] = d.DiscountPercent.Numeric
	}
	if d.DiscountStartDate.Valid {
		row[DiscountStartDate] = d.DiscountStartDate.Time
	}
	if d.DiscountEndDate.Valid {
		row[DiscountEndDate] = d.DiscountEndDate.Time
	}
	if d.ArchivedAt.Valid {
		row[ArchivedAt] = d.ArchivedAt.Time
	}

	return row
}

func (m *Model) FromRow(row *spanner.Row) (*Data, error) {
	d := &Data{}
	err := row.Columns(
		&d.ProductID, &d.Name, &d.Description, &d.Category,
		&d.BasePriceNumerator, &d.BasePriceDenominator,
		&d.DiscountPercent, &d.DiscountStartDate, &d.DiscountEndDate,
		&d.Status, &d.CreatedAt, &d.UpdatedAt, &d.ArchivedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// DiscountPercentRat returns the discount as *big.Rat, or nil if NULL.
func (d *Data) DiscountPercentRat() *big.Rat {
	if !d.DiscountPercent.Valid {
		return nil
	}
	return &d.DiscountPercent.Numeric
}
