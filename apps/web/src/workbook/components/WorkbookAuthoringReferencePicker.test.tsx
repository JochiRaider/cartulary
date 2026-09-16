import { partiesViewSchemaId } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { StrictMode } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { WorkbookCandidateAuthorityContext } from "../hooks/useWorkbookCandidateDiscovery";
import type {
  WorkbookAuthoringReadPort,
  WorkbookAuthoringSelection,
} from "../ports/WorkbookAuthoringReadPort";
import { WorkbookAuthoringReferenceControl } from "./WorkbookAuthoringReferenceControl";
import { WorkbookAuthoringReferencePicker } from "./WorkbookAuthoringReferencePicker";

afterEach(cleanup);
const view = partiesViewSchemaId;
const candidate = (id: string) => ({
  recordId: id,
  displayText: `Party ${id}`,
  viewSchemaId: view,
  row: {
    record_id: id,
    row_version: 3,
    cells: { secret: { value: "Unnecessary payload" } },
  },
});
const page = (
  number: number,
  nextCursor: string | null = `page-${number + 1}`,
) => ({
  kind: "accepted" as const,
  value: {
    candidates: Array.from({ length: 100 }, (_, index) =>
      candidate(`${number}-${index}`),
    ),
    hasMore: nextCursor !== null,
    nextCursor,
  },
});
function setup() {
  const reader: WorkbookAuthoringReadPort = {
    availableViews: vi.fn(async () => ({
      kind: "accepted" as const,
      value: [view],
    })),
    verify: vi.fn(async () => {}),
    page: vi.fn(async ({ cursor }) =>
      page(cursor ? Number(cursor.slice(5)) : 1),
    ),
  };
  const onApply = vi.fn(),
    onCancel = vi.fn();
  const props = {
    label: "Parties",
    targetKey: "draft:references",
    testId: "candidates",
    views: [view],
    multiple: true,
    maximum: 64,
    selected: [] as readonly WorkbookAuthoringSelection[],
    reader,
    revision: 0,
    disabled: false,
    onApply,
    onCancel,
  };
  return { reader, props, onApply, onCancel };
}
const select = (...ids: string[]) => {
  const list = screen.getByRole("listbox", {
    name: "Parties",
  }) as HTMLSelectElement;
  for (const option of list.options)
    option.selected = ids.includes(option.value);
  fireEvent.change(list);
};
const click = (name: string) =>
  fireEvent.click(screen.getByRole("button", { name }));
