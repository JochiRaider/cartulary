package timeline

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func projectRecord(record sourceRecord, replacementRecordID *uuid.UUID) projectedRecord {
	return workbookprojection.Derive(record, replacementRecordID)
}
