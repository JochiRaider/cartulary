import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useEffect } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { acceptedQueryMetadata } from "../../testing/workbookQueryTestSupport";
import { useWorkbookQueryController } from "../hooks/useWorkbookQueryController";
import { notesViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useGenericSurfaceQuery } from "./useGenericSurfaceQuery";
import { WorkbookQueryBrowsingProvider } from "./WorkbookQueryBrowsingContext";
import { WorkbookQueryBrowsingControls } from "./WorkbookQueryBrowsingControls";
import type {
  WorkbookViewQueryPort,
  WorkbookViewQueryResult,
} from "./WorkbookViewQueryPort";

const contract = requireViewContract(notesViewSchemaId);
const recordId = (number: number) =>
  `00000000-0000-4000-8000-${String(number).padStart(12, "0")}`;
function Harness({ port }: { readonly port: WorkbookViewQueryPort }) {
  const intent = useWorkbookQueryController({ surface: notesViewSchemaId });
  const query = useGenericSurfaceQuery({
    active: true,
    contract,
    queryState: intent.snapshot.genericQueryState,
    viewQuery: port,
    viewSchemaId: notesViewSchemaId,
    onAuthorityUncertain: undefined,
  });
  useEffect(() => {
    void query.refresh();
  }, [query.refresh]);
  return (
    <>
      <button
        type="button"
        onClick={() =>
          intent.snapshot.activeQueryControls.onSortChange([
            { fieldKey: "note.title", direction: "desc" },
          ])
        }
      >
        Sort titles descending
      </button>
      <WorkbookQueryBrowsingControls viewSchemaId={notesViewSchemaId} />
      <output aria-label="Applied sort">
        {query.acceptedQueryState.sort[0]?.direction ?? "default"}
      </output>
      <output aria-label="Visible rows">
        {query.rows
          .map((row) => String(row.cells["note.title"]?.value))
          .join(",")}
      </output>
    </>
  );
}
function fixture() {
  const cursors = new Map([
    [" token+/100= ", 100],
    [" token+/200= ", 200],
    [" token+/300= ", 300],
    [" token+/400= ", 400],
  ]);
  const query = vi.fn<WorkbookViewQueryPort["query"]>(async (input) => {
    const start =
      input.cursorToken === undefined ? 0 : cursors.get(input.cursorToken);
    if (start === undefined) throw new Error("Unknown fixture cursor");
    const end = Math.min(start + 100, 405);
    return {
      kind: "accepted",
      value: {
        ...acceptedQueryMetadata(notesViewSchemaId, input.queryState),
        incidentId: "incident",
        viewSchemaId: notesViewSchemaId,
        producingRequest: {
          queryState: input.queryState,
          limit: 100,
          ...(input.cursorToken === undefined
            ? {}
            : { cursorToken: input.cursorToken }),
        },
        rows: Array.from({ length: end - start }, (_, index) =>
          fullWorkbookViewRow(contract, recordId(start + index), 1, {
            "note.title": `Note ${start + index}`,
            "note.body": "body",
          }),
        ),
        paging: {
          limit: 100,
          hasMore: end < 405,
          nextCursor: end < 405 ? ` token+/${end}= ` : null,
        },
      },
    };
  });
  render(
    <WorkbookQueryBrowsingProvider>
      <Harness port={{ query }} />
    </WorkbookQueryBrowsingProvider>,
  );
  return query;
}
afterEach(cleanup);

it("reaches later records and returns by keyboard while preserving continuation control focus", async () => {
  const query = fixture();
  const user = userEvent.setup();
  await screen.findByText("100 records loaded; more available.");
  const more = screen.getByRole("button", { name: "Load more" });
  more.focus();
  await user.keyboard("{Enter}");
  await screen.findByText("200 records loaded; more available.");
  expect(document.activeElement).toBe(more);
  await user.keyboard("{Enter}");
  await screen.findByText("300 records loaded; more available.");
  await user.keyboard("{Enter}");
  await waitFor(() =>
    expect(screen.getByLabelText("Visible rows").textContent).toContain(
      "Note 399",
    ),
  );
  expect(screen.getByLabelText("Visible rows").textContent?.split(",")[0]).toBe(
    "Note 100",
  );
  expect(query).toHaveBeenCalledTimes(4);
  const earlier = screen.getByRole("button", { name: "Earlier rows" });
  earlier.focus();
  await user.keyboard("{Enter}");
  await waitFor(() =>
    expect(
      screen.getByLabelText("Visible rows").textContent?.split(",")[0],
    ).toBe("Note 0"),
  );
  expect(query.mock.calls.at(-1)?.[0].cursorToken).toBeUndefined();
  expect(document.activeElement).toBe(earlier);
});

it("retains accepted presentation on a failed replacement and offers local retry and revert", async () => {
  const query = fixture();
  const user = userEvent.setup();
  await screen.findByText("100 records loaded; more available.");
  const pending = deferred<WorkbookViewQueryResult>();
  query.mockImplementationOnce(() => pending.promise);
  await user.click(
    screen.getByRole("button", { name: "Sort titles descending" }),
  );
  expect(screen.getByLabelText("Applied sort").textContent).toBe("default");
  await act(() =>
    pending.resolve({
      kind: "rejected",
      failure: { kind: "retryable", message: "Offline" },
    }),
  );
  expect(screen.getByText("Query changes are unapplied.")).toBeTruthy();
  expect(
    screen
      .getByRole("button", { name: "Load more" })
      .getAttribute("aria-disabled"),
  ).toBe("true");
  await user.click(screen.getByRole("button", { name: "Revert" }));
  await waitFor(() =>
    expect(screen.queryByRole("button", { name: "Revert" })).toBeNull(),
  );
  expect(screen.getByLabelText("Applied sort").textContent).toBe("default");
  query.mockResolvedValueOnce({
    kind: "rejected",
    failure: { kind: "retryable", message: "Offline" },
  });
  await user.click(
    screen.getByRole("button", { name: "Sort titles descending" }),
  );
  await screen.findByRole("button", { name: "Retry" });
  await user.click(screen.getByRole("button", { name: "Retry" }));
  await waitFor(() =>
    expect(screen.getByLabelText("Applied sort").textContent).toBe("desc"),
  );
  expect(query.mock.calls.at(-1)?.[0].queryState.sort).toEqual([
    { fieldKey: "note.title", direction: "desc" },
  ]);
});
