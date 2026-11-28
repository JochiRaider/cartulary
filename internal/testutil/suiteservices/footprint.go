package suiteservices

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PostgresFootprintSchema identifies a broker generation's memory footprint.
const PostgresFootprintSchema = "cartulary.test_services.postgres_footprint.v1"

// PostgresFootprintFile is the footprint's name in its suite artifact root.
const PostgresFootprintFile = "postgres-footprint.json"

// PostgresFootprint is one broker generation's peak container memory, its
// memory-backed data directory included, against that directory's capacity.
// The work-estimate refresh sizes later generations of the same rows from it
// (TH-HARNESS-REQ-865).
type PostgresFootprint struct {
	SchemaID          string `json:"schema_id"`
	AllocationRef     string `json:"allocation_ref"`
	DataCapacityBytes int64  `json:"data_capacity_bytes"`
	MemoryPeakBytes   int64  `json:"memory_peak_bytes"`
}

// RecordPostgresFootprint writes the footprint of the suite's generation. Only
// a broker allocation, whose reference names a postgres: resident, has one.
func RecordPostgresFootprint(env map[string]string, dataCapacityBytes, memoryPeakBytes int64) error {
	ref := env[PGAllocationRefEnv]
	if !strings.HasPrefix(ref, "postgres:") || dataCapacityBytes <= 0 || memoryPeakBytes <= 0 {
		return nil
	}
	suiteDir, ok, err := ResolveSuiteArtifactDir(env)
	if err != nil || !ok {
		return err
	}
	if err := os.MkdirAll(suiteDir, 0o700); err != nil {
		return fmt.Errorf("create suite-service artifact dir: %w", err)
	}
	payload, err := json.MarshalIndent(PostgresFootprint{SchemaID: PostgresFootprintSchema, AllocationRef: ref, DataCapacityBytes: dataCapacityBytes, MemoryPeakBytes: memoryPeakBytes}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode postgres footprint: %w", err)
	}
	if err := writeFileAtomically(filepath.Join(suiteDir, PostgresFootprintFile), append(payload, '\n'), 0o600); err != nil {
		return fmt.Errorf("write postgres footprint: %w", err)
	}
	return nil
}
