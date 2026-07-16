import { Buffer } from "node:buffer";
import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  assertMarkerAnchoredToGridTarget,
  changeGrouping,
  createSavedViewFromCurrentSurface,
  scrollGridTargetIntoView,
  setCurrentSavedViewAsDefault,
  setCurrentSavedViewAsHome,
  setSavedViewDraftName,
} from "@cartulary/test-utils";
import {
  cartularyDefaultThemeId,
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  dataTestIdPrefixSelector,
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  gridGroupingSelectTestId,
  gridGroupRowTestId,
  gridRowGutterTestId,
  gridRowTestId,
  gridSavedRowsSelector,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentControlsPanelTestId,
  incidentMembershipListTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveCancelButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  rowPresenceMarkerTestId,
  savedViewSelectorTestId,
  savedViewStatusTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherTriggerTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorMessageTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
  type WorkbookSurface,
  workbookFilterPopoverTriggerTestId,
  workbookInlineDraftRowTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookResponsiveBandTestId,
  workbookShellReadyTestId,
  workbookShellSlots,
  workbookShellSlotTestId,
  workbookSortMenuTriggerTestId,
  workbookTopBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page, Route, TestInfo } from "@playwright/test";
import {
  createEvidenceFixtureRow,
  createUploadedEvidenceFixture,
  type EvidenceUploadOptions,
} from "./evidenceFixtureHelpers";
import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createIncidentMemberUser,
  createViewRow,
  csrfHeaders,
  gridSavedRows,
  holdBrowserApiRequest,
  patchRecord,
  queryViewRows,
  testRouteHeaders,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  collectionItems,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  hostRefsFieldKey,
  hostsViewSchemaId,
  lessonViewSchemaId,
  openTimelineInspector,
  partiesViewSchemaId,
  requireItemByRawText,
  resolvedRefPayload,
  seedHostMentionStateFixture,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./phase4Helpers";
import {
  driveRealTimelineSummaryConflict,
  editTimelineSummary,
  focusRemoteTimelineCellAndWaitForPresence,
  installIncidentSocketMonitor,
  installPatchController,
  installPatchTransportFailureController,
  openIncidentAsTrackedUserReady,
  successfulPatchCalls,
} from "./phase6Harness";
import { injectDesignFixture } from "./visualFixtureHelpers";

type ViewRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, unknown>;
};

type FrontendVisualFixture = {
  blocked_reason: string;
  browser_zoom_percent: number;
  capture_scope: { kind: string; selector?: string };
  density_id: string;
  device_scale_factor: number;
  dynamic_masks: string[];
  fixture_id: string;
  fixture_title: string;
  focus_state: Record<string, unknown>;
  editor_state: Record<string, unknown>;
  golden_artifacts: string[];
  golden_filename: string;
  inspector_state: Record<string, unknown>;
  no_dynamic_regions: boolean;
  owner_phase_ids: string[];
  owner_row_ids: string[];
  playwright_scenario_title: string;
  replacement_fixture_id: string;
  scroll_normalization: { kind: string; anchor?: string; reason?: string };
  seed_id: string;
  status: string;
  theme_id: string;
  viewport_css_px: string;
};

type FrontendVisualFixtureRegistry = {
  fixtures: FrontendVisualFixture[];
  guide_path: string;
  schema_id: string;
};

const expectedFeP11VisualFixtureIds = Array.from(
  { length: 21 },
  (_, index) => `FE-VFIX-${String(index + 1).padStart(2, "0")}`,
);

function findRepoRoot(): string {
  let candidate = process.cwd();
  while (true) {
    if (
      existsSync(
        path.join(candidate, "tools", "frontend_visual_fixture_registry.json"),
      )
    ) {
      return candidate;
    }
    const parent = path.dirname(candidate);
    if (parent === candidate) {
      throw new Error(
        "could not find tools/frontend_visual_fixture_registry.json",
      );
    }
    candidate = parent;
  }
}

function repoPath(relativePath: string): string {
  return path.join(findRepoRoot(), relativePath);
}

function loadFrontendVisualFixtureRegistry(): FrontendVisualFixtureRegistry {
  return JSON.parse(
    readFileSync(
      repoPath("tools/frontend_visual_fixture_registry.json"),
      "utf8",
    ),
  ) as FrontendVisualFixtureRegistry;
}

function frontendVisualFixtureRegistryDigest(): string {
  return createHash("sha256")
    .update(
      readFileSync(repoPath("tools/frontend_visual_fixture_registry.json")),
    )
    .digest("hex");
}

function expectCurrentFrontendVisualFixtureMetadata(
  fixture: FrontendVisualFixture,
) {
  expect(fixture.status).toBe("current");
  expect(fixture.fixture_title.length).toBeGreaterThan(0);
  expect(fixture.owner_phase_ids.length).toBeGreaterThan(0);
  expect(fixture.owner_row_ids.length).toBeGreaterThan(0);
  expect(fixture.playwright_scenario_title.length).toBeGreaterThan(0);
  expect(fixture.seed_id.length).toBeGreaterThan(0);
  expect(fixture.viewport_css_px).toMatch(/^[0-9]+x[0-9]+$/);
  expect(fixture.device_scale_factor).toBeGreaterThanOrEqual(1);
  expect(fixture.browser_zoom_percent).toBe(100);
  expect(fixture.theme_id).toBe("dark_graphite");
  expect(fixture.density_id.length).toBeGreaterThan(0);
  expect(fixture.capture_scope.kind).toMatch(
    /^(full_viewport|selector|region)$/,
  );
  expect(fixture.scroll_normalization.kind.length).toBeGreaterThan(0);
  expect(Array.isArray(fixture.dynamic_masks)).toBeTruthy();
  if (fixture.no_dynamic_regions) {
    expect(fixture.dynamic_masks).toEqual([]);
  }
  expect(fixture.blocked_reason).toBe("");
  expect(fixture.replacement_fixture_id).toBe("");
  for (const artifact of fixture.golden_artifacts) {
    expect(
      existsSync(repoPath(artifact)),
      `${artifact} should exist`,
    ).toBeTruthy();
  }
  expect(fixture.golden_artifacts).toContain(fixture.golden_filename);
}

type FeP9VisualHistoryItem = {
  available_rollback_actions: Array<
    "history_entry" | "change_set" | "row_restore"
  >;
  history_entry_ref?: string;
  history_item_ref: string;
};

type FeP9VisualHistoryData = {
  items: FeP9VisualHistoryItem[];
  row_version: number;
};

type GridVisualScrollLeft = "left" | "right" | number;

type GridVisualScrollState = {
  top: number;
  left: GridVisualScrollLeft;
};

type WorkbookGridVisualScrollSnapshot = {
  shellTop: number;
  shellLeft: number;
  scrollportTop: number;
  scrollportLeft: number;
};

type GridVisualAnchor = {
  kind: "timelineEvidenceActions";
  rowId: string;
  top: number;
};

function gridShellSelector(surface: string): string {
  return dataTestIdSelector(gridShellTestId(surface));
}

async function readTimelineGridFirstLayout(page: Page) {
  return page.evaluate(
    ({
      gridSelector,
      inspectorSelector,
      scrollportSelector,
      topBarSelector,
      viewBarSelector,
    }) => {
      const rectFor = (element: HTMLElement) => {
        const rect = element.getBoundingClientRect();
        return {
          bottom: Math.round(rect.bottom),
          height: Math.round(rect.height),
          left: Math.round(rect.left),
          right: Math.round(rect.right),
          top: Math.round(rect.top),
          width: Math.round(rect.width),
        };
      };
      const roundedRect = (selector: string) => {
        const element = document.querySelector<HTMLElement>(selector);
        if (element === null) {
          throw new Error(`Expected ${selector} to exist`);
        }
        return rectFor(element);
      };
      const optionalRoundedRect = (selector: string) => {
        const element = document.querySelector<HTMLElement>(selector);
        if (element === null) {
          return null;
        }
        return rectFor(element);
      };
      const scrollport =
        document.querySelector<HTMLElement>(scrollportSelector);
      if (scrollport === null) {
        throw new Error(`Expected ${scrollportSelector} to exist`);
      }
      const columnHeaders = Array.from(
        scrollport.querySelectorAll<HTMLElement>('[role="columnheader"]'),
      );
      const lastColumnElement =
        columnHeaders.length === 0
          ? null
          : columnHeaders[columnHeaders.length - 1];
      const lastColumn =
        lastColumnElement === null || lastColumnElement === undefined
          ? null
          : rectFor(lastColumnElement);
      return {
        grid: roundedRect(gridSelector),
        innerGrid: roundedRect(scrollportSelector),
        inspector: optionalRoundedRect(inspectorSelector),
        lastColumn,
        scrollport: {
          clientHeight: scrollport.clientHeight,
          clientWidth: scrollport.clientWidth,
          scrollTop: Math.round(scrollport.scrollTop),
        },
        topBar: roundedRect(topBarSelector),
        viewBar: roundedRect(viewBarSelector),
        windowY: Math.round(window.scrollY),
      };
    },
    {
      gridSelector: gridShellSelector(timelineViewSchemaId),
      inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
      scrollportSelector: `${gridShellSelector(
        timelineViewSchemaId,
      )} ${gridScrollportSelector()}`,
      topBarSelector: dataTestIdSelector(workbookShellSlotTestId("top-bar")),
      viewBarSelector: dataTestIdSelector(workbookShellSlotTestId("view-bar")),
    },
  );
}

async function expectWideWorkbookTopBarChrome(page: Page) {
  await expect(
    page.getByTestId(workbookResponsiveBandTestId()),
  ).toHaveAttribute("data-workbook-responsive-band", "base");
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(systemViewSwitcherTriggerTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookTopBarQueryControlsTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookSortMenuTriggerTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(workbookFilterPopoverTriggerTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(
    page.getByLabel("Account and application navigation"),
  ).toBeVisible();
}

async function openTimelineRowActions(page: Page, recordId: string) {
  const rowTestId = gridRowTestId(timelineViewSchemaId, recordId);
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: rowTestId,
  });
  await page.getByTestId(rowTestId).click({ button: "right" });
}

async function mountedGridTarget(
  page: Page,
  surface: WorkbookSurface,
  targetTestId: string,
): Promise<Locator> {
  await scrollGridTargetIntoView({ page, surface, targetTestId });
  return page.getByTestId(targetTestId);
}

async function mountedGridCell(
  page: Page,
  surface: WorkbookSurface,
  recordId: string,
  fieldKey: string,
): Promise<Locator> {
  return mountedGridTarget(page, surface, rowCellTestId(recordId, fieldKey));
}

async function clickTimelineRowAction(
  page: Page,
  recordId: string,
  actionTestId: string,
) {
  await openTimelineRowActions(page, recordId);
  await page.getByTestId(actionTestId).click();
}

async function blurActiveElement(page: Page) {
  await page.evaluate(() => {
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
  });
}

function tagActionsPayload(tagNames: string[]) {
  return {
    kind: "collection_actions_v1",
    actions: tagNames.map((tagName) => ({
      op: "add_tag",
      tag_name: tagName,
    })),
  };
}

type GridVisualRegressionOptions = {
  maxDiffPixels?: number;
  testInfo?: TestInfo;
} & (
  | { scroll: GridVisualScrollState; anchor?: never }
  | { anchor: GridVisualAnchor; scroll?: never }
);

type AuthVisualLoginMode =
  | "invalid_credentials"
  | "invalid_mfa"
  | "mfa_required"
  | "mfa_setup_required"
  | "pending"
  | "service_unavailable";

function releaseAuthVisualStep(release: (() => void) | null) {
  release?.();
}

test.describe("FE-P1 auth gateway visual readiness", () => {
  test("FE-V-P1-01 Capture auth gateway initial, focused, loading, invalid credentials, MFA required, invalid MFA, MFA setup required, service unavailable, mobile, reduced-motion, and 200%-zoom states.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await attachFontManifestDigest();

    let sessionPending = true;
    let releaseSession: (() => void) | null = null;
    let loginMode: AuthVisualLoginMode = "invalid_credentials";
    let pendingLoginResult: AuthVisualLoginMode = "invalid_credentials";
    let releaseLogin: (() => void) | null = null;

    await page.route("**/api/v1/auth/session", async (route) => {
      if (sessionPending) {
        await new Promise<void>((resolve) => {
          releaseSession = resolve;
        });
      }
      await fulfillAuthVisualError(route, {
        code: "session_required",
        message: "Session required.",
        status: 401,
      });
    });
    await page.route("**/api/v1/auth/providers", async (route) => {
      await fulfillAuthVisualJSON(route, { data: { providers: [] } });
    });
    await page.route("**/api/v1/auth/login", async (route) => {
      const mode = loginMode;
      if (mode === "pending") {
        await new Promise<void>((resolve) => {
          releaseLogin = resolve;
        });
        await fulfillAuthVisualLogin(route, pendingLoginResult);
        return;
      }
      await fulfillAuthVisualLogin(route, mode);
    });

    await page.goto("/");
    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "loading",
    );
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-loading");
    sessionPending = false;
    releaseAuthVisualStep(releaseSession);

    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "anonymous",
    );
    await expect(
      page.getByTestId(phase1AuthTestId("login-totp-code")),
    ).toHaveCount(0);
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-initial");

    await page.getByTestId(phase1AuthTestId("login-username")).focus();
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-focused");

    await fillAuthVisualCredentials(page);
    loginMode = "pending";
    pendingLoginResult = "invalid_credentials";
    await page.getByTestId(phase1AuthTestId("login-submit")).click();
    await expect(page.getByTestId(phase1AuthTestId("login-submit"))).toHaveText(
      "Signing in...",
    );
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-submitting");
    releaseAuthVisualStep(releaseLogin);
    await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
      "Email or password is incorrect.",
    );
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-invalid-credentials");

    loginMode = "mfa_required";
    await page.getByTestId(phase1AuthTestId("login-submit")).click();
    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "mfa_required",
    );
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-mfa-required");

    loginMode = "invalid_mfa";
    await page.getByTestId(phase1AuthTestId("login-totp-code")).fill("000000");
    await page.getByTestId(phase1AuthTestId("login-submit")).click();
    await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
      "The verification code is incorrect or expired.",
    );
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-invalid-mfa");

    loginMode = "mfa_setup_required";
    await page.reload();
    await fillAuthVisualCredentials(page);
    await page.getByTestId(phase1AuthTestId("login-submit")).click();
    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "mfa_setup_required",
    );
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-mfa-setup-required");

    loginMode = "service_unavailable";
    await page.reload();
    await fillAuthVisualCredentials(page);
    await page.getByTestId(phase1AuthTestId("login-submit")).click();
    await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toHaveText(
      "Authentication is temporarily unavailable. Try again.",
    );
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-service-unavailable");

    await page.setViewportSize({ width: 390, height: 844 });
    await page.reload();
    await expect(page.getByTestId(phase1AuthTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "anonymous",
    );
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-mobile");

    await page.setViewportSize({ width: 1440, height: 900 });
    await page.emulateMedia({ reducedMotion: "reduce" });
    await page.reload();
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-reduced-motion");
    await page.emulateMedia({ reducedMotion: "no-preference" });

    await page.evaluate(() => {
      document.documentElement.style.zoom = "200%";
    });
    await assertAuthGatewayVisual(page, "fe-v-p1-01-auth-200-zoom");
    await page.evaluate(() => {
      document.documentElement.style.zoom = "100%";
    });
  });
});

