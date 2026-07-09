package deleterestore

import recordsdeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"

func NewProvider() recordsdeleterestore.TableProvider {
	return recordsdeleterestore.TableProvider{
		SourceTable:        "evidence",
		SourceRecordCol:    "record_id",
		StaticViewSchemaID: "cartulary.view.evidence.v1",
	}
}
