import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import {
  mergeAuthority as authority,
  mergeEntityRow as entityRow,
  mergeIncidentId as incidentId,
  mergeLoserId as loserId,
  mergeReceipt,
  mergeSurvivorId as survivorId,
} from "../../../testing/entityMergeTestSupport";
import { createWorkbookEntityMergeAdapter } from "../../adapters/createWorkbookEntityMergeAdapter";
import type { EntityMergeTransportOutcome } from "./entityMergeOperation";
import { useEntityMergeController } from "./useEntityMergeController";
import { WorkbookEntityMergeOwner } from "./WorkbookEntityMergeOwner";

const otherId = "00000000-0000-4000-8000-000000007102";
type Props = Parameters<typeof useEntityMergeController>[0];
function accepted(type: "host" | "identity"): EntityMergeTransportOutcome {
  return { kind: "acknowledged", receipt: mergeReceipt(type) };
}

function setup(type: "host" | "identity" = "host") {
  const owner = new WorkbookEntityMergeOwner(
    incidentId,
    { create: () => "merge-id" },
    { canReserve: () => true, coordinate: async () => true },
  );
  owner.setAuthority(authority);
  const merge = vi.fn(
    async (): Promise<EntityMergeTransportOutcome> => accepted(type),
  );
  owner.configure({
    ...createWorkbookEntityMergeAdapter({ apiBase: undefined, incidentId }),
    send: merge,
  });
  const survivor = entityRow(survivorId, 7, type);
  const loser = entityRow(loserId, 2, type);
  const discard = vi.fn();
  const drafts = vi.fn(() => false);
  const refresh = vi.fn(async () => undefined);
  const preview = vi.fn(async () => undefined);
  owner.registerProjectionRefresh(refresh);
  let props: Props = {
    canMerge: true,
    discardAffectedDrafts: discard,
    hasAffectedDraft: drafts,
    lifecycleResetKey: "lifecycle",
    loadSurvivorPreview: preview,
    originSurface: type,
    owner,
    rows: [survivor, loser, entityRow(otherId, 1, type)],
    selectedEntity: survivor,
  };
  const hook = renderHook(
    (current: Props) => useEntityMergeController(current),
    { initialProps: props },
  );
  const rerender = (patch: Partial<Props>) => {
    props = { ...props, ...patch };
    hook.rerender(props);
  };
  const review = () => {
    act(() => hook.result.current.commands.selectCandidate(loserId));
    act(() => hook.result.current.commands.review());
    expect(hook.result.current.snapshot.reviewed).not.toBeNull();
  };
  return {
    ...hook,
    rerender,
    review,
    merge,
    owner,
    survivor,
    loser,
    discard,
    drafts,
    refresh,
    preview,
  };
}

