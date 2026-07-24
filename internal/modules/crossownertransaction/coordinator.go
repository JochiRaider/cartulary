package crossownertransaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type Options struct {
	Backend   Backend
	Catalog   []Descriptor
	Timeout   time.Duration
	Clock     func() int64
	FatalSink func(error)
}

type Coordinator struct {
	backend   Backend
	catalog   map[string]Descriptor
	timeout   time.Duration
	clock     func() int64
	fatalSink func(error)
}

var monotonicOrigin = time.Now()

func New(options Options) (*Coordinator, error) {
	if options.Backend == nil || options.FatalSink == nil {
		return nil, ErrUnavailable
	}
	if options.Timeout <= 0 {
		return nil, fmt.Errorf("%w: timeout", ErrUnavailable)
	}
	clock := options.Clock
	if clock == nil {
		clock = func() int64 { return time.Since(monotonicOrigin).Nanoseconds() }
	}
	catalog := make(map[string]Descriptor, len(options.Catalog))
	for _, descriptor := range options.Catalog {
		if err := validateDescriptor(descriptor); err != nil {
			return nil, err
		}
		if _, duplicate := catalog[descriptor.ParticipantID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate %s", ErrParticipantSet, descriptor.ParticipantID)
		}
		catalog[descriptor.ParticipantID] = cloneDescriptor(descriptor)
	}
	return &Coordinator{
		backend: options.Backend, catalog: catalog, timeout: options.Timeout,
		clock: clock, fatalSink: options.FatalSink,
	}, nil
}

