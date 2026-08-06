package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	collaborationIndexSchemaID    = "cartulary.collaboration_contract_index.v2"
	collaborationRegistrySchemaID = "cartulary.operator.collaboration_requeue_registry.v2"
	collaborationResultSchemaID   = "cartulary.operator.collaboration_requeue_result.v2"
)

var (
	canonicalUUIDPattern       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	collaborationResultMembers = []string{
		"schema_id", "operation_id", "operation", "result", "started_at",
		"completed_at", "incident_id", "requeued_intent_count", "error",
	}
	collaborationFailureRegistry = map[string]struct {
		exit    int
		reasons []string
	}{
		"invalid_operator_request":       {2, []string{"missing_required_flag", "invalid_flag_value", "duplicate_flag", "unknown_flag", "unexpected_argument", "local_config_invalid"}},
		"collaboration_requeue_rejected": {3, []string{"incident_not_quarantined", "repair_not_verified"}},
		"collaboration_requeue_failed":   {4, []string{"postgres_unavailable", "transaction_failed", "commit_outcome_unknown"}},
		"operation_timed_out":            {4, []string{"timeout_elapsed"}},
		"operation_cancelled":            {4, []string{"caller_cancelled"}},
	}
)

func validateCollaborationContractInput(relativePath string, value any) error {
	switch relativePath {
	case "index.json":
		return validateCollaborationIndex(value)
	case "operator-requeue-registry.v2.json":
		return validateCollaborationRegistry(value)
	case "operator-requeue-result.v2.schema.json":
		object, err := asObject(value, "contracts/collaboration/operator-requeue-result.v2.schema.json")
		if err != nil {
			return err
		}
		if err := requireDraftSchema(object, "contracts/collaboration/operator-requeue-result.v2.schema.json"); err != nil {
			return err
		}
		id, err := requiredString(object, "$id", "collaboration result schema")
		if err != nil {
			return err
		}
		if id != collaborationResultSchemaID {
			return fmt.Errorf("collaboration result schema $id must be %s", collaborationResultSchemaID)
		}
		return nil
	case "fixtures/operator-requeue-result.v2.success.json":
		return validateCollaborationResultFixture(value, true)
	case "fixtures/operator-requeue-result.v2.failure.json":
		return validateCollaborationResultFixture(value, false)
	case "fixtures/operator-requeue-negative.v2.json":
		return validateCollaborationNegativeFixtures(value)
	default:
		return fmt.Errorf("unexpected collaboration artifact %s", relativePath)
	}
}

func validateCollaborationIndex(value any) error {
	object, err := asObject(value, "contracts/collaboration/index.json")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet(
		"$schema", "schema_id", "family_id", "contract_major", "owner_requirements",
		"current_schema_ids", "historical_reader_schema_ids", "contract_files", "fixtures", "compatibility_policy",
	), "contracts/collaboration/index.json"); err != nil {
		return err
	}
	if err := requireDraftSchema(object, "contracts/collaboration/index.json"); err != nil {
		return err
	}
	schemaID, err := requiredString(object, "schema_id", "collaboration index")
	if err != nil || schemaID != collaborationIndexSchemaID {
		return fmt.Errorf("collaboration index schema_id must be %s", collaborationIndexSchemaID)
	}
	familyID, err := requiredString(object, "family_id", "collaboration index")
	if err != nil || familyID != "collaboration" {
		return fmt.Errorf("collaboration index family_id must be collaboration")
	}
	current, err := stringArray(object["current_schema_ids"], "collaboration index current_schema_ids", true)
	if err != nil || len(current) != 1 || current[0] != collaborationResultSchemaID {
		return fmt.Errorf("collaboration index must declare only the v2 result schema as current")
	}
	historical, err := stringArray(object["historical_reader_schema_ids"], "collaboration index historical_reader_schema_ids", false)
	if err != nil || len(historical) != 0 {
		return fmt.Errorf("collaboration index must declare no historical reader")
	}
	policy, err := asObject(object["compatibility_policy"], "collaboration compatibility_policy")
	if err != nil {
		return err
	}
	for _, key := range []string{"v1_reader", "parser_aliases", "dual_output"} {
		value, ok := policy[key].(bool)
		if !ok || value {
			return fmt.Errorf("collaboration compatibility_policy.%s must be false", key)
		}
	}
	return nil
}

