package jobapi_test

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentJobAuthorizationReDerivedAtRequestTime_Unit(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-jobapi-incident-auth", TestRouteMode: httptestx.TestRouteModeDisabled})
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-JOBAPI",
		"title":         "Enterprise integration job API auth",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	viewerPassword := "ViewerPassphrase11!"
	peerPassword := "PeerPassphrase11!"
	viewerUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-viewer@example.test", "ExtensionProfile Viewer", viewerPassword, false, false, true)
	peerUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-peer@example.test", "ExtensionProfile Peer", peerPassword, false, false, true)
	viewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, viewerUser.Email, viewerPassword, nil)
	peerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, peerUser.Email, peerPassword, nil)

	viewerMembership := scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-viewer-membership",
		"user_id":       viewerUser.ID.String(),
		"role":          "viewer",
	})
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-peer-membership",
		"user_id":       peerUser.ID.String(),
		"role":          "viewer",
	})

	job, err := createJobAPIFixture(context.Background(), harness.Jobs, jobs.EnqueueParams{
		JobKind:           "import.discovery_v1",
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: viewerUser.ID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create incident job: %v", err)
	}

	viewerRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(viewerLogin.Session))
	httptestx.RequireSuccessEnvelope(t, viewerRead, http.StatusOK)

	// REQ-01-704 / AC-612 (Core 01 §3.3.9.1): a member who neither submitted
	// the job nor administers its incident receives the missing job's
	// response, before the body is classified, and commits nothing.
	peerCancel := func(jobID string, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", body,
			httptestx.WithCookies(peerLogin.Session, peerLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, peerLogin.CSRF.Value))
	}
	beforePeerCancel := jobCancelEffects(t, harness.DB, job.JobID)
	missingJobID := uuid.NewString()
	requireJobResponsesAlike(t,
		peerCancel(job.JobID, map[string]any{"client_txn_id": "txn-extension_profile-jobapi-peer-cancel"}),
		peerCancel(missingJobID, map[string]any{"client_txn_id": "txn-extension_profile-jobapi-peer-cancel"}),
	)
	requireJobResponsesAlike(t, peerCancel(job.JobID, []any{1}), peerCancel(missingJobID, []any{1}))
	if after := jobCancelEffects(t, harness.DB, job.JobID); after != beforePeerCancel {
		t.Fatalf("concealed cancel committed an effect: before=%s after=%s", beforePeerCancel, after)
	}

	scenariotest.DeleteMembership(t, harness.Server, adminLogin, incidentID.String(), viewerUser.ID.String(), map[string]any{
		"base_membership_version": viewerMembership["membership_version"],
	})

	afterDeleteRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(viewerLogin.Session))
	httptestx.RequireErrorEnvelope(t, afterDeleteRead, http.StatusNotFound, "job_not_found")

	afterDeleteCancel := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-viewer-cancel-after-delete",
	}, httptestx.WithCookies(viewerLogin.Session, viewerLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, viewerLogin.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, afterDeleteCancel, http.StatusNotFound, "job_not_found")
}

