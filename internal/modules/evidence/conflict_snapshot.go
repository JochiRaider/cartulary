package evidence

import conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"

func newEvidenceConflictSnapshotProjector() conflicttokens.RevisionSnapshotProjector {
	projector, err := conflicttokens.NewRevisionSnapshotProjector(
		"cartulary.revisions.snapshot.evidence.v1",
		map[string]string{
			"evidence.title":                "title",
			"evidence.lifecycle_state":      "lifecycle_state",
			"evidence.requested_at":         "requested_at",
			"evidence.received_at":          "received_at",
			"evidence.storage_ref":          "storage_ref",
			"evidence.blob_hash":            "blob_hash",
			"evidence.collector_party_text": "collector_party_text",
			"evidence.collector_party_id":   "collector_party_id",
			"evidence.source_party_text":    "source_party_text",
			"evidence.source_party_id":      "source_party_id",
			"evidence.upload_state":         "upload_state",
		},
	)
	if err != nil {
		panic(err)
	}
	return projector
}
