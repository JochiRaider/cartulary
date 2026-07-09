package deleterestore

import recordsdeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"

func TaskRequestProvider() recordsdeleterestore.TableProvider {
	return recordsdeleterestore.TableProvider{
		SourceTable:        "task_requests",
		SourceRecordCol:    "record_id",
		StaticViewSchemaID: "cartulary.view.task_requests.v1",
	}
}

func DecisionProvider() recordsdeleterestore.TableProvider {
	return recordsdeleterestore.TableProvider{
		SourceTable:        "decisions",
		SourceRecordCol:    "record_id",
		StaticViewSchemaID: "cartulary.view.decisions.v1",
	}
}