test.describe("FE-P2 workbook visual readiness", () => {
  test("FE-V-P2-01 Capture Default Timeline workbook shell with view-bar query controls, compact sheet toolbar, dense Timeline grid, collapsed inspector default, explicit inspector opener, bottom draft row, and status strip.", async ({
    page,
  }) => {
    test.setTimeout(120_000);
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV2SHELL"),
      "FE-P2 visual default shell",
    );

    const rows: ViewRow[] = [];
    const fixtureRows = [
      "Login attempt with valid user",
      "Password spray from single source",
      "Failed MFA challenge",
      "Suspicious automation execution",
      "Outbound connection to uncommon provider",
      "New service installed",
      "Potential credential access",
      "User accessed sensitive share",
      "Data archived to temporary directory",
      "Archive staged for exfiltration",
      "Alert from endpoint rule triggered",
      "Host isolated by containment playbook",
      "Scheduled task removed",
      "Credential reset completed",
      "Investigation opened",
      "Containment review assigned",
      "Remote shell attempt blocked",
      "Cloud sign-in risk elevated",
      "Analyst comment added",
      "Final verification queued",
      ...Array.from(
        { length: 28 },
        (_, index) => `Follow-up chronology detail ${index + 1}`,
      ),
    ];
    for (const [index, summary] of fixtureRows.entries()) {
      rows.push(
        (await createViewRow(page, incidentId, timelineViewSchemaId, {
          client_txn_id: uniqueTxn(`FEV2SHELL-ROW-${index + 1}`),
          "timeline.activity_utc_text": new Date(
            Date.UTC(2026, 3, 18, 14, 12 + index * 2, 34),
          ).toISOString(),
          "timeline.activity_synopsis_text": summary,
          "timeline.raw_activity_text": `Default Timeline workbook shell fixture row ${
            index + 1
          }`,
          "timeline.host_refs": collectionActionsPayload([
            index % 3 === 0 ? "host-gamma" : "host-alpha",
          ]),
          "timeline.identity_refs": collectionActionsPayload([
            index % 2 === 0
              ? "identity-alpha@example.test"
              : "identity-beta@example.test",
          ]),
          "timeline.tags": tagActionsPayload([
            index % 4 === 0 ? "review" : "triage",
            index % 5 === 0 ? "evidence" : "timeline",
          ]),
        })) as ViewRow,
      );
    }
    const rowSummariesById = new Map(
      rows.map((row, index) => [row.record_id, fixtureRows[index] ?? ""]),
    );

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    const shell = page.getByTestId(workbookShellReadyTestId());
    await expect(shell).toBeVisible();
    await expect(shell).toHaveAttribute(
      "data-active-view-schema-id",
      timelineViewSchemaId,
    );
    for (const slot of workbookShellSlots) {
      if (slot === "inspector") {
        await expect(
          shell.locator(dataTestIdSelector(workbookShellSlotTestId(slot))),
        ).toHaveCount(0);
        continue;
      }
      await expect(
        shell.locator(dataTestIdSelector(workbookShellSlotTestId(slot))),
      ).toBeVisible();
    }
    await expect(page.getByTestId(incidentControlsPanelTestId())).toHaveCount(
      0,
    );
    await expect(page.getByTestId("incident-summary-key")).toHaveCount(0);
    await expect(page.getByTestId("incident-patch-button")).toHaveCount(0);
    await expect(page.getByTestId(incidentMembershipListTestId())).toHaveCount(
      0,
    );
    await expect(page.getByText("Phase 3 workbook")).toHaveCount(0);
    await expect(page.getByText(/Timeline mutation substrate/u)).toHaveCount(0);
    await expect(
      page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
    ).toHaveAttribute("aria-current", "page");
    await expect(
      page.getByTestId(systemViewSwitcherTriggerTestId()),
    ).toBeVisible();
    await expect(
      page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

    const grid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
    await expect(grid).toBeVisible();
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, incidentId, timelineViewSchemaId)).length,
      )
      .toBe(rows.length);
    await expect
      .poll(async () => gridSavedRows(page, timelineViewSchemaId).count())
      .toBeGreaterThanOrEqual(12);
    const renderedRecordIds = await grid
      .locator(gridSavedRowsSelector())
      .evaluateAll((rowElements) =>
        rowElements.map(
          (rowElement) => rowElement.getAttribute("data-grid-record-id") ?? "",
        ),
      );
    expect(rows.map((row) => row.record_id)).toEqual(
      expect.arrayContaining(renderedRecordIds),
    );
    const selectedRowId = renderedRecordIds[0];
    const selectedRow = rows.find((row) => row.record_id === selectedRowId);
    if (selectedRow === undefined) {
      throw new Error(`FE-V-P2 fixture selected unknown row ${selectedRowId}`);
    }
    const selectedGridRow = grid.locator(
      `[data-grid-record-id="${selectedRow.record_id}"]`,
    );
    await (
      await mountedGridCell(
        page,
        timelineViewSchemaId,
        selectedRow.record_id,
        "timeline.date_entered_text",
      )
    ).click();
    await expect(selectedGridRow).toHaveAttribute(
      "data-inspector-active",
      "true",
    );
    const selectedGridRowGutter = selectedGridRow.getByTestId(
      gridRowGutterTestId(timelineViewSchemaId, selectedRow.record_id),
    );
    await expect(selectedGridRowGutter).toBeVisible();
    await expect(selectedGridRowGutter).toHaveAttribute(
      "data-grid-field-key",
      "__cartulary_row_gutter__",
    );

    const defaultTimelineFields = [
      "timeline.date_entered_text",
      "timeline.analyst_text",
      "timeline.mitre_stage_text",
      "timeline.device_object_text",
      "timeline.ip_address_text",
      "timeline.activity_utc_text",
      "timeline.activity_local_text",
      "timeline.raw_activity_text",
      "timeline.activity_synopsis_text",
      "timeline.data_source_text",
    ];
    for (const fieldKey of defaultTimelineFields) {
      const header = await mountedGridTarget(
        page,
        timelineViewSchemaId,
        gridSortHeaderTestId(timelineViewSchemaId, fieldKey),
      );
      await expect(header).toHaveAttribute("data-grid-field-key", fieldKey);
      await expect(
        await mountedGridCell(
          page,
          timelineViewSchemaId,
          selectedRow.record_id,
          fieldKey,
        ),
      ).toHaveCount(1);
    }

    await expect(
      await mountedGridCell(
        page,
        timelineViewSchemaId,
        selectedRow.record_id,
        "timeline.activity_synopsis_text",
      ),
    ).toHaveValue(rowSummariesById.get(selectedRow.record_id) ?? "");

    const fixtureViewport = page.viewportSize() ?? { width: 1440, height: 900 };
    await expectWideWorkbookTopBarChrome(page);
    for (const shortHeight of [640, 560]) {
      await page.setViewportSize({
        width: fixtureViewport.width,
        height: shortHeight,
      });
      await expectWideWorkbookTopBarChrome(page);
      await expect
        .poll(async () => (await readTimelineGridFirstLayout(page)).windowY)
        .toBe(0);
    }
    await page.setViewportSize(fixtureViewport);
    await expectWideWorkbookTopBarChrome(page);

    await page.setViewportSize({ width: 2048, height: fixtureViewport.height });
    await expect
      .poll(async () => {
        const layout = await readTimelineGridFirstLayout(page);
        return (
          layout.innerGrid.right >= layout.grid.right - 2 &&
          layout.lastColumn !== null &&
          layout.lastColumn.right >= layout.grid.right - 2
        );
      })
      .toBe(true);
    const wideLayout = await readTimelineGridFirstLayout(page);
    expect(wideLayout.grid.width).toBeGreaterThan(1600);
    expect(wideLayout.innerGrid.right).toBeGreaterThanOrEqual(
      wideLayout.grid.right - 2,
    );
    const wideLastColumn = wideLayout.lastColumn;
    expect(wideLastColumn).not.toBeNull();
    if (wideLastColumn === null) {
      throw new Error("Expected wide Timeline grid to render a last column");
    }
    expect(wideLastColumn.right).toBeGreaterThanOrEqual(
      wideLayout.grid.right - 2,
    );
    await page
      .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
      .click();
    await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
    const wideDrawerOpenLayout = await readTimelineGridFirstLayout(page);
    expect(wideDrawerOpenLayout.grid).toEqual(wideLayout.grid);
    expect(wideDrawerOpenLayout.innerGrid).toEqual(wideLayout.innerGrid);
    expect(wideDrawerOpenLayout.inspector).not.toBeNull();
    const wideDrawerLastColumn = wideDrawerOpenLayout.lastColumn;
    expect(wideDrawerLastColumn).not.toBeNull();
    if (wideDrawerLastColumn === null) {
      throw new Error(
        "Expected wide Timeline grid with inspector to render a last column",
      );
    }
    expect(wideDrawerLastColumn.right).toBeGreaterThanOrEqual(
      wideDrawerOpenLayout.grid.right - 2,
    );
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    await page.setViewportSize(fixtureViewport);
    await expect
      .poll(async () => {
        const layout = await readTimelineGridFirstLayout(page);
        return layout.innerGrid.right >= layout.grid.right - 2;
      })
      .toBe(true);
    const closedLayout = await readTimelineGridFirstLayout(page);
    expect(closedLayout.windowY).toBe(0);
    expect(closedLayout.inspector).toBeNull();
    expect(closedLayout.innerGrid.right).toBeGreaterThanOrEqual(
      closedLayout.grid.right - 2,
    );
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    await expect(
      page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await page
      .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
      .click();
    await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
    const drawerOpenLayout = await readTimelineGridFirstLayout(page);
    expect(drawerOpenLayout.grid).toEqual(closedLayout.grid);
    expect(drawerOpenLayout.innerGrid).toEqual(closedLayout.innerGrid);
    expect(drawerOpenLayout.viewBar).toEqual(closedLayout.viewBar);
    expect(drawerOpenLayout.topBar).toEqual(closedLayout.topBar);
    expect(drawerOpenLayout.inspector).not.toBeNull();
    expect(drawerOpenLayout.inspector?.top).toBeGreaterThanOrEqual(
      drawerOpenLayout.viewBar.bottom - 1,
    );
    expect(drawerOpenLayout.inspector?.top).toBeLessThanOrEqual(
      drawerOpenLayout.grid.top + 1,
    );
    expect(drawerOpenLayout.inspector?.bottom).toBeLessThanOrEqual(
      drawerOpenLayout.grid.bottom,
    );
    expect(drawerOpenLayout.windowY).toBe(0);
    await expect(
      page.getByTestId(
        workbookInspectorCloseButtonTestId(timelineViewSchemaId),
      ),
    ).toBeVisible();
    await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
      rowSummariesById.get(selectedRow.record_id) ?? "Selected timeline row",
    );
    for (const section of [
      "operational-text",
      "relationships",
      "evidence",
      "history",
    ] as const) {
      await expect(
        page.getByTestId(timelineInspectorSectionTestId(section)),
      ).toBeVisible();
    }
    const evidenceLinkResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "PATCH" &&
        response.url().endsWith(`/api/v1/records/${selectedRow.record_id}`),
    );
    await page
      .getByTestId(timelineEvidenceFileInputTestId(selectedRow.record_id))
      .setInputFiles({
        name: "default-timeline-workbook-shell.png",
        mimeType: "image/png",
        buffer: tinyPNG(),
      });
    expect((await evidenceLinkResponse).ok()).toBe(true);
    await expect(page.getByTestId(timelineInspectorMessageTestId())).toHaveText(
      "Evidence attached.",
    );
    await expect(
      page.getByTestId(timelineInspectorSectionTestId("evidence")),
    ).toContainText("Attached evidence count: 1");
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .evaluateAll((elements) => {
        (elements[0] as HTMLElement | undefined)?.click();
      });
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

    const timelineScrollportSelector = `${dataTestIdSelector(
      gridShellTestId(timelineViewSchemaId),
    )} ${gridScrollportSelector()}`;
    await expect(
      page
        .getByTestId(gridShellTestId(timelineViewSchemaId))
        .locator(gridScrollportSelector()),
    ).toBeVisible();
    await page.evaluate((selector) => {
      document.querySelector<HTMLElement>(selector)?.scrollTo({ top: 240 });
    }, timelineScrollportSelector);
    await expect
      .poll(async () => (await readTimelineGridFirstLayout(page)).scrollport)
      .toMatchObject({ scrollTop: 240 });
    const scrolledLayout = await readTimelineGridFirstLayout(page);
    expect(scrolledLayout.topBar).toEqual(closedLayout.topBar);
    expect(scrolledLayout.viewBar).toEqual(closedLayout.viewBar);
    expect(scrolledLayout.windowY).toBe(0);

    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    const summaryCell = await mountedGridCell(
      page,
      timelineViewSchemaId,
      selectedRow.record_id,
      "timeline.activity_synopsis_text",
    );
    const summaryGridCell = summaryCell.locator(
      "xpath=ancestor::*[@role='gridcell'][1]",
    );
    await summaryGridCell.click();
    await summaryGridCell.focus();
    await expect(summaryGridCell).toBeFocused();
    await expect
      .poll(() =>
        page.evaluate(
          (selector) => ({
            gridLeft:
              document.querySelector<HTMLElement>(selector)?.scrollLeft ?? -1,
            windowY: window.scrollY,
          }),
          timelineScrollportSelector,
        ),
      )
      .toMatchObject({ windowY: 0 });

    await assertViewportVisualRegression(
      page,
      "fe-v-p2-01-default-timeline-workbook-shell",
    );
  });
});

test.describe("Phase 3 workbook visual evidence", () => {
  test("V-3-GRID-01 captures the Timeline default viewport with stable row version and save-state strip", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V3GRID01"),
      "Phase 3 visual default",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID01-ROW"),
        "timeline.activity_utc_text": "2025-02-17T09:12:00Z",
        "timeline.activity_synopsis_text": "Default visual row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    const summaryCell = await mountedGridCell(
      page,
      timelineViewSchemaId,
      timelineRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await expect(
      page.getByTestId(timelineRowVersionTestId(timelineRow.record_id)),
    ).toHaveText(String(timelineRow.row_version));
    await expect(summaryCell).toHaveValue("Default visual row");
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });

    await assertViewportVisualRegression(page, "v-3-grid-01-timeline-default", {
      renderSurface: timelineViewSchemaId,
    });
  });

  test("V-3-GRID-02 captures Timeline edit save-state visuals for active cell syncing saved and conflict states", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V3GRID02"),
      "Phase 3 visual edit state",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID02-ROW"),
        "timeline.activity_utc_text": "2025-01-01T00:00:00Z",
        "timeline.activity_synopsis_text": "Editable visual row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    const saveState = page.getByTestId(saveStateTestId());
    const summaryInput = await mountedGridCell(
      page,
      timelineViewSchemaId,
      timelineRow.record_id,
      "timeline.activity_synopsis_text",
    );

    await expect(saveState).toHaveText("Saved");
    await summaryInput.focus();
    await summaryInput.fill("Active visual edit");
    await assertWorkbookGridVisualRegression(
      page,
      "v-3-grid-02-active-edit-cell",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    const patchUrl = `**/api/v1/records/${timelineRow.record_id}`;
    const hold = await holdBrowserApiRequest(page, {
      method: "PATCH",
      path: `/api/v1/records/${timelineRow.record_id}`,
    });

    try {
      await (
        await mountedGridCell(
          page,
          timelineViewSchemaId,
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        )
      ).press("Enter");
      await hold.waitForHit;
      await expect(saveState).toHaveText("Syncing");
      await assertStatusStripVisualRegression(
        page,
        "v-3-grid-02-syncing-strip",
      );
      await hold.release();
      await expect(saveState).toHaveText("Saved");
      await assertStatusStripVisualRegression(page, "v-3-grid-02-saved-strip");
    } finally {
      await hold.dispose();
    }

    const conflictHandler = async (route: Route) => {
      if (route.request().method().toUpperCase() !== "PATCH") {
        await route.fallback();
        return;
      }
      await route.fulfill({
        status: 409,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            status: 409,
            code: "same_field_conflict",
            message: "same-field conflict",
            request_id: "visual-conflict",
            retryable: false,
            details: {},
            conflict: {
              conflict_token: "visual-conflict-token",
              record_id: timelineRow.record_id,
              field_key: "timeline.activity_synopsis_text",
              conflict_resolution_class: "text_compare_merge",
              base_row_version: timelineRow.row_version,
              current_row_version: timelineRow.row_version + 1,
              base_value: "Active visual edit",
              server_value: "Server visual edit",
              client_value: "Conflict visual edit",
              server_updated_by: "visual-remote-user",
              server_updated_at: "2026-06-02T12:00:00Z",
            },
          },
        }),
      });
    };

    await page.route(patchUrl, conflictHandler);
    try {
      const conflictInput = await mountedGridCell(
        page,
        timelineViewSchemaId,
        timelineRow.record_id,
        "timeline.activity_synopsis_text",
      );
      await conflictInput.fill("Conflict visual edit");
      await conflictInput.press("Enter");
      await expect(saveState).toHaveText("Conflict");
      await assertStatusStripVisualRegression(
        page,
        "v-3-grid-02-conflict-strip",
      );
    } finally {
      await page.unroute(patchUrl, conflictHandler);
    }
  });

  test("V-3-GRID-03 captures Timeline grouped rows and currently exposed grid chrome", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V3GRID03"),
      "Phase 3 visual grouped rows",
    );
    const firstRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID03-ROWA"),
        "timeline.activity_utc_text": "2025-02-17T11:00:00Z",
        "timeline.activity_synopsis_text": "Alpha grouped row",
      },
    )) as ViewRow;
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("V3GRID03-ROWB"),
      "timeline.activity_utc_text": "2025-02-17T11:05:00Z",
      "timeline.activity_synopsis_text": "Beta grouped row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await clickTimelineRowAction(
      page,
      firstRow.record_id,
      timelineRowMarkReviewedButtonTestId(firstRow.record_id),
    );
    await expect(
      await mountedGridCell(
        page,
        timelineViewSchemaId,
        firstRow.record_id,
        "timeline.capture_state",
      ),
    ).toHaveText("reviewed");

    await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
    await expect(
      page.getByTestId(
        gridGroupRowTestId(
          timelineViewSchemaId,
          "timeline.capture_state",
          "reviewed",
        ),
      ),
    ).toBeVisible();
    await expect(
      page.getByTestId(
        gridGroupRowTestId(
          timelineViewSchemaId,
          "timeline.capture_state",
          "rough",
        ),
      ),
    ).toBeVisible();
    await expect(
      page
        .getByTestId(gridShellTestId(timelineViewSchemaId))
        .getByText("Unassigned", { exact: true }),
    ).toHaveCount(0);

    await assertWorkbookGridVisualRegression(
      page,
      "v-3-grid-03-grouped-grid",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("FE-P3 visual readiness", () => {
  test("FE-V-P3-01 Capture frozen column, resize handle, fill-down handle, edit cell, group outline row, and empty successful query grid-adapter fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV3GRID"),
      "FE-P3 grid adapter visual fixture",
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("FEV3GRID-ROW"),
      "timeline.activity_utc_text": "2026-05-31T10:00:00Z",
      "timeline.activity_synopsis_text": "FE-P3 visual adapter row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await injectFeP3GridAdapterVisualFixture(page);

    const fixture = page.locator("[data-design-fixture='fe-p3-grid-adapter']");
    await expect(fixture).toBeVisible();
    for (const fixtureId of [
      "FE-VFIX-09",
      "FE-VFIX-10",
      "FE-VFIX-11",
      "FE-VFIX-12",
      "FE-VFIX-13",
      "FE-VFIX-15",
    ]) {
      await expect(
        fixture.locator(`[data-fixture-id='${fixtureId}']`),
      ).toBeVisible();
    }
    await assertVisualRegression(
      page,
      "fe-v-p3-01-grid-adapter-fixtures",
      fixture,
    );
  });
});

