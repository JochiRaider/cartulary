import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  mergeAuthority,
  mergeIncidentId,
  mergeReceipt,
  mergeReview,
} from "../../../testing/entityMergeTestSupport";
import { WorkbookRecoveryFixture } from "../../../testing/WorkbookRecoveryFixture";
import { createWorkbookEntityMergeAdapter } from "../../adapters/createWorkbookEntityMergeAdapter";
import { WorkbookHistoryContext } from "../../history/WorkbookHistoryContext";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { WorkbookEntityMergePort } from "./entityMergeOperation";
import { WorkbookEntityMergeRecovery } from "./WorkbookEntityMergeRecovery";

afterEach(cleanup);
function setup(type: "host" | "identity") {
  const runtime = new WorkbookMutationRuntime(
    { incidentId: mergeIncidentId, clientInstanceId: "merge-ui" },
    { create: () => "retained-merge" },
    { execute: vi.fn() },
  );
  const owner = runtime.entityMerge;
  owner.setAuthority(mergeAuthority);
  runtime.history.setAuthority(mergeAuthority);
  const send = vi.fn<WorkbookEntityMergePort["send"]>(async () => ({
    kind: "acknowledged",
    receipt: mergeReceipt(type),
  }));
  owner.configure({
    ...createWorkbookEntityMergeAdapter({
      apiBase: undefined,
      incidentId: mergeIncidentId,
    }),
    send,
  });
  const refresh = vi.fn(async () => {});
  owner.registerProjectionRefresh(refresh);
  const attempt = owner.admit(mergeReview(type), {
    isCurrent: () => false,
    matchesReview: () => true,
    acknowledged: vi.fn(),
    reconcile: vi.fn(),
  });
  // Admission is deliberately presentation-bound; detach after admission.
  expect(attempt).toBeNull();
  let current = true;
  const admitted = owner.admit(mergeReview(type), {
    isCurrent: () => current,
    matchesReview: () => true,
    acknowledged: vi.fn(),
    reconcile: vi.fn(),
  });
  expect(admitted).not.toBeNull();
  if (admitted === null) throw new Error("Expected merge admission");
  const view = render(
    <WorkbookRecoveryFixture>
      <WorkbookHistoryContext.Provider value={runtime}>
        <WorkbookEntityMergeRecovery runtime={runtime} />
      </WorkbookHistoryContext.Provider>
    </WorkbookRecoveryFixture>,
  );
  return {
    ...view,
    owner,
    runtime,
    send,
    refresh,
    attempt: admitted,
    detach: () => {
      current = false;
    },
  };
}
describe("Entity merge recovery presentation", () => {
  it("offers keyboard recovery, receipt completion and refresh-only retry for both entity types", async () => {
    for (const type of ["host", "identity"] as const) {
      const t = setup(type);
      t.send.mockResolvedValueOnce({ kind: "uncertain" });
      await act(() => t.owner.execute(t.attempt));
      t.detach();
      const trigger = screen.getByRole("button", { name: "Recovery (1)" });
      trigger.focus();
      fireEvent.click(trigger);
      if (screen.queryByText("Completed", { selector: "summary" }))
        fireEvent.click(screen.getByText("Completed", { selector: "summary" }));
      fireEvent.click(screen.getByRole("button", { name: /^Entity merge ·/ }));
      const panel = screen.getByRole("region", {
        name: "Merge action recovery",
      });
      expect(
        screen.getByRole("heading", { name: "Entity merge", level: 2 }),
      ).toBe(document.activeElement);
      expect(panel.textContent).toContain("outcome is unknown");
      t.refresh.mockRejectedValueOnce(new Error("Refresh failed"));
      fireEvent.click(
        within(panel).getByRole("button", {
          name: "Replay exact merge request",
        }),
      );
      await waitFor(() =>
        expect(panel.textContent).toContain("Refresh is still required"),
      );
      expect(panel.textContent).toContain(mergeReceipt(type).change_set_id);
      expect(panel.textContent).toContain(
        mergeReceipt(type).survivor_record_id,
      );
      expect(t.send).toHaveBeenCalledTimes(2);
      expect(
        within(panel).queryByRole("button", { name: "Dismiss merge action" }),
      ).toBeNull();
      fireEvent.click(
        within(panel).getByRole("button", { name: "Refresh completed merge" }),
      );
      await waitFor(() =>
        expect(panel.textContent).toContain("current projections refreshed"),
      );
      expect(t.send).toHaveBeenCalledTimes(2);
      expect(
        within(panel).getByRole("button", { name: "Review merge history" }),
      ).not.toBeNull();
      fireEvent.keyDown(panel, { key: "Escape" });
      expect(trigger).toBe(document.activeElement);
      expect(
        screen.queryByRole("region", {
          name: "Merge action recovery",
        }),
      ).toBeNull();
      fireEvent.click(trigger);
      if (screen.queryByText("Completed", { selector: "summary" }))
        fireEvent.click(screen.getByText("Completed", { selector: "summary" }));
      fireEvent.click(screen.getByRole("button", { name: /^Entity merge ·/ }));
      fireEvent.click(
        screen.getByRole("button", { name: "Dismiss merge action" }),
      );
      expect(screen.queryByRole("button", { name: "Recovery (1)" })).toBeNull();
      t.unmount();
      t.runtime.invalidate({ kind: "runtime_disposed" });
    }
  });
  it("conceals all protected recovery content during access loss and restores only the permitted memory lifetime", async () => {
    const t = setup("host");
    t.send.mockResolvedValueOnce({ kind: "uncertain" });
    await act(() => t.owner.execute(t.attempt));
    fireEvent.click(screen.getByRole("button", { name: "Recovery (1)" }));
    fireEvent.click(screen.getByRole("button", { name: /^Entity merge ·/ }));
    act(() => t.owner.suspend());
    expect(screen.queryByText(/Historical loser/)).toBeNull();
    expect(screen.queryByRole("button", { name: "Recovery (1)" })).toBeNull();
    act(() =>
      t.owner.setAuthority({ ...mergeAuthority, sessionIdentity: "recovered" }),
    );
    expect(screen.getByRole("button", { name: "Recovery (1)" })).not.toBeNull();
    expect(
      screen.queryByRole("region", {
        name: "Merge action recovery",
      }),
    ).toBeNull();
    act(() =>
      t.owner.setAuthority({ ...mergeAuthority, actorId: "different-account" }),
    );
    expect(screen.queryByRole("button", { name: "Recovery (1)" })).toBeNull();
    t.runtime.invalidate({ kind: "runtime_disposed" });
  });
});
