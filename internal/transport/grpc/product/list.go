package product

import (
	"context"

	"github.com/tshubham2/catalog-proj/internal/app/product/queries/list_products"
	pb "github.com/tshubham2/catalog-proj/proto/product/v1"
)

func (h *Handler) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsReply, error) {
	result, err := h.listProducts.Execute(ctx, list_products.Params{
		PageSize:  int(req.GetPageSize()),
		PageToken: req.GetPageToken(),
		Category:  req.GetCategory(),
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	reply := &pb.ListProductsReply{
		NextPageToken: result.NextPageToken,
		Products:      make([]*pb.ProductSummary, 0, len(result.Products)),
	}
	for _, s := range result.Products {
		reply.Products = append(reply.Products, productSummaryToProto(s))
	}

	return reply, nil
}