test.describe("FE-P4 visual readiness", () => {
  test("FE-V-P4-01 Capture save-state strip, pending replay indication, inline edit cell, and empty successful Timeline query fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV4VISUAL"),
      "FE-P4 visual readiness",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEV4VISUAL-ROW"),
        "timeline.activity_utc_text": "2026-06-03T10:00:00Z",
        "timeline.activity_synopsis_text": "FE-P4 visual editable row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

    const summaryInput = await mountedGridCell(
      page,
      timelineViewSchemaId,
      timelineRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await expect(summaryInput).toHaveValue("FE-P4 visual editable row");
    await summaryInput.focus();
    await summaryInput.fill("FE-P4 active visual edit");
    await assertWorkbookGridVisualRegression(
      page,
      "fe-v-p4-01-active-edit-cell",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    const patchController = await installPatchTransportFailureController(page);
    try {
      patchController.disconnect();
      await (
        await mountedGridCell(
          page,
          timelineViewSchemaId,
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        )
      ).press("Enter");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "1",
      );
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "left" },
      });
      await assertVisualRegression(page, "fe-v-p4-01-pending-replay-status");

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

    const emptyIncidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV4EMPTY"),
      "FE-P4 empty Timeline query",
    );
    await page.goto(`/?incident_id=${emptyIncidentId}`);
    await maskIncidentIdentity(page, emptyIncidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, emptyIncidentId, timelineViewSchemaId))
            .length,
      )
      .toBe(0);
    await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(0);
    await expect(
      page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Create timeline row" }),
    ).toBeVisible();
    await assertWorkbookGridVisualRegression(
      page,
      "fe-v-p4-01-empty-timeline-query",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("FE-P5 workbook visual readiness", () => {
  test("FE-V-P5-01 Capture unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual resolution metadata fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEVP501"),
      "FE-P5 visual mention chip states",
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
      unresolvedRawText,
      unresolvedRow,
    } = await seedHostMentionStateFixture(page, incidentId, {
      displayPrefix: "FE-V-P5",
      hostnamePrefix: "fevp501",
      occurredAt: {
        auto: "2026-06-06T15:15:00Z",
        dismissed: "2026-06-06T15:20:00Z",
        manual: "2026-06-06T15:10:00Z",
        resolved: "2026-06-06T15:05:00Z",
        unresolved: "2026-06-06T15:00:00Z",
      },
      rawTextPrefix: "FEVP501",
      summary: {
        auto: "FE-V-P5 auto chip state",
        dismissed: "FE-V-P5 dismissed chip state",
        manual: "FE-V-P5 manual resolution metadata",
        resolved: "FE-V-P5 resolved chip state",
        unresolved: "FE-V-P5 unresolved chip state",
      },
      txnPrefix: "fevp501",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

    await openTimelineInspector(page, unresolvedRow.record_id);
    await expect(
      page
        .getByTestId(
          relationshipItemsTestId(unresolvedRow.record_id, hostRefsFieldKey),
        )
        .getByLabel(`Unresolved ${unresolvedRawText}`),
    ).toBeVisible();
    await openTimelineInspector(page, resolvedRow.record_id);
    await expect(
      page
        .getByTestId(
          relationshipItemsTestId(resolvedRow.record_id, hostRefsFieldKey),
        )
        .getByTestId(relationshipChipTestId(String(resolvedMention.item_ref))),
    ).toContainText("FE-V-P5 Resolved Target");

    await openTimelineInspector(page, manualRow.record_id);
    await page
      .getByTestId(mentionItemTestId(String(manualMention.item_ref)))
      .click();
    await page
      .getByTestId(mentionResolveTargetSelectTestId())
      .selectOption(manualTarget.record_id);
    await page.getByTestId(mentionResolveExistingButtonTestId()).click();
    await expect(
      page
        .getByTestId(
          relationshipItemsTestId(manualRow.record_id, hostRefsFieldKey),
        )
        .getByTestId(relationshipChipTestId(String(manualMention.item_ref))),
    ).toContainText("Manual");

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
    await expect(
      page
        .getByTestId(
          relationshipItemsTestId(autoRow.record_id, hostRefsFieldKey),
        )
        .getByTestId(relationshipChipTestId(String(autoItem.item_ref))),
    ).toContainText("Auto");

    await openTimelineInspector(page, dismissedRow.record_id);
    await page
      .getByTestId(mentionItemTestId(String(dismissedMention.item_ref)))
      .click();
    await page
      .getByTestId(mentionResolveTargetSelectTestId())
      .selectOption(manualTarget.record_id);
    await page.getByTestId(mentionResolveExistingButtonTestId()).click();
    await page.getByTestId(mentionDismissButtonTestId()).click();
    await expect(
      page
        .getByTestId(mentionItemTestId(String(dismissedMention.item_ref)))
        .getByLabel(`Dismissed ${dismissedRawText}`),
    ).toBeVisible();

    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await blurActiveElement(page);
    await assertViewportVisualRegression(
      page,
      "fe-v-p5-01-mention-chip-states",
    );
  });
});

