package spanner

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"

	"github.com/tshubham2/commitplan"
)

type Committer struct {
	client *spanner.Client
}

func NewCommitter(client *spanner.Client) *Committer {
	return &Committer{client: client}
}

// Apply writes all mutations from the plan in a single Spanner transaction.
func (c *Committer) Apply(ctx context.Context, plan *commitplan.Plan) error {
	if plan.IsEmpty() {
		return nil
	}

	ms := make([]*spanner.Mutation, 0, len(plan.Mutations()))
	for _, m := range plan.Mutations() {
		sm, ok := m.(*spanner.Mutation)
		if !ok {
			return fmt.Errorf("commitplan/spanner: unexpected mutation type %T", m)
		}
		ms = append(ms, sm)
	}

	_, err := c.client.Apply(ctx, ms)
	return err
}
