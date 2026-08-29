package merge

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/mutationadmission"
	"github.com/JochiRaider/cartulary/internal/platform/strictjson"
)

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *mutationadmission.Failure) {
	raw, err := strictjson.DecodeObject(reader)
	if err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func invalidMutationPayload(field string, reasonCode mutationadmission.ReasonCode) *mutationadmission.Failure {
	return mutationadmission.New(field, reasonCode)
}

func entityVersionID(prefix string, recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("%s:%s:%d", prefix, recordID.String(), rowVersion)
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func stringPointer(value string) *string {
	cloned := value
	return &cloned
}