describe("authoring candidate presentation", () => {
  it("retains selected identities across twelve evicted pages with bounded options and reduced payloads", async () => {
    const { props, onApply } = setup();
    render(<WorkbookAuthoringReferencePicker {...props} />);
    await screen.findByRole("option", { name: "Party 1-0" });
    select("1-0");
    for (let number = 2; number <= 13; number++) {
      click("Next candidates");
      await screen.findByRole("option", { name: `Party ${number}-0` });
      expect(
        screen
          .getAllByRole("option")
          .filter(
            (option) =>
              option.parentElement?.getAttribute("data-testid") ===
              "candidates",
          ),
      ).toHaveLength(100);
    }
    select("13-0");
    expect(screen.queryByRole("option", { name: "Party 1-0" })).toBeNull();
    expect(
      screen.getByRole("button", { name: "Remove selected Parties Party 1-0" }),
    ).toBeTruthy();
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([
      {
        recordId: "1-0",
        displayText: "Party 1-0",
        viewSchemaId: view,
      },
      {
        recordId: "13-0",
        displayText: "Party 13-0",
        viewSchemaId: view,
      },
    ]);
    click("Remove selected Parties Party 1-0");
    click("Apply references");
    expect(onApply.mock.calls.at(-1)?.[0]).toHaveLength(1);
  });
  it("permits accepted-page selection and Apply during delayed and failed continuation then retries exactly once", async () => {
    const { props, reader, onApply } = setup();
    let finish!: (
      value: Awaited<ReturnType<WorkbookAuthoringReadPort["page"]>>,
    ) => void;
    vi.mocked(reader.page)
      .mockResolvedValueOnce(page(1))
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finish = resolve;
          }),
      )
      .mockResolvedValueOnce(page(2));
    render(<WorkbookAuthoringReferencePicker {...props} />);
    await screen.findByRole("option", { name: "Party 1-0" });
    click("Next candidates");
    click("Next candidates");
    select("1-0");
    click("Apply references");
    expect(onApply.mock.calls[0]?.[0][0].recordId).toBe("1-0");
    expect(reader.page).toHaveBeenCalledTimes(2);
    await act(async () =>
      finish({
        kind: "rejected",
        failure: { kind: "retryable", message: "Continuation offline" },
      }),
    );
    expect(screen.getByRole("alert").textContent).toContain(
      "accepted page remains available",
    );
    select("1-0", "1-1");
    click("Apply references");
    expect(onApply.mock.calls.at(-1)?.[0]).toHaveLength(2);
    click("Retry candidates");
    await screen.findByRole("option", { name: "Party 2-0" });
    expect(vi.mocked(reader.page).mock.calls[2]?.[0].cursor).toBe("page-2");
    expect(vi.mocked(reader.page).mock.calls[2]?.[0].queryState).toEqual(
      vi.mocked(reader.page).mock.calls[1]?.[0].queryState,
    );
  });
  it("applies query edits explicitly without clearing selection and rejects over-limit changes", async () => {
    const { props, reader, onApply } = setup();
    render(<WorkbookAuthoringReferencePicker {...props} maximum={1} />);
    await screen.findByRole("option", { name: "Party 1-0" });
    select("1-0");
    select("1-0", "1-1");
    expect(screen.getByRole("alert").textContent).toContain("at most 1");
    fireEvent.change(screen.getByLabelText("Parties order"), {
      target: { value: "party.display_name:desc" },
    });
    fireEvent.change(screen.getByLabelText("Parties filter field"), {
      target: { value: "party.display_name" },
    });
    fireEvent.change(screen.getByLabelText("Parties filter operator"), {
      target: { value: "eq" },
    });
    fireEvent.change(screen.getByLabelText("Parties filter value"), {
      target: { value: "Deliberate" },
    });
    click("Add filter");
    expect(reader.page).toHaveBeenCalledTimes(1);
    click("Apply candidate query");
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(2));
    expect(vi.mocked(reader.page).mock.calls[1]?.[0]).toMatchObject({
      cursor: null,
      queryState: {
        filters: [
          {
            fieldKey: "party.display_name",
            op: "eq",
            arg: { value: "Deliberate" },
          },
        ],
      },
    });
    click("Apply references");
    expect(
      onApply.mock.calls[0]?.[0].map(
        (item: WorkbookAuthoringSelection) => item.recordId,
      ),
    ).toEqual(["1-0"]);
  });
  it("cancels unapplied staging restores invoking focus and reopens with parent selection", async () => {
    const { props, onApply } = setup();
    render(
      <WorkbookAuthoringReferenceControl
        {...props}
        selected={[
          { recordId: "off-page", displayText: "Retained", viewSchemaId: view },
        ]}
      />,
    );
    click("Choose parties");
    await screen.findByRole("option", { name: "Party 1-0" });
    select("1-0");
    fireEvent.keyDown(screen.getByRole("listbox"), { key: "Escape" });
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Choose parties" }),
    );
    expect(onApply).not.toHaveBeenCalled();
    click("Choose parties");
    await screen.findByRole("option", { name: "Party 1-0" });
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([
      { recordId: "off-page", displayText: "Retained", viewSchemaId: view },
    ]);
  });
  it("conceals typed authorization failures and fences obsolete authority callbacks on target replacement", async () => {
    const { props, reader } = setup();
    const onAuthorityFailure = vi.fn();
    let finish!: (
      value: Awaited<ReturnType<WorkbookAuthoringReadPort["page"]>>,
    ) => void;
    vi.mocked(reader.page)
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finish = resolve;
          }),
      )
      .mockResolvedValueOnce(page(2));
    const boundary = (targetKey: string, identity = "session-a") => (
      <WorkbookCandidateAuthorityContext.Provider
        value={{ identity, canRead: true, onAuthorityFailure }}
      >
        <WorkbookAuthoringReferencePicker
          key={targetKey}
          {...props}
          targetKey={targetKey}
          selected={[
            {
              recordId: "private",
              displayText: "Protected label",
              viewSchemaId: view,
            },
          ]}
        />
      </WorkbookCandidateAuthorityContext.Provider>
    );
    const rendered = render(boundary("old"));
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(1));
    rendered.rerender(boundary("current"));
    await screen.findByRole("option", { name: "Party 2-0" });
    await act(async () =>
      finish({
        kind: "rejected",
        failure: { kind: "authentication_required", message: "Expired" },
      }),
    );
    expect(onAuthorityFailure).not.toHaveBeenCalled();
    vi.mocked(reader.page).mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "authentication_required", message: "Expired" },
    });
    click("Refresh candidates");
    await waitFor(() => expect(onAuthorityFailure).toHaveBeenCalledTimes(1));
    expect(screen.queryByText("Protected label")).toBeNull();
    expect(screen.queryByRole("option", { name: "Party 2-0" })).toBeNull();
    expect(
      screen
        .getByRole("button", { name: "Apply references" })
        .hasAttribute("disabled"),
    ).toBe(true);
    rendered.rerender(boundary("current", "session-a-restored"));
    await screen.findByRole("option", { name: "Party 1-0" });
    expect(screen.queryByText("Protected label")).toBeNull();
    rendered.unmount();
    for (const [kind, message] of [
      ["authentication_required", "Session authorization needs recovery."],
      ["authorization_lost", "Incident access needs verification."],
    ] as const) {
      onAuthorityFailure.mockClear();
      vi.mocked(reader.page).mockClear();
      vi.mocked(reader.availableViews).mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind, message: "Authority read failed" },
      });
      const inventoryFailure = render(boundary(`inventory-${kind}`));
      await screen.findByText(message);
      expect(onAuthorityFailure).toHaveBeenCalledTimes(1);
      expect(reader.page).not.toHaveBeenCalled();
      expect(screen.queryByText("Protected label")).toBeNull();
      expect(
        screen.queryByRole("button", { name: "Retry surfaces" }),
      ).toBeNull();
      expect(
        screen.getByRole("button", { name: "Apply references" }),
      ).toHaveProperty("disabled", true);
      inventoryFailure.unmount();
    }
  });
  it("starts and cleans up scoped reads under StrictMode without requesting record filters for membership", async () => {
    const { props, reader } = setup();
    let signal!: AbortSignal;
    vi.mocked(reader.page).mockImplementation(async (input) => {
      signal = input.signal;
      return page(1);
    });
    const rendered = render(
      <StrictMode>
        <WorkbookAuthoringReferencePicker
          {...props}
          views={["incident_members"]}
        />
      </StrictMode>,
    );
    await screen.findByRole("option", { name: "Party 1-0" });
    expect(reader.availableViews).not.toHaveBeenCalled();
    expect(screen.queryByText("Parties ordering and filters")).toBeNull();
    vi.mocked(reader.page).mockImplementation((input) => {
      signal = input.signal;
      return new Promise(() => {});
    });
    click("Next candidates");
    const count = vi.mocked(reader.page).mock.calls.length;
    rendered.unmount();
    expect(signal.aborted).toBe(true);
    expect(reader.page).toHaveBeenCalledTimes(count);
  });
  it("retains a reviewed source version and captures only source identity metadata for explicit replacement", async () => {
    const { props, onApply } = setup();
    render(
      <WorkbookAuthoringReferencePicker
        {...props}
        multiple={false}
        maximum={1}
        captureRowVersion
        selected={[
          {
            recordId: "1-0",
            displayText: "Reviewed source",
            viewSchemaId: view,
            rowVersion: 2,
          },
        ]}
      />,
    );
    await screen.findByRole("option", { name: "Party 1-0" });
    click("Next candidates");
    await screen.findByRole("option", { name: "Party 2-0" });
    click("Apply references");
    expect(onApply.mock.calls[0]?.[0]).toEqual([
      {
        recordId: "1-0",
        displayText: "Reviewed source",
        viewSchemaId: view,
        rowVersion: 2,
      },
    ]);
    fireEvent.change(screen.getByRole("combobox", { name: "Parties" }), {
      target: { value: "2-0" },
    });
    click("Apply references");
    expect(onApply.mock.calls[1]?.[0]).toEqual([
      {
        recordId: "2-0",
        displayText: "Party 2-0",
        viewSchemaId: view,
        rowVersion: 3,
      },
    ]);
  });
});
