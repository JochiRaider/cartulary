package identity

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
)

type Input struct {
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
}

type Canonical struct {
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DedupeKey       string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
}

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid indicator identity: %s %s", e.Field, e.ReasonCode)
}

// Evaluator binds one exact registry set for an operation. Identity owns record
// representation checks; Reference Data owns all value algorithms and dedupe.
type Evaluator func(string, string, string) (reference_data.IndicatorEvaluation, error)

type Consumer interface {
	ResolveCurrentPackSet(context.Context) reference_data.ConsumerResult[reference_data.PackSet]
	EvaluateIndicatorValue(context.Context, reference_data.EvaluateIndicatorRequest) reference_data.ConsumerResult[reference_data.IndicatorEvaluation]
}

func FromConsumer(ctx context.Context, consumer Consumer) (Evaluator, error) {
	if consumer == nil {
		return nil, errors.New("indicator evaluator unavailable")
	}
	set := consumer.ResolveCurrentPackSet(ctx)
	if set.Error != nil {
		return nil, set.Error
	}
	if set.Value == nil {
		return nil, errors.New("indicator set unavailable")
	}
	return func(typ, kind, raw string) (reference_data.IndicatorEvaluation, error) {
		result := consumer.EvaluateIndicatorValue(ctx, reference_data.EvaluateIndicatorRequest{PackSetID: set.Value.ID, IndicatorTypeID: typ, ValueKind: kind, RawValue: raw})
		if result.Error != nil {
			return reference_data.IndicatorEvaluation{}, result.Error
		}
		if result.Value == nil {
			return reference_data.IndicatorEvaluation{}, errors.New("indicator evaluation unavailable")
		}
		return *result.Value, nil
	}, nil
}

func ValidateShape(input Input) error {
	for _, required := range []struct{ field, value string }{{"indicator_type", input.IndicatorType}, {"value_kind", input.ValueKind}, {"display_value", input.DisplayValue}} {
		if strings.TrimSpace(required.value) == "" {
			return invalid(required.field, "missing_required_field")
		}
	}
	if !vocabulary.IsIndicatorType(input.IndicatorType) {
		return invalid("indicator_type", "invalid_value")
	}
	if !vocabulary.IsValueKind(input.ValueKind) {
		return invalid("value_kind", "invalid_value")
	}
	if !utf8.ValidString(input.DisplayValue) || utf8.RuneCountInString(input.DisplayValue) > 8192 {
		return invalid("display_value", "invalid_value")
	}
	if input.NormalizedValue != nil && (!utf8.ValidString(*input.NormalizedValue) || utf8.RuneCountInString(*input.NormalizedValue) > 8192) {
		return invalid("normalized_value", "invalid_value")
	}
	return nil
}

func Canonicalize(evaluate Evaluator, input Input) (Canonical, error) {
	if err := ValidateShape(input); err != nil {
		return Canonical{}, err
	}
	if evaluate == nil {
		return Canonical{}, errors.New("indicator evaluator unavailable")
	}
	result, err := evaluate(input.IndicatorType, input.ValueKind, input.DisplayValue)
	if err != nil {
		var rejected *reference_data.ConsumerError
		if errors.As(err, &rejected) {
			switch rejected.Code {
			case "indicator_type_unsupported":
				return Canonical{}, invalid("indicator_type", "invalid_value")
			case "indicator_value_kind_unsupported":
				return Canonical{}, invalid("value_kind", "invalid_value")
			}
		}
		return Canonical{}, err
	}
	if !result.Valid || result.Display == nil || result.Normalized == nil || result.Defanged == nil || result.DedupeKey == nil {
		return Canonical{}, invalid("display_value", "invalid_value")
	}
	if input.NormalizedValue != nil {
		supplied, err := evaluate(input.IndicatorType, input.ValueKind, *input.NormalizedValue)
		if err != nil {
			return Canonical{}, err
		}
		if !supplied.Valid || supplied.Normalized == nil || *supplied.Normalized != *result.Normalized {
			return Canonical{}, invalid("normalized_value", "invalid_value")
		}
	}
	algorithm, hash, err := normalizeHashPair(input.HashAlgorithm, input.HashValue)
	if err != nil {
		return Canonical{}, err
	}
	if isIPType(input.IndicatorType) && (algorithm != nil || hash != nil) {
		return Canonical{}, invalid("hash_algorithm", "invalid_value")
	}
	defanged := result.Defanged
	if input.DefangedValue != nil {
		defanged = input.DefangedValue
	}
	return Canonical{IndicatorType: result.Type, ValueKind: result.Kind, DisplayValue: *result.Display, NormalizedValue: cloneString(result.Normalized), DedupeKey: *result.DedupeKey, DefangedValue: cloneString(defanged), HashAlgorithm: algorithm, HashValue: hash, STIXPattern: cloneString(input.STIXPattern)}, nil
}

func NormalizeObservationCandidate(evaluate Evaluator, parsedType *string, normalizedCandidate *string, observedText string) (*string, *string, error) {
	if evaluate == nil {
		return nil, nil, errors.New("indicator evaluator unavailable")
	}
	if parsedType != nil {
		if !vocabulary.IsIndicatorType(*parsedType) {
			return nil, nil, invalid("indicator_type", "invalid_value")
		}
		raw := observedText
		if normalizedCandidate != nil {
			raw = *normalizedCandidate
		}
		result, err := evaluate(*parsedType, "atomic", raw)
		if err != nil {
			return nil, nil, err
		}
		if !result.Valid {
			return nil, nil, invalid("normalized_candidate", "invalid_value")
		}
		return cloneString(parsedType), cloneString(result.Normalized), nil
	}
	if normalizedCandidate != nil {
		return nil, nil, invalid("normalized_candidate", "invalid_value")
	}
	// Classification precedence is fixed. Every candidate is validated by the
	// same registry algorithms used for explicit record creation.
	for _, typ := range []string{"ipv4_addr", "ipv6_addr", "url", "sha256", "email_addr", "domain_name"} {
		result, err := evaluate(typ, "atomic", observedText)
		if err != nil {
			return nil, nil, err
		}
		if result.Valid {
			return stringPointer(typ), cloneString(result.Normalized), nil
		}
	}
	return nil, nil, nil
}

func normalizeHashPair(algorithm, value *string) (*string, *string, error) {
	if algorithm == nil && value == nil {
		return nil, nil, nil
	}
	if algorithm == nil || value == nil {
		return nil, nil, invalid("hash_value", "invalid_value")
	}
	a, v := strings.ToLower(strings.TrimSpace(*algorithm)), strings.ToLower(strings.TrimSpace(*value))
	if a == "" || v == "" {
		return nil, nil, invalid("hash_value", "invalid_value")
	}
	for _, c := range v {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return nil, nil, invalid("hash_value", "invalid_value")
		}
	}
	return &a, &v, nil
}
func invalid(field, code string) *ValidationError {
	return &ValidationError{Field: field, ReasonCode: code}
}
func isIPType(typ string) bool { return typ == "ipv4_addr" || typ == "ipv6_addr" }
func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
func stringPointer(value string) *string { return &value }
