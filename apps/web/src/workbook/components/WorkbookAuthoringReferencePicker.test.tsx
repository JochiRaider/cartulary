import { partiesViewSchemaId } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StrictMode, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { WorkbookCandidateAuthorityContext } from "../hooks/useWorkbookCandidateDiscovery";
import type {
  WorkbookAuthoringReadPort,
  WorkbookAuthoringSelection,
} from "../ports/WorkbookAuthoringReadPort";
import { WorkbookAuthoringReferenceControl } from "./WorkbookAuthoringReferenceControl";
import { WorkbookAuthoringReferencePicker } from "./WorkbookAuthoringReferencePicker";
import { WorkbookCandidateSelection } from "./WorkbookCandidateSelection";

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
const candidateName = (label: string) =>
  new RegExp(`^${label.trim().replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}\\s*\\(`);
const select = (...ids: string[]) => {
  const list = screen.getByRole("group", { name: "Parties" });
  for (const input of list.querySelectorAll<HTMLInputElement>(
    'input[type="checkbox"]',
  ))
    if (input.checked !== ids.includes(input.value)) fireEvent.click(input);
};
const click = (name: string) =>
  fireEvent.click(screen.getByRole("button", { name }));
function readGate() {
  let resolve!: (
    result: Awaited<ReturnType<WorkbookAuthoringReadPort["page"]>>,
  ) => void;
  const promise = new Promise<
    Awaited<ReturnType<WorkbookAuthoringReadPort["page"]>>
  >((complete) => {
    resolve = complete;
  });
  return { promise, resolve };
}
describe("authoring candidate presentation", () => {
  it("adds equal-label identities with plain clicks and Space while preserving editable context and metadata", async () => {
    const user = userEvent.setup();
    const { props, reader, onApply, onCancel } = setup();
    const source = { ...candidate("source"), rowVersion: 7 };
    const a = {
      ...candidate("a"),
      displayText: "Same complete candidate label",
    };
    const b = { ...candidate("b"), displayText: a.displayText };
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: {
        candidates: [source, a, b],
        nextCursor: null,
        hasMore: false,
      },
    });
    render(
      <WorkbookAuthoringReferencePicker
        {...props}
        selected={[source]}
        maximum={3}
      />,
    );
    const first = await screen.findByRole("checkbox", {
      name: `${a.displayText} (a)`,
    });
    const second = screen.getByRole("checkbox", {
      name: `${b.displayText} (b)`,
    });
    await user.click(first);
    await user.click(second);
    expect(screen.getByText(/3 selected \(maximum 3\)/)).toBeTruthy();
    first.focus();
    await user.tab();
    expect(document.activeElement).toBe(second);
    expect(first).toHaveProperty("checked", true);
    expect(second).toHaveProperty("checked", true);
    await user.keyboard(" ");
    expect(first).toHaveProperty("checked", true);
    expect(second).toHaveProperty("checked", false);
    expect(onApply).not.toHaveBeenCalled();
    expect(reader.page).toHaveBeenCalledTimes(1);
    await user.click(
      screen.getByRole("checkbox", { name: "Party source (source)" }),
    );
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([
      expect.objectContaining({
        recordId: "a",
        displayText: a.displayText,
        viewSchemaId: view,
      }),
    ]);
    expect(onCancel).not.toHaveBeenCalled();
  });
  it("stages timestamp filters only after local correction and retains selected Party identity through discovery", async () => {
    const user = userEvent.setup();
    const { props, reader, onApply } = setup();
    vi.mocked(reader.page)
      .mockResolvedValueOnce(page(1, null))
      .mockResolvedValue({
        kind: "accepted",
        value: { candidates: [], nextCursor: null, hasMore: false },
      });
    render(<WorkbookAuthoringReferencePicker {...props} />);
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
    select("1-0");
    await user.click(screen.getByText("Parties ordering and filters"));
    await user.selectOptions(
      screen.getByLabelText("Parties filter field"),
      "party.updated_at",
    );
    const value = screen.getByLabelText(
      "Parties filter value",
    ) as HTMLInputElement;
    await user.type(value, " tomorrow ");
    await user.click(screen.getByRole("button", { name: "Add filter" }));
    expect(value.value).toBe(" tomorrow ");
    expect(value.getAttribute("aria-invalid")).toBe("true");
    expect(
      document.getElementById(value.getAttribute("aria-describedby") ?? "")
        ?.textContent,
    ).toContain("numeric timezone offset");
    expect(reader.page).toHaveBeenCalledTimes(1);
    await user.clear(value);
    await user.type(value, "2026-04-18T00:00:00.000000001Z");
    expect(reader.page).toHaveBeenCalledTimes(1);
    await user.click(screen.getByRole("button", { name: "Add filter" }));
    expect(reader.page).toHaveBeenCalledTimes(1);
    await user.click(
      screen.getByRole("button", { name: "Apply candidate query" }),
    );
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(2));
    expect(
      vi.mocked(reader.page).mock.calls[1]?.[0].queryState.filters,
    ).toEqual([
      {
        fieldKey: "party.updated_at",
        op: "eq",
        arg: { value: "2026-04-18T00:00:00.000000001Z" },
      },
    ]);
    await waitFor(() =>
      expect(
        screen.queryByRole("checkbox", { name: candidateName("Party 1-0") }),
      ).toBeNull(),
    );
    expect(
      screen.getByRole("button", { name: "Remove selected Parties Party 1-0" }),
    ).toBeTruthy();
    expect(onApply).not.toHaveBeenCalled();
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([
      { recordId: "1-0", displayText: "Party 1-0", viewSchemaId: view },
    ]);
  });
  it("stages boolean reference filters before an explicit typed read", async () => {
    const user = userEvent.setup();
    const { props, reader, onApply } = setup();
    const targetView = "cartulary.view.task_requests.v1";
    const task = {
      recordId: "task-a",
      displayText: "Selected task",
      viewSchemaId: targetView,
    };
    vi.mocked(reader.availableViews).mockResolvedValue({
      kind: "accepted",
      value: [targetView],
    });
    vi.mocked(reader.page)
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { candidates: [task], nextCursor: null, hasMore: false },
      })
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "retryable", message: "Unavailable" },
      })
      .mockResolvedValue({
        kind: "accepted",
        value: { candidates: [], nextCursor: null, hasMore: false },
      });
    render(
      <WorkbookAuthoringReferencePicker
        {...props}
        label="Task references"
        views={[targetView]}
      />,
    );
    await screen.findByRole("checkbox", {
      name: candidateName("Selected task"),
    });
    await user.click(
      screen.getByRole("checkbox", { name: candidateName("Selected task") }),
    );
    await user.click(screen.getByText("Task references ordering and filters"));
    await user.selectOptions(
      screen.getByLabelText("Task references filter field"),
      "task.no_owner",
    );
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Task references filter value" }),
      "false",
    );
    await user.click(screen.getByRole("button", { name: "Add filter" }));
    expect(reader.page).toHaveBeenCalledTimes(1);
    expect(onApply).not.toHaveBeenCalled();
    await user.click(
      screen.getByRole("button", { name: "Apply candidate query" }),
    );
    await screen.findByRole("button", { name: "Retry candidates" });
    expect(
      vi.mocked(reader.page).mock.calls[1]?.[0].queryState.filters,
    ).toEqual([{ fieldKey: "task.no_owner", op: "eq", arg: { value: false } }]);
    await user.click(screen.getByRole("button", { name: "Retry candidates" }));
    await waitFor(() =>
      expect(
        screen.queryByRole("checkbox", {
          name: candidateName("Selected task"),
        }),
      ).toBeNull(),
    );
    expect(
      screen.getByRole("button", {
        name: "Remove selected Task references Selected task",
      }),
    ).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "Apply references" }));
    expect(onApply.mock.calls[0]?.[0]).toEqual([task]);
  });
  it("reconciles a retained member label locally and applies the exact identity without moving focus", async () => {
    const { props, reader, onApply, onCancel } = setup();
    const member = {
      recordId: "10000000-0000-4000-8000-000000000001",
      displayText: "",
      viewSchemaId: "incident_members",
    };
    const gate = readGate();
    vi.mocked(reader.page).mockImplementationOnce(() => gate.promise);
    render(
      <WorkbookAuthoringReferencePicker
        {...props}
        label="Owner"
        views={["incident_members"]}
        multiple={false}
        maximum={1}
        selected={[member]}
      />,
    );
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(1));
    const selector = screen.getByRole("combobox", {
      name: "Owner",
    });
    const cancel = screen.getByRole("button", { name: "Cancel references" });
    cancel.focus();
    expect(onApply).not.toHaveBeenCalled();
    await act(async () =>
      gate.resolve({
        kind: "accepted",
        value: {
          candidates: [{ ...member, displayText: "Review editor" }],
          hasMore: false,
          nextCursor: null,
        },
      }),
    );
    expect(screen.getByRole("combobox", { name: "Owner" })).toBe(selector);
    expect(document.activeElement).toBe(cancel);
    expect(selector).toHaveProperty("value", member.recordId);
    expect(
      screen.getByRole("option", { name: "Review editor" }),
    ).toHaveProperty("selected", true);
    const summary = screen.getByRole("list");
    expect(summary.textContent).toContain("Review editor");
    expect(summary.textContent).not.toContain(member.recordId);
    expect(
      screen
        .getByRole("button", { name: /^Remove selected Owner / })
        .getAttribute("aria-label"),
    ).toBe("Remove selected Owner Review editor");
    expect(onApply).not.toHaveBeenCalled();
    click("Cancel references");
    expect(onCancel).toHaveBeenCalledTimes(1);
    expect(onApply).not.toHaveBeenCalled();
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([
      { ...member, displayText: "Review editor" },
    ]);
    expect(member.displayText).toBe("");
    expect(reader.availableViews).not.toHaveBeenCalled();
  });

  it("clears labels on revision invalidation and recovers selected presentation from a new authorized page", async () => {
    const { props, reader, onApply } = setup();
    const gate = readGate();
    vi.mocked(reader.page)
      .mockResolvedValueOnce(page(1))
      .mockImplementationOnce(() => gate.promise);
    const retained = {
      recordId: "1-0",
      displayText: "Old label",
      viewSchemaId: view,
      rowVersion: 2,
    };
    const rendered = render(
      <WorkbookAuthoringReferencePicker {...props} selected={[retained]} />,
    );
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
    expect(
      screen
        .getByRole("button", { name: /^Remove selected Parties / })
        .getAttribute("aria-label"),
    ).toBe("Remove selected Parties Party 1-0");
    const selector = screen.getByRole("group", {
      name: "Parties",
    });
    rendered.rerender(
      <WorkbookAuthoringReferencePicker
        {...props}
        selected={[retained]}
        revision={1}
      />,
    );
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(2));
    expect(screen.getByRole("list").textContent).not.toContain("Party 1-0");
    expect(
      screen.getByRole("button", { name: "Remove selected Parties 1-0" }),
    ).toBeTruthy();
    expect(onApply).not.toHaveBeenCalled();
    await act(async () =>
      gate.resolve({
        kind: "accepted",
        value: {
          candidates: [{ ...candidate("1-0"), displayText: "Renamed Party" }],
          hasMore: false,
          nextCursor: null,
        },
      }),
    );
    expect(screen.getByRole("group", { name: "Parties" })).toBe(selector);
    expect(
      screen
        .getByRole("button", { name: /^Remove selected Parties / })
        .getAttribute("aria-label"),
    ).toBe("Remove selected Parties Renamed Party");
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([
      { ...retained, displayText: "Renamed Party" },
    ]);
    expect(reader.verify).not.toHaveBeenCalled();
  });

  it("refreshes only selected labels and preserves duplicate off-page identities and metadata", async () => {
    const { props, reader, onApply } = setup();
    const longName = "Duplicate readable name ".repeat(12);
    const items = [
      {
        recordId: "one",
        displayText: "Old one",
        viewSchemaId: "cartulary.view.timeline.v2",
        rowVersion: 2,
      },
      { recordId: "two", displayText: longName, viewSchemaId: view },
    ];
    const accepted = (name: string, nextCursor: string | null = "next") => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          { ...candidate("one"), displayText: name },
          candidate("unselected"),
        ],
        hasMore: nextCursor !== null,
        nextCursor,
      },
    });
    vi.mocked(reader.page)
      .mockResolvedValueOnce(accepted(longName))
      .mockResolvedValueOnce(page(2, null))
      .mockResolvedValueOnce(accepted("Updated Party"))
      .mockResolvedValueOnce(accepted(""))
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { candidates: [], hasMore: false, nextCursor: null },
      });
    render(
      <WorkbookAuthoringReferencePicker
        {...props}
        initialView={view}
        selected={items}
      />,
    );
    await screen.findByRole("checkbox", {
      name: candidateName(longName.trim()),
    });
    const remove = screen.getByRole("button", {
      name: `Remove selected Parties ${longName}(one)`,
    });
    expect(
      screen.getByRole("button", {
        name: `Remove selected Parties ${longName}(two)`,
      }),
    ).toBeTruthy();
    remove.focus();
    click("Next candidates");
    await screen.findByRole("checkbox", { name: candidateName("Party 2-0") });
    expect(document.activeElement).toBe(remove);
    expect(
      screen.queryByRole("checkbox", { name: candidateName(longName.trim()) }),
    ).toBeNull();
    click("Previous candidates");
    await screen.findByRole("checkbox", {
      name: candidateName("Updated Party"),
    });
    expect(document.activeElement).toBe(remove);
    expect(remove.getAttribute("aria-label")).toBe(
      "Remove selected Parties Updated Party",
    );
    click("Refresh candidates");
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(4));
    await waitFor(() =>
      expect(
        screen
          .getByRole("group", { name: "Parties" })
          .querySelector('input[value="one"]')
          ?.getAttribute("aria-label"),
      ).toBe("one (one)"),
    );
    expect(remove.getAttribute("aria-label")).toBe(
      "Remove selected Parties Updated Party",
    );
    click("Refresh candidates");
    await screen.findByText("No candidates match this query.");
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([
      { ...items[0], displayText: "Updated Party" },
      items[1],
    ]);
    fireEvent.click(remove);
    expect(document.activeElement).toBe(
      screen.getByRole("button", {
        name: `Remove selected Parties ${longName.trim()}`,
      }),
    );
    click("Apply references");
    expect(onApply.mock.calls.at(-1)?.[0]).toEqual([items[1]]);
  });

  it("keeps member and record label contexts separate for equal opaque IDs", async () => {
    const { props, onApply } = setup();
    const member = {
      recordId: "1-0",
      displayText: "Retained member",
      viewSchemaId: "incident_members",
    };
    render(
      <WorkbookAuthoringReferencePicker
        {...props}
        initialView={view}
        selected={[member]}
      />,
    );
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([member]);
  });

  it("cannot restore labels from obsolete accepted revisions authorities targets or disposed reads", async () => {
    const { props, reader, onApply } = setup();
    const obsolete = readGate(),
      concealed = readGate(),
      disposed = readGate();
    const accepted = (displayText: string) => ({
      kind: "accepted" as const,
      value: {
        candidates: [{ ...candidate("one"), displayText }],
        hasMore: false,
        nextCursor: null,
      },
    });
    vi.mocked(reader.page)
      .mockImplementationOnce(() => obsolete.promise)
      .mockResolvedValueOnce(accepted("Current label"))
      .mockImplementationOnce(() => concealed.promise)
      .mockResolvedValueOnce(accepted("Restored label"))
      .mockImplementationOnce(() => disposed.promise)
      .mockResolvedValueOnce(accepted("New target label"));
    const failure = vi.fn();
    const boundary = (
      revision: number,
      identity: string,
      canRead = true,
      target = "draft",
    ) => (
      <WorkbookCandidateAuthorityContext.Provider
        value={{ identity, canRead, onAuthorityFailure: failure }}
      >
        <WorkbookAuthoringReferencePicker
          key={target}
          {...props}
          targetKey={target}
          revision={revision}
          selected={[
            {
              recordId: "one",
              displayText: "Protected initial",
              viewSchemaId: view,
            },
          ]}
        />
      </WorkbookCandidateAuthorityContext.Provider>
    );
    const rendered = render(boundary(0, "a"));
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(1));
    rendered.rerender(boundary(1, "a"));
    await screen.findByRole("checkbox", {
      name: candidateName("Current label"),
    });
    await act(async () => obsolete.resolve(accepted("Obsolete label")));
    expect(screen.getByRole("list").textContent).toContain("Current label");
    expect(screen.queryByText("Obsolete label")).toBeNull();
    click("Refresh candidates");
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(3));
    rendered.rerender(boundary(1, "b", false));
    await act(async () => concealed.resolve(accepted("Protected late label")));
    expect(screen.getByRole("list").textContent).toContain(
      "Selected reference",
    );
    expect(screen.getByRole("list").textContent).not.toContain("label");
    rendered.rerender(boundary(1, "b"));
    await screen.findByRole("checkbox", {
      name: candidateName("Restored label"),
    });
    click("Refresh candidates");
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(5));
    rendered.rerender(boundary(1, "b", true, "new-draft"));
    await screen.findByRole("checkbox", {
      name: candidateName("New target label"),
    });
    await act(async () => disposed.resolve(accepted("Disposed label")));
    click("Apply references");
    expect(onApply).toHaveBeenCalledWith([
      { recordId: "one", displayText: "New target label", viewSchemaId: view },
    ]);
    expect(failure).not.toHaveBeenCalled();
  });

  it("keeps keyboard focus on held First Refresh Next and Previous reads and exhausted controls", async () => {
    const { props, reader } = setup();
    const user = userEvent.setup();
    const first = readGate();
    const refresh = readGate();
    const next = readGate();
    const previous = readGate();
    const empty = readGate();
    vi.mocked(reader.page)
      .mockResolvedValueOnce(page(1))
      .mockImplementationOnce(() => first.promise)
      .mockImplementationOnce(() => refresh.promise)
      .mockImplementationOnce(() => next.promise)
      .mockImplementationOnce(() => previous.promise)
      .mockImplementationOnce(() => empty.promise);
    render(<WorkbookAuthoringReferencePicker {...props} />);
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });

    for (const [name, gate, result, count] of [
      ["First candidates", first, page(1), 2],
      ["Refresh candidates", refresh, page(1), 3],
      ["Next candidates", next, page(2, null), 4],
      ["Previous candidates", previous, page(1), 5],
    ] as const) {
      const button = screen.getByRole("button", { name });
      button.focus();
      await user.keyboard("{Enter}");
      expect(reader.page).toHaveBeenCalledTimes(count);
      expect(document.activeElement).toBe(button);
      expect(button.getAttribute("aria-busy")).toBe("true");
      expect(button.getAttribute("aria-disabled")).toBe("true");
      expect(button.hasAttribute("disabled")).toBe(false);
      await user.keyboard("{Enter} ");
      fireEvent.click(button);
      expect(reader.page).toHaveBeenCalledTimes(count);
      await act(async () => gate.resolve(result));
      expect(document.activeElement).toBe(button);
      expect(button.getAttribute("aria-busy")).toBe("false");
      if (name === "Next candidates" || name === "Previous candidates") {
        expect(button.getAttribute("aria-disabled")).toBe("true");
        expect(button.hasAttribute("disabled")).toBe(false);
        await user.tab();
        expect(button.hasAttribute("disabled")).toBe(true);
      }
    }
    const emptyRefresh = screen.getByRole("button", {
      name: "Refresh candidates",
    });
    emptyRefresh.focus();
    await user.keyboard("{Enter}");
    expect(reader.page).toHaveBeenCalledTimes(6);
    await act(async () =>
      empty.resolve({
        kind: "accepted",
        value: { candidates: [], hasMore: false, nextCursor: null },
      }),
    );
    expect(document.activeElement).toBe(emptyRefresh);
    expect(screen.getByText("No candidates match this query.")).toBeTruthy();
  });

  it("keeps Retry connected through failure success and restart while blocking duplicate activation", async () => {
    const { props, reader, onApply } = setup();
    const user = userEvent.setup();
    const repeatedFailure = readGate();
    const success = readGate();
    const restart = readGate();
    const rejected = {
      kind: "rejected" as const,
      failure: { kind: "retryable" as const, message: "Read failed" },
    };
    vi.mocked(reader.page)
      .mockResolvedValueOnce(page(1))
      .mockResolvedValueOnce(rejected)
      .mockImplementationOnce(() => repeatedFailure.promise)
      .mockImplementationOnce(() => success.promise)
      .mockResolvedValueOnce(page(1))
      .mockResolvedValueOnce(rejected)
      .mockImplementationOnce(() => restart.promise);
    render(<WorkbookAuthoringReferencePicker {...props} />);
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
    select("1-0");
    click("Next candidates");
    await screen.findByRole("alert");
    const retry = screen.getByRole("button", { name: "Retry candidates" });
    retry.focus();
    await user.keyboard("{Enter}");
    expect(retry.isConnected).toBe(true);
    expect(retry.getAttribute("aria-busy")).toBe("true");
    expect(retry.getAttribute("aria-disabled")).toBe("true");
    expect(retry.hasAttribute("disabled")).toBe(false);
    await user.keyboard("{Enter} ");
    fireEvent.click(retry);
    expect(reader.page).toHaveBeenCalledTimes(3);
    await act(async () => repeatedFailure.resolve(rejected));
    expect(document.activeElement).toBe(retry);
    expect(retry.getAttribute("aria-disabled")).toBe("false");

    await user.keyboard("{Enter}");
    expect(reader.page).toHaveBeenCalledTimes(4);
    await act(async () => success.resolve(page(2, null)));
    expect(document.activeElement).toBe(retry);
    expect(retry.getAttribute("aria-disabled")).toBe("true");
    expect(retry.hasAttribute("disabled")).toBe(false);
    expect(onApply).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Apply references" }));
    expect(onApply).toHaveBeenCalledTimes(1);
    expect(retry.hasAttribute("disabled")).toBe(true);
    expect(
      screen
        .getByRole("button", { name: "Next candidates" })
        .hasAttribute("disabled"),
    ).toBe(true);

    click("First candidates");
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
    click("Next candidates");
    await screen.findByRole("alert");
    const retryAgain = screen.getByRole("button", {
      name: "Retry candidates",
    });
    retryAgain.focus();
    await user.keyboard("{Enter}");
    await act(async () =>
      restart.resolve({
        kind: "rejected",
        failure: {
          kind: "invalid_contract",
          message: "Restart from First",
        },
      }),
    );
    expect(document.activeElement).toBe(retryAgain);
    expect(retryAgain.getAttribute("aria-disabled")).toBe("true");
    expect(
      screen.getByText("Restart this query with First candidates."),
    ).toBeTruthy();
  });

  it("retires held read focus after outside interaction scope replacement and authority loss", async () => {
    const { props, reader, onApply } = setup();
    const user = userEvent.setup();
    const outside = readGate();
    const oldScope = readGate();
    const oldAuthority = readGate();
    vi.mocked(reader.page)
      .mockResolvedValueOnce(page(1))
      .mockImplementationOnce(() => outside.promise)
      .mockImplementationOnce(() => oldScope.promise)
      .mockResolvedValueOnce(page(2, null))
      .mockImplementationOnce(() => oldAuthority.promise);
    const failure = vi.fn();
    const boundary = (revision: number, identity: string, canRead = true) => (
      <WorkbookCandidateAuthorityContext.Provider
        value={{ identity, canRead, onAuthorityFailure: failure }}
      >
        <WorkbookAuthoringReferencePicker
          {...props}
          revision={revision}
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
    const rendered = render(boundary(0, "account-a"));
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
    const refresh = screen.getByRole("button", { name: "Refresh candidates" });
    refresh.focus();
    await user.keyboard("{Enter}");
    const apply = screen.getByRole("button", { name: "Apply references" });
    apply.focus();
    fireEvent.wheel(document);
    await act(async () => outside.resolve(page(1)));
    expect(document.activeElement).toBe(apply);
    expect(onApply).not.toHaveBeenCalled();

    refresh.focus();
    await user.keyboard("{Enter}");
    rendered.rerender(boundary(1, "account-a"));
    await screen.findByRole("checkbox", { name: candidateName("Party 2-0") });
    const newRefresh = screen.getByRole("button", {
      name: "Refresh candidates",
    });
    newRefresh.focus();
    await act(async () => oldScope.resolve(page(3, null)));
    expect(document.activeElement).toBe(newRefresh);
    expect(
      screen.queryByRole("checkbox", { name: candidateName("Party 3-0") }),
    ).toBeNull();

    await user.keyboard("{Enter}");
    rendered.rerender(boundary(1, "account-b", false));
    await act(async () =>
      oldAuthority.resolve({
        kind: "rejected",
        failure: { kind: "authentication_required", message: "Expired" },
      }),
    );
    expect(
      screen.queryByRole("button", { name: "Refresh candidates" }),
    ).toBeNull();
    expect(screen.queryByText("Protected label")).toBeNull();
    expect(failure).not.toHaveBeenCalled();
    expect(onApply).not.toHaveBeenCalled();
  });
  it("restores focused selected-reference removal to the next surviving identity", () => {
    function Selection() {
      const [selected, setSelected] = useState([
        { recordId: "one", displayText: "Duplicate" },
        { recordId: "two", displayText: "Duplicate" },
      ]);
      return (
        <WorkbookCandidateSelection
          candidates={[]}
          selected={selected}
          label="Linked Records"
          testId="focus-candidates"
          multiple
          maximum={64}
          disabled={false}
          onChange={(items) => setSelected([...items])}
        />
      );
    }
    render(<Selection />);
    const buttons = screen.getAllByRole("button", {
      name: /^Remove selected Linked Records Duplicate/,
    });
    const first = buttons[0];
    if (!first) throw new Error("Missing first Remove control");
    first.focus();
    fireEvent.click(first);
    expect(document.activeElement).toBe(buttons[1]);
  });
  it("uses prior identity order for first middle last and sole removals", () => {
    const values = ["a", "b", "c", "d"].map((recordId) => ({
      recordId,
      displayText: `Reference ${recordId}`,
    }));
    for (const [removed, expected] of [
      ["a", "b"],
      ["b", "c"],
      ["d", "c"],
      ["a", "selector"],
    ] as const) {
      const initial = expected === "selector" ? values.slice(0, 1) : values;
      function Selection() {
        const [selected, setSelected] = useState(initial);
        return (
          <WorkbookCandidateSelection
            candidates={[{ recordId: "page", displayText: "Current page" }]}
            selected={selected}
            label="Records"
            testId="records-selector"
            multiple
            maximum={64}
            disabled={false}
            onChange={(items) => setSelected([...items])}
          />
        );
      }
      render(<Selection />);
      const source = screen.getByRole("button", {
        name: `Remove selected Records Reference ${removed}`,
      });
      source.focus();
      fireEvent.click(source);
      expect(document.activeElement).toBe(
        expected === "selector"
          ? screen.getByRole("checkbox", {
              name: candidateName("Current page"),
            })
          : screen.getByRole("button", {
              name: `Remove selected Records Reference ${expected}`,
            }),
      );
      cleanup();
    }
  });
  it("uses an accessible field fallback when the sole candidate selector is disabled", () => {
    function Selection() {
      const [selected, setSelected] = useState([
        { recordId: "off-page", displayText: "Retained" },
      ]);
      return (
        <WorkbookCandidateSelection
          candidates={[]}
          selected={selected}
          label="Records"
          testId="empty-records"
          multiple
          maximum={64}
          disabled={false}
          onChange={(items) => setSelected([...items])}
        />
      );
    }
    render(<Selection />);
    const remove = screen.getByRole("button", {
      name: "Remove selected Records Retained",
    });
    remove.focus();
    fireEvent.click(remove);
    expect(screen.getByRole("group", { name: "Records" })).toHaveProperty(
      "disabled",
      true,
    );
    expect(document.activeElement).toBe(
      screen.getByRole("group", { name: "Records selected references" }),
    );
  });
  it("waits for a controlled parent and retires obsolete or external focus intentions", () => {
    const first = { recordId: "first", displayText: "First", rowVersion: 4 };
    const second = { recordId: "second", displayText: "Second", rowVersion: 9 };
    const onChange = vi.fn();
    const external = render(<button type="button">Elsewhere</button>).getByRole(
      "button",
    );
    const selection = (
      selected: readonly (typeof first)[],
      scopeKey = "field-a",
      candidates: readonly { recordId: string; displayText: string }[] = [],
      disabled = false,
    ) => (
      <WorkbookCandidateSelection
        candidates={candidates}
        selected={selected}
        label="Records"
        testId="controlled-records"
        scopeKey={scopeKey}
        multiple
        maximum={64}
        disabled={disabled}
        onChange={onChange}
      />
    );
    const rendered = render(selection([first, second]));
    const source = screen.getByRole("button", {
      name: "Remove selected Records First",
    });
    source.focus();
    fireEvent.click(source);
    expect(onChange).toHaveBeenCalledWith([second]);
    expect(document.activeElement).toBe(source);
    rendered.rerender(
      selection([first, second], "field-a", [
        { recordId: "page", displayText: "Page" },
      ]),
    );
    expect(document.activeElement).toBe(source);
    rendered.rerender(selection([second]));
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Remove selected Records Second" }),
    );
    rendered.rerender(
      selection([second], "field-a", [
        { recordId: "new-page", displayText: "New page" },
      ]),
    );
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Remove selected Records Second" }),
    );

    rendered.rerender(selection([first, second]));
    const again = screen.getByRole("button", {
      name: "Remove selected Records First",
    });
    again.focus();
    fireEvent.click(again);
    external.focus();
    rendered.rerender(selection([second]));
    expect(document.activeElement).toBe(external);

    rendered.rerender(selection([first, second]));
    const rejected = screen.getByRole("button", {
      name: "Remove selected Records First",
    });
    rejected.focus();
    fireEvent.click(rejected);
    rendered.rerender(selection([first, second], "field-b"));
    expect(document.activeElement).toBe(rejected);
    rendered.rerender(selection([second], "field-b"));
    expect(document.activeElement).not.toBe(
      screen.getByRole("button", { name: "Remove selected Records Second" }),
    );
    rendered.rerender(selection([first, second], "field-c"));
    const beforeDisable = screen.getByRole("button", {
      name: "Remove selected Records First",
    });
    beforeDisable.focus();
    fireEvent.click(beforeDisable);
    rendered.rerender(selection([first, second], "field-c", [], true));
    rendered.rerender(selection([second], "field-c"));
    expect(document.activeElement).not.toBe(
      screen.getByRole("button", { name: "Remove selected Records Second" }),
    );
  });
  it("keeps parent-list removal local without opening discovery or changing survivor metadata", () => {
    const { props, reader } = setup();
    const retained: WorkbookAuthoringSelection[] = [
      {
        recordId: "off-page-a",
        displayText: "First",
        viewSchemaId: view,
        rowVersion: 4,
      },
      {
        recordId: "off-page-b",
        displayText: "Last",
        viewSchemaId: view,
        rowVersion: 9,
      },
    ];
    const applied = vi.fn();
    function Control() {
      const [selected, setSelected] = useState(retained);
      return (
        <WorkbookAuthoringReferenceControl
          {...props}
          selected={selected}
          onApply={(items) => {
            applied(items);
            setSelected([...items]);
          }}
        />
      );
    }
    render(<Control />);
    const last = screen.getByRole("button", { name: "Remove Parties Last" });
    last.focus();
    fireEvent.click(last);
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Remove Parties First" }),
    );
    expect(applied).toHaveBeenCalledWith([retained[0]]);
    const first = screen.getByRole("button", { name: "Remove Parties First" });
    first.focus();
    fireEvent.click(first);
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Choose parties" }),
    );
    expect(reader.page).not.toHaveBeenCalled();
    expect(reader.availableViews).not.toHaveBeenCalled();
  });
  it("retains selected identities across twelve evicted pages with bounded options and reduced payloads", async () => {
    const { props, reader, onApply } = setup();
    render(<WorkbookAuthoringReferencePicker {...props} />);
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
    select("1-0");
    for (let number = 2; number <= 13; number++) {
      click("Next candidates");
      await screen.findByRole("checkbox", {
        name: candidateName(`Party ${number}-0`),
      });
      expect(
        Array.from(
          screen
            .getByRole("group", { name: "Parties" })
            .querySelectorAll('input[type="checkbox"]'),
        ),
      ).toHaveLength(100);
    }
    select("13-0");
    expect(
      screen.queryByRole("checkbox", { name: candidateName("Party 1-0") }),
    ).toBeNull();
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
    const readsBeforeRemoval = vi.mocked(reader.page).mock.calls.length;
    click("Remove selected Parties Party 1-0");
    expect(reader.page).toHaveBeenCalledTimes(readsBeforeRemoval);
    expect(onApply).toHaveBeenCalledTimes(1);
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
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
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
    await screen.findByRole("checkbox", { name: candidateName("Party 2-0") });
    expect(vi.mocked(reader.page).mock.calls[2]?.[0].cursor).toBe("page-2");
    expect(vi.mocked(reader.page).mock.calls[2]?.[0].queryState).toEqual(
      vi.mocked(reader.page).mock.calls[1]?.[0].queryState,
    );
  });
  it("applies query edits explicitly without clearing selection and rejects over-limit changes", async () => {
    const { props, reader, onApply } = setup();
    render(<WorkbookAuthoringReferencePicker {...props} maximum={1} />);
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
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
    select();
    expect(screen.queryByRole("alert")).toBeNull();
    select("1-1");
    click("Apply references");
    expect(
      onApply.mock.calls
        .at(-1)?.[0]
        .map((item: WorkbookAuthoringSelection) => item.recordId),
    ).toEqual(["1-1"]);
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
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
    select("1-0");
    fireEvent.keyDown(screen.getByRole("group", { name: "Parties" }), {
      key: "Escape",
    });
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Choose parties" }),
    );
    expect(onApply).not.toHaveBeenCalled();
    click("Choose parties");
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
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
    await screen.findByRole("checkbox", { name: candidateName("Party 2-0") });
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
    expect(
      screen.queryByRole("checkbox", { name: candidateName("Party 2-0") }),
    ).toBeNull();
    expect(
      screen
        .getByRole("button", { name: "Apply references" })
        .hasAttribute("disabled"),
    ).toBe(true);
    rendered.rerender(boundary("current", "session-a-restored"));
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
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
    await screen.findByRole("checkbox", { name: candidateName("Party 1-0") });
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
    const { props, onApply, onCancel } = setup();
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
    expect(
      screen
        .getByRole("button", { name: /^Remove selected Parties / })
        .getAttribute("aria-label"),
    ).toBe("Remove selected Parties Reviewed source");
    expect(onApply).not.toHaveBeenCalled();
    fireEvent.keyDown(screen.getByRole("combobox", { name: "Parties" }), {
      key: "Escape",
    });
    expect(onCancel).toHaveBeenCalledTimes(1);
    expect(onApply).not.toHaveBeenCalled();
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
