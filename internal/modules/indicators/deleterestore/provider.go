package deleterestore

import recordsdeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"

func NewProvider() recordsdeleterestore.TableProvider {
	return recordsdeleterestore.TableProvider{
		SourceTable:        "indicators",
		SourceRecordCol:    "record_id",
		StaticViewSchemaID: "cartulary.view.indicators.v1",
		SourceTombstone:    true,
	}
}
