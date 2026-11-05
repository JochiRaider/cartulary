package crossownertransaction

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type Options struct {
	Limits            Limits
	ValidationPolicy  ValidationPolicy
	Backend           Backend
	Catalog           []Descriptor
	Timeout           time.Duration
	CancellationGrace time.Duration
	Clock             func() int64
	FatalSink         func(error)
}

type Coordinator struct {
	limits            Limits
	validationPolicy  ValidationPolicy
	backend           Backend
	catalog           map[string]Descriptor
	timeout           time.Duration
	cancellationGrace time.Duration
	clock             func() int64
	fatalSink         func(error)
}

var monotonicOrigin = time.Now()

func New(options Options) (*Coordinator, error) {
	if options.Backend == nil || options.FatalSink == nil || options.ValidationPolicy == nil || !options.Limits.valid() {
		return nil, ErrUnavailable
	}
	if options.Timeout <= 0 || options.CancellationGrace < 0 {
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
		if !options.ValidationPolicy.AcceptDescriptor(descriptor) {
			return nil, ErrParticipantSet
		}
		if _, duplicate := catalog[descriptor.ParticipantID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate %s", ErrParticipantSet, descriptor.ParticipantID)
		}
		catalog[descriptor.ParticipantID] = cloneDescriptor(descriptor)
	}
	return &Coordinator{
		limits:           options.Limits,
		validationPolicy: options.ValidationPolicy,
		backend:          options.Backend, catalog: catalog, timeout: options.Timeout,
		clock: clock, fatalSink: options.FatalSink, cancellationGrace: options.CancellationGrace,
	}, nil
}

