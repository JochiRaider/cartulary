import { Buffer } from "node:buffer";
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  pasteGridMatrix,
} from "@cartulary/test-utils";
import {
  autoResolutionNoticeTestId,
  autoResolutionUndoButtonTestId,
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  currentIncidentRoleTestId,
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  landingIncidentCardTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
  phase1LandingTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  rowPresenceMarkerTestId,
  type SystemViewSwitcherGroupToken,
  savedViewCreateButtonTestId,
  savedViewNameInputTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStatusTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineInspectorSectionTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineScalarEditorTestId,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import type { APIRequestContext, Locator, Page, Route } from "@playwright/test";
import {
  indicatorsViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "../src/workbook/models/workbookSurfaceRegistry";
import {
  p1AccessibilityScenarioTitles,
  scenarioTitlesForAccessibilityRow,
} from "./a11yPhaseMap";
import {
  createLocalUser as createAuthLocalUser,
  revokeAllSessions,
} from "./authRuntime";
import {
  createEvidenceFixtureRow,
  createUploadedEvidenceFixture,
  type EvidenceUploadOptions,
} from "./evidenceFixtureHelpers";
import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createIncidentMembership,
  createIncidentMemberUser,
  createViewRow,
  csrfHeaders,
  enrollTotpViaBootstrap,
  generateTotpCode,
  patchTimelineRecord,
  safeUnroute,
  sessionCookieName,
  testRouteHeaders,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import { Phase1Page } from "./phase1Page";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  collectionItems,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  hostRefsFieldKey,
  lessonViewSchemaId,
  openTimelineInspector,
  partiesViewSchemaId,
  requireItemByRawText,
  seedHostMentionStateFixture,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "./phase4Helpers";
import {
  driveRealTimelineSummaryConflict,
  focusRemoteTimelineCellAndWaitForPresence,
  installIncidentSocketMonitor,
  installPatchController,
  installPatchTransportFailureController,
  openIncidentAsTrackedUserReady,
  requireRecordId,
  successfulPatchCalls,
} from "./phase6Harness";

type IncidentMembershipRecord = {
  membership_version: number;
  role: string;
  user_id: string;
};

type Phase9A11yHistoryItem = {
  available_rollback_actions: Array<
    "history_entry" | "change_set" | "row_restore"
  >;
  history_entry_ref?: string;
  history_item_ref: string;
};

type Phase9A11yHistoryData = {
  items: Phase9A11yHistoryItem[];
};

type ViewRow = {
  cells: Record<string, { value: unknown }>;
  record_id: string;
  row_version: number;
};

declare const phase1A11yAppLocalTestIdBrand: unique symbol;

type Phase1A11yAppLocalTestId = string & {
  readonly [phase1A11yAppLocalTestIdBrand]: "Phase1A11yAppLocalTestId";
};

const p2AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P2-01",
) as [string];
const p3AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P3-01",
) as [string];
const p4AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P4-01",
) as [string];
const p5AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P5-01",
) as [string];
const p6AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P6-01",
) as [string];
const p7AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P7-01",
) as [string];
const p8AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P8-01",
) as [string];
const p9AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P9-01",
) as [string];
const p10AccessibilityScenarioTitles = scenarioTitlesForAccessibilityRow(
  "FE-A11Y-P10-01",
) as [string];

if (p2AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P2-01 must declare exactly 1 scenario; found ${p2AccessibilityScenarioTitles.length}`,
  );
}
if (p3AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P3-01 must declare exactly 1 scenario; found ${p3AccessibilityScenarioTitles.length}`,
  );
}
if (p4AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P4-01 must declare exactly 1 scenario; found ${p4AccessibilityScenarioTitles.length}`,
  );
}
if (p5AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P5-01 must declare exactly 1 scenario; found ${p5AccessibilityScenarioTitles.length}`,
  );
}
if (p6AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P6-01 must declare exactly 1 scenario; found ${p6AccessibilityScenarioTitles.length}`,
  );
}
if (p7AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P7-01 must declare exactly 1 scenario; found ${p7AccessibilityScenarioTitles.length}`,
  );
}
if (p8AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P8-01 must declare exactly 1 scenario; found ${p8AccessibilityScenarioTitles.length}`,
  );
}
if (p9AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P9-01 must declare exactly 1 scenario; found ${p9AccessibilityScenarioTitles.length}`,
  );
}
if (p10AccessibilityScenarioTitles.length !== 1) {
  throw new Error(
    `FE-A11Y-P10-01 must declare exactly 1 scenario; found ${p10AccessibilityScenarioTitles.length}`,
  );
}

const phase1A11yAppLocalSelectors = Object.freeze({
  incidentPatchButton: {
    owner: "apps/web incident administration",
    reason:
      "Incident patch controls are app-local to the incident admin panel until later incident-surface selector promotion.",
    scope: "FE-P1 selected-incident accessibility recovery path",
    testId: "incident-patch-button" as Phase1A11yAppLocalTestId,
  },
});

const keyboardSentinelId = "a11y-keyboard-sentinel";
const contrastThreshold = 4.5;

const privateDiagnosticPatterns = [
  /bootstrap[_ -]?token/i,
  /secret[_ -]?base32/i,
  /PRIVATESECRET|ERRORSECRET|CREATESECRET|MFAREQUIREDSECRET/i,
  /otpauth/i,
  /request[_ -]?id/i,
  /\bstack\b/i,
  /\btraceback\b/i,
  /\/(?:home|var|tmp|usr|app|workspace)\//i,
  /\bselect\b[\s\S]{0,80}\bfrom\b/i,
  /\binsert\b[\s\S]{0,80}\binto\b/i,
  /\bupdate\b[\s\S]{0,80}\bset\b/i,
  /internal_path/i,
  /raw_request/i,
];

async function clearBrowserSession(page: Page) {
  await page.context().clearCookies();
}

async function hasSessionCookie(page: Page) {
  const cookies = await page.context().cookies(apiBase);
  return cookies.some((cookie) => cookie.name === sessionCookieName);
}

async function expectAllInteractiveControlsNamed(page: Page) {
  const unnamedControls = await page.locator("body").evaluate((body) => {
    function isVisible(element: Element) {
      const htmlElement = element as HTMLElement;
      const style = window.getComputedStyle(htmlElement);
      return (
        style.visibility !== "hidden" &&
        style.display !== "none" &&
        htmlElement.getClientRects().length > 0
      );
    }

    function labelledByText(element: Element) {
      const labelledBy = element.getAttribute("aria-labelledby") ?? "";
      return labelledBy
        .split(/\s+/u)
        .filter(Boolean)
        .map((id) => document.getElementById(id)?.textContent?.trim() ?? "")
        .join(" ")
        .trim();
    }

    function labelText(element: Element) {
      if (
        element instanceof HTMLInputElement ||
        element instanceof HTMLSelectElement ||
        element instanceof HTMLTextAreaElement
      ) {
        return Array.from(element.labels ?? [])
          .map((label) => label.textContent?.trim() ?? "")
          .join(" ")
          .trim();
      }
      return "";
    }

    function accessibleName(element: Element) {
      return (
        element.getAttribute("aria-label")?.trim() ||
        labelledByText(element) ||
        labelText(element) ||
        element.textContent?.trim() ||
        element.getAttribute("title")?.trim() ||
        element.getAttribute("placeholder")?.trim() ||
        ""
      );
    }

    return Array.from(
      body.querySelectorAll(
        'button:not([disabled]), a[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [role="button"]:not([disabled]), [tabindex]:not([tabindex="-1"])',
      ),
    )
      .filter(isVisible)
      .filter((element) => accessibleName(element) === "")
      .map((element) => {
        const tag = element.tagName.toLowerCase();
        const testId = element.getAttribute("data-testid") ?? "";
        const role = element.getAttribute("role") ?? "";
        return `${tag}${testId ? ` data-testid:${testId}` : ""}${role ? ` role:${role}` : ""}`;
      });
  });
  expect(unnamedControls).toEqual([]);
}

async function activeTestId(page: Page) {
  return page.evaluate(() => {
    const active = document.activeElement;
    if (!(active instanceof HTMLElement)) {
      return "";
    }
    return active.closest("[data-testid]")?.getAttribute("data-testid") ?? "";
  });
}

async function activeElementSignature(page: Page) {
  return page.evaluate(() => {
    const active = document.activeElement;
    if (!(active instanceof HTMLElement) || active === document.body) {
      return "";
    }
    const testId =
      active.closest("[data-testid]")?.getAttribute("data-testid") ?? "";
    const id = active.id === "" ? "" : `#${active.id}`;
    const tag = active.tagName.toLowerCase();
    const role = active.getAttribute("role") ?? "";
    const nameParts = [
      active.getAttribute("aria-label")?.trim() ?? "",
      active.getAttribute("title")?.trim() ?? "",
      active.textContent?.trim() ?? "",
    ].filter(Boolean);
    return [
      tag,
      id,
      testId === "" ? "" : ` data-testid:${testId}`,
      role === "" ? "" : ` role:${role}`,
      nameParts.length === 0 ? "" : `:${nameParts.join(" ")}`,
    ].join("");
  });
}

async function focusKeyboardSentinel(page: Page) {
  await page.evaluate((sentinelId) => {
    let sentinel = document.getElementById(sentinelId);
    if (sentinel === null) {
      sentinel = document.createElement("button");
      sentinel.id = sentinelId;
      sentinel.setAttribute("aria-label", "Keyboard traversal sentinel");
      sentinel.setAttribute("type", "button");
      sentinel.style.position = "fixed";
      sentinel.style.inlineSize = "1px";
      sentinel.style.blockSize = "1px";
      sentinel.style.opacity = "0";
      sentinel.style.pointerEvents = "none";
      document.body.prepend(sentinel);
    }
    sentinel.focus();
  }, keyboardSentinelId);
}

async function removeKeyboardSentinel(page: Page) {
  await page.evaluate((sentinelId) => {
    document.getElementById(sentinelId)?.remove();
  }, keyboardSentinelId);
}

async function expectTabOrderIncludes(
  page: Page,
  testIds: readonly string[],
  maxTabs = 180,
) {
  await focusKeyboardSentinel(page);
  try {
    const remaining = new Set<string>(testIds);
    for (let index = 0; index < maxTabs && remaining.size > 0; index += 1) {
      await page.keyboard.press("Tab");
      remaining.delete(await activeTestId(page));
    }
    expect([...remaining]).toEqual([]);
  } finally {
    await removeKeyboardSentinel(page);
  }
}

async function expectNoFocusTrap(page: Page) {
  await focusKeyboardSentinel(page);
  try {
    await page.keyboard.press("Tab");
    const first = await activeElementSignature(page);
    await page.keyboard.press("Tab");
    const second = await activeElementSignature(page);
    expect(first).not.toBe("");
    expect(second).not.toBe("");
    expect(second).not.toBe(first);
    await page.keyboard.press("Shift+Tab");
    expect(await activeElementSignature(page)).toBe(first);
  } finally {
    await removeKeyboardSentinel(page);
  }
}

async function expectTabTraversalAdvancesFrom(
  page: Page,
  origin: Locator,
  tabCount = 4,
) {
  await expectVisibleFocus(origin);
  const visited: string[] = [];
  for (let index = 0; index < tabCount; index += 1) {
    await page.keyboard.press("Tab");
    visited.push(await activeElementSignature(page));
  }
  expect(visited.every((signature) => signature !== "")).toBeTruthy();
  expect(new Set(visited).size).toBeGreaterThan(1);
}

