package deleterestore

import recordsdeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"

const timelineViewSchemaID = "cartulary.view.timeline.v2"

type Provider struct {
	recordsdeleterestore.TableProvider
}

func NewProvider() Provider {
	return Provider{
		TableProvider: recordsdeleterestore.TableProvider{
			SourceTable:        "timeline_events",
			SourceRecordCol:    "record_id",
			StaticViewSchemaID: timelineViewSchemaID,
		},
	}
}
