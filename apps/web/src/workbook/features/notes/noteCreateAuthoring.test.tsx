import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  renderHook,
  screen,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import { useGenericCreateDraft } from "../generic/useGenericCreateDraft";
import { NoteCreateContext } from "./NoteCreateContext";
import { NoteCreateForm } from "./NoteCreateForm";
import { NoteSourceControl } from "./NoteSourceControl";
import {
  type NoteCreateReader,
  noteCreateView,
  noteFeature,
  noteSourceViews,
  prepareNote,
} from "./noteCreateModel";
import { useNoteCreateAttachment } from "./useNoteCreateAttachment";
import { WorkbookNoteCreateOwner } from "./WorkbookNoteCreateOwner";

afterEach(cleanup);
const authority: WorkbookMutationAuthority = {
  actorId: "actor",
  incidentId: "incident",
  sessionIdentity: "session",
  role: "editor",
  closed: false,
};
const token = Symbol("test");
function fixture(view = noteSourceViews[0]) {
  const owner = new WorkbookNoteCreateOwner("incident", {
    create: () => "txn",
  });
  owner.setAuthority(authority);
  const reader: NoteCreateReader = {
    verifyNote: vi.fn(async () => {}),
    availableViews: vi.fn(async () => ({
      kind: "accepted" as const,
      value: noteSourceViews,
    })),
    page: vi.fn(async () => ({
      kind: "accepted" as const,
      value: { candidates: [], hasMore: false, nextCursor: null },
    })),
  };
  owner.configure(reader, async () => authority);
  const subject = {
    subject: {
      kind: "live" as const,
      recordId: "source",
      viewSchemaId: view,
      rowVersion: 1,
      label: "Original source",
      surfaceLabel: requireViewContract(view).title,
    },
    cells: {},
  };
  const feature = noteFeature(view);
  if (!feature) throw new Error("Expected Note feature");
  expect(
    owner.begin(subject, feature, { kind: "view_schema", id: view }, token),
  ).toBe(true);
  return { owner, reader, subject, feature };
}
describe("Note authoring", () => {
  it("detaches inspector presentations without retargeting and shares retained fields with the Notes grid", () => {
    const { owner, subject, feature } = fixture();
    owner.discard();
    let sheet = {
      kind: "view_schema" as const,
      id: noteSourceViews[0] as string,
    };
    const hook = renderHook(
      ({ source, open }) => ({
        attachment: useNoteCreateAttachment(source, open),
        grid: useGenericCreateDraft(
          requireViewContract(noteCreateView),
          authority.actorId,
        ),
      }),
      {
        initialProps: { source: subject, open: true },
        wrapper: ({ children }) => (
          <NoteCreateContext.Provider value={{ owner, sheetRef: sheet }}>
            {children}
          </NoteCreateContext.Provider>
        ),
      },
    );
    act(() => {
      hook.result.current.attachment.begin(feature);
      owner.update("note.body", "unfinished");
    });
    hook.rerender({
      source: { ...subject, subject: { ...subject.subject, rowVersion: 2 } },
      open: true,
    });
    expect(hook.result.current.attachment.workflow).not.toBeNull();
    expect(owner.getSnapshot().draft?.source?.rowVersion).toBe(1);
    hook.rerender({
      source: {
        ...subject,
        subject: { ...subject.subject, recordId: "other" },
      },
      open: true,
    });
    expect(hook.result.current.attachment.workflow).toBeNull();
    expect(owner.getSnapshot().draft?.source?.recordId).toBe("source");
    sheet = { kind: "view_schema", id: noteCreateView };
    hook.rerender({ source: subject, open: false });
    expect(hook.result.current.grid[0]).toEqual({ "note.body": "unfinished" });
    act(() =>
      hook.result.current.grid[1]((draft) => ({ ...draft, "note.title": "" })),
    );
    expect(owner.getSnapshot().draft?.values).toEqual({
      "note.title": "",
      "note.body": "unfinished",
    });
    hook.unmount();
    expect(owner.getSnapshot().draft?.source?.recordId).toBe("source");
  });
  it("captures all four contextual sources and preserves the ordinary sheet draft boundary", () => {
    for (const view of noteSourceViews) {
      const { owner } = fixture(view);
      expect(owner.getSnapshot().draft?.source).toMatchObject({
        recordId: "source",
        viewSchemaId: view,
        rowVersion: 1,
      });
      expect(owner.getSnapshot().draft?.values).toEqual({});
    }
    const owner = new WorkbookNoteCreateOwner("incident", {
      create: () => "txn",
    });
    owner.setAuthority(authority);
    expect(
      owner.beginSheet({ kind: "view_schema", id: noteCreateView }, token),
    ).toBe(true);
    expect(owner.getSnapshot().draft?.source).toBeNull();
  });
  it("retains raw fields and source across versions, detach, source replacement, and explicit clearing", () => {
    const { owner, subject, feature } = fixture();
    owner.update("note.title", "Draft");
    owner.update("note.body", "unfinished\nbody");
    owner.update("note.tags", "local");
    owner.observe("source", 2);
    owner.detach(token);
    expect(owner.getSnapshot().draft?.source?.rowVersion).toBe(1);
    expect(
      owner.begin(
        { ...subject, subject: { ...subject.subject, recordId: "other" } },
        feature,
        { kind: "view_schema", id: noteSourceViews[0] },
        Symbol(),
      ),
    ).toBe(false);
    owner.resume(token);
    owner.changeSource({
      recordId: "replacement",
      viewSchemaId: noteSourceViews[1],
      rowVersion: 4,
      label: "Replacement",
    });
    expect(owner.getSnapshot().draft?.source?.recordId).toBe("replacement");
    owner.changeSource(null);
    expect(owner.getSnapshot().draft?.values).toEqual({
      "note.title": "Draft",
      "note.body": "unfinished\nbody",
      "note.tags": "local",
    });
    expect(owner.getSnapshot().draft?.source).toBeNull();
    owner.discard();
    expect(
      owner.beginSheet({ kind: "view_schema", id: noteCreateView }, token),
    ).toBe(true);
  });
  it("normalizes title body and tags while preserving omission and explicit clearing", () => {
    const { owner } = fixture();
    owner.update("note.title", "\u2003Cafe\u0301\u2003");
    let draft = required(owner.getSnapshot().draft);
    expect(prepareNote(draft, "txn").request).toEqual({
      client_txn_id: "txn",
      "note.title": "Café",
    });
    owner.update("note.title", "");
    owner.update("note.body", " \r\nBody\rtext\t \u2003");
    owner.update("note.tags", " e\u0301 \n second ");
    draft = required(owner.getSnapshot().draft);
    expect(prepareNote(draft, "txn").request).toEqual({
      client_txn_id: "txn",
      "note.title": "",
      "note.body": "Body\ntext",
      "note.tags": {
        kind: "collection_actions_v1",
        actions: [
          { op: "add_tag", tag_name: "é" },
          { op: "add_tag", tag_name: "second" },
        ],
      },
    });
    owner.update("note.tags", "");
    expect(
      prepareNote(required(owner.getSnapshot().draft), "txn").request,
    ).not.toHaveProperty("note.tags");
    expect(owner.getSnapshot().draft?.values["note.body"]).toContain("\r");
  });
  it("rejects source-only tags-only whitespace controls and scalar overflow without destroying raw input", () => {
    const { owner } = fixture();
    expect(owner.validate()).toBe(false);
    owner.update("note.tags", "tag");
    owner.update("note.body", "\u2003 \n");
    expect(owner.validate()).toBe(false);
    for (const raw of [
      "title\u0000",
      "\tTitle",
      "x".repeat(513),
      "😀".repeat(513),
    ]) {
      owner.update("note.title", raw);
      expect(owner.validate()).toBe(false);
      expect(owner.getSnapshot().draft?.values["note.title"]).toBe(raw);
    }
    owner.update("note.title", "😀".repeat(512));
    expect(owner.validate()).toBe(true);
    owner.update("note.body", "x".repeat(16385));
    expect(owner.validate()).toBe(false);
    owner.update("note.body", "ok");
    owner.update("note.tags", "x".repeat(65));
    expect(owner.validate()).toBe(false);
  });
  it("conceals retained work on suspension and retires it on account replacement", () => {
    const { owner } = fixture();
    owner.update("note.title", "private");
    owner.suspend();
    expect(owner.getSnapshot().draft).toBeNull();
    owner.update("note.title", "obsolete callback");
    owner.setAuthority({ ...authority, sessionIdentity: "renewed" });
    expect(owner.getSnapshot().draft?.values["note.title"]).toBe("private");
    owner.closeIncident();
    expect(owner.canSubmit()).toBe(false);
    owner.setAuthority(authority);
    owner.setAuthority({ ...authority, actorId: "replacement" });
    expect(owner.getSnapshot().draft).toBeNull();
  });
  it("exposes invalid input accessibly in the shared Note form", () => {
    const { owner } = fixture();
    render(
      <NoteCreateForm
        owner={owner}
        attachment={token}
        onSubmit={() => owner.validate()}
      />,
    );
    fireEvent.change(screen.getByLabelText("Title"), {
      target: { value: "invalid\u0001" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create Note" }));
    expect(screen.getByLabelText("Title").getAttribute("aria-invalid")).toBe(
      "true",
    );
    expect(screen.getByRole("alert").textContent).toContain("control");
    fireEvent.click(screen.getByRole("button", { name: "Close draft" }));
    expect(owner.getSnapshot().draft?.values["note.title"]).toBe(
      "invalid\u0001",
    );
  });
  it("stages paged source selection, retains off-page identity, and restores focus on cancel", async () => {
    const { owner, reader } = fixture();
    vi.mocked(reader.page).mockImplementation(async (input) => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          {
            recordId: input.cursor ? "second" : "first",
            displayText: input.cursor ? "Second" : "First",
            viewSchemaId: input.viewSchemaId,
            row: {
              record_id: input.cursor ? "second" : "first",
              row_version: 3,
              cells: {},
            },
          },
        ],
        hasMore: !input.cursor,
        nextCursor: input.cursor ? null : "page2",
      },
    }));
    const onChange = vi.fn();
    const picker = render(
      <NoteSourceControl
        targetKey={"note-test"}
        source={required(owner.getSnapshot().draft).source}
        reader={reader}
        revision={0}
        disabled={false}
        onChange={onChange}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Choose source" }));
    await screen.findByRole("option", { name: "First" });
    expect(
      screen.getByRole("button", {
        name: "Remove selected Note source Original source",
      }),
    ).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Next candidates" }));
    await screen.findByRole("option", { name: "Second" });
    fireEvent.change(screen.getByLabelText("Note source"), {
      target: { value: "second" },
    });
    fireEvent.change(screen.getByLabelText("Source sheet"), {
      target: { value: noteSourceViews[1] },
    });
    picker.rerender(
      <NoteSourceControl
        targetKey={"note-test"}
        source={required(owner.getSnapshot().draft).source}
        reader={reader}
        revision={1}
        disabled={false}
        onChange={onChange}
      />,
    );
    await screen.findByRole("option", { name: "First" });
    expect(
      (screen.getByLabelText("Source sheet") as HTMLSelectElement).value,
    ).toBe(noteSourceViews[1]);
    expect(onChange).not.toHaveBeenCalled();
    fireEvent.keyDown(screen.getByLabelText("Note source"), { key: "Escape" });
    expect(onChange).not.toHaveBeenCalled();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Choose source" }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Choose source" }));
    await screen.findByRole("option", { name: "First" });
    fireEvent.change(screen.getByLabelText("Note source"), {
      target: { value: "first" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Apply source" }));
    expect(onChange).toHaveBeenCalledWith({
      recordId: "first",
      viewSchemaId: noteSourceViews[0],
      rowVersion: 3,
      label: "First",
    });
  });
  it("distinguishes discovery failures from successful empty pages and supports retry", async () => {
    const { reader } = fixture();
    vi.mocked(reader.availableViews).mockRejectedValueOnce(
      new Error("offline"),
    );
    render(
      <NoteSourceControl
        targetKey={"note-test"}
        source={null}
        reader={reader}
        revision={0}
        disabled={false}
        onChange={() => {}}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Choose source" }));
    await screen.findByRole("button", { name: "Retry surfaces" });
    expect(screen.queryByText("No available source sheets.")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Retry surfaces" }));
    await screen.findByText("No candidates match this query.");
    expect(reader.availableViews).toHaveBeenCalledTimes(2);
  });
});

function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Required fixture value");
  return value;
}
