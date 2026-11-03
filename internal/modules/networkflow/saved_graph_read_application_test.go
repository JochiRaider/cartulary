package networkflow

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"strings"
	"testing"
)

type savedReadProbe struct {
	savedGraphDeclarations
	savedGraphResults
	declaration graphViewDeclaration
	job         jobs.Resource
	err         error
	calls       int
}

func (p *savedReadProbe) GetGraphViewDeclaration(context.Context, uuid.UUID, string) (graphViewDeclaration, error) {
	return p.declaration, nil
}
func (p *savedReadProbe) Get(context.Context, uuid.UUID) (jobs.Resource, error) {
	p.calls++
	return p.job, p.err
}
func TestSavedGraphReadFailures_Unit(t *testing.T) {
	t.Run("derived availability follows authorization", func(t *testing.T) {
		for _, denied := range []bool{false, true} {
			access := &queryCompletionProbe{}
			if denied {
				access.admissionErr = &admission.Denied{Code: admission.DenialNotVisible}
			}
			probe := &savedReadProbe{declaration: graphViewDeclaration{DeclarationState: graphViewDeclarationStateActive, SelectedResult: &graphViewSelectedResultBinding{}}}
			readinessCalls := 0
			app := &savedGraphReadApplication{declarations: probe, results: probe, jobs: probe, incidentAccess: access, graphStateReady: func() bool { readinessCalls++; return false }}
			_, failure := app.result(t.Context(), readIdentity{ActorID: uuid.New(), IncidentID: uuid.New()}, "id")
			if failure == nil || probe.calls != 0 {
				t.Fatal("unavailable result consulted Jobs or succeeded")
			}
			if denied && (readinessCalls != 0 || semanticHTTPError(failure).Status != 404) {
				t.Fatal("availability preempted authorization")
			}
			if !denied && (readinessCalls != 1 || failure.kind != failureGraphProjectionFailed || failure.reason != "projection_unavailable") {
				t.Fatalf("unavailable classification=%#v", failure)
			}
		}
	})
	ctx := context.Background()
	id := uuid.New()
	failed := "failed"
	sentinel := errors.New("SENTINEL jobs database secret")
	for _, name := range []string{"no job", "retained failure", "missing history", "missing failed history", "queued", "running", "failed", "cancelled", "unexpected Jobs failure", "inconsistent success", "unknown status"} {
		t.Run(name, func(t *testing.T) {
			probe := &savedReadProbe{declaration: graphViewDeclaration{DeclarationState: graphViewDeclarationStateActive, LatestJobID: &id}}
			want := selectedResultPending
			internal := false
			switch name {
			case "no job":
				probe.declaration.LatestJobID = nil
			case "retained failure":
				probe.declaration.LatestJobID = nil
				probe.declaration.LastFailureCode = &failed
				want = selectedResultFailed
			case "missing history":
				probe.err = jobs.ErrNotFound
			case "missing failed history":
				probe.err = jobs.ErrNotFound
				probe.declaration.LastFailureCode = &failed
				want = selectedResultFailed
			case "queued":
				probe.job.Status = jobs.StatusQueued
			case "running":
				probe.job.Status = jobs.StatusRunning
			case "failed":
				probe.job.Status = jobs.StatusFailed
				want = selectedResultFailed
			case "cancelled":
				probe.job.Status = jobs.StatusCanceled
				want = selectedResultFailed
			case "unexpected Jobs failure":
				probe.err = sentinel
				probe.declaration.LastFailureCode = &failed
				internal = true
			case "inconsistent success":
				probe.job.Status = jobs.StatusSucceeded
				internal = true
			case "unknown status":
				probe.job.Status = "future status"
				internal = true
			}
			app := &savedGraphReadApplication{declarations: probe, results: probe, jobs: probe, incidentAccess: &queryCompletionProbe{}, graphStateReady: func() bool { return true }}
			who := readIdentity{ActorID: uuid.New(), IncidentID: uuid.New()}
			_, a := app.result(ctx, who, "id")
			_, b := app.contributors(ctx, who, "id", graphViewContributorRequest{})
			for _, failure := range []*semanticFailure{a, b} {
				if failure == nil {
					t.Fatal("missing result succeeded")
				}
				if internal && failure.kind != failureInternal || !internal && (failure.kind != failureGraphViewNotMaterialized || failure.reason != string(want)) {
					t.Fatalf("classification: %#v", failure)
				}
				if name == "unexpected Jobs failure" && !errors.Is(failure, sentinel) {
					t.Fatal("cause lost")
				}
				apiErr := semanticHTTPError(failure)
				if strings.Contains(apiErr.Message, "SENTINEL") {
					t.Fatal("cause disclosed")
				}
			}
			if probe.declaration.LatestJobID == nil && probe.calls != 0 {
				t.Fatal("unneeded jobs lookup")
			}
		})
	}
	probe := &savedReadProbe{declaration: graphViewDeclaration{SelectedResult: &graphViewSelectedResultBinding{}}}
	app := &savedGraphReadApplication{jobs: probe}
	if _, failure := app.missingSelectedResult(ctx, probe.declaration); failure == nil || probe.calls != 0 {
		t.Fatal("selected result manufactured job success")
	}
}
