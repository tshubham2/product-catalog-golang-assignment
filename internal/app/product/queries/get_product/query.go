package get_product

import (
	"context"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain/services"
	"github.com/tshubham2/catalog-proj/internal/pkg/clock"
)

type Handler struct {
	readModel contracts.ProductReadModel
	clock     clock.Clock
}

func NewHandler(rm contracts.ProductReadModel, clk clock.Clock) *Handler {
	return &Handler{readModel: rm, clock: clk}
}

func (h *Handler) Execute(ctx context.Context, productID string) (*ProductDTO, error) {
	view, err := h.readModel.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	return h.toDTO(view), nil
}

func (h *Handler) toDTO(v *contracts.ProductView) *ProductDTO {
	now := h.clock.Now()

	basePrice, _ := domain.NewMoney(v.BasePriceNumerator, v.BasePriceDenominator)

	var discount *domain.Discount
	if v.DiscountPercent != nil && v.DiscountStartDate != nil && v.DiscountEndDate != nil {
		discount, _ = domain.NewDiscount(v.DiscountPercent, *v.DiscountStartDate, *v.DiscountEndDate)
	}

	effectivePrice := services.CalculateEffectivePrice(basePrice, discount, now)

	dto := &ProductDTO{
		ID:             v.ID,
		Name:           v.Name,
		Description:    v.Description,
		Category:       v.Category,
		BasePrice:      basePrice.String(),
		EffectivePrice: effectivePrice.String(),
		Status:         v.Status,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}

	if v.DiscountPercent != nil {
		pct := v.DiscountPercent.FloatString(2)
		dto.DiscountPercent = &pct
	}

	return dto
}