describe("useEntityMergeController", () => {
  it("captures the reviewed pair and synchronously admits only one confirmation", async () => {
    for (const type of ["host", "identity"] as const) {
      const t = setup(type);
      t.review();
      const reviewed = t.result.current.snapshot.reviewed;
      expect(reviewed).toMatchObject({
        entityType: type,
        authority,
        originSurface: type,
        survivor: { recordId: survivorId, baseRowVersion: 7 },
        loser: { recordId: loserId, baseRowVersion: 2 },
      });
      expect(Object.isFrozen(reviewed)).toBe(true);
      await act(async () => {
        const confirm = t.result.current.commands.confirm;
        await Promise.all([confirm(), confirm()]);
      });
      expect(t.merge).toHaveBeenCalledTimes(1);
      expect(t.merge).toHaveBeenCalledWith(
        expect.objectContaining({
          id: "merge-id",
          review: reviewed,
        }),
        expect.any(AbortSignal),
      );
      expect(t.discard).not.toHaveBeenCalled();
      expect(t.refresh).toHaveBeenCalledTimes(1);
      expect(t.preview).toHaveBeenCalledWith(survivorId);
      await act(() => t.result.current.commands.confirm());
      expect(t.merge).toHaveBeenCalledTimes(1);
      t.unmount();
    }
  });

  it("invalidates confirmation synchronously for either participant, reason, eligibility, view or authority", async () => {
    for (const type of ["host", "identity"] as const)
      for (const change of [
        "survivor",
        "loser",
        "survivor_version",
        "loser_version",
        "deleted",
        "merged",
        "reason",
        "view",
        "role",
        "session",
        "actor",
        "closed",
        "closure",
        "access",
      ] as const) {
        const t = setup(type);
        t.review();
        const confirm = t.result.current.commands.confirm;
        act(() => {
          switch (change) {
            case "survivor":
              t.rerender({ selectedEntity: entityRow(otherId, 1, type) });
              break;
            case "loser":
              t.result.current.commands.selectCandidate(otherId);
              break;
            case "survivor_version": {
              const updated = entityRow(survivorId, 8, type);
              t.rerender({ selectedEntity: updated, rows: [updated, t.loser] });
              break;
            }
            case "loser_version":
              t.rerender({ rows: [t.survivor, entityRow(loserId, 3, type)] });
              break;
            case "deleted":
              t.rerender({ rows: [t.survivor] });
              break;
            case "merged":
              t.rerender({
                rows: [t.survivor, { ...t.loser, state: "merged" }],
              });
              break;
            case "reason":
              t.result.current.commands.setReason("Changed after review");
              break;
            case "view":
              t.rerender({ originSurface: "other-view" });
              break;
            case "role":
              t.owner.setAuthority({ ...authority, role: "viewer" });
              break;
            case "session":
              t.owner.setAuthority({
                ...authority,
                sessionIdentity: "replacement",
              });
              break;
            case "actor":
              t.owner.setAuthority({ ...authority, actorId: "different" });
              break;
            case "closed":
              t.owner.closeIncident();
              break;
            case "closure":
              t.result.current.commands.reset();
              break;
            case "access":
              t.owner.suspend();
              break;
          }
          void confirm();
        });
        await act(async () => {});
        expect(t.merge, `${type}:${change}`).not.toHaveBeenCalled();
        expect(t.result.current.snapshot.reviewed).toBeNull();
        t.unmount();
      }
  });

  it("requires explicit completion or scoped discard of participant drafts before review", () => {
    const t = setup();
    t.drafts.mockReturnValue(true);
    t.rerender({});
    act(() => t.result.current.commands.selectCandidate(loserId));
    act(() => t.result.current.commands.review());
    expect(t.result.current.snapshot.reviewed).toBeNull();
    expect(t.discard).not.toHaveBeenCalled();
    act(() => t.result.current.commands.discardDrafts());
    expect(t.discard).toHaveBeenCalledWith([survivorId, loserId]);
    t.drafts.mockReturnValue(false);
    t.rerender({});
    act(() => t.result.current.commands.review());
    expect(t.result.current.snapshot.reviewed).not.toBeNull();
    t.unmount();
  });

  it("fences presentation callbacks after closure or authorization loss", async () => {
    for (const detach of ["closure", "access"] as const) {
      const t = setup();
      let settle: ((value: EntityMergeTransportOutcome) => void) | undefined;
      t.merge.mockImplementation(
        () =>
          new Promise((resolve) => {
            settle = resolve;
          }),
      );
      t.review();
      let pending: Promise<void> | undefined;
      act(() => {
        pending = t.result.current.commands.confirm();
      });
      await act(async () => {
        await vi.waitFor(() => expect(settle).toBeDefined());
      });
      act(() => {
        if (detach === "closure") t.result.current.commands.reset();
        else t.owner.suspend();
      });
      await act(async () => {
        settle?.(accepted("host"));
        await pending;
      });
      expect(t.preview).not.toHaveBeenCalled();
      expect(t.discard).not.toHaveBeenCalled();
      t.unmount();
    }
  });
});
