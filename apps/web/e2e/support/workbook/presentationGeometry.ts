import {
  dataTestIdSelector,
  gridGroupingSelectTestId,
  networkAnalysisTestId,
  savedViewActionMenuTriggerTestId,
  savedViewSelectorTestId,
  workbookActiveSurfaceFocusTargetTestId,
  workbookAddRowButtonTestId,
  workbookConflictControlTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorToggleTestId,
  workbookSortMenuTriggerTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { expectQuerySummaryGeometry } from "./querySummary";

export async function expectTextSpacingViewBarResilience(
  page: Page,
  options: {
    readonly capacity: number;
    readonly height: number;
    readonly hiddenCount: number;
    readonly width: number;
  },
) {
  await page.setViewportSize({ height: options.height, width: options.width });
  const queryControls = page.getByTestId(
    workbookViewBarQueryControlsTestId(timelineViewSchemaId),
  );
  await expect(queryControls).toHaveAttribute(
    "data-query-chip-capacity",
    String(options.capacity),
  );
  await expect(queryControls).toHaveAttribute(
    "data-hidden-query-chip-count",
    String(options.hiddenCount),
  );
  await expectQuerySummaryGeometry(page, options.capacity);
  const geometry = await page.evaluate(
    ({
      actionMenuSelector,
      addRowSelector,
      filterSelector,
      groupingSelector,
      inspectorSelector,
      queryControlsSelector,
      savedViewSelector,
      sortSelector,
    }) => {
      const select = (selector: string) =>
        document.querySelector<HTMLElement>(selector);
      const requireElement = (element: HTMLElement | null, label: string) => {
        if (element === null) throw new Error(`Expected ${label} to exist`);
        return element;
      };
      const controls = requireElement(
        select(queryControlsSelector),
        "query controls",
      );
      const viewBar = requireElement(
        controls.closest<HTMLElement>(
          'section[aria-label="Workbook query and action controls"]',
        ),
        "workbook view bar",
      );
      const columns = requireElement(
        Array.from(controls.querySelectorAll<HTMLButtonElement>("button")).find(
          (button) => button.textContent?.trim() === "Columns",
        ) ?? null,
        "Columns button",
      );
      const mode = viewBar.dataset.chromeMode;
      const queryInMenu =
        mode === "compact_desktop" || mode === "below_supported_minimum";
      const orderedControls: ReadonlyArray<readonly [string, HTMLElement]> = [
        ["saved-view", requireElement(select(savedViewSelector), "saved view")],
        [
          "saved-view-actions",
          requireElement(select(actionMenuSelector), "saved view actions"),
        ],
        ...(queryInMenu
          ? [
              [
                "query-menu",
                requireElement(
                  viewBar.querySelector<HTMLButtonElement>(
                    'button[aria-label="Query controls"]',
                  ),
                  "Query menu",
                ),
              ] as const,
            ]
          : ([
              ["sort", requireElement(select(sortSelector), "Sort button")],
              [
                "filters",
                requireElement(select(filterSelector), "Filters button"),
              ],
              [
                "group",
                requireElement(select(groupingSelector), "Group select"),
              ],
            ] as const)),
        mode === "base"
          ? ["columns", columns]
          : [
              "view-options",
              requireElement(
                viewBar.querySelector<HTMLButtonElement>(
                  'button[aria-label="View options controls"]',
                ),
                "View options menu",
              ),
            ],
        ...Array.from(
          viewBar.querySelectorAll<HTMLButtonElement>("button"),
        ).flatMap((button) => {
          const label = button.getAttribute("aria-label");
          return label === "Find in loaded rows" || label === "Clear contents"
            ? [[label, button] as const]
            : [];
        }),
        [
          "inspector",
          requireElement(select(inspectorSelector), "Inspector button"),
        ],
        ["add-row", requireElement(select(addRowSelector), "Add row button")],
      ];
      const viewBarRect = viewBar.getBoundingClientRect();
      const filter = requireElement(select(filterSelector), "Filters button");
      return {
        controls: orderedControls.map(([name, element]) => {
          const rect = element.getBoundingClientRect();
          return {
            clientHeight: element.clientHeight,
            left: rect.left,
            name,
            right: rect.right,
            scrollHeight: element.scrollHeight,
          };
        }),
        document: {
          clientHeight: document.documentElement.clientHeight,
          clientWidth: document.documentElement.clientWidth,
          scrollHeight: document.documentElement.scrollHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
        filter: {
          clientWidth: filter.clientWidth,
          scrollWidth: filter.scrollWidth,
        },
        viewBar: { left: viewBarRect.left, right: viewBarRect.right },
      };
    },
    {
      actionMenuSelector: dataTestIdSelector(
        savedViewActionMenuTriggerTestId(timelineViewSchemaId),
      ),
      addRowSelector: dataTestIdSelector(
        workbookAddRowButtonTestId(timelineViewSchemaId),
      ),
      filterSelector: dataTestIdSelector(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
      groupingSelector: dataTestIdSelector(
        gridGroupingSelectTestId(timelineViewSchemaId),
      ),
      inspectorSelector: dataTestIdSelector(
        workbookInspectorToggleTestId(timelineViewSchemaId),
      ),
      queryControlsSelector: dataTestIdSelector(
        workbookViewBarQueryControlsTestId(timelineViewSchemaId),
      ),
      savedViewSelector: dataTestIdSelector(
        savedViewSelectorTestId(timelineViewSchemaId),
      ),
      sortSelector: dataTestIdSelector(
        workbookSortMenuTriggerTestId(timelineViewSchemaId),
      ),
    },
  );
  expect(geometry.document.scrollWidth).toBeLessThanOrEqual(
    geometry.document.clientWidth + 1,
  );
  expect(geometry.document.scrollHeight).toBeLessThanOrEqual(
    geometry.document.clientHeight + 1,
  );
  expect(geometry.filter.scrollWidth).toBeLessThanOrEqual(
    geometry.filter.clientWidth + 1,
  );
  for (const control of geometry.controls) {
    expect(
      control.left,
      `${control.name} left containment`,
    ).toBeGreaterThanOrEqual(geometry.viewBar.left - 1);
    expect(
      control.right,
      `${control.name} right containment`,
    ).toBeLessThanOrEqual(geometry.viewBar.right + 1);
    expect(
      control.scrollHeight,
      `${control.name} block clipping`,
    ).toBeLessThanOrEqual(control.clientHeight + 1);
  }
  geometry.controls.slice(1).forEach((control, index) => {
    const previous = geometry.controls[index];
    expect(
      control.left,
      `${previous?.name} before ${control.name}`,
    ).toBeGreaterThanOrEqual((previous?.right ?? 0) - 1);
  });
}

export async function expectRecoverySurfaceGeometry(
  page: Page,
  options: { readonly enforceDocumentBlockExtent?: boolean } = {},
) {
  const geometry = await page.evaluate(
    ({ activeSurfaceSelector, panelSelector }) => {
      const panel = document.querySelector<HTMLElement>(panelSelector);
      const activeSurface = document.querySelector<HTMLElement>(
        activeSurfaceSelector,
      );
      if (panel === null || activeSurface === null) {
        throw new Error("Expected recovery panel and active surface");
      }
      const detail =
        panel.querySelector<HTMLElement>(
          '[aria-label="Workbook edit recovery"]',
        ) ?? panel;
      const heading = detail.querySelector<HTMLElement>("h2");
      const message = panel.querySelector<HTMLElement>('[role="status"]');
      const buttons = Array.from(
        panel.querySelectorAll<HTMLButtonElement>("button"),
      );
      const actionRow = detail.querySelector("button")?.parentElement;
      if (heading === null || message === null || actionRow == null) {
        throw new Error("Expected recovery heading, message, and actions");
      }
      const rect = (element: Element) => {
        const value = element.getBoundingClientRect();
        return {
          bottom: value.bottom,
          left: value.left,
          right: value.right,
          top: value.top,
        };
      };
      const style = getComputedStyle(panel);
      return {
        activeSurface: rect(activeSurface),
        actions: rect(actionRow),
        backgroundColor: style.backgroundColor,
        buttons: buttons.map(rect),
        document: {
          clientHeight: document.documentElement.clientHeight,
          clientWidth: document.documentElement.clientWidth,
          scrollHeight: document.documentElement.scrollHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
        heading: rect(heading),
        message: rect(message),
        overflowX: style.overflowX,
        overflowY: style.overflowY,
        panel: rect(panel),
        panelClientWidth: panel.clientWidth,
        panelScrollWidth: panel.scrollWidth,
      };
    },
    {
      activeSurfaceSelector: dataTestIdSelector(
        workbookActiveSurfaceFocusTargetTestId(),
      ),
      panelSelector: '[aria-label="Workbook recovery"]',
    },
  );

  expect(geometry.backgroundColor).not.toBe("rgba(0, 0, 0, 0)");
  expect(geometry.backgroundColor).not.toBe("transparent");
  expect(geometry.overflowX).toBe("hidden");
  expect(["auto", "scroll"]).toContain(geometry.overflowY);
  expect(geometry.panel.left).toBeGreaterThanOrEqual(
    geometry.activeSurface.left - 1,
  );
  expect(geometry.panel.right).toBeLessThanOrEqual(
    geometry.activeSurface.right + 1,
  );
  expect(geometry.panel.top).toBeGreaterThanOrEqual(
    geometry.activeSurface.top - 1,
  );
  expect(geometry.panel.bottom).toBeLessThanOrEqual(
    geometry.activeSurface.bottom + 1,
  );
  expect(geometry.heading.bottom).toBeLessThanOrEqual(geometry.message.top + 1);
  expect(geometry.message.bottom).toBeLessThanOrEqual(geometry.actions.top + 1);
  expect(geometry.panelScrollWidth).toBeLessThanOrEqual(
    geometry.panelClientWidth + 1,
  );
  for (const button of geometry.buttons) {
    expect(button.left).toBeGreaterThanOrEqual(geometry.panel.left - 1);
    expect(button.right).toBeLessThanOrEqual(geometry.panel.right + 1);
  }
  expect(geometry.document.scrollWidth).toBeLessThanOrEqual(
    geometry.document.clientWidth + 1,
  );
  if (options.enforceDocumentBlockExtent !== false) {
    expect(geometry.document.scrollHeight).toBeLessThanOrEqual(
      geometry.document.clientHeight + 1,
    );
  }
}

export async function expectNetworkAnalysisChromeGeometry(
  page: Page,
  options: { readonly enforceDocumentBlockExtent?: boolean } = {},
) {
  await page.evaluate(() => window.scrollTo(0, 0));
  const geometry = await page.evaluate(
    ({ advancedSelector, workspaceSelector }) => {
      const workspace = document.querySelector<HTMLElement>(workspaceSelector);
      if (workspace === null) {
        throw new Error("Expected Network Analysis workspace");
      }
      const controls = Array.from(
        workspace.querySelectorAll<HTMLElement>(
          "[data-network-flow-control]:not([hidden])",
        ),
      ).filter((control) => control.getClientRects().length > 0);
      const popover = workspace.querySelector<HTMLElement>(
        `${advancedSelector} .network-flow-popover`,
      );
      const rect = (element: Element) => {
        const value = element.getBoundingClientRect();
        return {
          bottom: value.bottom,
          height: value.height,
          left: value.left,
          right: value.right,
          top: value.top,
          width: value.width,
        };
      };
      return {
        controls: controls.map((control) => {
          const style = getComputedStyle(control);
          return {
            backgroundColor: style.backgroundColor,
            className: control.className,
            color: style.color,
            label:
              control.getAttribute("aria-label") ??
              control.textContent?.trim() ??
              control.tagName,
            ownedHorizontalOverflow: (() => {
              let ancestor = control.parentElement;
              while (ancestor !== null && ancestor !== workspace) {
                const ancestorStyle = getComputedStyle(ancestor);
                if (
                  (ancestorStyle.overflowX === "auto" ||
                    ancestorStyle.overflowX === "scroll") &&
                  ancestor.scrollWidth > ancestor.clientWidth
                ) {
                  return true;
                }
                ancestor = ancestor.parentElement;
              }
              return false;
            })(),
            rect: rect(control),
          };
        }),
        document: {
          clientHeight: document.documentElement.clientHeight,
          clientWidth: document.documentElement.clientWidth,
          scrollHeight: document.documentElement.scrollHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
        popover: popover === null ? null : rect(popover),
        workspace: rect(workspace),
        workspaceClientWidth: workspace.clientWidth,
        workspaceScrollWidth: workspace.scrollWidth,
      };
    },
    {
      advancedSelector: dataTestIdSelector(
        networkAnalysisTestId("advanced-filters"),
      ),
      workspaceSelector: dataTestIdSelector(networkAnalysisTestId("workspace")),
    },
  );

  expect(geometry.controls.length).toBeGreaterThan(10);
  expect(geometry.workspaceScrollWidth).toBeLessThanOrEqual(
    geometry.workspaceClientWidth + 1,
  );
  expect(geometry.document.scrollWidth).toBeLessThanOrEqual(
    geometry.document.clientWidth + 1,
  );
  if (options.enforceDocumentBlockExtent !== false) {
    expect(geometry.document.scrollHeight).toBeLessThanOrEqual(
      geometry.document.clientHeight + 1,
    );
  }
  for (const control of geometry.controls) {
    expect(control.backgroundColor).not.toBe("rgb(255, 255, 255)");
    expect(control.backgroundColor).not.toBe("rgba(255, 255, 255, 1)");
    expect(control.color).not.toBe("rgb(0, 0, 0)");
    expect(control.rect.width).toBeGreaterThan(0);
    expect(control.rect.height).toBeGreaterThan(0);
    if (!control.ownedHorizontalOverflow) {
      expect(control.rect.left, JSON.stringify(control)).toBeGreaterThanOrEqual(
        geometry.workspace.left - 1,
      );
      expect(control.rect.right, JSON.stringify(control)).toBeLessThanOrEqual(
        geometry.workspace.right + 1,
      );
    }
  }
  if (geometry.popover !== null) {
    expect(geometry.popover.left).toBeGreaterThanOrEqual(0);
    expect(geometry.popover.right).toBeLessThanOrEqual(
      geometry.document.clientWidth + 1,
    );
    expect(geometry.popover.top).toBeGreaterThanOrEqual(0);
    expect(geometry.popover.bottom).toBeLessThanOrEqual(
      geometry.document.clientHeight + 1,
    );
  }
}

export async function expectResolverSurfaceGeometry(
  page: Page,
  options: { readonly enforceDocumentBlockExtent?: boolean } = {},
) {
  const resolver = page.getByRole("region", {
    name: "Workbook recovery",
    exact: true,
  });
  await resolver.evaluate((element) => {
    element.scrollTop = 0;
  });
  const geometry = await page.evaluate(
    ({ activeSurfaceSelector, resolverSelector }) => {
      const panel = document.querySelector<HTMLElement>(resolverSelector);
      const activeSurface = document.querySelector<HTMLElement>(
        activeSurfaceSelector,
      );
      if (panel === null || activeSurface === null) {
        throw new Error("Expected resolver and active surface");
      }
      const rect = (element: Element) => {
        const value = element.getBoundingClientRect();
        return {
          bottom: value.bottom,
          left: value.left,
          right: value.right,
          top: value.top,
        };
      };
      const style = getComputedStyle(panel);
      return {
        activeSurface: rect(activeSurface),
        backgroundColor: style.backgroundColor,
        document: {
          clientHeight: document.documentElement.clientHeight,
          clientWidth: document.documentElement.clientWidth,
          scrollHeight: document.documentElement.scrollHeight,
          scrollWidth: document.documentElement.scrollWidth,
        },
        overflowX: style.overflowX,
        overflowY: style.overflowY,
        panel: rect(panel),
        panelClientWidth: panel.clientWidth,
        panelScrollWidth: panel.scrollWidth,
      };
    },
    {
      activeSurfaceSelector: dataTestIdSelector(
        workbookActiveSurfaceFocusTargetTestId(),
      ),
      resolverSelector: '[aria-label="Workbook recovery"]',
    },
  );
  expect(geometry.backgroundColor).not.toBe("rgba(0, 0, 0, 0)");
  expect(geometry.backgroundColor).not.toBe("transparent");
  expect(geometry.overflowX).toBe("hidden");
  expect(["auto", "scroll"]).toContain(geometry.overflowY);
  expect(geometry.panel.left).toBeGreaterThanOrEqual(
    geometry.activeSurface.left - 1,
  );
  expect(geometry.panel.right).toBeLessThanOrEqual(
    geometry.activeSurface.right + 1,
  );
  expect(geometry.panel.top).toBeGreaterThanOrEqual(
    geometry.activeSurface.top - 1,
  );
  expect(geometry.panel.bottom).toBeLessThanOrEqual(
    geometry.activeSurface.bottom + 1,
  );
  expect(geometry.panelScrollWidth).toBeLessThanOrEqual(
    geometry.panelClientWidth + 1,
  );
  expect(geometry.document.scrollWidth).toBeLessThanOrEqual(
    geometry.document.clientWidth + 1,
  );
  if (options.enforceDocumentBlockExtent !== false) {
    expect(geometry.document.scrollHeight).toBeLessThanOrEqual(
      geometry.document.clientHeight + 1,
    );
  }

  const finalAction = page.getByTestId(
    workbookConflictControlTestId("use-merged"),
  );
  await finalAction.scrollIntoViewIfNeeded();
  await expect(finalAction).toBeVisible();
  const actionGeometry = await resolver.evaluate(
    (panel, actionSelector) => {
      const action = panel.querySelector<HTMLElement>(actionSelector);
      if (action === null) throw new Error("Expected resolver final action");
      const panelRect = panel.getBoundingClientRect();
      const actionRect = action.getBoundingClientRect();
      return {
        actionBottom: actionRect.bottom,
        actionLeft: actionRect.left,
        actionRight: actionRect.right,
        actionTop: actionRect.top,
        panelBottom: panelRect.bottom,
        panelLeft: panelRect.left,
        panelRight: panelRect.right,
        panelTop: panelRect.top,
      };
    },
    dataTestIdSelector(workbookConflictControlTestId("use-merged")),
  );
  expect(actionGeometry.actionLeft).toBeGreaterThanOrEqual(
    actionGeometry.panelLeft - 1,
  );
  expect(actionGeometry.actionRight).toBeLessThanOrEqual(
    actionGeometry.panelRight + 1,
  );
  expect(actionGeometry.actionTop).toBeGreaterThanOrEqual(
    actionGeometry.panelTop - 1,
  );
  expect(actionGeometry.actionBottom).toBeLessThanOrEqual(
    actionGeometry.panelBottom + 1,
  );
}
