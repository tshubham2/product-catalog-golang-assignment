package activate_product

import (
	"context"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases"
	"github.com/tshubham2/catalog-proj/internal/pkg/clock"
	"github.com/tshubham2/catalog-proj/internal/pkg/committer"
)

type Request struct {
	ProductID string
}

// --- Activate ---

type ActivateInteractor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer *committer.Committer
	clock     clock.Clock
}

func NewActivateInteractor(
	repo contracts.ProductRepository,
	outbox contracts.OutboxRepository,
	cm *committer.Committer,
	clk clock.Clock,
) *ActivateInteractor {
	return &ActivateInteractor{
		repo:      repo,
		outbox:    outbox,
		committer: cm,
		clock:     clk,
	}
}

func (it *ActivateInteractor) Execute(ctx context.Context, req Request) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	if err := product.Activate(it.clock.Now()); err != nil {
		return err
	}

	plan := committer.NewPlan()
	plan.Add(it.repo.UpdateMut(product))

	for _, event := range product.DomainEvents() {
		plan.Add(it.outbox.InsertMut(usecases.EnrichEvent(product.ID(), event)))
	}

	return it.committer.Apply(ctx, plan)
}

// --- Deactivate ---

type DeactivateInteractor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer *committer.Committer
	clock     clock.Clock
}

func NewDeactivateInteractor(
	repo contracts.ProductRepository,
	outbox contracts.OutboxRepository,
	cm *committer.Committer,
	clk clock.Clock,
) *DeactivateInteractor {
	return &DeactivateInteractor{
		repo:      repo,
		outbox:    outbox,
		committer: cm,
		clock:     clk,
	}
}

func (it *DeactivateInteractor) Execute(ctx context.Context, req Request) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	if err := product.Deactivate(it.clock.Now()); err != nil {
		return err
	}

	plan := committer.NewPlan()
	plan.Add(it.repo.UpdateMut(product))

	for _, event := range product.DomainEvents() {
		plan.Add(it.outbox.InsertMut(usecases.EnrichEvent(product.ID(), event)))
	}

	return it.committer.Apply(ctx, plan)
}

// --- Archive ---

type ArchiveInteractor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer *committer.Committer
	clock     clock.Clock
}

func NewArchiveInteractor(
	repo contracts.ProductRepository,
	outbox contracts.OutboxRepository,
	cm *committer.Committer,
	clk clock.Clock,
) *ArchiveInteractor {
	return &ArchiveInteractor{
		repo:      repo,
		outbox:    outbox,
		committer: cm,
		clock:     clk,
	}
}

func (it *ArchiveInteractor) Execute(ctx context.Context, req Request) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	if err := product.Archive(it.clock.Now()); err != nil {
		return err
	}

	plan := committer.NewPlan()
	plan.Add(it.repo.UpdateMut(product))

	for _, event := range product.DomainEvents() {
		plan.Add(it.outbox.InsertMut(usecases.EnrichEvent(product.ID(), event)))
	}

	return it.committer.Apply(ctx, plan)
}
