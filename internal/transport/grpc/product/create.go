package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/create_product"
	pb "github.com/tshubham2/catalog-proj/proto/product/v1"
)

func (h *Handler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductReply, error) {
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetCategory() == "" {
		return nil, status.Error(codes.InvalidArgument, "category is required")
	}
	if req.GetBasePrice() == "" {
		return nil, status.Error(codes.InvalidArgument, "base_price is required")
	}

	price, err := parseMoneyString(req.GetBasePrice())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	productID, err := h.createProduct.Execute(ctx, create_product.Request{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Category:    req.GetCategory(),
		BasePrice:   price,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &pb.CreateProductReply{ProductId: productID}, nil
}