test.describe("Phase 4 workbook visual evidence", () => {
  test("V-4-GRID-01 captures Timeline unresolved and resolved mention chips in the workbook grid", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V4GRID01"),
      "Phase 4 visual mention chips",
    );
    const hostRow = (await createViewRow(page, incidentId, hostsViewSchemaId, {
      client_txn_id: uniqueTxn("V4GRID01-HOST"),
      "host.display_name": "WS-023",
      "host.hostname": "ws-023.visual.example.test",
    })) as ViewRow;
    const unresolvedRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID01-UNRESOLVED"),
        "timeline.activity_utc_text": "2026-07-15T12:00:00Z",
        "timeline.activity_synopsis_text": "Unresolved mention visual row",
        [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
      },
    )) as ViewRow;
    const resolvedRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID01-RESOLVED"),
        "timeline.activity_utc_text": "2026-07-15T12:01:00Z",
        "timeline.activity_synopsis_text": "Resolved mention visual row",
        [hostRefsFieldKey]: resolvedRefPayload("WS-023", hostRow.record_id),
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await openTimelineInspector(page, unresolvedRow.record_id);
    await expect(
      page
        .getByTestId(
          relationshipItemsTestId(unresolvedRow.record_id, hostRefsFieldKey),
        )
        .getByLabel("Unresolved WS-023?"),
    ).toBeVisible();
    await openTimelineInspector(page, resolvedRow.record_id);
    await expect(
      page
        .getByTestId(
          relationshipItemsTestId(resolvedRow.record_id, hostRefsFieldKey),
        )
        .getByLabel(/^Resolved WS-023$/u),
    ).toBeVisible();

    await blurActiveElement(page);
    await assertWorkbookGridVisualRegression(
      page,
      "v-4-grid-01-mention-chips",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });

  test("V-4-GRID-02 captures Evidence access affordances on the required Evidence surface", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V4GRID02"),
      "Phase 4 visual evidence access",
    );
    const evidenceRow = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID02-EVIDENCE"),
        "evidence.title": "Visual evidence package",
        "evidence.storage_ref": "slot/visual",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      await mountedGridCell(
        page,
        evidenceViewSchemaId,
        evidenceRow.record_id,
        "evidence.title",
      ),
    ).toHaveText("Visual evidence package");
    await expect(
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidencePreviewButtonTestId(evidenceRow.record_id),
      ),
    ).toBeVisible();
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(evidenceRow.record_id)),
    ).toContainText("Requested");

    await assertWorkbookGridVisualRegression(
      page,
      "v-4-grid-02-evidence-access",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });

  test("V-4-GRID-03 captures Task Requests system view fields through the generic workbook grid", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V4GRID03"),
      "Phase 4 visual task requests",
    );
    const taskRow = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID03-TASK"),
        "task.title": "Visual task request",
        "task.task_kind": "collection",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        taskRequestsViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      await mountedGridCell(
        page,
        taskRequestsViewSchemaId,
        taskRow.record_id,
        "task.title",
      ),
    ).toHaveText("Visual task request");
    await expect(
      await mountedGridCell(
        page,
        taskRequestsViewSchemaId,
        taskRow.record_id,
        "task.status",
      ),
    ).toHaveText("open");

    await assertWorkbookGridVisualRegression(
      page,
      "v-4-grid-03-task-requests",
      taskRequestsViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("Phase 5 workbook visual evidence", () => {
  test("V-5-GRID-01 captures requested and available Evidence states on the required Evidence surface", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V5GRID01"),
      "Phase 5 visual evidence states",
    );
    const evidenceRow = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("V5GRID01-EVIDENCE"),
        "evidence.title": "Requested visual package",
        "evidence.storage_ref": "ticket://visual-request",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      await mountedGridCell(
        page,
        evidenceViewSchemaId,
        evidenceRow.record_id,
        "evidence.lifecycle_state",
      ),
    ).toHaveText("requested");
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-01-requested-evidence",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    await (
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidenceAttachFileInputTestId(evidenceRow.record_id),
      )
    ).setInputFiles({
      name: "visual-request.txt",
      mimeType: "text/plain",
      buffer: Buffer.from("phase5 visual evidence", "utf8"),
    });
    await expect(
      await mountedGridCell(
        page,
        evidenceViewSchemaId,
        evidenceRow.record_id,
        "evidence.lifecycle_state",
      ),
    ).toHaveText("available");
    await expect(
      await mountedGridCell(
        page,
        evidenceViewSchemaId,
        evidenceRow.record_id,
        "evidence.upload_state",
      ),
    ).toHaveText("available");
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-01-available-evidence",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });

  test("V-5-GRID-02 captures blocked preview feedback and Timeline evidence badges", async ({
    page,
  }, testInfo) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V5GRID02"),
      "Phase 5 visual evidence badges",
    );
    const blocked = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("V5GRID02-BLOCKED"),
        "evidence.title": "Blocked visual package",
        "evidence.storage_ref": "ticket://visual-blocked",
      },
    )) as ViewRow;
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V5GRID02-TIMELINE"),
        "timeline.activity_synopsis_text": "Visual evidence badge row",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidenceAccessMessageTestId(blocked.record_id),
      ),
    ).toContainText("Requested");
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-02-blocked-preview",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        timelineViewSchemaId,
      )}`,
    );
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await openTimelineInspector(page, timelineRow.record_id);
    await page
      .getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id))
      .setInputFiles({
        name: "visual-badge.png",
        mimeType: "image/png",
        buffer: tinyPNG(),
      });
    await expect(
      page.getByTestId(timelineInspectorSectionTestId("evidence")),
    ).toContainText("Attached evidence count: 1");
    await page.evaluate(() => {
      if (document.activeElement instanceof HTMLElement) {
        document.activeElement.blur();
      }
    });
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-02-timeline-evidence-badge",
      timelineViewSchemaId,
      {
        anchor: {
          kind: "timelineEvidenceActions",
          rowId: timelineRow.record_id,
          top: 0,
        },
        maxDiffPixels: 8_000,
        testInfo,
      },
    );
  });
});

test.describe("FE-P6 visual readiness", () => {
  test("FE-V-P6-01 Capture evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures.", async ({
    page,
  }, testInfo) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEVP601"),
      "FE-P6 visual evidence affordance",
    );
    const requested = await createVisualEvidenceRow(page, incidentId, {
      lifecycleState: "requested",
      requestedAt: "2026-05-01T10:00:00Z",
      storageRef: "case://fe-p6/requested",
      title: "01 requested evidence",
      txnPrefix: "FEVP601-REQUESTED",
    });
    const pending = await createVisualEvidenceRow(page, incidentId, {
      lifecycleState: "pending_receipt",
      requestedAt: "2026-05-01T10:05:00Z",
      storageRef: "case://fe-p6/pending",
      title: "02 pending evidence",
      txnPrefix: "FEVP601-PENDING",
    });
    const blocked = await createVisualEvidenceRow(page, incidentId, {
      lifecycleState: "quarantined",
      requestedAt: "2026-05-01T10:10:00Z",
      storageRef: "case://fe-p6/quarantined",
      title: "03 quarantined evidence",
      txnPrefix: "FEVP601-BLOCKED",
    });
    const availablePreview = await createUploadedVisualEvidence(
      page,
      incidentId,
      {
        body: Buffer.from("FE-V-P6 preview visual evidence\n", "utf8"),
        contentType: "text/plain",
        filename: "fe-v-p6-preview.txt",
        requestedAt: "2026-05-01T10:15:00Z",
        title: "04 available preview evidence",
        txnPrefix: "FEVP601-PREVIEW",
      },
    );
    const downloadHandle = await createUploadedVisualEvidence(
      page,
      incidentId,
      {
        body: Buffer.from("FE-V-P6 download handle visual evidence\n", "utf8"),
        contentType: "text/plain",
        filename: "fe-v-p6-download-handle.txt",
        requestedAt: "2026-05-01T10:20:00Z",
        title: "05 download handle evidence",
        txnPrefix: "FEVP601-DOWNLOAD",
      },
    );
    const previewBlocked = await createUploadedVisualEvidence(
      page,
      incidentId,
      {
        body: Buffer.from(
          "<!doctype html><title>FE-V-P6 unsupported preview</title>",
          "utf8",
        ),
        contentType: "text/html",
        filename: "fe-v-p6-preview-blocked.html",
        requestedAt: "2026-05-01T10:25:00Z",
        title: "06 preview blocked evidence",
        txnPrefix: "FEVP601-PREVIEW-BLOCKED",
      },
    );
    const failedHandle = await createUploadedVisualEvidence(page, incidentId, {
      body: Buffer.from("FE-V-P6 failed handle visual evidence\n", "utf8"),
      contentType: "text/plain",
      filename: "fe-v-p6-failed-handle.txt",
      requestedAt: "2026-05-01T10:30:00Z",
      title: "07 failed handle evidence",
      txnPrefix: "FEVP601-FAILED",
    });
    const inconsistentHandle = await createUploadedVisualEvidence(
      page,
      incidentId,
      {
        body: Buffer.from(
          "FE-V-P6 inconsistent handle visual evidence\n",
          "utf8",
        ),
        contentType: "text/plain",
        filename: "fe-v-p6-inconsistent-handle.txt",
        requestedAt: "2026-05-01T10:35:00Z",
        title: "08 inconsistent handle evidence",
        txnPrefix: "FEVP601-INCONSISTENT",
      },
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP601-TIMELINE"),
        "timeline.activity_utc_text": "2026-05-01T11:00:00Z",
        "timeline.activity_synopsis_text": "FE-P6 timeline evidence count",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeVisible();
    await expectVisualEvidenceState(page, requested.record_id, "requested");
    await expectVisualEvidenceState(page, pending.record_id, "pending_upload");
    await expectVisualEvidenceState(page, blocked.record_id, "blocked");
    await expectVisualEvidenceState(
      page,
      availablePreview.record_id,
      "available",
    );
    await expectVisualEvidenceState(
      page,
      downloadHandle.record_id,
      "available",
    );
    await expectVisualEvidenceState(
      page,
      previewBlocked.record_id,
      "available",
    );
    await expectVisualEvidenceState(page, failedHandle.record_id, "available");
    await expectVisualEvidenceState(
      page,
      inconsistentHandle.record_id,
      "available",
    );

    await (
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidencePreviewButtonTestId(availablePreview.record_id),
      )
    ).click();
    await expect(
      page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
    ).toBeVisible();
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
    ).toHaveText("Preview loaded inline.");
    await page
      .getByTestId(evidencePreviewPanelTestId())
      .getByRole("button", { name: "Close" })
      .click();

    const downloadPromise = page.waitForEvent("download");
    await (
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidenceDownloadButtonTestId(downloadHandle.record_id),
      )
    ).click();
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toBe("fe-v-p6-download-handle.txt");
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
    ).toHaveText("Download handle issued.");

    await (
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidencePreviewButtonTestId(previewBlocked.record_id),
      )
    ).click();
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(previewBlocked.record_id)),
    ).toContainText("evidence_access_unavailable: unsupported_preview");

    await armVisualPublicErrorFault(page, {
      path: `/api/v1/evidence-records/${failedHandle.record_id}/preview-handle`,
      reasonCode: "blob_failed",
    });
    await (
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidencePreviewButtonTestId(failedHandle.record_id),
      )
    ).click();
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(failedHandle.record_id)),
    ).toContainText("evidence_access_unavailable: blob_failed");

    await armVisualPublicErrorFault(page, {
      path: `/api/v1/evidence-records/${inconsistentHandle.record_id}/preview-handle`,
      reasonCode: "evidence_inconsistent",
    });
    await (
      await mountedGridTarget(
        page,
        evidenceViewSchemaId,
        evidencePreviewButtonTestId(inconsistentHandle.record_id),
      )
    ).click();
    await expect(
      page.getByTestId(
        evidenceAccessMessageTestId(inconsistentHandle.record_id),
      ),
    ).toContainText("evidence_access_unavailable: evidence_inconsistent");

    await assertEvidenceAccessVisualRegression(
      page,
      "fe-v-p6-01-evidence-affordance-states",
      availablePreview.record_id,
    );

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        timelineViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await openTimelineInspector(page, timelineRow.record_id);
    await expect(
      page.getByTestId(timelineInspectorSectionTestId("evidence")),
    ).toContainText("Evidence");
    await page
      .getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id))
      .setInputFiles({
        buffer: tinyPNG(),
        mimeType: "image/png",
        name: "fe-v-p6-timeline-evidence.png",
      });
    await expect(
      page.getByTestId(timelineInspectorSectionTestId("evidence")),
    ).toContainText("Attached evidence count: 1");
    await page.evaluate(() => {
      if (document.activeElement instanceof HTMLElement) {
        document.activeElement.blur();
      }
    });
    await assertWorkbookGridVisualRegression(
      page,
      "fe-v-p6-01-timeline-evidence-count",
      timelineViewSchemaId,
      {
        anchor: {
          kind: "timelineEvidenceActions",
          rowId: timelineRow.record_id,
          top: 0,
        },
        maxDiffPixels: 8_000,
        testInfo,
      },
    );
  });
});

test.describe("FE-P7 workbook visual readiness", () => {
  test("FE-V-P7-01 Capture row-gutter and cell presence markers.", async ({
    browser,
    page,
    sessionTracker,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEVP701"),
      "FE-P7 visual collaboration states",
    );
    const remote = await createIncidentMemberUser(page, incidentId, {
      display_name: "Visual Analyst",
      email: uniqueEmail("fe-v-p7-remote"),
      initial_password: "FeVP7RemotePass!",
      role: "editor",
    });
    const presenceRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP701-PRESENCE"),
        "timeline.activity_synopsis_text": "Presence visual row",
      },
    )) as ViewRow;
    const primarySocket = installIncidentSocketMonitor(page, incidentId);

    let remotePage: Page | null = null;
    try {
      await page.goto(`/?incident_id=${incidentId}`);
      await primarySocket.waitForAcceptedSocket();
      await maskIncidentIdentity(page, incidentId);

      const remoteSession = await openIncidentAsTrackedUserReady(
        browser,
        sessionTracker,
        {
          createdBy: "FE-V-P7-01",
          email: remote.email,
          incidentId,
          password: remote.initial_password,
          purpose: "FE-P7 visual presence analyst",
          readyRecordId: presenceRow.record_id,
          userId: remote.user_id,
        },
      );
      remotePage = remoteSession.page;
      await focusRemoteTimelineCellAndWaitForPresence({
        actorText: "VA",
        fieldKey: "timeline.activity_synopsis_text",
        primaryPage: page,
        recordId: presenceRow.record_id,
        remotePage,
        socketMonitor: primarySocket,
      });
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(
          presenceRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "row-gutter",
        markerTestId: rowPresenceMarkerTestId(presenceRow.record_id),
        page,
        surface: timelineViewSchemaId,
        targetTestId: gridRowGutterTestId(
          timelineViewSchemaId,
          presenceRow.record_id,
        ),
      });
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId: cellPresenceMarkerTestId(
          presenceRow.record_id,
          "timeline.activity_synopsis_text",
        ),
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(
          presenceRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });
      await assertWorkbookGridVisualRegression(
        page,
        "fe-v-p7-01-presence-markers",
        timelineViewSchemaId,
        { scroll: { top: 0, left: "left" } },
      );
    } finally {
      await remotePage?.context().close();
    }
  });

  test("FE-V-P7-01 Capture same-field conflict resolver.", async ({ page }) => {
    const fixture = await prepareFeP7ConflictVisual(page, {
      incidentKeyPrefix: "FEVP701RESOLVE",
      title: "FE-P7 visual conflict resolver",
    });
    try {
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(
          fixture.conflictRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });
      await expect(
        page.getByTestId(
          conflictMarkerTestId(
            fixture.conflictRow.record_id,
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toBeVisible();
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId: conflictMarkerTestId(
          fixture.conflictRow.record_id,
          "timeline.activity_synopsis_text",
        ),
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(
          fixture.conflictRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "right" },
      });
      await assertViewportVisualRegression(
        page,
        "fe-v-p7-01-conflict-resolver",
      );
    } finally {
      await fixture.patchController.dispose();
    }
  });

  test("FE-V-P7-01 Capture conflict save-state strip.", async ({ page }) => {
    const fixture = await prepareFeP7ConflictVisual(page, {
      incidentKeyPrefix: "FEVP701CONFLICTSTRIP",
      title: "FE-P7 visual conflict strip",
    });
    try {
      await page.getByTestId("conflict-close").click();
      await expect(page.getByTestId("conflict-resolver")).toHaveCount(0);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      await expect(
        page.getByTestId(workbookShellSlotTestId("status-strip")),
      ).toContainText("Conflict");
      await assertStatusStripVisualFixture(page, "fe-v-p7-01-conflict-strip");
    } finally {
      await fixture.patchController.dispose();
    }
  });

  test("FE-V-P7-01 Capture recovered saved-state strip.", async ({ page }) => {
    const fixture = await prepareFeP7ConflictVisual(page, {
      incidentKeyPrefix: "FEVP701RECOVERED",
      title: "FE-P7 visual recovered strip",
    });
    try {
      await page.getByTestId("conflict-keep-saved").click();
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
      await assertStatusStripVisualFixture(
        page,
        "fe-v-p7-01-recovered-saved-strip",
      );
    } finally {
      await fixture.patchController.dispose();
    }
  });

  test("FE-V-P7-01 Capture reset/invalidate refresh strip.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEVP701INVALIDATE"),
      "FE-P7 visual reset/invalidate strip",
    );
    const invalidateRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP701-INVALIDATE"),
        "timeline.activity_synopsis_text": "Invalidate visual base",
      },
    )) as ViewRow;
    const socketMonitor = installIncidentSocketMonitor(page, incidentId);

    await page.goto(`/?incident_id=${incidentId}`);
    await socketMonitor.waitForAcceptedSocket();
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await driveFeP7InvalidateRefreshVisual({
      incidentId,
      page,
      recordId: invalidateRow.record_id,
      rowVersion: invalidateRow.row_version,
      socketMonitor,
    });
  });
});

test.describe("FE-P8 workbook visual readiness", () => {
  test("FE-V-P8-01 Capture saved-view selector, active chips, grouped result, group row, default/startup state indicator, and empty successful query fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEVP801"),
      "FE-P8 visual saved view query controls",
    );
    const reviewedRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP801-REVIEWED"),
        "timeline.activity_utc_text": "2026-06-08T12:00:00Z",
        "timeline.activity_synopsis_text":
          "FE-P8 reviewed saved-view visual row",
      },
    )) as ViewRow;
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("FEVP801-ROUGH"),
      "timeline.activity_utc_text": "2026-06-08T12:05:00Z",
      "timeline.activity_synopsis_text": "FE-P8 rough grouped visual row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(
      page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
    ).toBeVisible();

    await clickTimelineRowAction(
      page,
      reviewedRow.record_id,
      timelineRowMarkReviewedButtonTestId(reviewedRow.record_id),
    );
    await expect(
      await mountedGridCell(
        page,
        timelineViewSchemaId,
        reviewedRow.record_id,
        "timeline.capture_state",
      ),
    ).toHaveText("reviewed");

    await applyFilterChip(
      page,
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    );
    await assertActiveFilterChipVisible(
      page,
      timelineViewSchemaId,
      "timeline.capture_state",
    );
    await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
    await expect(
      page.getByTestId(
        gridGroupRowTestId(
          timelineViewSchemaId,
          "timeline.capture_state",
          "reviewed",
        ),
      ),
    ).toBeVisible();

    await setSavedViewDraftName(
      page,
      timelineViewSchemaId,
      "FE-P8 visual saved-view state",
    );
    await createSavedViewFromCurrentSurface(page, timelineViewSchemaId);
    await expect(
      page.getByTestId(savedViewStatusTestId(timelineViewSchemaId)),
    ).toHaveText("Saved view created.");
    await setCurrentSavedViewAsHome(page, timelineViewSchemaId);
    await expect(
      page.getByTestId(savedViewStatusTestId(timelineViewSchemaId)),
    ).toHaveText("Home view updated.");
    await setCurrentSavedViewAsDefault(page, timelineViewSchemaId);
    await expect(
      page.getByTestId(savedViewStatusTestId(timelineViewSchemaId)),
    ).toHaveText("Default view updated.");

    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await assertViewportVisualRegression(
      page,
      "fe-v-p8-01-saved-view-query-controls",
    );

    const emptyIncidentId = await createIncident(
      page,
      uniqueIncidentKey("FEVP801EMPTY"),
      "FE-P8 empty successful Timeline query",
    );
    await page.goto(`/?incident_id=${emptyIncidentId}`);
    await maskIncidentIdentity(page, emptyIncidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, emptyIncidentId, timelineViewSchemaId))
            .length,
      )
      .toBe(0);
    await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(0);
    await expect(
      page.getByTestId(workbookInlineDraftRowTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await assertWorkbookGridVisualRegression(
      page,
      "fe-v-p8-01-empty-successful-query",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("FE-P9 workbook visual readiness", () => {
  test("FE-V-P9-01 Capture inspector Details, Relationships, Evidence, History, rollback preview, destructive confirmation, and public error fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEVP901"),
      "FE-P9 visual inspector actions",
    );
    const evidence = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP901-EVIDENCE"),
        "evidence.collector_party_text": "FE-P9 visual collector",
        "evidence.title": "FE-P9 visual attached evidence",
      },
    )) as ViewRow;
    const target = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        [hostRefsFieldKey]: collectionActionsPayload(["FE-P9 visual host"]),
        client_txn_id: uniqueTxn("FEVP901-TARGET"),
        "timeline.raw_activity_text": "FE-P9 visual inspector details",
        "timeline.activity_synopsis_text": "FE-P9 visual inspector target",
      },
    )) as ViewRow;
    const linkedTarget = (await patchRecord(page, target.record_id, {
      base_row_version: target.row_version,
      changes: [
        {
          action_payload: feP9VisualAttachedEvidencePayload(evidence.record_id),
          field_key: "timeline.attached_evidence_ids",
        },
      ],
      client_txn_id: uniqueTxn("FEVP901-LINK"),
      view_schema_id: timelineViewSchemaId,
    })) as ViewRow;
    const hostItem = requireItemByRawText(
      collectionItems(linkedTarget, hostRefsFieldKey),
      "FE-P9 visual host",
    );
    const history = await fetchFeP9VisualRecordHistory(page, target.record_id);
    const rollbackItem = requireFeP9VisualHistoryEntryAction(history);
    const rollbackAnchor = feP9VisualRollbackPreviewAnchor(
      rollbackItem,
      "history_entry",
    );

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        timelineViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await openTimelineInspector(page, target.record_id);
    for (const section of [
      "operational-text",
      "relationships",
      "evidence",
      "history",
    ] as const) {
      await expect(
        page.getByTestId(timelineInspectorSectionTestId(section)),
      ).toBeVisible();
    }
    await expect(
      page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.raw_activity_text",
          recordId: target.record_id,
          surface: "inspector",
        }),
      ),
    ).toHaveValue("FE-P9 visual inspector details");
    await expect(
      page
        .getByTestId(
          relationshipItemsTestId(target.record_id, hostRefsFieldKey),
        )
        .getByTestId(relationshipChipTestId(String(hostItem.item_ref))),
    ).toBeVisible();
    await expect(
      page.getByTestId(timelineInspectorSectionTestId("evidence")),
    ).toContainText("Attached evidence count: 0");

    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await blurActiveElement(page);
    await page
      .getByTestId(timelineInspectorSectionTestId("relationships"))
      .scrollIntoViewIfNeeded();
    await assertViewportVisualRegression(
      page,
      "fe-v-p9-01-inspector-relationships",
    );

    await page
      .getByTestId(timelineInspectorSectionTestId("history"))
      .scrollIntoViewIfNeeded();
    await clickTimelineRowAction(
      page,
      target.record_id,
      rowHistoryOpenButtonTestId(target.record_id),
    );
    await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
    await expect(
      page.getByTestId(
        feP9VisualHistoryActionTestId(rollbackItem, "history_entry"),
      ),
    ).toBeVisible();
    await page.getByTestId(rowHistoryPanelTestId()).scrollIntoViewIfNeeded();
    await assertViewportVisualRegression(page, "fe-v-p9-01-inspector-history");

    await page
      .getByTestId(feP9VisualHistoryActionTestId(rollbackItem, "history_entry"))
      .click();
    await expect(
      page.getByTestId(rowHistoryRollbackPreviewTestId(rollbackAnchor)),
    ).toContainText(rollbackItem.history_item_ref);
    await page
      .getByTestId(rowHistoryRollbackPreviewTestId(rollbackAnchor))
      .evaluate((element) => {
        element.classList.add("visual-row-history-rollback-preview");
      });
    await page
      .getByTestId(rowHistoryRollbackPreviewTestId(rollbackAnchor))
      .scrollIntoViewIfNeeded();
    await assertViewportVisualRegression(page, "fe-v-p9-01-rollback-preview");

    await page
      .getByTestId(rowHistoryRollbackCancelButtonTestId(rollbackAnchor))
      .click();
    await page.getByTestId(rowHistoryDeleteButtonTestId()).click();
    await expect(
      page.getByTestId(
        rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
      ),
    ).toContainText(target.record_id);
    await page
      .getByTestId(
        rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
      )
      .scrollIntoViewIfNeeded();
    await assertViewportVisualRegression(
      page,
      "fe-v-p9-01-destructive-confirmation",
    );

    await page
      .getByTestId(
        rowHistoryDestructiveCancelButtonTestId({ operation: "delete" }),
      )
      .click();
    await page
      .getByTestId(feP9VisualHistoryActionTestId(rollbackItem, "history_entry"))
      .click();
    await page.route(
      `**/api/v1/records/${target.record_id}/rollback`,
      (route) =>
        route.fulfill({
          contentType: "application/json",
          status: 409,
          body: JSON.stringify({
            error: {
              code: "row_version_conflict",
              message: "Rollback target is stale for FE-P9 visual fixture.",
              retryable: false,
            },
          }),
        }),
    );
    await page
      .getByTestId(rowHistoryRollbackConfirmButtonTestId(rollbackAnchor))
      .click();
    await expect(page.getByTestId(rowHistoryMessageTestId())).toContainText(
      "row_version_conflict",
    );
    await scrollVisualAnchorToScrollContainerTop(
      page,
      page.getByTestId(rowHistoryMessageTestId()),
    );
    await assertViewportVisualRegression(page, "fe-v-p9-01-public-error");
  });
});

function feP9VisualAttachedEvidencePayload(recordId: string) {
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

function feP9VisualHistoryActionTestId(
  item: FeP9VisualHistoryItem,
  action: FeP9VisualHistoryItem["available_rollback_actions"][number],
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

function feP9VisualRollbackPreviewAnchor(
  item: FeP9VisualHistoryItem,
  action: FeP9VisualHistoryItem["available_rollback_actions"][number],
) {
  return {
    action,
    historyItemRef: item.history_item_ref,
  };
}

function requireFeP9VisualHistoryEntryAction(history: FeP9VisualHistoryData) {
  const item =
    history.items.find(
      (candidate) =>
        candidate.available_rollback_actions.includes("history_entry") &&
        typeof candidate.history_entry_ref === "string" &&
        candidate.history_entry_ref.length > 0,
    ) ?? null;
  if (item === null) {
    throw new Error("missing FE-P9 visual history_entry rollback item");
  }
  return item;
}

async function fetchFeP9VisualRecordHistory(
  page: Page,
  recordId: string,
): Promise<FeP9VisualHistoryData> {
  const response = await page.request.get(
    `${apiBase}/api/v1/records/${recordId}/history`,
    { headers: await csrfHeaders(page) },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: FeP9VisualHistoryData }).data;
}

async function prepareFeP7ConflictVisual(
  page: Page,
  options: {
    incidentKeyPrefix: string;
    title: string;
  },
) {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(options.incidentKeyPrefix),
    options.title,
  );
  const conflictRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.incidentKeyPrefix}-CONFLICT`),
      "timeline.activity_utc_text": "2025-03-07T10:00:00Z",
      "timeline.activity_synopsis_text": "Conflict visual base",
    },
  )) as ViewRow;
  const queueRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.incidentKeyPrefix}-QUEUE`),
      "timeline.activity_utc_text": "2025-03-07T10:05:00Z",
      "timeline.activity_synopsis_text": "Pending visual base",
    },
  )) as ViewRow;
  const patchController = await installPatchController(page);

  await page.goto(`/?incident_id=${incidentId}`);
  await maskIncidentIdentity(page, incidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await driveRealTimelineSummaryConflict({
    afterLocalPatchHeld: async () => {
      await editTimelineSummary(
        page,
        queueRow.record_id,
        "Pending visual queued replay",
      );
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
    },
    baseRowVersion: conflictRow.row_version,
    localValue: "Conflict visual local",
    page,
    patchController,
    recordId: conflictRow.record_id,
    remoteValue: "Conflict visual server",
    txnPrefix: `${options.incidentKeyPrefix.toLowerCase()}-conflict`,
  });
  return { conflictRow, patchController };
}

async function driveFeP7InvalidateRefreshVisual({
  incidentId,
  page,
  recordId,
  rowVersion,
  socketMonitor,
}: {
  incidentId: string;
  page: Page;
  recordId: string;
  rowVersion: number;
  socketMonitor: ReturnType<typeof installIncidentSocketMonitor>;
}) {
  const removeStartAt = socketMonitor.messageCount();
  const deleteResponse = await page.request.delete(
    `${apiBase}/api/v1/records/${recordId}`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_row_version: rowVersion,
        client_txn_id: uniqueTxn("fevp701-delete"),
      },
    },
  );
  expect(deleteResponse.ok()).toBeTruthy();
  const removeMessage = await socketMonitor.waitForMessage("record_changed", {
    matches: (message) =>
      message.payload.record_id === recordId &&
      Array.isArray(message.payload.affected_views) &&
      message.payload.affected_views.some(
        (view: { change_kind?: string }) => view.change_kind === "remove",
      ),
    startAt: removeStartAt,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveCount(0);

  const queryHold = await holdBrowserApiRequest(page, {
    method: "POST",
    path: `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
  });
  let queryReleased = false;
  try {
    const invalidateStartAt = socketMonitor.messageCount();
    const restoreResponse = await page.request.post(
      `${apiBase}/api/v1/records/${recordId}/restore`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_row_version: Number(removeMessage.payload.row_version),
          client_txn_id: uniqueTxn("fevp701-restore"),
        },
      },
    );
    expect(restoreResponse.ok()).toBeTruthy();
    await socketMonitor.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload.record_id === recordId &&
        Array.isArray(message.payload.affected_views) &&
        message.payload.affected_views.some(
          (view: { change_kind?: string; view_schema_id?: string }) =>
            view.view_schema_id === timelineViewSchemaId &&
            view.change_kind === "invalidate",
        ),
      startAt: invalidateStartAt,
    });
    await queryHold.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    await expect(
      page.getByTestId(workbookShellSlotTestId("status-strip")),
    ).toContainText("Queued edits are waiting for workbook refresh.");
    await assertStatusStripVisualFixture(
      page,
      "fe-v-p7-01-reset-invalidate-notice",
    );
    const queryRefreshResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response
          .url()
          .endsWith(
            `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/query`,
          ),
    );
    queryHold.release();
    queryReleased = true;
    expect((await queryRefreshResponse).ok()).toBeTruthy();
  } finally {
    if (!queryReleased) {
      queryHold.release();
    }
    await queryHold.dispose();
  }
  await expect(
    await mountedGridCell(
      page,
      timelineViewSchemaId,
      recordId,
      "timeline.activity_synopsis_text",
    ),
  ).toHaveValue("Invalidate visual base", { timeout: 10_000 });
}

