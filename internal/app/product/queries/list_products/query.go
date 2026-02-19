package list_products

import (
	"context"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain/services"
	"github.com/tshubham2/catalog-proj/internal/pkg/clock"
)

const defaultPageSize = 20

type Handler struct {
	readModel contracts.ProductReadModel
	clock     clock.Clock
}

func NewHandler(rm contracts.ProductReadModel, clk clock.Clock) *Handler {
	return &Handler{readModel: rm, clock: clk}
}

type Params struct {
	PageSize  int
	PageToken string
	Category  string
}

func (h *Handler) Execute(ctx context.Context, params Params) (*ListResult, error) {
	size := params.PageSize
	if size <= 0 {
		size = defaultPageSize
	}

	views, nextToken, err := h.readModel.ListActive(ctx, size, params.PageToken, params.Category)
	if err != nil {
		return nil, err
	}

	now := h.clock.Now()
	result := &ListResult{
		Products:      make([]ProductSummary, 0, len(views)),
		NextPageToken: nextToken,
	}

	for _, v := range views {
		basePrice, _ := domain.NewMoney(v.BasePriceNumerator, v.BasePriceDenominator)

		var discount *domain.Discount
		if v.DiscountPercent != nil && v.DiscountStartDate != nil && v.DiscountEndDate != nil {
			discount, _ = domain.NewDiscount(v.DiscountPercent, *v.DiscountStartDate, *v.DiscountEndDate)
		}

		effectivePrice := services.CalculateEffectivePrice(basePrice, discount, now)

		result.Products = append(result.Products, ProductSummary{
			ID:             v.ID,
			Name:           v.Name,
			Category:       v.Category,
			BasePrice:      basePrice.String(),
			EffectivePrice: effectivePrice.String(),
			Status:         v.Status,
			CreatedAt:      v.CreatedAt,
		})
	}

	return result, nil
}
