import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { WorkbookRecoveryFixture } from "../testing/WorkbookRecoveryFixture";
import { CoordinationCreateRecovery } from "./features/coordination/CoordinationCreateRecovery";
import { coordinationFeature } from "./features/coordination/coordinationCreateModel";
import { NoteCreateRecovery } from "./features/notes/NoteCreateRecovery";
import { WorkbookMutationRuntime } from "./runtime/WorkbookMutationRuntime";

afterEach(cleanup);

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
