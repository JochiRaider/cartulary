package reference_data

import (
	"context"
	"errors"
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// OperationObserver receives only closed tokens. Request data and errors never
// cross this port; telemetry failure cannot change an operation's result.
type OperationObserver interface {
	BeginReferenceOperation(context.Context, string) (context.Context, func(string))
}

func ReferenceOperationNames() []string {
	return []string{"reference_pack.import", "reference_pack.verify", "reference_pack.reverify", "reference_pack.activate", "reference_pack.disable", "reference_pack.refresh", "reference_pack.reconcile", "reference_pack.remove", "reference_pack.invalidate", "reference_pack.collection", "reference_pack.lookup"}
}

func observeReferenceOperation(ctx context.Context, observer OperationObserver, operation string) (context.Context, func(string)) {
	if observer == nil || !slices.Contains(ReferenceOperationNames(), operation) {
		return ctx, func(string) {}
	}
	return observer.BeginReferenceOperation(ctx, operation)
}

func referenceOutcome(err error) string {
	if err == nil {
		return "success"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, jobs.ErrCancellationRequested) {
		return "canceled"
	}
	var content *ContentRejection
	var rejected *OperationRejection
	var consumer *ConsumerError
	if errors.As(err, &content) || errors.As(err, &rejected) {
		return "rejected"
	}
	if errors.As(err, &consumer) && consumer.Code != "pack_unavailable" {
		return "rejected"
	}
	return "failed"
}

type observedConsumer struct {
	Consumer
	observer OperationObserver
}

func observeConsumer(consumer Consumer, observer OperationObserver) Consumer {
	if observer == nil {
		return consumer
	}
	return &observedConsumer{consumer, observer}
}

func (c *observedConsumer) ResolveCurrentPackSet(ctx context.Context) ConsumerResult[PackSet] {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.lookup")
	result := c.Consumer.ResolveCurrentPackSet(ctx)
	end(consumerOutcome(result.Error))
	return result
}
func (c *observedConsumer) GetPackEntry(ctx context.Context, request GetPackEntryRequest) ConsumerResult[PackEntry] {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.lookup")
	result := c.Consumer.GetPackEntry(ctx, request)
	end(consumerOutcome(result.Error))
	return result
}
func (c *observedConsumer) LookupPackEntries(ctx context.Context, request LookupPackEntriesRequest) ConsumerResult[PackEntryPage] {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.lookup")
	result := c.Consumer.LookupPackEntries(ctx, request)
	end(consumerOutcome(result.Error))
	return result
}
func (c *observedConsumer) EvaluateIndicatorValue(ctx context.Context, request EvaluateIndicatorRequest) ConsumerResult[IndicatorEvaluation] {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.lookup")
	result := c.Consumer.EvaluateIndicatorValue(ctx, request)
	end(consumerOutcome(result.Error))
	return result
}
func (c *observedConsumer) GetPackProvenance(ctx context.Context, request GetPackProvenanceRequest) ConsumerResult[PackProvenance] {
	ctx, end := observeReferenceOperation(ctx, c.observer, "reference_pack.lookup")
	result := c.Consumer.GetPackProvenance(ctx, request)
	end(consumerOutcome(result.Error))
	return result
}
func consumerOutcome(err *ConsumerError) string {
	if err == nil {
		return "success"
	}
	return referenceOutcome(err)
}
