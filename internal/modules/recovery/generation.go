package recovery

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

type vNextRecoveryGeneration struct {
	id                  string
	captureCurrent      bool
	stateCatalog        *recoverystate.Catalog
	codecSchemaIDs      []string
	codecRegistrySHA256 string
	graph               VNextGraphProjectionRestoreArtifacts
	registrySchemaID    string
	bindingSchemaID     string
}

type vNextRecoveryGenerationRegistry struct {
	current *vNextRecoveryGeneration
	byPair  map[string]*vNextRecoveryGeneration
}

// RecoveryGenerationIdentity is the closed identity selected from a backup's
// exact recovery-state catalog and codec-registry pair.
type RecoveryGenerationIdentity struct {
	RecoveryStateCatalogSHA256       string
	CodecRegistrySHA256              string
	GraphSourceRegistrySHA256        string
	GraphImplementationBindingSHA256 string
	GraphAlgorithmID                 string
}

func (generation *vNextRecoveryGeneration) identity() RecoveryGenerationIdentity {
	if generation == nil {
		return RecoveryGenerationIdentity{}
	}
	return RecoveryGenerationIdentity{
		RecoveryStateCatalogSHA256:       generation.stateCatalog.DigestSHA256(),
		CodecRegistrySHA256:              generation.codecRegistrySHA256,
		GraphSourceRegistrySHA256:        generation.graph.SourceRegistrySHA256,
		GraphImplementationBindingSHA256: generation.graph.ImplementationBindingSHA256,
		GraphAlgorithmID:                 generation.graph.AlgorithmID,
	}
}

func loadVNextRecoveryGenerationRegistry() (*vNextRecoveryGenerationRegistry, error) {
	if contractrecovery.RecoveryGenerationRegistrySchemaID != "cartulary.recovery_generation_registry.v1" ||
		len(contractrecovery.RecoveryGenerations) != 1 {
		return nil, fmt.Errorf("%w: generated Recovery generation registry is malformed", ErrVNextBackup)
	}
	return loadVNextRecoveryGenerationRegistryFrom(contractrecovery.RecoveryGenerations)
}

