package jobapi_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/phase2test"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestPhase11_U_11_JOBS_01_IncidentJobAuthorizationReDerivedAtRequestTime(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-jobapi-incident-auth")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase11-jobapi-incident",
		"incident_key":  "IR-PHASE11-JOBAPI",
		"title":         "Phase 11 job API auth",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	viewerPassword := "ViewerPassphrase11!"
	peerPassword := "PeerPassphrase11!"
	viewerUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-job-viewer@example.test", "Phase11 Viewer", viewerPassword, false, false, true)
	peerUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-job-peer@example.test", "Phase11 Peer", peerPassword, false, false, true)
	viewerCookies, viewerCSRF := phase2test.LoginLocalUser(t, harness.Server, viewerUser.Email, viewerPassword)
	peerCookies, peerCSRF := phase2test.LoginLocalUser(t, harness.Server, peerUser.Email, peerPassword)

	viewerMembership := phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-phase11-jobapi-viewer-membership",
		"user_id":       viewerUser.ID.String(),
		"role":          "viewer",
	})
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-phase11-jobapi-peer-membership",
		"user_id":       peerUser.ID.String(),
		"role":          "viewer",
	})

	job, err := harness.Server.Runtime.Jobs.Create(context.Background(), jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: viewerUser.ID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create incident job: %v", err)
	}

	viewerRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, phase2test.WithCookies(viewerCookies))
	httptestx.RequireSuccessEnvelope(t, viewerRead, http.StatusOK)

	peerCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-phase11-jobapi-peer-cancel",
	}, phase2test.WithCookies(peerCookies, peerCSRF), phase2test.WithHeader(authn.CSRFHeaderName, peerCSRF.Value))
	httptestx.RequireErrorEnvelope(t, peerCancel, http.StatusForbidden, "authorization_denied")

	phase2test.DeleteMembership(t, harness.Server, adminLogin, incidentID.String(), viewerUser.ID.String(), map[string]any{
		"base_membership_version": viewerMembership["membership_version"],
	})

	afterDeleteRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, phase2test.WithCookies(viewerCookies))
	httptestx.RequireErrorEnvelope(t, afterDeleteRead, http.StatusNotFound, "job_not_found")

	afterDeleteCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-phase11-jobapi-viewer-cancel-after-delete",
	}, phase2test.WithCookies(viewerCookies, viewerCSRF), phase2test.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value))
	httptestx.RequireErrorEnvelope(t, afterDeleteCancel, http.StatusNotFound, "job_not_found")
}

func TestPhase11_U_11_JOBS_02_DeploymentJobAuthorizationReDerivedAtRequestTime(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-jobapi-deployment-auth")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	submitterPassword := "SubmitterPassphrase11!"
	otherPassword := "OtherPassphrase11!"
	submitterUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-job-submitter@example.test", "Phase11 Submitter", submitterPassword, false, false, true)
	otherUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-job-other@example.test", "Phase11 Other", otherPassword, false, false, true)
	submitterCookies, submitterCSRF := phase2test.LoginLocalUser(t, harness.Server, submitterUser.Email, submitterPassword)
	otherCookies, otherCSRF := phase2test.LoginLocalUser(t, harness.Server, otherUser.Email, otherPassword)

	job, err := harness.Server.Runtime.Jobs.Create(context.Background(), jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindDeployment},
		SubmittedByUserID: submitterUser.ID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create deployment job: %v", err)
	}

	adminRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, phase2test.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireSuccessEnvelope(t, adminRead, http.StatusOK)

	otherRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, phase2test.WithCookies(otherCookies))
	httptestx.RequireErrorEnvelope(t, otherRead, http.StatusNotFound, "job_not_found")

	otherCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-phase11-jobapi-other-cancel",
	}, phase2test.WithCookies(otherCookies, otherCSRF), phase2test.WithHeader(authn.CSRFHeaderName, otherCSRF.Value))
	httptestx.RequireErrorEnvelope(t, otherCancel, http.StatusNotFound, "job_not_found")

	submitterCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-phase11-jobapi-submitter-cancel",
	}, phase2test.WithCookies(submitterCookies, submitterCSRF), phase2test.WithHeader(authn.CSRFHeaderName, submitterCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, submitterCancel, http.StatusOK)
}

