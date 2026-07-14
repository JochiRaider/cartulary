package revisions

import (
	"errors"
	"fmt"
	"sort"

	recorddeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"
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

type RecordProviderContribution struct {
	SourceOwnerModule     SourceOwnerModule
	RecordType            string
	DeleteRestoreProvider recorddeleterestore.SourceProvider
	RowRollbackProvider   rollbackcontract.RowSourceProvider
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

func buildProviderCatalogs(contributions []ProviderContribution) (*DeleteRestoreProviderCatalog, *RowProviderCatalog, *NonRowProviderCatalog, error) {
	requiredOwners := map[SourceOwnerModule]struct{}{}
	for _, owner := range currentRecordProviderOwners {
		requiredOwners[owner] = struct{}{}
	}
	for _, owner := range currentNonRowProviderOwners {
		requiredOwners[owner] = struct{}{}
	}

	seenOwners := map[SourceOwnerModule]struct{}{}
	deleteRestore := make([]DeleteRestoreProviderRegistration, 0, len(currentRecordProviderOwners))
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
			deleteRestore = append(deleteRestore, DeleteRestoreProviderRegistration{RecordType: record.RecordType, Provider: record.DeleteRestoreProvider})
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
	deleteRestoreCatalog, err := NewDeleteRestoreProviderCatalog(recordTypes, deleteRestore...)
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

func sortedProviderRequirementKeys[T ~string](requirements map[string]T) []string {
	keys := make([]string, 0, len(requirements))
	for key := range requirements {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
