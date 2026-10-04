package packformat

import (
	"bytes"
	"encoding/json"
	"errors"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type consumerContract struct {
	request, result *shape
	order           []string
}

var consumerContracts = loadConsumerContracts()

func loadConsumerContracts() map[string]consumerContract {
	var catalog struct {
		Schema     string `json:"schema_id"`
		Operations []struct {
			Operation string   `json:"operation"`
			Request   string   `json:"request_schema"`
			Result    string   `json:"result_schema"`
			Order     []string `json:"field_order"`
		} `json:"operations"`
	}
	decoder := json.NewDecoder(bytes.NewReader(projection("consumer_operations.v1.json")))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil || catalog.Schema != "cartulary.reference_pack_consumer_operations.v1" || len(catalog.Operations) != 5 {
		panic("invalid consumer operation projection")
	}
	contracts := map[string]consumerContract{}
	for _, op := range catalog.Operations {
		if !projectionReferencePattern.MatchString(op.Request) || !projectionReferencePattern.MatchString(op.Result) {
			panic("invalid consumer schema reference")
		}
		request, result := compileProjection(op.Request), compileProjection(op.Result)
		if request.kind != "object" || !request.closed || len(request.properties) != len(op.Order) {
			panic("consumer request must be closed and ordered")
		}
		seen := map[string]bool{}
		for _, field := range op.Order {
			if request.properties[field] == nil || seen[field] {
				panic("invalid consumer field order")
			}
			seen[field] = true
		}
		if _, exists := contracts[op.Operation]; exists {
			panic("duplicate consumer operation")
		}
		contracts[op.Operation] = consumerContract{request, result, op.Order}
	}
	for _, operation := range []string{"ResolveCurrentPackSet", "GetPackEntry", "LookupPackEntries", "EvaluateIndicatorValue", "GetPackProvenance"} {
		if _, ok := contracts[operation]; !ok {
			panic("missing consumer operation")
		}
	}
	return contracts
}

// ValidateConsumerRequest is shared by native callers and strict JSON adapters.
// Shape and scalar checks precede any retained-set or profile resolution. Field
// order comes from the owner projection; hostile names/values never enter errors.
func ValidateConsumerRequest(operation string, value any) error {
	contract, known := consumerContracts[operation]
	object, ok := value.(map[string]any)
	if !known || !ok {
		return fail("invalid_pack_request")
	}
	for key, item := range object {
		if contract.request.properties[key] == nil || item == nil {
			return fail("invalid_pack_request")
		}
	}
	for _, key := range contract.request.required {
		if _, exists := object[key]; !exists {
			return fail("invalid_pack_request")
		}
	}
	for _, key := range contract.order {
		item, present := object[key]
		if !present {
			continue
		}
		if !contract.request.properties[key].matches(item) {
			if operation == "EvaluateIndicatorValue" && key == "raw_value" {
				if raw, ok := item.(string); ok && utf8.ValidString(raw) && utf8.RuneCountInString(raw) > 8192 {
					return fail("indicator_input_too_long")
				}
			}
			return fail("invalid_pack_request")
		}
	}
	return nil
}

func AdmitConsumerRequest(operation string, data []byte) error {
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return fail("invalid_pack_request")
	}
	return ValidateConsumerRequest(operation, value)
}

// ValidateConsumerResult checks the exact serialized projection. Indexed
// content and provenance are already validated on publication; this validator
// is also used by owner integration evidence against the actual consumer output.
func ValidateConsumerResult(operation string, data []byte) error {
	contract, ok := consumerContracts[operation]
	value, err := canonicaljson.DecodeStrict(data)
	if !ok || err != nil || !contract.result.matches(value) {
		return errors.New("reference pack: invalid consumer result")
	}
	// A successful entry/page must carry its own selected profile, not merely
	// an arbitrary valid object from another member of the closed item union.
	envelope := value.(map[string]any)
	if envelope["error"] != nil {
		return nil
	}
	switch operation {
	case "GetPackEntry", "LookupPackEntries":
		success := envelope["value"].(map[string]any)
		provenance := success["provenance"].(map[string]any)
		profile, found := profiles[provenance["pack_key"].(string)]
		if !found {
			return errors.New("reference pack: invalid consumer result")
		}
		kind := "entry"
		if profile.Shape == "objects_relationships" {
			kind = "object"
		}
		check := func(item any) bool { return contentShapes[profile.Key+"/"+kind].matches(item) }
		if operation == "GetPackEntry" {
			if !check(success["item"]) {
				return errors.New("reference pack: invalid consumer result")
			}
		} else {
			for _, item := range success["items"].([]any) {
				if !check(item) {
					return errors.New("reference pack: invalid consumer result")
				}
			}
		}
	case "EvaluateIndicatorValue":
		success := envelope["value"].(map[string]any)
		if success["provenance"].(map[string]any)["pack_key"] != "type_registry.indicator" {
			return errors.New("reference pack: invalid consumer result")
		}
	}
	return nil
}