func validateCollaborationRegistry(value any) error {
	object, err := asObject(value, "collaboration requeue registry")
	if err != nil {
		return err
	}
	schemaID, err := requiredString(object, "schema_id", "collaboration registry")
	if err != nil || schemaID != collaborationRegistrySchemaID {
		return fmt.Errorf("collaboration registry schema_id must be %s", collaborationRegistrySchemaID)
	}
	operation, err := requiredString(object, "operation", "collaboration registry")
	if err != nil || operation != "collaboration_requeue" {
		return fmt.Errorf("collaboration registry operation must be collaboration_requeue")
	}
	tokens, err := stringArray(object["command_tokens"], "collaboration registry command_tokens", true)
	if err != nil || strings.Join(tokens, " ") != "operator collaboration requeue" {
		return fmt.Errorf("collaboration registry command_tokens are not canonical")
	}
	memberOrder, err := stringArray(object["result_member_order"], "collaboration registry result_member_order", true)
	if err != nil || strings.Join(memberOrder, "\x00") != strings.Join(collaborationResultMembers, "\x00") {
		return fmt.Errorf("collaboration registry result_member_order is not canonical")
	}
	resultSchemaID, err := requiredString(object, "result_schema_id", "collaboration registry")
	if err != nil || resultSchemaID != collaborationResultSchemaID {
		return fmt.Errorf("collaboration registry result_schema_id must be %s", collaborationResultSchemaID)
	}
	failures, err := objectArray(object["failures"], "collaboration registry failures")
	if err != nil || len(failures) != len(collaborationFailureRegistry) {
		return fmt.Errorf("collaboration registry must contain exactly %d failure families", len(collaborationFailureRegistry))
	}
	seen := map[string]struct{}{}
	for index, failure := range failures {
		label := fmt.Sprintf("collaboration registry failures[%d]", index+1)
		code, err := requiredString(failure, "code", label)
		if err != nil {
			return err
		}
		expected, ok := collaborationFailureRegistry[code]
		if !ok {
			return fmt.Errorf("%s has unknown code %s", label, code)
		}
		if _, duplicate := seen[code]; duplicate {
			return fmt.Errorf("collaboration registry duplicates code %s", code)
		}
		seen[code] = struct{}{}
		exit, ok := jsonInteger(failure["exit"])
		if !ok || exit != expected.exit {
			return fmt.Errorf("%s exit is not canonical", label)
		}
		reasons, err := stringArray(failure["reason_codes"], label+".reason_codes", true)
		if err != nil || strings.Join(reasons, "\x00") != strings.Join(expected.reasons, "\x00") {
			return fmt.Errorf("%s reason_codes are not canonical", label)
		}
	}
	return nil
}

func validateCollaborationResultFixture(value any, success bool) error {
	object, err := asObject(value, "collaboration result fixture")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(object, stringSet(collaborationResultMembers...), "collaboration result fixture"); err != nil {
		return err
	}
	if len(object) != len(collaborationResultMembers) {
		return fmt.Errorf("collaboration result fixture must contain exactly %d members", len(collaborationResultMembers))
	}
	schemaID, err := requiredString(object, "schema_id", "collaboration result fixture")
	if err != nil || schemaID != collaborationResultSchemaID {
		return fmt.Errorf("collaboration result fixture schema_id must be v2")
	}
	for _, field := range []string{"operation_id", "incident_id"} {
		value, err := requiredString(object, field, "collaboration result fixture")
		if err != nil || !validCanonicalNonzeroUUID(value) {
			return fmt.Errorf("collaboration result fixture %s must be a canonical non-zero UUID", field)
		}
	}
	for _, field := range []string{"started_at", "completed_at"} {
		value, err := requiredString(object, field, "collaboration result fixture")
		if err != nil || !validUTCTimestamp(value) {
			return fmt.Errorf("collaboration result fixture %s must be an RFC3339 UTC timestamp", field)
		}
	}
	result, err := requiredString(object, "result", "collaboration result fixture")
	if err != nil {
		return err
	}
	if success {
		count, ok := jsonInteger(object["requeued_intent_count"])
		if result != "succeeded" || !ok || count < 0 || object["error"] != nil {
			return fmt.Errorf("collaboration success fixture has inconsistent terminal fields")
		}
		return nil
	}
	if result != "failed" || object["requeued_intent_count"] != nil {
		return fmt.Errorf("collaboration failure fixture has inconsistent terminal fields")
	}
	errorObject, err := asObject(object["error"], "collaboration result error")
	if err != nil {
		return err
	}
	if err := requireAllowedKeys(errorObject, stringSet("code", "reason_code", "message"), "collaboration result error"); err != nil {
		return err
	}
	code, err := requiredString(errorObject, "code", "collaboration result error")
	if err != nil {
		return err
	}
	reason, err := requiredString(errorObject, "reason_code", "collaboration result error")
	if err != nil {
		return err
	}
	expected, ok := collaborationFailureRegistry[code]
	if !ok || !containsString(expected.reasons, reason) {
		return fmt.Errorf("collaboration result error code/reason pair is not registered")
	}
	message, err := requiredString(errorObject, "message", "collaboration result error")
	if err != nil || len(message) > 256 {
		return fmt.Errorf("collaboration result error message is invalid")
	}
	return nil
}

