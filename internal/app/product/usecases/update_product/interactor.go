package update_product

import (
	"context"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases"
	"github.com/tshubham2/catalog-proj/internal/pkg/clock"
	"github.com/tshubham2/catalog-proj/internal/pkg/committer"
)

type Request struct {
	ProductID   string
	Name        *string
	Description *string
	Category    *string
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

func (it *Interactor) Execute(ctx context.Context, req Request) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	name := product.Name()
	if req.Name != nil {
		name = *req.Name
	}
	desc := product.Description()
	if req.Description != nil {
		desc = *req.Description
	}
	cat := product.Category()
	if req.Category != nil {
		cat = *req.Category
	}

	now := it.clock.Now()
	if err := product.UpdateDetails(name, desc, cat, now); err != nil {
		return err
	}

	plan := committer.NewPlan()
	plan.Add(it.repo.UpdateMut(product))

	for _, event := range product.DomainEvents() {
		plan.Add(it.outbox.InsertMut(usecases.EnrichEvent(product.ID(), event)))
	}

	if err := it.committer.Apply(ctx, plan); err != nil {
		return err
	}

	return nil
}
