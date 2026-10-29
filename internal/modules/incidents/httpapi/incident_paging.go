package httpapi

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

func parseIncidentListScope(rawQuery string) (listquery.Result, *platformhttpapi.APIError) {
	result, queryErr := listquery.Parse(rawQuery, listquery.Config{
		Search: true,
		ExactFilters: map[string]listquery.ExactFilter{
			"status": {Allowed: []string{"active", "closed"}},
		},
	})
	if queryErr == nil {
		return result, nil
	}
	if queryErr.Kind == listquery.ErrorKindPagination {
		return listquery.Result{}, invalidPaginationRequest(queryErr.ReasonCode)
	}
	return listquery.Result{}, invalidListQuery(queryErr.ReasonCode)
}

func (s *service) incidentListPageRequest(binding pagination.Binding, cursor *pagination.Cursor) (incidents.IncidentListPageRequest, string) {
	request := incidents.IncidentListPageRequest{
		Limit:        binding.Limit + 1,
		SearchTokens: strings.Fields(binding.Scope["search"]),
	}
	if status := binding.Scope["status"]; status != "" {
		request.Status = &status
	}
	if cursor == nil {
		return request, ""
	}
	if cursor.Mode != pagination.ModeKeyset {
		return incidents.IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor, err := time.Parse(time.RFC3339Nano, cursor.Position["anchor_updated_at"])
	if err != nil {
		return incidents.IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastUpdatedAt, err := time.Parse(time.RFC3339Nano, cursor.Position["last_updated_at"])
	if err != nil {
		return incidents.IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastID, err := uuid.Parse(cursor.Position["last_incident_id"])
	if err != nil {
		return incidents.IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor = anchor.UTC()
	request.AnchorUpdatedAt = &anchor
	request.After = &incidents.IncidentListPosition{UpdatedAt: lastUpdatedAt.UTC(), ID: lastID}
	return request, ""
}

func buildIncidentListPage(binding pagination.Binding, anchor time.Time, records []incidents.IncidentRecord) ([]json.RawMessage, *pagination.Cursor, error) {
	hasMore := len(records) > binding.Limit
	pageRecords := records
	if hasMore {
		pageRecords = records[:binding.Limit]
	}
	resources := make([]map[string]any, 0, len(pageRecords))
	for _, record := range pageRecords {
		resources = append(resources, incidents.BuildIncidentResource(record))
	}
	rows, err := pagination.MarshalResources(resources)
	if err != nil {
		return nil, nil, err
	}
	if !hasMore || len(pageRecords) == 0 {
		return rows, nil, nil
	}
	last := pageRecords[len(pageRecords)-1]
	return rows, &pagination.Cursor{
		Version:     pagination.CursorVersion,
		Mode:        pagination.ModeKeyset,
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       binding.Scope,
		Position: map[string]string{
			"anchor_updated_at": anchor.UTC().Format(time.RFC3339Nano),
			"last_updated_at":   last.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"last_incident_id":  last.ID.String(),
		},
	}, nil
}
