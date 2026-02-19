package product

import (
	"github.com/tshubham2/catalog-proj/internal/app/product/queries/get_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/queries/list_products"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/activate_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/apply_discount"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/create_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/update_product"
	pb "github.com/tshubham2/catalog-proj/proto/product/v1"
)

type Handler struct {
	pb.UnimplementedProductServiceServer

	createProduct  *create_product.Interactor
	updateProduct  *update_product.Interactor
	applyDiscount  *apply_discount.ApplyInteractor
	removeDiscount *apply_discount.RemoveInteractor
	activate       *activate_product.ActivateInteractor
	deactivate     *activate_product.DeactivateInteractor
	archive        *activate_product.ArchiveInteractor
	getProduct     *get_product.Handler
	listProducts   *list_products.Handler
}

func NewHandler(
	cp *create_product.Interactor,
	up *update_product.Interactor,
	ad *apply_discount.ApplyInteractor,
	rd *apply_discount.RemoveInteractor,
	act *activate_product.ActivateInteractor,
	deact *activate_product.DeactivateInteractor,
	arch *activate_product.ArchiveInteractor,
	gp *get_product.Handler,
	lp *list_products.Handler,
) *Handler {
	return &Handler{
		createProduct:  cp,
		updateProduct:  up,
		applyDiscount:  ad,
		removeDiscount: rd,
		activate:       act,
		deactivate:     deact,
		archive:        arch,
		getProduct:     gp,
		listProducts:   lp,
	}
}
