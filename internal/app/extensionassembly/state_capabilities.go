package extensionassembly

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
)

// Views hold a private reference rather than embedding the transaction. Their
// dynamic method sets cannot recover coordinator metadata or commit authority.
type stateReadCapability struct{ transaction *stateTransaction }
type stateWriteCapability struct{ stateReadCapability }

func (t *stateTransaction) ReadCapability() extensions.StateReadCapability {
	return &stateReadCapability{transaction: t}
}
func (t *stateTransaction) WriteCapability() extensions.StateWriteCapability {
	return &stateWriteCapability{stateReadCapability{transaction: t}}
}
func (t *stateTransaction) rejectCapabilityScope() error {
	t.scopeViolation.Store(true)
	return t.CapabilityError()
}
func (t *stateTransaction) CapabilityError() error {
	if t.scopeViolation.Load() {
		return errors.New("extension state capability scope violation")
	}
	return nil
}
func (c *stateReadCapability) CapabilityID() string     { return c.transaction.CapabilityID() }
func (c *stateReadCapability) StateFamilyIDs() []string { return c.transaction.StateFamilyIDs() }
func (c *stateReadCapability) FamilyCounts(ctx context.Context, families []string) (map[string]int64, error) {
	return c.transaction.FamilyCounts(ctx, families)
}
func (c *stateReadCapability) ValidateFamilyState(ctx context.Context, family string) error {
	return c.transaction.ValidateFamilyState(ctx, family)
}
func (c *stateWriteCapability) IsStateWriteCapability() bool { return true }