test.describe("Phase 6 workbook visual evidence", () => {
  test("V-6-GRID-01 regresses Phase 6 row-gutter and same-cell presence markers", async ({
    browser,
    page,
    sessionTracker,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V6GRID01"),
      "Phase 6 visual presence markers",
    );
    const remote = await createIncidentMemberUser(page, incidentId, {
      display_name: "Visual Analyst",
      email: uniqueEmail("phase6-v6grid01-remote"),
      initial_password: "Phase6V6Grid01!",
      role: "editor",
    });
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID01-ROW"),
        "timeline.activity_synopsis_text": "Presence visual row",
      },
    )) as ViewRow;
    const primarySocket = installIncidentSocketMonitor(page, incidentId);

    let remotePage: Page | null = null;
    try {
      await page.goto(`/?incident_id=${incidentId}`);
      await primarySocket.waitForAcceptedSocket();
      await maskIncidentIdentity(page, incidentId);
      await expect(
        await mountedGridCell(
          page,
          timelineViewSchemaId,
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      ).toHaveValue("Presence visual row");

      const remoteSession = await openIncidentAsTrackedUserReady(
        browser,
        sessionTracker,
        {
          createdBy: "V-6-GRID-01",
          email: remote.email,
          incidentId,
          password: remote.initial_password,
          purpose: "Phase 6 visual presence analyst",
          readyRecordId: timelineRow.record_id,
          userId: remote.user_id,
        },
      );
      remotePage = remoteSession.page;
      await focusRemoteTimelineCellAndWaitForPresence({
        actorText: "VA",
        fieldKey: "timeline.activity_synopsis_text",
        primaryPage: page,
        recordId: timelineRow.record_id,
        remotePage,
        socketMonitor: primarySocket,
      });
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "row-gutter",
        markerTestId: rowPresenceMarkerTestId(timelineRow.record_id),
        page,
        surface: timelineViewSchemaId,
        targetTestId: gridRowGutterTestId(
          timelineViewSchemaId,
          timelineRow.record_id,
        ),
      });
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId: cellPresenceMarkerTestId(
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        ),
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });

      await assertWorkbookGridVisualRegression(
        page,
        "v-6-grid-01-presence-markers",
        timelineViewSchemaId,
        { scroll: { top: 0, left: "left" } },
      );
    } finally {
      await remotePage?.context().close();
    }
  });

  test("V-6-GRID-02 regresses Phase 6 same-field conflict marker resolver and Conflict strip", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V6GRID02"),
      "Phase 6 visual conflict resolver",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID02-ROW"),
        "timeline.activity_synopsis_text": "Conflict visual base",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    const patchController = await installPatchController(page);
    try {
      await driveRealTimelineSummaryConflict({
        baseRowVersion: timelineRow.row_version,
        localValue: "Conflict visual local",
        page,
        patchController,
        recordId: timelineRow.record_id,
        remoteValue: "Conflict visual server",
        txnPrefix: "visual-phase6-conflict",
      });
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });
      await expect(
        page.getByTestId(
          conflictMarkerTestId(
            timelineRow.record_id,
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toBeVisible();
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId: conflictMarkerTestId(
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        ),
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });

      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "right" },
      });

      await assertViewportVisualRegression(
        page,
        "v-6-grid-02-conflict-resolver",
      );
    } finally {
      await patchController.dispose();
    }
  });

  test("V-6-GRID-03 regresses Phase 6 pending-queue save-state transitions", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V6GRID03"),
      "Phase 6 visual pending queue",
    );
    const syncRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-ROW"),
        "timeline.activity_utc_text": "2025-03-06T10:00:00Z",
        "timeline.activity_synopsis_text": "Pending visual base",
      },
    )) as ViewRow;
    const conflictRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-CONFLICT-ROW"),
        "timeline.activity_utc_text": "2025-03-06T10:05:00Z",
        "timeline.activity_synopsis_text": "Pending conflict visual base",
      },
    )) as ViewRow;
    const queuedRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-QUEUED-ROW"),
        "timeline.activity_utc_text": "2025-03-06T10:10:00Z",
        "timeline.activity_synopsis_text": "Pending queued visual base",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    const summaryInput = await mountedGridCell(
      page,
      timelineViewSchemaId,
      syncRow.record_id,
      "timeline.activity_synopsis_text",
    );
    const saveState = page.getByTestId(saveStateTestId());
    const patchController = await installPatchController(page);
    const hold = patchController.holdNextPatch({ recordId: syncRow.record_id });

    try {
      await summaryInput.fill("Pending visual syncing");
      await summaryInput.press("Enter");
      await hold.waitForHit;
      await expect(saveState).toHaveText("Syncing");
      await assertStatusStripVisualRegression(
        page,
        "v-6-grid-03-syncing-strip",
      );
      await hold.release();
      await hold.waitForCompletion;
      await expect(saveState).toHaveText("Saved");
      await assertStatusStripVisualRegression(page, "v-6-grid-03-saved-strip");

      await driveRealTimelineSummaryConflict({
        afterLocalPatchHeld: async () => {
          await editTimelineSummary(
            page,
            queuedRow.record_id,
            "Pending visual queued replay",
          );
          await expect(
            page.getByTestId(pendingQueueNoticeTestId()),
          ).toBeVisible();
        },
        baseRowVersion: conflictRow.row_version,
        localValue: "Pending visual blocked",
        page,
        patchController,
        recordId: conflictRow.record_id,
        remoteValue: "Pending visual server",
        txnPrefix: "visual-phase6-pending-conflict",
      });
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await page.getByTestId("conflict-resolver").scrollIntoViewIfNeeded();
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "right" },
      });
      await assertViewportVisualRegression(
        page,
        "v-6-grid-03-blocked-conflict",
      );

      await page.getByTestId("conflict-keep-saved").click();
      await expect(saveState).toHaveText("Saved");
    } finally {
      await patchController.dispose();
    }

    await expect(saveState).toHaveText("Saved");
    await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    await assertStatusStripVisualRegression(
      page,
      "v-6-grid-03-recovered-saved-strip",
    );
  });
});

test.describe("FE-P10 workbook visual readiness", () => {
  test("FE-V-P10-01 Capture Task Requests or Decisions, Parties link state, Communications Log, Handoff, Status Review, Lesson, keyboard focus, frozen column, resize handle, and fill-down fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.evaluate(() => {
      document.documentElement.style.zoom = "100%";
    });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEVP1001"),
      "FE-P10 coordination visual readiness",
    );
    const owner = await createIncidentMemberUser(page, incidentId, {
      display_name: "FE-P10 visual owner",
      email: uniqueEmail("fe-p10-visual-owner"),
      initial_password: "Phase10Visual1!",
      role: "editor",
    });
    const party = (await createViewRow(page, incidentId, partiesViewSchemaId, {
      client_txn_id: uniqueTxn("FEVP1001-PARTY"),
      "party.display_name": "FE-P10 Visual Party",
      "party.party_kind": "team",
    })) as ViewRow;
    const taskRow = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID03-TASK"),
        "task.task_kind": "collection",
        "task.title": "Visual task request",
      },
    )) as ViewRow;
    const decision = (await createViewRow(
      page,
      incidentId,
      decisionsViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP1001-DECISION"),
        "decision.decision_type": "containment",
        "decision.rationale": "FE-P10 visual rationale",
        "decision.summary": "FE-P10 visual decision",
      },
    )) as ViewRow;
    const comm = (await createViewRow(page, incidentId, commLogViewSchemaId, {
      client_txn_id: uniqueTxn("FEVP1001-COMM"),
      "comm_log.audience": "FE-P10 visual responders",
      "comm_log.channel_or_meeting": "FE-P10 visual bridge",
      "comm_log.comm_type": "briefing",
      "comm_log.decision_ids": {
        actions: [
          { linked_record_id: decision.record_id, op: "add_record_ref" },
        ],
        kind: "collection_actions_v1",
      },
      "comm_log.summary": "FE-P10 visual communication",
    })) as ViewRow;
    const handoff = (await createViewRow(
      page,
      incidentId,
      handoffViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP1001-HANDOFF"),
        "handoff.current_state_summary": "FE-P10 visual handoff state",
        "handoff.incoming_owner_user_id": owner.user_id,
      },
    )) as ViewRow;
    const status = (await createViewRow(
      page,
      incidentId,
      statusReviewViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP1001-STATUS"),
        "status_review.current_state_summary":
          "FE-P10 visual status review state",
      },
    )) as ViewRow;
    const lesson = (await createViewRow(page, incidentId, lessonViewSchemaId, {
      client_txn_id: uniqueTxn("FEVP1001-LESSON"),
      "lesson.summary": "FE-P10 visual lesson",
    })) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        taskRequestsViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      await mountedGridCell(
        page,
        taskRequestsViewSchemaId,
        taskRow.record_id,
        "task.title",
      ),
    ).toHaveText("Visual task request");
    await expect(
      await mountedGridCell(
        page,
        taskRequestsViewSchemaId,
        taskRow.record_id,
        "task.status",
      ),
    ).toHaveText("open");
    await normalizeWorkbookGridVisualState(page, taskRequestsViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await assertWorkbookGridVisualRegression(
      page,
      "v-4-grid-03-task-requests",
      taskRequestsViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    await injectFeP3GridAdapterVisualFixture(page);
    const fixture = page.locator("[data-design-fixture='fe-p3-grid-adapter']");
    await expect(fixture).toBeVisible();
    for (const fixtureId of [
      "FE-VFIX-09",
      "FE-VFIX-10",
      "FE-VFIX-11",
      "FE-VFIX-12",
      "FE-VFIX-13",
    ]) {
      await expect(
        fixture.locator(`[data-fixture-id='${fixtureId}']`),
      ).toBeVisible();
    }
    await assertVisualRegression(
      page,
      "fe-v-p3-01-grid-adapter-fixtures",
      fixture,
    );

    const linkedTask = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEVP1001-LINKED-TASK"),
        "task.requester_party_id": party.record_id,
        "task.requester_party_text": "FE-P10 requester",
        "task.task_kind": "follow_up",
        "task.title": "FE-P10 party-linked task",
      },
    )) as ViewRow;

    const surfaceExpectations = [
      {
        expected: "FE-P10 party-linked task",
        fieldKey: "task.title",
        recordId: linkedTask.record_id,
        surface: taskRequestsViewSchemaId,
      },
      {
        expected: "FE-P10 Visual Party",
        fieldKey: "party.display_name",
        recordId: party.record_id,
        surface: partiesViewSchemaId,
      },
      {
        expected: "FE-P10 visual communication",
        fieldKey: "comm_log.summary",
        recordId: comm.record_id,
        surface: commLogViewSchemaId,
      },
      {
        expected: "FE-P10 visual handoff state",
        fieldKey: "handoff.current_state_summary",
        recordId: handoff.record_id,
        surface: handoffViewSchemaId,
      },
      {
        expected: "FE-P10 visual status review state",
        fieldKey: "status_review.current_state_summary",
        recordId: status.record_id,
        surface: statusReviewViewSchemaId,
      },
      {
        expected: "FE-P10 visual lesson",
        fieldKey: "lesson.summary",
        recordId: lesson.record_id,
        surface: lessonViewSchemaId,
      },
    ] as const;
    for (const expectation of surfaceExpectations) {
      await page.goto(
        `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
          expectation.surface,
        )}`,
      );
      await maskIncidentIdentity(page, incidentId);
      await expect(
        page.getByTestId(gridShellTestId(expectation.surface)),
      ).toBeVisible();
      const cell = await mountedGridCell(
        page,
        expectation.surface,
        expectation.recordId,
        expectation.fieldKey,
      );
      await expect(cell).toHaveText(expectation.expected);
      await cell.focus();
      await normalizeWorkbookGridVisualState(page, expectation.surface, {
        scroll: { top: 0, left: "left" },
      });
    }

    const linkedRows = await queryViewRows(
      page,
      incidentId,
      taskRequestsViewSchemaId,
    );
    const linked = linkedRows.find(
      (candidate) => candidate.record_id === linkedTask.record_id,
    );
    const requesterPartyCell = linked?.cells["task.requester_party_id"] as
      | { value?: unknown }
      | undefined;
    expect(requesterPartyCell?.value).toBe(party.record_id);
    await test.info().attach("fe-v-p10-01-fixture-matrix.json", {
      body: Buffer.from(
        JSON.stringify(
          {
            browser_zoom_percent: 100,
            device_scale_factor: 1,
            dynamic_masks: [
              "incident identifiers",
              "generated record identifiers",
              "clock-derived labels when visible",
            ],
            editor_state:
              "FE-VFIX-12 readonly editor fixture captured in the grid-adapter fixture.",
            fixture_ids: [
              "FE-VFIX-07",
              "FE-VFIX-09",
              "FE-VFIX-10",
              "FE-VFIX-11",
              "FE-VFIX-12",
              "FE-VFIX-13",
            ],
            focus_state:
              "Each FE-P10 coordination/review surface focuses its row cell after deterministic query state is visible.",
            screenshot_scopes: [
              "task requests workbook grid",
              "grid-adapter design fixture",
            ],
            seed_id: "fe-v-p10-01-deterministic-seed",
            surfaces: surfaceExpectations.map((entry) => entry.surface),
            viewport_css_px: "1440x900",
          },
          null,
          2,
        ),
        "utf8",
      ),
      contentType: "application/json",
    });
  });
});

test.describe("FE-P11 visual readiness", () => {
  test("FE-V-P11-01 Run the owned-stack Playwright visual suite with deterministic seed data, viewport, zoom, fixture ordering, dynamic masks, scroll anchors, focus/editor state, inspector state, and post-scroll settle behavior.", async ({
    browserName: _browserName,
  }, testInfo) => {
    const registry = loadFrontendVisualFixtureRegistry();
    expect(registry.schema_id).toBe(
      "cartulary.frontend_visual_fixture_registry.v3",
    );
    expect(registry.guide_path).toBe(
      "docs/guides/cartulary_frontend_implementation_testing_guide.md",
    );
    expect(registry.fixtures.map((fixture) => fixture.fixture_id)).toEqual(
      expectedFeP11VisualFixtureIds,
    );
    for (const fixture of registry.fixtures) {
      if (fixture.status === "current") {
        expectCurrentFrontendVisualFixtureMetadata(fixture);
        continue;
      }
      expect(fixture.status).toBe("missing");
      expect(fixture.blocked_reason.length).toBeGreaterThan(0);
      expect(fixture.golden_artifacts).toEqual([]);
    }

    await testInfo.attach("fe-v-p11-01-owned-stack-visual-suite.json", {
      body: Buffer.from(
        JSON.stringify(
          {
            dynamic_masked_fixture_count: registry.fixtures.filter(
              (fixture) => !fixture.no_dynamic_regions,
            ).length,
            fixture_count: registry.fixtures.length,
            fixture_ids: registry.fixtures.map((fixture) => fixture.fixture_id),
            registry_sha256: frontendVisualFixtureRegistryDigest(),
            scroll_anchor_fixture_count: registry.fixtures.filter(
              (fixture) => fixture.scroll_normalization.anchor,
            ).length,
            selector_fixture_ids: registry.fixtures
              .filter((fixture) => fixture.capture_scope.kind === "selector")
              .map((fixture) => fixture.fixture_id),
          },
          null,
          2,
        ),
        "utf8",
      ),
      contentType: "application/json",
    });
  });

  test("FE-V-P11-02 Ensure the visual fixture matrix includes default Timeline workbook shell, unresolved/resolved entity state, same-field conflict, row-gutter presence, evidence affordance, grouped result, Task Requests or Decisions, save-state strip, frozen column, resize handle, fill-down handle, edit cell, group outline row, exposed theme states, and empty successful query.", async ({
    browserName: _browserName,
  }, testInfo) => {
    const registry = loadFrontendVisualFixtureRegistry();
    const fixturesById = new Map(
      registry.fixtures.map((fixture) => [fixture.fixture_id, fixture]),
    );
    expect([...fixturesById.keys()]).toEqual(expectedFeP11VisualFixtureIds);

    const defaultShell = fixturesById.get("FE-VFIX-01");
    expect(defaultShell?.fixture_title).toBe("Default Timeline workbook shell");
    expect(defaultShell?.capture_scope.kind).toBe("full_viewport");
    expect(defaultShell?.owner_row_ids).not.toContain("FE-V-P11-03");

    const exposedTheme = fixturesById.get("FE-VFIX-14");
    expect(exposedTheme?.fixture_title).toBe("Exposed theme states");
    expect(exposedTheme?.owner_phase_ids).toEqual(["FE-P11"]);
    expect(exposedTheme?.owner_row_ids).toEqual(["FE-V-P11-03"]);
    expect(exposedTheme?.capture_scope).toEqual({
      kind: "selector",
      selector: "[data-design-fixture='exposed-theme']",
    });
    expect(exposedTheme?.scroll_normalization.kind).toBe("not_applicable");
    expect(exposedTheme?.viewport_css_px).toBe("1280x720");
    expect(exposedTheme?.dynamic_masks).toEqual([]);
    expect(exposedTheme?.no_dynamic_regions).toBe(true);

    await testInfo.attach("fe-v-p11-02-visual-fixture-matrix.json", {
      body: Buffer.from(
        JSON.stringify(
          {
            fixture_ids: expectedFeP11VisualFixtureIds,
            matrix_titles: expectedFeP11VisualFixtureIds.map(
              (fixtureId) => fixturesById.get(fixtureId)?.fixture_title,
            ),
            non_claim_boundaries: [
              "FE-VFIX-14 does not satisfy FE-VFIX-01",
              "current fixture metadata does not close FE-P11 without row accounting",
              "visual evidence remains non-publication evidence",
            ],
            registry_sha256: frontendVisualFixtureRegistryDigest(),
          },
          null,
          2,
        ),
        "utf8",
      ),
      contentType: "application/json",
    });
  });

  test("FE-V-P11-03 Capture exposed dark_graphite token and theme states with deterministic density, color, component, focus, and semantic-state samples.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV11THEME"),
      "FE-P11 exposed theme visual fixture",
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("FEV11THEME-ROW"),
      "timeline.activity_utc_text": "2026-05-31T11:00:00Z",
      "timeline.activity_synopsis_text": "FE-P11 exposed theme fixture row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.locator("main.cartulary-shell").first()).toHaveAttribute(
      "data-cartulary-theme",
      cartularyDefaultThemeId,
    );
    await assertExposedThemeCssVariables(page);
    await injectExposedThemeVisualFixture(page);

    const fixture = page.locator("[data-design-fixture='exposed-theme']");
    await expect(fixture).toBeVisible();
    await assertVisualRegression(
      page,
      "fe-v-p11-03-exposed-theme-states",
      fixture,
    );
  });
});

type VisualEvidenceRowStateKey =
  | "available"
  | "blocked"
  | "pending_upload"
  | "requested";

async function createVisualEvidenceRow(
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
    collectorPartyText: "FE-P6 visual fixture",
    ...options,
  });
}

async function createUploadedVisualEvidence(
  page: Page,
  incidentId: string,
  options: EvidenceUploadOptions,
): Promise<ViewRow> {
  return createUploadedEvidenceFixture(page, incidentId, {
    collectorPartyText: "FE-P6 visual fixture",
    ...options,
    txnSuffixes: {
      attach: "ATTACH",
      blob: "BLOB",
      row: "ROW",
    },
  });
}

async function expectVisualEvidenceState(
  page: Page,
  recordId: string,
  stateKey: VisualEvidenceRowStateKey,
) {
  const previewButton = await mountedGridTarget(
    page,
    evidenceViewSchemaId,
    evidencePreviewButtonTestId(recordId),
  );
  const stateContainer = previewButton.locator(
    "xpath=ancestor::*[@data-evidence-state-key][1]",
  );
  await expect(stateContainer).toHaveAttribute(
    "data-evidence-state-key",
    stateKey,
  );
  if (stateKey === "available") {
    await expect(previewButton).toBeEnabled();
    await expect(
      page.getByTestId(evidenceDownloadButtonTestId(recordId)),
    ).toBeEnabled();
    return;
  }
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(recordId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(evidencePreviewButtonTestId(recordId)),
  ).toBeDisabled();
  await expect(
    page.getByTestId(evidenceDownloadButtonTestId(recordId)),
  ).toBeDisabled();
}

async function armVisualPublicErrorFault(
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
        message: "Evidence access failed for FE-P6 visual fixture.",
        method: "POST",
        path: options.path,
        retryable: false,
        status: 409,
      },
    },
  );
  expect(response.status()).toBe(201);
}

async function assertVisualRegression(
  page: Page,
  name: string,
  locator = page.getByRole("main"),
  options: { maxDiffPixels?: number; renderSurface?: string } = {},
) {
  await expect(locator).toBeVisible();
  await prepareVisualRegressionState(page);
  await attachVisualRenderDiagnostics(page, name, options.renderSurface);
  await expect(locator).toHaveScreenshot(`${name}.png`, {
    animations: "disabled",
    caret: "hide",
    ...(options.maxDiffPixels === undefined
      ? {}
      : { maxDiffPixels: options.maxDiffPixels }),
  });
}