async function expectVisibleFocus(locator: Locator) {
  await locator.focus();
  await expect(locator).toBeFocused();
  await expect
    .poll(() =>
      locator.evaluate((element) => {
        const style = window.getComputedStyle(element);
        const outlineVisible =
          style.outlineStyle !== "none" &&
          style.outlineWidth !== "0px" &&
          style.outlineColor !== "transparent";
        const shadowVisible = style.boxShadow !== "none";
        return outlineVisible || shadowVisible;
      }),
    )
    .toBeTruthy();
}

async function expectStatusRole(locator: Locator) {
  await expect(locator).toBeVisible();
  await expect(locator).toHaveAttribute("role", "status");
  await expect(locator).not.toHaveText("");
}

async function expectAlertRole(locator: Locator) {
  await expect(locator).toBeVisible();
  await expect(locator).toHaveAttribute("role", "alert");
  await expect(locator).not.toHaveText("");
}

async function expectNoPrivateDiagnostics(locator: Locator) {
  const text = (await locator.textContent()) ?? "";
  for (const pattern of privateDiagnosticPatterns) {
    expect(text).not.toMatch(pattern);
  }
}

function contrastRecordPath(title: string) {
  const dir = process.env.CARTULARY_FRONTEND_ACCESSIBILITY_CONTRAST_DIR;
  if (!dir) {
    return "";
  }
  const slug = title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "")
    .slice(0, 120);
  mkdirSync(dir, { recursive: true });
  return path.join(dir, `${slug}.json`);
}

async function collectContrastChecks(page: Page, testIds: readonly string[]) {
  const targets = [...new Set(testIds)].map((id) => ({
    id,
    selector: dataTestIdSelector(id),
  }));
  return page.evaluate(
    ({ targets, threshold }) => {
      type Rgba = { a: number; b: number; g: number; r: number };

      function parseRgba(value: string): Rgba | null {
        const match =
          /^rgba?\((\d+(?:\.\d+)?),\s*(\d+(?:\.\d+)?),\s*(\d+(?:\.\d+)?)(?:,\s*(\d+(?:\.\d+)?))?\)$/i.exec(
            value.trim(),
          );
        if (!match) {
          return null;
        }
        return {
          r: Number(match[1]),
          g: Number(match[2]),
          b: Number(match[3]),
          a: match[4] === undefined ? 1 : Number(match[4]),
        };
      }

      function rgbText(color: Rgba) {
        return `rgb(${Math.round(color.r)}, ${Math.round(color.g)}, ${Math.round(color.b)})`;
      }

      function channel(value: number) {
        const normalized = value / 255;
        if (normalized <= 0.03928) {
          return normalized / 12.92;
        }
        return ((normalized + 0.055) / 1.055) ** 2.4;
      }

      function luminance(color: Rgba) {
        return (
          0.2126 * channel(color.r) +
          0.7152 * channel(color.g) +
          0.0722 * channel(color.b)
        );
      }

      function contrastRatio(foreground: Rgba, background: Rgba) {
        const lighter = Math.max(luminance(foreground), luminance(background));
        const darker = Math.min(luminance(foreground), luminance(background));
        return (lighter + 0.05) / (darker + 0.05);
      }

      function backgroundFor(element: Element) {
        let candidate: Element | null = element;
        while (candidate) {
          const color = parseRgba(
            window.getComputedStyle(candidate).backgroundColor,
          );
          if (color && color.a > 0) {
            return color;
          }
          candidate = candidate.parentElement;
        }
        return { a: 1, b: 255, g: 255, r: 255 };
      }

      return targets
        .map(({ id, selector }) => {
          const element = document.querySelector(selector);
          if (!(element instanceof HTMLElement)) {
            return null;
          }
          const style = window.getComputedStyle(element);
          if (
            style.display === "none" ||
            style.visibility === "hidden" ||
            element.getClientRects().length === 0
          ) {
            return null;
          }
          const foreground = parseRgba(style.color);
          const background = backgroundFor(element);
          if (!foreground) {
            return null;
          }
          const ratio = contrastRatio(foreground, background);
          return {
            background: rgbText(background),
            foreground: rgbText(foreground),
            ratio: Math.round(ratio * 100) / 100,
            result: ratio >= threshold ? "pass" : "fail",
            target: id,
            threshold,
          };
        })
        .filter(Boolean);
    },
    { targets, threshold: contrastThreshold },
  );
}

async function expectAndRecordContrast(page: Page, testIds: readonly string[]) {
  const title = test.info().title;
  const checks = await collectContrastChecks(page, testIds);
  expect(checks.length).toBeGreaterThan(0);
  expect(checks.filter((check) => check?.result !== "pass")).toEqual([]);

  const recordPath = contrastRecordPath(title);
  if (recordPath) {
    writeFileSync(
      recordPath,
      `${JSON.stringify(
        {
          scenario_title: title,
          checks,
        },
        null,
        2,
      )}\n`,
      "utf8",
    );
  }
}

async function expectCellTextOrValue(locator: Locator, value: string) {
  const mode = await locator.evaluate((element) => {
    if (
      element instanceof HTMLInputElement ||
      element instanceof HTMLTextAreaElement ||
      element instanceof HTMLSelectElement
    ) {
      return "value";
    }
    return "text";
  });
  if (mode === "value") {
    await expect(locator).toHaveValue(value);
    return;
  }
  await expect(locator).toContainText(value);
}

async function openA11ySystemSurface(
  page: Page,
  options: {
    groupToken: SystemViewSwitcherGroupToken;
    viewSchemaId: string;
  },
) {
  const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
  await expectVisibleFocus(trigger);
  await trigger.press("Enter");
  await expect(trigger).toHaveAttribute("aria-expanded", "true");
  const menu = page.getByTestId(systemViewSwitcherMenuTestId());
  await expect(menu).toBeVisible();
  await expect(menu).toHaveAttribute("role", "menu");

  const option = page.getByTestId(
    systemViewSwitcherOptionTestId(options.groupToken, options.viewSchemaId),
  );
  await expect(option).toHaveAttribute("role", "menuitemradio");
  await expect(option).toHaveAttribute(
    "data-view-schema-id",
    options.viewSchemaId,
  );
  await expect(option).not.toHaveText("");
  await expectVisibleFocus(option);
  await option.press("Enter");
  await expect(menu).toHaveCount(0);
  await expect(trigger).toHaveAttribute("aria-expanded", "false");
  await expect(trigger).toHaveAttribute(
    "data-view-schema-id",
    options.viewSchemaId,
  );
  await expect(
    page.getByTestId(gridShellTestId(options.viewSchemaId)),
  ).toBeVisible();
}

function phase9A11yAttachedEvidencePayload(recordId: string) {
  return {
    kind: "collection_actions_v1",
    actions: [
      {
        op: "add_record_ref",
        linked_record_id: recordId,
      },
    ],
  };
}

function phase9A11yHistoryActionTestId(
  item: Phase9A11yHistoryItem,
  action: Phase9A11yHistoryItem["available_rollback_actions"][number],
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

function phase9A11yRollbackAnchor(
  item: Phase9A11yHistoryItem,
  action: Phase9A11yHistoryItem["available_rollback_actions"][number],
) {
  return {
    action,
    historyItemRef: item.history_item_ref,
  };
}

function requirePhase9A11yHistoryEntryAction(history: Phase9A11yHistoryData) {
  const item =
    history.items.find(
      (candidate) =>
        candidate.available_rollback_actions.includes("history_entry") &&
        typeof candidate.history_entry_ref === "string" &&
        candidate.history_entry_ref.length > 0,
    ) ?? null;
  if (item === null) {
    throw new Error("missing FE-A11Y-P9 history_entry rollback item");
  }
  return item;
}

async function fetchPhase9A11yRecordHistory(page: Page, recordId: string) {
  const response = await page.request.get(
    `${apiBase}/api/v1/records/${recordId}/history`,
    { headers: await csrfHeaders(page) },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: Phase9A11yHistoryData }).data;
}

async function expectP1SurfaceA11y(
  page: Page,
  options: {
    focusTestId?: string;
    tabStops?: readonly string[];
  } = {},
) {
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  if (options.focusTestId) {
    await expectVisibleFocus(page.getByTestId(options.focusTestId));
  }
  if (options.tabStops && options.tabStops.length > 0) {
    await expectTabOrderIncludes(page, options.tabStops);
  }
  await expectAndRecordContrast(page, [
    ...(options.focusTestId ? [options.focusTestId] : []),
    ...(options.tabStops ?? []),
  ]);
}

function phase1A11yAppLocalTestId(
  key: keyof typeof phase1A11yAppLocalSelectors,
): Phase1A11yAppLocalTestId {
  const entry = phase1A11yAppLocalSelectors[key];
  if (
    entry.owner.trim() === "" ||
    entry.reason.trim() === "" ||
    entry.scope.trim() === ""
  ) {
    throw new Error(`missing app-local selector ownership for ${key}`);
  }
  return entry.testId;
}

async function loadIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
) {
  const response = await authRequests.get(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { memberships: IncidentMembershipRecord[] };
  };
  const membership =
    body.data.memberships.find((candidate) => candidate.user_id === userId) ??
    null;
  if (membership === null) {
    throw new Error(`missing incident membership for ${userId}`);
  }
  return membership;
}

