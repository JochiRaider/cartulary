package incidentbundle

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
)

const (
	taskRequestsBundlePath = "data/task_requests.ndjson"
	decisionsBundlePath    = "data/decisions.ndjson"
)

func exportIncidentBundleFiles(ctx context.Context, input sourceport.ExportContext) error {
	for _, spec := range []sourceport.AttributedRows{
		{Path: taskRequestsBundlePath, Table: "task_requests", Identity: "record_id", ActorColumns: []string{"owner_user_id"}, Query: sqlshape.Literal(`SELECT to_jsonb(t) FROM task_requests t WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128`)},
		{Path: decisionsBundlePath, Table: "decisions", Identity: "record_id", ActorColumns: []string{"owner_user_id"}, Query: sqlshape.Literal(`SELECT to_jsonb(t) FROM decisions t WHERE incident_id=$1 AND ($2::uuid IS NULL OR record_id>$2) ORDER BY record_id LIMIT 128`)},
	} {
		if err := sourceport.ExportAttributedRows(ctx, input, spec); err != nil {
			return err
		}
	}
	return nil
}
