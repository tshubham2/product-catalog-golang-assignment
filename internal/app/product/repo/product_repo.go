package repo

import (
	"context"
	"math/big"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/api/iterator"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
	"github.com/tshubham2/catalog-proj/internal/models/m_product"
)

var _ contracts.ProductRepository = (*ProductRepo)(nil) // compile-time check

type ProductRepo struct {
	client *spanner.Client
	model  *m_product.Model
}

func NewProductRepo(client *spanner.Client) *ProductRepo {
	return &ProductRepo{
		client: client,
		model:  m_product.New(),
	}
}

func (r *ProductRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	row, err := r.client.Single().ReadRow(
		ctx, m_product.Table, spanner.Key{id}, m_product.AllColumns,
	)
	if err != nil {
		if spanner.ErrCode(err) == 5 {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	data, err := r.model.FromRow(row)
	if err != nil {
		return nil, err
	}
	return toDomain(data), nil
}

func (r *ProductRepo) InsertMut(p *domain.Product) *spanner.Mutation {
	values := map[string]interface{}{
		m_product.ProductID:            p.ID(),
		m_product.Name:                 p.Name(),
		m_product.Description:          p.Description(),
		m_product.Category:             p.Category(),
		m_product.BasePriceNumerator:   p.BasePrice().Numerator(),
		m_product.BasePriceDenominator: p.BasePrice().Denominator(),
		m_product.Status:               string(p.Status()),
		m_product.CreatedAt:            p.CreatedAt(),
		m_product.UpdatedAt:            p.UpdatedAt(),
	}

	if d := p.Discount(); d != nil {
		pct := d.Percentage()
		values[m_product.DiscountPercent] = *pct
		values[m_product.DiscountStartDate] = d.StartDate()
		values[m_product.DiscountEndDate] = d.EndDate()
	}
	if p.ArchivedAt() != nil {
		values[m_product.ArchivedAt] = *p.ArchivedAt()
	}

	return r.model.InsertMap(values)
}

func (r *ProductRepo) UpdateMut(p *domain.Product) *spanner.Mutation {
	ch := p.Changes()
	if !ch.HasChanges() {
		return nil
	}

	updates := map[string]interface{}{
		m_product.UpdatedAt: p.UpdatedAt(),
	}

	if ch.Dirty(domain.FieldName) {
		updates[m_product.Name] = p.Name()
	}
	if ch.Dirty(domain.FieldDescription) {
		updates[m_product.Description] = p.Description()
	}
	if ch.Dirty(domain.FieldCategory) {
		updates[m_product.Category] = p.Category()
	}
	if ch.Dirty(domain.FieldStatus) {
		updates[m_product.Status] = string(p.Status())
		if p.ArchivedAt() != nil {
			updates[m_product.ArchivedAt] = *p.ArchivedAt()
		}
	}
	if ch.Dirty(domain.FieldDiscount) {
		if d := p.Discount(); d != nil {
			pct := d.Percentage()
			updates[m_product.DiscountPercent] = *pct
			updates[m_product.DiscountStartDate] = d.StartDate()
			updates[m_product.DiscountEndDate] = d.EndDate()
		} else {
			updates[m_product.DiscountPercent] = nil
			updates[m_product.DiscountStartDate] = nil
			updates[m_product.DiscountEndDate] = nil
		}
	}

	return r.model.UpdateMap(p.ID(), updates)
}

// ProductReadModel reads directly from Spanner, bypassing the aggregate.

var _ contracts.ProductReadModel = (*ProductReadModel)(nil)

type ProductReadModel struct {
	client *spanner.Client
}

func NewProductReadModel(client *spanner.Client) *ProductReadModel {
	return &ProductReadModel{client: client}
}

func (rm *ProductReadModel) GetByID(ctx context.Context, id string) (*contracts.ProductView, error) {
	row, err := rm.client.Single().ReadRow(
		ctx, m_product.Table, spanner.Key{id}, m_product.AllColumns,
	)
	if err != nil {
		if spanner.ErrCode(err) == 5 {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}

	data, err := m_product.New().FromRow(row)
	if err != nil {
		return nil, err
	}
	return toView(data), nil
}

func (rm *ProductReadModel) ListActive(ctx context.Context, pageSize int, pageToken string, category string) ([]*contracts.ProductView, string, error) {
	stmt := spanner.Statement{}

	if category != "" {
		stmt.SQL = `SELECT ` + columnsCSV() + ` FROM products
			WHERE status = @status AND category = @category AND product_id > @after
			ORDER BY product_id ASC LIMIT @limit`
		stmt.Params = map[string]interface{}{
			"status":   string(domain.ProductStatusActive),
			"category": category,
			"after":    pageToken,
			"limit":    int64(pageSize + 1),
		}
	} else {
		stmt.SQL = `SELECT ` + columnsCSV() + ` FROM products
			WHERE status = @status AND product_id > @after
			ORDER BY product_id ASC LIMIT @limit`
		stmt.Params = map[string]interface{}{
			"status": string(domain.ProductStatusActive),
			"after":  pageToken,
			"limit":  int64(pageSize + 1),
		}
	}

	iter := rm.client.Single().Query(ctx, stmt)
	defer iter.Stop()

	model := m_product.New()
	var views []*contracts.ProductView

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, "", err
		}
		data, err := model.FromRow(row)
		if err != nil {
			return nil, "", err
		}
		views = append(views, toView(data))
	}

	var nextToken string
	if len(views) > pageSize {
		nextToken = views[pageSize].ID
		views = views[:pageSize]
	}

	return views, nextToken, nil
}

func toDomain(d *m_product.Data) *domain.Product {
	basePrice, _ := domain.NewMoney(d.BasePriceNumerator, d.BasePriceDenominator)

	var discount *domain.Discount
	if pct := d.DiscountPercentRat(); pct != nil && d.DiscountStartDate.Valid && d.DiscountEndDate.Valid {
		discount, _ = domain.NewDiscount(pct, d.DiscountStartDate.Time, d.DiscountEndDate.Time)
	}

	var archivedAt *time.Time
	if d.ArchivedAt.Valid {
		t := d.ArchivedAt.Time
		archivedAt = &t
	}

	return domain.Reconstitute(
		d.ProductID, d.Name, d.Description, d.Category,
		basePrice, discount,
		domain.ProductStatus(d.Status),
		d.CreatedAt, d.UpdatedAt,
		archivedAt,
	)
}

func toView(d *m_product.Data) *contracts.ProductView {
	v := &contracts.ProductView{
		ID:                   d.ProductID,
		Name:                 d.Name,
		Description:          d.Description,
		Category:             d.Category,
		BasePriceNumerator:   d.BasePriceNumerator,
		BasePriceDenominator: d.BasePriceDenominator,
		Status:               d.Status,
		CreatedAt:            d.CreatedAt,
		UpdatedAt:            d.UpdatedAt,
	}
	if pct := d.DiscountPercentRat(); pct != nil {
		v.DiscountPercent = new(big.Rat).Set(pct)
	}
	if d.DiscountStartDate.Valid {
		t := d.DiscountStartDate.Time
		v.DiscountStartDate = &t
	}
	if d.DiscountEndDate.Valid {
		t := d.DiscountEndDate.Time
		v.DiscountEndDate = &t
	}
	return v
}

func columnsCSV() string {
	out := m_product.AllColumns[0]
	for _, c := range m_product.AllColumns[1:] {
		out += ", " + c
	}
	return out
}
