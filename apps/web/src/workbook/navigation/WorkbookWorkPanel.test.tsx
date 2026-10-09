import {
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import {
  useWorkbookAuxiliaryDock,
  WorkbookAuxiliaryDockProvider,
} from "../layout/WorkbookAuxiliaryDock";
import { WorkbookSurfaceLayout } from "../layout/WorkbookSurfaceLayout";
import { buildSavedViewLayoutJson } from "../models/workbookQuery";
import {
  WorkbookSessionNavigation,
  type WorkbookSessionPin,
} from "./WorkbookSessionNavigation";
import {
  type WorkbookWorkbench,
  WorkbookWorkbenchContext,
} from "./WorkbookWorkbenchContext";
import { WorkbookWorkControls } from "./WorkbookWorkPanel";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
const label =
  "A collection script launched from a temporary directory; the parent process and command line were retained for review.";
const record = (id: string, name = label): WorkbookSessionPin => ({
  incidentId: "incident",
  sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
  recordId: id,
  label: name,
});
function Detach() {
  const dock = useWorkbookAuxiliaryDock();
  return (
    <button type="button" onClick={() => dock?.attachment?.close("navigation")}>
      Other destination
    </button>
  );
}
function fixture(
  pins = [record("a"), record("b"), record("c")],
  pinViewLabel: WorkbookWorkbench["pinViewLabel"] = "Pin view",
) {
  const session = new WorkbookSessionNavigation("incident");
  session.setReadable(true);
  for (const pin of pins) session.pin(pin);
  const value: WorkbookWorkbench = {
    session,
    navigationReady: false,
    inspectValue: null,
    requestInspectValue: () => false,
    acknowledgeInspectValue: vi.fn(),
    open: vi.fn(),
    openPin: vi.fn(),
    pinCurrentView: vi.fn(),
    pinRecord: vi.fn(),
    returnToOrigin: vi.fn(),
    registerInspector: () => () => {},
    registerInspectorFocus: () => () => {},
    pinViewLabel,
    cancelNavigation: () => session.cancel(),
    message: null,
    retry: null,
    openBase: null,
  };
  const view = render(
    <WorkbookWorkbenchContext value={value}>
      <WorkbookAuxiliaryDockProvider>
        <WorkbookWorkControls />
        <Detach />
        <WorkbookSurfaceLayout
          viewSchemaId={timelineViewSchemaId}
          viewBar={null}
          statusStrip={null}
          primaryGrid={<button type="button">Grid interaction</button>}
          chromeMode="base"
        />
      </WorkbookAuxiliaryDockProvider>
    </WorkbookWorkbenchContext>,
  );
  const trigger = screen.getByRole("button", { name: "Work" });
  fireEvent.click(trigger);
  const item = (id: string) => {
    const match = screen
      .getAllByRole("listitem")
      .find((node) => within(node).queryByText(id, { exact: true }));
    if (!match) throw new Error(`Missing pin ${id}`);
    return within(match);
  };
  const remove = (id: string) =>
    item(id).getByRole("button", { name: /^Remove / });
  return { ...view, session, value, trigger, item, remove };
}

it("Work identifies equal labels from descriptors and removes exactly one pin without reads or writes", async () => {
  const fetch = vi.fn();
  vi.stubGlobal("fetch", fetch);
  const f = fixture();
  const user = userEvent.setup();
  await user.click(f.item("b").getByLabelText(`Identity for ${label}`));
  expect(f.item("b").getByText("Record ID")).toBeTruthy();
  expect(f.item("b").getByText("Timeline · Record")).toBeTruthy();
  await user.click(f.remove("b"));
  expect(f.session.getSnapshot().pins.map((pin) => pin.recordId)).toEqual([
    "a",
    "c",
  ]);
  expect(document.activeElement).toBe(f.remove("c"));
  expect(fetch).not.toHaveBeenCalled();
  expect(f.value.openPin).not.toHaveBeenCalled();
  await user.click(f.item("a").getByRole("button", { name: label }));
  expect(f.value.openPin).toHaveBeenCalledExactlyOnceWith(record("a"));
});

it("Work restores first middle last and sole removal focus and closes immediately with Escape", async () => {
  const user = userEvent.setup();
  for (const [removed, next] of [
    ["a", "b"],
    ["b", "c"],
    ["c", "b"],
  ] as const) {
    const f = fixture();
    f.remove(removed).focus();
    await user.keyboard("{Enter}");
    expect(document.activeElement).toBe(f.remove(next));
    expect(f.remove(next).hasAttribute("data-work-pin-removal-focus")).toBe(
      true,
    );
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("region", { name: "Work" })).toBeNull();
    expect(document.activeElement).toBe(f.trigger);
    f.unmount();
  }
  const f = fixture([record("only")], "Pin base surface");
  f.remove("only").focus();
  await user.keyboard("{Enter}");
  expect(document.activeElement).toBe(
    screen.getByRole("button", { name: "Pin base surface" }),
  );
  expect(screen.getByText(/No pins yet/)).toBeTruthy();
  await user.keyboard("{Escape}");
  expect(document.activeElement).toBe(f.trigger);
});

