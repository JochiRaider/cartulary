package deleterestore

import recordsdeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"

func NewProvider() recordsdeleterestore.TableProvider {
	return recordsdeleterestore.TableProvider{
		SourceTable:        "timeline_events",
		SourceRecordCol:    "record_id",
		StaticViewSchemaID: "cartulary.view.timeline.v2",
	}
}
