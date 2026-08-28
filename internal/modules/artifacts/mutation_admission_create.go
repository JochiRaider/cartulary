package artifacts

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// AdmitCreate admits one of the fixed Artifact create surfaces. All
// view-specific normalization and minimum-create rules remain owner-local.
func AdmitCreate(viewSchemaID string, reader io.Reader) (CreateAdmission, *AdmissionError) {
	schema, ok := viewschema.Lookup(viewSchemaID)
	if !ok || !schema.CreateCapable || !isArtifactBackedView(viewSchemaID) {
		return CreateAdmission{}, newAdmissionError("view_schema_id", admissionUnknownViewSchema)
	}
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return CreateAdmission{}, newAdmissionError("", admissionRequestNotObject)
	}
	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, admitted := allowed[key]; !admitted {
			return CreateAdmission{}, newAdmissionError(key, admissionUnknownField)
		}
	}
	request := createRequest{
		ViewSchemaID: viewSchemaID,
		Values:       map[string]fieldValue{},
		Collections:  map[string]collectionActionPayload{},
	}
	if value, present := raw["client_txn_id"]; !present {
		return CreateAdmission{}, newAdmissionError("client_txn_id", admissionMissingRequiredField)
	} else if json.Unmarshal(value, &request.ClientTxnID) != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateAdmission{}, newAdmissionError("client_txn_id", admissionMissingRequiredField)
	}
	for fieldKey, field := range schema.Fields() {
		value, present := raw[fieldKey]
		if !present {
			continue
		}
		if field.ConflictResolutionClass == "collection_review" {
			payload, admissionErr := decodeArtifactCollectionActionPayload(fieldKey, value)
			if admissionErr != nil {
				return CreateAdmission{}, admissionErr
			}
			request.Collections[fieldKey] = payload
			continue
		}
		admitted, _, admissionErr := decodeArtifactValue(fieldKey, field, value, false)
		if admissionErr != nil {
			return CreateAdmission{}, admissionErr
		}
		request.Values[fieldKey] = admitted
	}
	if err := validateCreateParams(createParams{ViewSchemaID: viewSchemaID, Values: request.Values}); err != nil {
		var validation *ValidationError
		if errors.As(err, &validation) {
			return CreateAdmission{}, newAdmissionError(validation.Field, knownAdmissionReason(validation.ReasonCode))
		}
		return CreateAdmission{}, newAdmissionError("payload", admissionInvalidValue)
	}
	hash := createAdmissionHash(request)
	return CreateAdmission{request: cloneCreateRequest(request), hash: hash, admitted: true}, nil
}

func AdmitContextualNote(reader io.Reader) (ContextualNoteAdmission, *AdmissionError) {
	admission, admissionErr := AdmitCreate(NotesViewSchemaID, reader)
	if admissionErr != nil {
		return ContextualNoteAdmission{}, admissionErr
	}
	return ContextualNoteAdmission{request: admission.requestValue(), admitted: true}, nil
}
