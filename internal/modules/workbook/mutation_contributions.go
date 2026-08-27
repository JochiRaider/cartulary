package workbook

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type CreateCommand struct {
	Actor        authn.UserRecord
	IncidentID   uuid.UUID
	ViewSchemaID string
	RequestID    string
	Now          time.Time
}

type CreateOperation struct {
	execute func(context.Context, CreateCommand) (MutationOutcome, error)
}

func (operation CreateOperation) Execute(ctx context.Context, command CreateCommand) (MutationOutcome, error) {
	if operation.execute == nil {
		return MutationOutcome{}, errors.New("workbook create operation is not initialized")
	}
	return operation.execute(ctx, command)
}

type CreateProvider interface {
	DecodeCreate(io.Reader) (CreateOperation, *MutationFailure, error)
}

type neutralCreateProvider[T any] struct {
	decode  func(io.Reader) (T, bool, *MutationFailure, error)
	execute func(context.Context, CreateCommand, T) (MutationOutcome, error)
}

func NewCreateProvider[T any](
	decode func(io.Reader) (T, bool, *MutationFailure, error),
	execute func(context.Context, CreateCommand, T) (MutationOutcome, error),
) (CreateProvider, error) {
	if decode == nil || execute == nil {
		return nil, errors.New("create provider requires decode and create functions")
	}
	return &neutralCreateProvider[T]{decode: decode, execute: execute}, nil
}

func (provider *neutralCreateProvider[T]) DecodeCreate(reader io.Reader) (CreateOperation, *MutationFailure, error) {
	value, present, failure, err := provider.decode(reader)
	if validationErr := validateDecodedMutationValue("create", value, present, failure, err); validationErr != nil {
		return CreateOperation{}, nil, validationErr
	}
	if failure != nil {
		return CreateOperation{}, failure, nil
	}
	return CreateOperation{execute: func(ctx context.Context, command CreateCommand) (MutationOutcome, error) {
		return provider.execute(ctx, command, value)
	}}, nil, nil
}

type PatchCommand struct {
	Actor                   authn.UserRecord
	RecordID                uuid.UUID
	AuthoritativeRecordType string
	RequestID               string
	Now                     time.Time
}

type PatchOperation struct {
	viewSchemaID string
	execute      func(context.Context, PatchCommand) (MutationOutcome, error)
}

func (operation PatchOperation) AdmittedViewSchemaID() string {
	return operation.viewSchemaID
}

func (operation PatchOperation) Execute(ctx context.Context, command PatchCommand) (MutationOutcome, error) {
	if operation.execute == nil {
		return MutationOutcome{}, errors.New("workbook patch operation is not initialized")
	}
	return operation.execute(ctx, command)
}

type PatchProvider interface {
	DecodePatch(io.Reader) (PatchOperation, *MutationFailure, error)
}

type neutralPatchProvider[T any] struct {
	decode       func(io.Reader) (T, bool, *MutationFailure, error)
	viewSchemaID func(T) string
	execute      func(context.Context, PatchCommand, T) (MutationOutcome, error)
}

func NewPatchProvider[T any](
	decode func(io.Reader) (T, bool, *MutationFailure, error),
	viewSchemaID func(T) string,
	execute func(context.Context, PatchCommand, T) (MutationOutcome, error),
) (PatchProvider, error) {
	if decode == nil || viewSchemaID == nil || execute == nil {
		return nil, errors.New("patch provider requires decode, view-schema, and patch functions")
	}
	return &neutralPatchProvider[T]{decode: decode, viewSchemaID: viewSchemaID, execute: execute}, nil
}

func (provider *neutralPatchProvider[T]) DecodePatch(reader io.Reader) (PatchOperation, *MutationFailure, error) {
	value, present, failure, err := provider.decode(reader)
	if validationErr := validateDecodedMutationValue("patch", value, present, failure, err); validationErr != nil {
		return PatchOperation{}, nil, validationErr
	}
	if failure != nil {
		return PatchOperation{}, failure, nil
	}
	viewSchemaID := provider.viewSchemaID(value)
	if viewSchemaID == "" {
		return PatchOperation{}, nil, errors.New("workbook patch decoder returned an empty view_schema_id")
	}
	return PatchOperation{
		viewSchemaID: viewSchemaID,
		execute: func(ctx context.Context, command PatchCommand) (MutationOutcome, error) {
			return provider.execute(ctx, command, value)
		},
	}, nil, nil
}

type ConflictCommand struct {
	Actor                   authn.UserRecord
	RecordID                uuid.UUID
	AuthoritativeRecordType string
	Claims                  ConflictClaims
	RequestID               string
	Now                     time.Time
}

type ConflictOperation struct {
	execute func(context.Context, ConflictCommand) (MutationOutcome, error)
}

func (operation ConflictOperation) Execute(ctx context.Context, command ConflictCommand) (MutationOutcome, error) {
	if operation.execute == nil {
		return MutationOutcome{}, errors.New("workbook conflict operation is not initialized")
	}
	return operation.execute(ctx, command)
}

type ConflictProvider interface {
	DecodeConflict(io.Reader, string, ConflictClaims) (ConflictOperation, *MutationFailure, error)
}

type neutralConflictProvider[T any] struct {
	decode  func(io.Reader, string, ConflictClaims) (T, bool, *MutationFailure, error)
	execute func(context.Context, ConflictCommand, T) (MutationOutcome, error)
}

func NewConflictProvider[T any](
	decode func(io.Reader, string, ConflictClaims) (T, bool, *MutationFailure, error),
	execute func(context.Context, ConflictCommand, T) (MutationOutcome, error),
) (ConflictProvider, error) {
	if decode == nil || execute == nil {
		return nil, errors.New("conflict provider requires decode and resolve functions")
	}
	return &neutralConflictProvider[T]{decode: decode, execute: execute}, nil
}

func (provider *neutralConflictProvider[T]) DecodeConflict(
	reader io.Reader,
	token string,
	claims ConflictClaims,
) (ConflictOperation, *MutationFailure, error) {
	value, present, failure, err := provider.decode(reader, token, claims)
	if validationErr := validateDecodedMutationValue("conflict", value, present, failure, err); validationErr != nil {
		return ConflictOperation{}, nil, validationErr
	}
	if failure != nil {
		return ConflictOperation{}, failure, nil
	}
	return ConflictOperation{execute: func(ctx context.Context, command ConflictCommand) (MutationOutcome, error) {
		return provider.execute(ctx, command, value)
	}}, nil, nil
}

func validateDecodedMutationValue[T any](
	family string,
	value T,
	present bool,
	failure *MutationFailure,
	err error,
) error {
	if err != nil {
		if present || failure != nil {
			return fmt.Errorf("workbook %s decoder returned contradictory value, failure, and error state", family)
		}
		return err
	}
	if failure != nil {
		if present {
			return fmt.Errorf("workbook %s decoder returned both an admitted value and failure", family)
		}
		return nil
	}
	if !present || isNilContributionProvider(value) {
		return fmt.Errorf("workbook %s decoder returned no admitted value or failure", family)
	}
	return nil
}
