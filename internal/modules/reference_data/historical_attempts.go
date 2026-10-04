package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type historicalAttemptQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// Lifecycle evidence is authoritative even when no content was published. A
// restored abort must not become a content verdict or a successful attempt.
func validateHistoricalAttempts(ctx context.Context, db historicalAttemptQuery) error {
	after := uuid.Nil
	for {
		var id, operation uuid.UUID
		var kind string
		var outcome *string
		var completed, terminal bool
		var data []byte
		var selected, members, failures int64
		err := db.QueryRow(ctx, `SELECT a.attempt_id,a.operation_id,o.kind,a.completed_at IS NOT NULL,a.outcome,a.canonical_result,o.terminal_at IS NOT NULL,
(SELECT count(*) FROM reference_pack_operation_members m WHERE m.operation_id=a.operation_id),
(SELECT count(*) FROM reference_pack_attempt_members m WHERE m.attempt_id=a.attempt_id),
(SELECT count(*) FROM reference_pack_attempt_members m WHERE m.attempt_id=a.attempt_id AND m.verdict='content_rejected')
FROM reference_pack_attempts a JOIN reference_pack_operations o USING(operation_id) WHERE a.attempt_id>$1 ORDER BY a.attempt_id LIMIT 1`, after).Scan(&id, &operation, &kind, &completed, &outcome, &data, &terminal, &selected, &members, &failures)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		if kind == "import" && selected == 0 {
			selected = 1
		}
		if members > selected || completed != (outcome != nil) || completed != (data != nil) || terminal && !completed {
			return errHistoricalIntegrity
		}
		if completed {
			if err := packformat.ValidateAttemptResult(data); err != nil {
				return errHistoricalIntegrity
			}
			var result struct {
				Outcome string `json:"outcome"`
				Members int64  `json:"member_count"`
				Failed  int64  `json:"failed_count"`
			}
			if err := json.Unmarshal(data, &result); err != nil || result.Outcome != *outcome {
				return errHistoricalIntegrity
			}
			if result.Outcome != "interrupted" && result.Members != selected {
				return errHistoricalIntegrity
			}
			if result.Outcome == "succeeded" || result.Outcome == "content_rejected" {
				if members != selected || failures != result.Failed {
					return errHistoricalIntegrity
				}
			}
		}
		after = id
	}
	after = uuid.Nil
	for {
		var id uuid.UUID
		var kind string
		var input, result []byte
		var terminal bool
		var selected int64
		err := db.QueryRow(ctx, `SELECT operation_id,kind,frozen_input,final_outcome,terminal_at IS NOT NULL,
(SELECT count(*) FROM reference_pack_operation_members m WHERE m.operation_id=o.operation_id)
FROM reference_pack_operations o WHERE operation_id>$1 AND kind IN ('import','reverify','refresh') ORDER BY operation_id LIMIT 1`, after).Scan(&id, &kind, &input, &result, &terminal, &selected)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		frozen, err := decodeFrozenOperation(input)
		if err != nil || frozen.Kind != kind || terminal != (result != nil) {
			return errHistoricalIntegrity
		}
		if kind == "import" && selected == 0 {
			selected = 1
		}
		if terminal {
			if err := packformat.ValidateAttemptResult(result); err != nil {
				return errHistoricalIntegrity
			}
			var value struct {
				Outcome string `json:"outcome"`
				Members int64  `json:"member_count"`
			}
			if err := json.Unmarshal(result, &value); err != nil || value.Outcome == "interrupted" || value.Members != selected {
				return errHistoricalIntegrity
			}
			if value.Outcome == "succeeded" || value.Outcome == "content_rejected" {
				var completedResult []byte
				err := db.QueryRow(ctx, `SELECT canonical_result FROM reference_pack_attempts WHERE operation_id=$1 AND outcome=$2 ORDER BY started_at DESC,attempt_id DESC LIMIT 1`, id, value.Outcome).Scan(&completedResult)
				if err != nil || !bytes.Equal(result, completedResult) {
					return errHistoricalIntegrity
				}
			}
		}
		after = id
	}
}
