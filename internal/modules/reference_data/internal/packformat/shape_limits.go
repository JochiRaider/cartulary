package packformat

import "encoding/json"

var fixedShapeLimits = loadFixedShapeLimits()

func loadFixedShapeLimits() map[string]int {
	var catalog struct {
		SchemaID string         `json:"schema_id"`
		Fixed    map[string]int `json:"fixed"`
	}
	if err := json.Unmarshal(projection("limits.v3.json"), &catalog); err != nil || catalog.SchemaID != "cartulary.reference_pack_limits.v3" {
		panic("invalid reference pack limits catalog")
	}
	return catalog.Fixed
}
