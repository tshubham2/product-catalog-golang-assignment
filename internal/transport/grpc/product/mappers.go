package product

import (
	"fmt"
	"math/big"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/tshubham2/catalog-proj/internal/app/product/queries/get_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/queries/list_products"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/activate_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/apply_discount"
	pb "github.com/tshubham2/catalog-proj/proto/product/v1"
)

func parseMoneyString(s string) (*big.Rat, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, fmt.Errorf("invalid price format: %q", s)
	}
	if r.Sign() < 0 {
		return nil, fmt.Errorf("price must be non-negative")
	}
	return r, nil
}

func parsePercentageString(s string) (*big.Rat, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, fmt.Errorf("invalid percentage format: %q", s)
	}
	return r, nil
}

func productDTOToProto(dto *get_product.ProductDTO) *pb.Product {
	p := &pb.Product{
		Id:             dto.ID,
		Name:           dto.Name,
		Description:    dto.Description,
		Category:       dto.Category,
		BasePrice:      dto.BasePrice,
		EffectivePrice: dto.EffectivePrice,
		Status:         dto.Status,
		CreatedAt:      timestamppb.New(dto.CreatedAt),
		UpdatedAt:      timestamppb.New(dto.UpdatedAt),
	}
	if dto.DiscountPercent != nil {
		p.DiscountPercent = dto.DiscountPercent
	}
	return p
}

func productSummaryToProto(s list_products.ProductSummary) *pb.ProductSummary {
	return &pb.ProductSummary{
		Id:             s.ID,
		Name:           s.Name,
		Category:       s.Category,
		BasePrice:      s.BasePrice,
		EffectivePrice: s.EffectivePrice,
		Status:         s.Status,
		CreatedAt:      timestamppb.New(s.CreatedAt),
	}
}

func activate_product_request(productID string) activate_product.Request {
	return activate_product.Request{ProductID: productID}
}

func apply_discount_request(productID string, pct *big.Rat, start, end time.Time) apply_discount.ApplyRequest {
	return apply_discount.ApplyRequest{
		ProductID:  productID,
		Percentage: pct,
		StartDate:  start,
		EndDate:    end,
	}
}

func remove_discount_request(productID string) apply_discount.RemoveRequest {
	return apply_discount.RemoveRequest{ProductID: productID}
}
