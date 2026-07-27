package server

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

func TestRuntime_ExtensionPublication_OneServingEpoch(t *testing.T) {
	controller, plan, lifecycle := preparedPublicationController(t)
	if _, installed := controller.Summary(); installed ||
		len(controller.Discovery()) != 0 ||
		len(controller.Claims()) != 0 ||
		len(controller.Routes()) != 0 ||
		len(controller.Workspaces()) != 0 ||
		len(controller.Workers()) != 0 ||
		len(controller.JobKindContracts()) != 0 ||
		len(controller.Contributions()) != 0 ||
		len(controller.ImplementationBindings()) != 0 {
		t.Fatal("prepared plan became visible before commit")
	}
	if err := controller.Commit(); err != nil {
		t.Fatal(err)
	}
	acknowledgeAllPublicationComponents(t, controller)
	summary, installed := controller.Summary()
	if !installed || summary != plan.Summary() {
		t.Fatalf("installed epoch = %#v/%t; want %#v", summary, installed, plan.Summary())
	}
	if len(controller.Routes()) == 0 ||
		len(controller.Workspaces()) == 0 ||
		len(controller.Workers()) == 0 ||
		len(controller.JobKindContracts()) == 0 ||
		len(controller.Contributions()) == 0 ||
		len(controller.ImplementationBindings()) == 0 {
		t.Fatal("committed plan projections are incomplete")
	}
	if lifecycle.AdmissionOpen() {
		t.Fatal("committed epoch opened admission before serving")
	}
	if err := controller.Serve(); err != nil {
		t.Fatal(err)
	}
	if controller.State() != PublicationServing || !lifecycle.AdmissionOpen() {
		t.Fatalf("serving state = %s/%t", controller.State(), lifecycle.AdmissionOpen())
	}
	if err := controller.Serve(); err == nil {
		t.Fatal("second serving path was accepted")
	}
}

func TestRuntime_ExtensionPublication_CurrentAssemblyParity(t *testing.T) {
	coordinator, plan := generatedPublicationPlan(t)
	descriptors := coordinator.Descriptors()
	discovery := plan.Discovery()
	if len(discovery) != len(descriptors) {
		t.Fatalf("discovery profiles = %d, descriptors = %d", len(discovery), len(descriptors))
	}
	routeCount := 0
	for index, profile := range discovery {
		descriptor := descriptors[index]
		if profile.ProfileID != descriptor.ProfileID || !profile.Claimed ||
			profile.Claimable != descriptor.Claimable ||
			!reflect.DeepEqual(profile.RouteFamilies, descriptor.RouteFamilies) ||
			!reflect.DeepEqual(profile.WorkspaceKeys, descriptor.WorkspaceKeys) ||
			!reflect.DeepEqual(profile.Capabilities, descriptor.CapabilityIDs) {
			t.Fatalf("profile parity at %d: profile=%#v descriptor=%#v", index, profile, descriptor)
		}
		routeCount += len(profile.RouteFamilies)
	}
	if len(plan.Routes()) != routeCount {
		t.Fatalf("plan routes = %d, profile routes = %d", len(plan.Routes()), routeCount)
	}
}

func TestRuntime_ExtensionPublication_DequeueGate(t *testing.T) {
	lifecycle := processlifecycle.New()
	runner := jobs.NewRunner()
	runner.ConfigureDequeueGate(lifecycle)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := runner.Close(ctx); err != nil {
			t.Fatalf("close gated runner: %v", err)
		}
	})
	ran := make(chan struct{}, 1)
	if err := runner.Dispatch(func(context.Context) { ran <- struct{}{} }); !errors.Is(err, jobs.ErrDequeueGateClosed) {
		t.Fatalf("pre-serving dispatch error = %v, want dequeue gate closed", err)
	}
	select {
	case <-ran:
		t.Fatal("job work ran before serving")
	default:
	}
	if err := lifecycle.Publish(); err != nil {
		t.Fatal(err)
	}
	if err := runner.Dispatch(func(context.Context) { ran <- struct{}{} }); err != nil {
		t.Fatalf("serving dispatch failed: %v", err)
	}
	select {
	case <-ran:
	case <-time.After(time.Second):
		t.Fatal("serving dispatch did not run")
	}
}

