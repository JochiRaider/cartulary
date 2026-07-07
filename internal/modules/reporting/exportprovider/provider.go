package exportprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("reporting export provider: not found")

type IncidentSnapshot struct {
	ID           string
	Title        string
	Description  *string
	Status       string
	Severity     *string
	TLP          *string
	CurrentPhase *string
	Version      int64
}

type SourceBoundaryState struct {
	IncidentID               string  `json:"incident_id"`
	IncidentVersion          int64   `json:"incident_version"`
	LatestChangeSetID        *string `json:"latest_change_set_id"`
	LatestChangeSetCreatedAt *string `json:"latest_change_set_created_at"`
}

type Field struct {
	Path                    string
	ContentClass            string
	SourceFamily            string
	Value                   any
	DisclosurePartitionRefs []string
	SupportRefs             []string
	RawBlobSource           bool
	OpaqueBinary            bool
	GeneratedPresentation   bool
}

type FieldQuery struct {
	Prefix                       string
	SQL                          string
	DisclosurePartitionRefPrefix string
}

func CollectQueryFieldsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, supportRefs map[string][]string, queries []FieldQuery) ([]Field, error) {
	fields := []Field{}
	for _, query := range queries {
		rows, err := tx.Query(ctx, query.SQL, incidentID)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id string
			var sourceFamily string
			var contentClass string
			var raw []byte
			if err := rows.Scan(&id, &sourceFamily, &contentClass, &raw); err != nil {
				rows.Close()
				return nil, err
			}
			var value any
			if err := json.Unmarshal(raw, &value); err != nil {
				rows.Close()
				return nil, err
			}
			field := Field{
				Path:         fmt.Sprintf("/%s/%s", query.Prefix, id),
				ContentClass: contentClass,
				SourceFamily: sourceFamily,
				Value:        value,
				SupportRefs:  CloneStrings(supportRefs[id]),
			}
			if query.DisclosurePartitionRefPrefix != "" {
				field.DisclosurePartitionRefs = []string{query.DisclosurePartitionRefPrefix + id}
			}
			fields = append(fields, field)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Path < fields[j].Path
	})
	return fields, nil
}

func CloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return append([]string(nil), values...)
}
