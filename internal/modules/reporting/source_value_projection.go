package reporting

import (
	"regexp"
	"sort"
)

var reportingMemberName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,63}$`)

// Source providers may contribute a declared record-value object. That object
// is a source container, never a canonical Reporting field value. Project its
// immediate, named attributes before redaction so each value has its own exact
// selector, ordinal, classification and source binding. Nested objects are not
// recursively flattened or stringified.
func projectReportingSourceFields(source ExportField) ([]ReportingField, []string, error) {
	support, err := materializeReportingRefs(source.SupportRefs)
	if err != nil {
		return nil, nil, err
	}
	partitions := orderedDisclosurePartitions(source.DisclosurePartitionRefs)
	if validateDisclosurePartitionSet(partitions) != nil {
		return nil, nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
	}
	values, object := source.Value.(map[string]any)
	keys := []string{"value"}
	if object {
		keys = make([]string, 0, len(values))
		for key := range values {
			if !reportingMemberName.MatchString(key) {
				return nil, nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
			}
			keys = append(keys, key)
		}
		if len(keys) == 0 {
			return nil, nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
		}
		sort.Strings(keys)
	}
	fields := make([]ReportingField, 0, len(keys))
	refs := make([]string, 0, len(keys))
	for i, key := range keys {
		value, selector := source.Value, source.Path
		if object {
			value, selector = values[key], source.Path+"/"+key
		}
		value, valid := admitReportingFieldValue(value, true)
		if !valid {
			return nil, nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
		}
		fields = append(fields, ReportingField{
			SchemaID: "cartulary.reporting_field.v1", FieldKey: key, FieldOrdinal: i + 1,
			SourceValueState: sourceValueState(value), RedactedValueState: "unchanged", Value: value,
			SourceRefs: []string{selector}, SupportRefs: support,
			DisclosurePartitionRefs: partitions,
		})
		refs = append(refs, selector)
	}
	return fields, refs, nil
}

// Source-owner lists may repeat the same support target for distinct links.
// Only materialization coalesces those occurrences into canonical references;
// frozen-model readers never repair a retained null, duplicate or invalid ref.
func materializeReportingRefs(source []string) ([]string, error) {
	refs := make([]string, 0, len(source))
	seen := map[string]bool{}
	for _, ref := range source {
		if ref == "" || !validReportingText(ref, false) {
			return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
		}
		if !seen[ref] {
			refs = append(refs, ref)
			seen[ref] = true
		}
	}
	sort.Strings(refs)
	return refs, nil
}
