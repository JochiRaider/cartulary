// Package testsupport supplies owner-controlled fixtures to other owners.
// These fixtures construct records; conformance expectations remain separately
// authored vectors exercised through the real consumer and persistence ports.
package testsupport

import (
	"context"
	"errors"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/jackc/pgx/v5"
)

func EvaluateIndicator(typ, kind, raw string) (reference_data.IndicatorEvaluation, error) {
	for _, policy := range packformat.IndicatorPolicies() {
		if policy.Type == typ {
			value, err := packformat.Evaluate(policy, kind, raw)
			if err != nil {
				var rejected *packformat.Failure
				if errors.As(err, &rejected) {
					return reference_data.IndicatorEvaluation{}, &reference_data.ConsumerError{Code: rejected.Code}
				}
				return reference_data.IndicatorEvaluation{}, err
			}
			return reference_data.IndicatorEvaluation{Evaluation: value}, nil
		}
	}
	return reference_data.IndicatorEvaluation{}, &reference_data.ConsumerError{Code: "indicator_type_unsupported"}
}

// IndicatorRegistry is an in-memory fixture for caller tests which do not
// exercise Reference Data storage. Production composition must use the real
// NewConsumer and NewRegistryAssignments constructors.
type IndicatorRegistry struct{}

func (IndicatorRegistry) BeginTx(context.Context, pgx.Tx) (reference_data.RegistryAssignment, error) {
	return IndicatorRegistry{}, nil
}
func (IndicatorRegistry) RecordUsage(context.Context, string) error { return nil }
func (IndicatorRegistry) ResolveCurrentPackSet(context.Context) reference_data.ConsumerResult[reference_data.PackSet] {
	return reference_data.ConsumerResult[reference_data.PackSet]{Value: &reference_data.PackSet{ID: "rpset_" + strings.Repeat("0", 64)}}
}
func (IndicatorRegistry) EvaluateIndicatorValue(_ context.Context, request reference_data.EvaluateIndicatorRequest) reference_data.ConsumerResult[reference_data.IndicatorEvaluation] {
	value, err := EvaluateIndicator(request.IndicatorTypeID, request.ValueKind, request.RawValue)
	if err != nil {
		var rejected *reference_data.ConsumerError
		if errors.As(err, &rejected) {
			return reference_data.ConsumerResult[reference_data.IndicatorEvaluation]{Error: rejected}
		}
		return reference_data.ConsumerResult[reference_data.IndicatorEvaluation]{Error: &reference_data.ConsumerError{Code: "pack_unavailable"}}
	}
	return reference_data.ConsumerResult[reference_data.IndicatorEvaluation]{Value: &value}
}
func (IndicatorRegistry) GetPackEntry(context.Context, reference_data.GetPackEntryRequest) reference_data.ConsumerResult[reference_data.PackEntry] {
	return reference_data.ConsumerResult[reference_data.PackEntry]{Error: &reference_data.ConsumerError{Code: "entry_not_found"}}
}
func (IndicatorRegistry) LookupPackEntries(context.Context, reference_data.LookupPackEntriesRequest) reference_data.ConsumerResult[reference_data.PackEntryPage] {
	return reference_data.ConsumerResult[reference_data.PackEntryPage]{Error: &reference_data.ConsumerError{Code: "pack_unavailable"}}
}
func (IndicatorRegistry) GetPackProvenance(context.Context, reference_data.GetPackProvenanceRequest) reference_data.ConsumerResult[reference_data.PackProvenance] {
	return reference_data.ConsumerResult[reference_data.PackProvenance]{Error: &reference_data.ConsumerError{Code: "pack_unavailable"}}
}
