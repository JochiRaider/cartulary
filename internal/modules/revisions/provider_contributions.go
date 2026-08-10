package revisions

import (
	"errors"
	"fmt"
	"sort"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type SourceOwnerModule string

const (
	SourceOwnerArtifacts      SourceOwnerModule = "artifacts"
	SourceOwnerAssessments    SourceOwnerModule = "assessments"
	SourceOwnerEntities       SourceOwnerModule = "entities"
	SourceOwnerEvidence       SourceOwnerModule = "evidence"
	SourceOwnerIndicators     SourceOwnerModule = "indicators"
	SourceOwnerLinks          SourceOwnerModule = "links"
	SourceOwnerParties        SourceOwnerModule = "parties"
	SourceOwnerTasksDecisions SourceOwnerModule = "tasksdecisions"
	SourceOwnerTimeline       SourceOwnerModule = "timeline"
)

var (
	ErrDuplicateProviderContribution  = errors.New("revisions: duplicate provider contribution")
	ErrMissingProviderContribution    = errors.New("revisions: missing provider contribution")
	ErrUnexpectedProviderContribution = errors.New("revisions: unexpected provider contribution")
)

type ProviderContribution struct {
	SourceOwnerModule SourceOwnerModule
	Records           []RecordProviderContribution
	NonRowTargets     []NonRowProviderContribution
}

type RecordVariant struct {
	Kind  string
	Value string
}

type RecordViewRouteContribution struct {
	ContributionID string
	Variant        *RecordVariant
	ViewSchemaIDs  []string
}

type RecordProviderContribution struct {
	SourceOwnerModule SourceOwnerModule
	RecordType        string
	SnapshotSchemaID  string
	// HistoryTargetKinds are source-owner-declared change-set mutation target
	// kinds that resolve to this record provider. An empty set admits the
	// record type itself. The generic "record" target remains a Revisions
	// envelope target and is not repeated by source owners.
	HistoryTargetKinds  []string
	DeleteRestoreSource deleterestorecontract.DeleteRestoreSource
	RowRollbackProvider rollbackcontract.RowSourceProvider
	RecordViewRoutes    []RecordViewRouteContribution
}

type NonRowProviderContribution struct {
	SourceOwnerModule SourceOwnerModule
	TargetKind        string
	HistoryFacet      HistoryFacet
	RollbackProvider  rollbackcontract.NonRowTargetProvider
}

func buildDeleteRestoreSourceCatalog(contributions []ProviderContribution) (*DeleteRestoreSourceCatalog, error) {
	snapshotRequirements, err := currentSnapshotSchemaRequirements()
	if err != nil {
		return nil, err
	}
	targetRequirements, err := currentTargetSemanticsRequirements()
	if err != nil {
		return nil, err
	}
	return buildDeleteRestoreSourceCatalogForRequirements(contributions, snapshotRequirements, targetRequirements)
}

func buildDeleteRestoreSourceCatalogForRequirements(
	contributions []ProviderContribution,
	snapshotRequirements []snapshotSchemaRequirement,
	targetRequirements []targetSemanticsRequirement,
) (*DeleteRestoreSourceCatalog, error) {
	requiredRecords := make(map[string]snapshotSchemaRequirement, len(snapshotRequirements))
	requiredNonRows := make(map[string]targetSemanticsRequirement)
	requiredOwners := map[SourceOwnerModule]struct{}{}
	for _, requirement := range snapshotRequirements {
		if requirement.RecordType == "" || requirement.SourceOwner == "" || requirement.SnapshotSchemaID == "" {
			return nil, fmt.Errorf("%w: invalid snapshot requirement", ErrUnexpectedProviderContribution)
		}
		if _, duplicate := requiredRecords[requirement.RecordType]; duplicate {
			return nil, fmt.Errorf("%w: record type %q", ErrDuplicateProviderContribution, requirement.RecordType)
		}
		requiredRecords[requirement.RecordType] = requirement
		requiredOwners[requirement.SourceOwner] = struct{}{}
	}
	for _, requirement := range targetRequirements {
		if requirement.DispatchClass != rollbackcontract.DispatchNonRow {
			continue
		}
		owner := SourceOwnerModule(requirement.SourceOwnerID)
		if requirement.TargetKind == "" || owner == "" {
			return nil, fmt.Errorf("%w: invalid non-row target requirement", ErrUnexpectedProviderContribution)
		}
		if _, duplicate := requiredNonRows[requirement.TargetKind]; duplicate {
			return nil, fmt.Errorf("%w: target kind %q", ErrDuplicateProviderContribution, requirement.TargetKind)
		}
		requiredNonRows[requirement.TargetKind] = requirement
		requiredOwners[owner] = struct{}{}
	}

	seenOwners := map[SourceOwnerModule]struct{}{}
	seenRecords := map[string]struct{}{}
	seenNonRows := map[string]struct{}{}
	deleteRestore := make([]DeleteRestoreSourceRegistration, 0, len(requiredRecords))
	for _, contribution := range contributions {
		owner := contribution.SourceOwnerModule
		if _, required := requiredOwners[owner]; !required || owner == "" {
			return nil, fmt.Errorf("%w: source owner %q", ErrUnexpectedProviderContribution, owner)
		}
		if _, exists := seenOwners[owner]; exists {
			return nil, fmt.Errorf("%w: source owner %q", ErrDuplicateProviderContribution, owner)
		}
		seenOwners[owner] = struct{}{}

		for _, record := range contribution.Records {
			if record.SourceOwnerModule != owner {
				return nil, fmt.Errorf("%w: record type %q declares owner %q in module contribution %q", ErrUnexpectedProviderContribution, record.RecordType, record.SourceOwnerModule, owner)
			}
			requirement, required := requiredRecords[record.RecordType]
			if !required || requirement.SourceOwner != owner || requirement.SnapshotSchemaID != record.SnapshotSchemaID {
				return nil, fmt.Errorf("%w: record type %q owned by %q", ErrUnexpectedProviderContribution, record.RecordType, owner)
			}
			if _, duplicate := seenRecords[record.RecordType]; duplicate {
				return nil, fmt.Errorf("%w: record type %q", ErrDuplicateProviderContribution, record.RecordType)
			}
			seenRecords[record.RecordType] = struct{}{}
			seenTargetKinds := map[string]struct{}{}
			for _, targetKind := range record.HistoryTargetKinds {
				if targetKind == "" {
					return nil, fmt.Errorf("%w: record type %q has an empty history target kind", ErrUnexpectedProviderContribution, record.RecordType)
				}
				if _, duplicate := seenTargetKinds[targetKind]; duplicate {
					return nil, fmt.Errorf("%w: record type %q repeats history target kind %q", ErrDuplicateProviderContribution, record.RecordType, targetKind)
				}
				seenTargetKinds[targetKind] = struct{}{}
			}
			deleteRestore = append(deleteRestore, DeleteRestoreSourceRegistration{RecordType: record.RecordType, Source: record.DeleteRestoreSource})
		}
		for _, target := range contribution.NonRowTargets {
			if target.SourceOwnerModule != owner {
				return nil, fmt.Errorf("%w: target kind %q declares owner %q in module contribution %q", ErrUnexpectedProviderContribution, target.TargetKind, target.SourceOwnerModule, owner)
			}
			requirement, required := requiredNonRows[target.TargetKind]
			if !required || SourceOwnerModule(requirement.SourceOwnerID) != owner {
				return nil, fmt.Errorf("%w: target kind %q owned by %q", ErrUnexpectedProviderContribution, target.TargetKind, owner)
			}
			if _, duplicate := seenNonRows[target.TargetKind]; duplicate {
				return nil, fmt.Errorf("%w: target kind %q", ErrDuplicateProviderContribution, target.TargetKind)
			}
			seenNonRows[target.TargetKind] = struct{}{}
		}
	}

	missing := make([]string, 0)
	for owner := range requiredOwners {
		if _, exists := seenOwners[owner]; !exists {
			missing = append(missing, "source-owner:"+string(owner))
		}
	}
	for recordType := range requiredRecords {
		if _, exists := seenRecords[recordType]; !exists {
			missing = append(missing, "record-type:"+recordType)
		}
	}
	for targetKind := range requiredNonRows {
		if _, exists := seenNonRows[targetKind]; !exists {
			missing = append(missing, "target-kind:"+targetKind)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: %v", ErrMissingProviderContribution, missing)
	}

	recordTypes := make([]string, 0, len(requiredRecords))
	for recordType := range requiredRecords {
		recordTypes = append(recordTypes, recordType)
	}
	sort.Strings(recordTypes)
	deleteRestoreCatalog, err := NewDeleteRestoreSourceCatalog(recordTypes, deleteRestore...)
	if err != nil {
		return nil, fmt.Errorf("build delete/restore provider catalog: %w", err)
	}
	return deleteRestoreCatalog, nil
}

func ValidateProviderContributions(contributions []ProviderContribution) error {
	if _, err := buildDeleteRestoreSourceCatalog(contributions); err != nil {
		return err
	}
	_, err := NewTargetSemanticsCatalog(contributions)
	return err
}

// NewDeleteRestoreSourceCatalogFromContributions projects the source-owner
// snapshot adapters needed by delete/restore and live-material reconstruction.
// Rollback dispatch remains exclusively in TargetSemanticsCatalog.
func NewDeleteRestoreSourceCatalogFromContributions(contributions []ProviderContribution) (*DeleteRestoreSourceCatalog, error) {
	return buildDeleteRestoreSourceCatalog(contributions)
}