func TestDeploymentJobAuthorizationReDerivedAtRequestTime_Unit(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-jobapi-deployment-auth", TestRouteMode: httptestx.TestRouteModeDisabled})
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	submitterPassword := "SubmitterPassphrase11!"
	otherPassword := "OtherPassphrase11!"
	submitterUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-submitter@example.test", "ExtensionProfile Submitter", submitterPassword, false, false, true)
	otherUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-other@example.test", "ExtensionProfile Other", otherPassword, false, false, true)
	submitterLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, submitterUser.Email, submitterPassword, nil)
	otherLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, otherUser.Email, otherPassword, nil)

	job, err := createJobAPIFixture(context.Background(), harness.Jobs, jobs.EnqueueParams{
		JobKind:           "import.discovery_v1",
		Scope:             jobs.Scope{Kind: jobs.ScopeKindDeployment},
		SubmittedByUserID: submitterUser.ID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create deployment job: %v", err)
	}

	adminRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(adminLogin.Session))
	if read := httptestx.RequireSuccessData(t, adminRead, http.StatusOK); read["job_id"] != job.JobID || read["status"] != "queued" {
		t.Fatalf("admin job read = %#v, want queued %s", read, job.JobID)
	}

	otherRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(otherLogin.Session))
	httptestx.RequireErrorEnvelope(t, otherRead, http.StatusNotFound, "job_not_found")

	otherCancel := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-other-cancel",
	}, httptestx.WithCookies(otherLogin.Session, otherLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, otherLogin.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, otherCancel, http.StatusNotFound, "job_not_found")

	cancelURL := harness.Server.HTTP.URL + "/api/v1/jobs/" + job.JobID + "/cancel"
	submitterCancelRequest := func(body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, cancelURL, body,
			httptestx.WithCookies(submitterLogin.Session, submitterLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, submitterLogin.CSRF.Value))
	}
	requireErrorDetails := func(resp *http.Response, status int, code string, want map[string]any) {
		t.Helper()
		envelope := httptestx.RequireErrorEnvelope(t, resp, status, code)
		details, _ := envelope["error"].(map[string]any)["details"].(map[string]any)
		if !reflect.DeepEqual(details, want) {
			t.Fatalf("%s details = %#v, want %#v", code, details, want)
		}
	}
	for _, malformed := range []struct {
		body any
		want map[string]any
	}{
		{body: []any{1}, want: map[string]any{"reason_code": "request_not_object"}},
		{body: map[string]any{}, want: map[string]any{"field": "client_txn_id", "reason_code": "missing_required_field"}},
		{body: map[string]any{"client_txn_id": "txn-unknown-member", "priority": 1}, want: map[string]any{"field": "priority", "reason_code": "unknown_field"}},
	} {
		requireErrorDetails(submitterCancelRequest(malformed.body), http.StatusBadRequest, "invalid_mutation_payload", malformed.want)
	}

	submitterCancel := submitterCancelRequest(map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-submitter-cancel",
	})
	if canceled := httptestx.RequireSuccessData(t, submitterCancel, http.StatusOK); canceled["job_id"] != job.JobID || canceled["status"] != "cancel_requested" || canceled["cancelable"] != false {
		t.Fatalf("submitter cancel = %#v, want cancel_requested non-cancelable %s", canceled, job.JobID)
	}
	replay := submitterCancelRequest(map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-submitter-cancel",
		"reason":        nil,
	})
	if replayed := httptestx.RequireSuccessData(t, replay, http.StatusOK); replayed["job_id"] != job.JobID || replayed["status"] != "cancel_requested" {
		t.Fatalf("null-reason cancel replay = %#v, want current cancel_requested %s", replayed, job.JobID)
	}
	requireErrorDetails(submitterCancelRequest(map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-submitter-cancel",
		"reason":        "different reason",
	}), http.StatusConflict, "client_txn_conflict", map[string]any{"client_txn_id": "txn-extension_profile-jobapi-submitter-cancel"})
	requireErrorDetails(submitterCancelRequest(map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-submitter-second-cancel",
	}), http.StatusConflict, "job_cancel_rejected", map[string]any{"reason_code": "already_cancel_requested"})
}

