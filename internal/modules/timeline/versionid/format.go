package versionid

import (
	"strconv"

	"github.com/google/uuid"
)

const prefix = "timeline_record:"

// Format returns the sole current Timeline version identifier for a
// domain-valid record UUID and row version.
func Format(recordID uuid.UUID, rowVersion int64) string {
	return prefix + recordID.String() + ":" + strconv.FormatInt(rowVersion, 10)
}
