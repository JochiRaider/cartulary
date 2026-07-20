package extensions

import (
	"context"
	"errors"
	"fmt"
	"sort"
)

const (
	PortabilityExportResultSchema = "cartulary.extension_portability_export_result.v1"
	PortabilityImportResultSchema = "cartulary.extension_portability_import_preparation_result.v1"
)

var (
	ErrPortabilityPayloadMissing = errors.New("extension_portability_payload_missing")
	ErrPortabilityPayloadInvalid = errors.New("extension_portability_payload_invalid")
	ErrPortabilityInputLimit     = errors.New("extension_portability_input_limit")
	ErrPortabilityResultInvalid  = errors.New("extension_portability_result_invalid")
	ErrStagedOutputScope         = errors.New("extension_staged_output_scope_denied")
)

type PortabilityExportResult struct {
	SchemaID               string
	Kind                   string
	PayloadSchemaID        string
	PayloadContractMajor   int
	StateVersion           int
	CanonicalPayloadSHA256 string
	PayloadByteSize        int64
	PayloadRef             string
}

func (r PortabilityExportResult) Validate() error {
	if r.SchemaID != PortabilityExportResultSchema {
		return ErrPortabilityResultInvalid
	}
	if r.Kind == "omit" {
		if r.PayloadSchemaID != "" || r.PayloadRef != "" || r.PayloadByteSize != 0 {
			return ErrPortabilityResultInvalid
		}
		return nil
	}
	if r.Kind != "payload" || r.PayloadSchemaID == "" || r.PayloadContractMajor < 1 || r.StateVersion < 1 || r.CanonicalPayloadSHA256 == "" || r.PayloadByteSize < 0 || r.PayloadRef == "" {
		return ErrPortabilityResultInvalid
	}
	return nil
}

type PortabilityImportPreparation struct {
	SchemaID               string
	Status                 string
	ParticipantInput       []byte
	ParticipantInputSHA256 string
	StagedOutputRefs       []string
}

func (r PortabilityImportPreparation) Validate() error {
	if r.SchemaID != PortabilityImportResultSchema || r.Status != "prepared" || len(r.ParticipantInput) > TransactionByteLimit || r.ParticipantInputSHA256 == "" || !strictlySortedUnique(r.StagedOutputRefs) {
		return ErrPortabilityResultInvalid
	}
	return nil
}

type PortabilityImportParticipant interface {
	ID() string
	PrepareImport(context.Context, []byte, *StagedOutputScope) (PortabilityImportPreparation, error)
}

type StagedOutputAllocator interface {
	Allocate(context.Context, string, string, []byte) (string, error)
	Abandon(context.Context, string) error
}

// StagedOutputScope is process-local and operation-bound. It deliberately has
// no publication, redemption, storage-identity, or transaction methods.
type StagedOutputScope struct {
	operationID string
	profileID   string
	allocator   StagedOutputAllocator
	refs        []string
	closed      bool
}

func NewStagedOutputScope(operationID, profileID string, allocator StagedOutputAllocator) (*StagedOutputScope, error) {
	if operationID == "" || profileID == "" || allocator == nil {
		return nil, ErrStagedOutputScope
	}
	return &StagedOutputScope{operationID: operationID, profileID: profileID, allocator: allocator}, nil
}

func (s *StagedOutputScope) Allocate(ctx context.Context, operationID string, bytes []byte) (string, error) {
	if s == nil || s.closed || operationID != s.operationID || len(bytes) > TransactionByteLimit {
		return "", ErrStagedOutputScope
	}
	ref, err := s.allocator.Allocate(ctx, s.operationID, s.profileID, bytes)
	if err != nil {
		return "", err
	}
	if ref == "" {
		return "", ErrStagedOutputScope
	}
	s.refs = append(s.refs, ref)
	return ref, nil
}

func (s *StagedOutputScope) Refs() []string {
	if s == nil {
		return nil
	}
	refs := append([]string(nil), s.refs...)
	sort.Strings(refs)
	return refs
}

func (s *StagedOutputScope) Abandon(ctx context.Context) error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	var joined error
	for _, ref := range s.refs {
		joined = errors.Join(joined, s.allocator.Abandon(ctx, ref))
	}
	return joined
}

func (s *StagedOutputScope) TransferToTransaction(operationID string) ([]string, error) {
	if s == nil || s.closed || operationID != s.operationID {
		return nil, ErrStagedOutputScope
	}
	s.closed = true
	return s.Refs(), nil
}

type PortabilityImportRequest struct {
	OperationID string
	ProfileID   string
	Payload     []byte
	Participant PortabilityImportParticipant
	Allocator   StagedOutputAllocator
}

func PreparePortabilityImports(ctx context.Context, requests []PortabilityImportRequest) ([]PortabilityImportPreparation, []*StagedOutputScope, error) {
	if len(requests) == 0 {
		return nil, nil, nil
	}
	var aggregate int64
	previous := ""
	for _, request := range requests {
		if request.Participant == nil || request.Participant.ID() == "" || (previous != "" && previous >= request.Participant.ID()) {
			return nil, nil, ErrPortabilityPayloadInvalid
		}
		previous = request.Participant.ID()
		if request.Payload == nil {
			return nil, nil, ErrPortabilityPayloadMissing
		}
		if len(request.Payload) > TransactionByteLimit {
			return nil, nil, ErrPortabilityInputLimit
		}
		aggregate += int64(len(request.Payload))
		if aggregate > TransactionByteLimit {
			return nil, nil, ErrPortabilityInputLimit
		}
	}
	results := make([]PortabilityImportPreparation, 0, len(requests))
	scopes := make([]*StagedOutputScope, 0, len(requests))
	abandon := func() {
		for _, scope := range scopes {
			_ = scope.Abandon(context.WithoutCancel(ctx))
		}
	}
	var aggregatePrepared int64
	for _, request := range requests {
		if err := transactionContextError(ctx); err != nil {
			abandon()
			return nil, nil, err
		}
		scope, err := NewStagedOutputScope(request.OperationID, request.ProfileID, request.Allocator)
		if err != nil {
			abandon()
			return nil, nil, err
		}
		scopes = append(scopes, scope)
		result, err := request.Participant.PrepareImport(ctx, append([]byte(nil), request.Payload...), scope)
		if err != nil {
			abandon()
			return nil, nil, fmt.Errorf("%w: %s", ErrPortabilityResultInvalid, request.Participant.ID())
		}
		if err := result.Validate(); err != nil || !equalStringSlices(result.StagedOutputRefs, scope.Refs()) {
			abandon()
			return nil, nil, ErrPortabilityResultInvalid
		}
		aggregatePrepared += int64(len(result.ParticipantInput))
		if aggregatePrepared > TransactionByteLimit {
			abandon()
			return nil, nil, ErrPortabilityResultInvalid
		}
		results = append(results, result)
	}
	return results, scopes, nil
}

func equalStringSlices(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