it("Work skips disabled removal controls without moving focus for an unfocused removal", () => {
  const f = fixture();
  (f.remove("b") as HTMLButtonElement).disabled = true;
  f.remove("a").focus();
  fireEvent.click(f.remove("a"));
  expect(document.activeElement).toBe(f.remove("c"));
  const external = screen.getByRole("button", { name: "Grid interaction" });
  external.focus();
  fireEvent.click(f.remove("c"));
  expect(document.activeElement).toBe(external);
});

it("Work retires removal focus after deliberate interaction closure detachment and unmount", () => {
  for (const kind of [
    "focus",
    "pointer",
    "keyboard",
    "close",
    "detach",
    "unmount",
  ]) {
    const f = fixture();
    const apply = f.session.remove.bind(f.session);
    vi.spyOn(f.session, "remove").mockImplementation(() => {});
    f.remove("a").focus();
    fireEvent.click(f.remove("a"));
    const external = screen.getByRole("button", { name: "Grid interaction" });
    if (kind === "focus") external.focus();
    if (kind === "pointer") fireEvent.pointerDown(external);
    if (kind === "keyboard") fireEvent.keyDown(document, { key: "Tab" });
    if (kind === "close")
      fireEvent.click(screen.getByRole("button", { name: "Close Work" }));
    if (kind === "detach")
      fireEvent.click(
        screen.getByRole("button", { name: "Other destination" }),
      );
    if (kind === "unmount") f.unmount();
    const before = document.activeElement;
    act(() => apply(record("a")));
    if (
      kind === "focus" ||
      kind === "close" ||
      kind === "detach" ||
      kind === "unmount"
    )
      expect(document.activeElement).toBe(before);
    else expect(document.activeElement).toBe(document.body);
    expect(document.querySelector("[data-work-pin-removal-focus]")).toBeNull();
    f.unmount();
  }
});

it("Work fences removal focus across authority loss navigation and unexpected pin changes", async () => {
  for (const kind of ["authority", "navigation", "pins"]) {
    const f = fixture();
    const apply = f.session.remove.bind(f.session);
    vi.spyOn(f.session, "remove").mockImplementation(() => {});
    f.remove("a").focus();
    fireEvent.click(f.remove("a"));
    await act(async () => {
      if (kind === "authority") {
        // Even batched loss/recovery must retire the older intention.
        f.session.setReadable(false);
        f.session.setReadable(true);
      } else if (kind === "pins") f.session.pin(record("new"));
      else
        await f.session.navigate(
          {
            target: {
              sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
            },
            entry: "open",
            inspect: false,
          },
          {
            incidentId: "incident",
            sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
            viewSchemaId: timelineViewSchemaId,
            invoker: "work",
            query: { sort: [], filters: [], groupBy: null },
            layout: buildSavedViewLayoutJson(
              requireViewContract(timelineViewSchemaId),
            ),
          },
          async () => "same",
        );
      apply(record("a"));
    });
    expect(document.querySelector("[data-work-pin-removal-focus]")).toBeNull();
    expect(document.activeElement).not.toBe(f.remove("b"));
    f.unmount();
  }
});

it("Work conceals unavailable identity and labels while retaining generic removal", async () => {
  const f = fixture([
    record("protected-id", "Concealed source label"),
    record("readable-id", "Unavailable item"),
  ]);
  act(() => f.session.concealPin(record("protected-id")));
  const work = screen.getByRole("region", { name: "Work" });
  expect(work.innerHTML).not.toContain("protected-id");
  expect(work.innerHTML).not.toContain("Concealed source label");
  expect(
    f.item("readable-id").getByLabelText("Identity for Unavailable item"),
  ).toBeTruthy();
  const unavailable = within(work)
    .getAllByRole("listitem")
    .find((node) => !within(node).queryByText("Record ID"));
  if (!unavailable) throw new Error("Missing generic pin");
  await userEvent.setup().click(
    within(unavailable).getByRole("button", {
      name: "Remove Unavailable item from working set",
    }),
  );
  expect(f.session.getSnapshot().pins.map((pin) => pin.recordId)).toEqual([
    "readable-id",
  ]);
  act(() => f.session.setReadable(false));
  expect(screen.queryByRole("region", { name: "Work" })).toBeNull();
  expect(document.body.innerHTML).not.toContain("readable-id");
});

it("Work presents each supported descriptor kind without adding metadata observations", () => {
  const f = fixture([
    {
      incidentId: "incident",
      sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
      label: "Timeline",
    },
    {
      incidentId: "incident",
      sheetRef: { kind: "saved_view", id: "saved-id" },
      label: "Saved label",
    },
    {
      incidentId: "incident",
      sheetRef: {
        kind: "extension_workspace",
        extension_profile_id: "profile-id",
        workspace_key: "workspace-key",
      },
      label: "Workspace label",
    },
  ]);
  expect(screen.getByText("Timeline · Base surface")).toBeTruthy();
  expect(f.item("saved-id").getByText("Saved view ID")).toBeTruthy();
  expect(f.item("workspace-key").getByText("profile-id")).toBeTruthy();
  expect(f.value.open).not.toHaveBeenCalled();
});
