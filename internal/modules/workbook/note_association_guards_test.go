package workbook_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func TestNoteAssociationHTTPGuardsAndEffects(t *testing.T) {
	h, admin, adminID, incidentID := ConflictFixture(t, "note-association-guards", "IR-NOTE-GUARDS")
	// Stop asynchronous delivery, not transactional publication: durable intents
	// remain observable, and background sequencing cannot race effect snapshots.
	if err := h.Collaboration.CloseDispatcher(context.Background()); err != nil {
		t.Fatal(err)
	}
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := h.DB.ExecContext(context.Background(), query, args...); err != nil {
			t.Fatal(err)
		}
	}
	note := CreateNote(t, h, admin, incidentID, "guards-note", "Current Note", "Body")
	noteID := appsupport.MustUUID(t, note["record_id"].(string))
	otherIDs := []uuid.UUID{}
	for i := 0; i < 4; i++ {
		other := CreateNote(t, h, admin, incidentID, fmt.Sprintf("guards-other-%d", i), "Other Note", "Body")
		otherIDs = append(otherIDs, appsupport.MustUUID(t, other["record_id"].(string)))
	}
	routeFor := func(id uuid.UUID) string {
		return h.Server.HTTP.URL + "/api/v1/records/" + id.String() + "/note-associations"
	}
	route := routeFor(noteID)
	cookie := func(login appsupport.LoginResult) []func(*http.Request) {
		return []func(*http.Request){appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value)}
	}
	version := func(id uuid.UUID) int64 {
		t.Helper()
		var v int64
		if err := h.DB.QueryRowContext(context.Background(), `SELECT row_version FROM records WHERE record_id=$1`, id).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	payload := func(txn string, base int64, ids ...uuid.UUID) map[string]any {
		actions := []map[string]any{}
		for _, id := range ids {
			actions = append(actions, map[string]any{"op": "add", "counterpart_record_id": id.String()})
		}
		return map[string]any{"kind": "related_note", "base_row_version": base, "client_txn_id": txn, "actions": actions}
	}
	send := func(method, path string, body any, opts ...func(*http.Request)) *http.Response {
		t.Helper()
		return appsupport.DoJSON(t, method, path, body, opts...)
	}
	success := func(method, path string, body any, opts ...func(*http.Request)) map[string]any {
		t.Helper()
		return httptestx.RequireSuccessEnvelope(t, send(method, path, body, opts...), 200)["data"].(map[string]any)
	}
	reject := func(name, method, path string, body any, status int, code string, details map[string]any, opts ...func(*http.Request)) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			before := snapshotNoteAssociationEffects(t, h, incidentID)
			envelope := httptestx.RequireErrorEnvelope(t, send(method, path, body, opts...), status, code)
			if !reflect.DeepEqual(httptestx.RequireErrorDetails(t, envelope), details) {
				t.Fatalf("unexpected safe details for %s", code)
			}
			requireNoteAssociationEffects(t, before, snapshotNoteAssociationEffects(t, h, incidentID))
		})
	}
	empty := map[string]any{}
	roleDenied := map[string]any{"required_role": "editor|reviewer|admin"}
	malformed := map[string]any{"kind": "unknown", "actions": "invalid"}
	for _, method := range []string{http.MethodGet, http.MethodPost} {
		reject("HC01/"+method, method, route+"?kind=bad&limit=0", malformed, 401, "session_required", empty)
		t.Run("HC17/"+method, func(t *testing.T) {
			before := snapshotNoteAssociationEffects(t, h, incidentID)
			resp := send(method, h.Server.HTTP.URL+"/api/v1/records/not-a-uuid/note-associations?kind=bad", malformed)
			defer resp.Body.Close()
			if resp.StatusCode != 404 {
				t.Fatalf("malformed path status=%d", resp.StatusCode)
			}
			requireNoteAssociationEffects(t, before, snapshotNoteAssociationEffects(t, h, incidentID))
		})
	}
	type actor struct {
		id    uuid.UUID
		login appsupport.LoginResult
	}
	seed := func(name, role string, deploymentAdmin bool) actor {
		t.Helper()
		password := "AssociationGuardPassword1!"
		user := authflowtest.SeedLocalUserRecord(t, h.DB, "note-guard-"+name+"@example.test", name, password, false, deploymentAdmin, true)
		if role != "" {
			incidentstoretest.SeedMembership(t, h.DB, incidentID, user.ID, user.DisplayName, role, adminID)
		}
		return actor{user.ID, LoginLocalUserNoMFA(t, h, user.Email, password)}
	}
	viewer := seed("viewer", "viewer", false)
	editor := seed("editor", "editor", false)
	reviewer := seed("reviewer", "reviewer", false)
	for _, state := range []string{"missing", "expired", "revoked"} {
		a := seed(state, "editor", false)
		switch state {
		case "missing":
			exec(`DELETE FROM user_sessions WHERE user_id=$1`, a.id)
		case "expired":
			exec(`UPDATE user_sessions SET session_expires_at=authenticated_at-interval '1 second',idle_expires_at=authenticated_at-interval '1 second',absolute_expires_at=authenticated_at-interval '1 second' WHERE user_id=$1`, a.id)
		case "revoked":
			exec(`UPDATE user_sessions SET revoked_at=authenticated_at,revoke_reason_code='logout' WHERE user_id=$1`, a.id)
		}
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			reject("HC02/"+state+"/"+method, method, route+"?kind=bad", malformed, 401, "session_required", empty, cookie(a.login)...)
		}
	}
	for _, csrf := range []struct {
		name string
		opts []func(*http.Request)
	}{
		{"missing", []func(*http.Request){appsupport.WithCookies(admin.SessionCookie, admin.CSRFCookie)}},
		{"wrong", []func(*http.Request){appsupport.WithCookies(admin.SessionCookie, admin.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, "wrong")}},
		{"unbound", []func(*http.Request){appsupport.WithCookies(admin.SessionCookie, viewer.login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, viewer.login.CSRFCookie.Value)}},
	} {
		reject("HC03/"+csrf.name, http.MethodPost, routeFor(uuid.New()), malformed, 403, "csrf_verification_failed", empty, csrf.opts...)
	}
	for _, a := range []actor{seed("nonmember", "", false), seed("deployment-admin", "", true)} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			reject("HC05/"+a.id.String()+"/"+method, method, route+"?kind=bad", malformed, 404, "incident_not_found", empty, cookie(a.login)...)
		}
	}
	reject("HC06/viewer", http.MethodPost, route, malformed, 403, "authorization_denied", roleDenied, cookie(viewer.login)...)
	t.Run("HC04/viewer-read-without-CSRF", func(t *testing.T) {
		success(http.MethodGet, route+"?kind=source", nil, appsupport.WithCookies(viewer.login.SessionCookie))
	})
	t.Run("HC04/query-read-without-CSRF", func(t *testing.T) {
		success(http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.notes.v1/query", map[string]any{}, appsupport.WithCookies(viewer.login.SessionCookie))
	})
	var accepted map[string]any
	var acceptedBody map[string]any
	for i, a := range []actor{editor, reviewer, {adminID, admin}} {
		t.Run("HC07/"+[]string{"editor", "reviewer", "admin"}[i], func(t *testing.T) {
			success(http.MethodGet, route+"?kind=related_note", nil, cookie(a.login)...)
			body := payload(fmt.Sprintf("guards-allowed-%d", i), version(noteID), otherIDs[i])
			opts := cookie(a.login)
			if i == 0 {
				opts = []func(*http.Request){appsupport.WithHeader("Authorization", "Bearer "+a.login.SessionCookie.Value)}
			}
			result := success(http.MethodPost, route, body, opts...)
			if result["change_set_id"] == nil || result["row"].(map[string]any)["row_version"] != float64(body["base_row_version"].(int64)+1) {
				t.Fatal("incomplete mutation receipt")
			}
			if i == 2 {
				accepted, acceptedBody = result, body
			}
		})
	}
	// All three public page members must agree on continuation state.
	pageEnvelope := httptestx.RequireSuccessEnvelope(t, send(http.MethodGet, route+"?kind=related_note&limit=1", nil, cookie(admin)...), 200)
	page := pageEnvelope["data"].(map[string]any)
	cursor := page["next_cursor_token"].(string)
	requireNotePagingAgreement(t, pageEnvelope, 1, true)
	continuation := httptestx.RequireSuccessEnvelope(t, send(http.MethodGet, route+"?kind=related_note&cursor_token="+url.QueryEscape(cursor), nil, cookie(admin)...), 200)
	requireNotePagingAgreement(t, continuation, 1, true)
	terminalCursor := continuation["data"].(map[string]any)["next_cursor_token"].(string)
	terminal := httptestx.RequireSuccessEnvelope(t, send(http.MethodGet, route+"?kind=related_note&cursor_token="+url.QueryEscape(terminalCursor), nil, cookie(admin)...), 200)
	requireNotePagingAgreement(t, terminal, 1, false)
	reject("HC10/other-note", http.MethodGet, routeFor(otherIDs[0])+"?kind=related_note&cursor_token="+url.QueryEscape(cursor), nil, 400, "invalid_pagination_request", map[string]any{"reason_code": "cursor_query_mismatch"}, cookie(admin)...)
	reject("HC11/other-actor", http.MethodGet, route+"?kind=related_note&cursor_token="+url.QueryEscape(cursor), nil, 400, "invalid_pagination_request", map[string]any{"reason_code": "invalid_cursor_token"}, cookie(editor.login)...)
	reject("HC12/tamper", http.MethodGet, route+"?kind=related_note&cursor_token=x"+url.QueryEscape(cursor), nil, 400, "invalid_pagination_request", map[string]any{"reason_code": "invalid_cursor_token"}, cookie(admin)...)
	viewerPage := success(http.MethodGet, route+"?kind=related_note&limit=1", nil, cookie(viewer.login)...)
	exec(`DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incidentID, viewer.id)
	reject("HC11/revoked-between-pages", http.MethodGet, route+"?kind=related_note&cursor_token="+url.QueryEscape(viewerPage["next_cursor_token"].(string)), nil, 404, "incident_not_found", empty, cookie(viewer.login)...)
	incidentstoretest.SeedMembership(t, h.DB, incidentID, viewer.id, "viewer", "viewer", adminID)
	// An incoming related Note is readable, but cannot be removed by this Note.
	success(http.MethodPost, routeFor(otherIDs[3]), payload("guards-incoming", version(otherIDs[3]), noteID), cookie(admin)...)
	items := success(http.MethodGet, route+"?kind=related_note", nil, cookie(admin)...)["items"].([]any)
	var incoming any
	for _, raw := range items {
		item := raw.(map[string]any)
		if item["direction"] == "incoming" {
			incoming = item["item_ref"]
		}
	}
	if incoming == nil {
		t.Fatal("missing incoming association fixture")
	}
	mixed := payload("guards-atomic", version(noteID), otherIDs[3])
	mixed["actions"] = append(mixed["actions"].([]map[string]any), map[string]any{"op": "remove", "item_ref": incoming})
	reject("HC15/atomic-ineligible-removal", http.MethodPost, route, mixed, 400, "invalid_mutation_payload", map[string]any{"field": "actions", "reason_code": "invalid_value"}, cookie(admin)...)
	currentVersion := version(noteID)
	reject("HC18/stale", http.MethodPost, route, payload("guards-stale", 1, otherIDs[3]), 409, "row_version_conflict", map[string]any{"record_id": noteID.String(), "base_row_version": float64(1), "current_row_version": float64(currentVersion)}, cookie(admin)...)
	divergent := payload(acceptedBody["client_txn_id"].(string), currentVersion, otherIDs[3])
	reject("HC18/divergent-key", http.MethodPost, route, divergent, 409, "client_txn_conflict", map[string]any{"client_txn_id": acceptedBody["client_txn_id"]}, cookie(admin)...)
	for _, kind := range []any{nil, "", "unknown"} {
		body := payload("guards-kind", currentVersion, otherIDs[3])
		if kind == nil {
			delete(body, "kind")
		} else {
			body["kind"] = kind
		}
		reject(fmt.Sprintf("HC18/kind-%v", kind), http.MethodPost, route, body, 400, "invalid_mutation_payload", map[string]any{"field": "kind", "reason_code": "invalid_value"}, cookie(admin)...)
	}
	noopBody := payload("guards-noop", currentVersion, otherIDs[0])
	beforeNoop := snapshotNoteAssociationEffects(t, h, incidentID)
	noop := success(http.MethodPost, route, noopBody, cookie(admin)...)
	if _, present := noop["change_set_id"]; present {
		t.Fatal("no-op included change_set_id")
	}
	if noop["row"].(map[string]any)["row_version"] != float64(currentVersion) {
		t.Fatal("no-op changed version")
	}
	afterNoop := snapshotNoteAssociationEffects(t, h, incidentID)
	if beforeNoop["route_idempotency"] == afterNoop["route_idempotency"] {
		t.Fatal("no-op did not retain receipt")
	}
	beforeNoop["route_idempotency"] = afterNoop["route_idempotency"]
	requireNoteAssociationEffects(t, beforeNoop, afterNoop)
	var receipts int
	if err := h.DB.QueryRowContext(context.Background(), `SELECT count(*) FROM route_idempotency WHERE scope_key=$1 AND actor_user_id=$2 AND client_txn_id=$3`, noteID.String(), adminID, "guards-noop").Scan(&receipts); err != nil || receipts != 1 {
		t.Fatalf("no-op keyed receipt count=%d err=%v", receipts, err)
	}
	exec(`UPDATE incidents SET status='closed',closed_at=now() WHERE id=$1`, incidentID)
	for _, replay := range []struct{ body, receipt map[string]any }{{acceptedBody, accepted}, {noopBody, noop}} {
		t.Run("HC13/"+replay.body["client_txn_id"].(string), func(t *testing.T) {
			before := snapshotNoteAssociationEffects(t, h, incidentID)
			got := success(http.MethodPost, route, replay.body, cookie(admin)...)
			if !reflect.DeepEqual(got, replay.receipt) {
				t.Fatal("retained receipt changed after closure")
			}
			requireNoteAssociationEffects(t, before, snapshotNoteAssociationEffects(t, h, incidentID))
		})
	}
	reject("HC18/closed-fresh", http.MethodPost, route, payload("guards-closed", currentVersion, otherIDs[3]), 409, "incident_closed", empty, cookie(admin)...)
	reject("HC06/viewer-before-body-and-closure", http.MethodPost, route, malformed, 403, "authorization_denied", roleDenied, cookie(viewer.login)...)
	reject("HC14/replay-CSRF", http.MethodPost, route, acceptedBody, 403, "csrf_verification_failed", empty, appsupport.WithCookies(admin.SessionCookie, admin.CSRFCookie))
	exec(`UPDATE incident_memberships SET role='viewer' WHERE incident_id=$1 AND user_id=$2`, incidentID, adminID)
	reject("HC14/replay-downgraded", http.MethodPost, route, acceptedBody, 403, "authorization_denied", roleDenied, cookie(admin)...)
	exec(`DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incidentID, adminID)
	reject("HC14/replay-revoked", http.MethodPost, route, acceptedBody, 404, "incident_not_found", empty, cookie(admin)...)
}

