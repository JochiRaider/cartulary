package workbook

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type ClipboardCommand struct {
	Actor        authn.UserRecord
	IncidentID   uuid.UUID
	ViewSchemaID string
	RequestID    string
	Now          time.Time
}

type ClipboardOperation struct {
	execute func(context.Context, ClipboardCommand) (MutationOutcome, error)
}

func (operation ClipboardOperation) Execute(ctx context.Context, command ClipboardCommand) (MutationOutcome, error) {
	if operation.execute == nil {
		return MutationOutcome{}, errors.New("workbook clipboard operation is not initialized")
	}
	return operation.execute(ctx, command)
}

type ClipboardProvider interface {
	DecodeClipboard(io.Reader) (ClipboardOperation, *MutationFailure, error)
}

type clipboardProvider[T any] struct {
	decode  func(io.Reader) (T, bool, *MutationFailure, error)
	execute func(context.Context, ClipboardCommand, T) (MutationOutcome, error)
}

func NewClipboardProvider[T any](
	decode func(io.Reader) (T, bool, *MutationFailure, error),
	execute func(context.Context, ClipboardCommand, T) (MutationOutcome, error),
) (ClipboardProvider, error) {
	if decode == nil || execute == nil {
		return nil, errors.New("clipboard provider requires decode and apply functions")
	}
	return &clipboardProvider[T]{decode: decode, execute: execute}, nil
}

func (provider *clipboardProvider[T]) DecodeClipboard(reader io.Reader) (ClipboardOperation, *MutationFailure, error) {
	value, present, failure, err := provider.decode(reader)
	if validationErr := validateDecodedMutationValue("clipboard", value, present, failure, err); validationErr != nil {
		return ClipboardOperation{}, nil, validationErr
	}
	if failure != nil {
		return ClipboardOperation{}, failure, nil
	}
	return ClipboardOperation{execute: func(ctx context.Context, command ClipboardCommand) (MutationOutcome, error) {
		return provider.execute(ctx, command, value)
	}}, nil, nil
}

type BulkCommand struct {
	Actor        authn.UserRecord
	IncidentID   uuid.UUID
	ViewSchemaID string
	RequestID    string
	Now          time.Time
}

type BulkOperation struct {
	execute func(context.Context, BulkCommand) (MutationOutcome, error)
}

func (operation BulkOperation) Execute(ctx context.Context, command BulkCommand) (MutationOutcome, error) {
	if operation.execute == nil {
		return MutationOutcome{}, errors.New("workbook bulk operation is not initialized")
	}
	return operation.execute(ctx, command)
}

type BulkProvider interface {
	DecodeBulk(io.Reader) (BulkOperation, *MutationFailure, error)
}

type bulkProvider[T any] struct {
	decode  func(io.Reader) (T, bool, *MutationFailure, error)
	execute func(context.Context, BulkCommand, T) (MutationOutcome, error)
}

func NewBulkProvider[T any](
	decode func(io.Reader) (T, bool, *MutationFailure, error),
	execute func(context.Context, BulkCommand, T) (MutationOutcome, error),
) (BulkProvider, error) {
	if decode == nil || execute == nil {
		return nil, errors.New("bulk provider requires decode and apply functions")
	}
	return &bulkProvider[T]{decode: decode, execute: execute}, nil
}

func (provider *bulkProvider[T]) DecodeBulk(reader io.Reader) (BulkOperation, *MutationFailure, error) {
	value, present, failure, err := provider.decode(reader)
	if validationErr := validateDecodedMutationValue("bulk", value, present, failure, err); validationErr != nil {
		return BulkOperation{}, nil, validationErr
	}
	if failure != nil {
		return BulkOperation{}, failure, nil
	}
	return BulkOperation{execute: func(ctx context.Context, command BulkCommand) (MutationOutcome, error) {
		return provider.execute(ctx, command, value)
	}}, nil, nil
}

type LinkedNoteCommand struct {
	Actor     authn.UserRecord
	Target    RecordTarget
	RequestID string
	Now       time.Time
}

type LinkedNoteOperation struct {
	execute func(context.Context, LinkedNoteCommand) (MutationOutcome, error)
}

