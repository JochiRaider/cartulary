import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { acceptedQueryMetadata } from "../../testing/workbookQueryTestSupport";
import { ordinaryCreateContributions } from "../features/ordinary/ordinaryCreateContributions";
import { WorkbookOrdinaryCreateOwner } from "../features/ordinary/WorkbookOrdinaryCreateOwner";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { useEntitySurfaceQuery } from "./useEntitySurfaceQuery";
import { useGenericSurfaceQuery } from "./useGenericSurfaceQuery";
import type {
  WorkbookViewQueryPort,
  WorkbookViewQueryResult,
} from "./WorkbookViewQueryPort";

const incident = "10000000-0000-4000-8000-000000000001",
  record = "20000000-0000-4000-8000-000000000001";
const authority = {
  actorId: "99999999-9999-4999-8999-999999999999",
  incidentId: incident,
  sessionIdentity: "same-account",
  role: "admin" as const,
  closed: false,
};
afterEach(cleanup);
function fixture() {
  const owner = new WorkbookOrdinaryCreateOwner(
    incident,
    ordinaryCreateContributions,
  );
  owner.setAuthority(authority);
  const queryState = emptyWorkbookQueryState();
  return { owner, queryState };
}
describe("ordinary creation query reconciliation", () => {
  it("keeps all generic schema queries above accepted versions without inserting filtered records", async () => {
    for (const view of ordinaryCreateContributions
      .flatMap((item) => item.views)
      .filter(
        (view) =>
          !["cartulary.view.hosts.v1", "cartulary.view.identities.v1"].includes(
            view,
          ),
      )) {
      const { owner, queryState } = fixture(),
        contract = requireViewContract(view);
      let rows = [fullWorkbookViewRow(contract, record, 1, {})];
      const query = vi.fn<WorkbookViewQueryPort["query"]>(async () => ({
        kind: "accepted",
        value: {
          incidentId: incident,
          viewSchemaId: view,
          ...acceptedQueryMetadata(view),
          rows,
        },
      }));
      const hook = renderHook(() =>
        useGenericSurfaceQuery({
          active: true,
          contract,
          ordinaryCreateOwner: owner,
          queryState,
          viewSchemaId: view,
          viewQuery: { query },
          onAuthorityUncertain: undefined,
        }),
      );
      await act(() => hook.result.current.refresh());
      const accepted = fullWorkbookViewRow(contract, record, 4, {});
      act(() => {
        owner.acceptRow(accepted);
        owner.setAuthority({ ...authority, closed: true });
      });
      expect(hook.result.current.rows[0]?.row_version, view).toBe(4);
      await act(async () => {
        await expect(
          hook.result.current.refresh({ requireAcceptance: true }),
        ).rejects.toThrow();
      });
      expect(hook.result.current.rows[0]?.row_version).toBe(4);
      expect(hook.result.current.loadState.kind).toBe("stale_error");
      rows = [];
      await act(() => hook.result.current.refresh());
      expect(hook.result.current.rows).toEqual([]);
      act(() => {
        owner.acceptRow(fullWorkbookViewRow(contract, record, 5, {}));
        owner.setAuthority(authority);
      });
      expect(hook.result.current.rows).toEqual([]);
      hook.unmount();
    }
  });
  it("fences late generic reads immediately when authorization is concealed", async () => {
    const { owner, queryState } = fixture(),
      contract = requireViewContract("cartulary.view.evidence.v1");
    const response = deferred<WorkbookViewQueryResult>();
    const query = vi.fn(() => response.promise);
    const hook = renderHook(() =>
      useGenericSurfaceQuery({
        active: true,
        contract,
        ordinaryCreateOwner: owner,
        queryState,
        viewSchemaId: contract.viewSchemaId,
        viewQuery: { query },
        onAuthorityUncertain: undefined,
      }),
    );
    let loading: Promise<void> | undefined;
    act(() => {
      loading = hook.result.current.refresh();
      owner.suspend();
    });
    await act(async () => {
      response.resolve({
        kind: "accepted",
        value: {
          incidentId: incident,
          viewSchemaId: contract.viewSchemaId,
          ...acceptedQueryMetadata(contract.viewSchemaId),
          rows: [fullWorkbookViewRow(contract, record, 1, {})],
        },
      });
      await loading;
    });
    expect(hook.result.current.rows).toEqual([]);
    await act(() => hook.result.current.refresh());
    expect(query).toHaveBeenCalledTimes(1);
  });
  it("reconciles both Entity queries and prevents obsolete reads after suspension", async () => {
    const { owner, queryState } = fixture();
    let version = 1;
    const query = vi.fn<WorkbookViewQueryPort["query"]>(
      async ({ contract }) => ({
        kind: "accepted",
        value: {
          incidentId: incident,
          viewSchemaId: contract.viewSchemaId,
          ...acceptedQueryMetadata(contract.viewSchemaId),
          rows: [
            fullWorkbookViewRow(
              contract,
              contract.viewSchemaId.includes("hosts")
                ? record
                : "20000000-0000-4000-8000-000000000002",
              version,
              {},
            ),
          ],
        },
      }),
    );
    const hook = renderHook(() =>
      useEntitySurfaceQuery({
        ordinaryCreateOwner: owner,
        hostQueryState: queryState,
        identityQueryState: queryState,
        viewQuery: { query },
        onAuthorityUncertain: undefined,
      }),
    );
    await act(() => hook.result.current.refresh());
    act(() => {
      owner.acceptRow(
        fullWorkbookViewRow(
          requireViewContract("cartulary.view.hosts.v1"),
          record,
          3,
          {},
        ),
      );
      owner.acceptRow(
        fullWorkbookViewRow(
          requireViewContract("cartulary.view.identities.v1"),
          "20000000-0000-4000-8000-000000000002",
          3,
          {},
        ),
      );
      owner.setAuthority({ ...authority, closed: true });
    });
    expect(hook.result.current.hostRows[0]?.rowVersion).toBe(3);
    expect(hook.result.current.identityRows[0]?.rowVersion).toBe(3);
    await act(async () => {
      await expect(
        hook.result.current.refresh({ requireAcceptance: true }),
      ).rejects.toThrow();
    });
    expect(hook.result.current.hostRows[0]?.rowVersion).toBe(3);
    version = 4;
    await act(() => hook.result.current.refresh());
    expect(hook.result.current.hostRows[0]?.rowVersion).toBe(4);
    const pending = deferred<WorkbookViewQueryResult>();
    query.mockReturnValue(pending.promise);
    let loading: Promise<void> | undefined;
    act(() => {
      loading = hook.result.current.refresh();
      owner.suspend();
    });
    await act(async () => {
      pending.resolve({
        kind: "accepted",
        value: {
          incidentId: incident,
          viewSchemaId: "cartulary.view.hosts.v1",
          ...acceptedQueryMetadata("cartulary.view.hosts.v1"),
          rows: [
            fullWorkbookViewRow(
              requireViewContract("cartulary.view.hosts.v1"),
              record,
              5,
              {},
            ),
          ],
        },
      });
      await loading;
    });
    expect(hook.result.current.hostRows).toEqual([]);
    expect(hook.result.current.identityRows).toEqual([]);
  });
});
