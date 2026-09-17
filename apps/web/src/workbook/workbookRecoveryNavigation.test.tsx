import type { GridHandle } from "@cartulary/grid-adapter";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import { useLayoutEffect, useRef } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { WorkbookRecoveryDetail } from "../shared/WorkbookRecoveryBoundary";
import { WorkbookRecoveryNavigation } from "../shared/workbookRecoveryNavigation";
import { WorkbookRecoveryFixture } from "../testing/WorkbookRecoveryFixture";
import { WorkbookSameFieldConflictResolver } from "./components/WorkbookSameFieldConflictResolver";
import { CoordinationCreateRecovery } from "./features/coordination/CoordinationCreateRecovery";
import { coordinationFeature } from "./features/coordination/coordinationCreateModel";
import { NoteCreateRecovery } from "./features/notes/NoteCreateRecovery";
import {
  useWorkbookBrowsingRegistry,
  WorkbookQueryBrowsingProvider,
} from "./query/WorkbookQueryBrowsingContext";
import { WorkbookMutationRuntime } from "./runtime/WorkbookMutationRuntime";
import { workbookConflictEntry } from "./runtime/workbookConflictModel";

afterEach(cleanup);

it("restores the resolved cell only for the current recovery activation", async () => {
  for (const continuation of [
    "current",
    "closed",
    "retargeted",
    "newer-focus",
  ] as const) {
    const { runtime } = fixture();
    const navigation = new WorkbookRecoveryNavigation();
    const source = navigation.register("core", {});
    const view = "cartulary.view.timeline.v2";
    const conflict = workbookConflictEntry({
      viewSchemaId: view,
      surfaceLabel: "Timeline",
      rowLabel: "Original row",
      conflict: {
        record_id: "row-1",
        field_key: "timeline.activity_synopsis_text",
        base_row_version: 1,
        current_row_version: 2,
        conflict_token: "token",
        conflict_resolution_class: "text_compare_merge",
        base_value: "base",
        client_value: "local",
        server_value: "saved",
      },
    });
    source.update(
      ["original", "other"].map((id, order) => ({
        id,
        order,
        label: id,
        origin: "Timeline",
        summary: "Needs review",
        sheetRef: { kind: "view_schema", id: view },
        attention: "attention",
      })),
    );
    let complete!: (result: string | null) => void;
    vi.spyOn(runtime, "resolveConflict").mockImplementation(
      () =>
        new Promise((resolve) => {
          complete = resolve;
        }),
    );
    const gridRef: { current: GridHandle | null } = { current: null };
    const requestFocus = vi.fn<GridHandle["requestFocus"]>(async () => {
      const cell = screen.getByRole("button", { name: "Original cell" });
      const viewport = cell.parentElement;
      if (!viewport) throw new Error("Missing viewport");
      viewport.scrollTop = 100;
      viewport.scrollLeft = 300;
      // Selecting a cell republishes the handle while retaining the viewport.
      if (gridRef.current) gridRef.current = { ...gridRef.current };
      cell.focus();
      return "focused";
    });
    function Content() {
      const registry = useWorkbookBrowsingRegistry();
      const root = useRef<HTMLDivElement>(null);
      const summary = useRef<HTMLDivElement>(null);
      useLayoutEffect(() => {
        gridRef.current = {
          requestFocus,
          getScrollElement: () => root.current,
          activateEdit: () => false,
          cancelEdit: () => false,
          getAnchorRect: () => null,
          isAnchorRendered: () => false,
          moveFocus: () => null,
          planPasteTargets: () => null,
          scrollToAnchor: () => false,
        };
        return registry?.bindGrid(view, gridRef);
      }, [registry]);
      return (
        <>
          <div ref={root}>
            <button type="button">Original cell</button>
          </div>
          <button type="button">Newer work</button>
          <WorkbookRecoveryDetail source="core" item="original">
            <WorkbookSameFieldConflictResolver
              onClose={navigation.close}
              mutationRuntime={runtime}
              onActivateOrigin={() => {}}
              snapshot={{ conflicts: [conflict] }}
              summaryRef={summary}
            />
          </WorkbookRecoveryDetail>
        </>
      );
    }
    const rendered = render(
      <WorkbookQueryBrowsingProvider>
        <WorkbookRecoveryFixture navigation={navigation}>
          <Content />
        </WorkbookRecoveryFixture>
      </WorkbookQueryBrowsingProvider>,
    );
    const viewport = screen.getByRole("button", {
      name: "Original cell",
    }).parentElement;
    if (!viewport) throw new Error("Missing viewport");
    viewport.scrollTop = 42;
    viewport.scrollLeft = 80;
    fireEvent.click(screen.getByRole("button", { name: "Recovery (2)" }));
    fireEvent.click(
      screen.getByRole("button", { name: "original · Timeline" }),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Discard local draft" }),
    );
    if (continuation === "closed") {
      fireEvent.click(screen.getByRole("button", { name: "Close recovery" }));
      screen.getByRole("button", { name: "Newer work" }).focus();
    } else if (continuation === "retargeted") {
      fireEvent.click(screen.getByRole("button", { name: "All recovery" }));
      fireEvent.click(screen.getByRole("button", { name: "other · Timeline" }));
    } else if (continuation === "newer-focus") {
      screen.getByRole("button", { name: "Newer work" }).focus();
    }
    if (continuation === "current") {
      // The runtime may retire the conflict before the awaited submit returns.
      await act(async () =>
        source.update(
          navigation
            .getSnapshot()
            .entries.filter((entry) => entry.id === "other"),
        ),
      );
    }
    const previousFocus = document.activeElement;
    await act(async () => {
      complete(null);
    });
    if (continuation === "current") {
      expect(requestFocus).toHaveBeenCalledWith({
        kind: "cell",
        anchor: {
          surface: { kind: "view_schema", viewSchemaId: view },
          rowIdentity: { kind: "core_record", recordId: "row-1" },
          fieldKey: "timeline.activity_synopsis_text",
        },
      });
      expect(navigation.getSnapshot().open).toBe(false);
      expect(document.activeElement).toBe(
        screen.getByRole("button", { name: "Original cell" }),
      );
    } else {
      expect(requestFocus).not.toHaveBeenCalled();
      expect(document.activeElement).toBe(previousFocus);
      expect(navigation.getSnapshot().open).toBe(continuation !== "closed");
    }
    expect({ top: viewport.scrollTop, left: viewport.scrollLeft }).toEqual({
      top: 42,
      left: 80,
    });
    rendered.unmount();
    navigation.dispose();
  }
});