func TestRuntime_ExtensionPublication_MixedClaimProfileDomains(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	canonicalProfiles := []string{
		"enterprise_authentication",
		"import",
		"incident_portability",
		"reference_pack",
		"snapshot_reporting",
	}
	resolution, err := coordinator.ResolveClaims(canonicalProfiles)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	claimed := map[string]bool{}
	for _, profile := range plan.Discovery() {
		if profile.Claimed {
			claimed[profile.ProfileID] = true
		}
	}
	for _, profileID := range canonicalProfiles {
		if !claimed[profileID] {
			t.Errorf("canonical profile %q is absent from the claimed production epoch", profileID)
		}
	}
	if claimed["network_flow_activity"] {
		t.Fatal("mixed-claim production epoch independently claimed Network Flow")
	}

	routeProfiles := map[string]bool{}
	for _, route := range plan.Routes() {
		if route.DispatchState == "claimed" {
			routeProfiles[route.ProfileID] = true
			if route.ContributionID == nil || *route.ContributionID == "" {
				t.Errorf("claimed route has no exact contribution: %#v", route)
			}
		}
	}
	for _, profileID := range canonicalProfiles {
		if !routeProfiles[profileID] {
			t.Errorf("canonical profile %q has no claimed production route contribution", profileID)
		}
	}
	snapshotDescriptor, ok := coordinator.Descriptor("snapshot_reporting")
	if !ok ||
		!containsString(snapshotDescriptor.RouteFamilies, "/api/v1/snapshots") ||
		!containsString(snapshotDescriptor.RouteFamilies, "/api/v1/incidents/{incident_id}/report-compositions") {
		t.Fatalf("Snapshot/Reporting domains are not both represented: %#v/%t", snapshotDescriptor, ok)
	}
}

func TestRuntime_ExtensionPublication_PreparedComponentsAreQuiescent(t *testing.T) {
	controller, _, lifecycle := preparedPublicationController(t)
	if controller.State() != PublicationPrepared || lifecycle.AdmissionOpen() {
		t.Fatalf("prepared components are not quiescent: %s/%t", controller.State(), lifecycle.AdmissionOpen())
	}
	if _, installed := controller.Summary(); installed || len(controller.Routes()) != 0 ||
		len(controller.Workspaces()) != 0 || len(controller.Workers()) != 0 {
		t.Fatal("prepared publication exposed an epoch projection")
	}
}

func TestRuntime_ExtensionPublication_AtomicAdmissionGate(t *testing.T) {
	controller, _, lifecycle := preparedPublicationController(t)
	if err := controller.Commit(); err != nil {
		t.Fatal(err)
	}
	acknowledgeAllPublicationComponents(t, controller)
	if lifecycle.AdmissionOpen() {
		t.Fatal("commit opened admission")
	}
	if err := controller.Serve(); err != nil {
		t.Fatal(err)
	}
	if !lifecycle.AdmissionOpen() || lifecycle.State() != processlifecycle.StateRunning {
		t.Fatalf("atomic gate state = %t/%s", lifecycle.AdmissionOpen(), lifecycle.State())
	}
}

