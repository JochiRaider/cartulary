package mutationpolicy

import "slices"

const (
	MaxPatchChanges      = 32
	MaxCollectionActions = 64
	MaxVisibleTextRunes  = 32_768
)

var directWritableFieldKeys = []string{
	"timeline.activity_local_text",
	"timeline.activity_synopsis_text",
	"timeline.activity_utc_text",
	"timeline.analyst_text",
	"timeline.data_source_text",
	"timeline.date_entered_text",
	"timeline.device_object_text",
	"timeline.ip_address_text",
	"timeline.mitre_stage_text",
	"timeline.raw_activity_text",
}

func DirectWritableFieldKeys() []string {
	return slices.Clone(directWritableFieldKeys)
}

func IsDirectWritableField(fieldKey string) bool {
	_, found := slices.BinarySearch(directWritableFieldKeys, fieldKey)
	return found
}

func IsValidVisibleText(value string) bool {
	if len([]rune(value)) > MaxVisibleTextRunes {
		return false
	}
	for _, current := range value {
		if current == 0 ||
			((current < 0x20 || (current >= 0x7f && current <= 0x9f)) &&
				current != '\t' && current != '\n' && current != '\r') {
			return false
		}
	}
	return true
}