func TestExpiredJobReadAndCancelAreMasked_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "jobapi-expired-masking", TestRouteMode: httptestx.TestRouteModeDisabled})
	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	job, err := createJobAPIFixture(context.Background(), harness.Jobs, jobs.EnqueueParams{
		JobKind:           "import.discovery_v1",
		Scope:             jobs.Scope{Kind: jobs.ScopeKindDeployment},
		SubmittedByUserID: adminUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	clientTxnID := "txn-expired-job-cancel"
	cancelBody := map[string]any{"client_txn_id": clientTxnID}
	cancel := httptestx.DoJSON(
		t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel",
		cancelBody,
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireSuccessEnvelope(t, cancel, http.StatusOK)
	if _, err := harness.DB.Exec(`
UPDATE jobs
   SET status = 'canceled',
       finished_at = now() - interval '7 days',
       updated_at = now() - interval '7 days',
       retained_until = now() - interval '1 second',
       result_summary_json = '{"code":"job_canceled","message":"Job canceled."}',
       handler_attempt_id = NULL,
       handler_lease_expires_at = NULL,
       handler_next_attempt_at = NULL
 WHERE job_id = $1
`, uuid.MustParse(job.JobID)); err != nil {
		t.Fatal(err)
	}

	read := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(adminLogin.Session))
	httptestx.RequireErrorEnvelope(t, read, http.StatusNotFound, "job_not_found")
	replay := httptestx.DoJSON(
		t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel",
		cancelBody,
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, replay, http.StatusNotFound, "job_not_found")
}

func TestDeploymentAdminIncidentMembershipPolicy_Unit(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-jobapi-incident-admin-member-auth", TestRouteMode: httptestx.TestRouteModeDisabled})
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-incident",
		"incident_key":  "IR-EXTENSION-PROFILE-JOBAPI-ADMIN-MEMBER",
		"title":         "Enterprise integration admin-member job API auth",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	submitterPassword := "SubmitterAdminMemberPassphrase11!"
	deploymentViewerPassword := "DeploymentViewerPassphrase11!"
	incidentAdminPassword := "IncidentAdminPassphrase11!"
	deploymentNonMemberPassword := "DeploymentNonMemberPassphrase11!"
	submitterUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-admin-member-submitter@example.test", "ExtensionProfile Submitter Admin Member", submitterPassword, false, true, true)
	deploymentViewerUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-deployment-viewer@example.test", "ExtensionProfile Deployment Viewer", deploymentViewerPassword, false, true, true)
	incidentAdminUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-incident-admin@example.test", "ExtensionProfile Incident Admin", incidentAdminPassword, false, false, true)
	deploymentNonMemberUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-deployment-nonmember@example.test", "ExtensionProfile Deployment Nonmember", deploymentNonMemberPassword, false, true, true)
	submitterLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, submitterUser.Email, submitterPassword, nil)
	deploymentViewerLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, deploymentViewerUser.Email, deploymentViewerPassword, nil)
	incidentAdminLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, incidentAdminUser.Email, incidentAdminPassword, nil)
	deploymentNonMemberLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, deploymentNonMemberUser.Email, deploymentNonMemberPassword, nil)

	submitterMembership := scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-submitter-membership",
		"user_id":       submitterUser.ID.String(),
		"role":          "admin",
	})
	deploymentViewerMembership := scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-deployment-viewer-membership",
		"user_id":       deploymentViewerUser.ID.String(),
		"role":          "viewer",
	})
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-incident-admin-membership",
		"user_id":       incidentAdminUser.ID.String(),
		"role":          "admin",
	})

	readJob, err := createJobAPIFixture(context.Background(), harness.Jobs, jobs.EnqueueParams{
		JobKind:           "import.discovery_v1",
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: submitterUser.ID,
		AuthPolicy:        jobs.AuthPolicyDeploymentAdminIncidentMembership,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create admin-member read job: %v", err)
	}

	submitterRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, httptestx.WithCookies(submitterLogin.Session))
	httptestx.RequireSuccessEnvelope(t, submitterRead, http.StatusOK)

	deploymentViewerRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, httptestx.WithCookies(deploymentViewerLogin.Session))
	httptestx.RequireSuccessEnvelope(t, deploymentViewerRead, http.StatusOK)

	incidentAdminRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, httptestx.WithCookies(incidentAdminLogin.Session))
	httptestx.RequireErrorEnvelope(t, incidentAdminRead, http.StatusNotFound, "job_not_found")

	deploymentNonMemberRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, httptestx.WithCookies(deploymentNonMemberLogin.Session))
	httptestx.RequireErrorEnvelope(t, deploymentNonMemberRead, http.StatusNotFound, "job_not_found")

	viewerCancel := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-viewer-cancel",
	}, httptestx.WithCookies(deploymentViewerLogin.Session, deploymentViewerLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, deploymentViewerLogin.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, viewerCancel, http.StatusNotFound, "job_not_found")

	deploymentViewerAdmin := scenariotest.PatchMembership(t, harness.Server, adminLogin, incidentID.String(), deploymentViewerUser.ID.String(), map[string]any{
		"base_membership_version": deploymentViewerMembership["membership_version"],
		"role":                    "admin",
	})
	_ = deploymentViewerAdmin
	viewerAdminCancel := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-admin-cancel",
	}, httptestx.WithCookies(deploymentViewerLogin.Session, deploymentViewerLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, deploymentViewerLogin.CSRF.Value))
	httptestx.RequireSuccessEnvelope(t, viewerAdminCancel, http.StatusOK)

	demotedJob, err := createJobAPIFixture(context.Background(), harness.Jobs, jobs.EnqueueParams{
		JobKind:           "import.discovery_v1",
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: submitterUser.ID,
		AuthPolicy:        jobs.AuthPolicyDeploymentAdminIncidentMembership,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create admin-member demotion job: %v", err)
	}
	scenariotest.PatchMembership(t, harness.Server, adminLogin, incidentID.String(), submitterUser.ID.String(), map[string]any{
		"base_membership_version": submitterMembership["membership_version"],
		"role":                    "viewer",
	})
	if _, err := harness.DB.Exec(`UPDATE users SET is_deployment_admin = false WHERE id = $1`, submitterUser.ID); err != nil {
		t.Fatalf("demote submitter deployment admin flag: %v", err)
	}
	demotedRead := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+demotedJob.JobID, nil, httptestx.WithCookies(submitterLogin.Session))
	httptestx.RequireErrorEnvelope(t, demotedRead, http.StatusNotFound, "job_not_found")
	demotedCancel := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+demotedJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-demoted-cancel",
	}, httptestx.WithCookies(submitterLogin.Session, submitterLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, submitterLogin.CSRF.Value))
	httptestx.RequireErrorEnvelope(t, demotedCancel, http.StatusNotFound, "job_not_found")
}

