package committer

import (
	"context"

	"cloud.google.com/go/spanner"

	"github.com/tshubham2/commitplan"
	spannerdriver "github.com/tshubham2/commitplan/drivers/spanner"
)

// Plan is a typed wrapper that only accepts *spanner.Mutation, so callers
// don't have to deal with interface{} from the generic commitplan.Plan.
type Plan struct {
	inner *commitplan.Plan
}

func NewPlan() *Plan {
	return &Plan{inner: commitplan.NewPlan()}
}

func (p *Plan) Add(m *spanner.Mutation) {
	if m != nil {
		p.inner.Add(m)
	}
}

func (p *Plan) IsEmpty() bool {
	return p.inner.IsEmpty()
}

type Committer struct {
	driver *spannerdriver.Committer
}

func NewCommitter(client *spanner.Client) *Committer {
	return &Committer{driver: spannerdriver.NewCommitter(client)}
}

func (c *Committer) Apply(ctx context.Context, plan *Plan) error {
	return c.driver.Apply(ctx, plan.inner)
}