func requireNotePagingAgreement(t testing.TB, envelope map[string]any, limit int, hasMore bool) {
	t.Helper()
	data := envelope["data"].(map[string]any)
	expected := map[string]any{"limit": float64(limit), "has_more": hasMore, "next_cursor": data["next_cursor_token"]}
	if !reflect.DeepEqual(envelope["meta"].(map[string]any)["paging"], expected) {
		t.Fatal("data/meta paging disagreement")
	}
	if len(data["items"].([]any)) > limit || hasMore != (data["next_cursor_token"] != nil) {
		t.Fatal("invalid page continuation")
	}
}

// Full row content includes identities, provenance and tombstones, not just
// counts. Hashes keep raw Notes, cursor keys and hidden IDs out of diagnostics.
// Session/auth bookkeeping is permitted and intentionally outside domain state.
func snapshotNoteAssociationEffects(t testing.TB, h *appsupport.ServerHarness, incidentID uuid.UUID) map[string][32]byte {
	t.Helper()
	queries := map[string]string{}
	for _, table := range []string{"records", "artifacts", "record_links", "change_sets", "artifact_grid_projection", "evidence_grid_projection", "timeline_grid_projection"} {
		queries[table] = "SELECT * FROM " + table + " WHERE incident_id=$1"
	}
	queries["change_set_mutations"] = `SELECT m.* FROM change_set_mutations m JOIN change_sets c USING(change_set_id) WHERE c.incident_id=$1`
	queries["record_revisions"] = `SELECT v.* FROM record_revisions v JOIN records r USING(record_id) WHERE r.incident_id=$1`
	queries["record_history_entry_refs"] = `SELECT v.* FROM record_history_entry_refs v JOIN records r USING(record_id) WHERE r.incident_id=$1`
	queries["route_idempotency"] = `SELECT * FROM route_idempotency WHERE scope_key=$1::text OR scope_key IN (SELECT record_id::text FROM records WHERE incident_id=$1::uuid)`
	result := collaborationsupport.SnapshotPublicationState(t, h.DB, incidentID)
	for table, query := range queries {
		var data string
		if err := h.DB.QueryRowContext(context.Background(), `SELECT COALESCE(jsonb_agg(to_jsonb(x) ORDER BY to_jsonb(x)::text),'[]'::jsonb)::text FROM (`+query+`) x`, incidentID).Scan(&data); err != nil {
			t.Fatalf("snapshot %s: %v", table, err)
		}
		result[table] = sha256.Sum256([]byte(data))
	}
	return result
}

func requireNoteAssociationEffects(t testing.TB, before, after map[string][32]byte) {
	t.Helper()
	for table, expected := range before {
		if after[table] != expected {
			t.Errorf("unexpected Note domain effect in %s", table)
		}
	}
}
