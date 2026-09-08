import {
  savedViewActionMenuTriggerTestId,
  savedViewStatusTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { useLayoutEffect, useSyncExternalStore } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import {
  createSavedViewTestController,
  savedViewTestAuthority,
  savedViewTestResource,
} from "../../testing/workbookSavedViewTestSupport";
import { buildSavedViewLayoutJson } from "../models/workbookQuery";
import { workbookSavedViewsResource } from "../models/workbookSavedViewControl";
import type { SavedViewResource } from "../models/workbookSavedViews";
import type {
  SavedViewResult,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";
import { ActiveSurfaceSavedViewSelector } from "./ActiveSurfaceSavedViewSelector";

const surface = "cartulary.view.timeline.v2";
const saved = savedViewTestResource();
function Harness({
  controller,
  schema = surface,
  selectedId = saved.saved_view_id,
  onSelect = vi.fn(),
  onBase = vi.fn(),
}: {
  controller: WorkbookSavedViewController;
  schema?: string;
  selectedId?: string | null;
  onSelect?: (r: SavedViewResource) => void;
  onBase?: (id: string) => void;
}) {
  const snapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const selected = snapshot.resources.find(
    (r) => r.saved_view_id === selectedId,
  );
  const sheetRef =
    selectedId === null
      ? { kind: "view_schema" as const, id: schema }
      : { kind: "saved_view" as const, id: selectedId };
  useLayoutEffect(() => {
    controller.setBinding({
      incidentId: "incident-1",
      subject: {
        viewSchemaId: schema,
        savedViewId: selectedId,
        savedViewVersion: selected?.saved_view_version ?? null,
      },
      sheetRef,
      queryJson: { sort: [], filters: [] },
      layoutJson: buildSavedViewLayoutJson(requireViewContract(schema)),
      selectionGeneration: schema === surface ? 0 : 1,
      workingGeneration: 0,
      applyConfiguration: vi.fn(),
      select: onSelect,
      deleted: () => onBase(schema),
      authorizationRecovered: vi.fn(),
    });
  });
  useLayoutEffect(() => () => controller.setBinding(null), [controller]);
  return (
    <ActiveSurfaceSavedViewSelector
      controller={controller}
      activeViewSchemaId={schema}
      chromeMode="base"
      currentIncidentRole={snapshot.authority?.role ?? null}
      currentUserId={snapshot.authority?.actorId ?? null}
      isModified
      savedViewsResource={
        snapshot.list === "ready"
          ? workbookSavedViewsResource(snapshot.resources, sheetRef)
          : snapshot.list === "loading"
            ? { kind: "loading" }
            : {
                kind: "unavailable",
                message: snapshot.listProblem?.message ?? "Unavailable",
              }
      }
      selectedSheetRef={sheetRef}
      onSelectBaseSurface={onBase}
      onSelectSavedView={onSelect}
    />
  );
}
async function setup(
  overrides: Partial<WorkbookSavedViewPort> = {},
  resources = [saved],
) {
  let list = resources;
  const port: WorkbookSavedViewPort = {
    listPage: vi.fn<WorkbookSavedViewPort["listPage"]>(async () => ({
      kind: "accepted",
      value: { nextCursor: null, savedViews: list },
    })),
    create: vi.fn<WorkbookSavedViewPort["create"]>(async ({ definition }) => {
      const resource = savedViewTestResource({
        saved_view_id: "created-view",
        display_name: definition.displayName,
        scope: definition.scope,
        query_json: definition.queryJson,
        layout_json: definition.layoutJson,
      });
      list = [...list, resource];
      return { kind: "accepted", value: resource };
    }),
    patch: vi.fn<WorkbookSavedViewPort["patch"]>(),
    delete: vi.fn<WorkbookSavedViewPort["delete"]>(),
    ...overrides,
  };
  const controller = createSavedViewTestController(port);
  const onSelect = vi.fn();
  const view = render(<Harness controller={controller} onSelect={onSelect} />);
  await waitFor(() => expect(controller.getSnapshot().list).toBe("ready"));
  return { ...view, controller, port, onSelect };
}
function open(schema = surface) {
  fireEvent.click(screen.getByTestId(savedViewActionMenuTriggerTestId(schema)));
}
function button(name: string) {
  return screen.getByRole("button", { name });
}

describe("ActiveSurfaceSavedViewSelector", () => {
  afterEach(cleanup);
  it("uses a labelled dialog and restores keyboard focus on Escape", async () => {
    const h = await setup();
    open();
    expect(screen.getByRole("dialog", { name: "Saved view" })).toBeTruthy();
    expect(document.activeElement).toBe(
      screen.getByLabelText("Saved view name"),
    );
    fireEvent.keyDown(screen.getByLabelText("Saved view name"), {
      key: "Escape",
    });
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(document.activeElement).toBe(
      screen.getByTestId(savedViewActionMenuTriggerTestId(surface)),
    );
    h.controller.dispose();
  });
  it("retains late confirmation without reopening the panel or navigating after a surface switch", async () => {
    const pending = deferred<SavedViewResult<SavedViewResource>>();
    const h = await setup({ create: vi.fn(() => pending.promise) });
    open();
    fireEvent.click(button("Save current configuration as new view"));
    fireEvent.click(button("Save current configuration as new view"));
    await waitFor(() => expect(h.port.create).toHaveBeenCalledOnce());
    fireEvent.keyDown(screen.getByLabelText("Saved view name"), {
      key: "Escape",
    });
    h.rerender(
      <Harness
        controller={h.controller}
        schema="cartulary.view.hosts.v1"
        selectedId={null}
        onSelect={h.onSelect}
      />,
    );
    await act(async () =>
      pending.resolve({
        kind: "accepted",
        value: savedViewTestResource({ saved_view_id: "created-view" }),
      }),
    );
    expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(h.onSelect).not.toHaveBeenCalled();
    expect(screen.queryByRole("dialog")).toBeNull();
    open("cartulary.view.hosts.v1");
    expect(button("Open confirmed saved view")).toBeTruthy();
    h.controller.dispose();
  });
  it("keeps write confirmation local when complete-list refresh fails", async () => {
    const h = await setup();
    open();
    vi.mocked(h.port.listPage).mockResolvedValue({
      kind: "rejected",
      failure: { kind: "transport", message: "List offline" },
    });
    fireEvent.click(button("Save current configuration as new view"));
    await waitFor(() =>
      expect(screen.getByText(/The write is confirmed/)).toBeTruthy(),
    );
    expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(button("Refresh saved views for review")).toBeTruthy();
    h.controller.dispose();
  });
  it("requires explicit duplicate-risk confirmation before a new uncertain create attempt", async () => {
    const h = await setup({
      create: vi
        .fn<WorkbookSavedViewPort["create"]>()
        .mockResolvedValueOnce({
          kind: "uncertain",
          failure: { kind: "transport", message: "Response lost" },
        })
        .mockResolvedValueOnce({
          kind: "accepted",
          value: savedViewTestResource({ saved_view_id: "new-attempt" }),
        }),
    });
    open();
    fireEvent.click(button("Save current configuration as new view"));
    await waitFor(() =>
      expect(h.controller.getSnapshot().observation).not.toBeNull(),
    );
    expect(h.controller.getSnapshot().operation.kind).toBe("uncertain");
    expect(
      screen.queryByRole("button", {
        name: "Apply submitted changes to reviewed version",
      }),
    ).toBeNull();
    expect(
      (button("Make a new create attempt") as HTMLButtonElement).disabled,
    ).toBe(true);
    expect(h.port.create).toHaveBeenCalledOnce();
    fireEvent.keyDown(screen.getByLabelText("Saved view name"), {
      key: "Escape",
    });
    open();
    expect(
      screen.getByText(/matching name or configuration is not a receipt/),
    ).toBeTruthy();
    fireEvent.click(
      screen.getByRole("checkbox", { name: /may create a duplicate/ }),
    );
    fireEvent.click(button("Make a new create attempt"));
    await waitFor(() => expect(h.port.create).toHaveBeenCalledTimes(2));
    expect(h.onSelect).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("shows submitted and observed conflict configuration and waits for explicit review", async () => {
    const h = await setup({
      patch: vi
        .fn<WorkbookSavedViewPort["patch"]>()
        .mockResolvedValueOnce({
          kind: "rejected",
          failure: {
            kind: "conflict",
            publicCode: "saved_view_version_conflict",
            message: "Changed",
            conflict: {
              savedViewId: saved.saved_view_id,
              baseVersion: 1,
              currentVersion: 2,
            },
          },
        })
        .mockResolvedValueOnce({
          kind: "accepted",
          value: savedViewTestResource({
            display_name: "My draft",
            saved_view_version: 3,
            updated_at: "2026-08-01T00:02:00Z",
          }),
        }),
    });
    open();
    vi.mocked(h.port.listPage).mockResolvedValue({
      kind: "accepted",
      value: {
        nextCursor: null,
        savedViews: [
          savedViewTestResource({
            display_name: "Their edit",
            saved_view_version: 2,
            updated_at: "2026-08-01T00:01:00Z",
          }),
        ],
      },
    });
    fireEvent.change(screen.getByLabelText("Saved view name"), {
      target: { value: "My draft" },
    });
    fireEvent.click(button("Update selected view"));
    await waitFor(() =>
      expect(
        screen.getByText("Observed saved configuration (version 2)"),
      ).toBeTruthy(),
    );
    expect(screen.getByLabelText("Saved view name")).toHaveProperty(
      "value",
      "My draft",
    );
    expect(h.port.patch).toHaveBeenCalledOnce();
    fireEvent.click(button("Apply submitted changes to reviewed version"));
    await waitFor(() => expect(h.port.patch).toHaveBeenCalledTimes(2));
    expect(vi.mocked(h.port.patch).mock.calls[1]?.[0]).toMatchObject({
      base: { saved_view_version: 2 },
      changes: { displayName: "My draft" },
    });
    h.controller.dispose();
  });
  it("associates name validation and permits editing an invalid generated copy name", async () => {
    const h = await setup({}, [
      savedViewTestResource({ display_name: "A".repeat(256) }),
    ]);
    open();
    fireEvent.click(button("Duplicate selected view"));
    expect(
      screen
        .getByLabelText("Name for reviewed saved-view request")
        .getAttribute("aria-invalid"),
    ).toBe("true");
    expect(
      screen.getByLabelText("Name for reviewed saved-view request"),
    ).toHaveProperty("value", `${"A".repeat(256)} Copy`);
    fireEvent.change(
      screen.getByLabelText("Name for reviewed saved-view request"),
      { target: { value: "Short copy" } },
    );
    fireEvent.keyDown(
      screen.getByLabelText("Name for reviewed saved-view request"),
      { key: "Escape" },
    );
    open();
    expect(
      screen.getByLabelText("Name for reviewed saved-view request"),
    ).toHaveProperty("value", "Short copy");
    fireEvent.click(button("Save submitted configuration as a new view"));
    await waitFor(() => expect(h.port.create).toHaveBeenCalledOnce());
    expect(
      vi.mocked(h.port.create).mock.calls[0]?.[0].definition.displayName,
    ).toBe("Short copy");
    h.controller.dispose();
  });
  it("preserves system immutability and rechecks permission changes", async () => {
    const h = await setup({}, [
      savedViewTestResource({ scope: "system", owner_user_id: null }),
    ]);
    open();
    expect((button("Update selected view") as HTMLButtonElement).disabled).toBe(
      true,
    );
    expect((button("Delete selected view") as HTMLButtonElement).disabled).toBe(
      true,
    );
    expect(
      (button("Duplicate selected view") as HTMLButtonElement).disabled,
    ).toBe(false);
    act(() =>
      h.controller.setAuthority({ ...savedViewTestAuthority, role: "viewer" }),
    );
    expect((button("Update selected view") as HTMLButtonElement).disabled).toBe(
      true,
    );
    h.controller.dispose();
  });
  it("falls back from invalid selection with local notice while preserving list publication", async () => {
    const h = await setup();
    const onBase = vi.fn();
    h.rerender(
      <Harness
        controller={h.controller}
        selectedId="deleted-view"
        onBase={onBase}
      />,
    );
    await waitFor(() => expect(onBase).toHaveBeenCalledWith(surface));
    expect(screen.getByTestId(savedViewStatusTestId(surface))).toHaveProperty(
      "textContent",
      expect.stringContaining("no longer available"),
    );
    expect(
      within(screen.getByRole("combobox", { name: "Saved view" })).getByRole(
        "option",
        { name: saved.display_name },
      ),
    ).toBeTruthy();
    h.controller.dispose();
  });
});
