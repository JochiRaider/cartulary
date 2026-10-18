//go:build linux

package harnesscollect

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/leaseproof"

	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

// Deliberately excludes discovery and every mutation. Lifecycle owners provide
// exact private lease proof; this adapter never acquires or releases services.
type containerObservationClient interface {
	ContainerInspect(context.Context, string, dockerclient.ContainerInspectOptions) (dockerclient.ContainerInspectResult, error)
	ContainerStats(context.Context, string, dockerclient.ContainerStatsOptions) (dockerclient.ContainerStatsResult, error)
}
type containerObservationProof struct {
	Version        int              `json:"version"`
	RunID          string           `json:"run_id"`
	RuntimeLeaseID string           `json:"runtime_lease_id"`
	Ownership      string           `json:"ownership"`
	BorrowerFile   string           `json:"borrower_file"`
	BorrowerID     string           `json:"borrower_id"`
	Lease          leaseproof.Lease `json:"lease"`
}
type containerMetric struct {
	Value        *float64 `json:"value"`
	Availability string   `json:"availability"`
}
type containerObservation struct {
	Identity     string                     `json:"identity_digest"`
	Ownership    string                     `json:"ownership"`
	Availability string                     `json:"availability"`
	Metrics      map[string]containerMetric `json:"metrics"`
}
type containerObservationBatch struct {
	Availability string                 `json:"availability"`
	Observations []containerObservation `json:"observations"`
	Omitted      int                    `json:"omitted"`
	Overruns     int                    `json:"overruns"`
}

func observationJSON(raw []byte, value any) error {
	if _, err := canonicaljson.DecodeStrict(raw); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}
func privateObservationJSON(file string, value any) error {
	root, err := rootedfs.Open(filepath.Dir(file))
	if err != nil {
		return err
	}
	defer root.Close()
	raw, metadata, err := root.ReadRegular(rootedfs.MustParseReference(filepath.Base(file)), 1048576)
	if err != nil {
		return err
	}
	if metadata.Mode.Perm() != 0600 {
		return errors.New("unsafe observation proof")
	}
	return observationJSON(raw, value)
}

var errContainerProofRevoked = errors.New("container proof revoked")