// Timeout exposes the effective default for owner-required public diagnostics.
func (c *Coordinator) Timeout() time.Duration { return c.timeout }

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
	if parent, ok := ctx.Deadline(); ok {
		inherited := saturatingAdd(start, time.Until(parent))
		if inherited < deadline {
			deadline = inherited
		}
	}
	remaining := deadline - c.clock()
	if remaining < 0 {
		remaining = 0
	}
	ctx, cancelOperation := context.WithTimeout(ctx, time.Duration(remaining))
	defer cancelOperation()
	ctx = extensiondeadline.WithCancellationTracking(ctx)
	sample := func() error {
		if err := extensiondeadline.Sample(ctx, c.invocationPolicy(deadline)); err != nil {
			return interruptionOutcome(err)
		}
		return nil
	}
	if err := sample(); err != nil {
		return Result{}, err
	}

	inputs := make([]Input, len(participants))
	var aggregateInput int64
	for index, participant := range participants {
		if err := sample(); err != nil {
			return Result{}, err
		}
		input, buildErr := invokeParticipant(c, ctx, deadline, func(callCtx context.Context) (Input, error) {
			return participant.BuildInput(callCtx, OperationContext{
				OperationID: operation.OperationID, NormalizedRequestSHA256: operation.NormalizedRequestSHA256,
				DeadlineMonotonicNS: deadline,
			})
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
		total, withinLimit := boundedTotal(aggregateInput, int64(len(input.CanonicalBytes)), c.limits.ParticipantInputBytes, c.limits.AggregateInputBytes)
		if !withinLimit {
			return Result{}, fmt.Errorf("%w: %s bytes", ErrInput, participant.ID())
		}
		aggregateInput = total
		decoded, decodeErr := canonicaljson.DecodeExtension(input.CanonicalBytes)
		validInput := decodeErr == nil && c.validationPolicy.AcceptInput(descriptors[index], decoded)
		if err := sample(); err != nil {
			return Result{}, err
		}
		if !validInput {
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
		result, prepareErr := invokeParticipant(c, ctx, deadline, func(callCtx context.Context) (PrepareResult, error) { return participant.Prepare(callCtx, invocation) })
		if prepareErr != nil {
			return Result{}, fmt.Errorf("%w: %s: %w", ErrPrepare, participant.ID(), prepareErr)
		}
		if err := sample(); err != nil {
			return Result{}, err
		}
		validatedKeys, canonicalSize, validateErr := c.validatePrepare(descriptors[index], result)
		if err := sample(); err != nil {
			return Result{}, err
		}
		if validateErr != nil {
			return Result{}, validateErr
		}
		total, withinLimit := boundedTotal(aggregatePrepareBytes, canonicalSize, c.limits.PrepareResultBytes, c.limits.AggregatePrepareResultBytes)
		if !withinLimit {
			return Result{}, ErrPrepare
		}
		aggregatePrepareBytes = total
		if len(validatedKeys) > c.limits.AggregateSerializationKeys-len(keys) {
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
			_, _ = c.completeTransaction(ctx, deadline, func(callCtx context.Context) (CommitOutcome, error) {
				return tx.Rollback(context.WithoutCancel(callCtx))
			})
		}
	}()
	rollback := func(cause error) (Result, error) {
		if boundaryErr := sample(); boundaryErr != nil && cause == nil {
			cause = boundaryErr
		}
		completion, lifetimeErr := c.completeTransaction(ctx, deadline, func(callCtx context.Context) (CommitOutcome, error) {
			return tx.Rollback(context.WithoutCancel(callCtx))
		})
		finalized = true
		if lifetimeErr != nil {
			return Result{}, lifetimeErr
		}
		outcome, rollbackErr := completion.Value, completion.Err
		if outcome != CommitAbsent {
			return Result{}, c.fatal(rollbackErr)
		}
		if cause == nil {
			cause = rollbackErr
		}
		// After proven absence, preserve selected terminal outcomes before
		// applying the first interruption recorded through rollback completion.
		if IsFatalIntegrity(cause) || errors.Is(cause, ErrCanceled) || errors.Is(cause, ErrTimeout) {
			return Result{}, cause
		}
		if completion.Interruption != nil {
			return Result{}, interruptionOutcome(completion.Interruption)
		}
		var conflict *ConflictError
		if errors.As(cause, &conflict) {
			return Result{}, cause
		}
		return Result{}, classifyDatabaseError(cause)
	}
	for _, key := range keys {
		if err := sample(); err != nil {
			return rollback(err)
		}
		if err := tx.AcquireSerializationLock(ctx, key); err != nil {
			return rollback(err)
		}
		if err := sample(); err != nil {
			return rollback(err)
		}
	}
	if operation.Preconditions != nil {
		if err := sample(); err != nil {
			return rollback(err)
		}
		access, err := tx.PreconditionCapability()
		if err != nil || access == nil {
			return rollback(fmt.Errorf("%w: owner precondition capability", ErrValidation))
		}
		if _, err := invokeParticipant(c, ctx, deadline, func(callCtx context.Context) (struct{}, error) {
			return struct{}{}, operation.Preconditions.Check(callCtx, access)
		}); err != nil {
			return rollback(err)
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
		result, validationErr := invokeParticipant(c, ctx, deadline, func(callCtx context.Context) (ValidationResult, error) {
			return participant.Validate(callCtx, invocationFor("validate", operation, descriptors[index], inputs[index], deadline, access, nil, false))
		})
		if validationErr != nil {
			if IsFatalIntegrity(validationErr) || errors.Is(validationErr, ErrCanceled) || errors.Is(validationErr, ErrTimeout) {
				return rollback(validationErr)
			}
			return rollback(c.classifyValidation(descriptors[index], result, validationErr, timeout))
		}
		if err := sample(); err != nil {
			return rollback(err)
		}
		validateErr := c.classifyValidation(descriptors[index], result, nil, timeout)
		if err := sample(); err != nil {
			return rollback(err)
		}
		if validateErr != nil {
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
		result, writeErr := invokeParticipant(c, ctx, deadline, func(callCtx context.Context) (WriteResult, error) {
			return participant.Write(callCtx, invocationFor("write", operation, descriptors[index], inputs[index], deadline, access, access, false))
		})
		if writeErr != nil {
			return rollback(writeErr)
		}
		if err := sample(); err != nil {
			return rollback(err)
		}
		if !c.validWriteResult(descriptors[index], result) {
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
		if _, err := invokeParticipant(c, ctx, deadline, func(callCtx context.Context) (struct{}, error) {
			return struct{}{}, operation.Finalizer.Publish(callCtx, access, cloneValues(values))
		}); err != nil {
			return rollback(err)
		}
		if err := sample(); err != nil {
			return rollback(err)
		}
	}
	if err := sample(); err != nil {
		return rollback(err)
	}
	completion, lifetimeErr := c.completeTransaction(ctx, deadline, tx.Commit)
	finalized = true
	if lifetimeErr != nil {
		return Result{}, lifetimeErr
	}
	outcome, commitErr := completion.Value, completion.Err
	switch outcome {
	case CommitProven:
		return Result{ParticipantValues: values}, nil
	case CommitAbsent:
		if completion.Interruption != nil {
			return Result{}, interruptionOutcome(completion.Interruption)
		}
		if commitErr == nil {
			commitErr = ErrCommitAbsent
		}
		return Result{}, classifyDatabaseError(commitErr)
	default:
		return Result{}, c.fatal(commitErr)
	}
}

func (c *Coordinator) invocationPolicy(bound int64) extensiondeadline.InvocationPolicy {
	return extensiondeadline.InvocationPolicy{
		Grace:          c.cancellationGrace,
		Deadline:       &extensiondeadline.Deadline{MonotonicNS: bound},
		MonotonicNowNS: c.clock,
		Exhausted:      func(cause error) { c.fatalSink(&FatalIntegrityError{Cause: cause, cancellationGraceExhausted: true}) },
	}
}

func (c *Coordinator) completeTransaction(ctx context.Context, bound int64, finish func(context.Context) (CommitOutcome, error)) (extensiondeadline.Completion[CommitOutcome], error) {
	result, err := extensiondeadline.Complete(ctx, c.invocationPolicy(bound), finish)
	if err != nil {
		return result, &FatalIntegrityError{Cause: err, cancellationGraceExhausted: true}
	}
	return result, nil
}
func interruptionOutcome(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	return ErrCanceled
}

func invokeParticipant[T any](c *Coordinator, ctx context.Context, bound int64, invoke func(context.Context) (T, error)) (T, error) {
	result, err := extensiondeadline.Invoke(ctx, c.invocationPolicy(bound), invoke)
	if errors.Is(err, extensiondeadline.ErrGraceExhausted) {
		return result, &FatalIntegrityError{Cause: err, cancellationGraceExhausted: true}
	}
	if extensiondeadline.Interrupted(err) {
		if errors.Is(err, context.DeadlineExceeded) {
			return result, ErrTimeout
		}
		if errors.Is(err, context.Canceled) {
			return result, ErrCanceled
		}
	}
	return result, err
}

func (c *Coordinator) resolve(source []Participant) ([]Participant, []Descriptor, error) {
	if len(source) < c.limits.MinimumParticipants || len(source) > c.limits.Participants {
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
	if descriptor.ParticipantID == "" || descriptor.ContractSHA256 == "" ||
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

func (c *Coordinator) validatePrepare(descriptor Descriptor, result PrepareResult) ([]OrderedSerializationKey, int64, error) {
	if result.SerializationKeys == nil {
		return nil, 0, ErrPrepare
	}
	if len(result.SerializationKeys) > c.limits.SerializationKeysPerParticipant {
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
	for index := 1; index < len(keys); index++ {
		if compareKey(keys[index-1], keys[index]) >= 0 {
			return nil, 0, ErrSerializationKeys
		}
	}
	canonical, err := canonicaljson.MarshalExtension(map[string]any{
		"schema_id": PrepareResultSchema, "participant_id": descriptor.ParticipantID,
		"serialization_keys": prepareKeyObjects(keys),
	})
	if err != nil || len(canonical) < 1 || int64(len(canonical)) > c.limits.PrepareResultBytes {
		return nil, 0, ErrPrepare
	}
	return keys, int64(len(canonical)), nil
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
