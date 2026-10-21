//go:build linux

package harnessfixtures

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/leaseproof"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/startupproof"
)

func (p ProductionOptions) startupProof(lease string, env map[string]string) (startupproof.Document, error) {
	file := filepath.Join(filepath.Dir(lease), startupproof.Filename)
	d, err := startupproof.Read(file)
	if err != nil {
		return d, err
	}
	b := d.Binding
	if b.RuntimeRoot != p.Runtime.Root || b.RuntimeLease != p.Runtime.LeaseID || b.RunRoot != p.Runtime.RunRoot || b.RunID != p.Runtime.RunID || b.SuiteID != env[suiteservices.SuiteIDEnv] {
		return d, harnesssession.ErrUnresolved
	}
	resources, err := p.Runtime.Resources(4096)
	if err != nil {
		return d, err
	}
	for _, r := range resources {
		path, e := r.Path()
		if e == nil && path == lease && r.Kind == "managed_suite" && slices.Contains(r.RecoveryFiles, file) && slices.Equal(r.Dependencies, b.Dependencies) {
			return d, nil
		}
	}
	return d, harnesssession.ErrUnresolved
}
func (p ProductionOptions) settlePendingSuite(lease string, env map[string]string) error {
	d, err := p.startupProof(lease, env)
	if err != nil {
		return err
	}
	if !d.SettledWithoutLease() {
		return harnesssession.ErrUnresolved
	}
	return p.Runtime.RecordResource("managed_suite", lease, "released")
}

// Explicit recovery projects only returned identities into the existing physical
// owner. It cannot convert an ambiguous create request into an empty lease.
func (p ProductionOptions) startupCleanupLease(lease string, env map[string]string) ([]byte, error) {
	d, err := p.startupProof(lease, env)
	if err != nil {
		return nil, err
	}
	if !p.RecoveryOutput {
		return nil, harnesssession.ErrUnresolved
	}
	proof := leaseproof.Lease{SchemaID: "cartulary.test_services.lease.v3", LeaseID: p.Runtime.LeaseID, SuiteID: d.Binding.SuiteID, RunID: p.Runtime.RunID, ResultRoot: filepath.Dir(p.Runtime.RunRoot), RunRoot: p.Runtime.RunRoot, Target: p.Target, Mode: "owned", OwnershipMode: "owned", OwnerPID: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), CleanupState: "not_started", Resources: []leaseproof.LeaseResource{}, ProofLabels: map[string]string{}, ProofPrefixes: map[string]string{}}
	var helper harnesssession.ProcessProof
	if json.Unmarshal([]byte(d.Helper), &helper) != nil || helper.PID <= 0 {
		return nil, harnesssession.ErrUnresolved
	}
	proof.OwnerPID = helper.PID
	for _, a := range d.Attempts {
		if a.State == "settled" {
			continue
		}
		if a.State != "acquired" || a.ContainerID == "" {
			return nil, harnesssession.ErrUnresolved
		}
		if a.Labels[leaseproof.LabelSuiteID] != d.Binding.SuiteID || a.Labels[leaseproof.LabelRunID] != d.Binding.RunID || a.Labels[leaseproof.LabelService] != a.Service || a.Labels[leaseproof.LabelManaged] != leaseproof.ManagedValue {
			return nil, harnesssession.ErrUnresolved
		}
		image := pgtest.ContainerImage()
		if a.Service == suiteservices.ServiceObjectStore {
			image = s3test.ContainerImage()
		}
		proof.Resources = append(proof.Resources, leaseproof.LeaseResource{Kind: "container", Service: a.Service, ContainerID: a.ContainerID, Image: image, Labels: a.Labels})
	}
	if len(proof.Resources) == 0 {
		return nil, harnesssession.ErrUnresolved
	}
	// Each resource retains its original target label, independently of the
	// recovery command's output target.
	proof.Target = proof.Resources[0].Labels[leaseproof.LabelTarget]
	return json.Marshal(proof)
}