func TestPhase11_U_11_JOBS_03_DeploymentAdminIncidentMembershipPolicy(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-jobapi-incident-admin-member-auth")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase11-jobapi-admin-member-incident",
		"incident_key":  "IR-PHASE11-JOBAPI-ADMIN-MEMBER",
		"title":         "Phase 11 admin-member job API auth",
	})
	incidentID := uuid.MustParse(incident["incident_id"].(string))

	submitterPassword := "SubmitterAdminMemberPassphrase11!"
	deploymentViewerPassword := "DeploymentViewerPassphrase11!"
	incidentAdminPassword := "IncidentAdminPassphrase11!"
	deploymentNonMemberPassword := "DeploymentNonMemberPassphrase11!"
	submitterUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-job-admin-member-submitter@example.test", "Phase11 Submitter Admin Member", submitterPassword, false, true, true)
	deploymentViewerUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-job-deployment-viewer@example.test", "Phase11 Deployment Viewer", deploymentViewerPassword, false, true, true)
	incidentAdminUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-job-incident-admin@example.test", "Phase11 Incident Admin", incidentAdminPassword, false, false, true)
	deploymentNonMemberUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-job-deployment-nonmember@example.test", "Phase11 Deployment Nonmember", deploymentNonMemberPassword, false, true, true)
	submitterCookies, submitterCSRF := phase2test.LoginLocalUser(t, harness.Server, submitterUser.Email, submitterPassword)
	deploymentViewerCookies, deploymentViewerCSRF := phase2test.LoginLocalUser(t, harness.Server, deploymentViewerUser.Email, deploymentViewerPassword)
	incidentAdminCookies, _ := phase2test.LoginLocalUser(t, harness.Server, incidentAdminUser.Email, incidentAdminPassword)
	deploymentNonMemberCookies, _ := phase2test.LoginLocalUser(t, harness.Server, deploymentNonMemberUser.Email, deploymentNonMemberPassword)

	submitterMembership := phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-phase11-jobapi-admin-member-submitter-membership",
		"user_id":       submitterUser.ID.String(),
		"role":          "admin",
	})
	deploymentViewerMembership := phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-phase11-jobapi-admin-member-deployment-viewer-membership",
		"user_id":       deploymentViewerUser.ID.String(),
		"role":          "viewer",
	})
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID.String(), map[string]any{
		"client_txn_id": "txn-phase11-jobapi-admin-member-incident-admin-membership",
		"user_id":       incidentAdminUser.ID.String(),
		"role":          "admin",
	})

	readJob, err := harness.Server.Runtime.Jobs.Create(context.Background(), jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: submitterUser.ID,
		AuthPolicy:        jobs.AuthPolicyDeploymentAdminIncidentMembership,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create admin-member read job: %v", err)
	}

	submitterRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, phase2test.WithCookies(submitterCookies))
	httptestx.RequireSuccessEnvelope(t, submitterRead, http.StatusOK)

	deploymentViewerRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, phase2test.WithCookies(deploymentViewerCookies))
	httptestx.RequireSuccessEnvelope(t, deploymentViewerRead, http.StatusOK)

	incidentAdminRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, phase2test.WithCookies(incidentAdminCookies))
	httptestx.RequireErrorEnvelope(t, incidentAdminRead, http.StatusNotFound, "job_not_found")

	deploymentNonMemberRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, phase2test.WithCookies(deploymentNonMemberCookies))
	httptestx.RequireErrorEnvelope(t, deploymentNonMemberRead, http.StatusNotFound, "job_not_found")

	viewerCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-phase11-jobapi-admin-member-viewer-cancel",
	}, phase2test.WithCookies(deploymentViewerCookies, deploymentViewerCSRF), phase2test.WithHeader(authn.CSRFHeaderName, deploymentViewerCSRF.Value))
	httptestx.RequireErrorEnvelope(t, viewerCancel, http.StatusForbidden, "authorization_denied")

	deploymentViewerAdmin := phase2test.PatchMembership(t, harness.Server, adminLogin, incidentID.String(), deploymentViewerUser.ID.String(), map[string]any{
		"base_membership_version": deploymentViewerMembership["membership_version"],
		"role":                    "admin",
	})
	_ = deploymentViewerAdmin
	viewerAdminCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-phase11-jobapi-admin-member-admin-cancel",
	}, phase2test.WithCookies(deploymentViewerCookies, deploymentViewerCSRF), phase2test.WithHeader(authn.CSRFHeaderName, deploymentViewerCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, viewerAdminCancel, http.StatusOK)

	demotedJob, err := harness.Server.Runtime.Jobs.Create(context.Background(), jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: submitterUser.ID,
		AuthPolicy:        jobs.AuthPolicyDeploymentAdminIncidentMembership,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create admin-member demotion job: %v", err)
	}
	phase2test.PatchMembership(t, harness.Server, adminLogin, incidentID.String(), submitterUser.ID.String(), map[string]any{
		"base_membership_version": submitterMembership["membership_version"],
		"role":                    "viewer",
	})
	if _, err := harness.DB.Exec(`UPDATE users SET is_deployment_admin = false WHERE id = $1`, submitterUser.ID); err != nil {
		t.Fatalf("demote submitter deployment admin flag: %v", err)
	}
	demotedRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+demotedJob.JobID, nil, phase2test.WithCookies(submitterCookies))
	httptestx.RequireErrorEnvelope(t, demotedRead, http.StatusNotFound, "job_not_found")
	demotedCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+demotedJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-phase11-jobapi-admin-member-demoted-cancel",
	}, phase2test.WithCookies(submitterCookies, submitterCSRF), phase2test.WithHeader(authn.CSRFHeaderName, submitterCSRF.Value))
	httptestx.RequireErrorEnvelope(t, demotedCancel, http.StatusNotFound, "job_not_found")
}
