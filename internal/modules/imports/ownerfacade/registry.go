package ownerfacade

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ImportOwnerCreateCommand struct {
	Request           ImportOwnerCreateRequest
	ChangeSetID       uuid.UUID
	SequenceNo        int
	MutationSequencer *ImportMutationSequencer
	Now               time.Time
}

type ImportOwnerCreateBinding struct {
	TargetViewSchemaID string
	FacadeID           string
}

type ImportOwnerCreateFacade interface {
	ImportOwnerCreateBinding() ImportOwnerCreateBinding
	NormalizeImportField(
		fieldKey string,
		raw string,
		emptyValuePolicy string,
	) (ImportScalarValue, bool, error)
	CreateImportRowTx(context.Context, pgx.Tx, ImportOwnerCreateCommand) (ImportOwnerCreateResponse, error)
}

// ImportOwnerCreateTx is the caller-transaction create capability an import
// owner contributes before it is bound to normalization and registry metadata.
type ImportOwnerCreateTx interface {
	CreateImportRowTx(context.Context, pgx.Tx, ImportOwnerCreateCommand) (ImportOwnerCreateResponse, error)
}

type ImportOwnerCreateFunc func(
	context.Context,
	pgx.Tx,
	ImportOwnerCreateCommand,
) (ImportOwnerCreateResponse, error)

type ImportOwnerNormalizeFunc func(
	fieldKey string,
	raw string,
	emptyValuePolicy string,
) (ImportScalarValue, bool, error)

type boundImportOwnerCreateFacade struct {
	binding   ImportOwnerCreateBinding
	normalize ImportOwnerNormalizeFunc
	create    ImportOwnerCreateFunc
}

func NewImportOwnerCreateFacade(
	binding ImportOwnerCreateBinding,
	create ImportOwnerCreateFunc,
) (ImportOwnerCreateFacade, error) {
	return NewImportOwnerCreateFacadeWithNormalizer(
		binding,
		func(
			fieldKey string,
			raw string,
			emptyValuePolicy string,
		) (ImportScalarValue, bool, error) {
			return NormalizeImportScalar(
				binding.TargetViewSchemaID,
				fieldKey,
				raw,
				emptyValuePolicy,
			)
		},
		create,
	)
}

func NewImportOwnerCreateFacadeWithNormalizer(
	binding ImportOwnerCreateBinding,
	normalize ImportOwnerNormalizeFunc,
	create ImportOwnerCreateFunc,
) (ImportOwnerCreateFacade, error) {
	if err := validateImportOwnerCreateBinding(binding); err != nil {
		return nil, err
	}
	if normalize == nil {
		return nil, fmt.Errorf(
			"import owner-create facade %s for %s requires a normalizer",
			binding.FacadeID,
			binding.TargetViewSchemaID,
		)
	}
	if create == nil {
		return nil, fmt.Errorf(
			"import owner-create facade %s for %s requires an implementation",
			binding.FacadeID,
			binding.TargetViewSchemaID,
		)
	}
	return &boundImportOwnerCreateFacade{
		binding:   binding,
		normalize: normalize,
		create:    create,
	}, nil
}

func (f *boundImportOwnerCreateFacade) ImportOwnerCreateBinding() ImportOwnerCreateBinding {
	return f.binding
}

func (f *boundImportOwnerCreateFacade) NormalizeImportField(
	fieldKey string,
	raw string,
	emptyValuePolicy string,
) (ImportScalarValue, bool, error) {
	return f.normalize(fieldKey, raw, emptyValuePolicy)
}

func (f *boundImportOwnerCreateFacade) CreateImportRowTx(
	ctx context.Context,
	tx pgx.Tx,
	command ImportOwnerCreateCommand,
) (ImportOwnerCreateResponse, error) {
	if command.Request.TargetViewSchemaID != f.binding.TargetViewSchemaID {
		return ImportOwnerCreateResponse{}, fmt.Errorf(
			"import owner-create facade %s is bound to %s, not %s",
			f.binding.FacadeID,
			f.binding.TargetViewSchemaID,
			command.Request.TargetViewSchemaID,
		)
	}
	return f.create(ctx, tx, command)
}

type ImportMutationSequencer struct {
	next int
}

func NewImportMutationSequencer() *ImportMutationSequencer {
	return &ImportMutationSequencer{next: 1}
}

func (s *ImportMutationSequencer) Allocate(count int) (int, error) {
	if s == nil {
		return 0, fmt.Errorf("import mutation sequencer is required")
	}
	if count < 1 {
		return 0, fmt.Errorf("import mutation sequence allocation must be positive")
	}
	first := s.next
	s.next += count
	return first, nil
}