// TestJobRouteAdmissionOrder_Integration covers Core 01 REQ-01-707 and AC-617
// for the central-admission pilot: route resolution rejects undeclared methods,
// HEAD, unknown descendants and non-canonical paths with registered envelopes;
// authentication and CSRF precede path parameters; a malformed or encoded-slash
// job_id receives the response for a missing job, never reaching another
// operation; visibility precedes the singleton-query check; and a non-member
// still receives the missing job's response (REQ-01-704).
func TestJobRouteAdmissionOrder_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "jobapi-admission-order", TestRouteMode: httptestx.TestRouteModeDisabled})
	base := harness.Server.HTTP.URL
	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, base)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-jobapi-admission-order-incident",
		"incident_key":  "IR-JOBAPI-ADMISSION-ORDER",
		"title":         "Job route admission order",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	outsiderPassword := "OutsiderAdmissionPassphrase11!"
	outsiderUser := flowtest.SeedLocalUserRecord(t, harness.DB, "jobapi-admission-outsider@example.test", "Admission Outsider", outsiderPassword, false, false, true)
	outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, base, outsiderUser.Email, outsiderPassword, nil)
	job, err := createJobAPIFixture(context.Background(), harness.Jobs, jobs.EnqueueParams{
		JobKind:           "import.discovery_v1",
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: adminUserID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create incident job: %v", err)
	}
	jobURL := base + "/api/v1/jobs/" + job.JobID
	missingURL := base + "/api/v1/jobs/" + uuid.NewString()
	malformedURL := base + "/api/v1/jobs/not-a-job-id"
	// A redirect is never followed, so a 3xx would surface as itself.
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	// Each helper takes the calling (sub)test's t, so a failure inside t.Run
	// fails that subtest from its own goroutine.
	get := func(t testing.TB, url string, options ...func(*http.Request)) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, http.MethodGet, url, nil, options...)
	}
	cancel := func(t testing.TB, url string, login authcookietest.AuthCookies, body any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, client, http.MethodPost, url, body,
			httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value))
	}
	cancelBody := map[string]any{"client_txn_id": "txn-jobapi-admission-order-cancel"}
	requireDetails := func(t testing.TB, body map[string]any, want map[string]any) {
		t.Helper()
		if details, _ := body["error"].(map[string]any)["details"].(map[string]any); !reflect.DeepEqual(details, want) {
			t.Fatalf("details = %#v, want %#v", details, want)
		}
	}
	requireMethodNotAllowed := func(t testing.TB, response *http.Response, allow string) {
		t.Helper()
		if response.Header.Get("Allow") != allow {
			t.Fatalf("Allow = %q, want %q", response.Header.Get("Allow"), allow)
		}
		requireDetails(t, httptestx.RequireErrorEnvelope(t, response, http.StatusMethodNotAllowed, "method_not_allowed"), map[string]any{"allowed_methods": []any{allow}})
	}
	httptestx.RequireSuccessEnvelope(t, get(t, jobURL, httptestx.WithCookies(adminLogin.Session)), http.StatusOK)
	httptestx.RequireSuccessEnvelope(t, get(t, base+"/api/v1/jobs/"+strings.ToUpper(job.JobID), httptestx.WithCookies(adminLogin.Session)), http.StatusOK)
	before := jobCancelEffects(t, harness.DB, job.JobID)

	t.Run("authentication and CSRF precede path parameters", func(t *testing.T) {
		httptestx.RequireErrorEnvelope(t, get(t, malformedURL), http.StatusUnauthorized, "session_required")
		httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, client, http.MethodPost, malformedURL+"/cancel", cancelBody), http.StatusUnauthorized, "session_required")
		withoutCSRF := httptestx.DoJSON(t, client, http.MethodPost, malformedURL+"/cancel", cancelBody, httptestx.WithCookies(adminLogin.Session))
		httptestx.RequireErrorEnvelope(t, withoutCSRF, http.StatusForbidden, "csrf_verification_failed")
	})
	t.Run("a malformed job_id is a missing job", func(t *testing.T) {
		for _, malformed := range []string{malformedURL, base + "/api/v1/jobs/" + strings.ReplaceAll(job.JobID, "-", ""), base + "/api/v1/jobs/%7B" + job.JobID + "%7D"} {
			requireJobResponsesAlike(t, get(t, malformed, httptestx.WithCookies(adminLogin.Session)), get(t, missingURL, httptestx.WithCookies(adminLogin.Session)))
		}
		requireJobResponsesAlike(t, cancel(t, malformedURL+"/cancel", adminLogin, cancelBody), cancel(t, missingURL+"/cancel", adminLogin, cancelBody))
	})
	t.Run("an encoded slash stays inside job_id", func(t *testing.T) {
		requireJobResponsesAlike(t, get(t, jobURL+"%2Fcancel", httptestx.WithCookies(adminLogin.Session)), get(t, missingURL, httptestx.WithCookies(adminLogin.Session)))
		requireJobResponsesAlike(t, cancel(t, jobURL+"%2Fx/cancel", adminLogin, cancelBody), cancel(t, missingURL+"/cancel", adminLogin, cancelBody))
		// POST on the GET template: the encoded slash never selects cancel.
		requireMethodNotAllowed(t, cancel(t, jobURL+"%2Fcancel", adminLogin, cancelBody), "GET")
		if after := jobCancelEffects(t, harness.DB, job.JobID); after != before {
			t.Fatalf("encoded-slash requests committed an effect: before=%s after=%s", before, after)
		}
	})
	t.Run("route resolution rejections are enveloped", func(t *testing.T) {
		head := httptestx.DoJSON(t, client, http.MethodHead, jobURL, nil, httptestx.WithCookies(adminLogin.Session))
		body, err := io.ReadAll(head.Body)
		_ = head.Body.Close()
		if err != nil || head.StatusCode != http.StatusMethodNotAllowed || head.Header.Get("Allow") != "GET" || len(body) != 0 {
			t.Fatalf("HEAD = %d Allow=%q body=%q err=%v, want 405 Allow GET without a body", head.StatusCode, head.Header.Get("Allow"), body, err)
		}
		requireMethodNotAllowed(t, httptestx.DoJSON(t, client, http.MethodDelete, jobURL, nil, httptestx.WithCookies(adminLogin.Session)), "GET")
		requireMethodNotAllowed(t, get(t, jobURL+"/cancel", httptestx.WithCookies(adminLogin.Session)), "POST")
		requireDetails(t, httptestx.RequireErrorEnvelope(t, get(t, jobURL+"/unknown", httptestx.WithCookies(adminLogin.Session)), http.StatusNotFound, "route_not_found"), map[string]any{})
		for _, noncanonical := range []string{base + "/api/v1/jobs//" + job.JobID, base + "/api/v1/jobs/x/../" + job.JobID, base + "/api/v1/./jobs/" + job.JobID} {
			response := get(t, noncanonical, httptestx.WithCookies(adminLogin.Session))
			if location := response.Header.Get("Location"); location != "" {
				t.Fatalf("%s redirected to %s", noncanonical, location)
			}
			requireDetails(t, httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "route_not_found"), map[string]any{})
		}
	})
	t.Run("visibility precedes route handling", func(t *testing.T) {
		requireJobResponsesAlike(t, get(t, jobURL, httptestx.WithCookies(outsiderLogin.Session)), get(t, missingURL, httptestx.WithCookies(outsiderLogin.Session)))
		requireJobResponsesAlike(t, cancel(t, jobURL+"/cancel", outsiderLogin, []any{1}), cancel(t, missingURL+"/cancel", outsiderLogin, []any{1}))
		requireJobResponsesAlike(t, get(t, jobURL+"?limit=1", httptestx.WithCookies(outsiderLogin.Session)), get(t, missingURL+"?limit=1", httptestx.WithCookies(outsiderLogin.Session)))
		httptestx.RequireErrorEnvelope(t, get(t, jobURL+"?limit=1", httptestx.WithCookies(adminLogin.Session)), http.StatusBadRequest, "invalid_pagination_request")
		if after := jobCancelEffects(t, harness.DB, job.JobID); after != before {
			t.Fatalf("concealed requests committed an effect: before=%s after=%s", before, after)
		}
	})
}

