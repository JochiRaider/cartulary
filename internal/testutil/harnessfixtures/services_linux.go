//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// Cluster identity is independent of database isolation. Measurement builders
// and their clones belong to one fixture; ordinary exclusive work belongs to
// its unit, and reusable browser work belongs to its allocation lifetime.
func clusterKey(request Request) string {
	if request.PostgresIsolation == "none" {
		return ""
	}
	if request.PostgresIsolation == "shared" {
		return "postgres:shared"
	}
	if request.FixtureProfile != "" {
		// A consumer may enable a different application runtime than the seed
		// builder. The sealed snapshot still belongs to that exact builder.
		return "postgres:measurement:" + request.BuilderUnit + ":" + request.FixtureProfile + ":" + request.SnapshotKey
	}
	if request.Capability == "browser_stack" {
		return "postgres:" + stackKey(request)
	}
	return "postgres:unit:" + request.UnitID
}

func objectStoreBindings(environment map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range environment {
		if strings.HasPrefix(key, "CARTULARY_S3TEST_") || key == "SSL_CERT_FILE" {
			result[key] = value
		}
	}
	return result
}

// serviceEnvironment constructs a new binding from validated descriptors. A
// parent invocation's incidental credentials cannot select a physical service.
func serviceEnvironment(base, binding map[string]string) map[string]string {
	result := map[string]string{}
	for key, value := range base {
		if strings.HasPrefix(key, "CARTULARY_PGTEST_") || strings.HasPrefix(key, "CARTULARY_S3TEST_") || key == "SSL_CERT_FILE" || key == "CARTULARY_TEST_SERVICES_BINDING_LEASE" {
			continue
		}
		result[key] = value
	}
	maps.Copy(result, binding)
	return result
}

func (p *Production) services(ctx context.Context, request Request) (*Suite, error) {
	if err := validateServiceRequest(request); err != nil {
		return nil, err
	}
	var objects *Suite
	if slices.Contains(request.Services, "object_store") {
		var err error
		objects, err = p.Suite.Ensure(ctx)
		if err != nil {
			return nil, err
		}
	}
	if request.PostgresIsolation == "none" {
		if objects == nil {
			return nil, fixtureFailure(errors.New("service fixture has no declared service"))
		}
		return objects, nil
	}
	if p.Options.Environment["CARTULARY_TEST_SERVICES_MODE"] == "attach" && request.PostgresIsolation == "shared" {
		attached, err := p.Suite.Ensure(ctx)
		if err != nil {
			return nil, err
		}
		if attached.Environment[suiteservices.PGIsolationEnv] != "shared" {
			return nil, fixtureFailure(errors.New("incompatible shared PostgreSQL attachment"))
		}
		return attached, nil
	}
	key := clusterKey(request)
	p.servicesMu.Lock()
	owner := p.clusters[key]
	if owner == nil {
		options := p.Options
		options.ServiceKey = key
		options.ServiceID = "postgres-" + identifier()
		options.ServiceDependencies = []string{"postgres"}
		options.PostgresIsolation = request.PostgresIsolation
		options.PostgresAllocation = key
		owner = &SuiteOwner{Acquire: options.StartSuite, ReleaseResidency: func(ctx context.Context) error {
			return p.Options.residency.close(ctx, key)
		}}
		p.clusters[key] = owner
	}
	p.servicesMu.Unlock()
	suite, err := owner.Ensure(ctx)
	if err != nil || objects == nil {
		return suite, err
	}
	owner.mu.Lock()
	defer owner.mu.Unlock()
	if owner.closed {
		return nil, fixtureFailure(errors.New("cluster closed before service binding"))
	}
	if owner.binding != nil {
		return owner.binding, nil
	}
	bound := *suite
	bound.Environment = maps.Clone(suite.Environment)
	bound.ResourceRefs = append(slices.Clone(suite.ResourceRefs), objects.ResourceRef)
	proof, err := readObject(suite.LeaseFile, "cartulary.test_services.lease.v3", 1048576)
	if err != nil {
		return nil, err
	}
	if err := composeServiceBinding(&bound, objects, proof); err != nil {
		return nil, err
	}
	owner.binding = &bound
	return &bound, nil
}

func (p *Production) releaseCluster(ctx context.Context, request Request) error {
	if request.PostgresIsolation != "exclusive" || request.FixtureProfile != "" {
		return nil
	}
	return p.retireCluster(ctx, clusterKey(request))
}

func (p *Production) retireCluster(ctx context.Context, key string) error {
	p.servicesMu.Lock()
	owner := p.clusters[key]
	p.servicesMu.Unlock()
	if owner == nil {
		return nil
	}
	if err := owner.Close(ctx); err != nil {
		return err
	}
	p.servicesMu.Lock()
	if p.clusters[key] == owner {
		delete(p.clusters, key)
	}
	p.servicesMu.Unlock()
	return nil
}

func composeServiceBinding(suite, objects *Suite, proof map[string]any) error {
	if objects == nil {
		return nil
	}
	objectProof, err := readObject(objects.LeaseFile, "cartulary.test_services.lease.v3", 1048576)
	if err != nil {
		return err
	}
	objectResources := []any{}
	for _, raw := range objectProof["resources"].([]any) {
		resource := raw.(map[string]any)
		if resource["service"] == "object_store" {
			objectResources = append(objectResources, resource)
		}
	}
	if len(objectResources) != 1 {
		return errors.New("object-store binding lacks exact physical proof")
	}
	binding := maps.Clone(proof)
	binding["mode"], binding["ownership_mode"] = "attach", "attach"
	binding["resources"] = append(slices.Clone(proof["resources"].([]any)), objectResources...)
	file := filepath.Join(filepath.Dir(suite.LeaseFile), "binding-lease.json")
	if err := privateJSON(file, "cartulary.test_services.lease.v3", binding, false); err != nil {
		return err
	}
	maps.Copy(suite.Environment, objectStoreBindings(objects.Environment))
	suite.Environment["CARTULARY_TEST_SERVICES_BINDING_LEASE"] = file
	suite.LeaseFile = file
	return nil
}
