import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import { buildSavedViewLayoutJson } from "../models/workbookQuery";
import { WorkbookSessionNavigation } from "./WorkbookSessionNavigation";
import {
  type WorkbookWorkbench,
  WorkbookWorkbenchContext,
} from "./WorkbookWorkbenchContext";
import { WorkbookNavigationStatus } from "./WorkbookWorkPanel";

afterEach(cleanup);

it("navigation details retain keyboard access and stable status controls across outcomes", async () => {
  const session = new WorkbookSessionNavigation("incident");
  session.setReadable(true);
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
    pinViewLabel: "Pin view",
    cancelNavigation: () => session.cancel(),
    message: null,
    retry: null,
    openBase: null,
  };
  const view = render(
    <WorkbookWorkbenchContext value={value}>
      <WorkbookNavigationStatus />
    </WorkbookWorkbenchContext>,
  );
  const user = userEvent.setup();
  const trigger = screen.getByRole("button", {
    name: "Navigation",
  });
  await act(async () => {
    await session.navigate(
      {
        target: { sheetRef: { kind: "view_schema", id: "view" } },
        entry: "open",
        inspect: false,
      },
      {
        incidentId: "incident",
        sheetRef: { kind: "view_schema", id: "view" },
        viewSchemaId: "view",
        query: { sort: [], filters: [], groupBy: null },
        layout: buildSavedViewLayoutJson(
          requireViewContract("cartulary.view.timeline.v2"),
        ),
        invoker: "view",
      },
      async () => "same",
    );
  });
  expect(
    screen.getByRole("status", { name: "Navigation updates" }).textContent,
  ).toBe("Opening destination…");
  trigger.focus();
  await user.keyboard("{Enter}");
  expect(document.activeElement).toBe(
    screen.getByRole("dialog", { name: "Navigation details" }),
  );
  await user.keyboard("{Escape}");
  expect(document.activeElement).toBe(trigger);
  expect(session.getSnapshot().outcome).toBe("admitted");
  await user.keyboard("{Enter}");
  await user.click(screen.getByRole("button", { name: "Cancel navigation" }));
  expect(session.getSnapshot().outcome).toBe("cancelled");
  expect(document.activeElement).toBe(
    screen.getByRole("dialog", { name: "Navigation details" }),
  );
  expect(screen.getByRole("button", { name: "Navigation" })).toBe(trigger);
  view.rerender(
    <WorkbookWorkbenchContext
      value={{
        ...value,
        message: "A destination is unavailable.",
        retry: vi.fn(),
        openBase: vi.fn(),
      }}
    >
      <WorkbookNavigationStatus compact />
    </WorkbookWorkbenchContext>,
  );
  expect(screen.getByRole("button", { name: "Navigation" })).toBe(trigger);
  expect(screen.getByRole("button", { name: "Retry navigation" })).toBeTruthy();
  expect(screen.getByRole("button", { name: "Open base view" })).toBeTruthy();
  act(() => session.setReadable(false));
  expect(screen.queryByRole("dialog")).toBeNull();
  expect(screen.queryByRole("button", { name: "Navigation" })).toBeNull();
});
