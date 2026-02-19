package product

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/update_product"
	pb "github.com/tshubham2/catalog-proj/proto/product/v1"
)

func (h *Handler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.UpdateProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	err := h.updateProduct.Execute(ctx, update_product.Request{
		ProductID:   req.GetProductId(),
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &pb.UpdateProductReply{}, nil
}

func (h *Handler) ActivateProduct(ctx context.Context, req *pb.ActivateProductRequest) (*pb.ActivateProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	if err := h.activate.Execute(ctx, activate_product_request(req.GetProductId())); err != nil {
		return nil, mapDomainError(err)
	}
	return &pb.ActivateProductReply{}, nil
}

func (h *Handler) DeactivateProduct(ctx context.Context, req *pb.DeactivateProductRequest) (*pb.DeactivateProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	if err := h.deactivate.Execute(ctx, activate_product_request(req.GetProductId())); err != nil {
		return nil, mapDomainError(err)
	}
	return &pb.DeactivateProductReply{}, nil
}

func (h *Handler) ArchiveProduct(ctx context.Context, req *pb.ArchiveProductRequest) (*pb.ArchiveProductReply, error) {
	if req.GetProductId() == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	if err := h.archive.Execute(ctx, activate_product_request(req.GetProductId())); err != nil {
		return nil, mapDomainError(err)
	}
	return &pb.ArchiveProductReply{}, nil
}
