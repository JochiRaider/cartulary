package sourceport

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

type ExportFunc func(context.Context, ExportContext) error
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
	instance         *adapterInstance
	descriptor       Descriptor
	export           ExportFunc
	prepare          PrepareFunc
	apply            ApplyFunc
	validate         ValidateFunc
	validateContract ContractValidateFunc
}

// A non-zero-sized private token distinguishes constructions even when a caller
// overwrites an Adapter value at an existing address.
type adapterInstance struct{ marker byte }

func NewAdapter(options AdapterOptions) *Adapter {
	return &Adapter{
		instance:         &adapterInstance{},
		descriptor:       cloneDescriptor(options.Descriptor),
		export:           options.Export,
		prepare:          options.Prepare,
		apply:            options.Apply,
		validate:         options.Validate,
		validateContract: options.ValidateContract,
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

func (a *Adapter) Export(ctx context.Context, exportContext ExportContext) error {
	if a == nil || a.export == nil || exportContext.Query == nil || exportContext.Members == nil || exportContext.IncidentID == uuid.Nil {
		return fmt.Errorf("%w: missing export", ErrInvalidCatalog)
	}
	return a.export(ctx, exportContext)
}

func QueryExport(export func(context.Context, incidentportability.Queryer, uuid.UUID, incidentportability.MemberWriter) error) ExportFunc {
	return func(ctx context.Context, input ExportContext) error {
		return export(ctx, input.Query, input.IncidentID, input.Members)
	}
}

func (a *Adapter) PrepareImport(ctx context.Context, bundle Bundle, importContext ImportContext) (Prepared, error) {
	if a == nil || a.prepare == nil || importContext.OperationID == "" ||
		importContext.IncidentID == uuid.Nil || importContext.ActorUserID == uuid.Nil || a.descriptor.ContractMajor != ContractMajor {
		return Prepared{}, fmt.Errorf("%w: missing preparation", ErrInvalidCatalog)
	}
	admitted := false
	for _, path := range a.descriptor.Paths {
		admitted = admitted || containsVersion(path.Versions, importContext.BundleVersion)
	}
	if !admitted {
		return Prepared{}, fmt.Errorf("%w: unsupported source version", ErrInvalidCatalog)
	}
	value, err := a.prepare(ctx, bundle, importContext)
	if err != nil {
		return Prepared{}, err
	}
	return newPrepared(a, importContext, value), nil
}

func (a *Adapter) ApplyImportTx(ctx context.Context, tx pgx.Tx, prepared Prepared, importContext ImportContext) error {
	if a == nil || a.apply == nil {
		return fmt.Errorf("%w: missing apply", ErrInvalidCatalog)
	}
	value, err := prepared.valueFor(a, importContext)
	if err != nil {
		return err
	}
	err = a.apply(ctx, tx, value, importContext)
	if err == nil {
		return nil
	}
	if logicalPath, ok := incidentportability.FixedImportFailurePath(err); ok {
		return a.declaredPathFailure(logicalPath)
	}
	return err
}

func (a *Adapter) declaredPathFailure(logicalPath string) error {
	if a == nil {
		return ErrInvalidCatalog
	}
	for _, path := range a.descriptor.Paths {
		if path.LogicalPath == logicalPath {
			return a.descriptor.DeclaredFailure(path.StableIdentityInvariantID)
		}
	}
	return fmt.Errorf("%w: fixed import path is undeclared", ErrInvalidCatalog)
}

func (a *Adapter) ValidateImportTx(ctx context.Context, tx pgx.Tx, prepared Prepared, importContext ImportContext) error {
	if a == nil || a.validate == nil {
		return fmt.Errorf("%w: missing validation", ErrInvalidCatalog)
	}
	value, err := prepared.valueFor(a, importContext)
	if err != nil {
		return err
	}
	return a.validate(ctx, tx, value, importContext)
}

// PrepareMembers validates canonical source identities without retaining rows.
func PrepareMembers(ctx context.Context, descriptor Descriptor, bundle incidentportability.MemberReader, version int) (*incidentportability.BoundMembers, error) {
	paths := []string{}
	for _, path := range descriptor.Paths {
		if containsVersion(path.Versions, version) {
			paths = append(paths, path.LogicalPath)
		}
	}
	if len(paths) == 0 {
		return nil, ErrInvalidCatalog
	}
	prepared, err := incidentportability.BindMembers(bundle, paths)
	if err != nil {
		return nil, err
	}
	for _, path := range descriptor.Paths {
		if !containsVersion(path.Versions, version) || path.ContentRole == "singleton_json" {
			continue
		}
		var previous []any
		err := incidentportability.WalkPath(ctx, prepared, path.LogicalPath, func(row map[string]any) error {
			identity := make([]any, len(path.StableIdentity))
			for i, field := range path.StableIdentity {
				value, ok := row[field]
				if !ok || incidentportability.StringFromAny(value) == "" {
					return descriptor.DeclaredFailure(path.StableIdentityInvariantID)
				}
				identity[i] = value
			}
			if _, err := compareIdentity(identity, identity); err != nil {
				return descriptor.DeclaredFailure(path.StableIdentityInvariantID)
			}
			if previous != nil {
				order, err := compareIdentity(previous, identity)
				if err != nil || order >= 0 {
					return descriptor.DeclaredFailure(path.StableIdentityInvariantID)
				}
			}
			previous = identity
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return prepared, nil
}

// Identity ordering follows native integer order and UTF-8 string order.
// Stable identity fields never admit floating point or compound JSON values.
func compareIdentity(left, right []any) (int, error) {
	for i, a := range left {
		var order int
		switch value := a.(type) {
		case string:
			other, ok := right[i].(string)
			if !ok {
				return 0, ErrInvalidCatalog
			}
			order = strings.Compare(value, other)
		case json.Number:
			other, ok := right[i].(json.Number)
			if !ok {
				return 0, ErrInvalidCatalog
			}
			first, ok := new(big.Int).SetString(value.String(), 10)
			if !ok {
				return 0, ErrInvalidCatalog
			}
			second, ok := new(big.Int).SetString(other.String(), 10)
			if !ok {
				return 0, ErrInvalidCatalog
			}
			order = first.Cmp(second)
		default:
			return 0, ErrInvalidCatalog
		}
		if order != 0 {
			return order, nil
		}
	}
	return 0, nil
}