func TestRuntime_ExtensionPublication_AcknowledgmentValidation(t *testing.T) {
	t.Run("early", func(t *testing.T) {
		controller := NewPublicationController(processlifecycle.New())
		if err := controller.Acknowledge("http", strings.Repeat("a", 64), nil); err == nil || controller.State() != PublicationFailed {
			t.Fatalf("early acknowledgment = %v/%s", err, controller.State())
		}
	})
	t.Run("before commit", func(t *testing.T) {
		controller, _, _ := preparedPublicationController(t)
		digest := controller.ExpectedComponents()["http"]
		if err := controller.Acknowledge("http", digest, nil); err == nil || controller.State() != PublicationFailed {
			t.Fatalf("pre-commit acknowledgment = %v/%s", err, controller.State())
		}
	})
	t.Run("unknown", func(t *testing.T) {
		controller, _, _ := preparedPublicationController(t)
		if err := controller.Commit(); err != nil {
			t.Fatal(err)
		}
		if err := controller.Acknowledge("unknown", strings.Repeat("a", 64), nil); err == nil {
			t.Fatal("unknown acknowledgment accepted")
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		controller, _, _ := preparedPublicationController(t)
		if err := controller.Commit(); err != nil {
			t.Fatal(err)
		}
		digest := controller.ExpectedComponents()["http"]
		if err := controller.Acknowledge("http", digest, nil); err != nil {
			t.Fatal(err)
		}
		if err := controller.Acknowledge("http", digest, nil); err == nil {
			t.Fatal("duplicate acknowledgment accepted")
		}
	})
	t.Run("failed", func(t *testing.T) {
		controller, _, _ := preparedPublicationController(t)
		if err := controller.Commit(); err != nil {
			t.Fatal(err)
		}
		digest := controller.ExpectedComponents()["http"]
		if err := controller.Acknowledge("http", digest, errors.New("prepare failed")); err == nil {
			t.Fatal("failed acknowledgment accepted")
		}
	})
	t.Run("digest mismatch", func(t *testing.T) {
		controller, _, _ := preparedPublicationController(t)
		if err := controller.Commit(); err != nil {
			t.Fatal(err)
		}
		if err := controller.Acknowledge("http", strings.Repeat("0", 64), nil); err == nil {
			t.Fatal("mismatched acknowledgment accepted")
		}
	})
	t.Run("missing", func(t *testing.T) {
		controller, _, _ := preparedPublicationController(t)
		if err := controller.Commit(); err != nil {
			t.Fatal(err)
		}
		if err := controller.Serve(); err == nil {
			t.Fatal("missing acknowledgments accepted")
		}
	})
}

func TestRuntime_ExtensionPublication_WorkspaceAndWorkerClaimFiltering(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims([]string{"import", "network_flow_activity"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	claimed := map[string]bool{"import": true, "network_flow_activity": true}
	for _, workspace := range plan.Workspaces() {
		if !claimed[workspace.ProfileID] {
			t.Fatalf("unclaimed workspace published: %#v", workspace)
		}
	}
	for _, worker := range plan.Workers() {
		if !claimed[worker.ProfileID] {
			t.Fatalf("unclaimed worker published: %#v", worker)
		}
	}
	routes := plan.Routes()
	if len(routes) == 0 {
		t.Fatal("reserved route projection is empty")
	}
	for _, route := range routes {
		if route.DispatchState == "claimed" && !claimed[route.ProfileID] {
			t.Fatalf("unclaimed route published: %#v", route)
		}
		if route.ProfileID == incidentbundles.ProfileID && route.DispatchState == "claimed" {
			t.Fatalf("unclaimed Incident Portability route published: %#v", route)
		}
	}
}

func TestRuntime_ExtensionPublication_NoIndependentRederivation(t *testing.T) {
	source, err := os.ReadFile("runtime.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, obsolete := range []string{
		"configuredExtensionClaimIDs",
		"extensionHTTPProfiles",
		"revisionExtensionClaims",
		"switch descriptor.ProfileID",
		"normalizedCfg.EnterpriseAuthentication.Claimed",
	} {
		if strings.Contains(string(source), obsolete) {
			t.Fatalf("obsolete independent derivation %s remains", obsolete)
		}
	}
	for _, required := range []string{
		"extensionassembly.ResolveClaimRequest",
		"extensionassembly.NewPublicationCatalog",
		"publicationHTTPProjections",
		"provider.publication.Discovery()",
		"provider.publication.Claims()",
		"provider.publication.Routes()",
		"provider.publication.Workspaces()",
	} {
		if !strings.Contains(string(source), required) {
			t.Fatalf("runtime consumers are not visibly bound to the immutable publication projection %q", required)
		}
	}
}

func TestRuntime_ExtensionPublication_FailureNoExposure(t *testing.T) {
	controller, _, lifecycle := preparedPublicationController(t)
	if err := controller.Commit(); err != nil {
		t.Fatal(err)
	}
	digest := controller.ExpectedComponents()["http"]
	if err := controller.Acknowledge("http", digest, errors.New("failed")); err == nil {
		t.Fatal("preparation failure was accepted")
	}
	if controller.State() != PublicationFailed || lifecycle.AdmissionOpen() {
		t.Fatalf("failed publication exposure = %s/%t", controller.State(), lifecycle.AdmissionOpen())
	}
	if _, installed := controller.Summary(); installed || len(controller.Discovery()) != 0 {
		t.Fatal("failed publication retained its plan")
	}
}

func TestRuntime_ExtensionPublication_PublishedComponentLoss(t *testing.T) {
	controller, _, lifecycle := preparedPublicationController(t)
	if err := controller.Commit(); err != nil {
		t.Fatal(err)
	}
	acknowledgeAllPublicationComponents(t, controller)
	if err := controller.Serve(); err != nil {
		t.Fatal(err)
	}
	if !controller.ComponentLost("http") {
		t.Fatal("published component loss did not enter the fatal lifecycle")
	}
	signal := <-lifecycle.FatalEvents()
	if signal.ExitCode != 70 || signal.ReasonCode != "published_component_lost" ||
		controller.State() != PublicationFailed || lifecycle.AdmissionOpen() {
		t.Fatalf("component-loss result = %#v/%s/%t", signal, controller.State(), lifecycle.AdmissionOpen())
	}
	if _, installed := controller.Summary(); installed {
		t.Fatal("component-loss path retained a rebuildable plan")
	}
}

func TestRuntime_ExtensionPublication_PreServingComponentLoss(t *testing.T) {
	controller, _, lifecycle := preparedPublicationController(t)
	if err := controller.Commit(); err != nil {
		t.Fatal(err)
	}
	if !controller.ComponentLost("http") {
		t.Fatal("committed component loss was ignored")
	}
	if controller.State() != PublicationFailed || lifecycle.AdmissionOpen() {
		t.Fatalf("pre-serving component loss = %s/%t", controller.State(), lifecycle.AdmissionOpen())
	}
	if _, installed := controller.Summary(); installed {
		t.Fatal("pre-serving component loss retained the installed epoch")
	}
	select {
	case signal := <-lifecycle.FatalEvents():
		t.Fatalf("pre-serving startup failure emitted serving fatal signal %#v", signal)
	default:
	}
}

func TestRuntime_ExtensionPublication_PlanNotPersistedOrLogged(t *testing.T) {
	_, plan := generatedPublicationPlan(t)
	planType := reflect.TypeOf(plan)
	for index := 0; index < planType.NumField(); index++ {
		if planType.Field(index).IsExported() {
			t.Fatalf("publication plan exposes field %s", planType.Field(index).Name)
		}
	}
	for _, path := range []string{"runtime.go", "publication_controller.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"CanonicalJSON", "canonicalJSON", "PersistPublication", "LogPublication"} {
			if strings.Contains(string(source), forbidden) {
				t.Fatalf("%s contains forbidden publication byte path %s", path, forbidden)
			}
		}
	}
}

func TestRuntime_ExtensionPublication_NoSecondPublicationPath(t *testing.T) {
	controller, _, lifecycle := preparedPublicationController(t)
	if err := controller.Commit(); err != nil {
		t.Fatal(err)
	}
	acknowledgeAllPublicationComponents(t, controller)
	if lifecycle.AdmissionOpen() {
		t.Fatal("a path other than PublicationController.Serve opened admission")
	}
	if err := controller.Serve(); err != nil {
		t.Fatal(err)
	}
	if err := lifecycle.Publish(); err == nil {
		t.Fatal("lifecycle accepted a second publication")
	}
}

func generatedPublicationPlan(t testing.TB) (*extensions.Coordinator, extensions.PublicationPlan) {
	t.Helper()
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	descriptors := coordinator.Descriptors()
	profileIDs := make([]string, len(descriptors))
	for index, descriptor := range descriptors {
		profileIDs[index] = descriptor.ProfileID
	}
	resolution, err := coordinator.ResolveClaims(profileIDs)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	return coordinator, plan
}

func preparedPublicationController(t testing.TB) (*PublicationController, extensions.PublicationPlan, *processlifecycle.Controller) {
	t.Helper()
	_, plan := generatedPublicationPlan(t)
	lifecycle := processlifecycle.New()
	controller := NewPublicationController(lifecycle)
	if err := controller.Prepare(plan); err != nil {
		t.Fatal(err)
	}
	return controller, plan, lifecycle
}

func acknowledgeAllPublicationComponents(t testing.TB, controller *PublicationController) {
	t.Helper()
	for componentID, digest := range controller.ExpectedComponents() {
		if err := controller.Acknowledge(componentID, digest, nil); err != nil {
			t.Fatalf("acknowledge %s: %v", componentID, err)
		}
	}
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
