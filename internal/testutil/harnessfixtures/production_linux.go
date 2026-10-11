//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"maps"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

// Production owns one dependency tree: leases -> physical allocations -> suite.
// Closing dependent allocations never closes the shared suite underneath them.
type Production struct {
	Broker  *Broker
	Suite   *SuiteOwner
	Options ProductionOptions
}

func NewProduction(p ProductionOptions, options Options) *Production {
	suite := &SuiteOwner{Acquire: p.StartSuite}
	if p.Environment["CARTULARY_TEST_SERVICES_MODE"] == "attach" {
		suite.Acquire = p.AttachSuite
	}
	providers := map[string]Provider{}
	for _, capability := range []string{"postgres_transaction", "postgres_dedicated", "postgres_migration", "managed_process"} {
		providers[capability] = &ProviderFuncs{AcquireFunc: func(ctx context.Context, _ Request) (*Allocation, error) {
			s, err := suite.Ensure(ctx)
			if err != nil {
				return nil, err
			}
			return &Allocation{Ownership: "borrowed", ResourceIDs: []string{"suite:" + capability}, Environment: maps.Clone(s.Environment), Detach: func(context.Context) error { return nil }}, nil
		}}
	}
	providers["browser_stack"] = &browserProvider{options: p, suite: suite, allocations: map[string]int{}, clones: map[string]int{}}
	providers["object_store_namespace"] = &proxyProvider{options: p, suite: suite}
	options.Providers = providers
	if options.AllocationID == nil {
		options.AllocationID = func() string { return "allocation:" + identifier() }
	}
	if options.Publish == nil {
		options.Publish = func(record LeaseRecord) error {
			directory := filepath.Join(p.RunRoot, "_shared", "fixture-leases")
			parent, err := harnesslocal.PrivateDirectory(directory)
			if err != nil {
				return err
			}
			defer parent.Close()
			bytes, err := validateJSON(record.Schema, record)
			if err != nil {
				return err
			}
			return harnesslocal.ReplaceFileAt(parent, record.ID+".json", append(bytes, '\n'))
		}
	}
	return &Production{Broker: New(options), Suite: suite, Options: p}
}
func (p *Production) Close(ctx context.Context) error {
	err := p.Broker.Close(ctx)
	resources, proofErr := p.Options.Runtime.Resources(4096)
	unresolved := p.Broker.Unresolved() || proofErr != nil
	for _, resource := range resources {
		if resource.Kind == "browser_stack" || resource.Kind == "object_store_proxy" {
			unresolved = true
		}
	}
	if unresolved {
		return errors.Join(err, proofErr, CleanupFailure(errors.New("dependent fixture cleanup unresolved")))
	}
	return errors.Join(err, p.Suite.Close(ctx))
}
