import { describe, expect, it, vi } from "vitest";
import {
  mergeAuthority,
  mergeEntityRow,
  mergeIncidentId,
  mergeReceipt,
  mergeReview,
} from "../../testing/entityMergeTestSupport";
import { createWorkbookEntityMergeAdapter } from "../adapters/createWorkbookEntityMergeAdapter";
import type { WorkbookEntityMergePort } from "../features/entities/entityMergeOperation";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";

function setup(type: "host" | "identity") {
  let id = 0;
  const transactionIds = { create: () => `txn-${++id}` };
  let settle!: (
    value: Awaited<ReturnType<WorkbookPendingMutationPort["execute"]>>,
  ) => void;
  const execute = vi.fn<WorkbookPendingMutationPort["execute"]>(
    () =>
      new Promise((resolve) => {
        settle = resolve;
      }),
  );
  const runtime = new WorkbookMutationRuntime(
    { incidentId: mergeIncidentId, clientInstanceId: "merge-test" },
    transactionIds,
    { execute },
  );
  runtime.entityMerge.setAuthority(mergeAuthority);
  runtime.history.setAuthority(mergeAuthority);
  const send = vi.fn<WorkbookEntityMergePort["send"]>(async () => ({
    kind: "acknowledged",
    receipt: mergeReceipt(type),
  }));
  runtime.entityMerge.configure({
    ...createWorkbookEntityMergeAdapter({
      apiBase: undefined,
      incidentId: mergeIncidentId,
    }),
    send,
  });
  runtime.entityMerge.registerProjectionRefresh(async () => {});
  const review = mergeReview(
    type,
    runtime.entityMerge.getSnapshot().generation,
  );
  const viewSchemaId =
    type === "host"
      ? "cartulary.view.hosts.v1"
      : "cartulary.view.identities.v1";
  const patch = (recordId: string) => ({
    baseRowVersion: recordId === review.loser.recordId ? 2 : 7,
    changes: [{ field_key: `${type}.display_name`, value: "Edited" }],
    fieldKey: `${type}.display_name`,
    localValue: "Edited",
    recordId,
    rowLabel: "Entity",
    surfaceLabel: "Entities",
    viewSchemaId,
  });
  const admit = () => {
    const attempt = runtime.entityMerge.admit(review, {
      isCurrent: () => true,
      matchesReview: () => true,
      acknowledged: () => {},
      reconcile: async () => {},
    });
    expect(attempt).not.toBeNull();
    if (attempt === null) throw new Error("Expected merge admission");
    return attempt;
  };
  return {
    runtime,
    review,
    send,
    execute,
    patch,
    admit,
    settle: (version: number, recordId = review.survivor.recordId) =>
      settle({
        kind: "accepted",
        value: {
          changeSetId: "change",
          viewSchemaId,
          row: {
            ...mergeEntityRow(recordId, version, type).rawRow,
            view_schema_id: viewSchemaId,
          },
        },
      }),
  };
}
describe("Entity merge record admission", () => {
  it("coordinates queued edits to either participant and refuses changed versions without replacing the review", async () => {
    for (const type of ["host", "identity"] as const)
      for (const participant of ["survivor", "loser"] as const) {
        const t = setup(type);
        expect(
          t.runtime.enqueuePatch(t.patch(t.review[participant].recordId)),
        ).toMatchObject({ kind: "admitted" });
        await vi.waitFor(() => expect(t.execute).toHaveBeenCalledTimes(1));
        const attempt = t.admit();
        const pending = t.runtime.entityMerge.execute(attempt);
        expect(t.send).not.toHaveBeenCalled();
        expect(
          t.runtime.enqueuePatch(t.patch(t.review.loser.recordId)),
        ).toMatchObject({ kind: "rejected_mutation" });
        t.settle(
          t.review[participant].baseRowVersion + 1,
          t.review[participant].recordId,
        );
        await pending;
        expect(t.send).not.toHaveBeenCalled();
        expect(attempt.review.survivor.baseRowVersion).toBe(7);
        expect(t.runtime.entityMerge.getSnapshot().entries[0]?.phase).toBe(
          "rejected",
        );
        t.runtime.invalidate({ kind: "runtime_disposed" });
      }
  });
  it("coordinates direct writes and synchronously rejects overlapping commands while unrelated records remain usable", async () => {
    for (const type of ["host", "identity"] as const) {
      const t = setup(type);
      t.runtime.explicitPatches.setAuthority(t.review.authority);
      t.runtime.explicitPatches.configure(
        { send: vi.fn() },
        undefined,
        async (_view, id) => t.runtime.explicitPatches.latestRow(id),
      );
      const release = t.runtime.beginEntityWrite({
        recordIds: [t.review.loser.recordId],
      });
      expect(release).not.toBeNull();
      const attempt = t.admit();
      const pending = t.runtime.entityMerge.execute(attempt);
      expect(t.send).not.toHaveBeenCalled();
      expect(
        t.runtime.beginEntityWrite({ recordIds: [t.review.survivor.recordId] }),
      ).toBeNull();
      expect(
        t.runtime.beginEntityWrite({ recordIds: [], unknownEntityType: type }),
      ).toBeNull();
      const unrelated = t.runtime.beginEntityWrite({
        recordIds: ["unrelated"],
      });
      expect(unrelated).not.toBeNull();
      unrelated?.();
      expect(
        await t.runtime.explicitPatches.submit({
          ...t.patch(t.review.survivor.recordId),
          baseline: {
            record_id: t.review.survivor.recordId,
            row_version: 7,
            cells: {},
          },
          sheetRef: { kind: "view_schema", id: t.patch("id").viewSchemaId },
          surfaceLabel: "Entities",
          purpose: "entity-edit",
        }),
      ).toMatchObject({
        phase: "preparation_failed",
        failure: { kind: "stale_target" },
      });
      expect(
        t.runtime.history.admit(
          {
            subject: {
              kind: "live",
              recordId: t.review.survivor.recordId,
              rowVersion: 7,
              label: "Entity",
              surfaceLabel: "Entities",
              viewSchemaId: t.patch("id").viewSchemaId,
            },
            pending: {
              kind: "destructive",
              operation: "delete",
              recordId: t.review.survivor.recordId,
              rowVersion: 7,
            },
          },
          {
            isCurrent: () => true,
            coordinate: async () => 7,
            acknowledged: () => {},
            reconcile: async () => {},
          },
        ),
      ).toBeNull();
      release?.();
      await pending;
      expect(t.send).toHaveBeenCalledTimes(1);
      expect(t.runtime.entityMerge.getSnapshot().entries[0]?.phase).toBe(
        "acknowledged",
      );
      t.runtime.invalidate({ kind: "runtime_disposed" });
    }
  });
});
