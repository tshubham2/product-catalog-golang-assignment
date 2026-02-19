package create_product

import (
	"context"
	"math/big"

	"github.com/google/uuid"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases"
	"github.com/tshubham2/catalog-proj/internal/pkg/clock"
	"github.com/tshubham2/catalog-proj/internal/pkg/committer"
)

type Request struct {
	Name        string
	Description string
	Category    string
	BasePrice   *big.Rat
}

type Interactor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer *committer.Committer
	clock     clock.Clock
}

func NewInteractor(
	repo contracts.ProductRepository,
	outbox contracts.OutboxRepository,
	cm *committer.Committer,
	clk clock.Clock,
) *Interactor {
	return &Interactor{
		repo:      repo,
		outbox:    outbox,
		committer: cm,
		clock:     clk,
	}
}

func (it *Interactor) Execute(ctx context.Context, req Request) (string, error) {
	basePrice, err := domain.NewMoneyFromRat(req.BasePrice)
	if err != nil {
		return "", err
	}

	now := it.clock.Now()
	product, err := domain.NewProduct(uuid.NewString(), req.Name, req.Description, req.Category, basePrice, now)
	if err != nil {
		return "", err
	}

	plan := committer.NewPlan()
	plan.Add(it.repo.InsertMut(product))

	for _, event := range product.DomainEvents() {
		plan.Add(it.outbox.InsertMut(usecases.EnrichEvent(product.ID(), event)))
	}

	if err := it.committer.Apply(ctx, plan); err != nil {
		return "", err
	}

	return product.ID(), nil
}