func loadVNextRecoveryGenerationRegistryFrom(
	projections []contractrecovery.RecoveryGeneration,
) (*vNextRecoveryGenerationRegistry, error) {
	if len(projections) != 1 {
		return nil, fmt.Errorf("%w: Recovery generation registry must contain exactly one current entry", ErrVNextBackup)
	}
	registry := &vNextRecoveryGenerationRegistry{
		byPair: make(map[string]*vNextRecoveryGeneration, len(projections)),
	}
	seenIDs := make(map[string]struct{}, len(projections))
	for _, projected := range projections {
		if strings.TrimSpace(projected.GenerationID) == "" {
			return nil, fmt.Errorf("%w: Recovery generation ID is missing", ErrVNextBackup)
		}
		if _, duplicate := seenIDs[projected.GenerationID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate Recovery generation ID", ErrVNextBackup)
		}
		seenIDs[projected.GenerationID] = struct{}{}
		if digestBytes([]byte(projected.CatalogJSON)) != projected.CatalogCanonicalSHA256 ||
			digestBytes([]byte(projected.GraphSourceRegistryJSON)) != projected.GraphSourceRegistrySHA256 ||
			digestBytes([]byte(projected.GraphImplementationBindingJSON)) != projected.GraphImplementationBindingSHA256 {
			return nil, fmt.Errorf("%w: Recovery generation canonical artifact digest mismatch", ErrVNextBackup)
		}
		catalog, err := recoverystate.NewFrozenCatalogJSON(
			[]byte(projected.CatalogJSON),
			recoverystate.FrozenShape{
				ContributionCount: projected.ContributionCount, AuthoredTableCount: projected.AuthoredTableCount,
				RequiredTableCount: projected.RequiredTableCount, ObjectFamilyCount: projected.ObjectFamilyCount,
			},
		)
		if err != nil {
			return nil, fmt.Errorf("%w: generation %s: %v", ErrVNextBackup, projected.GenerationID, err)
		}
		if catalog.DigestSHA256() != projected.CatalogDigestSHA256 {
			return nil, fmt.Errorf("%w: generation %s catalog digest mismatch", ErrVNextBackup, projected.GenerationID)
		}
		codecSchemaIDs := append([]string(nil), projected.CodecSchemaIDs...)
		requiredCodecs := contractrecovery.RequiredCodecSchemaIDs()
		emittedCodecs := []string{BackupArtifactEnvelopeSchemaID, BackupIntegrityManifestV4SchemaID,
			PostgresSnapshotArtifactV3SchemaID, PostgresSnapshotUnitV1SchemaID, PostgresBindingProofV1SchemaID,
			ObjectStoreBackupManifestV2SchemaID, ObjectStoreBackupSummaryV2SchemaID, RecoveryJournalPayloadSchemaID,
			projected.GraphSourceRegistrySchemaID, projected.GraphImplementationBindingSchemaID}
		slices.Sort(emittedCodecs)
		if !slices.Equal(codecSchemaIDs, requiredCodecs) || !slices.Equal(codecSchemaIDs, emittedCodecs) {
			return nil, fmt.Errorf("%w: generation codec membership differs from required emitted formats", ErrVNextBackup)
		}
		if err := validatePackagedRecoveryCodecs(requiredCodecs, contractrecovery.Artifacts); err != nil {
			return nil, err
		}
		if !sort.StringsAreSorted(codecSchemaIDs) || hasDuplicateStrings(codecSchemaIDs) ||
			vNextCodecRegistrySHA256(codecSchemaIDs) != projected.CodecRegistrySHA256 {
			return nil, fmt.Errorf("%w: generation %s codec registry mismatch", ErrVNextBackup, projected.GenerationID)
		}
		var binding struct {
			SchemaID                   string `json:"schema_id"`
			AlgorithmID                string `json:"algorithm_id"`
			RecoveryStateCatalogSHA256 string `json:"recovery_state_catalog_sha256"`
			SourceRegistrySHA256       string `json:"source_registry_sha256"`
		}
		if err := json.Unmarshal([]byte(projected.GraphImplementationBindingJSON), &binding); err != nil ||
			binding.SchemaID != projected.GraphImplementationBindingSchemaID ||
			binding.AlgorithmID != projected.GraphAlgorithmID ||
			binding.RecoveryStateCatalogSHA256 != projected.CatalogDigestSHA256 ||
			binding.SourceRegistrySHA256 != projected.GraphSourceRegistrySHA256 {
			return nil, fmt.Errorf("%w: generation %s Graph implementation binding mismatch", ErrVNextBackup, projected.GenerationID)
		}
		var sourceRegistry struct {
			SchemaID string `json:"schema_id"`
		}
		if err := json.Unmarshal([]byte(projected.GraphSourceRegistryJSON), &sourceRegistry); err != nil ||
			sourceRegistry.SchemaID != projected.GraphSourceRegistrySchemaID {
			return nil, fmt.Errorf("%w: generation %s Graph source registry mismatch", ErrVNextBackup, projected.GenerationID)
		}
		generation := &vNextRecoveryGeneration{
			id: projected.GenerationID, captureCurrent: projected.CaptureCurrent,
			stateCatalog: catalog, codecSchemaIDs: codecSchemaIDs,
			codecRegistrySHA256: projected.CodecRegistrySHA256,
			registrySchemaID:    projected.GraphSourceRegistrySchemaID,
			bindingSchemaID:     projected.GraphImplementationBindingSchemaID,
			graph: VNextGraphProjectionRestoreArtifacts{
				AlgorithmID: projected.GraphAlgorithmID, RecoveryStateCatalogSHA256: projected.CatalogDigestSHA256,
				SourceRegistryJSON:          []byte(projected.GraphSourceRegistryJSON),
				SourceRegistrySHA256:        projected.GraphSourceRegistrySHA256,
				ImplementationBindingJSON:   []byte(projected.GraphImplementationBindingJSON),
				ImplementationBindingSHA256: projected.GraphImplementationBindingSHA256,
			},
		}
		pair := vNextRecoveryGenerationPair(projected.CatalogDigestSHA256, projected.CodecRegistrySHA256)
		if _, duplicate := registry.byPair[pair]; duplicate {
			return nil, fmt.Errorf("%w: duplicate Recovery generation lookup pair", ErrVNextBackup)
		}
		registry.byPair[pair] = generation
		if generation.captureCurrent {
			if registry.current != nil {
				return nil, fmt.Errorf("%w: multiple current Recovery generations", ErrVNextBackup)
			}
			registry.current = generation
		}
	}
	if registry.current == nil {
		return nil, fmt.Errorf("%w: current Recovery generation is missing", ErrVNextBackup)
	}
	return registry, nil
}

func (registry *vNextRecoveryGenerationRegistry) lookup(
	catalogSHA256 string,
	codecRegistrySHA256 string,
) (*vNextRecoveryGeneration, bool) {
	if registry == nil {
		return nil, false
	}
	generation, ok := registry.byPair[vNextRecoveryGenerationPair(catalogSHA256, codecRegistrySHA256)]
	return generation, ok
}

func vNextRecoveryGenerationPair(catalogSHA256 string, codecRegistrySHA256 string) string {
	return catalogSHA256 + "\x00" + codecRegistrySHA256
}

func vNextCodecRegistrySHA256(schemaIDs []string) string {
	sum := sha256.Sum256([]byte(vNextCodecRegistryDomain + strings.Join(schemaIDs, "\n") + "\n"))
	return hex.EncodeToString(sum[:])
}

func digestBytes(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func hasDuplicateStrings(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1] == values[index] {
			return true
		}
	}
	return false
}

// Resolve the selected codecs against the packaged schema bytes, independently
// of the generation's self-consistent ID list and digest.
func validatePackagedRecoveryCodecs(required []string, artifacts []contractrecovery.Artifact) error {
	resolved := make(map[string]bool, len(required))
	for _, artifact := range artifacts {
		if !strings.HasSuffix(artifact.Path, ".schema.json") {
			continue
		}
		var schema struct {
			ID string `json:"$id"`
		}
		if err := json.Unmarshal([]byte(artifact.JSON), &schema); err != nil {
			return fmt.Errorf("%w: malformed packaged Recovery schema", ErrVNextBackup)
		}
		if !slices.Contains(required, schema.ID) {
			continue
		}
		if resolved[schema.ID] || digestBytes([]byte(artifact.JSON)) != artifact.SHA256 {
			return fmt.Errorf("%w: packaged Recovery codec integrity mismatch", ErrVNextBackup)
		}
		resolved[schema.ID] = true
	}
	if len(resolved) != len(required) {
		return fmt.Errorf("%w: required Recovery codec is not packaged", ErrVNextBackup)
	}
	return nil
}
