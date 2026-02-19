package apply_discount

import (
	"context"
	"math/big"
	"time"

	"github.com/tshubham2/catalog-proj/internal/app/product/contracts"
	"github.com/tshubham2/catalog-proj/internal/app/product/domain"
	"github.com/tshubham2/catalog-proj/internal/app/product/usecases"
	"github.com/tshubham2/catalog-proj/internal/pkg/clock"
	"github.com/tshubham2/catalog-proj/internal/pkg/committer"
)

// --- Apply ---

type ApplyRequest struct {
	ProductID  string
	Percentage *big.Rat
	StartDate  time.Time
	EndDate    time.Time
}

type ApplyInteractor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer *committer.Committer
	clock     clock.Clock
}

func NewApplyInteractor(
	repo contracts.ProductRepository,
	outbox contracts.OutboxRepository,
	cm *committer.Committer,
	clk clock.Clock,
) *ApplyInteractor {
	return &ApplyInteractor{
		repo:      repo,
		outbox:    outbox,
		committer: cm,
		clock:     clk,
	}
}

func (it *ApplyInteractor) Execute(ctx context.Context, req ApplyRequest) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	discount, err := domain.NewDiscount(req.Percentage, req.StartDate, req.EndDate)
	if err != nil {
		return err
	}

	now := it.clock.Now()
	if err := product.ApplyDiscount(discount, now); err != nil {
		return err
	}

	plan := committer.NewPlan()
	plan.Add(it.repo.UpdateMut(product))

	for _, event := range product.DomainEvents() {
		plan.Add(it.outbox.InsertMut(usecases.EnrichEvent(product.ID(), event)))
	}

	return it.committer.Apply(ctx, plan)
}

// --- Remove ---

type RemoveRequest struct {
	ProductID string
}

type RemoveInteractor struct {
	repo      contracts.ProductRepository
	outbox    contracts.OutboxRepository
	committer *committer.Committer
	clock     clock.Clock
}

func NewRemoveInteractor(
	repo contracts.ProductRepository,
	outbox contracts.OutboxRepository,
	cm *committer.Committer,
	clk clock.Clock,
) *RemoveInteractor {
	return &RemoveInteractor{
		repo:      repo,
		outbox:    outbox,
		committer: cm,
		clock:     clk,
	}
}

func (it *RemoveInteractor) Execute(ctx context.Context, req RemoveRequest) error {
	product, err := it.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return err
	}

	if err := product.RemoveDiscount(it.clock.Now()); err != nil {
		return err
	}

	plan := committer.NewPlan()
	plan.Add(it.repo.UpdateMut(product))

	for _, event := range product.DomainEvents() {
		plan.Add(it.outbox.InsertMut(usecases.EnrichEvent(product.ID(), event)))
	}

	return it.committer.Apply(ctx, plan)
}
