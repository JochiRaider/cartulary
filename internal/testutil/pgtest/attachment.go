package pgtest

import (
	"context"
	"errors"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"maps"
	"os"
	"strings"
	"sync"
)

type attachmentOperations struct {
	inspect func(context.Context, string) (string, error)
	ping    func(context.Context, string, func(context.Context)) error
}

// Identity includes allocation and lease scope as well as connection and schema
// identities. A cached generation never bypasses admission of the current binding.
type attachmentIdentity struct {
	admin, templateDSN, template, schema, provisioning, templateDigest string
	generation, isolation, allocation, suite, lease, run               string
}
type admittedAttachment struct {
	identity attachmentIdentity
	binding  fixtureBinding
	fixture  *Harness
	done     chan struct{}
	err      error
}
type attachmentRegistry struct {
	mu          sync.Mutex
	operations  attachmentOperations
	generations map[string]*admittedAttachment
}

func newAttachmentRegistry() *attachmentRegistry {
	return &attachmentRegistry{operations: attachmentOperations{inspectClusterGeneration, pingAdminDSN}, generations: make(map[string]*admittedAttachment)}
}

// Process reuse is intentional; tests construct independent registries.
var sharedAttachments = newAttachmentRegistry()

// fixtureBinding is an immutable snapshot of inputs used for admission,
// execution and event publication. Only entrypoints read process environment.
type fixtureBinding struct{ inputs map[string]string }

var fixtureInputKeys = [...]string{
	suiteservices.PGAdminDSNEnv, suiteservices.PGDSNTemplateEnv, suiteservices.PGTemplateDBEnv,
	suiteservices.PGSchemaHashEnv, suiteservices.PGProvisioningDigestEnv, suiteservices.PGTemplateDigestEnv,
	suiteservices.PGGenerationEnv, suiteservices.PGIsolationEnv, suiteservices.PGAllocationRefEnv,
	suiteservices.SuiteIDEnv, suiteservices.TargetEnv, suiteservices.HarnessServiceDependenciesEnv,
	suiteservices.SuiteRuntimeRootEnv, suiteservices.SuiteRuntimeLeaseIDEnv, suiteservices.SuiteRuntimeRunIDEnv,
	suiteservices.CallModeEnv, suiteservices.LifecycleModeEnv, suiteservices.PersistentBorrowerEnv,
	"CARTULARY_HARNESS_EXECUTION_ENABLED", "CARTULARY_HARNESS_LAUNCH_CONTEXT", "CARTULARY_HARNESS_ALLOCATION_REF",
	"CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID", "CARTULARY_FIXTURE_CAPABILITY",
	postgresFixturePolicyTestsEnv, postgresFixturePolicyPackagesEnv, postgresFixturePolicyDefaultEnv,
}

func processFixtureInputs() map[string]string {
	inputs := make(map[string]string)
	for _, key := range fixtureInputKeys {
		if value, ok := os.LookupEnv(key); ok {
			inputs[key] = value
		}
	}
	return inputs
}

func bindFixture(inputs map[string]string) fixtureBinding {
	copy := make(map[string]string)
	for _, key := range fixtureInputKeys {
		if value, ok := inputs[key]; ok {
			copy[key] = value
		}
	}
	return fixtureBinding{inputs: copy}
}

func (binding fixtureBinding) environment() map[string]string {
	if binding.inputs == nil {
		return map[string]string{}
	}
	return binding.inputs
}

func (binding fixtureBinding) identity() (attachmentIdentity, error) {
	get := func(key string) string { return strings.TrimSpace(binding.inputs[key]) }
	b := attachmentIdentity{admin: get(suiteservices.PGAdminDSNEnv), templateDSN: get(suiteservices.PGDSNTemplateEnv), template: get(suiteservices.PGTemplateDBEnv), schema: get(suiteservices.PGSchemaHashEnv), provisioning: get(suiteservices.PGProvisioningDigestEnv), templateDigest: get(suiteservices.PGTemplateDigestEnv), generation: get(suiteservices.PGGenerationEnv), isolation: get(suiteservices.PGIsolationEnv), allocation: get(suiteservices.PGAllocationRefEnv), suite: get(suiteservices.SuiteIDEnv), lease: get(suiteservices.SuiteRuntimeLeaseIDEnv), run: get(suiteservices.SuiteRuntimeRunIDEnv)}
	if b.admin == "" && b.templateDSN == "" {
		return b, errors.New("postgres requires an explicit suite attachment; dedicated lifecycle fixtures must use StartOwned")
	}
	if b.admin == "" || !strings.Contains(b.templateDSN, "{database}") {
		return b, errors.New("attach postgres harness: incomplete connection binding")
	}
	if b.schema != pgschema.MustHash() || b.provisioning != ProvisioningDigest() || b.templateDigest != TemplateDigest(b.schema) || b.template == "" {
		return b, errors.New("attach postgres harness: incompatible provisioning or template identity")
	}
	if (b.isolation != "shared" && b.isolation != "exclusive") || b.allocation == "" || b.generation == "" {
		return b, errors.New("attach postgres harness: isolation or physical generation ownership is missing")
	}
	return b, nil
}
func (r *attachmentRegistry) attach(ctx context.Context, inputs map[string]string) (*Harness, error) {
	if ctx == nil {
		return nil, errors.New("postgres attachment requires a context")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	binding := bindFixture(inputs)
	if err := suiteservices.CheckServiceDependencies(binding.environment(), "postgres"); err != nil {
		return nil, err
	}
	b, err := binding.identity()
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	if existing, ok := r.generations[b.generation]; ok {
		r.mu.Unlock()
		if b != existing.identity || !maps.Equal(binding.inputs, existing.binding.inputs) {
			return nil, errors.New("attach postgres harness: admitted generation binding changed")
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-existing.done:
			if existing.err != nil {
				return nil, existing.err
			}
			if err := existing.fixture.admitPreparation(); err != nil {
				return nil, err
			}
			return existing.fixture, nil
		}
	}
	admission := &admittedAttachment{identity: b, binding: binding, done: make(chan struct{})}
	r.generations[b.generation] = admission
	r.mu.Unlock()
	actual, err := r.operations.inspect(ctx, b.admin)
	verified := err == nil && actual == b.generation
	var h *Harness
	if !verified {
		err = errors.Join(errors.New("attach postgres harness: physical generation mismatch"), err, ctx.Err())
	} else {
		h = &Harness{binding: binding, adminDSN: b.admin, dsnTemplate: b.templateDSN, templateDB: b.template, schemaHash: b.schema, generation: b.generation, suiteHash: suiteservices.SuiteHash(binding.environment()), processHash: suiteservices.ProcessHash()}
		if err = r.operations.ping(ctx, b.admin, nil); err != nil {
			err = fmt.Errorf("attach postgres harness: ping admin dsn: %w", err)
		}
		err = errors.Join(err, ctx.Err())
		if err == nil {
			err = h.recordEvent(suiteservices.Event{Type: suiteservices.EventPostgresAttach, Details: map[string]any{"schema_hash": h.schemaHash}})
		}
		err = errors.Join(err, ctx.Err())
		if err != nil {
			h.quarantined.Store(true)
			err = errors.Join(err, errGenerationQuarantined)
		}
	}
	r.mu.Lock()
	admission.fixture, admission.err = h, err
	if !verified {
		delete(r.generations, b.generation)
	}
	close(admission.done)
	r.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return h, nil
}
