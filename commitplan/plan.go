package commitplan

// Mutation is intentionally interface{} — concrete types (like *spanner.Mutation)
// are supplied by callers and only interpreted by the driver.
type Mutation interface{}

// Plan collects mutations that should be applied atomically.
type Plan struct {
	mutations []Mutation
}

func NewPlan() *Plan { return &Plan{} }

func (p *Plan) Add(m Mutation) {
	if m == nil {
		return
	}
	p.mutations = append(p.mutations, m)
}

func (p *Plan) Mutations() []Mutation { return p.mutations }
func (p *Plan) IsEmpty() bool         { return len(p.mutations) == 0 }
