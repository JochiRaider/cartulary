//go:build linux

package harnesscollect

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
	dockerclient "github.com/moby/moby/client"
)

const checkpointSchema = "cartulary.harness_postgres_checkpoints.v2"
const checkpointRecordsLimit = 4096
const checkpointBytesLimit = 2 * 1024 * 1024

type checkpointEvidence struct {
	records               []postgrescleanup.CheckpointCompletion
	services, unavailable int
	truncated             bool
	causes                []string
	serviceResults        []checkpointService
	seen                  map[string]bool
}

type checkpointService struct {
	Ordinal    int      `json:"service_ordinal"`
	ScopeRef   string   `json:"service_scope_ref"`
	Generation string   `json:"generation"`
	Identity   *string  `json:"identity_digest"`
	Status     string   `json:"status"`
	Causes     []string `json:"unavailable_causes"`
	Truncated  bool     `json:"truncated"`
}

func (e *checkpointEvidence) unavailableBecause(cause string) {
	e.unavailable++
	if !slices.Contains(e.causes, cause) {
		e.causes = append(e.causes, cause)
		slices.Sort(e.causes)
	}
}
func checkpointCause(err error, fallback string) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "deadline_exceeded"
	}
	if errors.Is(err, context.Canceled) {
		return "cancelled"
	}
	return fallback
}

type checkpointLogClient interface {
	containerObservationClient
	ContainerLogs(context.Context, string, dockerclient.ContainerLogsOptions) (dockerclient.ContainerLogsResult, error)
}

// The existing collector handles this one bounded read before the lifecycle
// owner revokes its proof and tears down the service. There is no streaming
// subscriber, timer, new process, or observation outside the quiet fence.
func (c *collector) captureCheckpoints(ctx context.Context) error {
	return c.captureCheckpointServices(ctx, "")
}
func (c *collector) captureCheckpointServices(ctx context.Context, unavailable string) error {
	if c.containers == nil {
		return nil
	}
	roster, err := readContainerObservationProofs(c.containers.file, c.containers.runID, c.containers.leaseID)
	if err != nil || roster.RetiringService == "" {
		return nil
	}
	if c.checkpoints == nil {
		c.checkpoints = &checkpointEvidence{}
	}
	e := c.checkpoints
	if e.seen == nil {
		e.seen = map[string]bool{}
	}
	for _, proof := range roster.Services {
		if proof.Lease.LeaseID != roster.RetiringService || proof.Ownership != "owned" {
			continue
		}
		for _, resource := range proof.Lease.Resources {
			if resource.Service != "postgres" {
				continue
			}
			key := proof.Lease.LeaseID + ":" + resource.ContainerID
			if e.seen[key] {
				continue
			}
			e.seen[key] = true
			if e.services >= 256 {
				e.truncated = true
				return nil
			}
			e.services++
			service := checkpointService{Ordinal: e.services, ScopeRef: "_shared/test-services/" + proof.Lease.SuiteID + "/service-scope.json", Status: "not_observed", Causes: []string{}}
			if proof.Lease.PostgresProvisioning != nil {
				service.Generation = proof.Lease.PostgresProvisioning.Generation
			}
			observe := func() string {
				if service.Generation == "" {
					return "ownership_unavailable"
				}
				if unavailable != "" {
					return unavailable
				}
				if c.containers.client == nil {
					client, err := dockerclient.New(dockerclient.FromEnv)
					if err != nil {
						return "client_unavailable"
					}
					c.containers.client, c.containers.closeClient = client, client.Close
				}
				client, ok := c.containers.client.(checkpointLogClient)
				if !ok {
					return "client_unavailable"
				}
				deadline, cancel := context.WithTimeout(ctx, c.policy.duration("container_batch_deadline_ms"))
				defer cancel()
				before, err := client.ContainerInspect(deadline, resource.ContainerID, dockerclient.ContainerInspectOptions{})
				if err != nil {
					return checkpointCause(err, "inspection_failed")
				}
				identity, err := containerStartProof(before.Container, proof.Lease, resource)
				if err != nil || before.Container.Config.Tty {
					return "ownership_unavailable"
				}
				service.Identity = &identity
				start, err := time.Parse(time.RFC3339Nano, before.Container.State.StartedAt)
				if err != nil {
					return "clock_unavailable"
				}
				body, err := client.ContainerLogs(deadline, resource.ContainerID, dockerclient.ContainerLogsOptions{ShowStderr: true, Timestamps: true, Follow: false})
				if err != nil {
					return checkpointCause(err, "log_request_failed")
				}
				records, cause, truncated := readCheckpointFrames(body, start, checkpointRecordsLimit-len(e.records))
				closeErr := body.Close()
				after, inspectErr := client.ContainerInspect(deadline, resource.ContainerID, dockerclient.ContainerInspectOptions{})
				afterIdentity, proofErr := containerStartProof(after.Container, proof.Lease, resource)
				if closeErr != nil || inspectErr != nil || proofErr != nil || identity != afterIdentity {
					return checkpointCause(errors.Join(closeErr, inspectErr, proofErr), "settlement_unavailable")
				}
				service.Truncated = truncated
				if len(records) > 0 {
					service.Status = "observed"
				}
				for _, record := range records {
					record.ServiceOrdinal = service.Ordinal
					e.records = append(e.records, record)
				}
				return cause
			}
			if cause := observe(); cause != "" {
				e.unavailableBecause(cause)
				service.Status = "unavailable"
				service.Causes = append(service.Causes, cause)
			}
			if service.Truncated {
				service.Status = "truncated"
				e.truncated = true
			}
			e.serviceResults = append(e.serviceResults, service)
		}
	}
	return nil
}

