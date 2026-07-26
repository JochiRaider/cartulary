package timeline

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func projectRecord(record sourcerepository.Snapshot, replacementRecordID *uuid.UUID) workbookprojection.DerivedRecord {
	return workbookprojection.Derive(record, replacementRecordID)
}
