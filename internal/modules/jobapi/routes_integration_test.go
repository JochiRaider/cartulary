package jobapi_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentJobAuthorizationReDerivedAtRequestTime_Unit(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-jobapi-incident-auth")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
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
	viewerCookies, viewerCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, viewerUser.Email, viewerPassword, nil)
	peerCookies, peerCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, peerUser.Email, peerPassword, nil)

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

	job, err := harness.Server.Runtime.Jobs.Create(context.Background(), jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID},
		SubmittedByUserID: viewerUser.ID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create incident job: %v", err)
	}

	viewerRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(viewerCookies))
	httptestx.RequireSuccessEnvelope(t, viewerRead, http.StatusOK)

	peerCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-peer-cancel",
	}, httptestx.WithCookies(peerCookies, peerCSRF), httptestx.WithHeader(authn.CSRFHeaderName, peerCSRF.Value))
	httptestx.RequireErrorEnvelope(t, peerCancel, http.StatusForbidden, "authorization_denied")

	scenariotest.DeleteMembership(t, harness.Server, adminLogin, incidentID.String(), viewerUser.ID.String(), map[string]any{
		"base_membership_version": viewerMembership["membership_version"],
	})

	afterDeleteRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(viewerCookies))
	httptestx.RequireErrorEnvelope(t, afterDeleteRead, http.StatusNotFound, "job_not_found")

	afterDeleteCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-viewer-cancel-after-delete",
	}, httptestx.WithCookies(viewerCookies, viewerCSRF), httptestx.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value))
	httptestx.RequireErrorEnvelope(t, afterDeleteCancel, http.StatusNotFound, "job_not_found")
}

func TestDeploymentJobAuthorizationReDerivedAtRequestTime_Unit(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-jobapi-deployment-auth")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	submitterPassword := "SubmitterPassphrase11!"
	otherPassword := "OtherPassphrase11!"
	submitterUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-submitter@example.test", "ExtensionProfile Submitter", submitterPassword, false, false, true)
	otherUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-job-other@example.test", "ExtensionProfile Other", otherPassword, false, false, true)
	submitterCookies, submitterCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, submitterUser.Email, submitterPassword, nil)
	otherCookies, otherCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, otherUser.Email, otherPassword, nil)

	job, err := harness.Server.Runtime.Jobs.Create(context.Background(), jobs.CreateParams{
		Scope:             jobs.Scope{Kind: jobs.ScopeKindDeployment},
		SubmittedByUserID: submitterUser.ID,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0},
	})
	if err != nil {
		t.Fatalf("create deployment job: %v", err)
	}

	adminRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireSuccessEnvelope(t, adminRead, http.StatusOK)

	otherRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID, nil, httptestx.WithCookies(otherCookies))
	httptestx.RequireErrorEnvelope(t, otherRead, http.StatusNotFound, "job_not_found")

	otherCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-other-cancel",
	}, httptestx.WithCookies(otherCookies, otherCSRF), httptestx.WithHeader(authn.CSRFHeaderName, otherCSRF.Value))
	httptestx.RequireErrorEnvelope(t, otherCancel, http.StatusNotFound, "job_not_found")

	submitterCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+job.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-submitter-cancel",
	}, httptestx.WithCookies(submitterCookies, submitterCSRF), httptestx.WithHeader(authn.CSRFHeaderName, submitterCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, submitterCancel, http.StatusOK)
}

func TestDeploymentAdminIncidentMembershipPolicy_Unit(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "extension_profile-jobapi-incident-admin-member-auth")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
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
	submitterCookies, submitterCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, submitterUser.Email, submitterPassword, nil)
	deploymentViewerCookies, deploymentViewerCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, deploymentViewerUser.Email, deploymentViewerPassword, nil)
	incidentAdminCookies, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, incidentAdminUser.Email, incidentAdminPassword, nil)
	deploymentNonMemberCookies, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, deploymentNonMemberUser.Email, deploymentNonMemberPassword, nil)

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

	submitterRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, httptestx.WithCookies(submitterCookies))
	httptestx.RequireSuccessEnvelope(t, submitterRead, http.StatusOK)

	deploymentViewerRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, httptestx.WithCookies(deploymentViewerCookies))
	httptestx.RequireSuccessEnvelope(t, deploymentViewerRead, http.StatusOK)

	incidentAdminRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, httptestx.WithCookies(incidentAdminCookies))
	httptestx.RequireErrorEnvelope(t, incidentAdminRead, http.StatusNotFound, "job_not_found")

	deploymentNonMemberRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID, nil, httptestx.WithCookies(deploymentNonMemberCookies))
	httptestx.RequireErrorEnvelope(t, deploymentNonMemberRead, http.StatusNotFound, "job_not_found")

	viewerCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-viewer-cancel",
	}, httptestx.WithCookies(deploymentViewerCookies, deploymentViewerCSRF), httptestx.WithHeader(authn.CSRFHeaderName, deploymentViewerCSRF.Value))
	httptestx.RequireErrorEnvelope(t, viewerCancel, http.StatusForbidden, "authorization_denied")

	deploymentViewerAdmin := scenariotest.PatchMembership(t, harness.Server, adminLogin, incidentID.String(), deploymentViewerUser.ID.String(), map[string]any{
		"base_membership_version": deploymentViewerMembership["membership_version"],
		"role":                    "admin",
	})
	_ = deploymentViewerAdmin
	viewerAdminCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+readJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-admin-cancel",
	}, httptestx.WithCookies(deploymentViewerCookies, deploymentViewerCSRF), httptestx.WithHeader(authn.CSRFHeaderName, deploymentViewerCSRF.Value))
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
	scenariotest.PatchMembership(t, harness.Server, adminLogin, incidentID.String(), submitterUser.ID.String(), map[string]any{
		"base_membership_version": submitterMembership["membership_version"],
		"role":                    "viewer",
	})
	if _, err := harness.DB.Exec(`UPDATE users SET is_deployment_admin = false WHERE id = $1`, submitterUser.ID); err != nil {
		t.Fatalf("demote submitter deployment admin flag: %v", err)
	}
	demotedRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+demotedJob.JobID, nil, httptestx.WithCookies(submitterCookies))
	httptestx.RequireErrorEnvelope(t, demotedRead, http.StatusNotFound, "job_not_found")
	demotedCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+demotedJob.JobID+"/cancel", map[string]any{
		"client_txn_id": "txn-extension_profile-jobapi-admin-member-demoted-cancel",
	}, httptestx.WithCookies(submitterCookies, submitterCSRF), httptestx.WithHeader(authn.CSRFHeaderName, submitterCSRF.Value))
	httptestx.RequireErrorEnvelope(t, demotedCancel, http.StatusNotFound, "job_not_found")
}
