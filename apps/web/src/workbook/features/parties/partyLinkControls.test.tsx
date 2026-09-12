import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { PartyLinkControls } from "./PartyLinkControls";
import {
  type PartyLinkReadPort,
  partyChanges,
  partyPairs,
  partyState,
  supportedPartyPairs,
} from "./partyLinkModel";
import { usePartyCandidates } from "./usePartyCandidates";

afterEach(cleanup);
const reader: PartyLinkReadPort = {
  page: async () => ({ rows: [], hasMore: false, nextCursor: null }),
  source: vi.fn(),
};
it("Party controls preserve all pair states and submit only the selected independent clear", async () => {
  for (const pair of partyPairs) {
    expect(
      supportedPartyPairs(requireViewContract(pair.viewSchemaId)),
    ).toContainEqual(pair);
    for (const text of ["", "Original source wording"])
      for (const reference of ["", "party-id"]) {
        const row = {
          record_id: "source",
          row_version: 1,
          cells: {
            [pair.textFieldKey]: { value: text || null },
            [pair.refFieldKey]: { value: reference || null },
          },
        };
        const patch = vi.fn();
        const create = vi.fn();
        const { unmount } = render(
          <PartyLinkControls
            pair={pair}
            row={row}
            reader={reader}
            scopeKey="source"
            disabled={false}
            onCreate={create}
            onPatch={patch}
          />,
        );
        expect(screen.getByText(partyState(row, pair))).toBeTruthy();
        await waitFor(() =>
          expect(screen.getByText("No eligible Parties found.")).toBeTruthy(),
        );
        fireEvent.click(
          screen.getByRole("button", { name: "Clear party link" }),
        );
        expect(patch.mock.calls).toEqual(reference ? [["clear_link"]] : []);
        expect(create).not.toHaveBeenCalled();
        expect(partyChanges(pair, "clear_link")).toEqual([
          { field_key: pair.refFieldKey, value: null },
        ]);
        expect(partyChanges(pair, "clear_text")).toEqual([
          { field_key: pair.textFieldKey, value: null },
        ]);
        expect(partyChanges(pair, "clear_both")).toEqual([
          { field_key: pair.textFieldKey, value: null },
          { field_key: pair.refFieldKey, value: null },
        ]);
        expect(row.cells[pair.textFieldKey]?.value).toBe(text || null);
        unmount();
      }
  }
});
it("Party creation requires deliberate kind and opt-in editable email with optional details", async () => {
  const pair = partyPairs[0],
    create = vi.fn();
  render(
    <PartyLinkControls
      pair={pair}
      row={{
        record_id: "source",
        row_version: 1,
        cells: { [pair.textFieldKey]: { value: "Ops <ops@example.test>" } },
      }}
      reader={reader}
      scopeKey="source"
      disabled={false}
      onCreate={create}
      onPatch={vi.fn()}
    />,
  );
  fireEvent.click(
    screen.getByRole("button", { name: "Create party from text" }),
  );
  expect(
    (
      screen.getByRole("button", {
        name: "Save Party and link",
      }) as HTMLButtonElement
    ).disabled,
  ).toBe(true);
  expect(
    (
      screen.getByRole("textbox", {
        name: "Display Name value",
      }) as HTMLInputElement
    ).value,
  ).toBe("Ops <ops@example.test>");
  fireEvent.change(screen.getByRole("combobox", { name: "Kind value" }), {
    target: { value: "team" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Save Party and link" }));
  expect(create.mock.calls[0]?.[0]).toMatchObject({
    "party.party_kind": "team",
    "party.primary_email": "",
  });
  fireEvent.click(screen.getByRole("checkbox", { name: /Include email/ }));
  fireEvent.change(screen.getByRole("textbox", { name: "Email value" }), {
    target: { value: "reviewed@example.test" },
  });
  fireEvent.click(screen.getByRole("button", { name: "Save Party and link" }));
  expect(create.mock.calls[1]?.[0]["party.primary_email"]).toBe(
    "reviewed@example.test",
  );
  expect(screen.getByText("Optional Party details")).toBeTruthy();
});
it("Party candidate paging retains earlier rows after failure and retries the same missing page", async () => {
  const page = vi
    .fn<PartyLinkReadPort["page"]>()
    .mockResolvedValueOnce({
      rows: [{ record_id: "first", row_version: 1, cells: {} }],
      nextCursor: "next",
      hasMore: true,
    })
    .mockRejectedValueOnce(new Error("network"))
    .mockResolvedValueOnce({
      rows: [{ record_id: "second", row_version: 1, cells: {} }],
      nextCursor: null,
      hasMore: false,
    });
  const port = { ...reader, page };
  const { result } = renderHook(() => usePartyCandidates(port, "origin"));
  await waitFor(() => expect(result.current.phase).toBe("ready"));
  await act(() => result.current.loadMore());
  expect(result.current.phase).toBe("failed");
  expect(result.current.rows.map((row) => row.record_id)).toEqual(["first"]);
  await act(() => result.current.retry());
  expect(page.mock.calls.map((call) => call[0])).toEqual([
    null,
    "next",
    "next",
  ]);
  expect(result.current.rows.map((row) => row.record_id)).toEqual([
    "first",
    "second",
  ]);
});

it("Party review Escape restores its trigger and cyclic paging remains stale until reload", async () => {
  const pair = partyPairs[0];
  const mounted = render(
    <PartyLinkControls
      pair={pair}
      row={{
        record_id: "source",
        row_version: 1,
        cells: { [pair.textFieldKey]: { value: "Source" } },
      }}
      reader={reader}
      scopeKey="source"
      disabled={false}
      onCreate={vi.fn()}
      onPatch={vi.fn()}
    />,
  );
  const trigger = screen.getByRole("button", {
    name: "Create party from text",
  });
  fireEvent.click(trigger);
  expect(document.activeElement).toBe(
    screen.getByRole("textbox", { name: "Display Name value" }),
  );
  fireEvent.keyDown(
    screen.getByRole("textbox", { name: "Display Name value" }),
    { key: "Escape" },
  );
  expect(
    screen.queryByRole("form", { name: "Review Party creation" }),
  ).toBeNull();
  expect(document.activeElement).toBe(trigger);
  mounted.unmount();
  const page = vi
    .fn<PartyLinkReadPort["page"]>()
    .mockResolvedValueOnce({
      rows: [{ record_id: "first", row_version: 1, cells: {} }],
      nextCursor: "repeat",
      hasMore: true,
    })
    .mockResolvedValueOnce({ rows: [], nextCursor: "repeat", hasMore: true })
    .mockResolvedValueOnce({ rows: [], nextCursor: null, hasMore: false });
  const port = { ...reader, page };
  const { result } = renderHook(() => usePartyCandidates(port, "origin"));
  await waitFor(() => expect(result.current.phase).toBe("ready"));
  await act(() => result.current.loadMore());
  expect(result.current.phase).toBe("failed");
  expect(result.current.rows).toHaveLength(1);
  await act(() => result.current.reload());
  expect(result.current.phase).toBe("ready");
  expect(result.current.rows).toEqual([]);
});

it("Party socket invalidation retains loaded candidates as stale until an authorized reload", async () => {
  const page = vi
    .fn<PartyLinkReadPort["page"]>()
    .mockResolvedValueOnce({
      rows: [{ record_id: "first", row_version: 1, cells: {} }],
      nextCursor: null,
      hasMore: false,
    })
    .mockRejectedValueOnce(new Error("read failed"))
    .mockResolvedValueOnce({ rows: [], nextCursor: null, hasMore: false });
  const port = { ...reader, page };
  const { result, rerender } = renderHook(
    ({ revision }) => usePartyCandidates(port, "origin", revision),
    { initialProps: { revision: 0 } },
  );
  await waitFor(() => expect(result.current.phase).toBe("ready"));
  rerender({ revision: 1 });
  await waitFor(() => expect(result.current.phase).toBe("failed"));
  expect(result.current.rows.map((row) => row.record_id)).toEqual(["first"]);
  await act(() => result.current.reload());
  expect(result.current.rows).toEqual([]);
  expect(result.current.phase).toBe("ready");

  const currentPage = {
    rows: [{ record_id: "first", row_version: 2, cells: {} }],
    nextCursor: null,
    hasMore: false,
  };
  const currentReader = {
    ...reader,
    page: vi.fn().mockResolvedValue(currentPage),
  };
  const controls = (revision: number) => (
    <PartyLinkControls
      pair={partyPairs[0]}
      row={{ record_id: "origin", row_version: 1, cells: {} }}
      reader={currentReader}
      scopeKey="origin"
      candidateRevision={revision}
      disabled={false}
      onCreate={vi.fn()}
      onPatch={vi.fn()}
    />
  );
  const panel = render(controls(0));
  await waitFor(() =>
    expect(screen.getByRole("combobox")).toHaveProperty("disabled", false),
  );
  fireEvent.change(screen.getByRole("combobox"), {
    target: { value: "first" },
  });
  expect(
    screen.getByRole("button", { name: "Link existing party" }),
  ).toHaveProperty("disabled", false);
  panel.rerender(controls(1));
  await waitFor(() => expect(currentReader.page).toHaveBeenCalledTimes(2));
  await waitFor(() =>
    expect(screen.getByRole("combobox")).toHaveProperty("disabled", false),
  );
  expect(screen.getByRole("combobox")).toHaveProperty("value", "");
  expect(
    screen.getByRole("button", { name: "Link existing party" }),
  ).toHaveProperty("disabled", true);
});
