package evidence

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// AdmitCreateJSON admits the fixed Evidence create surface, including the
// optional initial-blob input. Evidence policy remains owner-side.
func AdmitCreateJSON(reader io.Reader) (CreateAdmission, *AdmissionFailure) {
	schema, ok := viewschema.Lookup(ViewSchemaID)
	if !ok || !schema.CreateCapable {
		return CreateAdmission{}, newAdmissionFailure("view_schema_id", admissionUnknownViewSchema)
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return CreateAdmission{}, newAdmissionFailure("", admissionRequestNotObject)
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for _, input := range schema.CreateInputs {
		allowed[input.InputKey] = struct{}{}
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return CreateAdmission{}, newAdmissionFailure(key, admissionUnknownField)
		}
	}

	request := createRequest{ViewSchemaID: ViewSchemaID, Values: map[string]FieldValue{}}
	if value, present := raw["client_txn_id"]; !present {
		return CreateAdmission{}, newAdmissionFailure("client_txn_id", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateAdmission{}, newAdmissionFailure("client_txn_id", admissionMissingRequiredField)
	}
	for fieldKey, field := range schema.Fields() {
		value, present := raw[fieldKey]
		if !present {
			continue
		}
		admitted, _, failure := decodeEvidenceValue(fieldKey, field, value, false)
		if failure != nil {
			return CreateAdmission{}, failure
		}
		request.Values[fieldKey] = admitted
	}
	if rawBlobID, present := raw["evidence.initial_object_blob_id"]; present {
		if string(rawBlobID) == "null" {
			return CreateAdmission{}, newAdmissionFailure("evidence.initial_object_blob_id", admissionFieldNotNullable)
		}
		var value string
		if json.Unmarshal(rawBlobID, &value) != nil {
			return CreateAdmission{}, newAdmissionFailure("evidence.initial_object_blob_id", admissionInvalidValue)
		}
		blobID, parseErr := uuid.Parse(value)
		if parseErr != nil || blobID.String() != value {
			return CreateAdmission{}, newAdmissionFailure("evidence.initial_object_blob_id", admissionInvalidValue)
		}
		request.InitialObjectBlobID = &blobID
	}
	hash := createAdmissionHash(request)
	return CreateAdmission{request: cloneCreateRequest(request), hash: hash, admitted: true}, nil
}

func createAdmissionHash(request createRequest) [sha256.Size]byte {
	values := make(map[string]any, len(request.Values)+1)
	for fieldKey, value := range request.Values {
		values[fieldKey] = canonicalEvidenceValue(value)
	}
	if lifecycle, present := request.Values["evidence.lifecycle_state"]; !present {
		values["evidence.lifecycle_state"] = "requested"
	} else if lifecycle.Text != nil {
		values["evidence.lifecycle_state"] = *lifecycle.Text
	}
	inputs := map[string]any{}
	if request.InitialObjectBlobID != nil {
		inputs["evidence.initial_object_blob_id"] = request.InitialObjectBlobID.String()
	}
	return sha256.Sum256(canonicalHashBytes(map[string]any{
		"view_schema_id": ViewSchemaID, "values": values,
		"collection_ops": map[string]any{}, "create_inputs": inputs,
	}))
}
