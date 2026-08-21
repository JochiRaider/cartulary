package workbook

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type CreateAdmission interface {
	ClientTransactionID() string
}

type CreateCommand struct {
	Actor        authn.UserRecord
	IncidentID   uuid.UUID
	ViewSchemaID string
	Admission    CreateAdmission
	RequestHash  []byte
	RequestID    string
	Now          time.Time
}

type CreateProvider interface {
	ValidateWorkbookContribution() error
	DecodeCreate(io.Reader) (CreateAdmission, *MutationFailure, error)
	Create(context.Context, CreateCommand) (MutationOutcome, error)
}

type neutralCreateProvider struct {
	decode  func(io.Reader) (CreateAdmission, *MutationFailure, error)
	execute func(context.Context, CreateCommand) (MutationOutcome, error)
}

func NewCreateProvider(
	decode func(io.Reader) (CreateAdmission, *MutationFailure, error),
	execute func(context.Context, CreateCommand) (MutationOutcome, error),
) (CreateProvider, error) {
	provider := &neutralCreateProvider{decode: decode, execute: execute}
	if err := provider.ValidateWorkbookContribution(); err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *neutralCreateProvider) ValidateWorkbookContribution() error {
	if provider == nil || provider.decode == nil || provider.execute == nil {
		return errors.New("create provider requires decode and create functions")
	}
	return nil
}

func (provider *neutralCreateProvider) DecodeCreate(reader io.Reader) (CreateAdmission, *MutationFailure, error) {
	return provider.decode(reader)
}

func (provider *neutralCreateProvider) Create(ctx context.Context, command CreateCommand) (MutationOutcome, error) {
	return provider.execute(ctx, command)
}

type PatchAdmission interface {
	ClientTransactionID() string
	AdmittedViewSchemaID() string
	AdmittedBaseRowVersion() int64
}

type PatchCommand struct {
	Actor                   authn.UserRecord
	RecordID                uuid.UUID
	AuthoritativeRecordType string
	Admission               PatchAdmission
	RequestHash             []byte
	RequestID               string
	Now                     time.Time
}

type PatchProvider interface {
	ValidateWorkbookContribution() error
	DecodePatch(io.Reader) (PatchAdmission, *MutationFailure, error)
	Patch(context.Context, PatchCommand) (MutationOutcome, error)
}

type neutralPatchProvider struct {
	decode  func(io.Reader) (PatchAdmission, *MutationFailure, error)
	execute func(context.Context, PatchCommand) (MutationOutcome, error)
}

func NewPatchProvider(
	decode func(io.Reader) (PatchAdmission, *MutationFailure, error),
	execute func(context.Context, PatchCommand) (MutationOutcome, error),
) (PatchProvider, error) {
	provider := &neutralPatchProvider{decode: decode, execute: execute}
	if err := provider.ValidateWorkbookContribution(); err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *neutralPatchProvider) ValidateWorkbookContribution() error {
	if provider == nil || provider.decode == nil || provider.execute == nil {
		return errors.New("patch provider requires decode and patch functions")
	}
	return nil
}

func (provider *neutralPatchProvider) DecodePatch(reader io.Reader) (PatchAdmission, *MutationFailure, error) {
	return provider.decode(reader)
}

func (provider *neutralPatchProvider) Patch(ctx context.Context, command PatchCommand) (MutationOutcome, error) {
	return provider.execute(ctx, command)
}

type ConflictAdmission interface {
	ClientTransactionID() string
}

type ConflictCommand struct {
	Actor                   authn.UserRecord
	RecordID                uuid.UUID
	AuthoritativeRecordType string
	Claims                  ConflictClaims
	Admission               ConflictAdmission
	RequestHash             []byte
	RequestID               string
	Now                     time.Time
}

type ConflictProvider interface {
	ValidateWorkbookContribution() error
	DecodeConflict(
		io.Reader,
		string,
		ConflictClaims,
	) (ConflictAdmission, *MutationFailure, error)
	ResolveConflict(context.Context, ConflictCommand) (MutationOutcome, error)
}

type neutralConflictProvider struct {
	decode  func(io.Reader, string, ConflictClaims) (ConflictAdmission, *MutationFailure, error)
	execute func(context.Context, ConflictCommand) (MutationOutcome, error)
}

func NewConflictProvider(
	decode func(io.Reader, string, ConflictClaims) (ConflictAdmission, *MutationFailure, error),
	execute func(context.Context, ConflictCommand) (MutationOutcome, error),
) (ConflictProvider, error) {
	provider := &neutralConflictProvider{decode: decode, execute: execute}
	if err := provider.ValidateWorkbookContribution(); err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *neutralConflictProvider) ValidateWorkbookContribution() error {
	if provider == nil || provider.decode == nil || provider.execute == nil {
		return errors.New("conflict provider requires decode and resolve functions")
	}
	return nil
}

func (provider *neutralConflictProvider) DecodeConflict(
	reader io.Reader,
	token string,
	claims ConflictClaims,
) (ConflictAdmission, *MutationFailure, error) {
	return provider.decode(reader, token, claims)
}

func (provider *neutralConflictProvider) ResolveConflict(
	ctx context.Context,
	command ConflictCommand,
) (MutationOutcome, error) {
	return provider.execute(ctx, command)
}
