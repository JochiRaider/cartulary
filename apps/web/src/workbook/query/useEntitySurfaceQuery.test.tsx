import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
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
import {
  renderWithWorkbookQueryBrowsing as render,
  renderHookWithWorkbookQueryBrowsing as renderHook,
  workbookQueryMeta,
} from "../../testing/workbookQueryTestSupport";
import { createWorkbookViewQueryAdapter } from "../adapters/createWorkbookViewQueryAdapter";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { useEntitySurfaceQuery } from "./useEntitySurfaceQuery";
import type {
  WorkbookViewQueryPort,
  WorkbookViewQueryResult,
} from "./WorkbookViewQueryPort";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

it("browses only the active Entity sheet and releases its rows independently on departure", async () => {
  const fetch = vi.fn((url: RequestInfo | URL) =>
    Promise.resolve(
      String(url).includes(hostsViewSchemaId)
        ? queryResponse(hostsViewSchemaId, [hostRow(hostCurrentId, 1, "Host")])
        : queryResponse(identitiesViewSchemaId, [
            identityRow(identityCurrentId, 1, "Identity"),
          ]),
    ),
  );
  vi.stubGlobal("fetch", fetch);
  const hook = renderHook(
    ({ activeViewSchemaId }) =>
      useEntitySurfaceQuery({
        activeViewSchemaId,
        hostQueryState: emptyWorkbookQueryState(),
        identityQueryState: emptyWorkbookQueryState(),
        viewQuery,
        onAuthorityUncertain: undefined,
      }),
    { initialProps: { activeViewSchemaId: hostsViewSchemaId as string } },
  );
  await act(() => hook.result.current.refresh());
  expect(fetch).toHaveBeenCalledTimes(1);
  expect(hook.result.current.hostRows).toHaveLength(1);
  expect(hook.result.current.identityRows).toEqual([]);
  hook.rerender({ activeViewSchemaId: identitiesViewSchemaId });
  expect(hook.result.current.hostRows).toEqual([]);
  await act(() => hook.result.current.refresh());
  expect(fetch).toHaveBeenCalledTimes(2);
  expect(hook.result.current.identityRows).toHaveLength(1);
  expect(hook.result.current.hosts.browsing.accepted).toBeNull();
  hook.rerender({ activeViewSchemaId: hostsViewSchemaId });
  expect(hook.result.current.identityRows).toEqual([]);
  await act(() => hook.result.current.refresh());
  expect(fetch).toHaveBeenCalledTimes(3);
  hook.rerender({ activeViewSchemaId: notesViewSchemaId });
  expect(hook.result.current.entityIndex).toEqual({});
  await act(() => hook.result.current.refresh());
  await act(async () => {
    await expect(
      hook.result.current.refresh({ requireAcceptance: true }),
    ).rejects.toThrow("active reader");
  });
  expect(fetch).toHaveBeenCalledTimes(3);
  hook.unmount();
  const empty = async (): Promise<WorkbookViewQueryResult> => ({
    kind: "accepted",
    value: {
      incidentId,
      viewSchemaId: hostsViewSchemaId,
      rows: [],
      canonicalQuery: { filters: [], sort: [] },
      paging: { limit: 100, hasMore: false, nextCursor: null },
      producingRequest: { queryState: emptyWorkbookQueryState(), limit: 100 },
    },
  });
  const query = vi.fn(empty);
  const reader = { query };
  const references = renderHook(
    ({ reader }: { reader: WorkbookViewQueryPort }) =>
      useEntitySurfaceQuery({
        activeViewSchemaId: timelineViewSchemaId,
        hostQueryState: emptyWorkbookQueryState(),
        identityQueryState: emptyWorkbookQueryState(),
        viewQuery: reader,
        onAuthorityUncertain: undefined,
      }),
    { initialProps: { reader } },
  );
  const readReferences = references.result.current.refresh;
  await act(() => readReferences());
  const revalidatedReader = { query: vi.fn(empty) };
  references.rerender({ reader: revalidatedReader });
  expect(references.result.current.refresh).toBe(readReferences);
  expect(revalidatedReader.query).not.toHaveBeenCalled();
  await act(() => readReferences());
  expect(query).toHaveBeenCalledTimes(2);
  expect(revalidatedReader.query).toHaveBeenCalledTimes(2);
  // Mention creation must retain refresh recovery when the bounded reference
  // observation is rejected, even though ordinary reference reads are silent.
  const denied = {
    query: vi.fn(
      async (): Promise<WorkbookViewQueryResult> => ({ kind: "aborted" }),
    ),
  };
  references.rerender({ reader: denied });
  await act(async () => {
    await expect(
      references.result.current.refresh({ requireAcceptance: true }),
    ).rejects.toThrow("not accepted");
  });
  expect(references.result.current.references.hosts).toEqual([]);
  references.rerender({ reader: revalidatedReader });
  // Replacement resumes the unaccepted reference observation once.
  await waitFor(() => expect(revalidatedReader.query).toHaveBeenCalledTimes(4));
  await act(() =>
    references.result.current.refresh({ requireAcceptance: true }),
  );
  expect(revalidatedReader.query).toHaveBeenCalledTimes(6);
  references.unmount();
  const interrupted = deferred<WorkbookViewQueryResult>();
  const pendingReader = { query: vi.fn(() => interrupted.promise) };
  const startup = renderHook(
    ({ reader }: { reader: WorkbookViewQueryPort }) =>
      useEntitySurfaceQuery({
        activeViewSchemaId: timelineViewSchemaId,
        hostQueryState: emptyWorkbookQueryState(),
        identityQueryState: emptyWorkbookQueryState(),
        viewQuery: reader,
        onAuthorityUncertain: undefined,
      }),
    { initialProps: { reader: pendingReader } },
  );
  let pendingRead = Promise.resolve();
  act(() => {
    pendingRead = startup.result.current.refresh();
  });
  const recoveredReader = { query: vi.fn(empty) };
  startup.rerender({ reader: recoveredReader });
  await waitFor(() => expect(recoveredReader.query).toHaveBeenCalledTimes(2));
  await act(async () => {
    interrupted.resolve(await empty());
    await pendingRead;
  });
  expect(startup.result.current.references.hosts).toEqual([]);
  expect(recoveredReader.query).toHaveBeenCalledTimes(2);
  startup.unmount();
  vi.unstubAllGlobals();
});
const incidentId = "00000000-0000-4000-8000-000000000001";
const hostCurrentId = "00000000-0000-4000-8000-000000000201";
const hostObsoleteId = "00000000-0000-4000-8000-000000000202";
const identityCurrentId = "00000000-0000-4000-8000-000000000203";
const identityObsoleteId = "00000000-0000-4000-8000-000000000204";
const viewQuery = createWorkbookViewQueryAdapter({
  apiBase: undefined,
  incidentId,
});

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