func (c ImportOwnerCreateCommand) AllocateMutationSequence(count int) (int, error) {
	if c.MutationSequencer != nil {
		return c.MutationSequencer.Allocate(count)
	}
	if count == 1 && c.SequenceNo > 0 {
		return c.SequenceNo, nil
	}
	return 0, fmt.Errorf("import owner-create command requires a mutation sequencer")
}

type ImportOwnerCreateRegistry struct {
	byTarget map[string]ImportOwnerCreateFacade
	bindings []ImportOwnerCreateBinding
}

func NewImportOwnerCreateRegistry(
	expected []ImportOwnerCreateBinding,
	facades ...ImportOwnerCreateFacade,
) (*ImportOwnerCreateRegistry, error) {
	expectedByTarget := make(map[string]ImportOwnerCreateBinding, len(expected))
	for _, binding := range expected {
		if err := validateImportOwnerCreateBinding(binding); err != nil {
			return nil, fmt.Errorf("invalid expected import owner-create binding: %w", err)
		}
		if previous, exists := expectedByTarget[binding.TargetViewSchemaID]; exists {
			return nil, fmt.Errorf(
				"duplicate expected import owner-create binding for %s: %s and %s",
				binding.TargetViewSchemaID,
				previous.FacadeID,
				binding.FacadeID,
			)
		}
		expectedByTarget[binding.TargetViewSchemaID] = binding
	}
	if len(expectedByTarget) == 0 {
		return nil, fmt.Errorf("import owner-create registry requires at least one expected binding")
	}

	byTarget := make(map[string]ImportOwnerCreateFacade, len(facades))
	for _, facade := range facades {
		if facade == nil {
			return nil, fmt.Errorf("import owner-create registry received a nil facade")
		}
		binding := facade.ImportOwnerCreateBinding()
		if err := validateImportOwnerCreateBinding(binding); err != nil {
			return nil, fmt.Errorf("invalid import owner-create facade binding: %w", err)
		}
		expectedBinding, exists := expectedByTarget[binding.TargetViewSchemaID]
		if !exists {
			return nil, fmt.Errorf(
				"unexpected import owner-create facade %s for %s",
				binding.FacadeID,
				binding.TargetViewSchemaID,
			)
		}
		if expectedBinding.FacadeID != binding.FacadeID {
			return nil, fmt.Errorf(
				"import owner-create facade mismatch for %s: got %s, want %s",
				binding.TargetViewSchemaID,
				binding.FacadeID,
				expectedBinding.FacadeID,
			)
		}
		if previous, exists := byTarget[binding.TargetViewSchemaID]; exists {
			return nil, fmt.Errorf(
				"duplicate import owner-create facade for %s: %s and %s",
				binding.TargetViewSchemaID,
				previous.ImportOwnerCreateBinding().FacadeID,
				binding.FacadeID,
			)
		}
		byTarget[binding.TargetViewSchemaID] = facade
	}
	for targetViewSchemaID, binding := range expectedByTarget {
		if _, exists := byTarget[targetViewSchemaID]; !exists {
			return nil, fmt.Errorf(
				"missing import owner-create facade %s for %s",
				binding.FacadeID,
				targetViewSchemaID,
			)
		}
	}

	bindings := make([]ImportOwnerCreateBinding, 0, len(expectedByTarget))
	for _, binding := range expectedByTarget {
		bindings = append(bindings, binding)
	}
	sort.Slice(bindings, func(i, j int) bool {
		return bindings[i].TargetViewSchemaID < bindings[j].TargetViewSchemaID
	})
	return &ImportOwnerCreateRegistry{byTarget: byTarget, bindings: bindings}, nil
}

func (r *ImportOwnerCreateRegistry) Resolve(
	targetViewSchemaID string,
	facadeID string,
) (ImportOwnerCreateFacade, bool) {
	if r == nil {
		return nil, false
	}
	facade, exists := r.byTarget[targetViewSchemaID]
	if !exists || facade.ImportOwnerCreateBinding().FacadeID != facadeID {
		return nil, false
	}
	return facade, true
}

func (r *ImportOwnerCreateRegistry) Bindings() []ImportOwnerCreateBinding {
	if r == nil {
		return nil
	}
	return append([]ImportOwnerCreateBinding(nil), r.bindings...)
}

func validateImportOwnerCreateBinding(binding ImportOwnerCreateBinding) error {
	if strings.TrimSpace(binding.TargetViewSchemaID) == "" {
		return fmt.Errorf("target view schema id is required")
	}
	if strings.TrimSpace(binding.FacadeID) == "" {
		return fmt.Errorf("facade id is required for %s", binding.TargetViewSchemaID)
	}
	return nil
}