function fixture() {
  const incidentId = "40000000-0000-4000-8000-000000000004";
  const runtime = new WorkbookMutationRuntime(
    { incidentId, clientInstanceId: "tab" },
    { create: () => crypto.randomUUID() },
    { execute: vi.fn() },
  );
  const authority = {
    incidentId,
    actorId: "10000000-0000-4000-8000-000000000001",
    sessionIdentity: "session",
    role: "editor" as const,
    closed: false,
  };
  const sheet = {
    kind: "view_schema" as const,
    id: "cartulary.view.timeline.v2",
  };
  const token = Symbol("original authoring");
  runtime.noteCreate.setAuthority(authority);
  runtime.coordinationCreate.setAuthority(authority);
  runtime.noteCreate.beginSheet(
    { kind: "view_schema", id: "cartulary.view.notes.v1" },
    token,
  );
  runtime.noteCreate.update("note.title", "Retained note text");
  const feature = coordinationFeature(sheet.id, "lesson");
  if (!feature) throw new Error("Missing lesson feature");
  runtime.coordinationCreate.begin(
    {
      subject: {
        kind: "live",
        viewSchemaId: sheet.id,
        recordId: "20000000-0000-4000-8000-000000000002",
        rowVersion: 1,
        label: "Original Timeline record",
        surfaceLabel: "Timeline",
      },
      cells: {},
    },
    feature,
    sheet,
    token,
  );
  runtime.coordinationCreate.update("lesson.summary", "Retained lesson text");
  runtime.noteCreate.detach(token);
  runtime.coordinationCreate.detach(token);
  return { runtime, authority };
}