func (c *Coordinator) Execute(ctx context.Context, operation Operation) (Result, error) {
	if c == nil || c.backend == nil || c.clock == nil || c.fatalSink == nil {
		return Result{}, ErrUnavailable
	}
	participants, descriptors, err := c.resolve(operation.Participants)
	if err != nil {
		return Result{}, err
	}
	if operation.OperationID == "" || operation.NormalizedRequestSHA256 == "" {
		return Result{}, ErrInput
	}
	timeout := operation.Timeout
	if timeout <= 0 {
		timeout = c.timeout
	}
	start := c.clock()
	deadline := saturatingAdd(start, timeout)
	sample := func() error { return sampleContext(ctx, c.clock(), deadline) }
	if err := sample(); err != nil {
		return Result{}, err
	}

	inputs := make([]Input, len(participants))
	var aggregateInput int64
	for index, participant := range participants {
		if err := sample(); err != nil {
			return Result{}, err
		}
		input, buildErr := participant.BuildInput(ctx, OperationContext{
			OperationID: operation.OperationID, NormalizedRequestSHA256: operation.NormalizedRequestSHA256,
			DeadlineMonotonicNS: deadline,
		})
		if buildErr != nil {
			return Result{}, fmt.Errorf("%w: %s: %w", ErrInput, participant.ID(), buildErr)
		}
		if err := sample(); err != nil {
			return Result{}, err
		}
		if input.SchemaID != descriptors[index].InputSchemaID {
			return Result{}, fmt.Errorf("%w: %s schema", ErrInput, participant.ID())
		}
		size := int64(len(input.CanonicalBytes))
		if size < 1 || size > ParticipantInputByteLimit || aggregateInput > math.MaxInt64-size {
			return Result{}, fmt.Errorf("%w: %s bytes", ErrInput, participant.ID())
		}
		aggregateInput += size
		if aggregateInput > AggregateInputByteLimit {
			return Result{}, ErrInput
		}
		input.CanonicalBytes = bytes.Clone(input.CanonicalBytes)
		inputs[index] = input
	}

	prepared := make([]PrepareResult, len(participants))
	keys := make([]OrderedSerializationKey, 0)
	var aggregatePrepareBytes int64
	for index, participant := range participants {
		if err := sample(); err != nil {
			return Result{}, err
		}
		invocation := invocationFor("prepare", operation, descriptors[index], inputs[index], deadline, nil, nil, false)
		result, prepareErr := participant.Prepare(ctx, invocation)
		if prepareErr != nil {
			return Result{}, fmt.Errorf("%w: %s: %w", ErrPrepare, participant.ID(), prepareErr)
		}
		if err := sample(); err != nil {
			return Result{}, err
		}
		validatedKeys, canonicalSize, validateErr := validatePrepare(descriptors[index], result)
		if validateErr != nil {
			return Result{}, validateErr
		}
		aggregatePrepareBytes += canonicalSize
		if aggregatePrepareBytes > AggregateInputByteLimit || len(keys)+len(validatedKeys) > AggregateSerializationKeyLimit {
			return Result{}, ErrSerializationKeys
		}
		keys = append(keys, validatedKeys...)
		prepared[index] = result
	}
	sort.Slice(keys, func(i, j int) bool { return compareKey(keys[i], keys[j]) < 0 })
	for index := 1; index < len(keys); index++ {
		if compareKey(keys[index-1], keys[index]) == 0 {
			return Result{}, ErrSerializationKeys
		}
	}
	if err := sample(); err != nil {
		return Result{}, err
	}
	tx, err := c.backend.Begin(ctx, descriptors)
	if err != nil {
		return Result{}, classifyDatabaseError(err)
	}
	finalized := false
	defer func() {
		if !finalized {
			_, _ = tx.Rollback(context.WithoutCancel(ctx))
		}
	}()
	rollback := func(cause error) (Result, error) {
		if boundaryErr := sample(); boundaryErr != nil && cause == nil {
			cause = boundaryErr
		}
		outcome, rollbackErr := tx.Rollback(context.WithoutCancel(ctx))
		finalized = true
		if outcome != CommitAbsent {
			return Result{}, c.fatal(rollbackErr)
		}
		if cause != nil {
			return Result{}, cause
		}
		return Result{}, rollbackErr
	}
	for _, key := range keys {
		if err := sample(); err != nil {
			return rollback(err)
		}
		if err := tx.AcquireSerializationLock(ctx, key); err != nil {
			return rollback(classifyDatabaseError(err))
		}
		if err := sample(); err != nil {
			return rollback(err)
		}
	}
	for index, participant := range participants {
		if err := sample(); err != nil {
			return rollback(err)
		}
		access, accessErr := tx.ReadCapability(participant.ID())
		if accessErr != nil || access == nil || access.ParticipantScope() != participant.ID() {
			return rollback(fmt.Errorf("%w: %s read capability", ErrValidation, participant.ID()))
		}
		result, validationErr := participant.Validate(ctx, invocationFor("validate", operation, descriptors[index], inputs[index], deadline, access, nil, false))
		if validationErr != nil {
			return rollback(fmt.Errorf("%w: %s: %w", ErrValidation, participant.ID(), validationErr))
		}
		if err := sample(); err != nil {
			return rollback(err)
		}
		if validateErr := validateValidation(participant.ID(), result); validateErr != nil {
			return rollback(validateErr)
		}
	}
	values := make(map[string]any, len(participants))
	for index, participant := range participants {
		if err := sample(); err != nil {
			return rollback(err)
		}
		access, accessErr := tx.WriteCapability(participant.ID())
		if accessErr != nil || access == nil || access.ParticipantScope() != participant.ID() {
			return rollback(fmt.Errorf("%w: %s write capability", ErrWrite, participant.ID()))
		}
		result, writeErr := participant.Write(ctx, invocationFor("write", operation, descriptors[index], inputs[index], deadline, access, access, false))
		if writeErr != nil {
			return rollback(writeErr)
		}
		if err := sample(); err != nil {
			return rollback(err)
		}
		if result.Status != "written" {
			return rollback(fmt.Errorf("%w: %s", ErrWrite, participant.ID()))
		}
		values[participant.ID()] = result.Value
	}
	if operation.Finalizer != nil {
		if err := sample(); err != nil {
			return rollback(err)
		}
		access, accessErr := tx.FinalizationCapability()
		if accessErr != nil || access == nil {
			return rollback(fmt.Errorf("%w: finalization capability", ErrWrite))
		}
		if err := operation.Finalizer.Publish(ctx, access, cloneValues(values)); err != nil {
			return rollback(err)
		}
		if err := sample(); err != nil {
			return rollback(err)
		}
	}
	if err := sample(); err != nil {
		return rollback(err)
	}
	outcome, commitErr := tx.Commit(ctx)
	finalized = true
	switch outcome {
	case CommitProven:
		return Result{ParticipantValues: values}, nil
	case CommitAbsent:
		if commitErr == nil {
			commitErr = ErrCommitAbsent
		}
		return Result{}, classifyDatabaseError(commitErr)
	default:
		return Result{}, c.fatal(commitErr)
	}
}

