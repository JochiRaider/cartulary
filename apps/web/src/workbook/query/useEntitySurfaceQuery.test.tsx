import { requireViewContract } from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  deferred,
  errorResponse,
  jsonResponse,
} from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { useEntitySurfaceQuery } from "./useEntitySurfaceQuery";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

function hostRow(recordId: string, rowVersion: number, label: string) {
  return fullWorkbookViewRow(hostsContract, recordId, rowVersion, {
    "host.display_name": label,
    "host.hostname": `${recordId}.example.test`,
  });
}

function identityRow(recordId: string, rowVersion: number, label: string) {
  return fullWorkbookViewRow(identitiesContract, recordId, rowVersion, {
    "identity.display_name": label,
    "identity.upn": `${recordId}@example.test`,
  });
}

function queryResponse(viewSchemaId: string, rows: readonly unknown[]) {
  return jsonResponse({
    data: {
      incident_id: "incident-1",
      view_schema_id: viewSchemaId,
      rows,
    },
  });
}

function EntityQueryHarness({
  onIncidentAccessLost,
}: {
  readonly onIncidentAccessLost?: (() => void) | undefined;
}) {
  const query = useEntitySurfaceQuery({
    apiBase: undefined,
    hostQueryState: emptyWorkbookQueryState(),
    identityQueryState: emptyWorkbookQueryState(),
    incidentId: "incident-1",
    onIncidentAccessLost,
  });
  return (
    <>
      <button onClick={() => void query.refresh()} type="button">
        refresh
      </button>
      <button
        onClick={() => query.invalidate({ kind: "incident_access_lost" })}
        type="button"
      >
        clear
      </button>
      <button
        onClick={() =>
          query.applyRecordChanged(
            {
              record_id: "host-current",
              row_version: 2,
              change_set_id: "change-1",
              client_txn_id: "txn-1",
              actor_user_id: "user-2",
              changed_field_keys: ["host.display_name"],
              affected_views: [
                {
                  view_schema_id: hostsViewSchemaId,
                  change_kind: "patch",
                  patch_cells: {
                    record_id: "host-current",
                    row_version: 2,
                    cells: {
                      "host.display_name": { value: "Patched host" },
                    },
                  },
                },
              ],
            },
            hostsViewSchemaId,
          )
        }
        type="button"
      >
        patch
      </button>
      <output aria-label="entity-load-state">{query.loadState.kind}</output>
      <output aria-label="host-rows">
        {query.hostRows.map((row) => `${row.recordId}:${row.label}`).join(",")}
      </output>
      <output aria-label="identity-rows">
        {query.identityRows
          .map((row) => `${row.recordId}:${row.label}`)
          .join(",")}
      </output>
      <output aria-label="entity-index">
        {Object.keys(query.entityIndex).sort().join(",")}
      </output>
    </>
  );
}

