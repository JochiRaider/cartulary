package deleterestore

import recordsdeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"

func NewProvider() recordsdeleterestore.TableProvider {
	return recordsdeleterestore.TableProvider{
		SourceTable:        "assessments",
		SourceRecordCol:    "record_id",
		StaticViewSchemaID: "cartulary.view.assessments.v1",
		SourceTombstone:    true,
	}
}