// Docker multiplexed frame lengths are checked before allocation. Raw input is
// transient, bounded, and never passed to error messages or artifact writers.
func readCheckpointFrames(input io.Reader, start time.Time, remaining int) ([]postgrescleanup.CheckpointCompletion, string, bool) {
	reader := &io.LimitedReader{R: input, N: checkpointBytesLimit + 1}
	records := []postgrescleanup.CheckpointCompletion{}
	cause, truncated := "", false
	var pending strings.Builder
	for {
		var header [8]byte
		_, err := io.ReadFull(reader, header[:])
		if err == io.EOF {
			break
		}
		if err != nil {
			cause = "incomplete_stream"
			truncated = reader.N <= 0
			break
		}
		length := int(binary.BigEndian.Uint32(header[4:]))
		if header[0] != 2 || header[1] != 0 || header[2] != 0 || header[3] != 0 || length > 65536 {
			cause = "invalid_framing"
			break
		}
		payload := make([]byte, length)
		if _, err := io.ReadFull(reader, payload); err != nil {
			cause = "incomplete_stream"
			truncated = reader.N <= 0
			break
		}
		pending.Write(payload)
		if pending.Len() > 65536 {
			cause = "line_limit"
			break
		}
		text := pending.String()
		pending.Reset()
		lines := strings.Split(text, "\n")
		pending.WriteString(lines[len(lines)-1])
		for _, line := range lines[:len(lines)-1] {
			if !strings.Contains(line, "checkpoint complete:") {
				continue
			}
			record, ok := postgrescleanup.ParseCheckpointCompletion(line, start)
			if !ok {
				cause = "unsupported_format"
				continue
			}
			if len(records) >= remaining {
				truncated = true
				return records, cause, truncated
			}
			records = append(records, record)
		}
		if reader.N <= 0 {
			truncated = true
			break
		}
	}
	if pending.Len() > 0 {
		cause = "incomplete_stream"
	}
	return records, cause, truncated
}

func (c *collector) publishCheckpoints() error {
	e := c.checkpoints
	if e == nil {
		e = &checkpointEvidence{}
	}
	records := e.records
	if records == nil {
		records = []postgrescleanup.CheckpointCompletion{}
	}
	status := "not_observed"
	if len(records) > 0 {
		status = "observed"
	}
	if e.unavailable > 0 {
		status = "unavailable"
	}
	if e.truncated {
		status = "truncated"
	}
	causes := e.causes
	if causes == nil {
		causes = []string{}
	}
	services := e.serviceResults
	if services == nil {
		services = []checkpointService{}
	}
	value := object{"service_results": services, "schema_id": checkpointSchema, "clock": "owned_service_start", "status": status, "records": records, "services": e.services, "unavailable": e.unavailable, "unavailable_causes": causes, "truncated": e.truncated}
	for key, valuePart := range c.identity {
		value[key] = valuePart
	}
	// Convert typed records to schema-validator input; no raw log data survives.
	encoded, err := json.Marshal(value)
	if err != nil || len(encoded) > checkpointBytesLimit {
		return errors.New("checkpoint evidence limit")
	}
	var normalized object
	if json.Unmarshal(encoded, &normalized) != nil {
		return errors.New("checkpoint evidence encoding")
	}
	encoded, err = encodeValidated(checkpointSchema, normalized)
	if err != nil {
		return err
	}
	return harnesslocal.PublishNew(filepath.Join(c.store.directory, "postgres-checkpoints.json"), encoded)
}
