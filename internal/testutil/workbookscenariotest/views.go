package workbookscenariotest

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func QueryViewEnvelope(t testing.TB, serverURL string, incidentID string, viewSchemaID string, login appsupport.LoginResult) map[string]any {
	t.Helper()

	resp := appsupport.DoJSON(
		t,
		http.MethodPost,
		serverURL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/query",
		map[string]any{},
		appsupport.WithCookies(login.SessionCookie),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func QueryViewRows(t testing.TB, serverURL string, incidentID string, viewSchemaID string, login appsupport.LoginResult) []map[string]any {
	t.Helper()

	data := QueryViewEnvelope(t, serverURL, incidentID, viewSchemaID, login)["data"].(map[string]any)
	rawRows := data["rows"].([]any)
	rows := make([]map[string]any, 0, len(rawRows))
	for _, rawRow := range rawRows {
		rows = append(rows, rawRow.(map[string]any))
	}
	return rows
}

func FindRow(t testing.TB, rows []map[string]any, recordID string) map[string]any {
	t.Helper()

	for _, row := range rows {
		if row["record_id"] == recordID {
			return row
		}
	}
	t.Fatalf("expected row record_id=%q in %#v", recordID, rows)
	return nil
}

func CollectionItems(t testing.TB, row map[string]any, fieldKey string) []map[string]any {
	t.Helper()

	cells := row["cells"].(map[string]any)
	cell := cells[fieldKey].(map[string]any)
	value := cell["value"].(map[string]any)
	switch rawItems := value["items"].(type) {
	case []any:
		items := make([]map[string]any, 0, len(rawItems))
		for _, rawItem := range rawItems {
			items = append(items, rawItem.(map[string]any))
		}
		return items
	case []map[string]any:
		items := make([]map[string]any, 0, len(rawItems))
		items = append(items, rawItems...)
		return items
	default:
		t.Fatalf("unexpected collection item payload for %s: %#v", fieldKey, value["items"])
	}
	return nil
}

func RequireSingleCollectionItem(t testing.TB, row map[string]any, fieldKey string) map[string]any {
	t.Helper()

	items := CollectionItems(t, row, fieldKey)
	if len(items) != 1 {
		t.Fatalf("expected exactly one %s item, got %#v", fieldKey, items)
	}
	return items[0]
}