async function assertViewportVisualRegression(
  page: Page,
  name: string,
  options: { renderSurface?: string } = {},
) {
  await prepareVisualRegressionState(page);
  await attachVisualRenderDiagnostics(page, name, options.renderSurface);
  await expect(page).toHaveScreenshot(`${name}.png`, {
    animations: "disabled",
    caret: "hide",
    fullPage: false,
  });
}

async function scrollVisualAnchorToScrollContainerTop(
  page: Page,
  locator: Locator,
) {
  await locator.evaluate((element) => {
    const scrollableOverflow = new Set(["auto", "scroll", "overlay"]);
    let container = element.parentElement;
    while (container !== null) {
      const style = window.getComputedStyle(container);
      if (
        container.scrollHeight > container.clientHeight &&
        scrollableOverflow.has(style.overflowY)
      ) {
        break;
      }
      container = container.parentElement;
    }

    const elementRect = element.getBoundingClientRect();
    if (container === null) {
      window.scrollBy({ top: elementRect.top, left: 0, behavior: "instant" });
      return;
    }

    const containerRect = container.getBoundingClientRect();
    container.scrollTop += elementRect.top - containerRect.top;
  });
  await waitForVisualLayoutFrame(page);
}

async function assertAuthGatewayVisual(page: Page, name: string) {
  await assertViewportVisualRegression(page, name);
  await test.info().attach(`${name}.png`, {
    body: await page.screenshot({ fullPage: false }),
    contentType: "image/png",
  });
}

async function assertStatusStripVisualRegression(page: Page, name: string) {
  const statusStrip = page.getByTestId(workbookShellSlotTestId("status-strip"));
  await expectStatusStripFocusAnchorVisuallyHidden(statusStrip);
  await statusStrip.scrollIntoViewIfNeeded();
  await assertVisualRegression(page, name, statusStrip);
}

async function fillAuthVisualCredentials(page: Page) {
  await page
    .getByTestId(phase1AuthTestId("login-username"))
    .fill("visual.operator@example.test");
  await page
    .getByTestId(phase1AuthTestId("login-password"))
    .fill("VisualPass1!");
}

async function fulfillAuthVisualJSON(
  route: Route,
  body: Record<string, unknown>,
  status = 200,
) {
  await route.fulfill({
    contentType: "application/json",
    status,
    body: JSON.stringify(body),
  });
}

async function fulfillAuthVisualError(
  route: Route,
  options: {
    code: string;
    details?: Record<string, unknown>;
    message: string;
    status: number;
  },
) {
  await fulfillAuthVisualJSON(
    route,
    {
      error: {
        code: options.code,
        status: options.status,
        message: options.message,
        details: options.details ?? {},
      },
    },
    options.status,
  );
}

async function fulfillAuthVisualLogin(route: Route, mode: AuthVisualLoginMode) {
  if (mode === "mfa_required") {
    await fulfillAuthVisualError(route, {
      code: "mfa_required",
      details: {
        required_second_factor_kinds: ["totp"],
      },
      message: "MFA is required.",
      status: 401,
    });
    return;
  }
  if (mode === "mfa_setup_required") {
    await fulfillAuthVisualError(route, {
      code: "mfa_setup_required",
      details: {
        bootstrap_token: "visual-bootstrap-token",
        required_setup_kinds: ["totp"],
      },
      message: "Authenticator setup is required.",
      status: 401,
    });
    return;
  }
  if (mode === "invalid_mfa") {
    await fulfillAuthVisualError(route, {
      code: "invalid_second_factor",
      message: "Invalid second factor.",
      status: 401,
    });
    return;
  }
  if (mode === "service_unavailable") {
    await fulfillAuthVisualError(route, {
      code: "service_unavailable",
      message: "Authentication service unavailable.",
      status: 503,
    });
    return;
  }
  await fulfillAuthVisualError(route, {
    code: "invalid_credentials",
    message: "Invalid credentials.",
    status: 401,
  });
}

async function assertStatusStripVisualFixture(page: Page, name: string) {
  const cloneTestId = `visual-status-strip-${name}`;
  await page.evaluate(
    ({ cloneSelector, cloneTestId, sourceSelector, sourceTestId }) => {
      document.querySelector(cloneSelector)?.remove();
      const source = document.querySelector(sourceSelector);
      if (!(source instanceof HTMLElement)) {
        throw new Error(`missing status strip fixture source ${sourceTestId}`);
      }
      const sourceRect = source.getBoundingClientRect();
      const clone = source.cloneNode(true) as HTMLElement;
      clone.setAttribute("aria-hidden", "true");
      clone.setAttribute("data-testid", cloneTestId);
      clone.style.position = "fixed";
      clone.style.insetInlineStart = "0";
      clone.style.insetBlockStart = "0";
      clone.style.inlineSize = `${Math.round(sourceRect.width)}px`;
      clone.style.margin = "0";
      clone.style.zIndex = "2147483647";
      document.body.append(clone);
    },
    {
      cloneSelector: dataTestIdSelector(cloneTestId),
      cloneTestId,
      sourceSelector: dataTestIdSelector(
        workbookShellSlotTestId("status-strip"),
      ),
      sourceTestId: workbookShellSlotTestId("status-strip"),
    },
  );
  try {
    await assertVisualRegression(page, name, page.getByTestId(cloneTestId));
  } finally {
    await page.evaluate((selector) => {
      document.querySelector(selector)?.remove();
    }, dataTestIdSelector(cloneTestId));
  }
}

async function expectStatusStripFocusAnchorVisuallyHidden(
  statusStrip: Locator,
) {
  const focusAnchor = statusStrip.getByTestId("workbook-focus-anchor");
  await expect(focusAnchor).toHaveCount(1);
  await expect
    .poll(
      async () =>
        focusAnchor.evaluate((node) => {
          const rect = node.getBoundingClientRect();
          const style = window.getComputedStyle(node);
          return {
            blockSize: Math.round(rect.height),
            clipPath: style.clipPath,
            inlineSize: Math.round(rect.width),
            overflow: style.overflow,
            position: style.position,
          };
        }),
      {
        message:
          "Expected status-strip focus anchor to remain present but visually hidden",
      },
    )
    .toEqual({
      blockSize: 1,
      clipPath: "inset(50%)",
      inlineSize: 1,
      overflow: "hidden",
      position: "absolute",
    });
}

const exposedThemeCssVars = [
  "--ct-colors-accent",
  "--ct-colors-canvas",
  "--ct-colors-surface-1",
  "--ct-colors-surface-2",
  "--ct-colors-ink",
  "--ct-colors-ink-muted",
  "--ct-colors-semantic-success",
  "--ct-colors-semantic-caution",
  "--ct-colors-semantic-conflict",
  "--ct-colors-semantic-destructive",
  "--ct-density-default-rowHeight",
  "--ct-density-default-cellPadding",
  "--ct-component-button-primary-backgroundColor",
  "--ct-component-button-primary-textColor",
  "--ct-component-button-secondary-backgroundColor",
  "--ct-component-button-secondary-border",
  "--ct-component-button-danger-textColor",
  "--ct-component-text-input-backgroundColor",
  "--ct-component-chip-backgroundColor",
  "--ct-component-grid-cell-padding",
  "--ct-component-focus-ring-border",
] as const;

async function assertExposedThemeCssVariables(page: Page) {
  const missingVars = await page.evaluate(
    (varNames) => {
      const styles = window.getComputedStyle(document.documentElement);
      return varNames.filter((name) => !styles.getPropertyValue(name).trim());
    },
    [...exposedThemeCssVars],
  );
  expect(missingVars).toEqual([]);
}

async function injectFeP3GridAdapterVisualFixture(page: Page) {
  await injectDesignFixture(page, {
    ariaLabel: "FE-P3 grid adapter visual fixtures",
    fixtureName: "fe-p3-grid-adapter",
    missingMainMessage: "Expected workbook shell main before FE-P3 fixture",
    styleText: `
      [data-design-fixture='fe-p3-grid-adapter'] {
        position: fixed;
        inset: var(--ct-spacing-xl);
        box-sizing: border-box;
        display: grid;
        grid-template-columns: minmax(0, 1fr) 20rem;
        gap: var(--ct-spacing-md);
        overflow: hidden;
        background: var(--ct-colors-canvas);
        color: var(--ct-colors-ink);
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-lg);
        padding: var(--ct-spacing-lg);
        box-shadow: var(--ct-elevation-panel);
        font-family: var(--ct-typography-ui-fontFamily);
        font-size: var(--ct-typography-ui-fontSize);
        font-weight: var(--ct-typography-ui-fontWeight);
        letter-spacing: var(--ct-typography-ui-letterSpacing);
        line-height: var(--ct-typography-ui-lineHeight);
        z-index: 1000;
      }

      [data-design-fixture='fe-p3-grid-adapter'] * {
        box-sizing: border-box;
      }

      .fe-p3-grid-fixture-table {
        min-width: 0;
        overflow: hidden;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        background: var(--ct-colors-surface-1);
      }

      .fe-p3-grid-fixture-row {
        display: grid;
        grid-template-columns: 10rem 16rem 12rem 14rem 14rem;
        min-width: 66rem;
      }

      .fe-p3-grid-fixture-head,
      .fe-p3-grid-fixture-cell {
        position: relative;
        min-width: 0;
        min-height: 3.75rem;
        display: flex;
        align-items: center;
        gap: var(--ct-spacing-xs);
        overflow: hidden;
        padding: var(--ct-density-default-cellPadding);
        border-inline-end: var(--ct-border-hairline);
        border-block-end: var(--ct-border-hairline);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
      }

      .fe-p3-grid-fixture-head {
        min-height: 3rem;
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
        font-weight: var(--ct-typography-metadata-fontWeight);
      }

      .fe-p3-grid-fixture-frozen {
        position: sticky;
        left: 0;
        z-index: 2;
        background: var(--ct-colors-surface-2);
        box-shadow: 0.75rem 0 1rem rgba(0, 0, 0, 0.28);
      }

      .fe-p3-grid-fixture-resize-handle {
        position: absolute;
        inset-block: 0.45rem;
        inset-inline-end: 0.2rem;
        width: 0.25rem;
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-hairline-focus);
      }

      .fe-p3-grid-fixture-active {
        outline: var(--ct-component-focus-ring-border);
        outline-offset: -0.2rem;
        background: var(--ct-colors-surface-3);
      }

      .fe-p3-grid-fixture-editor {
        width: 100%;
        min-width: 0;
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-sm);
        padding: 0.45rem 0.55rem;
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
        font: inherit;
      }

      .fe-p3-grid-fixture-fill {
        position: absolute;
        right: 0.2rem;
        bottom: 0.2rem;
        width: 0.65rem;
        height: 0.65rem;
        border: 0.15rem solid var(--ct-colors-hairline-focus);
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-surface-1);
      }

      .fe-p3-grid-fixture-group {
        grid-column: 1 / -1;
        min-height: 3.5rem;
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink-muted);
        font-weight: 600;
      }

      .fe-p3-grid-fixture-tree-toggle {
        display: inline-grid;
        place-items: center;
        width: 1.35rem;
        height: 1.35rem;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
      }

      .fe-p3-grid-fixture-side {
        display: grid;
        grid-template-rows: auto 1fr;
        gap: var(--ct-spacing-sm);
        min-width: 0;
      }

      .fe-p3-grid-fixture-caption {
        margin: 0;
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
      }

      .fe-p3-grid-fixture-empty {
        display: grid;
        place-items: center;
        min-height: 16rem;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink-muted);
        text-align: center;
      }

      .fe-p3-grid-fixture-empty strong {
        display: block;
        margin-block-end: var(--ct-spacing-xs);
        color: var(--ct-colors-ink);
      }
    `,
    html: `
      <div class="fe-p3-grid-fixture-table" role="grid" aria-label="Adapter fixture grid">
        <div class="fe-p3-grid-fixture-row" role="row">
          <div class="fe-p3-grid-fixture-head fe-p3-grid-fixture-frozen" role="columnheader" data-fixture-id="FE-VFIX-09">Record</div>
          <div class="fe-p3-grid-fixture-head" role="columnheader" data-fixture-id="FE-VFIX-10">Summary<span class="fe-p3-grid-fixture-resize-handle" aria-hidden="true"></span></div>
          <div class="fe-p3-grid-fixture-head" role="columnheader">State</div>
          <div class="fe-p3-grid-fixture-head" role="columnheader">Assignee</div>
          <div class="fe-p3-grid-fixture-head" role="columnheader">Last edit</div>
        </div>
        <div class="fe-p3-grid-fixture-row" role="row">
          <div class="fe-p3-grid-fixture-cell fe-p3-grid-fixture-group" role="rowheader" data-fixture-id="FE-VFIX-13"><span class="fe-p3-grid-fixture-tree-toggle" aria-hidden="true">v</span> reviewed group, 2 rows</div>
        </div>
        <div class="fe-p3-grid-fixture-row" role="row">
          <div class="fe-p3-grid-fixture-cell fe-p3-grid-fixture-frozen" role="rowheader">record-1</div>
          <div class="fe-p3-grid-fixture-cell fe-p3-grid-fixture-active" role="gridcell" data-fixture-id="FE-VFIX-12"><input class="fe-p3-grid-fixture-editor" value="Edit cell adapter" aria-label="Summary editor" readonly><span class="fe-p3-grid-fixture-fill" data-fixture-id="FE-VFIX-11" aria-hidden="true"></span></div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">reviewed</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">Analyst</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">saved</div>
        </div>
        <div class="fe-p3-grid-fixture-row" role="row">
          <div class="fe-p3-grid-fixture-cell fe-p3-grid-fixture-frozen" role="rowheader">record-2</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">Frozen column remains pinned</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">rough</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">Unassigned</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">clean</div>
        </div>
      </div>
      <aside class="fe-p3-grid-fixture-side" aria-label="Empty successful query fixture">
        <p class="fe-p3-grid-fixture-caption">Adapter-owned visual states only. Row-gutter presence remains FE-P7 and grouped-result query ownership remains FE-P8.</p>
        <div class="fe-p3-grid-fixture-empty" data-fixture-id="FE-VFIX-15">
          <span><strong>No rows match this query</strong>Successful empty result</span>
        </div>
      </aside>
    `,
  });
}

async function injectExposedThemeVisualFixture(page: Page) {
  await injectDesignFixture(page, {
    ariaLabel: "Exposed theme token state fixture",
    fixtureName: "exposed-theme",
    missingMainMessage: "Expected workbook shell main before theme fixture",
    styleText: `
      [data-design-fixture='exposed-theme'] {
        position: fixed;
        inset: var(--ct-spacing-xl);
        box-sizing: border-box;
        display: grid;
        grid-template-columns: 1.1fr 0.9fr;
        gap: var(--ct-spacing-md);
        overflow: hidden;
        background: var(--ct-colors-canvas);
        color: var(--ct-colors-ink);
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-lg);
        padding: var(--ct-spacing-lg);
        box-shadow: var(--ct-elevation-panel);
        font-family: var(--ct-typography-ui-fontFamily);
        font-size: var(--ct-typography-ui-fontSize);
        font-weight: var(--ct-typography-ui-fontWeight);
        letter-spacing: var(--ct-typography-ui-letterSpacing);
        line-height: var(--ct-typography-ui-lineHeight);
        z-index: 1000;
      }

      [data-design-fixture='exposed-theme'] * {
        box-sizing: border-box;
      }

      .theme-fixture-panel {
        display: grid;
        gap: var(--ct-spacing-sm);
        align-content: start;
        min-width: 0;
        background: var(--ct-colors-surface-1);
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        padding: var(--ct-spacing-md);
      }

      .theme-fixture-title {
        margin: 0;
        color: var(--ct-colors-ink);
        font-family: var(--ct-typography-surface-title-fontFamily);
        font-size: var(--ct-typography-surface-title-fontSize);
        font-weight: var(--ct-typography-surface-title-fontWeight);
        letter-spacing: var(--ct-typography-surface-title-letterSpacing);
        line-height: var(--ct-typography-surface-title-lineHeight);
      }

      .theme-fixture-note {
        margin: 0;
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
        font-weight: var(--ct-typography-metadata-fontWeight);
        letter-spacing: var(--ct-typography-metadata-letterSpacing);
        line-height: var(--ct-typography-metadata-lineHeight);
      }

      .theme-fixture-swatches,
      .theme-fixture-components,
      .theme-fixture-states {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: var(--ct-spacing-xs);
      }

      .theme-swatch,
      .theme-state {
        min-height: var(--ct-density-default-rowHeight);
        display: flex;
        align-items: center;
        gap: var(--ct-spacing-xs);
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-sm);
        padding: var(--ct-density-default-cellPadding);
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink);
      }

      .theme-swatch::before,
      .theme-state::before {
        content: "";
        inline-size: var(--ct-component-icon-inline-size);
        block-size: var(--ct-component-icon-inline-size);
        border-radius: var(--ct-rounded-pill);
        border: var(--ct-border-hairline);
        flex: 0 0 auto;
      }

      .theme-swatch[data-token='accent']::before {
        background: var(--ct-colors-accent);
      }

      .theme-swatch[data-token='surface']::before {
        background: var(--ct-colors-surface-3);
      }

      .theme-swatch[data-token='ink']::before {
        background: var(--ct-colors-ink);
      }

      .theme-swatch[data-token='hairline']::before {
        background: var(--ct-colors-hairline-strong);
      }

      .theme-state[data-state='success']::before {
        background: var(--ct-colors-semantic-success);
      }

      .theme-state[data-state='caution']::before {
        background: var(--ct-colors-semantic-caution);
      }

      .theme-state[data-state='conflict']::before {
        background: var(--ct-colors-semantic-conflict);
      }

      .theme-state[data-state='destructive']::before {
        background: var(--ct-colors-semantic-destructive);
      }

      .theme-button {
        min-height: var(--ct-density-default-rowHeight);
        border: 0;
        border-radius: var(--ct-component-button-primary-rounded);
        padding: var(--ct-component-button-primary-padding);
        font-family: var(--ct-typography-button-fontFamily);
        font-size: var(--ct-typography-button-fontSize);
        font-weight: var(--ct-typography-button-fontWeight);
        letter-spacing: var(--ct-typography-button-letterSpacing);
        line-height: var(--ct-typography-button-lineHeight);
      }

      .theme-button-primary {
        background: var(--ct-component-button-primary-backgroundColor);
        color: var(--ct-component-button-primary-textColor);
      }

      .theme-button-secondary {
        background: var(--ct-component-button-secondary-backgroundColor);
        color: var(--ct-component-button-secondary-textColor);
        border: var(--ct-component-button-secondary-border);
      }

      .theme-button-danger {
        background: var(--ct-component-button-danger-backgroundColor);
        color: var(--ct-component-button-danger-textColor);
        border: var(--ct-border-hairline);
      }

      .theme-input,
      .theme-grid-cell {
        min-height: var(--ct-density-default-rowHeight);
        width: 100%;
        border: var(--ct-component-text-input-border);
        border-radius: var(--ct-component-text-input-rounded);
        background: var(--ct-component-text-input-backgroundColor);
        color: var(--ct-component-text-input-textColor);
        padding: var(--ct-component-text-input-padding);
        font: inherit;
      }

      .theme-chip {
        display: inline-flex;
        align-items: center;
        width: max-content;
        border: var(--ct-component-chip-border);
        border-radius: var(--ct-component-chip-rounded);
        background: var(--ct-component-chip-backgroundColor);
        color: var(--ct-component-chip-textColor);
        padding: var(--ct-component-chip-padding);
      }

      .theme-grid-cell {
        display: flex;
        align-items: center;
        background: var(--ct-component-grid-cell-backgroundColor);
        color: var(--ct-component-grid-cell-textColor);
        padding: var(--ct-component-grid-cell-padding);
        font-family: var(--ct-typography-grid-cell-fontFamily);
        font-size: var(--ct-typography-grid-cell-fontSize);
        font-weight: var(--ct-typography-grid-cell-fontWeight);
        letter-spacing: var(--ct-typography-grid-cell-letterSpacing);
        line-height: var(--ct-typography-grid-cell-lineHeight);
      }

      .theme-focus-sample {
        outline: var(--ct-component-focus-ring-border);
        outline-offset: var(--ct-component-focus-ring-offset);
      }
    `,
    html: `
      <div class="theme-fixture-panel">
        <h2 class="theme-fixture-title">dark_graphite token states</h2>
        <p class="theme-fixture-note">Generated CSS variables applied through the workbook runtime.</p>
        <div class="theme-fixture-swatches" aria-label="Color token samples">
          <div class="theme-swatch" data-token="accent">Accent</div>
          <div class="theme-swatch" data-token="surface">Surface</div>
          <div class="theme-swatch" data-token="ink">Ink</div>
          <div class="theme-swatch" data-token="hairline">Hairline</div>
        </div>
        <div class="theme-fixture-states" aria-label="Semantic state samples">
          <div class="theme-state" data-state="success">Success state</div>
          <div class="theme-state" data-state="caution">Caution state</div>
          <div class="theme-state" data-state="conflict">Conflict state</div>
          <div class="theme-state" data-state="destructive">Destructive state</div>
        </div>
      </div>
      <div class="theme-fixture-panel">
        <h2 class="theme-fixture-title">Component and density states</h2>
        <p class="theme-fixture-note">Default density, buttons, input, chip, focus, and grid-cell tokens.</p>
        <div class="theme-fixture-components">
          <button class="theme-button theme-button-primary" type="button">Primary</button>
          <button class="theme-button theme-button-secondary theme-focus-sample" type="button">Secondary focus</button>
          <button class="theme-button theme-button-danger" type="button">Danger</button>
          <span class="theme-chip">Evidence chip</span>
        </div>
        <input class="theme-input" value="Readonly token input" readonly />
        <div class="theme-grid-cell">Grid cell typography and default density</div>
      </div>
    `,
  });
}

