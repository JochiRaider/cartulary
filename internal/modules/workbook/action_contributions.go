package workbook

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

// ClipboardAdmission exposes only the routing facts Workbook owns. The
// concrete decoded value remains private to the contributing adapter.
type ClipboardAdmission interface {
	ClientTransactionID() string
	AdmittedViewSchemaID() string
}

type ClipboardCommand struct {
	Actor        authn.UserRecord
	IncidentID   uuid.UUID
	ViewSchemaID string
	Admission    ClipboardAdmission
	RequestID    string
	Now          time.Time
}

type ClipboardProvider interface {
	ValidateWorkbookContribution() error
	DecodeClipboard(io.Reader) (ClipboardAdmission, *MutationFailure, error)
	ApplyClipboard(context.Context, ClipboardCommand) (MutationOutcome, error)
}

type ClipboardDecodeFunc func(io.Reader) (ClipboardAdmission, *MutationFailure, error)
type ClipboardExecuteFunc func(context.Context, ClipboardCommand) (MutationOutcome, error)

type clipboardProvider struct {
	decode  ClipboardDecodeFunc
	execute ClipboardExecuteFunc
}

func NewClipboardProvider(decode ClipboardDecodeFunc, execute ClipboardExecuteFunc) (ClipboardProvider, error) {
	provider := &clipboardProvider{decode: decode, execute: execute}
	if err := provider.ValidateWorkbookContribution(); err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *clipboardProvider) ValidateWorkbookContribution() error {
	if provider == nil || provider.decode == nil || provider.execute == nil {
		return fmt.Errorf("clipboard provider requires decode and apply functions")
	}
	return nil
}

func (provider *clipboardProvider) DecodeClipboard(reader io.Reader) (ClipboardAdmission, *MutationFailure, error) {
	return provider.decode(reader)
}

func (provider *clipboardProvider) ApplyClipboard(ctx context.Context, command ClipboardCommand) (MutationOutcome, error) {
	return provider.execute(ctx, command)
}

type BulkAdmission interface {
	ClientTransactionID() string
	AdmittedViewSchemaID() string
}

type BulkCommand struct {
	Actor        authn.UserRecord
	IncidentID   uuid.UUID
	ViewSchemaID string
	Admission    BulkAdmission
	RequestID    string
	Now          time.Time
}

type BulkProvider interface {
	ValidateWorkbookContribution() error
	DecodeBulk(io.Reader) (BulkAdmission, *MutationFailure, error)
	ApplyBulk(context.Context, BulkCommand) (MutationOutcome, error)
}

type BulkDecodeFunc func(io.Reader) (BulkAdmission, *MutationFailure, error)
type BulkExecuteFunc func(context.Context, BulkCommand) (MutationOutcome, error)

type bulkProvider struct {
	decode  BulkDecodeFunc
	execute BulkExecuteFunc
}

func NewBulkProvider(decode BulkDecodeFunc, execute BulkExecuteFunc) (BulkProvider, error) {
	provider := &bulkProvider{decode: decode, execute: execute}
	if err := provider.ValidateWorkbookContribution(); err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *bulkProvider) ValidateWorkbookContribution() error {
	if provider == nil || provider.decode == nil || provider.execute == nil {
		return fmt.Errorf("bulk provider requires decode and apply functions")
	}
	return nil
}

func (provider *bulkProvider) DecodeBulk(reader io.Reader) (BulkAdmission, *MutationFailure, error) {
	return provider.decode(reader)
}

func (provider *bulkProvider) ApplyBulk(ctx context.Context, command BulkCommand) (MutationOutcome, error) {
	return provider.execute(ctx, command)
}

type LinkedNoteAdmission interface {
	ClientTransactionID() string
}

type LinkedNoteCommand struct {
	Actor     authn.UserRecord
	Target    RecordTarget
	Admission LinkedNoteAdmission
	RequestID string
	Now       time.Time
}

type LinkedNoteProvider interface {
	ValidateWorkbookContribution() error
	DecodeLinkedNote(io.Reader) (LinkedNoteAdmission, *MutationFailure, error)
	CreateLinkedNote(context.Context, LinkedNoteCommand) (MutationOutcome, error)
}

type LinkedNoteDecodeFunc func(io.Reader) (LinkedNoteAdmission, *MutationFailure, error)
type LinkedNoteExecuteFunc func(context.Context, LinkedNoteCommand) (MutationOutcome, error)

type linkedNoteProvider struct {
	decode  LinkedNoteDecodeFunc
	execute LinkedNoteExecuteFunc
}

func NewLinkedNoteProvider(decode LinkedNoteDecodeFunc, execute LinkedNoteExecuteFunc) (LinkedNoteProvider, error) {
	provider := &linkedNoteProvider{decode: decode, execute: execute}
	if err := provider.ValidateWorkbookContribution(); err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *linkedNoteProvider) ValidateWorkbookContribution() error {
	if provider == nil || provider.decode == nil || provider.execute == nil {
		return fmt.Errorf("linked-note provider requires decode and create functions")
	}
	return nil
}

func (provider *linkedNoteProvider) DecodeLinkedNote(reader io.Reader) (LinkedNoteAdmission, *MutationFailure, error) {
	return provider.decode(reader)
}

func (provider *linkedNoteProvider) CreateLinkedNote(ctx context.Context, command LinkedNoteCommand) (MutationOutcome, error) {
	return provider.execute(ctx, command)
}

type SupersedeAdmission interface {
	ClientTransactionID() string
	AdmittedBaseRowVersion() int64
}

type SupersedeCommand struct {
	Actor     authn.UserRecord
	Target    RecordTarget
	Admission SupersedeAdmission
	RequestID string
	Now       time.Time
}

type SupersedeProvider interface {
	ValidateWorkbookContribution() error
	DecodeSupersede(io.Reader) (SupersedeAdmission, *MutationFailure, error)
	Supersede(context.Context, SupersedeCommand) (MutationOutcome, error)
}

type SupersedeDecodeFunc func(io.Reader) (SupersedeAdmission, *MutationFailure, error)
type SupersedeExecuteFunc func(context.Context, SupersedeCommand) (MutationOutcome, error)

type supersedeProvider struct {
	decode  SupersedeDecodeFunc
	execute SupersedeExecuteFunc
}

func NewSupersedeProvider(decode SupersedeDecodeFunc, execute SupersedeExecuteFunc) (SupersedeProvider, error) {
	provider := &supersedeProvider{decode: decode, execute: execute}
	if err := provider.ValidateWorkbookContribution(); err != nil {
		return nil, err
	}
	return provider, nil
}

func (provider *supersedeProvider) ValidateWorkbookContribution() error {
	if provider == nil || provider.decode == nil || provider.execute == nil {
		return fmt.Errorf("supersede provider requires decode and supersede functions")
	}
	return nil
}

func (provider *supersedeProvider) DecodeSupersede(reader io.Reader) (SupersedeAdmission, *MutationFailure, error) {
	return provider.decode(reader)
}

func (provider *supersedeProvider) Supersede(ctx context.Context, command SupersedeCommand) (MutationOutcome, error) {
	return provider.execute(ctx, command)
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
