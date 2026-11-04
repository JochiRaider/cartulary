package graphprojection

import (
	"encoding/json"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
)

var projectedMergeMatrixV2 = map[string]map[string]struct{}{
	"boolean":          mergeSetV2("single_value", "first", "last"),
	"integer":          mergeSetV2("single_value", "first", "last", "min", "max", "sum", "count"),
	"number":           mergeSetV2("single_value", "first", "last", "min", "max", "sum"),
	"string":           mergeSetV2("single_value", "first", "last", "min", "max"),
	"timestamp":        mergeSetV2("single_value", "first", "last", "min", "max"),
	"identifier":       mergeSetV2("single_value", "first", "last", "min", "max"),
	"boolean_array":    mergeSetV2("single_value", "first", "last", "set", "ordered_list"),
	"integer_array":    mergeSetV2("single_value", "first", "last", "set", "ordered_list"),
	"number_array":     mergeSetV2("single_value", "first", "last", "set", "ordered_list"),
	"string_array":     mergeSetV2("single_value", "first", "last", "set", "ordered_list"),
	"timestamp_array":  mergeSetV2("single_value", "first", "last", "set", "ordered_list"),
	"identifier_array": mergeSetV2("single_value", "first", "last", "set", "ordered_list"),
}

func mergeSetV2(values ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func validProjectedType(value string) bool {
	_, ok := projectedMergeMatrixV2[value]
	return ok
}

func validProjectedMergeV2(projectedType, mergeBehavior string) bool {
	_, ok := projectedMergeMatrixV2[projectedType][mergeBehavior]
	return ok
}

func normalizeProjectedValueV2(projectedType string, value any) (any, bool) {
	if strings.HasSuffix(projectedType, "_array") {
		return normalizeProjectedArrayV2(strings.TrimSuffix(projectedType, "_array"), value)
	}
	return normalizeProjectedScalarV2(projectedType, value)
}

func normalizeProjectedArrayV2(elementType string, value any) (any, bool) {
	var values []any
	switch typed := value.(type) {
	case []any:
		values = typed
	case []string:
		values = make([]any, len(typed))
		for index, item := range typed {
			values[index] = item
		}
	default:
		return nil, false
	}
	if len(values) > graphProjectionLimits.MaxArrayItems {
		return nil, false
	}
	normalized := make([]any, len(values))
	for index, item := range values {
		if item == nil {
			return nil, false
		}
		value, ok := normalizeProjectedScalarV2(elementType, item)
		if !ok {
			return nil, false
		}
		normalized[index] = value
	}
	return normalized, true
}

func normalizeProjectedScalarV2(projectedType string, value any) (any, bool) {
	switch projectedType {
	case "boolean":
		normalized, ok := value.(bool)
		return normalized, ok
	case "integer":
		switch typed := value.(type) {
		case int:
			return int64(typed), true
		case int64:
			return typed, true
		case json.Number:
			if strings.ContainsAny(typed.String(), ".eE") {
				return nil, false
			}
			normalized, err := strconv.ParseInt(typed.String(), 10, 64)
			return normalized, err == nil
		default:
			return nil, false
		}
	case "number":
		switch typed := value.(type) {
		case int:
			return int64(typed), true
		case int64:
			return typed, true
		case float64:
			return typed, !math.IsNaN(typed) && !math.IsInf(typed, 0)
		case json.Number:
			normalized, err := normalizedNumberV2(typed)
			return normalized, err == nil
		default:
			return nil, false
		}
	case "string":
		normalized, ok := value.(string)
		return normalized, ok && len(normalized) <= graphProjectionLimits.MaxStringBytes
	case "timestamp":
		normalized, ok := value.(string)
		if !ok || len(normalized) > graphProjectionLimits.MaxStringBytes {
			return nil, false
		}
		_, err := parseTimestamp(normalized)
		return normalized, err == nil
	case "identifier":
		normalized, ok := value.(string)
		return normalized, ok && validIdentifierV2(normalized)
	default:
		return nil, false
	}
}

// Array mergers consume one normalized contributor at a time. Only distinct
// set keys or admitted list elements survive the contributor iteration.
type projectedMerge struct {
	behavior string
	values   []any
	unique   map[string]any
	seen     bool
	conflict bool
}

func (merge *projectedMerge) add(value any) error {
	merge.seen = true
	if merge.behavior != "ordered_list" && merge.behavior != "set" {
		merge.values = append(merge.values, value)
		return nil
	}
	items, ok := value.([]any)
	if !ok {
		merge.conflict = true
		return nil
	}
	limit := graphProjectionLimits.MaxArrayItems
	if merge.behavior == "ordered_list" {
		if len(items) > limit-len(merge.values) {
			return resourceProjectionErrorV2("maximum_array_items", limit, limit+1)
		}
		merge.values = append(merge.values, items...)
		return nil
	}
	if merge.unique == nil {
		merge.unique = make(map[string]any)
	}
	for _, item := range items {
		key := canonicalValueKey(item)
		if _, exists := merge.unique[key]; !exists && len(merge.unique) == limit {
			return resourceProjectionErrorV2("maximum_array_items", limit, limit+1)
		}
		merge.unique[key] = item
	}
	return nil
}

func (merge *projectedMerge) result(projectedType string) (any, bool, bool) {
	behavior, values := merge.behavior, merge.values
	if !validProjectedMergeV2(projectedType, behavior) || !merge.seen {
		return nil, false, false
	}
	if merge.conflict {
		return nil, false, true
	}
	switch behavior {
	case "first":
		return values[0], true, false
	case "last":
		return values[len(values)-1], true, false
	case "single_value":
		first := canonicalValueKey(values[0])
		for _, value := range values[1:] {
			if canonicalValueKey(value) != first {
				return nil, false, true
			}
		}
		return values[0], true, false
	case "count":
		for _, value := range values {
			if _, ok := value.(int64); !ok {
				return nil, false, true
			}
		}
		return int64(len(values)), true, false
	case "min", "max":
		selected := values[0]
		for _, value := range values[1:] {
			comparison, ok := compareProjectedScalarsV2(projectedType, value, selected)
			if !ok {
				return nil, false, true
			}
			if behavior == "min" && comparison < 0 || behavior == "max" && comparison > 0 {
				selected = value
			}
		}
		return selected, true, false
	case "sum":
		if projectedType == "integer" {
			var sum int64
			for _, value := range values {
				integer, ok := value.(int64)
				if !ok || integer > 0 && sum > math.MaxInt64-integer || integer < 0 && sum < math.MinInt64-integer {
					return nil, false, true
				}
				sum += integer
			}
			return sum, true, false
		}
		var sum float64
		for _, value := range values {
			number, ok := projectedNumberFloat64V2(value)
			if !ok {
				return nil, false, true
			}
			sum += number
			if math.IsNaN(sum) || math.IsInf(sum, 0) {
				return nil, false, true
			}
		}
		if sum == 0 {
			sum = 0
		}
		return sum, true, false
	case "set":
		unique := merge.unique
		keys := make([]string, 0, len(unique))
		for key := range unique {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		merged := make([]any, 0, len(keys))
		for _, key := range keys {
			merged = append(merged, unique[key])
		}
		return merged, true, false
	case "ordered_list":
		if values == nil {
			return []any{}, true, false
		}
		return values, true, false
	default:
		return nil, false, false
	}
}

func compareProjectedScalarsV2(projectedType string, left, right any) (int, bool) {
	switch projectedType {
	case "integer":
		leftValue, leftOK := left.(int64)
		rightValue, rightOK := right.(int64)
		if !leftOK || !rightOK {
			return 0, false
		}
		if leftValue < rightValue {
			return -1, true
		}
		if leftValue > rightValue {
			return 1, true
		}
		return 0, true
	case "number":
		leftValue, leftOK := projectedNumberRatV2(left)
		rightValue, rightOK := projectedNumberRatV2(right)
		if !leftOK || !rightOK {
			return 0, false
		}
		return leftValue.Cmp(rightValue), true
	case "timestamp":
		leftValue, leftOK := left.(string)
		rightValue, rightOK := right.(string)
		if !leftOK || !rightOK {
			return 0, false
		}
		leftTime, leftErr := parseTimestamp(leftValue)
		rightTime, rightErr := parseTimestamp(rightValue)
		if leftErr != nil || rightErr != nil {
			return 0, false
		}
		if leftTime.Before(rightTime) {
			return -1, true
		}
		if leftTime.After(rightTime) {
			return 1, true
		}
		return 0, true
	case "string", "identifier":
		leftValue, leftOK := left.(string)
		rightValue, rightOK := right.(string)
		if !leftOK || !rightOK {
			return 0, false
		}
		return strings.Compare(leftValue, rightValue), true
	default:
		return 0, false
	}
}

func projectedNumberRatV2(value any) (*big.Rat, bool) {
	switch typed := value.(type) {
	case int64:
		return new(big.Rat).SetInt64(typed), true
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return nil, false
		}
		return new(big.Rat).SetFloat64(typed), true
	default:
		return nil, false
	}
}

func projectedNumberFloat64V2(value any) (float64, bool) {
	var number float64
	switch typed := value.(type) {
	case int64:
		number = float64(typed)
	case float64:
		number = typed
	default:
		return 0, false
	}
	return number, !math.IsNaN(number) && !math.IsInf(number, 0)
}
