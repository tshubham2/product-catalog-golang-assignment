package domain

// Field name constants used by both the ChangeTracker and the repository
// layer for building targeted UPDATE mutations.
const (
	FieldName        = "name"
	FieldDescription = "description"
	FieldCategory    = "category"
	FieldBasePrice   = "base_price"
	FieldDiscount    = "discount"
	FieldStatus      = "status"
)

// ChangeTracker keeps track of which aggregate fields have been modified
// since the last load. The repo reads this to build partial updates.
type ChangeTracker struct {
	dirtyFields map[string]bool
}

func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{dirtyFields: make(map[string]bool)}
}

func (ct *ChangeTracker) MarkDirty(field string) { ct.dirtyFields[field] = true }
func (ct *ChangeTracker) Dirty(field string) bool { return ct.dirtyFields[field] }
func (ct *ChangeTracker) HasChanges() bool        { return len(ct.dirtyFields) > 0 }
