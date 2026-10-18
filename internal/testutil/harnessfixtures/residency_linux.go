//go:build linux

package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

type activityKey struct{}

func WithActivity(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, activityKey{}, token)
}

type residentRecord struct {
	Schema     string                    `json:"schema_id"`
	Root       string                    `json:"admission_root"`
	Allocation harnessadmission.Resident `json:"allocation"`
}
type residency struct {
	mu          sync.Mutex
	options     ProductionOptions
	root        string
	parent      *string
	capacities  harnessadmission.Quantities
	allocations map[string]*residentRecord
}

func quantities(input map[string]int64) harnessadmission.Quantities {
	result := harnessadmission.Quantities{}
	for _, key := range []string{"browser_stack", "cpu", "io", "memory_mb", "object_store", "port_lane", "postgres", "process", "service_stack", "volume"} {
		result[key] = float64(input[key])
	}
	return result
}
func (r *residency) ensure(key string, claims map[string]int64) (*residentRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if prior := r.allocations[key]; prior != nil {
		return prior, nil
	}
	token := strings.ReplaceAll(identifier(), "-", "")
	file, err := r.options.Runtime.RecoveryPath("resident-allocations", token+".json")
	if err != nil {
		return nil, err
	}
	record := &residentRecord{Schema: "cartulary.harness_resident_allocation.v1", Root: r.root, Allocation: harnessadmission.Resident{Token: token, Reference: file, Claims: quantities(claims), Capacities: maps.Clone(r.capacities), Parent: r.parent}}
	directory, err := harnesslocal.PrivateDirectory(filepath.Dir(file))
	if err != nil {
		return nil, err
	}
	directory.Close()
	if err := privateJSON(file, record.Schema, record, false); err != nil {
		return nil, err
	}
	if err := r.options.Runtime.RecordDependentResource("resident_allocation", file, "pending", []string{}, file); err != nil {
		return nil, err
	}
	r.allocations[key] = record
	return record, nil
}
func (r *residency) reference(key string) []string {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	record := r.allocations[key]
	if record == nil {
		return nil
	}
	ref, err := r.options.Runtime.ResourceReference("resident_allocation", record.Allocation.Reference)
	if err != nil {
		return nil
	}
	return []string{ref}
}
func stackKey(request Request) string {
	return "stack:" + request.Affinity + ":" + request.FixtureProfile + ":" + request.SnapshotKey
}

func readResident(file string) (*residentRecord, error) {
	data, err := harnesslocal.ReadFile(file, 65536, false)
	if err != nil {
		return nil, err
	}
	value, err := harnessschema.Decode(data)
	if err != nil || harnessschema.Validate("cartulary.harness_resident_allocation.v1", value) != nil {
		return nil, harnesslocal.ErrUnsafe
	}
	record := &residentRecord{}
	if err := json.Unmarshal(data, record); err != nil {
		return nil, err
	}
	if record.Schema != "cartulary.harness_resident_allocation.v1" || record.Allocation.Reference != file {
		return nil, harnesslocal.ErrUnsafe
	}
	return record, nil
}
func releaseResident(ctx context.Context, file string) error {
	record, err := readResident(file)
	if err != nil {
		return err
	}
	// Preserve the immutable proof until runtime finalization. A journal write
	// failure after capacity release must remain safely recoverable/idempotent.
	return harnessadmission.ReleaseResident(ctx, record.Root, record.Allocation)
}

func (r *residency) close(ctx context.Context, key string) error {
	r.mu.Lock()
	record := r.allocations[key]
	r.mu.Unlock()
	if record == nil {
		return nil
	}
	err := r.options.closeResource(ctx, "resident_allocation", record.Allocation.Reference, nil)
	if err == nil {
		r.mu.Lock()
		delete(r.allocations, key)
		r.mu.Unlock()
	}
	return err
}

func (r *residency) drain(ctx context.Context, records []*residentRecord, operation func(context.Context) error) error {
	residents := []harnessadmission.Resident{}
	for _, record := range records {
		residents = append(residents, record.Allocation)
	}
	lease, err := harnessadmission.AcquireDrain(ctx, r.root, quantities(map[string]int64{"cpu": 1, "io": 1, "memory_mb": 64, "process": 1}), r.capacities, r.parent, residents)
	if err != nil {
		return err
	}
	result := operation(WithActivity(ctx, lease.Token))
	release, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()
	return errors.Join(result, lease.Release(release))
}

func (p ProductionOptions) withResidentActivity(ctx context.Context, operation func(context.Context) error) error {
	if p.Runtime == nil {
		return operation(ctx)
	}
	if token, _ := ctx.Value(activityKey{}).(string); token != "" {
		return operation(ctx)
	}
	resources, err := p.Runtime.Resources(4096)
	if err != nil {
		return err
	}
	records := []*residentRecord{}
	for _, resource := range resources {
		if resource.Kind != "resident_allocation" {
			continue
		}
		file, err := resource.Path()
		if err != nil {
			return err
		}
		record, err := readResident(file)
		if err != nil {
			return err
		}
		granted, err := harnessadmission.Transact(ctx, harnessadmission.Request{Root: record.Root, Operation: "resident_granted", Token: record.Allocation.Token, Reference: file})
		if err != nil {
			return err
		}
		if granted {
			records = append(records, record)
		}
	}
	if len(records) == 0 {
		return operation(ctx)
	}
	first := records[0]
	for _, record := range records[1:] {
		if record.Root != first.Root || !reflect.DeepEqual(record.Allocation.Parent, first.Allocation.Parent) || !maps.Equal(record.Allocation.Capacities, first.Allocation.Capacities) {
			return harnesslocal.ErrUnsafe
		}
	}
	r := &residency{root: first.Root, parent: first.Allocation.Parent, capacities: first.Allocation.Capacities}
	return r.drain(ctx, records, operation)
}