async function assertWorkbookGridVisualRegression(
  page: Page,
  name: string,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  try {
    await prepareVisualRegressionState(page);
    await normalizeWorkbookGridVisualState(page, surface, options);
    await normalizeWorkbookInspectorVisualState(page, options);
    await assertVisualRegression(
      page,
      name,
      page.getByTestId(gridShellTestId(surface)),
      options.maxDiffPixels === undefined
        ? { renderSurface: surface }
        : { maxDiffPixels: options.maxDiffPixels, renderSurface: surface },
    );
  } catch (error) {
    try {
      await attachWorkbookGridVisualDiagnostics(page, name, surface, options);
    } catch {
      // Preserve the assertion failure when the page is already torn down.
    }
    throw error;
  }
}

async function normalizeWorkbookInspectorVisualState(
  page: Page,
  options: GridVisualRegressionOptions,
) {
  if (!("anchor" in options)) {
    return;
  }
  switch (options.anchor.kind) {
    case "timelineEvidenceActions": {
      const evidenceSection = page.getByTestId(
        timelineInspectorSectionTestId("evidence"),
      );
      await evidenceSection.scrollIntoViewIfNeeded();
      await expect(evidenceSection).toContainText("Attached evidence count: 1");
      await waitForVisualLayoutFrame(page);
      break;
    }
  }
}

async function assertEvidenceAccessVisualRegression(
  page: Page,
  name: string,
  actionRecordId: string,
) {
  try {
    await installFeP6EvidenceAccessVisualStyle(page);
    await prepareVisualRegressionState(page);
    await setWorkbookGridScroll(page, evidenceViewSchemaId, {
      top: 0,
      left: "right",
    });
    const actionButton = await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidencePreviewButtonTestId(actionRecordId),
    );
    await waitForVisualLayoutFrame(page);
    await expect(actionButton).toBeVisible();
    await assertVisualRegression(
      page,
      name,
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
      { renderSurface: evidenceViewSchemaId },
    );
  } catch (error) {
    try {
      await attachWorkbookGridVisualDiagnostics(
        page,
        name,
        evidenceViewSchemaId,
        { scroll: { top: 0, left: "left" } },
      );
    } catch {
      // Preserve the assertion failure when the page is already torn down.
    }
    throw error;
  }
}

async function installFeP6EvidenceAccessVisualStyle(page: Page) {
  await page.evaluate((gridTestId) => {
    const styleId = "fe-p6-evidence-access-visual-style";
    document.getElementById(styleId)?.remove();
    const style = document.createElement("style");
    style.id = styleId;
    style.textContent = `
      [data-testid='${gridTestId}'] .cartulary-grid {
        grid-auto-rows: minmax(4.35rem, auto) !important;
      }

      [data-testid='${gridTestId}'] [data-grid-field-key='__cartulary_actions__'] {
        min-block-size: 4.35rem !important;
        overflow: hidden !important;
      }

      [data-testid='${gridTestId}'] [role='gridcell'][data-grid-field-key='record_id'] {
        color: transparent !important;
      }

      [data-testid='${gridTestId}'] [data-evidence-state-key] {
        align-content: start !important;
        gap: 0.12rem !important;
      }

      [data-testid='${gridTestId}'] [data-evidence-state-key] button {
        padding: 0.22rem 0.38rem !important;
        font-size: 0.72rem !important;
        line-height: 1.05 !important;
      }

      [data-testid='${gridTestId}'] [data-evidence-state-key] label {
        gap: 0 !important;
        font-size: 0.66rem !important;
        line-height: 1.05 !important;
      }

      [data-testid='${gridTestId}'] [data-evidence-state-key] input[type='file'] {
        block-size: 1px !important;
        inline-size: 1px !important;
        opacity: 0 !important;
        position: absolute !important;
      }

      [data-testid='${gridTestId}'] [data-evidence-state-key] > span {
        display: block !important;
        font-size: 0.62rem !important;
        line-height: 1.05 !important;
        overflow-wrap: anywhere !important;
      }
    `;
    document.head.append(style);
  }, gridShellTestId(evidenceViewSchemaId));
}

async function prepareVisualRegressionState(page: Page) {
  await page.evaluate(() => {
    document.documentElement.dataset.visualSnapshot = "true";
  });
  await waitForVendoredFonts(page);
  await attachFontManifestDigest();
  await maskVisualDynamicText(page);
}

async function waitForVendoredFonts(page: Page) {
  await page.evaluate(async () => {
    await Promise.all([
      document.fonts.load('400 12px "Inter"'),
      document.fonts.load('400 12px "JetBrains Mono"'),
    ]);
    await document.fonts.ready;
    const faces = Array.from(document.fonts);
    for (const family of ["Inter", "JetBrains Mono"]) {
      const familyFaces = faces.filter((face) => face.family === family);
      if (familyFaces.length === 0) {
        throw new Error(`missing vendored font-face for ${family}`);
      }
      const failedFace = familyFaces.find((face) => face.status === "error");
      if (failedFace) {
        throw new Error(`vendored font ${family} failed to load`);
      }
      if (!document.fonts.check(`400 12px "${family}"`)) {
        throw new Error(`vendored font ${family} is not ready`);
      }
    }
  });
}

async function attachFontManifestDigest() {
  await test.info().attach("font-manifest-sha256", {
    body: Buffer.from(`${fontManifestSha256()}\n`, "utf8"),
    contentType: "text/plain",
  });
}

function fontManifestSha256(): string {
  const manifest = readFileSync(
    new URL("../public/assets/fonts/FONT_MANIFEST.json", import.meta.url),
  );
  return createHash("sha256").update(manifest).digest("hex");
}

async function attachVisualRenderDiagnostics(
  page: Page,
  name: string,
  surface?: string,
) {
  const browser = page.context().browser();
  const diagnostics = await readVisualRenderDiagnostics(page, surface);
  await test.info().attach(`${name}-render-diagnostics`, {
    body: JSON.stringify(
      {
        ...diagnostics,
        fontManifestSha256: fontManifestSha256(),
        playwright: {
          browserName: browser?.browserType().name() ?? null,
          browserVersion: browser?.version() ?? null,
          nodeArch: process.arch,
          nodePlatform: process.platform,
          nodeVersion: process.version,
          viewport: page.viewportSize(),
        },
      },
      null,
      2,
    ),
    contentType: "application/json",
  });
}

async function readVisualRenderDiagnostics(page: Page, surface?: string) {
  return page.evaluate(
    ({ chipSelector, scrollportSelector, shellSelector }) => {
      const round = (value: number) => Number(value.toFixed(2));
      const rectFor = (element: Element) => {
        const rect = element.getBoundingClientRect();
        return {
          bottom: round(rect.bottom),
          height: round(rect.height),
          left: round(rect.left),
          right: round(rect.right),
          top: round(rect.top),
          width: round(rect.width),
          x: round(rect.x),
          y: round(rect.y),
        };
      };
      const textBoundsFor = (element: HTMLElement) => {
        const range = document.createRange();
        range.selectNodeContents(element);
        const rect = range.getBoundingClientRect();
        range.detach();
        return {
          bottom: round(rect.bottom),
          height: round(rect.height),
          left: round(rect.left),
          right: round(rect.right),
          text: (element.textContent ?? "").replace(/\s+/g, " ").trim(),
          top: round(rect.top),
          width: round(rect.width),
        };
      };
      const fontStyleFor = (element: HTMLElement | null) => {
        if (element === null) {
          return null;
        }
        const style = getComputedStyle(element);
        return {
          fontFamily: style.fontFamily,
          fontFeatureSettings: style.fontFeatureSettings,
          fontKerning: style.fontKerning,
          fontSize: style.fontSize,
          fontStretch: style.fontStretch,
          fontStyle: style.fontStyle,
          fontSynthesis: style.fontSynthesis,
          fontVariantLigatures: style.fontVariantLigatures,
          fontVariationSettings: style.fontVariationSettings,
          fontWeight: style.fontWeight,
          letterSpacing: style.letterSpacing,
          lineHeight: style.lineHeight,
          textRendering: style.textRendering,
          webkitFontSmoothing: style.getPropertyValue("-webkit-font-smoothing"),
        };
      };
      const shell =
        shellSelector === null
          ? null
          : document.querySelector<HTMLElement>(shellSelector);
      const scrollport =
        shell?.querySelector<HTMLElement>(scrollportSelector) ?? null;
      const gridStyleSource = scrollport ?? shell;
      const gridTextSamples = shell
        ? Array.from(
            shell.querySelectorAll<HTMLElement>(
              '[role="columnheader"], [role="rowheader"], [role="gridcell"]',
            ),
          )
            .filter((element) => {
              const rect = element.getBoundingClientRect();
              return (
                rect.width > 0 &&
                rect.height > 0 &&
                (element.textContent ?? "").trim() !== ""
              );
            })
            .slice(0, 8)
            .map((element) => ({
              ariaColIndex: element.getAttribute("aria-colindex"),
              fieldKey: element.getAttribute("data-grid-field-key"),
              rect: rectFor(element),
              role: element.getAttribute("role"),
              testId: element.getAttribute("data-testid"),
              textBounds: textBoundsFor(element),
              typography: fontStyleFor(element),
            }))
        : [];
      const chipSamples = Array.from(
        document.querySelectorAll<HTMLElement>(chipSelector),
      )
        .filter((element) => {
          const rect = element.getBoundingClientRect();
          return rect.width > 0 && rect.height > 0;
        })
        .slice(0, 8)
        .map((element) => ({
          ariaLabel: element.getAttribute("aria-label"),
          rect: rectFor(element),
          role: element.getAttribute("role"),
          testId: element.getAttribute("data-testid"),
          textBounds: textBoundsFor(element),
          typography: fontStyleFor(element),
        }));
      const userAgentData =
        "userAgentData" in navigator
          ? {
              brands: (
                navigator as Navigator & {
                  userAgentData?: {
                    brands?: unknown;
                    mobile?: unknown;
                    platform?: unknown;
                  };
                }
              ).userAgentData?.brands,
              mobile: (
                navigator as Navigator & {
                  userAgentData?: {
                    brands?: unknown;
                    mobile?: unknown;
                    platform?: unknown;
                  };
                }
              ).userAgentData?.mobile,
              platform: (
                navigator as Navigator & {
                  userAgentData?: {
                    brands?: unknown;
                    mobile?: unknown;
                    platform?: unknown;
                  };
                }
              ).userAgentData?.platform,
            }
          : null;
      return {
        computed: {
          chip: fontStyleFor(
            chipSamples.length > 0
              ? document.querySelector<HTMLElement>(chipSelector)
              : null,
          ),
          gridCell: fontStyleFor(
            gridTextSamples.length > 0
              ? (shell?.querySelector<HTMLElement>(
                  '[role="gridcell"], [role="rowheader"], [role="columnheader"]',
                ) ?? null)
              : null,
          ),
          scrollport: fontStyleFor(scrollport),
          shell: fontStyleFor(shell),
        },
        cssVars: gridStyleSource
          ? {
              cellPadding: getComputedStyle(gridStyleSource)
                .getPropertyValue("--cartulary-grid-cell-padding")
                .trim(),
              density: getComputedStyle(gridStyleSource)
                .getPropertyValue("--cartulary-grid-density")
                .trim(),
              gridCellLineHeight: getComputedStyle(gridStyleSource)
                .getPropertyValue("--ct-typography-grid-cell-lineHeight")
                .trim(),
              rowHeight: getComputedStyle(gridStyleSource)
                .getPropertyValue("--cartulary-grid-row-height")
                .trim(),
            }
          : null,
        devicePixelRatio: window.devicePixelRatio,
        fontFaces: Array.from(document.fonts).map((face) => ({
          display: face.display,
          family: face.family,
          status: face.status,
          stretch: face.stretch,
          style: face.style,
          weight: face.weight,
        })),
        fontChecks: {
          inter400: document.fonts.check('400 12px "Inter"'),
          inter500: document.fonts.check('500 12px "Inter"'),
          jetBrainsMono400: document.fonts.check('400 12px "JetBrains Mono"'),
        },
        gridTextSamples,
        navigator: {
          hardwareConcurrency: navigator.hardwareConcurrency,
          language: navigator.language,
          languages: navigator.languages,
          platform: navigator.platform,
          userAgent: navigator.userAgent,
          userAgentData,
        },
        shell: shell ? rectFor(shell) : null,
        chipSamples,
        scrollport: scrollport ? rectFor(scrollport) : null,
        visualViewport:
          window.visualViewport === null
            ? null
            : {
                height: round(window.visualViewport.height),
                offsetLeft: round(window.visualViewport.offsetLeft),
                offsetTop: round(window.visualViewport.offsetTop),
                pageLeft: round(window.visualViewport.pageLeft),
                pageTop: round(window.visualViewport.pageTop),
                scale: window.visualViewport.scale,
                width: round(window.visualViewport.width),
              },
        viewport: {
          innerHeight: window.innerHeight,
          innerWidth: window.innerWidth,
          outerHeight: window.outerHeight,
          outerWidth: window.outerWidth,
          screenHeight: window.screen.height,
          screenWidth: window.screen.width,
        },
      };
    },
    {
      chipSelector: dataTestIdPrefixSelector("chip-"),
      scrollportSelector: gridScrollportSelector(),
      shellSelector: surface === undefined ? null : gridShellSelector(surface),
    },
  );
}

async function normalizeWorkbookGridVisualState(
  page: Page,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  if ("anchor" in options) {
    await normalizeWorkbookGridAnchorVisualState(page, surface, options.anchor);
    return;
  }
  const { scroll } = options;
  await setWorkbookGridScroll(page, surface, scroll);
  await waitForVisualLayoutFrame(page);
  const expected = await setWorkbookGridScroll(page, surface, scroll);
  await expect
    .poll(() => readWorkbookGridScroll(page, surface), {
      message: `Expected ${surface} grid visual scroll to normalize shell and scrollport state`,
    })
    .toEqual(expected);
  await expect
    .poll(
      async () => {
        const headerText = await page
          .getByTestId(gridShellTestId(surface))
          .locator('[role="columnheader"]')
          .allTextContents();
        return headerText.some((text) => text.trim() !== "");
      },
      {
        message: `Expected ${surface} grid visual headers to finish rendering after scroll normalization`,
      },
    )
    .toBe(true);
  await waitForVisualLayoutFrame(page);
}

async function normalizeWorkbookGridAnchorVisualState(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
) {
  await setWorkbookGridAnchor(page, surface, anchor);
  await waitForVisualLayoutFrame(page);
  await expect
    .poll(
      async () => {
        const expected = await setWorkbookGridAnchor(page, surface, anchor);
        const state = await readWorkbookGridAnchorState(page, surface, anchor);
        return (
          state.ready &&
          state.diagnostics.scroll.shell.left === expected.shellLeft &&
          state.diagnostics.scroll.shell.top === expected.shellTop &&
          state.diagnostics.scroll.scrollport.left ===
            expected.scrollportLeft &&
          state.diagnostics.scroll.scrollport.top === expected.scrollportTop
        );
      },
      {
        message: `Expected ${surface} grid visual anchor ${anchor.kind} to reach stable geometry with normalized shell and scrollport state`,
        timeout: 6_000,
      },
    )
    .toBe(true);
  await waitForVisualLayoutFrame(page);
}

async function waitForVisualLayoutFrame(page: Page) {
  await page.evaluate(() => {
    return new Promise<void>((resolve) => {
      requestAnimationFrame(() => {
        requestAnimationFrame(() => resolve());
      });
    });
  });
}