function withoutLocalViewSchema(row: unknown): unknown {
  const { view_schema_id: _viewSchemaId, ...wireRow } = row as Record<
    string,
    unknown
  >;
  return wireRow;
}

function queryResponse(viewSchemaId: string, rows: readonly unknown[]) {
  return jsonResponse({
    data: {
      incident_id: incidentId,
      view_schema_id: viewSchemaId,
      rows: rows.map(withoutLocalViewSchema),
    },
    meta: workbookQueryMeta(viewSchemaId),
  });
}

function EntityQueryHarness({
  activeViewSchemaId,
  onAuthorityUncertain,
}: {
  readonly activeViewSchemaId: string;
  readonly onAuthorityUncertain?: (() => void) | undefined;
}) {
  const query = useEntitySurfaceQuery({
    activeViewSchemaId,
    hostQueryState: emptyWorkbookQueryState(),
    identityQueryState: emptyWorkbookQueryState(),
    onAuthorityUncertain,
    viewQuery,
  });
  const isHost = activeViewSchemaId === hostsViewSchemaId;
  const recordId = isHost ? hostCurrentId : identityCurrentId;
  const fieldKey = isHost ? "host.display_name" : "identity.display_name";
  const rowVersion = isHost ? 2 : 4;
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
              record_id: recordId,
              row_version: rowVersion,
              change_set_id: "change-1",
              client_txn_id: "txn-1",
              actor_user_id: "user-2",
              changed_field_keys: [fieldKey],
              affected_views: [
                {
                  view_schema_id: activeViewSchemaId,
                  change_kind: "patch",
                  patch_cells: {
                    record_id: recordId,
                    row_version: rowVersion,
                    cells: {
                      [fieldKey]: {
                        value: isHost ? "Patched host" : "Patched identity",
                      },
                    },
                  },
                },
              ],
            },
            activeViewSchemaId,
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

  it("requires an accepted current query before authorization recovery can resume", async () => {
    for (const activeViewSchemaId of [
      hostsViewSchemaId,
      identitiesViewSchemaId,
    ]) {
      const onAuthorityUncertain = vi.fn();
      const query = vi.fn().mockResolvedValue({
        kind: "rejected",
        failure: { kind: "invalid_contract", message: "Malformed query" },
      });
      const view = renderHook(() =>
        useEntitySurfaceQuery({
          activeViewSchemaId,
          hostQueryState: emptyWorkbookQueryState(),
          identityQueryState: emptyWorkbookQueryState(),
          onAuthorityUncertain,
          viewQuery: { query },
        }),
      );
      await act(async () => {
        await expect(
          view.result.current.refresh({ requireAcceptance: true }),
        ).rejects.toMatchObject({
          recovery: { kind: "unavailable", failure: "contract" },
        });
      });
      expect(query).toHaveBeenCalledTimes(1);
      expect(onAuthorityUncertain).not.toHaveBeenCalled();
      query.mockResolvedValue({ kind: "aborted" });
      await act(async () => {
        await expect(
          view.result.current.refresh({ requireAcceptance: true }),
        ).rejects.toMatchObject({ recovery: { kind: "cancelled" } });
      });
      view.unmount();
    }
  });

  it("owns explicit host and identity admission, indexing, live patching, and cleanup", async () => {
    const fetch = vi.fn((input: RequestInfo | URL) =>
      Promise.resolve(
        String(input).includes(`/views/${hostsViewSchemaId}/query`)
          ? queryResponse(hostsViewSchemaId, [
              hostRow(hostCurrentId, 1, "Current host"),
            ])
          : queryResponse(identitiesViewSchemaId, [
              identityRow(identityCurrentId, 3, "Current identity"),
            ]),
      ),
    );
    vi.stubGlobal("fetch", fetch);
    for (const activeViewSchemaId of [
      hostsViewSchemaId,
      identitiesViewSchemaId,
    ]) {
      const isHost = activeViewSchemaId === hostsViewSchemaId;
      const label = isHost ? "host" : "identity";
      const recordId = isHost ? hostCurrentId : identityCurrentId;
      const view = render(
        <EntityQueryHarness activeViewSchemaId={activeViewSchemaId} />,
      );
      fireEvent.click(screen.getByRole("button", { name: "refresh" }));
      await waitFor(() =>
        expect(screen.getByLabelText("entity-load-state").textContent).toBe(
          "ready",
        ),
      );
      expect(screen.getByLabelText(`${label}-rows`).textContent).toBe(
        `${recordId}:Current ${label}`,
      );
      expect(
        screen.getByLabelText(isHost ? "identity-rows" : "host-rows")
          .textContent,
      ).toBe("");
      expect(screen.getByLabelText("entity-index").textContent).toBe(recordId);
      fireEvent.click(screen.getByRole("button", { name: "patch" }));
      await waitFor(() =>
        expect(screen.getByLabelText(`${label}-rows`).textContent).toBe(
          `${recordId}:Patched ${label}`,
        ),
      );
      fireEvent.click(screen.getByRole("button", { name: "clear" }));
      expect(screen.getByLabelText("host-rows").textContent).toBe("");
      expect(screen.getByLabelText("identity-rows").textContent).toBe("");
      expect(screen.getByLabelText("entity-index").textContent).toBe("");
      view.unmount();
    }
    expect(fetch).toHaveBeenCalledTimes(2);
  });

  it("rejects obsolete Entity results after rapid refresh and schema departure", async () => {
    for (const activeViewSchemaId of [
      hostsViewSchemaId,
      identitiesViewSchemaId,
    ]) {
      const isHost = activeViewSchemaId === hostsViewSchemaId;
      const recordId = isHost ? hostCurrentId : identityCurrentId;
      const opposite = isHost ? identitiesViewSchemaId : hostsViewSchemaId;
      const first = deferred<Response>(),
        departed = deferred<Response>();
      const signals: AbortSignal[] = [];
      let calls = 0;
      const fetch = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
        calls++;
        if (init?.signal) signals.push(init.signal);
        if (calls === 1) return first.promise;
        if (calls === 3) return departed.promise;
        return Promise.resolve(
          String(input).includes(`/views/${hostsViewSchemaId}/query`)
            ? queryResponse(hostsViewSchemaId, [
                hostRow(hostCurrentId, 2, "Current host"),
              ])
            : queryResponse(identitiesViewSchemaId, [
                identityRow(identityCurrentId, 2, "Current identity"),
              ]),
        );
      });
      vi.stubGlobal("fetch", fetch);
      const view = render(
        <EntityQueryHarness activeViewSchemaId={activeViewSchemaId} />,
      );
      fireEvent.click(screen.getByRole("button", { name: "refresh" }));
      await waitFor(() => expect(calls).toBe(1));
      fireEvent.click(screen.getByRole("button", { name: "refresh" }));
      await waitFor(() =>
        expect(screen.getByLabelText("entity-index").textContent).toBe(
          recordId,
        ),
      );
      expect(signals[0]?.aborted).toBe(true);
      await act(async () => {
        first.resolve(
          queryResponse(activeViewSchemaId, [
            isHost
              ? hostRow(hostObsoleteId, 1, "Obsolete host")
              : identityRow(identityObsoleteId, 1, "Obsolete identity"),
          ]),
        );
      });
      expect(screen.getByLabelText("entity-index").textContent).toBe(recordId);
      fireEvent.click(screen.getByRole("button", { name: "refresh" }));
      await waitFor(() => expect(calls).toBe(3));
      view.rerender(<EntityQueryHarness activeViewSchemaId={opposite} />);
      expect(signals[2]?.aborted).toBe(true);
      expect(screen.getByLabelText("entity-index").textContent).toBe("");
      fireEvent.click(screen.getByRole("button", { name: "refresh" }));
      await waitFor(() =>
        expect(screen.getByLabelText("entity-load-state").textContent).toBe(
          "ready",
        ),
      );
      await act(async () => {
        departed.resolve(errorResponse("authorization_denied", 403));
      });
      expect(screen.getByLabelText("entity-index").textContent).toBe(
        isHost ? identityCurrentId : hostCurrentId,
      );
      expect(screen.getByLabelText("entity-load-state").textContent).toBe(
        "ready",
      );
      view.unmount();
    }
  });

  it("clears protected rows on access loss and aborts active queries on teardown", async () => {
    for (const activeViewSchemaId of [
      hostsViewSchemaId,
      identitiesViewSchemaId,
    ]) {
      const onAuthorityUncertain = vi.fn();
      const pending = deferred<Response>();
      const signals: AbortSignal[] = [];
      let denied = false,
        held = false;
      vi.stubGlobal(
        "fetch",
        vi.fn((_input: RequestInfo | URL, init?: RequestInit) => {
          if (denied)
            return Promise.resolve(errorResponse("authorization_denied", 403));
          if (init?.signal) signals.push(init.signal);
          if (held) return pending.promise;
          return Promise.resolve(
            queryResponse(activeViewSchemaId, [
              activeViewSchemaId === hostsViewSchemaId
                ? hostRow(hostCurrentId, 1, "Current host")
                : identityRow(identityCurrentId, 1, "Current identity"),
            ]),
          );
        }),
      );
      const view = render(
        <EntityQueryHarness
          activeViewSchemaId={activeViewSchemaId}
          onAuthorityUncertain={onAuthorityUncertain}
        />,
      );
      fireEvent.click(screen.getByRole("button", { name: "refresh" }));
      await waitFor(() =>
        expect(screen.getByLabelText("entity-load-state").textContent).toBe(
          "ready",
        ),
      );
      denied = true;
      fireEvent.click(screen.getByRole("button", { name: "refresh" }));
      await waitFor(() =>
        expect(screen.getByLabelText("entity-load-state").textContent).toBe(
          "permission_denied",
        ),
      );
      expect(onAuthorityUncertain).toHaveBeenCalledTimes(1);
      expect(screen.getByLabelText("entity-index").textContent).toBe("");
      denied = false;
      held = true;
      fireEvent.click(screen.getByRole("button", { name: "refresh" }));
      await waitFor(() => expect(signals).toHaveLength(2));
      view.unmount();
      expect(signals[1]?.aborted).toBe(true);
      await act(async () => {
        pending.resolve(queryResponse(activeViewSchemaId, []));
      });
    }
  });

  it("conceals protected rows before rejecting an acceptance-required merge refresh", async () => {
    for (const activeViewSchemaId of [
      hostsViewSchemaId,
      identitiesViewSchemaId,
    ]) {
      const onAuthorityUncertain = vi.fn();
      const fetch = vi.fn(() =>
        Promise.resolve(
          queryResponse(activeViewSchemaId, [
            activeViewSchemaId === hostsViewSchemaId
              ? hostRow(hostCurrentId, 1, "Current host")
              : identityRow(identityCurrentId, 1, "Current identity"),
          ]),
        ),
      );
      vi.stubGlobal("fetch", fetch);
      const query = renderHook(() =>
        useEntitySurfaceQuery({
          activeViewSchemaId,
          hostQueryState: emptyWorkbookQueryState(),
          identityQueryState: emptyWorkbookQueryState(),
          onAuthorityUncertain,
          viewQuery,
        }),
      );
      await act(() =>
        query.result.current.refresh({ requireAcceptance: true }),
      );
      expect(
        activeViewSchemaId === hostsViewSchemaId
          ? query.result.current.hostRows
          : query.result.current.identityRows,
      ).toHaveLength(1);
      fetch.mockImplementation(() =>
        Promise.resolve(errorResponse("authorization_denied", 403)),
      );
      await act(async () => {
        await expect(
          query.result.current.refresh({ requireAcceptance: true }),
        ).rejects.toThrow();
      });
      expect(query.result.current.hostRows).toEqual([]);
      expect(query.result.current.identityRows).toEqual([]);
      expect(query.result.current.loadState.kind).toBe("permission_denied");
      expect(onAuthorityUncertain).toHaveBeenCalledTimes(1);
      query.unmount();
    }
  });
});