func validateCollaborationNegativeFixtures(value any) error {
	object, err := asObject(value, "collaboration negative fixtures")
	if err != nil {
		return err
	}
	schemaID, err := requiredString(object, "schema_id", "collaboration negative fixtures")
	if err != nil || schemaID != "cartulary.operator.collaboration_requeue_negative_fixtures.v2" {
		return fmt.Errorf("collaboration negative fixture schema_id is invalid")
	}
	grammar, err := objectArray(object["grammar_cases"], "collaboration negative grammar_cases")
	if err != nil || len(grammar) != 18 {
		return fmt.Errorf("collaboration negative fixtures must contain exactly 18 grammar cases")
	}
	results, err := objectArray(object["result_cases"], "collaboration negative result_cases")
	if err != nil || len(results) != 10 {
		return fmt.Errorf("collaboration negative fixtures must contain exactly 10 result cases")
	}
	caseIDs := map[string]struct{}{}
	for _, cases := range [][]map[string]any{grammar, results} {
		for _, current := range cases {
			caseID, err := requiredString(current, "case_id", "collaboration negative fixture")
			if err != nil {
				return err
			}
			if _, duplicate := caseIDs[caseID]; duplicate {
				return fmt.Errorf("collaboration negative fixture duplicates case_id %s", caseID)
			}
			caseIDs[caseID] = struct{}{}
		}
	}
	return nil
}

func validateCollaborationContractFamily(root string) error {
	base := filepath.Join(root, "contracts", "collaboration")
	expected := []string{
		"fixtures/operator-requeue-negative.v2.json",
		"fixtures/operator-requeue-result.v2.failure.json",
		"fixtures/operator-requeue-result.v2.success.json",
		"index.json",
		"operator-requeue-registry.v2.json",
		"operator-requeue-result.v2.schema.json",
	}
	actual := []string{}
	if err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		actual = append(actual, filepath.ToSlash(relative))
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(actual)
	if err := compareStringSlices(expected, actual, "collaboration contract paths"); err != nil {
		return err
	}
	for _, fixture := range []string{
		"fixtures/operator-requeue-result.v2.success.json",
		"fixtures/operator-requeue-result.v2.failure.json",
	} {
		raw, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(fixture)))
		if err != nil {
			return err
		}
		if err := requireJSONMemberOrder(raw, collaborationResultMembers); err != nil {
			return fmt.Errorf("%s: %w", fixture, err)
		}
	}
	return nil
}

func requireJSONMemberOrder(raw []byte, members []string) error {
	position := -1
	for _, member := range members {
		index := bytes.Index(raw, []byte(`"`+member+`"`))
		if index <= position {
			return fmt.Errorf("member %s is absent or out of canonical order", member)
		}
		position = index
	}
	return nil
}

func jsonInteger(value any) (int, bool) {
	number, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	parsed, err := number.Int64()
	return int(parsed), err == nil && int64(int(parsed)) == parsed
}

func validCanonicalNonzeroUUID(value string) bool {
	return value != "00000000-0000-0000-0000-000000000000" && canonicalUUIDPattern.MatchString(value)
}

func validUTCTimestamp(value string) bool {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return err == nil && strings.HasSuffix(value, "Z") && parsed.Location() == time.UTC
}