func readContainerObservationProof(file, runID, leaseID string) (containerObservationProof, error) {
	var published *containerObservationProof
	if err := privateObservationJSON(file, &published); err != nil {
		return containerObservationProof{}, err
	}
	if published == nil {
		return containerObservationProof{}, errContainerProofRevoked
	}
	proof := *published
	lease := proof.Lease
	if proof.Version != 1 || proof.RunID != runID || runID == "" || proof.RuntimeLeaseID != leaseID || leaseID == "" ||
		lease.SchemaID != "cartulary.test_services.lease.v2" || lease.SuiteID == "" || lease.LeaseID == "" || lease.RunID == "" || lease.CleanupState != "not_started" || len(lease.Resources) > 256 {
		return proof, errors.New("foreign or inactive observation proof")
	}
	switch proof.Ownership {
	case "owned":
		if lease.RunID != runID || lease.OwnershipMode != "owned" || proof.BorrowerFile != "" || proof.BorrowerID != "" {
			return proof, errors.New("foreign owned lease")
		}
	case "borrowed":
		var borrower struct {
			SchemaID   string `json:"schema_id"`
			SessionID  string `json:"session_id"`
			BorrowerID string `json:"borrower_id"`
			RunID      string `json:"run_id"`
			PID        int    `json:"pid"`
			Start      string `json:"process_start_proof"`
			CreatedAt  string `json:"created_at"`
		}
		if proof.BorrowerFile == "" || privateObservationJSON(proof.BorrowerFile, &borrower) != nil || borrower.SchemaID != "cartulary.test_services.local_session_borrower.v1" || borrower.RunID != runID || borrower.SessionID != lease.SuiteID || borrower.BorrowerID != proof.BorrowerID || borrower.BorrowerID == "" {
			return proof, errors.New("unproved borrower")
		}
	default:
		return proof, errors.New("unknown observation ownership")
	}
	seen := map[string]bool{}
	for _, resource := range lease.Resources {
		if resource.Kind != "container" || resource.ContainerID == "" || seen[resource.ContainerID] {
			return proof, errors.New("invalid exact container roster")
		}
		seen[resource.ContainerID] = true
		if resource.Labels[leaseproof.LabelRunID] != lease.RunID || resource.Labels[leaseproof.LabelSuiteID] != lease.SuiteID || resource.Labels[leaseproof.LabelManaged] != leaseproof.ManagedValue {
			return proof, errors.New("foreign resource proof")
		}
	}
	return proof, nil
}
func containerStartProof(value dockercontainer.InspectResponse, lease leaseproof.Lease, resource leaseproof.LeaseResource) (string, error) {
	if value.ID != resource.ContainerID || value.Config == nil || value.State == nil || !value.State.Running || value.State.StartedAt == "" {
		return "", errors.New("container absent or replaced")
	}
	for key, expected := range leaseproof.RequiredLeaseLabels(lease, resource) {
		if value.Config.Labels[key] != expected {
			return "", errors.New("container ownership conflict")
		}
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(value.ID+"\x00"+value.State.StartedAt))), nil
}
func observeContainer(ctx context.Context, client containerObservationClient, proof containerObservationProof, resource leaseproof.LeaseResource) (containerObservation, error) {
	before, err := client.ContainerInspect(ctx, resource.ContainerID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return containerObservation{}, err
	}
	identity, err := containerStartProof(before.Container, proof.Lease, resource)
	if err != nil {
		return containerObservation{}, err
	}
	result := containerObservation{Identity: identity, Ownership: proof.Ownership, Availability: "container_gone", Metrics: map[string]containerMetric{}}
	response, err := client.ContainerStats(ctx, resource.ContainerID, dockerclient.ContainerStatsOptions{Stream: false, IncludePreviousSample: false})
	if err != nil {
		return result, err
	}
	if response.Body == nil {
		return result, errors.New("missing container statistics")
	}
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, 1048577))
	closeErr := response.Body.Close()
	if readErr != nil || closeErr != nil || len(raw) > 1048576 {
		return result, errors.New("invalid container statistics")
	}
	var stats dockercontainer.StatsResponse
	if _, err := canonicaljson.DecodeStrict(raw); err != nil {
		return result, err
	}
	if err := json.Unmarshal(raw, &stats); err != nil {
		return result, err
	}
	if (stats.ID != "" && stats.ID != resource.ContainerID) || (stats.OSType != "" && stats.OSType != "linux") {
		return result, errors.New("foreign or unsupported container statistics")
	}
	after, err := client.ContainerInspect(ctx, resource.ContainerID, dockerclient.ContainerInspectOptions{})
	if err != nil {
		return result, err
	}
	endIdentity, err := containerStartProof(after.Container, proof.Lease, resource)
	if err != nil || identity != endIdentity {
		return result, errors.New("container restarted during observation")
	}
	// Presence is distinct from a zero counter. No working-set/RSS translation,
	// Docker percentage, previous-sample wait or process/container sum is used.
	add := func(name string, value uint64, present bool, divisor float64) {
		metric := containerMetric{Availability: "not_observed"}
		if present && value <= 1<<53-1 {
			v := float64(value) / divisor
			metric.Value = &v
			metric.Availability = "available"
		}
		result.Metrics[name] = metric
	}
	present := func(keys ...string) bool {
		value := json.RawMessage(raw)
		for _, key := range keys {
			var object map[string]json.RawMessage
			if json.Unmarshal(value, &object) != nil {
				return false
			}
			var ok bool
			value, ok = object[key]
			if !ok || string(value) == "null" {
				return false
			}
		}
		return true
	}
	cpu := present("cpu_stats", "cpu_usage", "total_usage")
	memoryUsage := present("memory_stats", "usage")
	memoryLimit := present("memory_stats", "limit")
	pids := present("pids_stats", "current")
	add("container_cpu_us", stats.CPUStats.CPUUsage.TotalUsage, cpu, 1000)
	add("container_memory_usage_bytes", stats.MemoryStats.Usage, memoryUsage, 1)
	add("container_memory_limit_bytes", stats.MemoryStats.Limit, memoryLimit, 1)
	add("container_pids", stats.PidsStats.Current, pids, 1)
	var read, written uint64
	readSeen, writeSeen := false, false
	for _, entry := range stats.BlkioStats.IoServiceBytesRecursive {
		switch entry.Op {
		case "Read", "read":
			read += entry.Value
			readSeen = true
		case "Write", "write":
			written += entry.Value
			writeSeen = true
		}
	}
	add("container_block_read_bytes", read, readSeen, 1)
	add("container_block_write_bytes", written, writeSeen, 1)
	result.Availability = "available"
	return result, nil
}
func collectContainerBatch(ctx context.Context, client containerObservationClient, proof containerObservationProof) containerObservationBatch {
	result := containerObservationBatch{Availability: "available", Observations: []containerObservation{}}
	resources := proof.Lease.Resources
	if len(resources) > 32 {
		result.Omitted = len(resources) - 32
		resources = resources[:32]
	}
	observations := make([]containerObservation, len(resources))
	failures := make([]bool, len(resources))
	var group sync.WaitGroup
	for i, resource := range resources {
		group.Add(1)
		go func(i int, resource leaseproof.LeaseResource) {
			defer group.Done()
			var err error
			observations[i], err = observeContainer(ctx, client, proof, resource)
			failures[i] = err != nil
			if err != nil && observations[i].Identity != "" {
				observations[i].Availability = "collector_failed"
				if errors.Is(err, context.DeadlineExceeded) {
					observations[i].Availability = "deadline_exceeded"
				}
				if errors.Is(err, context.Canceled) {
					observations[i].Availability = "paused_measurement"
				}
			}
		}(i, resource)
	}
	group.Wait()
	for i, observation := range observations {
		if failures[i] {
			result.Omitted++
			if observation.Identity == "" {
				continue
			}
		}
		result.Observations = append(result.Observations, observation)
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.Overruns = 1
	}
	return result
}