describe("useEntitySurfaceQuery", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("owns dual host and identity admission, indexing, live patching, and explicit cleanup", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes(`/views/${hostsViewSchemaId}/query`)) {
          return Promise.resolve(
            queryResponse(hostsViewSchemaId, [
              hostRow("host-current", 1, "Current host"),
            ]),
          );
        }
        return Promise.resolve(
          queryResponse(identitiesViewSchemaId, [
            identityRow("identity-current", 3, "Current identity"),
          ]),
        );
      }),
    );
    render(<EntityQueryHarness />);

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("entity-load-state").textContent).toBe(
        "ready",
      ),
    );
    expect(screen.getByLabelText("host-rows").textContent).toBe(
      "host-current:Current host",
    );
    expect(screen.getByLabelText("identity-rows").textContent).toBe(
      "identity-current:Current identity",
    );
    expect(screen.getByLabelText("entity-index").textContent).toBe(
      "host-current,identity-current",
    );

    fireEvent.click(screen.getByRole("button", { name: "patch" }));
    await waitFor(() =>
      expect(screen.getByLabelText("host-rows").textContent).toBe(
        "host-current:Patched host",
      ),
    );

    fireEvent.click(screen.getByRole("button", { name: "clear" }));
    expect(screen.getByLabelText("host-rows").textContent).toBe("");
    expect(screen.getByLabelText("identity-rows").textContent).toBe("");
    expect(screen.getByLabelText("entity-index").textContent).toBe("");
  });

  it("rejects obsolete dual-query results after a rapid refresh", async () => {
    const firstHost = deferred<Response>();
    const firstIdentity = deferred<Response>();
    const firstSignals: AbortSignal[] = [];
    let callCount = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
        callCount += 1;
        if (callCount <= 2 && init?.signal) {
          firstSignals.push(init.signal);
        }
        const viewSchemaId = String(input).includes(
          `/views/${hostsViewSchemaId}/query`,
        )
          ? hostsViewSchemaId
          : identitiesViewSchemaId;
        if (callCount === 1) return firstHost.promise;
        if (callCount === 2) return firstIdentity.promise;
        return Promise.resolve(
          viewSchemaId === hostsViewSchemaId
            ? queryResponse(viewSchemaId, [
                hostRow("host-current", 2, "Current host"),
              ])
            : queryResponse(viewSchemaId, [
                identityRow("identity-current", 2, "Current identity"),
              ]),
        );
      }),
    );
    render(<EntityQueryHarness />);

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() => expect(callCount).toBe(2));
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("entity-index").textContent).toBe(
        "host-current,identity-current",
      ),
    );
    expect(firstSignals).toHaveLength(2);
    expect(firstSignals.every((signal) => signal.aborted)).toBe(true);

    firstHost.resolve(
      queryResponse(hostsViewSchemaId, [
        hostRow("host-obsolete", 1, "Obsolete host"),
      ]),
    );
    firstIdentity.resolve(
      queryResponse(identitiesViewSchemaId, [
        identityRow("identity-obsolete", 1, "Obsolete identity"),
      ]),
    );
    await Promise.resolve();
    await Promise.resolve();
    expect(screen.getByLabelText("entity-index").textContent).toBe(
      "host-current,identity-current",
    );
  });

  it("clears protected rows on access loss and aborts both queries on teardown", async () => {
    const onIncidentAccessLost = vi.fn();
    const pendingSignals: AbortSignal[] = [];
    let accessDenied = false;
    vi.stubGlobal(
      "fetch",
      vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
        const viewSchemaId = String(input).includes(
          `/views/${hostsViewSchemaId}/query`,
        )
          ? hostsViewSchemaId
          : identitiesViewSchemaId;
        if (accessDenied) {
          return Promise.resolve(errorResponse("authorization_denied", 403));
        }
        if (init?.signal) pendingSignals.push(init.signal);
        return Promise.resolve(
          viewSchemaId === hostsViewSchemaId
            ? queryResponse(viewSchemaId, [
                hostRow("host-current", 1, "Current host"),
              ])
            : queryResponse(viewSchemaId, [
                identityRow("identity-current", 1, "Current identity"),
              ]),
        );
      }),
    );
    const rendered = render(
      <EntityQueryHarness onIncidentAccessLost={onIncidentAccessLost} />,
    );

    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("entity-index").textContent).toBe(
        "host-current,identity-current",
      ),
    );
    accessDenied = true;
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() =>
      expect(screen.getByLabelText("entity-load-state").textContent).toBe(
        "permission_denied",
      ),
    );
    expect(onIncidentAccessLost).toHaveBeenCalledOnce();
    expect(screen.getByLabelText("entity-index").textContent).toBe("");

    accessDenied = false;
    fireEvent.click(screen.getByRole("button", { name: "refresh" }));
    await waitFor(() => expect(pendingSignals).toHaveLength(4));
    const teardownSignals = pendingSignals.slice(-2);
    rendered.unmount();
    expect(teardownSignals.every((signal) => signal.aborted)).toBe(true);
  });
});
