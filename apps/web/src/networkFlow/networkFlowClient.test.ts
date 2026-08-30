import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ExtensionAvailabilityUnavailableError } from "../extensions/extensionAvailability";
import { csrfHeaderName } from "../services/browserApi";
import { NetworkFlowContractDecodeError } from "../services/networkFlowContractAdapter";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import {
  listNetworkFlowTables,
  renameNetworkFlowTable,
} from "./networkFlowClient";

const incidentId = "incident-1";
const tableId = "nft_11111111111111111111111111111111";

describe("Network Flow client transport boundary", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=network-flow-csrf",
    );
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("uses the owner-neutral browser transport for decoding, cancellation, and CSRF", async () => {
    fetchMock
      .mockResolvedValueOnce(
        jsonResponse({
          schema_id: "cartulary.network_flow.table_list.v1",
          tables: [tableResource()],
          meta: { count: 1 },
        }),
      )
      .mockResolvedValueOnce(
        jsonResponse({
          schema_id: "cartulary.network_flow.table_mutation_result.v1",
          table: { ...tableResource(), display_name: "renamed.csv" },
        }),
      );
    const availability = readyExtensionAvailability(incidentId);
    const abortController = new AbortController();

    const tables = await listNetworkFlowTables({
      availability,
      incidentId,
      signal: abortController.signal,
    });
    const renamed = await renameNetworkFlowTable({
      availability,
      baseTableVersion: 1,
      displayName: "renamed.csv",
      incidentId,
      tableId,
    });

    expect(tables[0]?.network_flow_table_id).toBe(tableId);
    expect(renamed.display_name).toBe("renamed.csv");
    const listInit = fetchMock.mock.calls[0]?.[1] as RequestInit | undefined;
    expect(listInit?.signal).toBe(abortController.signal);
    expect(listInit?.credentials).toBe("include");
    expect(new Headers(listInit?.headers).get("Content-Type")).toBeNull();
    const renameInit = fetchMock.mock.calls[1]?.[1] as RequestInit | undefined;
    expect(renameInit?.method).toBe("PATCH");
    expect(new Headers(renameInit?.headers).get(csrfHeaderName)).toBe(
      "network-flow-csrf",
    );
    expect(JSON.parse(String(renameInit?.body))).toMatchObject({
      base_table_version: 1,
      display_name: "renamed.csv",
    });
  });

  it("fails before transport when the discovered Network Flow route is no longer claimed", async () => {
    const availability = readyExtensionAvailability(incidentId);
    availability.setDiscovery([
      {
        profile_id: "network_flow_activity",
        claimed: false,
        contract_major: 5,
        route_families: ["/api/v1/incidents/{incident_id}/network-flow"],
        workspace_keys: ["network_analysis"],
        capabilities: [],
      },
    ]);
    const tag = availability.reserve();
    if (
      tag === null ||
      !availability.acceptWorkbookStartup(tag, {
        schema_id: "cartulary.extension_workspace_availability.v1",
        incident_id: incidentId,
        workspaces: [
          {
            extension_profile_id: "network_flow_activity",
            workspace_key: "network_analysis",
          },
        ],
      })
    ) {
      throw new Error("failed to arrange unclaimed Network Flow route");
    }

    await expect(
      listNetworkFlowTables({ availability, incidentId }),
    ).rejects.toBeInstanceOf(ExtensionAvailabilityUnavailableError);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("keeps malformed success responses behind the generated decoder", async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({
        schema_id: "cartulary.network_flow.table_list.v1",
        tables: [{ network_flow_table_id: tableId }],
        meta: { count: 1 },
      }),
    );

    await expect(
      listNetworkFlowTables({
        availability: readyExtensionAvailability(incidentId),
        incidentId,
      }),
    ).rejects.toBeInstanceOf(NetworkFlowContractDecodeError);
  });
});

function tableResource() {
  const digest = "a".repeat(64);
  return {
    network_flow_table_id: tableId,
    incident_id: "11111111-1111-4111-8111-111111111111",
    display_name: "flows.csv",
    table_version: 1,
    table_status: "active",
    source_import_session_id: "import-session-1",
    source_import_unit_id: "import-unit-1",
    source_content_sha256: digest,
    source_filename_display: "flows.csv",
    source_filename_digest: digest,
    source_filename_digest_key_id: "filename-key-1",
    mapping_fingerprint: digest,
    source_profile_id: "cisco_sna_netflow_csv_v1",
    parser_profile_id: "rfc4180_headered_csv_v1",
    row_count_accepted: 1,
    row_count_rejected: 0,
    diagnostics_truncated: false,
    created_by_user_id: "user-1",
    created_at: "2026-07-10T12:00:00Z",
    updated_at: "2026-07-10T12:00:00Z",
    deleted_at: null,
  };
}

function jsonResponse(payload: unknown, status = 200) {
  const envelope =
    status >= 200 &&
    status < 300 &&
    payload !== null &&
    typeof payload === "object" &&
    !Array.isArray(payload) &&
    "schema_id" in payload
      ? { data: payload, meta: { request_id: "request-network-flow" } }
      : payload;
  return new Response(JSON.stringify(envelope), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
