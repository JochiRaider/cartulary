package reference_data

import (
	"bytes"
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

func decodeSuccessfulEnvelope(data []byte) (successfulEnvelope, error) {
	if err := packformat.ValidateSuccessfulEnvelope(data); err != nil {
		return successfulEnvelope{}, err
	}
	var envelope successfulEnvelope
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		return successfulEnvelope{}, err
	}
	if envelope.ContainerRef != nil {
		if len(*envelope.ContainerRef) > 4096 {
			return successfulEnvelope{}, ErrInvalidReferencePackStorageReference
		}
		if _, err := ParseStorageRef(*envelope.ContainerRef); err != nil {
			return successfulEnvelope{}, err
		}
	}
	return envelope, nil
}
