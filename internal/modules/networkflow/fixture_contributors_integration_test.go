package networkflow_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
	"github.com/google/uuid"
)

// Preserve unexpected rejected queries for complete frozen comparison instead
// of skipping a case or substituting an expected error.
func (r *publicationFixtureRuntime) observedPost(path string, body any) map[string]any {
	response := httptestx.DoJSON(r.t, http.DefaultClient, http.MethodPost, r.h.Server.HTTP.URL+path, body, r.requestOptions...)
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		r.t.Fatal(err)
	}
	response.Body.Close()
	response.Body = io.NopCloser(bytes.NewReader(raw))
	r.transport = append(r.transport, map[string]any{"method": http.MethodPost, "path": path, "body": body, "status": response.StatusCode, "response": fixtureResponseEvidence(raw)})
	result := map[string]any{"status": response.StatusCode, "data": nil, "error": nil}
	if response.StatusCode >= 200 && response.StatusCode < 300 {
		result["data"] = httptestx.RequireSuccessEnvelope(r.t, response, response.StatusCode)["data"]
	} else {
		var envelope struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(raw, &envelope); err != nil {
			r.t.Fatal(err)
		}
		failure := httptestx.RequireErrorEnvelope(r.t, response, response.StatusCode, envelope.Error.Code)["error"].(map[string]any)
		failure["request_id"] = "$request-id"
		result["error"] = failure
	}
	return result
}

func TestNetworkFlowContributorsFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-025-graph-contributors", 3, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runGraphFixtureFamily(t, bundle, func(r *publicationFixtureRuntime, input graphFixtureCase, result map[string]any) error {
			ctx := context.Background()
			root := "/api/v1/incidents/" + input.IncidentID + "/network-flow"
			path := root + "/graphs/contributors/query"
			graph := result["graph_queries"].([]map[string]any)[0]["response"].(map[string]any)
			vertex := graphEndpointSelector(graph, "192.0.2.10")
			if vertex == nil {
				return fmt.Errorf("missing contributor endpoint")
			}
			edges := graph["edge_annotations"].([]any)
			if len(edges) != 1 {
				return fmt.Errorf("expected one imported shared edge")
			}
			edge := edges[0].(map[string]any)["selector"]
			initial := graphContributorRequest(graph, edge, 1)
			next := initial
			pages := []map[string]any{}
			payloads := []any{}
			firstToken := ""
			for index := 0; index < 5; index++ {
				page, token, payload := r.normalizedPage(r.request(http.MethodPost, path, next, 200), fmt.Sprintf("edge-%d", index+1))
				pages = append(pages, page)
				payloads = append(payloads, payload)
				if index == 0 {
					firstToken = token
				}
				if token == "" {
					break
				}
				if index == 4 {
					return fmt.Errorf("contributor pagination did not terminate")
				}
				next = map[string]any{"schema_id": "cartulary.network_flow.graph_contributor_query_continuation.v1", "cursor_token": token}
			}
			if len(pages) != 4 || firstToken == "" {
				return fmt.Errorf("expected four accepted contributors across two tables")
			}
			vertexRequest := graphContributorRequest(graph, vertex, 200)
			vertexPage := r.request(http.MethodPost, path, vertexRequest, 200)
			restoreActor := r.alternateActor(input.IncidentID)
			viewerPage := r.request(http.MethodPost, path, vertexRequest, 200)
			restoreActor()
			omitted := graphContributorRequest(graph, vertex, 200)
			delete(omitted, "limit")
			defaulted := r.observedPost(path, omitted)
			stale := graphContributorRequest(graph, edge, 1)
			stale["graph_query_digest"] = strings.Repeat("f", 64)
			staleError := r.observedError(http.MethodPost, path, stale)
			missing := graphContributorRequest(graph, input.MissingSelector, 1)
			missingError := r.observedError(http.MethodPost, path, missing)
			binding := networkflowsupport.AuthorizationBinding{ActorRef: "actor:contributors", IncidentRef: "incident:contributors", ResourceKind: hc.NetworkFlowAuthResourceNetworkFlowContributors, ResourceRef: "contributors:fixture", ActorID: uuid.MustParse(input.ActorID), IncidentID: uuid.MustParse(input.IncidentID), SessionID: r.authSession, TableID: "nft_" + strings.ReplaceAll(input.TableUUID, "-", "")}
			fixture, err := networkflowsupport.NewAuthorizationFixture(ctx, r.controls.Transitions, r.h.Pool, binding, r.h.Server.Clock.Now)
			if err != nil {
				return err
			}
			transition := func(kind string) error {
				boundary := hc.NetworkFlowAuthTransitionBoundaryRouteBeforeAuthorization
				body := map[string]any{"boundary": boundary, "transition_kind": kind, "actor_ref": binding.ActorRef, "incident_ref": binding.IncidentRef, "resource_kind": binding.ResourceKind, "resource_ref": binding.ResourceRef, "correlation_key": "fixture-contributors", "consume_once": true}
				httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, r.h.Server.HTTP.URL+"/api/v1/test/runtime/network-flow-auth-transitions", body, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), 201)
				consumed, err := fixture.Before(ctx, boundary, "fixture-contributors")
				if err != nil {
					return err
				}
				if !consumed {
					return fmt.Errorf("authorization transition was not consumed")
				}
				r.transport = append(r.transport, map[string]any{"authorization_transition": body, "consumed": consumed})
				return nil
			}
			before := r.durableState()
			continuation := map[string]any{"schema_id": "cartulary.network_flow.graph_contributor_query_continuation.v1", "cursor_token": firstToken}
			if err := transition(hc.NetworkFlowAuthTransitionKindIncidentMembershipRevoked); err != nil {
				return err
			}
			denied := map[string]any{"graph": r.observedError(http.MethodPost, root+"/graphs/query", input.Queries[0].Request), "initial": r.observedError(http.MethodPost, path, initial), "continuation": r.observedError(http.MethodPost, path, continuation)}
			if !reflect.DeepEqual(before, r.durableState()) {
				return fmt.Errorf("unauthorized queries mutated owner state")
			}
			if err := transition(hc.NetworkFlowAuthTransitionKindIncidentMembershipRestored); err != nil {
				return err
			}
			recovered, _, recoveredPayload := r.normalizedPage(r.request(http.MethodPost, path, continuation, 200), "edge-2")
			if !reflect.DeepEqual(recovered, pages[1]) || !reflect.DeepEqual(recoveredPayload, payloads[1]) {
				return fmt.Errorf("authorization restoration changed live contributor continuation")
			}
			result["contributors"] = map[string]any{"edge_request": initial, "pages": pages, "cursor_payloads": payloads, "vertex_request": vertexRequest, "vertex_page": vertexPage, "viewer_page": viewerPage, "omitted_limit": defaulted, "stale_digest": staleError, "missing_vertex": missingError, "authorization_loss": denied, "recovered_page": recovered, "recovered_cursor": recoveredPayload}
			return nil
		})
	})
}
