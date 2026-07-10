import { describe, expect, it } from "vitest";
import {
  networkAnalysisURLSelected,
  writeNetworkAnalysisURL,
} from "./networkFlowClient";

describe("networkFlowClient route identity", () => {
  it("uses extension workspace route state for Network Analysis", () => {
    const params = new URLSearchParams({
      sheet_ref_kind: "extension_workspace",
      extension_profile_id: "network_flow_activity",
      sheet_ref_id: "network_analysis",
    });
    expect(networkAnalysisURLSelected(params)).toBe(true);

    window.history.replaceState(
      {},
      "",
      "/?incident_id=old&view_schema_id=cartulary.view.timeline.v2&sheet_ref_id=legacy",
    );
    writeNetworkAnalysisURL("incident-1");

    const next = new URLSearchParams(window.location.search);
    expect(next.get("incident_id")).toBe("incident-1");
    expect(next.get("sheet_ref_kind")).toBe("extension_workspace");
    expect(next.get("extension_profile_id")).toBe("network_flow_activity");
    expect(next.get("sheet_ref_id")).toBe("network_analysis");
    expect(next.has("view_schema_id")).toBe(false);
    expect(next.has("workspace_key")).toBe(false);
  });
});
