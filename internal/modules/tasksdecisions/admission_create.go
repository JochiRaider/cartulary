package tasksdecisions

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func AdmitCreateJSON(viewSchemaID string, reader io.Reader) (CreateRequest, *AdmissionFailure) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok || !schema.CreateCapable || !isMutationView(viewSchemaID) {
		return CreateRequest{}, invalidAdmission("view_schema_id", "unknown_view_schema")
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return CreateRequest{}, invalidAdmission("", "request_not_object")
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return CreateRequest{}, invalidAdmission(key, "unknown_field")
		}
	}
	request := CreateRequest{
		ViewSchemaID: viewSchemaID,
		Values:       map[string]FieldValue{},
		Collections:  map[string]CollectionActionPayload{},
	}
	if value, present := raw["client_txn_id"]; !present {
		return CreateRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateRequest{}, invalidAdmission("client_txn_id", "missing_required_field")
	}
	for fieldKey, field := range schema.Fields() {
		value, present := raw[fieldKey]
		if !present {
			continue
		}
		if field.ConflictResolutionClass == "collection_review" {
			payload, apiErr := decodeMutationCollectionPayload(fieldKey, value)
			if apiErr != nil {
				return CreateRequest{}, apiErr
			}
			request.Collections[fieldKey] = payload
			continue
		}
		admitted, _, apiErr := decodeMutationValue(fieldKey, field, value, false)
		if apiErr != nil {
			return CreateRequest{}, apiErr
		}
		request.Values[fieldKey] = admitted
	}
	return request, nil
}