// requireJobResponsesAlike requires a concealed cancel to match a missing
// job's cancel in everything but its request identity (REQ-01-704).
func requireJobResponsesAlike(t testing.TB, existing, missing *http.Response) {
	t.Helper()
	bodies := make([]map[string]any, 0, 2)
	for _, response := range []*http.Response{existing, missing} {
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "job_not_found")
		delete(body["error"].(map[string]any), "request_id")
		bodies = append(bodies, body)
	}
	if !reflect.DeepEqual(bodies[0], bodies[1]) {
		t.Fatalf("existing-job cancel response %v differs from missing-job response %v", bodies[0], bodies[1])
	}
}

// jobCancelEffects renders the job row, cancel receipts and Collaboration
// intents a cancel could commit.
func jobCancelEffects(t testing.TB, db *sql.DB, jobID string) string {
	t.Helper()
	var effects string
	if err := db.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'job',(SELECT to_jsonb(j) FROM jobs j WHERE job_id=$1::uuid),
 'receipts',(SELECT coalesce(jsonb_agg(to_jsonb(i) ORDER BY id),'[]') FROM route_idempotency i WHERE route_key='jobs.cancel')
 )::text`, jobID).Scan(&effects); err != nil {
		t.Fatal(err)
	}
	return effects + " collaboration_intents=" + strconv.Itoa(collaborationsupport.CountIntents(t, db, collaborationsupport.IntentSelector{}))
}

func createJobAPIFixture(
	ctx context.Context,
	capability *httptestx.JobsCapability,
	params jobs.EnqueueParams,
) (jobs.Resource, error) {
	scopeKey := jobs.ScopeKindDeployment
	if params.Scope.IncidentID != nil {
		scopeKey = params.Scope.IncidentID.String()
	}
	clientTxnID := "jobapi-fixture-" + uuid.NewString()
	normalizedRequest := []byte(`{"client_txn_id":"` + clientTxnID + `"}`)
	admission, err := jobs.NewExtensionJobAdmission(
		"import",
		jobs.NewRouteIdempotencyKey("test.jobapi.enqueue", params.SubmittedByUserID, scopeKey, clientTxnID),
		params.Scope,
		normalizedRequest,
	)
	if err != nil {
		return jobs.Resource{}, err
	}
	params.Extension = admission
	return capability.Create(ctx, params)
}