it("coordinates independent recovery owners without losing retained authoring", () => {
  const { runtime } = fixture();
  render(
    <WorkbookRecoveryFixture>
      <NoteCreateRecovery owner={runtime.noteCreate} />
      <CoordinationCreateRecovery owner={runtime.coordinationCreate} />
    </WorkbookRecoveryFixture>,
  );
  fireEvent.click(screen.getByRole("button", { name: "Recovery (2)" }));
  fireEvent.click(screen.getByRole("button", { name: /Note draft ·/ }));
  expect(
    screen.getByRole("region", { name: "Retained Note authoring" }),
  ).toBeTruthy();
  expect(
    screen.queryByRole("region", { name: "Retained Coordination authoring" }),
  ).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Resume Note draft" }));
  fireEvent.click(screen.getByRole("button", { name: "All recovery" }));
  fireEvent.click(screen.getByRole("button", { name: /Coordination draft ·/ }));
  expect(
    screen.getByRole("region", { name: "Retained Coordination authoring" }),
  ).toBeTruthy();
  expect(
    screen.queryByRole("region", { name: "Retained Note authoring" }),
  ).toBeNull();
  fireEvent.click(screen.getByRole("button", { name: "Close recovery" }));
  expect(screen.getByRole("button", { name: "Recovery (2)" })).toBe(
    document.activeElement,
  );
  fireEvent.click(screen.getByRole("button", { name: "Recovery (2)" }));
  fireEvent.click(screen.getByRole("button", { name: /Note draft ·/ }));
  expect(runtime.noteCreate.getSnapshot().draft?.values["note.title"]).toBe(
    "Retained note text",
  );
  expect(
    runtime.coordinationCreate.getSnapshot().draft?.values["lesson.summary"],
  ).toBe("Retained lesson text");
  expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
  act(() => runtime.invalidate({ kind: "runtime_disposed" }));
});

it("conceals revoked owner summaries and keeps focus in the recovery list", () => {
  const { runtime, authority } = fixture();
  render(
    <WorkbookRecoveryFixture>
      <NoteCreateRecovery owner={runtime.noteCreate} />
    </WorkbookRecoveryFixture>,
  );
  fireEvent.click(screen.getByRole("button", { name: "Recovery (1)" }));
  fireEvent.click(screen.getByRole("button", { name: /Note draft ·/ }));
  const resume = screen.getByRole("button", { name: "Resume Note draft" });
  resume.focus();
  // Browsers blur synchronously when an action disables its own button.
  resume.setAttribute("disabled", "");
  fireEvent.blur(resume, { relatedTarget: document.body });
  act(() => runtime.noteCreate.suspend());
  expect(
    screen.queryByRole("region", { name: "Retained Note authoring" }),
  ).toBeNull();
  expect(screen.getByRole("button", { name: "Recovery (0)" })).toBeTruthy();
  expect(
    within(screen.getByRole("region", { name: "Recovery navigation" }))
      .getAllByRole("heading", { level: 2 })
      .find((heading) => heading.tabIndex === -1),
  ).toBe(document.activeElement);
  expect(runtime.noteCreate.getSnapshot().draft).toBeNull();
  act(() => runtime.noteCreate.setAuthority(authority));
  expect(runtime.noteCreate.getSnapshot().draft?.values["note.title"]).toBe(
    "Retained note text",
  );
  expect(
    screen.queryByRole("region", { name: "Retained Note authoring" }),
  ).toBeNull();
  act(() => runtime.invalidate({ kind: "runtime_disposed" }));
});
