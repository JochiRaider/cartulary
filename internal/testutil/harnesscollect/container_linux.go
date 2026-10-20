//go:build linux

package harnesscollect

import (
	"context"
	"errors"
	"os"

	dockerclient "github.com/moby/moby/client"
)

type containerBatch struct {
	availability      string
	observations      []object
	omitted, overruns int
}

// This adapter runs only within the collector's existing observation fence.
// It owns a read-only client, never a process, a socket server, or a timer loop.
type containerAdapter struct {
	file, runID, leaseID string
	policy               policy
	client               containerObservationClient
	closeClient          func() error
}

func (c *containerAdapter) close() error {
	if c.closeClient == nil {
		return nil
	}
	closeClient := c.closeClient
	c.closeClient = nil
	c.client = nil
	return closeClient()
}

func (c *containerAdapter) sample(ctx context.Context) (containerBatch, error) {
	empty := containerBatch{availability: "not_observed", observations: []object{}}
	if c.file == "" {
		return empty, nil
	}
	batchContext, cancel := context.WithTimeout(ctx, c.policy.duration("container_batch_deadline_ms"))
	defer cancel()
	proof, err := readContainerObservationProofs(c.file, c.runID, c.leaseID)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) && !errors.Is(err, errContainerProofRevoked) {
			empty.omitted = 1
		}
		return empty, nil
	}
	if c.client == nil {
		client, err := dockerclient.New(dockerclient.FromEnv)
		if err != nil {
			empty.availability, empty.omitted = "collector_failed", 1
			return empty, nil
		}
		c.client, c.closeClient = client, client.Close
	}
	// collectContainerBatch joins every request before returning, including on
	// cancellation. The control loop cannot acknowledge pause/stop before drain.
	batch := collectContainerBatch(batchContext, c.client, proof.Services)
	result := containerBatch{availability: batch.Availability, omitted: batch.Omitted, overruns: batch.Overruns, observations: []object{}}
	for _, observation := range batch.Observations {
		metrics := object{}
		for name, metric := range observation.Metrics {
			var value any
			if metric.Value != nil {
				value = *metric.Value
			}
			metrics[name] = object{"value": value, "availability": metric.Availability}
		}
		result.observations = append(result.observations, object{"identity_digest": observation.Identity, "ownership": observation.Ownership, "availability": observation.Availability, "metrics": metrics})
	}
	return result, nil
}
