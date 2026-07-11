package deleterestore

import recordsdeleterestore "github.com/JochiRaider/cartulary/internal/modules/records/deleterestore"

func HostProvider() recordsdeleterestore.TableProvider {
	return recordsdeleterestore.TableProvider{
		SourceTable:        "hosts",
		SourceRecordCol:    "record_id",
		StaticViewSchemaID: "cartulary.view.hosts.v1",
	}
}

func IdentityProvider() recordsdeleterestore.TableProvider {
	return recordsdeleterestore.TableProvider{
		SourceTable:        "identities",
		SourceRecordCol:    "record_id",
		StaticViewSchemaID: "cartulary.view.identities.v1",
	}
}
