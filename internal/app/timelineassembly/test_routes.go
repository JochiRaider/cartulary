package timelineassembly

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func registerTestControlRoutes(mux *http.ServeMux, deps httpapi.DependencySet) error {
	if !httpapi.TestRoutesEnabled(deps.Env) {
		return nil
	}
	guard, err := httpapi.NewTestRouteGuard(deps.Env)
	if err != nil {
		return err
	}
	mux.HandleFunc("/api/v1/test/timeline/record-changes", guard.Protect(recordChangeSnapshotHandler(deps.PostgresHandle())))
	mux.HandleFunc("GET /api/v1/test/timeline/records/{record_id}/substrate", guard.Protect(recordSubstrateHandler(deps.PostgresHandle())))
	return nil
}

func recordChangeSnapshotHandler(db postgres.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(), `
SELECT canonical_payload
  FROM collaboration_replay_events
 WHERE event_family = 'record_changed'
 ORDER BY stream_seq, event_id
`)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		items := make([]map[string]any, 0)
		for rows.Next() {
			var raw []byte
			if err := rows.Scan(&raw); err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			var payload map[string]any
			if err := json.Unmarshal(raw, &payload); err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			items = append(items, payload)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{"record_changes": items})
	}
}

func recordSubstrateHandler(db postgres.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		recordID, err := uuid.Parse(r.PathValue("record_id"))
		if err != nil {
			http.Error(w, "record not found", http.StatusNotFound)
			return
		}
		var (
			rowVersion          int64
			captureState        string
			replacementRecordID pgtype.UUID
			revisionCount       int
		)
		err = db.QueryRow(r.Context(), `
SELECT
    r.row_version,
    e.capture_state,
    (
        SELECT l.src_record_id
          FROM active_record_links_v1 l
         WHERE l.dst_record_id = e.record_id
           AND l.link_type = 'supersedes'
         ORDER BY l.created_at DESC, l.record_link_id DESC
         LIMIT 1
    ),
    (SELECT COUNT(*) FROM record_revisions rr WHERE rr.record_id = e.record_id)
  FROM timeline_events e
  JOIN records r
    ON r.record_id = e.record_id
   AND r.incident_id = e.incident_id
 WHERE e.record_id = $1
   AND r.record_type = 'timeline_event'
   AND r.deleted_at IS NULL
`, recordID).Scan(&rowVersion, &captureState, &replacementRecordID, &revisionCount)
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "record not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		var replacement any
		if replacementRecordID.Valid {
			replacement = uuid.UUID(replacementRecordID.Bytes).String()
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
			"record_id":             recordID.String(),
			"row_version":           rowVersion,
			"capture_state":         captureState,
			"replacement_record_id": replacement,
			"record_revision_count": revisionCount,
		})
	}
}
