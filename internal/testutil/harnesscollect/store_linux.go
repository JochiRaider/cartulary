//go:build linux

package harnesscollect

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"path/filepath"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

type policy map[string]float64

func (p policy) integer(key string) int            { return int(p[key]) }
func (p policy) duration(key string) time.Duration { return time.Duration(p[key]) * time.Millisecond }
func encodeValidated(id string, v object) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	normalized, err := harnessdiag.Decode(data)
	if err != nil {
		return nil, err
	}
	if err = harnessdiag.Validate(id, normalized); err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

type sampleStream interface {
	Write([]byte) error
	Flush() error
	Close() error
	Abort() error
}
type resourceStore struct {
	directory               string
	policy                  policy
	stream                  sampleStream
	hash                    hash.Hash
	samples, bytes, omitted int
	truncated, failed       bool
}

func newStore(root string, p policy) (*resourceStore, error) {
	directory := filepath.Join(root, "diagnostics")
	stream, err := harnesslocal.NewStream(filepath.Join(directory, "resource-samples.ndjson"), p.integer("flush_bytes"))
	if err != nil {
		return nil, err
	}
	return &resourceStore{directory: directory, policy: p, stream: stream, hash: sha256.New()}, nil
}
func (s *resourceStore) append(record object) error {
	if s.failed {
		return fmt.Errorf("resource writer failed")
	}
	record["schema_id"] = "cartulary.harness_resource_sample.v2"
	record["seq"] = s.samples + 1
	data, err := encodeValidated("cartulary.harness_resource_sample.v2", record)
	if err != nil {
		return err
	}
	if len(data) > s.policy.integer("maximum_record_bytes") || s.bytes+len(data) > s.policy.integer("maximum_sample_bytes") {
		s.omitted++
		s.truncated = true
		return nil
	}
	if err = s.stream.Write(data); err != nil {
		s.failed = true
		return err
	}
	s.hash.Write(data)
	s.samples++
	s.bytes += len(data)
	return nil
}
func (s *resourceStore) flush() error {
	if s.failed {
		return nil
	}
	err := s.stream.Flush()
	if err != nil {
		s.failed = true
	}
	return err
}
func (s *resourceStore) finish(facts object, start time.Time) error {
	var sampleDigest any
	if !s.failed {
		if err := s.stream.Close(); err != nil {
			s.failed = true
		} else {
			sampleDigest = fmt.Sprintf("sha256:%x", s.hash.Sum(nil))
		}
	}
	if s.failed {
		if err := s.stream.Abort(); err != nil {
			return err
		}
		facts["status"] = "collector_failed"
	} else if s.truncated {
		facts["status"] = "truncated"
	}
	facts["samples"] = s.samples
	facts["sample_bytes"] = s.bytes
	facts["sample_digest"] = sampleDigest
	facts["omitted_observations"] = facts["omitted_observations"].(int) + s.omitted
	observer := facts["observer"].(object)
	observer["write_bytes"] = s.bytes
	observer["shutdown_ms"] = float64(time.Since(start)) / float64(time.Millisecond)
	data, err := encodeValidated("cartulary.harness_resource_index.v3", facts)
	if err != nil {
		return err
	}
	if len(data) > s.policy.integer("maximum_metadata_bytes") {
		return fmt.Errorf("resource metadata limit")
	}
	return harnesslocal.PublishNew(filepath.Join(s.directory, "resource-index.json"), data)
}
