package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/tshubham2/catalog-proj/proto/product/v1"
)

func (h *Handler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	dto, err := h.getProduct.Execute(ctx, req.GetProductId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &pb.GetProductReply{
		Product: productDTOToProto(dto),
	}, nil
}

func (h *Handler) ApplyDiscount(ctx context.Context, req *pb.ApplyDiscountRequest) (*pb.ApplyDiscountReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.GetStartDate() == nil || req.GetEndDate() == nil {
		return nil, status.Error(codes.InvalidArgument, "start_date and end_date are required")
	}

	pct, err := parsePercentageString(req.GetPercentage())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = h.applyDiscount.Execute(ctx, apply_discount_request(
		req.GetProductId(), pct,
		req.GetStartDate().AsTime(),
		req.GetEndDate().AsTime(),
	))
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &pb.ApplyDiscountReply{}, nil
}

func (h *Handler) RemoveDiscount(ctx context.Context, req *pb.RemoveDiscountRequest) (*pb.RemoveDiscountReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	if err := h.removeDiscount.Execute(ctx, remove_discount_request(req.GetProductId())); err != nil {
		return nil, mapDomainError(err)
	}

	return &pb.RemoveDiscountReply{}, nil
}