async function setWorkbookGridScroll(
  page: Page,
  surface: string,
  scroll: GridVisualScrollState,
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ left, scrollportSelector, shellSelector, surface, top }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      const scrollport = scrollports[0];
      const maxLeft = Math.max(
        0,
        scrollport.scrollWidth - scrollport.clientWidth,
      );
      const maxTop = Math.max(
        0,
        scrollport.scrollHeight - scrollport.clientHeight,
      );
      const expectedLeft =
        left === "left" ? 0 : left === "right" ? maxLeft : left;
      const expectedTop = Math.min(Math.max(0, top), maxTop);
      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      scrollport.scrollTop = expectedTop;
      scrollport.scrollLeft = Math.min(Math.max(0, expectedLeft), maxLeft);
      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      return {
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollport.scrollTop),
        scrollportLeft: Math.round(scrollport.scrollLeft),
      };
    },
    {
      left: scroll.left,
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
      top: scroll.top,
    },
  );
}

function buildWorkbookGridAnchorSelectors(
  surface: string,
  anchor: GridVisualAnchor,
) {
  switch (anchor.kind) {
    case "timelineEvidenceActions":
      return {
        fieldKeys: [],
        scrollTargetTestId: gridRowGutterTestId(surface, anchor.rowId),
        requiredTestIds: {
          rowGutter: gridRowGutterTestId(surface, anchor.rowId),
        },
      };
  }
}

async function setWorkbookGridAnchor(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ anchor, scrollportSelector, selectors, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      const scrollport = scrollports[0];
      const maxLeft = Math.max(
        0,
        scrollport.scrollWidth - scrollport.clientWidth,
      );
      const maxTop = Math.max(
        0,
        scrollport.scrollHeight - scrollport.clientHeight,
      );
      const expectedTop = Math.min(Math.max(0, anchor.top), maxTop);
      const byTestId = (testId: string) =>
        Array.from(shell.querySelectorAll<HTMLElement>("[data-testid]")).find(
          (element) => element.getAttribute("data-testid") === testId,
        ) ?? null;

      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      scrollport.scrollTop = expectedTop;
      const scrollTarget = byTestId(selectors.scrollTargetTestId);
      if (scrollTarget === null) {
        shell.scrollLeft = Math.max(0, shell.scrollWidth - shell.clientWidth);
        scrollport.scrollLeft = maxLeft;
      } else {
        scrollTarget.scrollIntoView({ block: "nearest", inline: "center" });
      }
      shell.scrollTop = 0;

      const requiredElements = Object.values(selectors.requiredTestIds)
        .map((testId) => byTestId(testId))
        .filter((element): element is HTMLElement => element !== null);
      if (requiredElements.length > 0) {
        const shellRect = shell.getBoundingClientRect();
        const leftMost = Math.min(
          ...requiredElements.map(
            (element) => element.getBoundingClientRect().left,
          ),
        );
        const rightMost = Math.max(
          ...requiredElements.map(
            (element) => element.getBoundingClientRect().right,
          ),
        );
        const padding = 8;
        if (leftMost < shellRect.left + padding) {
          shell.scrollLeft = Math.max(
            0,
            shell.scrollLeft - (shellRect.left + padding - leftMost),
          );
        } else if (rightMost > shellRect.right - padding) {
          shell.scrollLeft = Math.min(
            Math.max(0, shell.scrollWidth - shell.clientWidth),
            shell.scrollLeft + (rightMost - (shellRect.right - padding)),
          );
        }
      }
      shell.scrollTop = 0;

      return {
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollport.scrollTop),
        scrollportLeft: Math.round(scrollport.scrollLeft),
      };
    },
    {
      anchor,
      scrollportSelector: gridScrollportSelector(),
      selectors: buildWorkbookGridAnchorSelectors(surface, anchor),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function readWorkbookGridAnchorState(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
) {
  return page.evaluate(
    async ({
      anchor,
      inspectorSelector,
      scrollportSelector,
      selectors,
      shellSelector,
      surface,
    }) => {
      const readDiagnostics = () => {
        const shell = document.querySelector<HTMLElement>(shellSelector);
        if (shell === null) {
          throw new Error(`Expected ${surface} grid shell to exist`);
        }
        const scrollports = Array.from(
          shell.querySelectorAll<HTMLElement>(scrollportSelector),
        );
        if (scrollports.length !== 1 || scrollports[0] === undefined) {
          throw new Error(
            `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
          );
        }
        const scrollport = scrollports[0];
        const shellRect = shell.getBoundingClientRect();
        const byTestId = (testId: string) =>
          Array.from(shell.querySelectorAll<HTMLElement>("[data-testid]")).find(
            (element) => element.getAttribute("data-testid") === testId,
          ) ?? null;
        const roundedRect = (element: HTMLElement | null) => {
          if (element === null) {
            return null;
          }
          const rect = element.getBoundingClientRect();
          const visible =
            rect.width > 0 &&
            rect.height > 0 &&
            rect.right <= shellRect.right - 1 &&
            rect.left >= shellRect.left + 1 &&
            rect.bottom >= shellRect.top + 1 &&
            rect.top <= shellRect.bottom - 1;
          return {
            bottom: Math.round(rect.bottom),
            height: Math.round(rect.height),
            left: Math.round(rect.left),
            right: Math.round(rect.right),
            top: Math.round(rect.top),
            visible,
            width: Math.round(rect.width),
          };
        };
        const requiredRects = Object.fromEntries(
          Object.entries(selectors.requiredTestIds).map(([key, testId]) => [
            key,
            roundedRect(byTestId(testId)),
          ]),
        );
        const visibleFieldKeys = Array.from(
          shell.querySelectorAll<HTMLElement>("[data-grid-field-key]"),
        )
          .filter((element) => {
            const rect = element.getBoundingClientRect();
            return (
              rect.width > 0 &&
              rect.height > 0 &&
              rect.right >= shellRect.left + 1 &&
              rect.left <= shellRect.right - 1 &&
              rect.bottom >= shellRect.top + 1 &&
              rect.top <= shellRect.bottom - 1
            );
          })
          .map((element) => element.getAttribute("data-grid-field-key") ?? "")
          .filter((fieldKey, index, fieldKeys) => {
            return fieldKey !== "" && fieldKeys.indexOf(fieldKey) === index;
          });
        const missingTestIds = Object.entries(selectors.requiredTestIds)
          .filter(([, testId]) => byTestId(testId) === null)
          .map(([key]) => key);
        const hiddenTestIds = Object.entries(requiredRects)
          .filter(([, rect]) => rect === null || !rect.visible)
          .map(([key]) => key);
        const missingFieldKeys = selectors.fieldKeys.filter(
          (fieldKey) => !visibleFieldKeys.includes(fieldKey),
        );
        return {
          activeElementTestId:
            document.activeElement instanceof HTMLElement
              ? document.activeElement.getAttribute("data-testid")
              : null,
          anchorKind: anchor.kind,
          inspectorOpen: document.querySelector(inspectorSelector) !== null,
          missingFieldKeys,
          missingTestIds,
          ready:
            missingTestIds.length === 0 &&
            hiddenTestIds.length === 0 &&
            missingFieldKeys.length === 0,
          requiredRects,
          screenshotTargetTestId: `${surface}-grid-shell`,
          scroll: {
            shell: {
              clientHeight: shell.clientHeight,
              clientWidth: shell.clientWidth,
              left: Math.round(shell.scrollLeft),
              maxLeft: Math.max(0, shell.scrollWidth - shell.clientWidth),
              maxTop: Math.max(0, shell.scrollHeight - shell.clientHeight),
              scrollHeight: shell.scrollHeight,
              scrollWidth: shell.scrollWidth,
              top: Math.round(shell.scrollTop),
            },
            scrollport: {
              clientHeight: scrollport.clientHeight,
              clientWidth: scrollport.clientWidth,
              left: Math.round(scrollport.scrollLeft),
              maxLeft: Math.max(
                0,
                scrollport.scrollWidth - scrollport.clientWidth,
              ),
              maxTop: Math.max(
                0,
                scrollport.scrollHeight - scrollport.clientHeight,
              ),
              scrollHeight: scrollport.scrollHeight,
              scrollWidth: scrollport.scrollWidth,
              top: Math.round(scrollport.scrollTop),
            },
          },
          surface,
          visibleFieldKeys,
        };
      };
      const nextFrame = () =>
        new Promise<void>((resolve) => {
          requestAnimationFrame(() => resolve());
        });
      const samples = [readDiagnostics()];
      await nextFrame();
      samples.push(readDiagnostics());
      await nextFrame();
      samples.push(readDiagnostics());
      const signature = (sample: (typeof samples)[number]) =>
        JSON.stringify({
          rects: sample.requiredRects,
          scroll: sample.scroll,
          visibleFieldKeys: sample.visibleFieldKeys,
        });
      const firstSample = samples[0];
      if (firstSample === undefined) {
        throw new Error("Expected grid visual anchor diagnostics to sample");
      }
      const lastSample = samples[samples.length - 1];
      if (lastSample === undefined) {
        throw new Error(
          "Expected grid visual anchor diagnostics to retain a final sample",
        );
      }
      return {
        diagnostics: lastSample,
        ready:
          samples.every((sample) => sample.ready) &&
          samples.every(
            (sample) => signature(sample) === signature(firstSample),
          ),
        samples,
      };
    },
    {
      anchor,
      inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
      scrollportSelector: gridScrollportSelector(),
      selectors: buildWorkbookGridAnchorSelectors(surface, anchor),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function attachWorkbookGridVisualDiagnostics(
  page: Page,
  name: string,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  const testInfo = options.testInfo ?? test.info();
  const diagnostics =
    "anchor" in options
      ? await readWorkbookGridAnchorState(page, surface, options.anchor)
      : await readWorkbookGridDiagnostics(page, surface);
  await testInfo.attach(`${name}-grid-diagnostics`, {
    body: JSON.stringify(
      {
        ...diagnostics,
        presentation: await readWorkbookGridPresentationDiagnostics(
          page,
          surface,
        ),
      },
      null,
      2,
    ),
    contentType: "application/json",
  });
}

async function readWorkbookGridPresentationDiagnostics(
  page: Page,
  surface: string,
) {
  return page.evaluate(
    ({ chipSelector, scrollportSelector, shellSelector, surface }) => {
      const round = (value: number) => Number(value.toFixed(2));
      const rectFor = (element: Element) => {
        const rect = element.getBoundingClientRect();
        return {
          bottom: round(rect.bottom),
          height: round(rect.height),
          left: round(rect.left),
          right: round(rect.right),
          top: round(rect.top),
          width: round(rect.width),
        };
      };
      const typographyFor = (element: HTMLElement | null) => {
        if (element === null) {
          return null;
        }
        const style = getComputedStyle(element);
        return {
          fontFamily: style.fontFamily,
          fontSize: style.fontSize,
          fontWeight: style.fontWeight,
          letterSpacing: style.letterSpacing,
          lineHeight: style.lineHeight,
          paddingBlockEnd: style.paddingBlockEnd,
          paddingBlockStart: style.paddingBlockStart,
          paddingInlineEnd: style.paddingInlineEnd,
          paddingInlineStart: style.paddingInlineStart,
        };
      };
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollport = shell.querySelector<HTMLElement>(scrollportSelector);
      if (scrollport === null) {
        throw new Error(
          `Expected ${surface} grid shell to contain ${scrollportSelector}`,
        );
      }
      const gridStyle = getComputedStyle(scrollport);
      const rowHeightVar = gridStyle
        .getPropertyValue("--cartulary-grid-row-height")
        .trim();
      const firstCell = shell.querySelector<HTMLElement>(
        '[role="gridcell"], [role="rowheader"], [role="columnheader"]',
      );
      const firstDataCell =
        shell.querySelector<HTMLElement>(
          '[data-grid-record-id] [role="gridcell"], [data-grid-record-id] [role="rowheader"]',
        ) ?? firstCell;
      const chipBounds = Array.from(
        shell.querySelectorAll<HTMLElement>(chipSelector),
      )
        .filter((element) => {
          const rect = element.getBoundingClientRect();
          return rect.width > 0 && rect.height > 0;
        })
        .slice(0, 8)
        .map((element) => ({
          ariaLabel: element.getAttribute("aria-label"),
          rect: rectFor(element),
          testId: element.getAttribute("data-testid"),
          text: (element.textContent ?? "").replace(/\s+/g, " ").trim(),
          typography: typographyFor(element),
        }));
      return {
        chipBounds,
        computedLineHeight: firstCell
          ? getComputedStyle(firstCell).lineHeight
          : null,
        computedRowHeight:
          rowHeightVar === "" ? null : round(Number.parseFloat(rowHeightVar)),
        densityId: gridStyle
          .getPropertyValue("--cartulary-grid-density")
          .trim(),
        gridCell: typographyFor(firstCell),
        observedFirstCellHeight: firstDataCell
          ? round(firstDataCell.getBoundingClientRect().height)
          : null,
        rowHeightVar,
        cellPaddingVar: gridStyle
          .getPropertyValue("--cartulary-grid-cell-padding")
          .trim(),
        tokenGridCellLineHeight: gridStyle
          .getPropertyValue("--ct-typography-grid-cell-lineHeight")
          .trim(),
      };
    },
    {
      chipSelector: dataTestIdPrefixSelector("chip-"),
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function readWorkbookGridDiagnostics(page: Page, surface: string) {
  return page.evaluate(
    ({ inspectorSelector, scrollportSelector, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      const scrollport = scrollports[0];
      const scrollportRect = scrollport.getBoundingClientRect();
      const visibleFieldKeys = Array.from(
        shell.querySelectorAll<HTMLElement>("[data-grid-field-key]"),
      )
        .filter((element) => {
          const rect = element.getBoundingClientRect();
          return (
            rect.width > 0 &&
            rect.height > 0 &&
            rect.right >= scrollportRect.left + 1 &&
            rect.left <= scrollportRect.right - 1 &&
            rect.bottom >= scrollportRect.top + 1 &&
            rect.top <= scrollportRect.bottom - 1
          );
        })
        .map((element) => element.getAttribute("data-grid-field-key") ?? "")
        .filter((fieldKey, index, fieldKeys) => {
          return fieldKey !== "" && fieldKeys.indexOf(fieldKey) === index;
        });
      return {
        activeElementTestId:
          document.activeElement instanceof HTMLElement
            ? document.activeElement.getAttribute("data-testid")
            : null,
        inspectorOpen: document.querySelector(inspectorSelector) !== null,
        ready: true,
        requiredRects: {},
        screenshotTargetTestId: `${surface}-grid-shell`,
        scroll: {
          shell: {
            clientHeight: shell.clientHeight,
            clientWidth: shell.clientWidth,
            left: Math.round(shell.scrollLeft),
            maxLeft: Math.max(0, shell.scrollWidth - shell.clientWidth),
            maxTop: Math.max(0, shell.scrollHeight - shell.clientHeight),
            scrollHeight: shell.scrollHeight,
            scrollWidth: shell.scrollWidth,
            top: Math.round(shell.scrollTop),
          },
          scrollport: {
            clientHeight: scrollport.clientHeight,
            clientWidth: scrollport.clientWidth,
            left: Math.round(scrollport.scrollLeft),
            maxLeft: Math.max(
              0,
              scrollport.scrollWidth - scrollport.clientWidth,
            ),
            maxTop: Math.max(
              0,
              scrollport.scrollHeight - scrollport.clientHeight,
            ),
            scrollHeight: scrollport.scrollHeight,
            scrollWidth: scrollport.scrollWidth,
            top: Math.round(scrollport.scrollTop),
          },
        },
        surface,
        visibleFieldKeys,
      };
    },
    {
      inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function readWorkbookGridScroll(
  page: Page,
  surface: string,
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ scrollportSelector, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      return {
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollports[0].scrollTop),
        scrollportLeft: Math.round(scrollports[0].scrollLeft),
      };
    },
    {
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function maskVisualDynamicText(page: Page) {
  await page.evaluate(() => {
    const styleId = "visual-dynamic-input-mask";
    if (!document.getElementById(styleId)) {
      const style = document.createElement("style");
      style.id = styleId;
      style.textContent = `
        [data-testid="conflict-server-actor"],
        [data-testid="conflict-server-updated-at"] {
          color: transparent !important;
          caret-color: transparent !important;
        }

        html[data-visual-snapshot="true"]
          .visual-row-history-rollback-preview > p:first-child {
          block-size: 4.5rem !important;
          color: transparent !important;
          inline-size: 100% !important;
          overflow: hidden !important;
          position: relative !important;
        }

        html[data-visual-snapshot="true"]
          .visual-row-history-rollback-preview > p:first-child::after {
          color: var(--ct-colors-ink-muted);
          content: "Preview rollback history_entry for history item hitem.VISUAL-FIXTURE on record 00000000-0000-0000-0000-000000000000 at row version 2.";
          inset: 0;
          position: absolute;
        }
      `;
      document.head.append(style);
    }
    const timestampReplacement: [RegExp, string] = [
      /\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z\b/g,
      "2025-01-01T00:00:00Z",
    ];
    const replacements: Array<[RegExp, string]> = [
      [
        /\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b/gi,
        "00000000-0000-0000-0000-000000000000",
      ],
      timestampReplacement,
      [/hitem\.[^\s<>"']+/g, "hitem.VISUAL-FIXTURE"],
      [/\bIR-[A-Z0-9-]+\b/g, "IR-VISUAL-FIXTURE"],
      [/Playwright Worker Admin \d+/g, "Playwright Worker Admin"],
    ];
    const formControlReplacements = replacements.filter(
      (replacement) => replacement !== timestampReplacement,
    );
    const walker = document.createTreeWalker(
      document.body,
      NodeFilter.SHOW_TEXT,
    );
    for (let node = walker.nextNode(); node; node = walker.nextNode()) {
      let text = node.textContent ?? "";
      for (const [pattern, replacement] of replacements) {
        text = text.replace(pattern, replacement);
      }
      node.textContent = text;
    }
    for (const element of document.querySelectorAll("input, textarea")) {
      if (
        !(element instanceof HTMLInputElement) &&
        !(element instanceof HTMLTextAreaElement)
      ) {
        continue;
      }
      let value = element.value;
      // Controlled inputs repaint their fixture values; do not race React by
      // replacing timestamp values in form controls during screenshot prep.
      for (const [pattern, replacement] of formControlReplacements) {
        value = value.replace(pattern, replacement);
      }
      element.value = value;
    }
  });
}

async function maskIncidentIdentity(page: Page, incidentId: string) {
  await page.evaluate((id) => {
    for (const node of document.querySelectorAll("p")) {
      if (node.textContent?.includes(id)) {
        node.textContent = "Incident visual-fixture";
      }
    }
  }, incidentId);
}

function tinyPNG() {
  return Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
    "base64",
  );
}
