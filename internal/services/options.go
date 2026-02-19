package services

import (
	"cloud.google.com/go/spanner"

	"github.com/tshubham2/catalog-proj/internal/app/product/queries/get_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/queries/list_products"
	"github.com/tshubham2/catalog-proj/internal/app/product/repo"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/activate_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/apply_discount"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/create_product"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases/update_product"
	"github.com/tshubham2/catalog-proj/internal/pkg/clock"
	"github.com/tshubham2/catalog-proj/internal/pkg/committer"
	transport "github.com/tshubham2/catalog-proj/internal/transport/grpc/product"
)

type Container struct {
	Handler *transport.Handler
}

func NewContainer(spannerClient *spanner.Client) *Container {
	clk := clock.RealClock{}
	cm := committer.NewCommitter(spannerClient)

	productRepo := repo.NewProductRepo(spannerClient)
	outboxRepo := repo.NewOutboxRepo()
	readModel := repo.NewProductReadModel(spannerClient)

	createUC := create_product.NewInteractor(productRepo, outboxRepo, cm, clk)
	updateUC := update_product.NewInteractor(productRepo, outboxRepo, cm, clk)
	applyUC := apply_discount.NewApplyInteractor(productRepo, outboxRepo, cm, clk)
	removeUC := apply_discount.NewRemoveInteractor(productRepo, outboxRepo, cm, clk)
	activateUC := activate_product.NewActivateInteractor(productRepo, outboxRepo, cm, clk)
	deactivateUC := activate_product.NewDeactivateInteractor(productRepo, outboxRepo, cm, clk)
	archiveUC := activate_product.NewArchiveInteractor(productRepo, outboxRepo, cm, clk)

	getQ := get_product.NewHandler(readModel, clk)
	listQ := list_products.NewHandler(readModel, clk)

	handler := transport.NewHandler(
		createUC, updateUC, applyUC, removeUC,
		activateUC, deactivateUC, archiveUC,
		getQ, listQ,
	)

	return &Container{Handler: handler}
}