func (c *Coordinator) resolve(source []Participant) ([]Participant, []Descriptor, error) {
	if len(source) < 1 || len(source) > ParticipantLimit {
		return nil, nil, ErrParticipantSet
	}
	participants := append([]Participant(nil), source...)
	sort.Slice(participants, func(i, j int) bool {
		if participants[i] == nil {
			return false
		}
		if participants[j] == nil {
			return true
		}
		return participants[i].ID() < participants[j].ID()
	})
	descriptors := make([]Descriptor, len(participants))
	previous := ""
	for index, participant := range participants {
		if participant == nil || participant.ID() == "" || participant.ID() == previous {
			return nil, nil, ErrParticipantSet
		}
		descriptor, admitted := c.catalog[participant.ID()]
		if !admitted {
			return nil, nil, fmt.Errorf("%w: unadmitted %s", ErrParticipantSet, participant.ID())
		}
		participants[index] = participant
		descriptors[index] = cloneDescriptor(descriptor)
		previous = participant.ID()
	}
	return participants, descriptors, nil
}

func (c *Coordinator) fatal(cause error) error {
	if cause == nil {
		cause = ErrCommitIndeterminate
	}
	fatal := &FatalIntegrityError{Cause: cause}
	c.fatalSink(fatal)
	return fatal
}

func validateDescriptor(descriptor Descriptor) error {
	if descriptor.ParticipantID == "" || descriptor.OwnerProfileID == "" || descriptor.ContractSHA256 == "" ||
		descriptor.InputSchemaID == "" || descriptor.PrepareAlgorithmID == "" ||
		descriptor.ValidationAlgorithmID == "" || descriptor.WriteAlgorithmID == "" ||
		len(descriptor.SerializationKeyKinds) < 1 || len(descriptor.SerializationKeyKinds) > 32 ||
		len(descriptor.OwnedStateFamilyIDs) < 1 || len(descriptor.OwnedStateFamilyIDs) > 64 ||
		!strictlySortedUnique(descriptor.SerializationKeyKinds) ||
		!strictlySortedUnique(descriptor.OwnedStateFamilyIDs) {
		return fmt.Errorf("%w: descriptor %s", ErrParticipantSet, descriptor.ParticipantID)
	}
	return nil
}

func validatePrepare(descriptor Descriptor, result PrepareResult) ([]OrderedSerializationKey, int64, error) {
	if len(result.SerializationKeys) > SerializationKeysPerOwner {
		return nil, 0, ErrSerializationKeys
	}
	declared := make(map[string]struct{}, len(descriptor.SerializationKeyKinds))
	for _, kind := range descriptor.SerializationKeyKinds {
		declared[kind] = struct{}{}
	}
	keys := make([]OrderedSerializationKey, len(result.SerializationKeys))
	for index, key := range result.SerializationKeys {
		if _, ok := declared[key.KeyKind]; !ok || key.Key == "" || len(key.Key) > 512 ||
			!utf8.ValidString(key.Key) || strings.IndexByte(key.Key, 0) >= 0 {
			return nil, 0, ErrSerializationKeys
		}
		keys[index] = OrderedSerializationKey{ParticipantID: descriptor.ParticipantID, KeyKind: key.KeyKind, Key: key.Key}
	}
	sort.Slice(keys, func(i, j int) bool { return compareKey(keys[i], keys[j]) < 0 })
	for index := 1; index < len(keys); index++ {
		if compareKey(keys[index-1], keys[index]) == 0 {
			return nil, 0, ErrSerializationKeys
		}
	}
	canonical, _ := json.Marshal(map[string]any{
		"schema_id": PrepareResultSchema, "participant_id": descriptor.ParticipantID,
		"serialization_keys": keys,
	})
	if len(canonical) < 1 || len(canonical) > AggregateInputByteLimit {
		return nil, 0, ErrPrepare
	}
	return keys, int64(len(canonical)), nil
}

