package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func TestCanonicalCoordinatorExecutionScope_Integration(t *testing.T) {
	f := newCanonicalCoordinatorFixture(t, "reference-pack-execution-scope")
	accepted, err := f.coordinator.Import(t.Context(), f.actor, "selected-input", bytes.NewReader(f.container))
	if err != nil {
		t.Fatal(err)
	}
	jobID := uuid.MustParse(accepted.Job.JobID)
	execution, claimed, err := f.manager.Claim(t.Context(), jobID)
	if err != nil || !claimed {
		t.Fatal("claim", claimed, err)
	}

	t.Run("unrelated portability history is outside execution", func(t *testing.T) {
		maintenance, err := f.pool.Begin(t.Context())
		if err != nil {
			t.Fatal(err)
		}
		defer maintenance.Rollback(context.Background())
		if _, err := maintenance.Exec(t.Context(), `LOCK TABLE reference_pack_portable_catalogs, reference_pack_portable_preparations IN ACCESS EXCLUSIVE MODE`); err != nil {
			t.Fatal(err)
		}
		// This is a deadlock bound, not a performance threshold. A global
		// compatibility rescan cannot finish until maintenance releases its lock.
		ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
		defer cancel()
		if err := f.coordinator.Execute(ctx, execution); err != nil {
			t.Fatal("selected import depended on unrelated retained catalogs", err)
		}
		job, err := f.manager.Get(t.Context(), jobID)
		if err != nil || job.Status != jobs.StatusSucceeded {
			t.Fatal("import did not complete through the production worker", job.Status, err)
		}
	})

	var retained []byte
	if err := f.pool.QueryRow(t.Context(), `SELECT frozen_input FROM reference_pack_operations WHERE job_id=$1`, jobID).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"schema_id", "pack_contract_version"} {
		t.Run("selected input rejects incompatible "+field, func(t *testing.T) {
			var input map[string]any
			if err := json.Unmarshal(retained, &input); err != nil {
				t.Fatal(err)
			}
			input[field] = "unsupported"
			raw, err := canonicaljson.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			tx, err := f.pool.Begin(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(context.Background())
			operation := uuid.New()
			if _, err := tx.Exec(t.Context(), `INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,frozen_input) VALUES($1,'import','system',now(),$2)`, operation, raw); err != nil {
				t.Fatal(err)
			}
			if _, err := beginAttemptTx(t.Context(), tx, operation, time.Now().UTC()); err == nil {
				t.Fatal("worker accepted incompatible selected input")
			}
			var count int
			if err := tx.QueryRow(t.Context(), `SELECT count(*) FROM reference_pack_attempts WHERE operation_id=$1`, operation).Scan(&count); err != nil || count != 0 {
				t.Fatal("incompatible input wrote an execution attempt", count, err)
			}
		})
	}
}
