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

type LiveRecordChangePolicy string

const (
	LiveRecordChangeRequired LiveRecordChangePolicy = "required"
	LiveRecordChangeNone     LiveRecordChangePolicy = "none"
)

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
	SourceOwnerModule      SourceOwnerModule
	RecordType             string
	DeleteRestoreSource    deleterestorecontract.DeleteRestoreSource
	RowRollbackProvider    rollbackcontract.RowSourceProvider
	LiveRecordChangePolicy LiveRecordChangePolicy
	RecordViewRoutes       []RecordViewRouteContribution
}

type NonRowProviderContribution struct {
	SourceOwnerModule SourceOwnerModule
	TargetKind        string
	RollbackProvider  rollbackcontract.NonRowTargetProvider
}

var currentRecordProviderOwners = map[string]SourceOwnerModule{
	"artifact":       SourceOwnerArtifacts,
	"assessment":     SourceOwnerAssessments,
	"decision":       SourceOwnerTasksDecisions,
	"evidence":       SourceOwnerEvidence,
	"host":           SourceOwnerEntities,
	"identity":       SourceOwnerEntities,
	"indicator":      SourceOwnerIndicators,
	"party":          SourceOwnerParties,
	"task_request":   SourceOwnerTasksDecisions,
	"timeline_event": SourceOwnerTimeline,
}

var currentNonRowProviderOwners = map[string]SourceOwnerModule{
	"entity_alias":                SourceOwnerEntities,
	"entity_mention":              SourceOwnerEntities,
	"entity_preserved_identifier": SourceOwnerEntities,
	"indicator_observation":       SourceOwnerIndicators,
	"indicator_state_interval":    SourceOwnerIndicators,
	"record_link":                 SourceOwnerLinks,
	"record_tag":                  SourceOwnerLinks,
}

func buildProviderCatalogs(contributions []ProviderContribution) (*DeleteRestoreSourceCatalog, *RowProviderCatalog, *NonRowProviderCatalog, error) {
	requiredOwners := map[SourceOwnerModule]struct{}{}
	for _, owner := range currentRecordProviderOwners {
		requiredOwners[owner] = struct{}{}
	}
	for _, owner := range currentNonRowProviderOwners {
		requiredOwners[owner] = struct{}{}
	}

	seenOwners := map[SourceOwnerModule]struct{}{}
	deleteRestore := make([]DeleteRestoreSourceRegistration, 0, len(currentRecordProviderOwners))
	rowRollback := make([]RowProviderRegistration, 0, len(currentRecordProviderOwners))
	nonRowRollback := make([]NonRowProviderRegistration, 0, len(currentNonRowProviderOwners))
	for _, contribution := range contributions {
		owner := contribution.SourceOwnerModule
		if _, required := requiredOwners[owner]; !required || owner == "" {
			return nil, nil, nil, fmt.Errorf("%w: source owner %q", ErrUnexpectedProviderContribution, owner)
		}
		if _, exists := seenOwners[owner]; exists {
			return nil, nil, nil, fmt.Errorf("%w: source owner %q", ErrDuplicateProviderContribution, owner)
		}
		seenOwners[owner] = struct{}{}

		for _, record := range contribution.Records {
			if record.SourceOwnerModule != owner {
				return nil, nil, nil, fmt.Errorf("%w: record type %q declares owner %q in module contribution %q", ErrUnexpectedProviderContribution, record.RecordType, record.SourceOwnerModule, owner)
			}
			expectedOwner, required := currentRecordProviderOwners[record.RecordType]
			if !required || expectedOwner != owner {
				return nil, nil, nil, fmt.Errorf("%w: record type %q owned by %q", ErrUnexpectedProviderContribution, record.RecordType, owner)
			}
			deleteRestore = append(deleteRestore, DeleteRestoreSourceRegistration{RecordType: record.RecordType, Source: record.DeleteRestoreSource})
			rowRollback = append(rowRollback, RowProviderRegistration{RecordType: record.RecordType, Provider: record.RowRollbackProvider})
		}
		for _, target := range contribution.NonRowTargets {
			if target.SourceOwnerModule != owner {
				return nil, nil, nil, fmt.Errorf("%w: target kind %q declares owner %q in module contribution %q", ErrUnexpectedProviderContribution, target.TargetKind, target.SourceOwnerModule, owner)
			}
			expectedOwner, required := currentNonRowProviderOwners[target.TargetKind]
			if !required || expectedOwner != owner {
				return nil, nil, nil, fmt.Errorf("%w: target kind %q owned by %q", ErrUnexpectedProviderContribution, target.TargetKind, owner)
			}
			nonRowRollback = append(nonRowRollback, NonRowProviderRegistration{TargetKind: target.TargetKind, Provider: target.RollbackProvider})
		}
	}

	missingOwners := make([]string, 0)
	for owner := range requiredOwners {
		if _, exists := seenOwners[owner]; !exists {
			missingOwners = append(missingOwners, string(owner))
		}
	}
	if len(missingOwners) > 0 {
		sort.Strings(missingOwners)
		return nil, nil, nil, fmt.Errorf("%w: source owners %v", ErrMissingProviderContribution, missingOwners)
	}

	recordTypes := sortedProviderRequirementKeys(currentRecordProviderOwners)
	targetKinds := sortedProviderRequirementKeys(currentNonRowProviderOwners)
	deleteRestoreCatalog, err := NewDeleteRestoreSourceCatalog(recordTypes, deleteRestore...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build delete/restore provider catalog: %w", err)
	}
	rowRollbackCatalog, err := NewRowProviderCatalog(recordTypes, rowRollback...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build row rollback provider catalog: %w", err)
	}
	nonRowRollbackCatalog, err := NewNonRowProviderCatalog(targetKinds, nonRowRollback...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("build non-row rollback provider catalog: %w", err)
	}
	return deleteRestoreCatalog, rowRollbackCatalog, nonRowRollbackCatalog, nil
}

func ValidateProviderContributions(contributions []ProviderContribution) error {
	_, _, _, err := buildProviderCatalogs(contributions)
	return err
}

func sortedProviderRequirementKeys[T ~string](requirements map[string]T) []string {
	keys := make([]string, 0, len(requirements))
	for key := range requirements {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
