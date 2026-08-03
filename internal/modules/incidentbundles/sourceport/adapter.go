package sourceport

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type ExportFunc func(context.Context, ExportContext) ([]incidentportability.File, error)
type PrepareFunc func(context.Context, Bundle, ImportContext) (any, error)
type ApplyFunc func(context.Context, pgx.Tx, any, ImportContext) error
type ValidateFunc func(context.Context, pgx.Tx, any, ImportContext) error
type ContractValidateFunc func() error

type AdapterOptions struct {
	Descriptor       Descriptor
	Export           ExportFunc
	Prepare          PrepareFunc
	Apply            ApplyFunc
	Validate         ValidateFunc
	ValidateContract ContractValidateFunc
}

type Adapter struct {
	descriptor       Descriptor
	export           ExportFunc
	prepare          PrepareFunc
	apply            ApplyFunc
	validate         ValidateFunc
	validateContract ContractValidateFunc
	portKey          string
}

func NewAdapter(options AdapterOptions) *Adapter {
	return &Adapter{
		descriptor:       cloneDescriptor(options.Descriptor),
		export:           options.Export,
		prepare:          options.Prepare,
		apply:            options.Apply,
		validate:         options.Validate,
		validateContract: options.ValidateContract,
		portKey:          options.Descriptor.OwnerID + ":" + options.Descriptor.FamilyID,
	}
}

func (a *Adapter) Descriptor() Descriptor {
	if a == nil {
		return Descriptor{}
	}
	return cloneDescriptor(a.descriptor)
}

func (a *Adapter) ValidateSourcePortContract() error {
	if a == nil || a.export == nil || a.prepare == nil || a.apply == nil || a.validate == nil {
		return fmt.Errorf("%w: incomplete source port implementation", ErrInvalidCatalog)
	}
	if a.validateContract != nil {
		if err := a.validateContract(); err != nil {
			return fmt.Errorf("%w: source port contract rejected", ErrInvalidCatalog)
		}
	}
	return nil
}

func (a *Adapter) Export(ctx context.Context, exportContext ExportContext) ([]incidentportability.File, error) {
	if a == nil || a.export == nil || exportContext.Query == nil || exportContext.IncidentID == uuid.Nil {
		return nil, fmt.Errorf("%w: missing export", ErrInvalidCatalog)
	}
	return a.export(ctx, exportContext)
}

func QueryExport(
	export func(context.Context, incidentportability.Queryer, uuid.UUID) ([]incidentportability.File, error),
) ExportFunc {
	return func(ctx context.Context, exportContext ExportContext) ([]incidentportability.File, error) {
		return export(ctx, exportContext.Query, exportContext.IncidentID)
	}
}

func (a *Adapter) PrepareImport(ctx context.Context, bundle Bundle, importContext ImportContext) (Prepared, error) {
	if a == nil || a.prepare == nil || importContext.OperationID == "" {
		return Prepared{}, fmt.Errorf("%w: missing preparation", ErrInvalidCatalog)
	}
	value, err := a.prepare(ctx, bundle, importContext)
	if err != nil {
		return Prepared{}, err
	}
	return NewPrepared(a.portKey, importContext.OperationID, value), nil
}

func (a *Adapter) ApplyImportTx(ctx context.Context, tx pgx.Tx, prepared Prepared, importContext ImportContext) error {
	if a == nil || a.apply == nil {
		return fmt.Errorf("%w: missing apply", ErrInvalidCatalog)
	}
	value, err := prepared.ValueFor(a.portKey, importContext.OperationID)
	if err != nil {
		return err
	}
	err = a.apply(ctx, tx, value, importContext)
	if err == nil {
		return nil
	}
	var verification *incidentportability.VerificationFailure
	if errors.As(err, &verification) && verification.ReasonCode == "duplicate_source_row" {
		return a.sourceFailure()
	}
	return err
}

func (a *Adapter) sourceFailure() error {
	invariantID := ""
	if a != nil && len(a.descriptor.InvariantIDs) > 0 {
		invariantID = a.descriptor.InvariantIDs[0]
	}
	return &Failure{FamilyID: a.descriptor.FamilyID, InvariantID: invariantID}
}

func (a *Adapter) ValidateImportTx(ctx context.Context, tx pgx.Tx, prepared Prepared, importContext ImportContext) error {
	if a == nil || a.validate == nil {
		return fmt.Errorf("%w: missing validation", ErrInvalidCatalog)
	}
	value, err := prepared.ValueFor(a.portKey, importContext.OperationID)
	if err != nil {
		return err
	}
	return a.validate(ctx, tx, value, importContext)
}

type PreparedFiles map[string][]byte

func PrepareFiles(descriptor Descriptor, bundle Bundle, version int) (PreparedFiles, error) {
	prepared := PreparedFiles{}
	for _, path := range descriptor.Paths {
		if !containsVersion(path.Versions, version) {
			continue
		}
		payload, ok := bundle.File(path.LogicalPath)
		if !ok {
			return nil, &incidentportability.VerificationFailure{ReasonCode: "missing_required_file"}
		}
		prepared[path.LogicalPath] = append([]byte(nil), payload...)
		if path.ContentRole != "singleton_json" {
			rows, err := incidentportability.DecodeNDJSON(payload)
			if err != nil {
				return nil, err
			}
			seen := map[string]struct{}{}
			for _, row := range rows {
				identity := stableIdentity(row, path.StableIdentity)
				if identity == "" {
					return nil, &Failure{FamilyID: descriptor.FamilyID, InvariantID: descriptor.InvariantIDs[0]}
				}
				if _, duplicate := seen[identity]; duplicate {
					return nil, &Failure{FamilyID: descriptor.FamilyID, InvariantID: descriptor.InvariantIDs[0]}
				}
				seen[identity] = struct{}{}
			}
		}
	}
	return prepared, nil
}

func stableIdentity(row map[string]any, fields []string) string {
	identity := ""
	for _, field := range fields {
		value := incidentportability.StringFromAny(row[field])
		if value == "" {
			return ""
		}
		identity += "\x00" + value
	}
	return identity
}
