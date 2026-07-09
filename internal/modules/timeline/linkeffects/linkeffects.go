package linkeffects

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
)

func LoadTimelineInvalidationsTx(ctx context.Context, tx pgx.Tx, linkTypesByRecord map[uuid.UUID][]string) ([]mentioneffects.TimelineInvalidation, error) {
	fieldKeysByRecord := make(map[uuid.UUID][]string, len(linkTypesByRecord))
	for recordID, linkTypes := range linkTypesByRecord {
		for _, linkType := range linkTypes {
			fieldKey := timelineFieldKeyForLinkType(linkType)
			if fieldKey == "" {
				continue
			}
			fieldKeysByRecord[recordID] = append(fieldKeysByRecord[recordID], fieldKey)
		}
	}
	return mentioneffects.LoadTimelineInvalidationsTx(ctx, tx, fieldKeysByRecord)
}

func timelineFieldKeyForLinkType(linkType string) string {
	switch linkType {
	case "observed_on_host":
		return "timeline.host_refs"
	case "observed_as_identity":
		return "timeline.identity_refs"
	default:
		return ""
	}
}
