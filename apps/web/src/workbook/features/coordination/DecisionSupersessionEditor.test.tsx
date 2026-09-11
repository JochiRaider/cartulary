import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import {
  decisionAuthority,
  decisionReplacementId,
  decisionRow,
  decisionTargetId,
} from "../../../testing/decisionSupersessionTestSupport";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { DecisionSupersessionEditor } from "./DecisionSupersessionEditor";
import {
  decisionIneligibility,
  normalizeDecisionReason,
  reviewedDecision,
} from "./decisionSupersessionModel";
import type {
  DecisionSupersessionOwnerPort,
  DecisionSupersessionSnapshot,
} from "./decisionSupersessionOperation";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});
function host() {
  const snapshot: DecisionSupersessionSnapshot = {
    authority: decisionAuthority,
    generation: 1,
    revision: 0,
    entries: [],
  };
  const port: DecisionSupersessionOwnerPort = {
    getSnapshot: () => snapshot,
    subscribe: () => () => {},
    canSubmit: () => true,
    latestRow: () => null,
    latestVersion: () => null,
    acceptRow: (row) => row,
    blocksRecord: () => false,
    page: vi.fn<DecisionSupersessionOwnerPort["page"]>(async () => ({
      kind: "accepted",
      value: {
        rows: [decisionRow(decisionReplacementId, "approved", 6)],
        hasMore: false,
        nextCursor: null,
      },
    })),
    admit: vi.fn(() => null),
    execute: vi.fn(async () => {}),
    replay: vi.fn(async () => {}),
    refresh: vi.fn(async () => {}),
    dismiss: vi.fn(),
  };
  return port;
}
function editor(
  owner: DecisionSupersessionOwnerPort,
  row = decisionRow(),
  lifecycleKey = "target",
) {
  return (
    <DecisionSupersessionEditor
      owner={owner}
      row={row}
      lifecycleKey={lifecycleKey}
      originSurface="decisions"
      onCancel={() => {}}
      reconcile={async () => {}}
    />
  );
}
it("Decision review uses an outside-grid replacement and invalidates changed inputs and record versions", async () => {
  const owner = host();
  const view = render(editor(owner));
  await screen.findByText("All visible Decision candidates loaded.");
  expect(
    screen.getByRole("option", { name: new RegExp(decisionReplacementId) })
      .textContent,
  ).toContain("approved");
  fireEvent.change(screen.getByLabelText("Superseding Decision"), {
    target: { value: decisionReplacementId },
  });
  fireEvent.change(screen.getByLabelText("Decision supersession reason"), {
    target: { value: "Later evidence" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  expect(
    screen.getByRole("region", { name: "Review Decision supersession" })
      .textContent,
  ).toContain(decisionTargetId);
  fireEvent.change(screen.getByLabelText("Decision supersession reason"), {
    target: { value: "Changed reason" },
  });
  expect(
    screen.queryByRole("button", { name: "Confirm supersession" }),
  ).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  view.rerender(editor(owner, decisionRow(decisionTargetId, "approved", 5)));
  expect(
    screen.queryByRole("button", { name: "Confirm supersession" }),
  ).toBeNull();
  expect(owner.admit).not.toHaveBeenCalled();
});
it("Decision candidates distinguish partial failure retry exhaustion and disabled duplicate-label records", async () => {
  const owner = host();
  const proposed = decisionRow("00000000-0000-4000-8000-000000000422");
  const page = vi
    .fn<DecisionSupersessionOwnerPort["page"]>()
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        rows: [decisionRow(), proposed],
        hasMore: true,
        nextCursor: "next",
      },
    })
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Candidate read failed." },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        rows: [decisionRow(decisionReplacementId, "executed", 6)],
        hasMore: false,
        nextCursor: null,
      },
    });
  owner.page = page;
  render(editor(owner));
  await screen.findByText(/These results are incomplete/);
  expect(
    screen
      .getByRole("option", { name: new RegExp(proposed.record_id) })
      .hasAttribute("disabled"),
  ).toBe(true);
  expect(
    screen
      .getByRole("option", { name: new RegExp(decisionTargetId) })
      .hasAttribute("disabled"),
  ).toBe(true);
  expect(screen.queryByText(/No eligible replacement/)).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Load more Decisions" }));
  await screen.findByText("Candidate read failed.");
  expect(
    screen.getByRole("option", { name: new RegExp(proposed.record_id) }),
  ).not.toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Retry candidates" }));
  await screen.findByText("All visible Decision candidates loaded.");
  expect(page.mock.calls.map(([cursor]) => cursor)).toEqual([
    null,
    "next",
    "next",
  ]);
  expect(
    screen.getByRole("option", { name: new RegExp(decisionReplacementId) })
      .textContent,
  ).toContain("executed");
});
it("Decision review cancellation retains draft and detached candidate callbacks cannot publish", async () => {
  const owner = host();
  const view = render(editor(owner, decisionRow(decisionTargetId, "executed")));
  await screen.findByText("All visible Decision candidates loaded.");
  fireEvent.change(screen.getByLabelText("Superseding Decision"), {
    target: { value: decisionReplacementId },
  });
  fireEvent.change(screen.getByLabelText("Decision supersession reason"), {
    target: { value: "Later evidence" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Review supersession" }));
  expect(
    screen.getByRole("region", { name: "Review Decision supersession" })
      .textContent,
  ).toContain("remains executed");
  fireEvent.click(screen.getByRole("button", { name: "Cancel review" }));
  expect(
    (
      screen.getByLabelText(
        "Decision supersession reason",
      ) as HTMLTextAreaElement
    ).value,
  ).toBe("Later evidence");
  let resolve!: (rows: WorkbookQueryRow[]) => void;
  owner.page = vi.fn<DecisionSupersessionOwnerPort["page"]>(
    () =>
      new Promise((done) => {
        resolve = (rows) =>
          done({
            kind: "accepted",
            value: { rows, hasMore: false, nextCursor: null },
          });
      }),
  );
  owner.acceptRow = vi.fn((row) => row);
  fireEvent.click(screen.getByRole("button", { name: "Refresh candidates" }));
  await waitFor(() => expect(owner.page).toHaveBeenCalledOnce());
  view.unmount();
  await act(async () => resolve([decisionRow()]));
  expect(owner.acceptRow).not.toHaveBeenCalled();
});
it("Decision eligibility and reason normalization preserve the executed distinction and fail closed", () => {
  const target = reviewedDecision(decisionRow(), decisionAuthority.incidentId);
  for (const status of ["proposed", "approved", "executed"])
    expect(
      decisionIneligibility(
        reviewedDecision(
          decisionRow(decisionTargetId, status),
          target.incidentId,
        ),
        "target",
      ),
    ).toBeNull();
  for (const status of ["rejected", "superseded", "unknown"])
    expect(
      decisionIneligibility(
        reviewedDecision(
          decisionRow(decisionTargetId, status),
          target.incidentId,
        ),
        "target",
      ),
    ).not.toBeNull();
  expect(
    decisionIneligibility({ ...target, ownerId: null }, "target"),
  ).not.toBeNull();
  expect(
    decisionIneligibility({ ...target, isSuperseded: true }, "target"),
  ).not.toBeNull();
  expect(
    decisionIneligibility(
      {
        ...target,
        recordId: decisionReplacementId,
        status: "executed",
        isSuperseded: true,
      },
      "replacement",
      target,
    ),
  ).toBeNull();
  expect(normalizeDecisionReason("\u0085 e\u0301\r\nwhy\t \u0085")).toBe(
    "é\nwhy",
  );
  for (const value of ["", " \n", "bad\u0000reason", "x".repeat(4097)])
    expect(normalizeDecisionReason(value)).toBeNull();
});
