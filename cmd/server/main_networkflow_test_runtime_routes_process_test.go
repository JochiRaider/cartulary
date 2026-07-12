package main

import (
	"net/http"
	"testing"

	networkflowharnesscontrol "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestNetworkFlowHarnessRuntimeRouteServerProcessContribution(t *testing.T) {
	addr := reserveHarnessRuntimeProcessAddress(t)
	apiOrigin := "http://" + addr
	publicOrigin := "http://127.0.0.1:4173"
	server := startHarnessRuntimeServerProcess(t, "network-flow-test-runtime", map[string]string{
		"CARTULARY_HTTP_ADDR":             addr,
		"CARTULARY_ENABLE_TEST_ROUTES":    "1",
		"CARTULARY_TEST_RUNTIME_MARKER":   "harness-owned",
		"CARTULARY_TEST_ROUTE_TOKEN":      httptestx.TestRouteToken,
		"CARTULARY_WEB_E2E_API_ORIGIN":    apiOrigin,
		"CARTULARY_WEB_E2E_PUBLIC_ORIGIN": publicOrigin,
	})

	invalidFault := doHarnessRuntimeJSON(t, server, http.MethodPost, "/api/v1/test/runtime/network-flow-faults", map[string]any{
		"boundary":     "network_flow.invalid",
		"fault_kind":   networkflowharnesscontrol.NetworkFlowFaultKindReturnError,
		"error_code":   "server_process_probe",
		"consume_once": true,
	}, httptestx.TestRouteToken, publicOrigin, "")
	httptestx.RequireErrorEnvelope(t, invalidFault, http.StatusBadRequest, "invalid_network_flow_fault_request")

	armFault := doHarnessRuntimeJSON(t, server, http.MethodPost, "/api/v1/test/runtime/network-flow-faults", networkFlowFaultProcessBody(), httptestx.TestRouteToken, publicOrigin, "")
	httptestx.RequireSuccessEnvelope(t, armFault, http.StatusCreated)

	reset := doHarnessRuntimeJSON(t, server, http.MethodPost, "/api/v1/test/runtime/reset", nil, httptestx.TestRouteToken, publicOrigin, "")
	httptestx.RequireSuccessEnvelope(t, reset, http.StatusOK)

	rearmFault := doHarnessRuntimeJSON(t, server, http.MethodPost, "/api/v1/test/runtime/network-flow-faults", networkFlowFaultProcessBody(), httptestx.TestRouteToken, publicOrigin, "")
	httptestx.RequireSuccessEnvelope(t, rearmFault, http.StatusCreated)
}

func networkFlowFaultProcessBody() map[string]any {
	return map[string]any{
		"boundary":     networkflowharnesscontrol.NetworkFlowFaultBoundaryImportBeforeTransactionCommit,
		"fault_kind":   networkflowharnesscontrol.NetworkFlowFaultKindReturnError,
		"error_code":   "server_process_probe",
		"consume_once": true,
	}
}
