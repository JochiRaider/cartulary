package reference_data

import (
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

func decodeRetainedProvenance(data []byte, set PackSet) ([]PackProvenance, error) {
	if err := packformat.ValidateProvenance(data, set); err != nil {
		return nil, consumerError("pack_unavailable")
	}
	var anchors []PackProvenance
	if err := json.Unmarshal(data, &anchors); err != nil {
		return nil, consumerError("pack_unavailable")
	}
	return anchors, nil
}