func (operation LinkedNoteOperation) Execute(ctx context.Context, command LinkedNoteCommand) (MutationOutcome, error) {
	if operation.execute == nil {
		return MutationOutcome{}, errors.New("workbook linked-note operation is not initialized")
	}
	return operation.execute(ctx, command)
}

type LinkedNoteProvider interface {
	DecodeLinkedNote(io.Reader) (LinkedNoteOperation, *MutationFailure, error)
}

type linkedNoteProvider[T any] struct {
	decode  func(io.Reader) (T, bool, *MutationFailure, error)
	execute func(context.Context, LinkedNoteCommand, T) (MutationOutcome, error)
}

func NewLinkedNoteProvider[T any](
	decode func(io.Reader) (T, bool, *MutationFailure, error),
	execute func(context.Context, LinkedNoteCommand, T) (MutationOutcome, error),
) (LinkedNoteProvider, error) {
	if decode == nil || execute == nil {
		return nil, errors.New("linked-note provider requires decode and create functions")
	}
	return &linkedNoteProvider[T]{decode: decode, execute: execute}, nil
}

func (provider *linkedNoteProvider[T]) DecodeLinkedNote(reader io.Reader) (LinkedNoteOperation, *MutationFailure, error) {
	value, present, failure, err := provider.decode(reader)
	if validationErr := validateDecodedMutationValue("linked-note", value, present, failure, err); validationErr != nil {
		return LinkedNoteOperation{}, nil, validationErr
	}
	if failure != nil {
		return LinkedNoteOperation{}, failure, nil
	}
	return LinkedNoteOperation{execute: func(ctx context.Context, command LinkedNoteCommand) (MutationOutcome, error) {
		return provider.execute(ctx, command, value)
	}}, nil, nil
}

type SupersedeCommand struct {
	Actor     authn.UserRecord
	Target    RecordTarget
	RequestID string
	Now       time.Time
}

type SupersedeOperation struct {
	execute func(context.Context, SupersedeCommand) (MutationOutcome, error)
}

func (operation SupersedeOperation) Execute(ctx context.Context, command SupersedeCommand) (MutationOutcome, error) {
	if operation.execute == nil {
		return MutationOutcome{}, errors.New("workbook supersede operation is not initialized")
	}
	return operation.execute(ctx, command)
}

type SupersedeProvider interface {
	DecodeSupersede(io.Reader) (SupersedeOperation, *MutationFailure, error)
}

type supersedeProvider[T any] struct {
	decode  func(io.Reader) (T, bool, *MutationFailure, error)
	execute func(context.Context, SupersedeCommand, T) (MutationOutcome, error)
}

func NewSupersedeProvider[T any](
	decode func(io.Reader) (T, bool, *MutationFailure, error),
	execute func(context.Context, SupersedeCommand, T) (MutationOutcome, error),
) (SupersedeProvider, error) {
	if decode == nil || execute == nil {
		return nil, errors.New("supersede provider requires decode and supersede functions")
	}
	return &supersedeProvider[T]{decode: decode, execute: execute}, nil
}

func (provider *supersedeProvider[T]) DecodeSupersede(reader io.Reader) (SupersedeOperation, *MutationFailure, error) {
	value, present, failure, err := provider.decode(reader)
	if validationErr := validateDecodedMutationValue("supersede", value, present, failure, err); validationErr != nil {
		return SupersedeOperation{}, nil, validationErr
	}
	if failure != nil {
		return SupersedeOperation{}, failure, nil
	}
	return SupersedeOperation{execute: func(ctx context.Context, command SupersedeCommand) (MutationOutcome, error) {
		return provider.execute(ctx, command, value)
	}}, nil, nil
}

type ClipboardContribution struct {
	ViewSchemaID string
	Provider     ClipboardProvider
}

type BulkContribution struct {
	ViewSchemaID string
	Provider     BulkProvider
}

type LinkedNoteContribution struct {
	RecordType string
	Provider   LinkedNoteProvider
}

type SupersedeContribution struct {
	RecordType string
	Provider   SupersedeProvider
}

type MutationActionContributions struct {
	Clipboard  []ClipboardContribution
	Bulk       []BulkContribution
	LinkedNote []LinkedNoteContribution
	Supersede  []SupersedeContribution
}