async function createIncidentWithRequest(
  authRequests: APIRequestContext,
  incidentKey: string,
  title: string,
) {
  const response = await authRequests.post(`${apiBase}/api/v1/incidents`, {
    data: {
      client_txn_id: uniqueTxn("a11y-incident"),
      incident_key: incidentKey,
      title,
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as { data: { incident_id: string } };
  return body.data.incident_id;
}

async function createIncidentMembershipWithRequest(
  authRequests: APIRequestContext,
  incidentId: string,
  email: string,
  role: string,
) {
  const response = await authRequests.post(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
    {
      data: {
        client_txn_id: uniqueTxn("a11y-membership"),
        email,
        role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

async function patchIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  options: {
    baseMembershipVersion: number;
    role: string;
    userId: string;
  },
  headers: Record<string, string> = {},
) {
  const response = await authRequests.patch(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${options.userId}`,
    {
      headers,
      data: {
        base_membership_version: options.baseMembershipVersion,
        role: options.role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: IncidentMembershipRecord }).data;
}

async function deleteIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
  baseMembershipVersion: number,
  headers: Record<string, string> = {},
) {
  const response = await authRequests.delete(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${userId}`,
    {
      headers,
      data: {
        base_membership_version: baseMembershipVersion,
      },
    },
  );
  expect(response.status()).toBe(204);
}

type EvidenceA11yRowStateKey =
  | "available"
  | "blocked"
  | "pending_upload"
  | "requested";

async function createA11yEvidenceRow(
  page: Page,
  incidentId: string,
  options: {
    lifecycleState: string;
    requestedAt: string;
    storageRef: string;
    title: string;
    txnPrefix: string;
  },
): Promise<ViewRow> {
  return createEvidenceFixtureRow(page, incidentId, {
    collectorPartyText: "FE-P6 accessibility fixture",
    ...options,
  });
}

async function createUploadedA11yEvidence(
  page: Page,
  incidentId: string,
  options: EvidenceUploadOptions,
): Promise<ViewRow> {
  return createUploadedEvidenceFixture(page, incidentId, {
    collectorPartyText: "FE-P6 accessibility fixture",
    ...options,
  });
}

async function expectEvidenceAccessState(
  page: Page,
  recordId: string,
  stateKey: EvidenceA11yRowStateKey,
  options: {
    liveRole?: "alert" | "status";
    messageText?: RegExp | string;
  } = {},
) {
  await expect(evidenceAccessStateContainer(page, recordId)).toHaveAttribute(
    "data-evidence-state-key",
    stateKey,
  );
  const previewButton = page.getByTestId(evidencePreviewButtonTestId(recordId));
  const downloadButton = page.getByTestId(
    evidenceDownloadButtonTestId(recordId),
  );
  await expect(previewButton).toHaveText("Preview");
  await expect(downloadButton).toHaveText("Download");
  if (stateKey === "available") {
    await expect(previewButton).toBeEnabled();
    await expect(downloadButton).toBeEnabled();
  } else {
    await expect(previewButton).toBeDisabled();
    await expect(downloadButton).toBeDisabled();
  }
  if (options.messageText !== undefined || options.liveRole !== undefined) {
    const message = page.getByTestId(evidenceAccessMessageTestId(recordId));
    await expect(message).toBeVisible();
    if (options.messageText !== undefined) {
      await expect(message).toContainText(options.messageText);
    }
    if (options.liveRole !== undefined) {
      await expect(message).toHaveAttribute("role", options.liveRole);
      await expect(message).toHaveAttribute(
        "aria-live",
        options.liveRole === "alert" ? "assertive" : "polite",
      );
    }
  }
}

function evidenceAccessStateContainer(page: Page, recordId: string): Locator {
  return page
    .getByTestId(evidencePreviewButtonTestId(recordId))
    .locator("xpath=ancestor::*[@data-evidence-state-key][1]");
}

async function armA11yPublicErrorFault(
  page: Page,
  options: {
    path: string;
    reasonCode: "blob_failed" | "evidence_inconsistent";
  },
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/test/runtime/public-error-faults`,
    {
      headers: testRouteHeaders(),
      data: {
        code: "evidence_access_unavailable",
        consume_once: true,
        details: {
          reason_code: options.reasonCode,
        },
        message: "Evidence access failed for FE-P6 accessibility fixture.",
        method: "POST",
        path: options.path,
        retryable: false,
        status: 409,
      },
    },
  );
  expect(response.status()).toBe(201);
}

async function holdSinglePublicAPIResponse(
  page: Page,
  options: {
    method: string;
    path: string;
  },
) {
  const routePattern = `**${options.path}`;
  const expectedMethod = options.method.toUpperCase();
  let releaseHold: (() => void) | null = null;
  let resolveHit: (() => void) | null = null;
  const waitForHit = new Promise<void>((resolve) => {
    resolveHit = resolve;
  });
  const hold = new Promise<void>((resolve) => {
    releaseHold = resolve;
  });
  let released = false;

  const routeHandler = async (route: Route) => {
    if (route.request().method().toUpperCase() !== expectedMethod || released) {
      await route.fallback();
      return;
    }
    released = true;
    resolveHit?.();
    await hold;
    const response = await route.fetch();
    await route.fulfill({ response });
  };

  await page.route(routePattern, routeHandler);
  return {
    dispose: async () => {
      releaseHold?.();
      await safeUnroute(page, routePattern, routeHandler);
    },
    release: () => {
      releaseHold?.();
    },
    waitForHit,
  };
}

async function fulfillPublicError(
  route: Route,
  options: {
    code: string;
    details?: Record<string, unknown>;
    message?: string;
    status: number;
  },
) {
  await route.fulfill({
    contentType: "application/json",
    status: options.status,
    body: JSON.stringify({
      error: {
        code: options.code,
        message: options.message,
        status: options.status,
        request_id: "raw-request-id-must-not-render",
        details: {
          reason_code: "safe_reason",
          ...(options.details ?? {}),
          bootstrap_token: "private-bootstrap-token",
          internal_path: "/home/cartulary/private/file.go",
          raw_request_id: "raw-request-id-must-not-render",
          secret_base32: "PRIVATESECRETBASE32",
          sql: "select * from private_table",
          stack: "stack frame at /home/cartulary/private/file.go:1",
        },
      },
    }),
  });
}

test.describe("FE-P2 accessibility readiness", () => {
  test(p2AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP2"),
      "FE-A11Y-P2-01 workbook shell",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p2-01-row"),
        "timeline.occurred_at": "2026-05-31T09:00:00Z",
        "timeline.summary": "FE-P2 accessibility shell row",
        "timeline.details": "Inspector control coverage",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    const shell = page.getByTestId(workbookShellReadyTestId());
    await expect(shell).toBeVisible();
    await expect(
      page.getByRole("region", { name: "Workbook shell" }),
    ).toHaveCount(1);

    for (const slot of workbookShellSlots) {
      const label = workbookShellSlotLabel(slot);
      const slotByTestId = shell.locator(
        dataTestIdSelector(workbookShellSlotTestId(slot)),
      );
      await expect(slotByTestId).toBeVisible();
      await expect(slotByTestId).toHaveAttribute("aria-label", label);
      await expect(shell.getByRole("region", { name: label })).toHaveCount(1);
    }

    for (const surface of requiredBuiltInWorkbookSurfaceIds) {
      const tab = page.getByTestId(surfaceTabTestId(surface));
      await expect(tab).toBeVisible();
      await expect(tab).not.toHaveText("");
    }
    await expect(
      page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
    ).toHaveAttribute("aria-current", "page");

    const trigger = page.getByTestId(systemViewSwitcherTriggerTestId());
    await expect(trigger).toBeVisible();
    await expect(trigger).toHaveAttribute("aria-label", "System views");
    await expectVisibleFocus(trigger);
    await page.keyboard.press("Enter");

    const menu = page.getByTestId(systemViewSwitcherMenuTestId());
    await expect(menu).toBeVisible();
    await expect(menu).toHaveAttribute("role", "menu");
    const indicatorOption = page.getByTestId(
      systemViewSwitcherOptionTestId(
        "scope-assessment",
        indicatorsViewSchemaId,
      ),
    );
    await expect(indicatorOption).toBeFocused();
    await expect(indicatorOption).toHaveAttribute("role", "menuitemradio");
    await expect(indicatorOption).toHaveAttribute("aria-checked", "false");
    await page.keyboard.press("Escape");
    await expect(menu).toHaveCount(0);
    await expect(trigger).toBeFocused();

    await expect(
      page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(gridFilterFieldTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(gridFilterApplyTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
    ).toBeVisible();

    const inspectButton = page.getByTestId(
      rowInspectButtonTestId(timelineRow.record_id),
    );
    await expect(inspectButton).toBeVisible();
    await expectVisibleFocus(inspectButton);
    await inspectButton.click();

    const inspector = page.getByTestId("timeline-inspector");
    await expect(inspector).toBeVisible();
    await expect(inspector).toHaveAttribute("aria-label", "Timeline inspector");
    await expect(
      page.getByTestId(
        rowInspectorFieldTestId(timelineRow.record_id, "timeline.details"),
      ),
    ).toBeVisible();

    const saveState = page.getByTestId(saveStateTestId());
    await expectStatusRole(saveState);
    await expect(saveState).toHaveText("Saved");

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      workbookShellSlotTestId("top-bar"),
      workbookShellSlotTestId("status-strip"),
      systemViewSwitcherTriggerTestId(),
      savedViewSelectorTestId(timelineViewSchemaId),
      gridFilterApplyTestId(timelineViewSchemaId),
      rowInspectButtonTestId(timelineRow.record_id),
      saveStateTestId(),
    ]);
  });
});

test.describe("FE-P3 accessibility readiness", () => {
  test(p3AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP3"),
      "FE-A11Y-P3-01 grid adapter",
    );
    const alphaRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p3-01-alpha"),
        "timeline.occurred_at": "2026-05-31T10:00:00Z",
        "timeline.summary": "Alpha accessibility row",
        "timeline.details": "Keyboard grid coverage",
      },
    )) as ViewRow;
    const betaRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p3-01-beta"),
        "timeline.occurred_at": "2026-05-31T10:05:00Z",
        "timeline.summary": "Beta accessibility row",
        "timeline.details": "Grouped grid coverage",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(
      page.getByTestId(rowCellTestId(alphaRow.record_id, "timeline.summary")),
    ).toHaveValue("Alpha accessibility row");

    const betaMarkReviewed = page.getByTestId(
      timelineRowMarkReviewedButtonTestId(betaRow.record_id),
    );
    await expectVisibleFocus(betaMarkReviewed);
    await betaMarkReviewed.click();
    await expect(betaMarkReviewed).toBeDisabled();
    await expect(
      page.getByTestId(
        rowCellTestId(betaRow.record_id, "timeline.capture_state"),
      ),
    ).toHaveText("reviewed");

    await page
      .getByTestId(gridGroupingSelectTestId(timelineViewSchemaId))
      .selectOption("timeline.capture_state");
    const reviewedGroup = page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    );
    await expect(reviewedGroup).toBeVisible();
    await expect(reviewedGroup).toContainText("reviewed");

    const betaSummary = page.getByTestId(
      rowCellTestId(betaRow.record_id, "timeline.summary"),
    );
    await expect(betaSummary).toHaveAttribute(
      "aria-label",
      `Summary ${betaRow.record_id}`,
    );
    await expectVisibleFocus(betaSummary);
    await betaSummary.fill("Beta accessibility active edit");
    await expect(betaSummary).toHaveValue("Beta accessibility active edit");
    await expectStatusRole(page.getByTestId(saveStateTestId()));

    await expect(
      page.getByTestId(
        gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"),
      ),
    ).toContainText("Summary");
    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      gridGroupingSelectTestId(timelineViewSchemaId),
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
      gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"),
      rowCellTestId(betaRow.record_id, "timeline.summary"),
      timelineRowMarkReviewedButtonTestId(betaRow.record_id),
      saveStateTestId(),
    ]);
  });
});

test.describe("FE-P4 accessibility readiness", () => {
  test(p4AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP4"),
      "FE-A11Y-P4-01 Timeline accessibility",
    );
    const editRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p4-01-edit"),
        "timeline.occurred_at": "2026-06-03T10:00:00Z",
        "timeline.summary": "FE-P4 edit accessibility row",
        "timeline.details": "Escape priority details",
      },
    )) as ViewRow;
    const pasteRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p4-01-paste"),
        "timeline.occurred_at": "2026-06-03T10:05:00Z",
        "timeline.summary": "FE-P4 paste accessibility row",
      },
    )) as ViewRow;
    const pendingRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p4-01-pending"),
        "timeline.occurred_at": "2026-06-03T10:10:00Z",
        "timeline.summary": "FE-P4 pending accessibility row",
      },
    )) as ViewRow;
    const validationRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p4-01-validation"),
        "timeline.occurred_at": "2026-06-03T10:15:00Z",
        "timeline.summary": "FE-P4 validation accessibility row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expectStatusRole(page.getByTestId(saveStateTestId()));
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectTabOrderIncludes(page, [
      gridFilterFieldTestId(timelineViewSchemaId),
      gridFilterApplyTestId(timelineViewSchemaId),
      gridGroupingSelectTestId(timelineViewSchemaId),
      rowInspectButtonTestId(editRow.record_id),
    ]);

    const editSummary = page.getByTestId(
      rowCellTestId(editRow.record_id, "timeline.summary"),
    );
    await expect(editSummary).toHaveAttribute(
      "aria-label",
      `Summary ${editRow.record_id}`,
    );
    await expectVisibleFocus(editSummary);
    await editSummary.fill("FE-P4 accessibility committed edit");
    await editSummary.press("Enter");
    await expect(editSummary).toHaveValue("FE-P4 accessibility committed edit");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

    await pasteGridMatrix({
      fieldKey: "timeline.summary",
      matrix: [["FE-P4 accessibility pasted summary", "a11y-host.example"]],
      page,
      recordId: pasteRow.record_id,
      surface: timelineViewSchemaId,
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(rowCellTestId(pasteRow.record_id, "timeline.summary")),
    ).toHaveValue("FE-P4 accessibility pasted summary");

    const patchController = await installPatchTransportFailureController(page);
    try {
      patchController.disconnect();
      const pendingSummary = page.getByTestId(
        rowCellTestId(pendingRow.record_id, "timeline.summary"),
      );
      await expectVisibleFocus(pendingSummary);
      await pendingSummary.fill("FE-P4 accessibility pending replay");
      await pendingSummary.press("Enter");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
      await expectStatusRole(page.getByTestId(pendingQueueNoticeTestId()));
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "1",
      );

      patchController.connect();
      await expect
        .poll(() => successfulPatchCalls(patchController.calls).length)
        .toBe(1);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    } finally {
      patchController.connect();
      await patchController.dispose();
    }

    const originSummary = page.getByTestId(
      rowCellTestId(editRow.record_id, "timeline.summary"),
    );
    await expectVisibleFocus(originSummary);
    const inspectButton = page.getByTestId(
      rowInspectButtonTestId(editRow.record_id),
    );
    await expectVisibleFocus(inspectButton);
    await inspectButton.click();
    const inspectorDetails = page.getByTestId(
      rowInspectorFieldTestId(editRow.record_id, "timeline.details"),
    );
    await expectVisibleFocus(inspectorDetails);
    await page.keyboard.press("Escape");
    await expect(originSummary).toBeFocused();

    const validationCell = page.getByTestId(
      rowCellTestId(validationRow.record_id, "timeline.occurred_at"),
    );
    await expectVisibleFocus(validationCell);
    await validationCell.fill("not-a-timestamp");
    await validationCell.press("Enter");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    const validationNotice = page.getByTestId(pendingQueueNoticeTestId());
    await expectStatusRole(validationNotice);
    await expectNoPrivateDiagnostics(validationNotice);
    await expect(validationCell).toHaveValue("not-a-timestamp");

    await expect(
      page.getByTestId(
        gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"),
      ),
    ).toContainText("Summary");
    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      workbookShellSlotTestId("status-strip"),
      gridFilterApplyTestId(timelineViewSchemaId),
      gridGroupingSelectTestId(timelineViewSchemaId),
      gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"),
      rowCellTestId(editRow.record_id, "timeline.summary"),
      rowCellTestId(validationRow.record_id, "timeline.occurred_at"),
      rowInspectButtonTestId(editRow.record_id),
      pendingQueueNoticeTestId(),
      saveStateTestId(),
    ]);
  });
});

test.describe("FE-P5 accessibility readiness", () => {
  test(p5AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP5"),
      "FE-A11Y-P5-01 mention states",
    );
    const {
      autoRawText,
      autoRow,
      dismissedMention,
      dismissedRawText,
      dismissedRow,
      manualMention,
      manualRow,
      manualTarget,
      resolvedMention,
      resolvedRow,
      unresolvedMention,
      unresolvedRawText,
      unresolvedRow,
    } = await seedHostMentionStateFixture(page, incidentId, {
      displayPrefix: "FE-A11Y-P5",
      hostnamePrefix: "fe-a11y-p5",
      occurredAt: {
        auto: "2026-06-06T16:15:00Z",
        dismissed: "2026-06-06T16:20:00Z",
        manual: "2026-06-06T16:10:00Z",
        resolved: "2026-06-06T16:05:00Z",
        unresolved: "2026-06-06T16:00:00Z",
      },
      rawTextPrefix: "FEA11YP5",
      summary: {
        auto: "FE-A11Y-P5 auto chip",
        dismissed: "FE-A11Y-P5 dismissed chip",
        manual: "FE-A11Y-P5 manual chip",
        resolved: "FE-A11Y-P5 resolved chip",
        unresolved: "FE-A11Y-P5 unresolved chip",
      },
      txnPrefix: "fe-a11y-p5",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

    const unresolvedChip = page
      .getByTestId(
        relationshipItemsTestId(unresolvedRow.record_id, hostRefsFieldKey),
      )
      .getByTestId(relationshipChipTestId(String(unresolvedMention.item_ref)));
    await expect(unresolvedChip).toHaveAttribute(
      "aria-label",
      `Unresolved ${unresolvedRawText}`,
    );
    await expect(unresolvedChip).toContainText("Unresolved");
    await expectVisibleFocus(unresolvedChip);

    const resolvedChip = page
      .getByTestId(
        relationshipItemsTestId(resolvedRow.record_id, hostRefsFieldKey),
      )
      .getByTestId(relationshipChipTestId(String(resolvedMention.item_ref)));
    await expect(resolvedChip).toHaveAttribute(
      "aria-label",
      /^Resolved FE-A11Y-P5 Resolved Target$/u,
    );
    await expect(resolvedChip).toContainText("Resolved");
    await expectVisibleFocus(resolvedChip);

    await openTimelineInspector(page, manualRow.record_id);
    const manualMentionItem = page.getByTestId(
      mentionItemTestId(String(manualMention.item_ref)),
    );
    await expectVisibleFocus(manualMentionItem);
    await manualMentionItem.click();
    const resolveSelect = page.getByTestId(mentionResolveTargetSelectTestId());
    await expectVisibleFocus(resolveSelect);
    await resolveSelect.selectOption(manualTarget.record_id);
    const resolveButton = page.getByTestId(
      mentionResolveExistingButtonTestId(),
    );
    await expectVisibleFocus(resolveButton);
    await resolveButton.click();

    const manualChip = page
      .getByTestId(
        relationshipItemsTestId(manualRow.record_id, hostRefsFieldKey),
      )
      .getByTestId(relationshipChipTestId(String(manualMention.item_ref)));
    await expect(manualChip).toHaveAttribute(
      "aria-label",
      /^Resolved .*manual resolution/u,
    );
    await expect(manualChip).toContainText("Manual");
    await expectVisibleFocus(manualChip);
    await expectVisibleFocus(page.getByTestId(mentionDismissButtonTestId()));
    await expectVisibleFocus(
      page.getByTestId(mentionRestoreUnresolvedButtonTestId()),
    );

    const autoEnvelope = await addRelationshipTokenViaUI(
      page,
      autoRow.record_id,
      "hostRefs",
      autoRawText,
    );
    const autoItem = requireItemByRawText(
      collectionItems(autoEnvelope.data.row, hostRefsFieldKey),
      autoRawText,
    );
    const autoChip = page
      .getByTestId(relationshipItemsTestId(autoRow.record_id, hostRefsFieldKey))
      .getByTestId(relationshipChipTestId(String(autoItem.item_ref)));
    await expect(autoChip).toHaveAttribute(
      "aria-label",
      /^Auto-resolved .*matched/u,
    );
    await expect(autoChip).toContainText("Auto");
    await expectVisibleFocus(autoChip);
    const autoNotice = page.getByTestId(
      autoResolutionNoticeTestId(String(autoItem.item_ref)),
    );
    await expect(autoNotice).toContainText("FE-A11Y-P5 Auto Target");
    await expectVisibleFocus(
      autoNotice.getByTestId(
        autoResolutionUndoButtonTestId(String(autoItem.item_ref)),
      ),
    );

    await openTimelineInspector(page, dismissedRow.record_id);
    await page
      .getByTestId(mentionItemTestId(String(dismissedMention.item_ref)))
      .click();
    await page
      .getByTestId(mentionResolveTargetSelectTestId())
      .selectOption(manualTarget.record_id);
    await page.getByTestId(mentionResolveExistingButtonTestId()).click();
    await page.getByTestId(mentionDismissButtonTestId()).click();
    const dismissedMentionItem = page.getByTestId(
      mentionItemTestId(String(dismissedMention.item_ref)),
    );
    await expect(
      dismissedMentionItem.getByLabel(`Dismissed ${dismissedRawText}`),
    ).toBeVisible();
    await expect(dismissedMentionItem).toContainText("Dismissed");
    await expectVisibleFocus(dismissedMentionItem);
    await expectVisibleFocus(
      page.getByTestId(mentionRestoreUnresolvedButtonTestId()),
    );

    await expectTabOrderIncludes(page, [
      rowInspectButtonTestId(unresolvedRow.record_id),
      rowInspectButtonTestId(resolvedRow.record_id),
      rowInspectButtonTestId(manualRow.record_id),
      rowInspectButtonTestId(autoRow.record_id),
      rowInspectButtonTestId(dismissedRow.record_id),
    ]);
    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      relationshipChipTestId(String(unresolvedMention.item_ref)),
      relationshipChipTestId(String(resolvedMention.item_ref)),
      relationshipChipTestId(String(manualMention.item_ref)),
      relationshipChipTestId(String(autoItem.item_ref)),
      mentionItemTestId(String(dismissedMention.item_ref)),
      mentionRestoreUnresolvedButtonTestId(),
    ]);
  });
});

test.describe("FE-P6 accessibility readiness", () => {
  test(p6AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP6"),
      "FE-A11Y-P6-01 evidence access",
    );
    const requested = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "requested",
      requestedAt: "2026-06-07T10:00:00Z",
      storageRef: "case://fe-a11y-p6/requested",
      title: "01 requested evidence",
      txnPrefix: "fe-a11y-p6-requested",
    });
    const pending = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "pending_receipt",
      requestedAt: "2026-06-07T10:05:00Z",
      storageRef: "case://fe-a11y-p6/pending",
      title: "02 pending evidence",
      txnPrefix: "fe-a11y-p6-pending",
    });
    const blocked = await createA11yEvidenceRow(page, incidentId, {
      lifecycleState: "quarantined",
      requestedAt: "2026-06-07T10:10:00Z",
      storageRef: "case://fe-a11y-p6/quarantined",
      title: "03 quarantined evidence",
      txnPrefix: "fe-a11y-p6-blocked",
    });
    const availablePreview = await createUploadedA11yEvidence(
      page,
      incidentId,
      {
        body: Buffer.from("FE-A11Y-P6 preview evidence\n", "utf8"),
        contentType: "text/plain",
        filename: "fe-a11y-p6-preview.txt",
        requestedAt: "2026-06-07T10:15:00Z",
        title: "04 available preview evidence",
        txnPrefix: "fe-a11y-p6-preview",
      },
    );
    const downloadHandle = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from("FE-A11Y-P6 download evidence\n", "utf8"),
      contentType: "text/plain",
      filename: "fe-a11y-p6-download.txt",
      requestedAt: "2026-06-07T10:20:00Z",
      title: "05 download handle evidence",
      txnPrefix: "fe-a11y-p6-download",
    });
    const previewBlocked = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from(
        "<!doctype html><title>FE-A11Y-P6 unsupported preview</title>",
        "utf8",
      ),
      contentType: "text/html",
      filename: "fe-a11y-p6-preview-blocked.html",
      requestedAt: "2026-06-07T10:25:00Z",
      title: "06 preview blocked evidence",
      txnPrefix: "fe-a11y-p6-preview-blocked",
    });
    const failedHandle = await createUploadedA11yEvidence(page, incidentId, {
      body: Buffer.from("FE-A11Y-P6 failed handle evidence\n", "utf8"),
      contentType: "text/plain",
      filename: "fe-a11y-p6-failed.txt",
      requestedAt: "2026-06-07T10:30:00Z",
      title: "07 failed handle evidence",
      txnPrefix: "fe-a11y-p6-failed",
    });
    const inconsistentHandle = await createUploadedA11yEvidence(
      page,
      incidentId,
      {
        body: Buffer.from("FE-A11Y-P6 inconsistent handle evidence\n", "utf8"),
        contentType: "text/plain",
        filename: "fe-a11y-p6-inconsistent.txt",
        requestedAt: "2026-06-07T10:35:00Z",
        title: "08 inconsistent handle evidence",
        txnPrefix: "fe-a11y-p6-inconsistent",
      },
    );

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await expect(
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeVisible();

    await expectEvidenceAccessState(page, requested.record_id, "requested", {
      liveRole: "status",
      messageText: /^Requested:/u,
    });
    await expectEvidenceAccessState(page, pending.record_id, "pending_upload", {
      liveRole: "status",
      messageText: /pending/u,
    });
    await expectEvidenceAccessState(page, blocked.record_id, "blocked", {
      liveRole: "alert",
      messageText: /^Blocked:/u,
    });

    await expectEvidenceAccessState(
      page,
      availablePreview.record_id,
      "available",
    );
    const previewButton = page.getByTestId(
      evidencePreviewButtonTestId(availablePreview.record_id),
    );
    await expectVisibleFocus(
      page.getByTestId(
        evidenceDownloadButtonTestId(availablePreview.record_id),
      ),
    );
    await expectVisibleFocus(previewButton);
    await previewButton.click();
    await expect(
      page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
    ).toBeVisible();
    await expect(
      page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
    ).toHaveAttribute("title", /Evidence preview/u);
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
    ).toHaveAttribute("role", "status");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
    ).toHaveText("Preview loaded inline.");
    await page
      .getByTestId(evidencePreviewPanelTestId())
      .getByRole("button", { name: "Close" })
      .click();

    await expectEvidenceAccessState(
      page,
      downloadHandle.record_id,
      "available",
    );
    const downloadButton = page.getByTestId(
      evidenceDownloadButtonTestId(downloadHandle.record_id),
    );
    await expectVisibleFocus(
      page.getByTestId(evidencePreviewButtonTestId(downloadHandle.record_id)),
    );
    await expectVisibleFocus(downloadButton);
    const downloadPromise = page.waitForEvent("download");
    await downloadButton.click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toBe("fe-a11y-p6-download.txt");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
    ).toHaveAttribute("role", "status");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
    ).toHaveText("Download handle issued.");

    await expectEvidenceAccessState(
      page,
      previewBlocked.record_id,
      "available",
    );
    const unsupportedPreviewButton = page.getByTestId(
      evidencePreviewButtonTestId(previewBlocked.record_id),
    );
    await expectVisibleFocus(unsupportedPreviewButton);
    await expectVisibleFocus(
      page.getByTestId(evidenceDownloadButtonTestId(previewBlocked.record_id)),
    );
    await unsupportedPreviewButton.click();
    const previewBlockedMessage = page.getByTestId(
      evidenceAccessMessageTestId(previewBlocked.record_id),
    );
    await expect(previewBlockedMessage).toHaveAttribute("role", "alert");
    await expect(previewBlockedMessage).toContainText(
      "evidence_access_unavailable: unsupported_preview",
    );
    await expectNoPrivateDiagnostics(previewBlockedMessage);

    await armA11yPublicErrorFault(page, {
      path: `/api/v1/evidence-records/${failedHandle.record_id}/preview-handle`,
      reasonCode: "blob_failed",
    });
    await page
      .getByTestId(evidencePreviewButtonTestId(failedHandle.record_id))
      .click();
    const failedMessage = page.getByTestId(
      evidenceAccessMessageTestId(failedHandle.record_id),
    );
    await expect(failedMessage).toHaveAttribute("role", "alert");
    await expect(failedMessage).toContainText(
      "evidence_access_unavailable: blob_failed",
    );
    await expectNoPrivateDiagnostics(failedMessage);

    await armA11yPublicErrorFault(page, {
      path: `/api/v1/evidence-records/${inconsistentHandle.record_id}/preview-handle`,
      reasonCode: "evidence_inconsistent",
    });
    await page
      .getByTestId(evidencePreviewButtonTestId(inconsistentHandle.record_id))
      .click();
    const inconsistentMessage = page.getByTestId(
      evidenceAccessMessageTestId(inconsistentHandle.record_id),
    );
    await expect(inconsistentMessage).toHaveAttribute("role", "alert");
    await expect(inconsistentMessage).toContainText(
      "evidence_access_unavailable: evidence_inconsistent",
    );
    await expectNoPrivateDiagnostics(inconsistentMessage);

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      evidenceAccessMessageTestId(requested.record_id),
      evidenceAccessMessageTestId(pending.record_id),
      evidenceAccessMessageTestId(blocked.record_id),
      evidencePreviewButtonTestId(availablePreview.record_id),
      evidenceDownloadButtonTestId(downloadHandle.record_id),
      evidenceAccessMessageTestId(availablePreview.record_id),
      evidenceAccessMessageTestId(downloadHandle.record_id),
      evidenceAccessMessageTestId(previewBlocked.record_id),
      evidenceAccessMessageTestId(failedHandle.record_id),
      evidenceAccessMessageTestId(inconsistentHandle.record_id),
    ]);
  });
});

test.describe("FE-P7 accessibility readiness", () => {
  test(
    p7AccessibilityScenarioTitles[0],
    async ({ browser, page, sessionTracker }) => {
      await page.setViewportSize({ width: 1440, height: 900 });
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11YP7"),
        "FE-A11Y-P7 conflict accessibility",
      );
      const remote = await createIncidentMemberUser(page, incidentId, {
        display_name: "Accessible Analyst",
        email: uniqueEmail("fe-a11y-p7-remote"),
        initial_password: "FeA11yP7RemotePass!",
        role: "editor",
      });
      const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("fe-a11y-p7-row"),
        "timeline.summary": "FE-A11Y-P7 conflict base",
      });
      const recordId = requireRecordId(row);
      const patchController = await installPatchController(page);
      const primarySocket = installIncidentSocketMonitor(page, incidentId);

      let remotePage: Page | null = null;
      try {
        await page.goto(`/?incident_id=${incidentId}`);
        await expect(
          page.getByTestId(workbookShellReadyTestId()),
        ).toBeVisible();
        await primarySocket.waitForAcceptedSocket();

        const remoteSession = await openIncidentAsTrackedUserReady(
          browser,
          sessionTracker,
          {
            createdBy: "FE-A11Y-P7-01",
            email: remote.email,
            incidentId,
            password: remote.initial_password,
            purpose: "FE-A11Y-P7 remote presence analyst",
            readyRecordId: recordId,
            userId: remote.user_id,
          },
        );
        remotePage = remoteSession.page;
        await focusRemoteTimelineCellAndWaitForPresence({
          actorText: "AA",
          fieldKey: "timeline.summary",
          primaryPage: page,
          recordId,
          remotePage,
          socketMonitor: primarySocket,
        });

        await driveRealTimelineSummaryConflict({
          baseRowVersion: 1,
          localValue: "FE-A11Y-P7 local draft",
          page,
          patchController,
          recordId,
          remotePatchPage: remotePage,
          remoteValue: "FE-A11Y-P7 saved value",
          txnPrefix: "fea11yp7-conflict",
        });
        const resolver = page.getByTestId("conflict-resolver");
        await expect(resolver).toHaveAttribute(
          "aria-label",
          "Same-field conflict resolver",
        );
        const summary = page.getByTestId("conflict-resolver-summary");
        await expect(summary).toBeFocused();
        await expect(page.getByTestId("conflict-field-key")).toHaveValue(
          "timeline.summary",
        );
        await expect(page.getByTestId("conflict-server-value")).toHaveValue(
          "FE-A11Y-P7 saved value",
        );
        await expect(page.getByTestId("conflict-local-value")).toHaveValue(
          "FE-A11Y-P7 local draft",
        );
        await expect(page.getByRole("button", { name: "Close" })).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Keep saved value" }),
        ).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Use my unsaved value" }),
        ).toBeVisible();
        await expect(
          page.getByRole("button", { name: "Use merged value" }),
        ).toBeVisible();

        await summary.press("Enter");
        await expect(resolver).toBeVisible();
        await expect(page.getByTestId(saveStateTestId())).toHaveText(
          "Conflict",
        );
        await expect(
          page.getByTestId(conflictMarkerTestId(recordId, "timeline.summary")),
        ).toBeVisible();
        await expectAllInteractiveControlsNamed(page);
        await expectNoFocusTrap(page);
        await expectAndRecordContrast(page, [
          saveStateTestId(),
          rowPresenceMarkerTestId(recordId),
          cellPresenceMarkerTestId(recordId, "timeline.summary"),
          "conflict-close",
          "conflict-keep-saved",
          "conflict-use-unsaved",
          "conflict-use-merged",
        ]);

        await page.keyboard.press("Escape");
        await expect(resolver).toHaveCount(0);
        await expect(page.getByTestId(saveStateTestId())).toHaveText(
          "Conflict",
        );
        await expect(
          page.getByTestId(conflictMarkerTestId(recordId, "timeline.summary")),
        ).toBeVisible();
      } finally {
        await patchController.dispose();
        await remotePage?.context().close();
      }
    },
  );
});

test.describe("FE-P8 accessibility readiness", () => {
  test(p8AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP8"),
      "FE-A11Y-P8 query controls",
    );
    const reviewedRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p8-reviewed"),
        "timeline.summary": "FE-A11Y-P8 reviewed row",
      },
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p8-rough"),
      "timeline.summary": "FE-A11Y-P8 rough row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

    const summarySortHeader = page.getByTestId(
      gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"),
    );
    await expectVisibleFocus(summarySortHeader);
    await summarySortHeader.press("Enter");
    await expect(summarySortHeader).toContainText("Asc");

    await page
      .getByTestId(timelineRowMarkReviewedButtonTestId(reviewedRow.record_id))
      .click();
    await expect(
      page.getByTestId(
        rowCellTestId(reviewedRow.record_id, "timeline.capture_state"),
      ),
    ).toHaveText("reviewed");

    const filterField = page.getByTestId(
      gridFilterFieldTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(filterField);
    await filterField.selectOption("timeline.capture_state");
    const filterValue = page.getByTestId(
      gridFilterValueTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(filterValue);
    try {
      await filterValue.selectOption("reviewed");
    } catch {
      await filterValue.fill("reviewed");
    }
    const filterApply = page.getByTestId(
      gridFilterApplyTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(filterApply);
    await filterApply.press("Enter");
    await assertActiveFilterChipVisible(
      page,
      timelineViewSchemaId,
      "timeline.capture_state",
    );
    const filterChip = page.getByTestId(
      gridFilterChipTestId(timelineViewSchemaId, "timeline.capture_state"),
    );
    await expectVisibleFocus(filterChip);

    const groupingSelect = page.getByTestId(
      gridGroupingSelectTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(groupingSelect);
    await groupingSelect.selectOption("timeline.capture_state");
    const reviewedGroup = page.getByTestId(
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
    );
    await expect(reviewedGroup).toBeVisible();
    await expectVisibleFocus(reviewedGroup);
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");
    await reviewedGroup.press("Enter");
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "false");
    await reviewedGroup.press("Enter");
    await expect(reviewedGroup).toHaveAttribute("aria-expanded", "true");

    const savedViewSelector = page.getByTestId(
      savedViewSelectorTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(savedViewSelector);
    await expect(savedViewSelector).toHaveAttribute(
      "data-selected-sheet-ref-kind",
      "view_schema",
    );
    const savedViewNameInput = page.getByTestId(
      savedViewNameInputTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(savedViewNameInput);
    await savedViewNameInput.fill("FE-A11Y-P8 keyboard saved view");
    const createSavedViewButton = page.getByTestId(
      savedViewCreateButtonTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(createSavedViewButton);
    await createSavedViewButton.press("Enter");
    const savedViewStatus = page.getByTestId(
      savedViewStatusTestId(timelineViewSchemaId),
    );
    await expect(savedViewStatus).toHaveAttribute("aria-live", "polite");
    await expect(savedViewStatus).toHaveText("Saved view created.");
    await expect(savedViewSelector).toHaveAttribute(
      "data-selected-sheet-ref-kind",
      "saved_view",
    );

    const homeButton = page.getByTestId(
      savedViewSetHomeButtonTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(homeButton);
    await homeButton.press("Enter");
    await expect(savedViewStatus).toHaveText("Home view updated.");
    const defaultButton = page.getByTestId(
      savedViewSetDefaultButtonTestId(timelineViewSchemaId),
    );
    await expectVisibleFocus(defaultButton);
    await defaultButton.press("Enter");
    await expect(savedViewStatus).toHaveText("Default view updated.");

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      gridSortHeaderTestId(timelineViewSchemaId, "timeline.summary"),
      gridFilterFieldTestId(timelineViewSchemaId),
      gridFilterValueTestId(timelineViewSchemaId),
      gridFilterApplyTestId(timelineViewSchemaId),
      gridFilterChipTestId(timelineViewSchemaId, "timeline.capture_state"),
      gridGroupingSelectTestId(timelineViewSchemaId),
      gridGroupRowTestId(
        timelineViewSchemaId,
        "timeline.capture_state",
        "reviewed",
      ),
      savedViewSelectorTestId(timelineViewSchemaId),
      savedViewNameInputTestId(timelineViewSchemaId),
      savedViewCreateButtonTestId(timelineViewSchemaId),
      savedViewSetHomeButtonTestId(timelineViewSchemaId),
      savedViewSetDefaultButtonTestId(timelineViewSchemaId),
      savedViewStatusTestId(timelineViewSchemaId),
    ]);
  });
});

test.describe("FE-P9 accessibility readiness", () => {
  test(p9AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP9"),
      "FE-A11Y-P9 inspector actions",
    );
    const evidence = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p9-evidence"),
        "evidence.collector_party_text": "FE-A11Y-P9 collector",
        "evidence.title": "FE-A11Y-P9 evidence",
      },
    )) as ViewRow;
    const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
      [hostRefsFieldKey]: collectionActionsPayload(["FE-A11Y-P9 host"]),
      client_txn_id: uniqueTxn("fe-a11y-p9-row"),
      "timeline.details": "FE-A11Y-P9 inspector details",
      "timeline.summary": "FE-A11Y-P9 selected row",
    })) as ViewRow;
    const linkedRow = (await patchTimelineRecord(page, row.record_id, {
      base_row_version: row.row_version,
      changes: [
        {
          action_payload: phase9A11yAttachedEvidencePayload(evidence.record_id),
          field_key: "timeline.attached_evidence_ids",
        },
      ],
      client_txn_id: uniqueTxn("fe-a11y-p9-link"),
      view_schema_id: timelineViewSchemaId,
    })) as ViewRow;
    const hostItem = requireItemByRawText(
      collectionItems(linkedRow, hostRefsFieldKey),
      "FE-A11Y-P9 host",
    );
    const history = await fetchPhase9A11yRecordHistory(page, row.record_id);
    const rollbackItem = requirePhase9A11yHistoryEntryAction(history);
    const rollbackAnchor = phase9A11yRollbackAnchor(
      rollbackItem,
      "history_entry",
    );

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    const inspectButton = page.getByTestId(
      rowInspectButtonTestId(row.record_id),
    );
    await expectVisibleFocus(inspectButton);
    await inspectButton.press("Enter");

    for (const section of [
      "details",
      "relationships",
      "evidence",
      "history",
    ] as const) {
      await expect(
        page.getByTestId(timelineInspectorSectionTestId(section)),
      ).toBeVisible();
    }
    const detailsEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.details",
        recordId: row.record_id,
        surface: "inspector",
      }),
    );
    await expectVisibleFocus(detailsEditor);
    await expect(detailsEditor).toHaveValue("FE-A11Y-P9 inspector details");

    const relationshipChip = page
      .getByTestId(relationshipItemsTestId(row.record_id, hostRefsFieldKey))
      .getByTestId(relationshipChipTestId(String(hostItem.item_ref)));
    await expect(relationshipChip).toContainText("Unresolved");
    await expectVisibleFocus(relationshipChip);

    const openHistory = page.getByTestId(
      rowHistoryOpenButtonTestId(row.record_id),
    );
    await expectVisibleFocus(openHistory);
    await openHistory.press("Enter");
    await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
    const rollbackAction = page.getByTestId(
      phase9A11yHistoryActionTestId(rollbackItem, "history_entry"),
    );
    await expectVisibleFocus(rollbackAction);
    await rollbackAction.press("Enter");

    const rollbackPreview = page.getByTestId(
      rowHistoryRollbackPreviewTestId(rollbackAnchor),
    );
    await expect(rollbackPreview).toHaveAttribute("role", "dialog");
    await expect(rollbackPreview).toHaveAttribute("aria-modal", "true");
    await expect(rollbackPreview).toContainText(/rollback/i);
    const rollbackCancel = page.getByTestId(
      rowHistoryRollbackCancelButtonTestId(rollbackAnchor),
    );
    const rollbackConfirm = page.getByTestId(
      rowHistoryRollbackConfirmButtonTestId(rollbackAnchor),
    );
    await expectVisibleFocus(rollbackCancel);
    await expectVisibleFocus(rollbackConfirm);

    await page.route(`**/api/v1/records/${row.record_id}/rollback`, (route) =>
      route.fulfill({
        contentType: "application/json",
        status: 409,
        body: JSON.stringify({
          error: {
            code: "row_version_conflict",
            message: "Rollback target is stale for FE-A11Y-P9.",
            retryable: false,
          },
        }),
      }),
    );
    await rollbackConfirm.press("Enter");
    const historyMessage = page.getByTestId(rowHistoryMessageTestId());
    await expectAlertRole(historyMessage);
    await expect(historyMessage).toHaveAttribute("aria-live", "assertive");
    await expect(historyMessage).toContainText("row_version_conflict");
    await expectNoPrivateDiagnostics(historyMessage);

    const deleteButton = page.getByTestId(rowHistoryDeleteButtonTestId());
    await expectVisibleFocus(deleteButton);
    await deleteButton.press("Enter");
    const deletePanel = page.getByTestId(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
    );
    await expect(deletePanel).toHaveAttribute("role", "alertdialog");
    await expect(deletePanel).toHaveAttribute("aria-modal", "true");
    await expect(deletePanel).toContainText(row.record_id);
    const deleteConfirm = page.getByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    );
    const deleteCancel = page.getByTestId(
      rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
    );
    await expectVisibleFocus(deleteConfirm);
    await expectVisibleFocus(deleteCancel);
    await expectAndRecordContrast(page, [
      rowInspectButtonTestId(row.record_id),
      timelineScalarEditorTestId({
        fieldKey: "timeline.details",
        recordId: row.record_id,
        surface: "inspector",
      }),
      relationshipChipTestId(String(hostItem.item_ref)),
      rowHistoryOpenButtonTestId(row.record_id),
      phase9A11yHistoryActionTestId(rollbackItem, "history_entry"),
      rowHistoryRollbackConfirmButtonTestId(rollbackAnchor),
      rowHistoryRollbackCancelButtonTestId(rollbackAnchor),
      rowHistoryMessageTestId(),
      rowHistoryDeleteButtonTestId(),
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
      rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
    ]);
    await deleteCancel.press("Enter");
    await expect(deletePanel).toHaveCount(0);

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
  });
});

test.describe("FE-P10 accessibility readiness", () => {
  test(p10AccessibilityScenarioTitles[0], async ({ page }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YP10"),
      "FE-A11Y-P10 coordination accessibility",
    );
    const owner = await createIncidentMemberUser(page, incidentId, {
      display_name: "FE-P10 accessibility owner",
      email: uniqueEmail("fe-p10-a11y-owner"),
      initial_password: "Phase10A11y1!",
      role: "editor",
    });
    const party = (await createViewRow(page, incidentId, partiesViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p10-party"),
      "party.display_name": "FE-A11Y-P10 response party",
      "party.party_kind": "team",
    })) as ViewRow;
    const task = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-task"),
        "task.priority": "normal",
        "task.requester_party_id": party.record_id,
        "task.task_kind": "collection",
        "task.title": "FE-A11Y-P10 task alpha",
      },
    )) as ViewRow;
    const urgentTask = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-urgent-task"),
        "task.priority": "urgent",
        "task.task_kind": "follow_up",
        "task.title": "FE-A11Y-P10 task urgent",
      },
    )) as ViewRow;
    const clipboardRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-clipboard"),
        "timeline.occurred_at": "2026-06-12T10:00:00Z",
        "timeline.summary": "FE-A11Y-P10 clipboard row",
      },
    )) as ViewRow;
    const decision = (await createViewRow(
      page,
      incidentId,
      decisionsViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-decision"),
        "decision.decision_type": "containment",
        "decision.rationale": "FE-A11Y-P10 coordination rationale",
        "decision.summary": "FE-A11Y-P10 decision summary",
      },
    )) as ViewRow;
    const comm = (await createViewRow(page, incidentId, commLogViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p10-comm"),
      "comm_log.audience": "FE-A11Y-P10 responders",
      "comm_log.channel_or_meeting": "FE-A11Y-P10 bridge",
      "comm_log.comm_type": "briefing",
      "comm_log.decision_ids": {
        actions: [
          { linked_record_id: decision.record_id, op: "add_record_ref" },
        ],
        kind: "collection_actions_v1",
      },
      "comm_log.summary": "FE-A11Y-P10 communications log",
    })) as ViewRow;
    const handoff = (await createViewRow(
      page,
      incidentId,
      handoffViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-handoff"),
        "handoff.current_state_summary": "FE-A11Y-P10 handoff state",
        "handoff.incoming_owner_user_id": owner.user_id,
      },
    )) as ViewRow;
    const status = (await createViewRow(
      page,
      incidentId,
      statusReviewViewSchemaId,
      {
        client_txn_id: uniqueTxn("fe-a11y-p10-status"),
        "status_review.current_state_summary":
          "FE-A11Y-P10 status review state",
      },
    )) as ViewRow;
    const lesson = (await createViewRow(page, incidentId, lessonViewSchemaId, {
      client_txn_id: uniqueTxn("fe-a11y-p10-lesson"),
      "lesson.summary": "FE-A11Y-P10 lesson summary",
    })) as ViewRow;

    const surfaces = [
      {
        expected: "FE-A11Y-P10 task alpha",
        fieldKey: "task.title",
        groupToken: "coordination",
        label: "Task Requests",
        row: task,
        viewSchemaId: taskRequestsViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 decision summary",
        fieldKey: "decision.summary",
        groupToken: "coordination",
        label: "Decisions",
        row: decision,
        viewSchemaId: decisionsViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 response party",
        fieldKey: "party.display_name",
        groupToken: "scope-assessment",
        label: "Parties",
        row: party,
        viewSchemaId: partiesViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 communications log",
        fieldKey: "comm_log.summary",
        groupToken: "coordination",
        label: "Communications Log",
        row: comm,
        viewSchemaId: commLogViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 handoff state",
        fieldKey: "handoff.current_state_summary",
        groupToken: "coordination",
        label: "Handoff",
        row: handoff,
        viewSchemaId: handoffViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 status review state",
        fieldKey: "status_review.current_state_summary",
        groupToken: "review-learning",
        label: "Status Review",
        row: status,
        viewSchemaId: statusReviewViewSchemaId,
      },
      {
        expected: "FE-A11Y-P10 lesson summary",
        fieldKey: "lesson.summary",
        groupToken: "review-learning",
        label: "Lesson",
        row: lesson,
        viewSchemaId: lessonViewSchemaId,
      },
    ] as const;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        taskRequestsViewSchemaId,
      )}`,
    );
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.getByTestId(currentIncidentRoleTestId())).toContainText(
      "admin",
    );
    await expectStatusRole(page.getByTestId(saveStateTestId()));
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectTabTraversalAdvancesFrom(
      page,
      page.getByTestId(systemViewSwitcherTriggerTestId()),
    );

    for (const surface of surfaces) {
      await test.step(`${surface.label} accessibility surface`, async () => {
        await openA11ySystemSurface(page, surface);
        await expect(page).toHaveURL(
          new RegExp(
            `view_schema_id=${encodeURIComponent(surface.viewSchemaId)}`,
          ),
        );
        await expect(
          page.getByTestId(savedViewSelectorTestId(surface.viewSchemaId)),
        ).toHaveAttribute("data-selected-sheet-ref-kind", "view_schema");
        await expect(
          page.getByTestId(gridFilterFieldTestId(surface.viewSchemaId)),
        ).toBeVisible();
        await expect(
          page.getByTestId(gridFilterApplyTestId(surface.viewSchemaId)),
        ).toBeVisible();
        await expect(
          page.getByTestId(gridGroupingSelectTestId(surface.viewSchemaId)),
        ).toBeVisible();
        await expect(page.getByTestId("workbook-focus-anchor")).toContainText(
          surface.viewSchemaId,
        );

        const sortHeader = page.getByTestId(
          gridSortHeaderTestId(surface.viewSchemaId, surface.fieldKey),
        );
        await expectVisibleFocus(sortHeader);
        await expect(sortHeader).not.toHaveText("");

        const cell = page.getByTestId(
          rowCellTestId(surface.row.record_id, surface.fieldKey),
        );
        await expectCellTextOrValue(cell, surface.expected);
        await expectVisibleFocus(cell);
      });
    }

    await openA11ySystemSurface(page, {
      groupToken: "coordination",
      viewSchemaId: taskRequestsViewSchemaId,
    });

    const taskTitle = page.getByTestId(
      rowCellTestId(task.record_id, "task.title"),
    );
    await expectVisibleFocus(taskTitle);
    const copiedTaskTitle = await taskTitle.evaluate((element) => {
      const data = new DataTransfer();
      element.dispatchEvent(
        new ClipboardEvent("copy", {
          bubbles: true,
          cancelable: true,
          clipboardData: data,
        }),
      );
      return data.getData("text/plain");
    });
    expect(copiedTaskTitle).toBe("FE-A11Y-P10 task alpha");

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    const clipboardSummary = page.getByTestId(
      rowCellTestId(clipboardRow.record_id, "timeline.summary"),
    );
    await expectVisibleFocus(clipboardSummary);
    await pasteGridMatrix({
      fieldKey: "timeline.summary",
      matrix: [["FE-A11Y-P10 pasted timeline", "fe-a11y-p10-host"]],
      page,
      recordId: clipboardRow.record_id,
      surface: timelineViewSchemaId,
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(clipboardSummary).toHaveValue("FE-A11Y-P10 pasted timeline");

    await openA11ySystemSurface(page, {
      groupToken: "coordination",
      viewSchemaId: taskRequestsViewSchemaId,
    });

    await applyFilterChip(
      page,
      taskRequestsViewSchemaId,
      "task.priority",
      "urgent",
    );
    await assertActiveFilterChipVisible(
      page,
      taskRequestsViewSchemaId,
      "task.priority",
    );
    const priorityChip = page.getByTestId(
      gridFilterChipTestId(taskRequestsViewSchemaId, "task.priority"),
    );
    await expect(priorityChip).toContainText("urgent");
    await expectVisibleFocus(priorityChip);

    const urgentTitle = page.getByTestId(
      rowCellTestId(urgentTask.record_id, "task.title"),
    );
    await expectCellTextOrValue(urgentTitle, "FE-A11Y-P10 task urgent");

    const savedViewName = page.getByTestId(
      savedViewNameInputTestId(taskRequestsViewSchemaId),
    );
    await expectVisibleFocus(savedViewName);
    await savedViewName.fill("FE-A11Y-P10 keyboard saved view");
    const createSavedView = page.getByTestId(
      savedViewCreateButtonTestId(taskRequestsViewSchemaId),
    );
    await expectVisibleFocus(createSavedView);
    await createSavedView.press("Enter");
    const savedStatus = page.getByTestId(
      savedViewStatusTestId(taskRequestsViewSchemaId),
    );
    await expect(savedStatus).toHaveAttribute("aria-live", "polite");
    await expect(savedStatus).toHaveText("Saved view created.");

    await test.info().attach("fe-a11y-p10-01-readiness-matrix.json", {
      contentType: "application/json",
      body: Buffer.from(
        `${JSON.stringify(
          {
            checks: [
              "keyboard reachability",
              "visible focus",
              "accessible names",
              "system-view menu ARIA state",
              "status and saved-view live regions",
              "clipboard copy and paste",
              "non-color-only filter chip state",
            ],
            scenario_title: test.info().title,
            stable_identity_scope: "view_schema_id + record_id + field_key",
            surfaces: surfaces.map((surface) => ({
              field_key: surface.fieldKey,
              record_id: surface.row.record_id,
              surface: surface.label,
              view_schema_id: surface.viewSchemaId,
            })),
            viewport: "1440x900",
            zoom: "100%",
          },
          null,
          2,
        )}\n`,
      ),
    });

    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      systemViewSwitcherTriggerTestId(),
      savedViewSelectorTestId(taskRequestsViewSchemaId),
      savedViewNameInputTestId(taskRequestsViewSchemaId),
      savedViewCreateButtonTestId(taskRequestsViewSchemaId),
      savedViewStatusTestId(taskRequestsViewSchemaId),
      gridFilterFieldTestId(taskRequestsViewSchemaId),
      gridFilterApplyTestId(taskRequestsViewSchemaId),
      gridFilterChipTestId(taskRequestsViewSchemaId, "task.priority"),
      gridGroupingSelectTestId(taskRequestsViewSchemaId),
      gridSortHeaderTestId(taskRequestsViewSchemaId, "task.title"),
      rowCellTestId(urgentTask.record_id, "task.title"),
      saveStateTestId(),
    ]);
  });
});

test.describe("FE-P1 accessibility readiness", () => {
  test(p1AccessibilityScenarioTitles[0], async ({ page }) => {
    const phase1 = new Phase1Page(page);
    await clearBrowserSession(page);
    const heldSession = await holdSinglePublicAPIResponse(page, {
      method: "GET",
      path: "/api/v1/auth/session",
    });
    await phase1.goto();
    await heldSession.waitForHit;

    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "loading",
    );
    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "aria-busy",
      "true",
    );
    await expectStatusRole(page.getByTestId(phase1AuthTestId("status")));
    await expect(page.getByTestId(phase1AuthTestId("status"))).toContainText(
      "Checking current session",
    );
    await expectP1SurfaceA11y(page, {
      focusTestId: phase1AuthTestId("login-submit"),
      tabStops: [
        phase1AuthTestId("login-username"),
        phase1AuthTestId("login-password"),
        phase1AuthTestId("login-totp-code"),
        phase1AuthTestId("login-submit"),
      ],
    });

    try {
      heldSession.release();
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "anonymous",
      );
    } finally {
      await heldSession.dispose();
    }
  });

  test(
    p1AccessibilityScenarioTitles[1],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const phase1 = new Phase1Page(page);
      const email = uniqueEmail("a11y-p1-login");
      const password = "A11yP1LoginPass!";
      const user = await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Login",
        initial_password: password,
        mfa_required: false,
      });

      await clearBrowserSession(page);
      await phase1.goto();
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "anonymous",
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("login-username"),
        tabStops: [
          phase1AuthTestId("login-username"),
          phase1AuthTestId("login-password"),
          phase1AuthTestId("login-totp-code"),
          phase1AuthTestId("login-submit"),
        ],
      });

      await phase1.login(email, password);
      await expect(
        page.getByTestId(phase1LandingTestId("shell")),
      ).toBeVisible();
      await expectStatusRole(page.getByTestId(phase1LandingTestId("status")));
      await expectStatusRole(page.getByTestId(phase1AccountTestId("status")));
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1LandingTestId("refresh"),
        tabStops: [
          phase1LandingTestId("refresh"),
          phase1LandingTestId("incident-key"),
          phase1LandingTestId("incident-title"),
          phase1LandingTestId("create-button"),
          phase1AccountTestId("refresh-state"),
          phase1AccountTestId("logout"),
        ],
      });
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 anonymous login",
        userId: user.user_id,
      });
    },
  );

  test(
    p1AccessibilityScenarioTitles[2],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const phase1 = new Phase1Page(page);
      const email = uniqueEmail("a11y-p1-mfa");
      const password = "A11yP1MfaPass!";
      const user = await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 MFA",
        initial_password: password,
        mfa_required: true,
      });
      const secretBase32 = await enrollTotpViaBootstrap(email, password);

      await clearBrowserSession(page);
      await phase1.goto();
      await phase1.login(email, password);
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "mfa_required",
      );
      await expectAlertRole(page.getByTestId(phase1ErrorCodeTestId("auth")));
      await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
        "mfa_required",
      );
      await expectStatusRole(page.getByTestId(phase1AuthTestId("status")));
      await expect(page.getByTestId(phase1AuthTestId("status"))).toContainText(
        "TOTP code",
      );
      await expectNoPrivateDiagnostics(
        page.getByTestId(phase1ErrorSummaryTestIds("auth").container),
      );
      expect(await hasSessionCookie(page)).toBeFalsy();
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("login-totp-code"),
        tabStops: [
          phase1AuthTestId("login-username"),
          phase1AuthTestId("login-password"),
          phase1AuthTestId("login-totp-code"),
          phase1AuthTestId("login-submit"),
        ],
      });

      await phase1.login(email, password, generateTotpCode(secretBase32));
      await expect(
        page.getByTestId(phase1AccountTestId("session-user-id")),
      ).toHaveText(user.user_id);
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 mfa_required retry",
        userId: user.user_id,
      });
    },
  );

  test(
    p1AccessibilityScenarioTitles[3],
    async ({ page, workerAdminRequest }) => {
      const phase1 = new Phase1Page(page);
      const email = uniqueEmail("a11y-p1-mfa-setup");
      const password = "A11yP1SetupPass!";
      await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 MFA Setup",
        initial_password: password,
        mfa_required: true,
      });

      await clearBrowserSession(page);
      await phase1.goto();
      await phase1.login(email, password);
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "mfa_setup_required",
      );
      await expectAlertRole(page.getByTestId(phase1ErrorCodeTestId("auth")));
      await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
        "mfa_setup_required",
      );
      await expect(
        page.getByTestId(phase1AuthTestId("bootstrap-token")),
      ).toHaveText("Stored for TOTP setup requests.");
      await expectNoPrivateDiagnostics(
        page.getByTestId(phase1ErrorSummaryTestIds("auth").container),
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("bootstrap-begin"),
        tabStops: [
          phase1AuthTestId("login-username"),
          phase1AuthTestId("login-password"),
          phase1AuthTestId("login-totp-code"),
          phase1AuthTestId("login-submit"),
          phase1AuthTestId("bootstrap-begin"),
          phase1AuthTestId("bootstrap-complete-code"),
          phase1AuthTestId("bootstrap-complete"),
        ],
      });

      await phase1.beginBootstrapEnrollment();
      await expectStatusRole(page.getByTestId(phase1AuthTestId("status")));
      const secretBase32 = await phase1.requireText(
        phase1AuthTestId("bootstrap-secret-base32"),
      );
      await phase1.completeBootstrapEnrollment(generateTotpCode(secretBase32));
      await expect(page.getByTestId(phase1AuthTestId("status"))).toHaveText(
        "TOTP enrollment completed. Sign in with your TOTP code.",
      );
      expect(await hasSessionCookie(page)).toBeFalsy();
    },
  );

  test(p1AccessibilityScenarioTitles[4], async ({ page }) => {
    const phase1 = new Phase1Page(page);
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("A11YLAND"),
      "A11Y authenticated landing",
    );

    await phase1.goto();
    await expect(page.getByTestId(phase1LandingTestId("shell"))).toBeVisible();
    await expect(
      page.getByTestId(landingIncidentCardTestId(incidentId)),
    ).toBeVisible();
    await expectStatusRole(page.getByTestId(phase1LandingTestId("status")));
    await expectStatusRole(page.getByTestId(phase1AccountTestId("status")));
    await expectStatusRole(page.getByTestId(phase1AdminTestId("status")));
    await expectP1SurfaceA11y(page, {
      focusTestId: phase1LandingTestId("create-button"),
      tabStops: [
        phase1LandingTestId("refresh"),
        phase1LandingTestId("incident-key"),
        phase1LandingTestId("incident-title"),
        phase1LandingTestId("create-button"),
        phase1AccountTestId("refresh-state"),
        phase1AccountTestId("logout"),
        phase1AdminTestId("create-email"),
        phase1AdminTestId("create-user"),
      ],
    });
  });

  test(
    p1AccessibilityScenarioTitles[5],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const phase1 = new Phase1Page(page);
      const email = uniqueEmail("a11y-p1-incident");
      const password = "A11yP1IncidentPass!";
      const user = await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Incident",
        initial_password: password,
        mfa_required: false,
      });

      await clearBrowserSession(page);
      await phase1.goto();
      await phase1.login(email, password);
      await expect(
        page.getByTestId(phase1LandingTestId("empty-state")),
      ).toContainText("No incidents are visible");
      await expectStatusRole(page.getByTestId(phase1LandingTestId("status")));

      const selectedIncidentId = await createIncidentWithRequest(
        workerAdminRequest,
        uniqueIncidentKey("A11YEMPTYA"),
        "A11Y selected incident",
      );
      await createIncidentMembershipWithRequest(
        workerAdminRequest,
        selectedIncidentId,
        email,
        "admin",
      );
      const alternateIncidentId = await createIncidentWithRequest(
        workerAdminRequest,
        uniqueIncidentKey("A11YEMPTYB"),
        "A11Y alternate incident",
      );
      await createIncidentMembershipWithRequest(
        workerAdminRequest,
        alternateIncidentId,
        email,
        "admin",
      );

      await phase1.refreshLanding();
      await expect(
        page.getByTestId(landingIncidentCardTestId(selectedIncidentId)),
      ).toBeVisible();
      await expect(
        page.getByTestId(landingIncidentCardTestId(alternateIncidentId)),
      ).toBeVisible();

      await phase1.openIncident(selectedIncidentId);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expect(page.getByTestId(currentIncidentRoleTestId())).toContainText(
        "admin",
      );
      await page.getByTestId(incidentControlsTriggerTestId()).click();
      await expect(
        page.getByTestId(incidentControlsPanelTestId()),
      ).toBeVisible();
      await expectStatusRole(page.getByTestId("incident-admin-status"));
      await expectVisibleFocus(
        page.getByTestId(phase1A11yAppLocalTestId("incidentPatchButton")),
      );
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1LandingTestId("return"),
        tabStops: [phase1LandingTestId("return")],
      });

      await phase1.returnToLanding();
      await phase1.openIncident(selectedIncidentId);

      const selectedMembership = await loadIncidentMembership(
        workerAdminRequest,
        selectedIncidentId,
        user.user_id,
      );
      await deleteIncidentMembership(
        workerAdminRequest,
        selectedIncidentId,
        user.user_id,
        selectedMembership.membership_version,
      );
      await page.reload();
      await expect(
        page.getByTestId(phase1LandingTestId("shell")),
      ).toBeVisible();
      await expect(
        page.getByTestId(phase1LandingTestId("status")),
      ).toContainText("no longer visible");
      await expect(
        page.getByTestId(landingIncidentCardTestId(selectedIncidentId)),
      ).toHaveCount(0);
      await expect(
        page.getByTestId(landingIncidentCardTestId(alternateIncidentId)),
      ).toBeVisible();

      await phase1.openIncident(alternateIncidentId);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      await expect(page.getByTestId(currentIncidentRoleTestId())).toContainText(
        "admin",
      );
      await page.getByTestId(incidentControlsTriggerTestId()).click();
      await expect(
        page.getByTestId(incidentControlsPanelTestId()),
      ).toBeVisible();
      await expect(page.getByTestId("incident-patch-tlp")).toBeVisible();
      const alternateMembership = await loadIncidentMembership(
        workerAdminRequest,
        alternateIncidentId,
        user.user_id,
      );
      await patchIncidentMembership(workerAdminRequest, alternateIncidentId, {
        baseMembershipVersion: alternateMembership.membership_version,
        role: "editor",
        userId: user.user_id,
      });
      await phase1.patchIncidentFields({
        currentPhase: "containment",
        externalCase: "CASE-A11Y",
        tlp: "amber",
      });
      await expectAlertRole(page.getByTestId("incident-admin-error-code"));
      await expect(page.getByTestId("incident-admin-error-code")).toHaveText(
        "authorization_denied",
      );
      await expectNoPrivateDiagnostics(
        page.getByTestId("incident-admin-error-code"),
      );
      await expectVisibleFocus(page.getByTestId(phase1LandingTestId("return")));
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 incident states",
        userId: user.user_id,
      });
    },
  );

  test(p1AccessibilityScenarioTitles[6], async ({ page }) => {
    const routePattern = "**/api/v1/incidents";
    const routeHandler = async (route: Route) => {
      if (route.request().method().toUpperCase() !== "GET") {
        await route.fallback();
        return;
      }
      await fulfillPublicError(route, {
        code: "authorization_denied",
        details: {
          required_role: "incident_admin",
        },
        message: "Access denied.",
        status: 403,
      });
    };

    await page.route(routePattern, routeHandler);
    await page.goto("/");
    await expect(page.getByTestId(phase1LandingTestId("shell"))).toBeVisible();
    await expect(
      page.getByTestId(phase1LandingTestId("shell")),
    ).toHaveAttribute("data-bootstrap-state", "forbidden");
    await expectStatusRole(page.getByTestId(phase1LandingTestId("status")));
    await expectAlertRole(page.getByTestId(phase1ErrorCodeTestId("landing")));
    await expect(page.getByTestId(phase1ErrorCodeTestId("landing"))).toHaveText(
      "authorization_denied",
    );
    await expectNoPrivateDiagnostics(
      page.getByTestId(phase1ErrorSummaryTestIds("landing").container),
    );
    await expectP1SurfaceA11y(page, {
      focusTestId: phase1LandingTestId("refresh"),
      tabStops: [
        phase1LandingTestId("refresh"),
        phase1AccountTestId("refresh-state"),
        phase1AccountTestId("logout"),
        phase1AdminTestId("create-email"),
      ],
    });

    await safeUnroute(page, routePattern, routeHandler);
    await page.getByTestId(phase1LandingTestId("refresh")).click();
    await expect(page.getByTestId(phase1ErrorCodeTestId("landing"))).toHaveText(
      "",
    );
  });

  test(
    p1AccessibilityScenarioTitles[7],
    async ({ page, sessionTracker, workerAdminRequest }) => {
      const phase1 = new Phase1Page(page);
      const incidentId = await createIncident(
        page,
        uniqueIncidentKey("A11YREVOKE"),
        "A11Y revoked incident",
      );
      const email = uniqueEmail("a11y-p1-revoked");
      const password = "A11yP1RevokedPass!";
      const user = await createAuthLocalUser(workerAdminRequest, {
        email,
        display_name: "A11Y P1 Revoked",
        initial_password: password,
        mfa_required: false,
      });
      await createIncidentMembership(page, incidentId, email, "viewer");

      await clearBrowserSession(page);
      await phase1.goto();
      await phase1.login(email, password);
      await expect(
        page.getByTestId(phase1AccountTestId("session-user-id")),
      ).toHaveText(user.user_id);
      await expect(
        page.getByTestId(landingIncidentCardTestId(incidentId)),
      ).toBeVisible();
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 revoked session before revoke-all",
        userId: user.user_id,
      });

      await revokeAllSessions(
        workerAdminRequest,
        user.user_id,
        "FE-A11Y-P1-01 revoked-session",
      );
      await phase1.refreshLanding();
      await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
        "data-bootstrap-state",
        "revoked",
      );
      await expect(
        page.getByTestId(phase1AuthTestId("shell-message")),
      ).toContainText("Sign in again");
      await expectStatusRole(page.getByTestId(phase1AuthTestId("status")));
      await expectP1SurfaceA11y(page, {
        focusTestId: phase1AuthTestId("login-submit"),
        tabStops: [
          phase1AuthTestId("login-username"),
          phase1AuthTestId("login-password"),
          phase1AuthTestId("login-totp-code"),
          phase1AuthTestId("login-submit"),
        ],
      });

      await phase1.login(email, password);
      await expect(
        page.getByTestId(phase1AccountTestId("session-user-id")),
      ).toHaveText(user.user_id);
      await sessionTracker.captureCurrentSession(page, {
        createdBy: "phase1 accessibility",
        email,
        purpose: "FE-A11Y-P1-01 revoked session re-auth",
        userId: user.user_id,
      });
    },
  );

  test(p1AccessibilityScenarioTitles[8], async ({ page }) => {
    const routePattern = "**/api/v1/auth/credential-state";
    const routeHandler = async (route: Route) => {
      await fulfillPublicError(route, {
        code: "credential_state_unavailable",
        details: {
          field: "credential_state",
          reason_code: "temporary_failure",
        },
        message: "select private_column from local_credentials",
        status: 500,
      });
    };

    await page.route(routePattern, routeHandler);
    await page.goto("/");
    await expect(page.getByTestId(phase1ErrorCodeTestId("account"))).toHaveText(
      "credential_state_unavailable",
    );
    await expectAlertRole(page.getByTestId(phase1ErrorCodeTestId("account")));
    await expectAlertRole(
      page.getByTestId(phase1ErrorSummaryTestIds("account").container),
    );
    await expect(
      page.getByTestId(phase1ErrorSummaryTestIds("account").message),
    ).toHaveText("Request failed.");
    await expect(
      page.getByTestId(phase1ErrorSummaryTestIds("account").details),
    ).toContainText("Reason: temporary_failure");
    await expect(
      page.getByTestId(phase1ErrorSummaryTestIds("account").details),
    ).toContainText("Field: credential_state");
    await expectNoPrivateDiagnostics(
      page.getByTestId(phase1ErrorSummaryTestIds("account").container),
    );
    await expectP1SurfaceA11y(page, {
      focusTestId: phase1AccountTestId("refresh-state"),
      tabStops: [
        phase1LandingTestId("refresh"),
        phase1AccountTestId("refresh-state"),
        phase1AccountTestId("logout"),
      ],
    });

    await safeUnroute(page, routePattern, routeHandler);
    await page.getByTestId(phase1AccountTestId("refresh-state")).click();
    await expect(page.getByTestId(phase1AccountTestId("status"))).toHaveText(
      "Refreshed account security.",
    );
  });
});
