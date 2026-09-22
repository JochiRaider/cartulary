import { fireEvent, render, within } from "@testing-library/react";
import { createElement } from "react";
import { afterEach, expect, it, vi } from "vitest";
import {
  decisionAuthority,
  decisionReceipt,
  decisionReplacementId,
  decisionReview,
  decisionRow,
  decisionTargetId,
} from "../../../testing/decisionSupersessionTestSupport";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { WorkbookRecoveryFixture } from "../../../testing/WorkbookRecoveryFixture";
import { createWorkbookDecisionSupersessionAdapter } from "../../adapters/createWorkbookDecisionSupersessionAdapter";
import type { WorkbookPendingMutationPort } from "../../ports/WorkbookPendingMutationPort";
import { createWorkbookMutationRuntime } from "../../runtime/createWorkbookMutationRuntime";
import { decisionViewId } from "./decisionSupersessionModel";
import type { DecisionSupersessionTransportPort } from "./decisionSupersessionOperation";
import { WorkbookDecisionSupersessionRecovery } from "./WorkbookDecisionSupersessionRecovery";

afterEach(() => vi.useRealTimers());
it("Decision runtime coordinates queued and direct writes without clearing unrelated drafts", async () => {
  for (const kind of ["queued", "direct"] as const) {
    vi.useFakeTimers();
    const earlier =
      deferred<Awaited<ReturnType<WorkbookPendingMutationPort["execute"]>>>();
    const execute = vi.fn<WorkbookPendingMutationPort["execute"]>(
      () => earlier.promise,
    );
    let id = 0;
    const runtime = createWorkbookMutationRuntime(
      { incidentId: decisionAuthority.incidentId, clientInstanceId: "tab" },
      { create: () => `runtime-${++id}` },
      { execute },
    );
    const owner = runtime.decisionSupersession;
    owner.setAuthority(decisionAuthority);
    const send = vi.fn<DecisionSupersessionTransportPort["send"]>(async () => ({
      kind: "acknowledged",
      receipt: decisionReceipt(),
    }));
    owner.configure({
      ...createWorkbookDecisionSupersessionAdapter({
        apiBase: undefined,
        incidentId: decisionAuthority.incidentId,
      }),
      send,
    });
    const patch = {
      baseRowVersion: 4,
      changes: [{ field_key: "decision.rationale", value: "Earlier draft" }],
      fieldKey: "decision.rationale",
      localValue: "Earlier draft",
      recordId: decisionTargetId,
      rowLabel: "Target",
      surfaceLabel: "Decisions",
      viewSchemaId: decisionViewId,
    };
    let release: (() => void) | null = null;
    if (kind === "queued") {
      runtime.enqueuePatch(patch);
      await vi.advanceTimersByTimeAsync(1);
    } else release = runtime.beginDecisionWrite([decisionReplacementId]);
    const unrelated = {
      ...patch,
      recordId: "unrelated",
      localValue: "Unrelated draft",
    };
    runtime.enqueuePatch(unrelated);
    const review = decisionReview("proposed", owner.getSnapshot().generation);
    const attempt = owner.admit(review, {
      isCurrent: () => true,
      matchesReview: () => true,
      reconcile: async () => {},
    });
    if (!attempt) throw new Error("Expected admission");
    const running = owner.execute(attempt);
    await vi.advanceTimersByTimeAsync(1);
    expect(send).not.toHaveBeenCalled();
    expect(runtime.enqueuePatch(patch).kind).toBe("rejected_mutation");
    expect(runtime.beginDecisionWrite([decisionTargetId])).toBeNull();
    if (kind === "queued")
      earlier.resolve({
        kind: "accepted",
        value: {
          changeSetId: "earlier",
          viewSchemaId: decisionViewId,
          row: {
            ...decisionRow(decisionTargetId, "proposed", 5),
            view_schema_id: decisionViewId,
          },
        },
      });
    else {
      owner.acceptRow(decisionRow(decisionReplacementId, "approved", 7));
      release?.();
    }
    await vi.advanceTimersByTimeAsync(32);
    await running;
    expect(send).not.toHaveBeenCalled();
    expect(owner.getSnapshot().entries[0]?.phase).toBe("rejected");
    const recovery = render(
      createElement(
        WorkbookRecoveryFixture,
        null,
        createElement(WorkbookDecisionSupersessionRecovery, { runtime }),
      ),
    );
    fireEvent.click(recovery.getByRole("button", { name: /^Recovery \(/ }));
    fireEvent.click(
      recovery.getByRole("button", {
        name: /Decision supersession ·/,
      }),
    );
    expect(
      within(
        recovery.getByRole("region", { name: "Decision action recovery" }),
      ).getByRole("status").textContent,
    ).toBe(
      "Supersession rejected. Refresh and review again before a new attempt.",
    );
    recovery.unmount();
    // Supersession never discards the FIFO, even when its review becomes stale.
    expect(
      execute.mock.calls.some(
        ([input]) => input.unit.recordId === "unrelated",
      ) ||
        runtime.visibleEdit(
          decisionViewId,
          "unrelated",
          "decision.rationale",
        ) !== undefined,
    ).toBe(true);
    runtime.invalidate({ kind: "runtime_disposed" });
    vi.useRealTimers();
  }
});

it("Decision admission includes versions already accepted by History", () => {
  const ids = { create: vi.fn(() => "unused") };
  const runtime = createWorkbookMutationRuntime(
    { incidentId: decisionAuthority.incidentId, clientInstanceId: "tab" },
    ids,
    { execute: vi.fn() },
  );
  runtime.decisionSupersession.configure(
    createWorkbookDecisionSupersessionAdapter({
      apiBase: undefined,
      incidentId: decisionAuthority.incidentId,
    }),
  );
  runtime.decisionSupersession.setAuthority(decisionAuthority);
  runtime.history.acceptVersion(decisionReplacementId, 7);
  expect(
    runtime.decisionSupersession.admit(decisionReview(), {
      isCurrent: () => true,
      matchesReview: () => true,
      reconcile: async () => {},
    }),
  ).toBeNull();
  expect(ids.create).not.toHaveBeenCalled();
  runtime.invalidate({ kind: "runtime_disposed" });
});
