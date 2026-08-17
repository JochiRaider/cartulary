package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

const maxPerformanceFixtureTimelineRows = 100_000

type PerformanceFixtureRow struct {
	Summary     string
	HostRef     string
	IdentityRef string
	Tag         string
	DataSource  string
}

type PerformanceFixtureCommand struct {
	Actor      authn.UserRecord
	IncidentID uuid.UUID
	Rows       []PerformanceFixtureRow
	Now        time.Time
}

type PerformanceFixtureResult struct {
	RowCount         int
	RelationshipRows int
}

func (f *Facade) CreatePerformanceFixtureRows(ctx context.Context, command PerformanceFixtureCommand) (PerformanceFixtureResult, error) {
	if f == nil || f.store == nil {
		return PerformanceFixtureResult{}, errors.New("timeline performance fixture facade is required")
	}
	if command.Actor.ID == uuid.Nil || command.IncidentID == uuid.Nil || command.Now.IsZero() ||
		len(command.Rows) == 0 || len(command.Rows) > maxPerformanceFixtureTimelineRows {
		return PerformanceFixtureResult{}, errors.New("timeline performance fixture command is invalid")
	}
	seenSummaries := make(map[string]struct{}, len(command.Rows))
	for index, row := range command.Rows {
		if !validPerformanceFixtureScalar("timeline.activity_synopsis_text", row.Summary) ||
			(row.DataSource != "" && !validPerformanceFixtureScalar("timeline.data_source_text", row.DataSource)) {
			return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture row %d has invalid scalar data", index+1)
		}
		if _, duplicate := seenSummaries[row.Summary]; duplicate {
			return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture row %d duplicates its summary", index+1)
		}
		seenSummaries[row.Summary] = struct{}{}
		relationshipFields := 0
		for _, value := range []string{row.HostRef, row.IdentityRef, row.Tag, row.DataSource} {
			if strings.TrimSpace(value) != "" {
				relationshipFields++
			}
		}
		if relationshipFields != 0 && relationshipFields != 4 {
			return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture row %d has a partial relationship set", index+1)
		}
		for fieldKey, value := range map[string]string{
			"timeline.host_refs": row.HostRef, "timeline.identity_refs": row.IdentityRef, "timeline.tags": row.Tag,
		} {
			if value == "" {
				continue
			}
			if _, ok := clipboardValueToPatchChange(fieldKey, value); !ok {
				return PerformanceFixtureResult{}, fmt.Errorf("timeline performance fixture row %d has invalid %s", index+1, fieldKey)
			}
		}
	}
	return f.store.createPerformanceFixtureRows(ctx, command)
}

func (f *Facade) ValidatePerformanceFixtureRows(ctx context.Context, incidentID uuid.UUID, expected PerformanceFixtureResult) error {
	if f == nil || f.store == nil || incidentID == uuid.Nil || expected.RowCount < 1 || expected.RelationshipRows < 0 {
		return errors.New("timeline performance fixture validation request is invalid")
	}
	actual, err := f.store.performanceFixtureCounts(ctx, incidentID)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("timeline performance fixture validation got rows=%d relationships=%d, want rows=%d relationships=%d", actual.RowCount, actual.RelationshipRows, expected.RowCount, expected.RelationshipRows)
	}
	return nil
}

func validPerformanceFixtureScalar(fieldKey string, value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return false
	}
	_, ok := normalizeFieldTextValue(fieldKey, raw)
	return ok
}
