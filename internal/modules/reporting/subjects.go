package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"slices"

	strictjson "github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const tokenizableSubjectSchemaID = "cartulary.tokenizable_subject.v1"

// The canonical subject carries identity and a display-value digest. The raw
// display value stays in private source content and the sensitive reveal map;
// it must not become a second unchecked presentation copy in subjects[].
func materializeReportingSubject(record reportingSourceGroup, family string) (TokenizableSubject, error) {
	id := record.SourceRecordRef.SourceRecordID
	subject := TokenizableSubject{
		SchemaID: tokenizableSubjectSchemaID, SubjectFamily: family,
		StableSubjectRef: canonicalSubjectRef(family, id), SourceRecordID: &id, SourceRecordType: &family,
		DisclosurePartitionRefs: orderedDisclosurePartitions(record.DisclosurePartitionRefs),
		SourceRefs:              append([]string{}, record.SourceRefs...), SupportRefs: append([]string{}, record.SupportRefs...),
	}
	if record.SourceRecordRef.SourceFamily == "entity_mention" {
		subject.StableSubjectRef = "subject:" + family + ":mention:" + id
		subject.SourceRecordID, subject.SourceRecordType, subject.EntityMentionID = nil, nil, &id
	}
	digest, err := subjectDisplayValueDigest(subject.StableSubjectRef, record.DisplayName)
	subject.CanonicalDisplayValueSHA256 = digest
	return subject, err
}

func subjectDisplayValueDigest(ref string, value *string) (string, error) {
	raw, err := canonicalJSON(struct {
		SchemaID         string  `json:"schema_id"`
		StableSubjectRef string  `json:"stable_subject_ref"`
		Value            *string `json:"value"`
	}{tokenizableSubjectSchemaID, ref, value})
	if err != nil {
		return "", err
	}
	return hashHex(raw), nil
}

// Resolve only the closed identity union. Neither a raw value nor a source
// pointer spelling can turn an unresolved mention into a canonical subject.
func subjectSourceIdentity(subject TokenizableSubject) (family, id string, err error) {
	invalid := newRenderValidationError("token_manifest_invalid", "token_subject_not_stable")
	if subject.SchemaID != tokenizableSubjectSchemaID || (subject.SubjectFamily != "host" && subject.SubjectFamily != "identity" && subject.SubjectFamily != "party") {
		return "", "", invalid
	}
	if subject.SourceRecordID != nil && subject.SourceRecordType != nil && subject.EntityMentionID == nil {
		family, id = *subject.SourceRecordType, *subject.SourceRecordID
		if family != subject.SubjectFamily || !validReportingIdentifier(id) || subject.StableSubjectRef != canonicalSubjectRef(family, id) {
			return "", "", invalid
		}
	} else if subject.SourceRecordID == nil && subject.SourceRecordType == nil && subject.EntityMentionID != nil && subject.SubjectFamily != "party" {
		family, id = "entity_mention", *subject.EntityMentionID
		if !validReportingIdentifier(id) || subject.StableSubjectRef != "subject:"+subject.SubjectFamily+":mention:"+id {
			return "", "", invalid
		}
	} else {
		return "", "", invalid
	}
	return family, id, nil
}

func validateSnapshotSubjects(model SnapshotContent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	groups, err := snapshotSourceGroups(model)
	if err != nil {
		return err
	}
	records := map[[2]string]reportingSourceGroup{}
	ambiguous := map[[2]string]bool{}
	originRecords := map[string]bool{}
	for _, record := range groups {
		key := [2]string{record.SourceRecordRef.SourceFamily, record.SourceRecordRef.SourceRecordID}
		if _, exists := records[key]; exists {
			ambiguous[key] = true
		}
		records[key] = record
		if key[0] == "record_envelope" || coreTypeForSourceFamily(key[0]) != "" {
			originRecords[key[1]] = true
		}
	}
	seen := map[string]bool{}
	for _, subject := range model.Subjects {
		family, id, err := subjectSourceIdentity(subject)
		if err != nil {
			issues.add(err)
			continue
		}
		if seen[subject.StableSubjectRef] {
			issues.add(newRenderValidationError("token_manifest_invalid", "token_subject_not_unique"))
			continue
		}
		seen[subject.StableSubjectRef] = true
		// The record binding validator owns the duplicate error. Never choose
		// one of the ambiguous values to produce a traversal-dependent digest.
		if ambiguous[[2]string{family, id}] {
			continue
		}
		record, exists := records[[2]string{family, id}]
		if !exists {
			issues.add(newRenderValidationError("export_model_invalid", "dangling_source_ref"))
			continue
		}
		digest, err := subjectDisplayValueDigest(subject.StableSubjectRef, record.DisplayName)
		if err != nil || digest != subject.CanonicalDisplayValueSHA256 ||
			!slices.Equal(subject.SourceRefs, record.SourceRefs) || !slices.Equal(subject.SupportRefs, record.SupportRefs) ||
			!slices.Equal(subject.DisclosurePartitionRefs, record.DisclosurePartitionRefs) || subject.SourceRefs == nil || subject.SupportRefs == nil || subject.DisclosurePartitionRefs == nil {
			issues.add(newRenderValidationError("export_model_invalid", "invalid_scalar_value"))
		}
	}
	for _, record := range groups {
		family := record.SourceRecordRef.SourceFamily
		switch family {
		case "host", "identity", "party":
			if !seen[canonicalSubjectRef(family, record.RecordID)] {
				issues.add(newRenderValidationError("token_manifest_invalid", "token_subject_not_stable"))
			}
		case "entity_mention":
			binding, err := reportingMentionSource(record)
			if err != nil {
				issues.add(err)
				continue
			}
			if !originRecords[binding.OriginRecordID] {
				issues.add(newRenderValidationError("export_model_invalid", "dangling_source_ref"))
			}
			mentionRef := "subject:" + binding.Family + ":mention:" + record.RecordID
			if binding.ResolvedRecordID == nil {
				if !seen[mentionRef] {
					issues.add(newRenderValidationError("token_manifest_invalid", "token_subject_not_stable"))
				}
			} else if seen[mentionRef] || !seen[canonicalSubjectRef(binding.Family, *binding.ResolvedRecordID)] {
				issues.add(newRenderValidationError("token_manifest_invalid", "token_subject_not_stable"))
			}
		}
	}
	return issues.finish()
}

func (subject *TokenizableSubject) UnmarshalJSON(data []byte) error {
	decoded, err := strictjson.DecodeStrict(data)
	if err != nil {
		return err
	}
	object, ok := decoded.(map[string]any)
	if !ok || len(object) != 10 {
		return errors.New("invalid subject object")
	}
	for _, key := range []string{"schema_id", "stable_subject_ref", "subject_family", "source_record_id", "source_record_type", "entity_mention_id", "canonical_display_value_sha256", "disclosure_partition_refs", "source_refs", "support_refs"} {
		value, exists := object[key]
		if !exists || value == nil && key != "source_record_id" && key != "source_record_type" && key != "entity_mention_id" {
			return errors.New("invalid subject member")
		}
	}
	type wire TokenizableSubject
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*subject = TokenizableSubject(value)
	return nil
}
