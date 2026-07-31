import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { EntityRow } from "../../models/entityWorkbookModel";
import type {
  EntityMergeOutcome,
  EntityMutationCommandPort,
} from "../../mutations/workbookMutationCommandPorts";
import { useEntityMergeController } from "./useEntityMergeController";

const survivorId = "00000000-0000-4000-8000-000000007100";
const loserId = "00000000-0000-4000-8000-000000007101";

function entityRow(
  recordId: string,
  label: string,
  rowVersion: number,
): EntityRow {
  return {
    aliasTexts: [],
    aliases: [],
    entityType: "host",
    identifiers: [],
    label,
    linkedEventCount: 0,
    rawRow: { cells: {}, record_id: recordId, row_version: rowVersion },
    recordId,
    reusableIdentifiers: [],
    rowVersion,
    secondaryText: "",
    state: "canonical",
  };
}

function mergePort(
  merge: EntityMutationCommandPort["merge"],
): EntityMutationCommandPort {
  return {
    canCreateRecord: () => false,
    createRecord: async () => ({
      kind: "rejected",
      failure: { kind: "terminal", message: "not used" },
    }),
    merge,
    pasteCreate: async () => ({
      kind: "rejected",
      failure: { kind: "terminal", message: "not used" },
    }),
    patchRecord: async () => ({
      kind: "rejected",
      failure: { kind: "terminal", message: "not used" },
    }),
  };
}

function acceptedMerge(): EntityMergeOutcome {
  return {
    kind: "accepted",
    value: {
      changeSetId: "00000000-0000-4000-8000-000000007200",
      loserRecordId: loserId,
      loserRowVersion: 2,
      mergedIntoRecordId: survivorId,
      recordType: "host",
      survivorRecordId: survivorId,
      survivorRowVersion: 8,
    },
  };
}

describe("useEntityMergeController", () => {
  it("owns accepted merge cleanup, refresh, preview, and survivor continuity", async () => {
    const survivor = entityRow(survivorId, "Survivor", 7);
    const loser = entityRow(loserId, "Loser", 2);
    const clearDrafts = vi.fn();
    const loadSurvivorPreview = vi.fn(async () => undefined);
    const onRefreshEntities = vi.fn(async () => undefined);
    const retargetSurvivor = vi.fn();
    const merge = vi.fn(async () => acceptedMerge());
    const { result } = renderHook(() =>
      useEntityMergeController({
        canMerge: true,
        clearDrafts,
        lifecycleResetKey: "incident-1:authorized",
        loadSurvivorPreview,
        mutationCommands: mergePort(merge),
        onRefreshEntities,
        retargetSurvivor,
        rows: [survivor, loser],
        selectedEntity: survivor,
      }),
    );

    act(() => result.current.commands.selectCandidate(loserId));
    await act(async () => result.current.commands.confirm());

    expect(merge).toHaveBeenCalledWith({
      loserBaseRowVersion: 2,
      loserRecordId: loserId,
      reason: "Merge duplicate entity",
      survivorBaseRowVersion: 7,
      survivorRecordId: survivorId,
    });
    expect(clearDrafts).toHaveBeenCalledTimes(1);
    expect(onRefreshEntities).toHaveBeenCalledTimes(1);
    expect(loadSurvivorPreview).toHaveBeenCalledWith(survivorId);
    expect(retargetSurvivor).toHaveBeenCalledWith(survivorId);
    expect(result.current.snapshot.candidateId).toBe("");
    expect(result.current.snapshot.message).toBe(
      "Merged Loser into Survivor (host).",
    );
  });

  it("rejects a late merge completion after lifecycle authorization changes", async () => {
    const survivor = entityRow(survivorId, "Survivor", 7);
    const loser = entityRow(loserId, "Loser", 2);
    let resolveMerge: ((outcome: EntityMergeOutcome) => void) | undefined;
    const merge = vi.fn(
      () =>
        new Promise<EntityMergeOutcome>((resolve) => {
          resolveMerge = resolve;
        }),
    );
    const clearDrafts = vi.fn();
    const onRefreshEntities = vi.fn(async () => undefined);
    const retargetSurvivor = vi.fn();
    const { result, rerender } = renderHook(
      ({ lifecycleResetKey }: { lifecycleResetKey: string }) =>
        useEntityMergeController({
          canMerge: true,
          clearDrafts,
          lifecycleResetKey,
          loadSurvivorPreview: async () => undefined,
          mutationCommands: mergePort(merge),
          onRefreshEntities,
          retargetSurvivor,
          rows: [survivor, loser],
          selectedEntity: survivor,
        }),
      { initialProps: { lifecycleResetKey: "authorized" } },
    );

    act(() => result.current.commands.selectCandidate(loserId));
    let completion: Promise<void> | undefined;
    await act(async () => {
      completion = result.current.commands.confirm();
      await Promise.resolve();
    });
    rerender({ lifecycleResetKey: "authorization-lost" });
    await act(async () => {
      resolveMerge?.(acceptedMerge());
      await completion;
    });

    expect(clearDrafts).not.toHaveBeenCalled();
    expect(onRefreshEntities).not.toHaveBeenCalled();
    expect(retargetSurvivor).not.toHaveBeenCalled();
    expect(result.current.snapshot.message).toBeNull();
    expect(result.current.snapshot.candidateId).toBe("");
  });
});
