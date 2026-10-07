import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { useState } from "react";
import { afterEach, expect, it, vi } from "vitest";
import {
  WorkbookAuxiliaryDock,
  WorkbookAuxiliaryDockProvider,
} from "../layout/WorkbookAuxiliaryDock";
import { WorkbookSurfaceLayout } from "../layout/WorkbookSurfaceLayout";
import {
  useWorkbookBrowsingRegistry,
  WorkbookQueryBrowsingProvider,
} from "./WorkbookQueryBrowsingContext";
import { WorkbookQuerySurfaceLayout } from "./WorkbookQuerySurfaceLayout";

afterEach(cleanup);

const surface = {
  viewSchemaId: "composition",
  viewBar: null,
  statusStrip: null,
  primaryGrid: <button type="button">Background grid</button>,
};

it("isolated layout hosts one dock without query providers and restores focus across responsive closure", async () => {
  function Subject({ overlay }: { readonly overlay: boolean }) {
    const [open, setOpen] = useState(false);
    return (
      <WorkbookAuxiliaryDockProvider>
        <button type="button" onClick={() => setOpen(true)}>
          Open recovery
        </button>
        {open ? (
          <WorkbookAuxiliaryDock
            label="Recovery"
            onClose={() => setOpen(false)}
          >
            <button type="button">Recovery action</button>
          </WorkbookAuxiliaryDock>
        ) : null}
        <WorkbookSurfaceLayout
          {...surface}
          chromeMode={overlay ? "compact_desktop" : "base"}
        />
      </WorkbookAuxiliaryDockProvider>
    );
  }
  const mounted = render(<Subject overlay={false} />);
  const invoker = screen.getByRole("button", { name: "Open recovery" });
  invoker.focus();
  fireEvent.click(invoker);
  expect(screen.getAllByRole("region", { name: "Recovery" })).toHaveLength(1);
  expect(
    screen.getByRole("button", { name: "Background grid" }).closest("[inert]"),
  ).toBeNull();
  mounted.rerender(<Subject overlay />);
  expect(screen.getAllByRole("region", { name: "Recovery" })).toHaveLength(1);
  expect(
    screen.getByRole("button", { name: "Background grid" }).closest("[inert]"),
  ).not.toBeNull();
  const action = screen.getByRole("button", { name: "Recovery action" });
  action.focus();
  fireEvent.keyDown(action, { key: "Escape" });
  expect(screen.queryByRole("region", { name: "Recovery" })).toBeNull();
  await waitFor(() => expect(document.activeElement).toBe(invoker));
  expect(
    screen.getByRole("button", { name: "Background grid" }).closest("[inert]"),
  ).toBeNull();
});

it("query surfaces and auxiliary destinations reject missing required providers", () => {
  const error = vi.spyOn(console, "error").mockImplementation(() => {});
  try {
    expect(() => render(<WorkbookQuerySurfaceLayout {...surface} />)).toThrow(
      "WorkbookQueryBrowsingProvider is required",
    );
    expect(() =>
      render(
        <WorkbookAuxiliaryDock label="Recovery" onClose={() => {}}>
          Recovery
        </WorkbookAuxiliaryDock>,
      ),
    ).toThrow("WorkbookAuxiliaryDockProvider is required");
  } finally {
    error.mockRestore();
  }
});

it("query presentation cleanup fences replaced mounts and retires changed surface identities", () => {
  let registry: ReturnType<typeof useWorkbookBrowsingRegistry> | undefined;
  function Probe() {
    registry = useWorkbookBrowsingRegistry();
    return null;
  }
  function Subject({
    earlier,
    view,
  }: {
    readonly earlier: boolean;
    readonly view: string;
  }) {
    return (
      <WorkbookQueryBrowsingProvider>
        <Probe />
        {earlier ? (
          <WorkbookQuerySurfaceLayout key="earlier" {...surface} />
        ) : null}
        <WorkbookQuerySurfaceLayout
          key="current"
          {...surface}
          viewSchemaId={view}
        />
      </WorkbookQueryBrowsingProvider>
    );
  }
  const mounted = render(<Subject earlier view="composition" />);
  expect(registry?.presentationReady("composition")).toBe(true);
  mounted.rerender(<Subject earlier={false} view="composition" />);
  expect(registry?.presentationReady("composition")).toBe(true);
  mounted.rerender(<Subject earlier={false} view="replacement" />);
  expect(registry?.presentationReady("composition")).toBe(false);
  expect(registry?.presentationReady("replacement")).toBe(true);
  act(() => mounted.unmount());
  expect(registry?.presentationReady("replacement")).toBe(false);
});
