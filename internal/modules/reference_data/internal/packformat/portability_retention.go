package packformat

import (
	"bytes"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

var portabilityInputShape = compileProjection("portability_input.v1.schema.json")
var portabilityResolutionShape = compileProjection("portability_resolution.v1.schema.json")
var errPortabilityRetention = errors.New("reference pack: invalid portability retention evidence")

type PortabilityInput struct {
	IncidentID        uuid.UUID
	SourceOperationID uuid.UUID
	Content           PortableContent
}

func decodePortabilityObject(data []byte, rule *shape) (map[string]any, error) {
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !rule.matches(value) {
		return nil, errPortabilityRetention
	}
	canonical, err := canonicaljson.Marshal(value)
	if err != nil || !bytes.Equal(canonical, data) {
		return nil, errPortabilityRetention
	}
	return value.(map[string]any), nil
}

// Admission and recovery validate frozen input without fabricating a terminal
// resolution for a private or aborted cohort.
func DecodePortabilityInput(input []byte, refs PortableReferences) (PortabilityInput, error) {
	invalid := func() (PortabilityInput, error) { return PortabilityInput{}, errPortabilityRetention }
	in, err := decodePortabilityObject(input, portabilityInputShape)
	if err != nil {
		return invalid()
	}
	incident, err := uuid.Parse(in["incident_id"].(string))
	if err != nil || incident == uuid.Nil || incident.String() != in["incident_id"] {
		return invalid()
	}
	source, err := uuid.Parse(in["source_operation_id"].(string))
	if err != nil || source == uuid.Nil || source.String() != in["source_operation_id"] {
		return invalid()
	}
	canonical, err := canonicaljson.Marshal(refs)
	if err != nil || in["catalog_sha256"] != Digest(canonical) {
		return invalid()
	}
	contentBytes, err := canonicaljson.Marshal(in["content_manifest"])
	if err != nil {
		return invalid()
	}
	content, err := DecodePortableContent(contentBytes, refs)
	if err != nil {
		return invalid()
	}
	return PortabilityInput{incident, source, content}, nil
}

func ValidatePortabilityRetention(input, result []byte, refs PortableReferences) error {
	in, err := DecodePortabilityInput(input, refs)
	if err != nil {
		return err
	}
	required := RequiredPortableVersions(in.Content, refs)
	out, err := decodePortabilityObject(result, portabilityResolutionShape)
	if err != nil {
		return err
	}
	versions := out["versions"].([]any)
	if len(versions) != len(refs.Versions) {
		return errPortabilityRetention
	}
	for index, item := range versions {
		v := item.(map[string]any)
		if v["pack_key"] != refs.Versions[index].Key || v["pack_version"] != refs.Versions[index].Version || v["available"].(bool) != (v["reason_code"] == nil) || required[refs.Versions[index].Key+"\x00"+refs.Versions[index].Version] && !v["available"].(bool) {
			return errPortabilityRetention
		}
	}
	return nil
}
