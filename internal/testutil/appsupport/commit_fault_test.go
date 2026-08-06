package appsupport

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentCreateCommitFaultRequiresValidatedTestRuntimeAdmission(t *testing.T) {
	validEnv := map[string]string{
		httpapi.TestRoutesEnabledEnv: "1",
		httpapi.TestRuntimeMarkerEnv: httpapi.TestRuntimeMarkerValue,
		httpapi.TestRouteTokenEnv:    httptestx.TestRouteToken,
	}
	if capability, err := newIncidentCreateCommitFaultCapability(
		validEnv,
		httptestx.TestRouteModeDisabled,
		nil,
	); err == nil || capability != nil {
		t.Fatalf("disabled runtime installed fault: capability=%v err=%v", capability, err)
	}
	if capability, err := newIncidentCreateCommitFaultCapability(
		nil,
		httptestx.TestRouteModeCustomEnv,
		nil,
	); err == nil || capability != nil {
		t.Fatalf("unvalidated custom runtime installed fault: capability=%v err=%v", capability, err)
	}
	capability, err := newIncidentCreateCommitFaultCapability(
		nil,
		httptestx.TestRouteModeHarnessOwned,
		nil,
	)
	if err != nil {
		t.Fatalf("harness-owned runtime admission failed: %v", err)
	}
	if capability == nil {
		t.Fatal("harness-owned runtime did not install typed fault")
	}
}