func validateValidation(participantID string, result ValidationResult) error {
	if len(result.Findings) > ValidationFindingLimit {
		return ErrValidation
	}
	if (result.Status == "valid" && len(result.Findings) != 0) ||
		(result.Status == "invalid" && len(result.Findings) == 0) ||
		(result.Status != "valid" && result.Status != "invalid") {
		return ErrValidation
	}
	previous := Finding{}
	for index, finding := range result.Findings {
		if finding.Path == "" || finding.ReasonCode == "" || finding.Message == "" {
			return ErrValidation
		}
		if index > 0 && compareFinding(previous, finding) >= 0 {
			return ErrValidation
		}
		previous = finding
	}
	canonical, _ := json.Marshal(map[string]any{
		"schema_id": ValidationResultSchema, "participant_id": participantID,
		"status": result.Status, "findings": result.Findings,
	})
	if len(canonical) > ResultByteLimit {
		return ErrValidation
	}
	if result.Status == "invalid" {
		return ErrValidation
	}
	return nil
}

func invocationFor(phase string, operation Operation, descriptor Descriptor, input Input, deadline int64, read ReadCapability, write WriteCapability, canceled bool) Invocation {
	return Invocation{
		SchemaID: ParticipantContextSchema, Phase: phase, OperationID: operation.OperationID,
		ParticipantID: descriptor.ParticipantID, OwnerProfileID: descriptor.OwnerProfileID,
		NormalizedRequestSHA256: operation.NormalizedRequestSHA256, Input: input,
		CancellationRequested: canceled, DeadlineMonotonicNS: deadline,
		ReadAccess: read, WriteAccess: write,
	}
}

func sampleContext(ctx context.Context, now int64, deadline int64) error {
	if now >= deadline {
		return ErrTimeout
	}
	if ctx == nil {
		return ErrCanceled
	}
	if err := ctx.Err(); errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	} else if err != nil {
		return ErrCanceled
	}
	return nil
}

func classifyDatabaseError(err error) error {
	if err == nil {
		return nil
	}
	type sqlState interface{ SQLState() string }
	var state sqlState
	if errors.As(err, &state) {
		switch state.SQLState() {
		case "40P01":
			return &ConflictError{ReasonCode: "deadlock_detected", Cause: err}
		case "40001":
			return &ConflictError{ReasonCode: "serialization_failure", Cause: err}
		}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	if errors.Is(err, context.Canceled) {
		return ErrCanceled
	}
	return err
}

func cloneDescriptor(source Descriptor) Descriptor {
	source.SerializationKeyKinds = append([]string(nil), source.SerializationKeyKinds...)
	source.OwnedStateFamilyIDs = append([]string(nil), source.OwnedStateFamilyIDs...)
	return source
}

func cloneValues(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func compareKey(left, right OrderedSerializationKey) int {
	if value := strings.Compare(left.ParticipantID, right.ParticipantID); value != 0 {
		return value
	}
	if value := strings.Compare(left.KeyKind, right.KeyKind); value != 0 {
		return value
	}
	return strings.Compare(left.Key, right.Key)
}

func compareFinding(left, right Finding) int {
	if value := strings.Compare(left.Path, right.Path); value != 0 {
		return value
	}
	if value := strings.Compare(left.ReasonCode, right.ReasonCode); value != 0 {
		return value
	}
	return bytes.Compare(left.Details, right.Details)
}

func strictlySortedUnique(values []string) bool {
	for index, value := range values {
		if value == "" || (index > 0 && values[index-1] >= value) {
			return false
		}
	}
	return true
}

func saturatingAdd(now int64, timeout time.Duration) int64 {
	delta := int64(timeout)
	if delta <= 0 {
		return now
	}
	if now > math.MaxInt64-delta {
		return math.MaxInt64
	}
	return now + delta
}
