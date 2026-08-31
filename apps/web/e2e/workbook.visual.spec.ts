import { Buffer } from "node:buffer";
import { createHash } from "node:crypto";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import type {
  CollectionActionsV1,
  EvidenceCreateRequest,
  RecordHistoryData,
  RecordHistoryItem,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import {
  applyFilterChip,
  assertActiveFilterChipVisible,
  assertMarkerAnchoredToGridTarget,
  changeGrouping,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  authTestId,
  cartularyDefaultThemeId,
  cellPresenceMarkerTestId,
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
  incidentAdministrationTestId,
  incidentControlsPanelTestId,
  incidentMembershipListTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  networkAnalysisTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  publicErrorCodeTestId,
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
  savedViewModifiedTestId,
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
  workbookAddRowButtonTestId,
  workbookConflictControlTestId,
  workbookConflictResolverTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryTestId,
  workbookFilterPopoverTriggerTestId,
  workbookFocusAnchorTestId,
  workbookInlineDraftRowTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookPresenceSummaryTestId,
  workbookResponsiveBandTestId,
  workbookShellReadyTestId,
  workbookShellSlots,
  workbookShellSlotTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  lessonViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page, Route, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";
import { gridSavedRows } from "./pages/workbookInspector";
import {
  driveRealTimelineSummaryConflict,
  focusRemoteTimelineCellAndWaitForPresence,
  installPatchController,
  installPatchTransportFailureController,
  openIncidentAsTrackedUserReady,
  successfulPatchCalls,
} from "./support/collaboration/replay";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
  requireItemByRawText,
  resolvedRefPayload,
  seedHostMentionStateFixture,
} from "./support/entities/mentions";
import {
  createEvidenceFixtureRow,
  createUploadedEvidenceFixture,
  type EvidenceUploadOptions,
} from "./support/evidence/fixtures";
import {
  importNetworkFlowCSV,
  networkFlowMinimalCSV,
  openClaimedNetworkAnalysis,
} from "./support/extensions/network_flow_activity/workspace";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { holdBrowserRequest as holdBrowserApiRequest } from "./support/transport/requestInterception";
import { createEnvironmentTestControlClient } from "./support/transport/testControlEnvironment";
import { injectDesignFixture } from "./support/visual/fixtures";
import { fetchRecordHistory } from "./support/workbook/history";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";
import {
  createSavedView,
  createSavedViewFromCurrentSurface,
  selectSavedView,
  setCurrentSavedViewAsDefault,
  setCurrentSavedViewAsHome,
  setSavedViewDraftName,
} from "./support/workbook/savedViews";

type FrontendVisualFixture = {
  blocked_reason: string;
  browser_zoom_percent: number;
  capture_scope: { kind: string; selector?: string };
  design_contract_id?: string;
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
  catalog_row_ids: string[];
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
  owner_id: string;
  schema_id: string;
  verification_id: string;
};

const expectedFrontendVisualFixtureIds = [
  "visual.fixture.claimed_network_analysis_workspace_states",
  "visual.fixture.base_inspector",
  "visual.fixture.default_timeline_workbook_shell",
  "visual.fixture.compact_desktop_workbook_shell",
  "visual.fixture.destructive_actions",
  "visual.fixture.delayed_initial_loading",
  "visual.fixture.error_presentation_loci",
  "visual.fixture.drag_fill_handle",
  "visual.fixture.edit_cell",
  "visual.fixture.empty_successful_query",
  "visual.fixture.evidence_affordance",
  "visual.fixture.component_state_matrix",
  "visual.fixture.frozen_column",
  "visual.fixture.mention_chip_state_matrix",
  "visual.fixture.narrow_desktop_workbook_shell",
  "visual.fixture.resize_handle",
  "visual.fixture.presence_overflow",
  "visual.fixture.same_field_conflict",
  "visual.fixture.save_state_strip",
  "visual.fixture.saved_view_query_controls_and_grouped_result",
  "visual.fixture.task_requests_or_decisions",
  "visual.fixture.tree_group_row",
] as const;

const expectedDesignContractIds = Array.from(
  { length: 14 },
  (_, index) => `D-VFIX-${String(index + 1).padStart(3, "0")}`,
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
  const updatingSnapshots =
    process.env.CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS === "1";
  expect(fixture.status).toBe("current");
  expect(fixture.fixture_title.length).toBeGreaterThan(0);
  expect(fixture.catalog_row_ids.length).toBeGreaterThan(0);
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
    if (!updatingSnapshots) {
      expect(
        existsSync(repoPath(artifact)),
        `${artifact} should exist`,
      ).toBeTruthy();
    }
  }
  expect(fixture.golden_artifacts).toContain(fixture.golden_filename);
}

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
    page.getByTestId(workbookViewBarQueryControlsTestId(timelineViewSchemaId)),
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
  surface: string,
  targetTestId: string,
): Promise<Locator> {
  await scrollGridTargetIntoView({ page, surface, targetTestId });
  return page.getByTestId(targetTestId);
}

async function mountedGridCell(
  page: Page,
  surface: string,
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

function tagActionsPayload(
  tagNames: readonly [string, ...string[]],
): CollectionActionsV1 {
  const [firstTagName, ...remainingTagNames] = tagNames;
  return {
    kind: "collection_actions_v1",
    actions: [
      { op: "add_tag", tag_name: firstTagName },
      ...remainingTagNames.map((tagName) => ({
        op: "add_tag" as const,
        tag_name: tagName,
      })),
    ],
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

test.describe("browser.incident-selection auth gateway visual readiness", () => {
  test("Capture auth gateway initial, focused, loading, invalid credentials, MFA required, invalid MFA, MFA setup required, service unavailable, mobile, reduced-motion, and 200%-zoom states.", async ({
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
    await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "loading",
    );
    await assertAuthGatewayVisual(page, "auth-loading");
    sessionPending = false;
    releaseAuthVisualStep(releaseSession);

    await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "anonymous",
    );
    await expect(page.getByTestId(authTestId("login-totp-code"))).toHaveCount(
      0,
    );
    await assertAuthGatewayVisual(page, "auth-initial");

    await page.getByTestId(authTestId("login-username")).focus();
    await assertAuthGatewayVisual(page, "auth-focused");

    await fillAuthVisualCredentials(page);
    loginMode = "pending";
    pendingLoginResult = "invalid_credentials";
    await page.getByTestId(authTestId("login-submit")).click();
    await expect(page.getByTestId(authTestId("login-submit"))).toHaveText(
      "Signing in...",
    );
    await assertAuthGatewayVisual(page, "auth-submitting");
    releaseAuthVisualStep(releaseLogin);
    await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toHaveText(
      "Email or password is incorrect.",
    );
    await assertAuthGatewayVisual(page, "auth-invalid-credentials");

    loginMode = "mfa_required";
    await page.getByTestId(authTestId("login-submit")).click();
    await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "mfa_required",
    );
    await assertAuthGatewayVisual(page, "auth-mfa-required");

    loginMode = "invalid_mfa";
    await page.getByTestId(authTestId("login-totp-code")).fill("000000");
    await page.getByTestId(authTestId("login-submit")).click();
    await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toHaveText(
      "The verification code is incorrect or expired.",
    );
    await assertAuthGatewayVisual(page, "auth-invalid-mfa");

    loginMode = "mfa_setup_required";
    await page.reload();
    await fillAuthVisualCredentials(page);
    await page.getByTestId(authTestId("login-submit")).click();
    await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "mfa_setup_required",
    );
    await assertAuthGatewayVisual(page, "auth-mfa-setup-required");

    loginMode = "service_unavailable";
    await page.reload();
    await fillAuthVisualCredentials(page);
    await page.getByTestId(authTestId("login-submit")).click();
    await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toHaveText(
      "Authentication is temporarily unavailable. Try again.",
    );
    await assertAuthGatewayVisual(page, "auth-service-unavailable");

    await page.setViewportSize({ width: 390, height: 844 });
    await page.reload();
    await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "anonymous",
    );
    await assertAuthGatewayVisual(page, "auth-mobile");

    await page.setViewportSize({ width: 1440, height: 900 });
    await page.emulateMedia({ reducedMotion: "reduce" });
    await page.reload();
    await assertAuthGatewayVisual(page, "auth-reduced-motion");
    await page.emulateMedia({ reducedMotion: "no-preference" });

    await page.evaluate(() => {
      document.documentElement.style.zoom = "200%";
    });
    await assertAuthGatewayVisual(page, "auth-200-zoom");
    await page.evaluate(() => {
      document.documentElement.style.zoom = "100%";
    });
  });
});

test.describe("browser.workbook-shell workbook visual readiness", () => {
  test("Capture Default Timeline workbook shell with view-bar query controls, compact sheet toolbar, dense Timeline grid, collapsed inspector default, explicit inspector opener, bottom draft row, and status strip.", async ({
    page,
  }) => {
    test.setTimeout(120_000);
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALWORKBOOKSHELL"),
      "browser.workbook-shell visual default shell",
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
        await createViewRow(page, incidentId, timelineViewSchemaId, {
          client_txn_id: uniqueTxn(`VISUALWORKBOOKSHELL-ROW-${index + 1}`),
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
        }),
      );
    }
    const rowSummariesById = new Map(
      rows.map((row, index) => [row.record_id, fixtureRows[index] ?? ""]),
    );
    const longQuerySavedView = await createSavedView(page, incidentId, {
      display_name:
        "Workbook view-bar visual resilience with a deliberately long selected saved-view name",
      query_json: {
        group_by: "timeline.capture_state",
        sort: [
          { direction: "asc", field_key: "timeline.activity_sort_ts" },
          { direction: "desc", field_key: "timeline.date_entered_sort_day" },
          { direction: "asc", field_key: "timeline.activity_synopsis_text" },
          { direction: "desc", field_key: "timeline.analyst_text" },
          { direction: "asc", field_key: "timeline.mitre_stage_text" },
          { direction: "desc", field_key: "timeline.device_object_text" },
          { direction: "asc", field_key: "timeline.ip_address_text" },
          { direction: "desc", field_key: "timeline.capture_state" },
        ],
      },
      view_schema_id: timelineViewSchemaId,
    });

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
    await expect(
      page.getByTestId(incidentAdministrationTestId("summary-key")),
    ).toHaveCount(0);
    await expect(
      page.getByTestId(incidentAdministrationTestId("patch-button")),
    ).toHaveCount(0);
    await expect(page.getByTestId(incidentMembershipListTestId())).toHaveCount(
      0,
    );
    await expect(page.getByText("Timeline mutation workbook")).toHaveCount(0);
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
      throw new Error(
        `visual.workbook-shell fixture selected unknown row ${selectedRowId}`,
      );
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
    }

    await expect(
      await mountedGridCell(
        page,
        timelineViewSchemaId,
        selectedRow.record_id,
        "timeline.activity_synopsis_text",
      ),
    ).toHaveText(rowSummariesById.get(selectedRow.record_id) ?? "");

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
    expect(wideDrawerOpenLayout.inspector).not.toBeNull();
    expect(wideDrawerOpenLayout.grid.left).toBe(wideLayout.grid.left);
    expect(wideDrawerOpenLayout.grid.right).toBeLessThan(wideLayout.grid.right);
    expect(wideDrawerOpenLayout.grid.width).toBeGreaterThan(0);
    expect(wideDrawerOpenLayout.inspector?.left).toBeGreaterThanOrEqual(
      wideDrawerOpenLayout.grid.right,
    );
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
    expect(drawerOpenLayout.grid.left).toBe(closedLayout.grid.left);
    expect(drawerOpenLayout.grid.right).toBeLessThan(closedLayout.grid.right);
    expect(drawerOpenLayout.grid.width).toBeGreaterThan(0);
    expect(drawerOpenLayout.viewBar).toEqual(closedLayout.viewBar);
    expect(drawerOpenLayout.topBar).toEqual(closedLayout.topBar);
    expect(drawerOpenLayout.inspector).not.toBeNull();
    expect(drawerOpenLayout.inspector?.left).toBeGreaterThanOrEqual(
      drawerOpenLayout.grid.right,
    );
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
      "incident-directory-default-timeline-workbook-shell",
    );

    await selectSavedView(
      page,
      timelineViewSchemaId,
      longQuerySavedView.saved_view_id,
    );
    await expect(
      page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
    ).toHaveValue(longQuerySavedView.saved_view_id);
    await expect(
      page.locator('[data-grid-data-state="refreshing"]'),
    ).toHaveCount(0);
    await expect(
      page.locator(
        '[data-grid-data-state="stale_error"], [data-grid-data-state="unavailable"]',
      ),
    ).toHaveCount(0);
    await expect(
      page.getByTestId(
        workbookViewBarQueryControlsTestId(timelineViewSchemaId),
      ),
    ).toHaveAttribute("data-hidden-query-chip-count", "1");

    await page.setViewportSize({ width: 1024, height: 720 });
    await expect(
      page.getByTestId(workbookResponsiveBandTestId()),
    ).toHaveAttribute("data-workbook-responsive-band", "narrow_desktop");
    await expect(
      page.getByTestId(
        workbookViewBarQueryControlsTestId(timelineViewSchemaId),
      ),
    ).toHaveAttribute("data-query-chip-capacity", "6");
    await expect(
      page.getByTestId(
        workbookViewBarQueryControlsTestId(timelineViewSchemaId),
      ),
    ).toHaveAttribute("data-hidden-query-chip-count", "3");
    await expect(
      page.getByTestId(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
    ).toHaveAttribute("aria-label", "Filters, 3 hidden");
    await assertViewportVisualRegression(
      page,
      "incident-directory-narrow-desktop-workbook-shell",
    );

    await page.setViewportSize({ width: 768, height: 640 });
    await expect(
      page.getByTestId(workbookResponsiveBandTestId()),
    ).toHaveAttribute("data-workbook-responsive-band", "compact_desktop");
    await expect(
      page.getByTestId(
        workbookViewBarQueryControlsTestId(timelineViewSchemaId),
      ),
    ).toHaveAttribute("data-query-chip-capacity", "0");
    await expect(
      page.getByTestId(
        workbookViewBarQueryControlsTestId(timelineViewSchemaId),
      ),
    ).toHaveAttribute("data-hidden-query-chip-count", "9");
    await expect(
      page.getByTestId(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
    ).toHaveAttribute("aria-label", "Filters, 9 hidden");
    await expect(
      page
        .getByTestId(workbookShellSlotTestId("status-strip"))
        .getByTestId(workbookPresenceSummaryTestId()),
    ).toBeVisible();
    await assertViewportVisualRegression(
      page,
      "incident-directory-compact-desktop-workbook-shell",
    );
  });
});

test.describe("workbook visual evidence", () => {
  test("captures the Timeline default viewport with stable row version and save-state strip", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALTIMELINEDEFAULT"),
      "Timeline mutation visual default",
    );
    const timelineRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALTIMELINEDEFAULT-ROW"),
        "timeline.activity_utc_text": "2025-02-17T09:12:00Z",
        "timeline.activity_synopsis_text": "Default visual row",
      },
    );

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
    await expect(summaryCell).toHaveText("Default visual row");
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });

    await assertViewportVisualRegression(
      page,
      "timeline-grid-timeline-default",
      {
        renderSurface: timelineViewSchemaId,
      },
    );
  });

  test("captures Timeline edit save-state visuals for active cell syncing saved and conflict states", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALTIMELINEEDIT"),
      "Timeline mutation visual edit state",
    );
    const timelineRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALTIMELINEEDIT-ROW"),
        "timeline.activity_utc_text": "2025-01-01T00:00:00Z",
        "timeline.activity_synopsis_text": "Editable visual row",
      },
    );

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
    await summaryInput.click();
    const summaryEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId: timelineRow.record_id,
        surface: "grid",
      }),
    );
    await expect(summaryEditor).toBeFocused();
    await summaryEditor.fill("Active visual edit");
    await assertWorkbookGridVisualRegression(
      page,
      "timeline-grid-active-edit-cell",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    const patchUrl = `**/api/v1/records/${timelineRow.record_id}`;
    const hold = await holdBrowserApiRequest(page, {
      method: "PATCH",
      path: `/api/v1/records/${timelineRow.record_id}`,
    });

    try {
      await summaryEditor.press("Enter");
      await hold.waitForHit;
      await expect(saveState).toHaveText("Syncing");
      await assertStatusStripVisualRegression(
        page,
        "timeline-grid-syncing-strip",
      );
      await hold.release();
      await expect(saveState).toHaveText("Saved");
      await assertStatusStripVisualRegression(
        page,
        "timeline-grid-saved-strip",
      );
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
      await conflictInput.click();
      const conflictEditor = page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: timelineRow.record_id,
          surface: "grid",
        }),
      );
      await expect(conflictEditor).toBeFocused();
      await conflictEditor.fill("Conflict visual edit");
      await conflictEditor.press("Enter");
      await expect(saveState).toHaveText("Conflict");
      await assertStatusStripVisualRegression(
        page,
        "timeline-grid-conflict-strip",
      );
    } finally {
      await page.unroute(patchUrl, conflictHandler);
    }
  });

  test("captures Timeline grouped rows and currently exposed grid chrome", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALTIMELINEGROUPED"),
      "Timeline mutation visual grouped rows",
    );
    const firstRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALTIMELINEGROUPED-ROWA"),
        "timeline.activity_utc_text": "2025-02-17T11:00:00Z",
        "timeline.activity_synopsis_text": "Alpha grouped row",
      },
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALTIMELINEGROUPED-ROWB"),
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
      "timeline-grid-grouped-grid",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("browser.grid-interaction visual readiness", () => {
  test("Capture frozen column, resize handle, fill-down handle, edit cell, group outline row, and empty successful query grid-adapter fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALGRIDADAPTER"),
      "browser.grid-interaction grid adapter visual fixture",
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALGRIDADAPTER-ROW"),
      "timeline.activity_utc_text": "2026-05-31T10:00:00Z",
      "timeline.activity_synopsis_text":
        "browser.grid-interaction visual adapter row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await injectFeP3GridAdapterVisualFixture(page);

    const fixture = page.locator(
      "[data-design-fixture='grid-interaction-grid-adapter']",
    );
    await expect(fixture).toBeVisible();
    for (const fixtureId of [
      "visual.fixture.frozen_column",
      "visual.fixture.resize_handle",
      "visual.fixture.drag_fill_handle",
      "visual.fixture.edit_cell",
      "visual.fixture.tree_group_row",
      "visual.fixture.empty_successful_query",
    ]) {
      await expect(
        fixture.locator(`[data-fixture-id='${fixtureId}']`),
      ).toBeVisible();
    }
    await assertVisualRegression(
      page,
      "timeline-grid-adapter-fixtures",
      fixture,
    );
  });
});

test.describe("browser.mutation-lifecycle visual readiness", () => {
  test("Capture save-state, pending replay, transaction recovery, inline edit, and empty Timeline fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALMUTATION"),
      "browser.mutation-lifecycle visual readiness",
    );
    const timelineRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALMUTATION-ROW"),
        "timeline.activity_utc_text": "2026-06-03T10:00:00Z",
        "timeline.activity_synopsis_text":
          "browser.mutation-lifecycle visual editable row",
      },
    );

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
    await expect(summaryInput).toHaveText(
      "browser.mutation-lifecycle visual editable row",
    );
    await summaryInput.click();
    const summaryEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId: timelineRow.record_id,
        surface: "grid",
      }),
    );
    await expect(summaryEditor).toBeFocused();
    await summaryEditor.fill("browser.mutation-lifecycle active visual edit");
    await assertWorkbookGridVisualRegression(
      page,
      "timeline-mutation-active-edit-cell",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    const patchController = await installPatchTransportFailureController(page);
    try {
      patchController.disconnect();
      await summaryEditor.press("Enter");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "1",
      );
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "left" },
      });
      await assertVisualRegression(
        page,
        "timeline-mutation-pending-replay-status",
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

    const conflictController = await installPatchController(page);
    try {
      conflictController.failNextPatch(409, "client_txn_conflict", {
        recordId: timelineRow.record_id,
      });
      const blockedSummary = await mountedGridCell(
        page,
        timelineViewSchemaId,
        timelineRow.record_id,
        "timeline.activity_synopsis_text",
      );
      await blockedSummary.click();
      const blockedEditor = page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: timelineRow.record_id,
          surface: "grid",
        }),
      );
      await blockedEditor.fill(
        "browser.mutation-lifecycle blocked transaction edit",
      );
      await blockedEditor.press("Enter");
      await expect.poll(() => conflictController.calls.length).toBe(1);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      await expect(
        page.getByTestId(workbookEditRecoveryTestId()),
      ).toBeVisible();
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "left" },
      });
      await assertVisualRegression(
        page,
        "timeline-mutation-transaction-recovery-panel",
      );

      await page.setViewportSize({ width: 1024, height: 720 });
      await expect(
        page.getByTestId(workbookEditRecoveryTestId()),
      ).toBeVisible();
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "left" },
      });
      await assertViewportVisualRegression(
        page,
        "timeline-mutation-transaction-recovery-panel-narrow",
      );

      await page.setViewportSize({ width: 768, height: 640 });
      await expect(
        page.getByTestId(workbookEditRecoveryTestId()),
      ).toBeVisible();
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "left" },
      });
      await assertViewportVisualRegression(
        page,
        "timeline-mutation-transaction-recovery-panel-compact",
      );
      await page.setViewportSize({ width: 1440, height: 900 });

      await page.getByTestId(workbookEditRecoveryDiscardButtonTestId()).click();
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(workbookEditRecoveryTestId())).toHaveCount(
        0,
      );
    } finally {
      await conflictController.dispose();
    }

    const emptyIncidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALEMPTYQUERY"),
      "browser.mutation-lifecycle empty Timeline query",
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
      "timeline-mutation-empty-timeline-query",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("browser.entity-linking workbook visual readiness", () => {
  test("Capture unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual resolution metadata fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALENTITYLINKING"),
      "browser.entity-linking visual mention chip states",
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
      displayPrefix: "visual.entity-linking",
      hostnamePrefix: "visual-entity-linking",
      occurredAt: {
        auto: "2026-06-06T15:15:00Z",
        dismissed: "2026-06-06T15:20:00Z",
        manual: "2026-06-06T15:10:00Z",
        resolved: "2026-06-06T15:05:00Z",
        unresolved: "2026-06-06T15:00:00Z",
      },
      rawTextPrefix: "VISUALENTITYLINKING",
      summary: {
        auto: "visual.entity-linking auto chip state",
        dismissed: "visual.entity-linking dismissed chip state",
        manual: "visual.entity-linking manual resolution metadata",
        resolved: "visual.entity-linking resolved chip state",
        unresolved: "visual.entity-linking unresolved chip state",
      },
      txnPrefix: "visual-entity-linking",
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
    ).toContainText("visual.entity-linking Resolved Target");

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
    const dismissedMentionItem = page.getByTestId(
      mentionItemTestId(String(dismissedMention.item_ref)),
    );
    await dismissedMentionItem.click();
    await expect(
      page.getByTestId(mentionRestoreUnresolvedButtonTestId()),
    ).toBeVisible();
    await scrollVisualAnchorToScrollContainerTop(page, dismissedMentionItem, {
      clipTopPixels: 7,
    });
    await blurActiveElement(page);
    const chipFixture = page.locator("main.cartulary-shell").first();
    await chipFixture.evaluate((element) => {
      element.setAttribute("data-design-fixture", "chips");
    });
    await assertVisualRegression(
      page,
      "entity-mention-chip-states",
      chipFixture,
    );
  });
});

test.describe("workbook visual evidence", () => {
  test("captures Timeline unresolved and resolved mention chips in the workbook grid", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALENTITYLINKINGAUX"),
      "Entity linking visual mention chips",
    );
    const hostRow = await createViewRow(page, incidentId, hostsViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALENTITYLINKINGAUX-HOST"),
      "host.display_name": "WS-023",
      "host.hostname": "ws-023.visual.example.test",
    });
    const unresolvedRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALENTITYLINKINGAUX-UNRESOLVED"),
        "timeline.activity_utc_text": "2026-07-15T12:00:00Z",
        "timeline.activity_synopsis_text": "Unresolved mention visual row",
        [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
      },
    );
    const resolvedRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALENTITYLINKINGAUX-RESOLVED"),
        "timeline.activity_utc_text": "2026-07-15T12:01:00Z",
        "timeline.activity_synopsis_text": "Resolved mention visual row",
        [hostRefsFieldKey]: resolvedRefPayload("WS-023", hostRow.record_id),
      },
    );

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
      "record-relationships-mention-chips",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });

  test("captures Evidence access affordances on the required Evidence surface", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALEVIDENCEACCESS"),
      "Entity linking visual evidence access",
    );
    const evidenceRow = await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALEVIDENCEACCESS-EVIDENCE"),
        "evidence.title": "Visual evidence package",
        "evidence.storage_ref": "slot/visual",
      },
    );

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
      "record-relationships-evidence-access",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("workbook visual evidence", () => {
  test("captures requested and available Evidence states on the required Evidence surface", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALEVIDENCEAVAILABLE"),
      "Evidence lifecycle visual evidence states",
    );
    const evidenceRow = await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALEVIDENCEAVAILABLE-EVIDENCE"),
        "evidence.title": "Requested visual package",
        "evidence.storage_ref": "ticket://visual-request",
      },
    );

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
      "evidence-grid-requested-evidence",
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
      buffer: Buffer.from("evidence_lifecycle visual evidence", "utf8"),
    });
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(evidenceRow.record_id)),
    ).toHaveText("Evidence attached.", { timeout: 30_000 });
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
      "evidence-grid-available-evidence",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });

  test("captures blocked preview feedback and Timeline evidence badges", async ({
    page,
  }, testInfo) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALEVIDENCEBLOCKED"),
      "Evidence lifecycle visual evidence badges",
    );
    const blocked = await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALEVIDENCEBLOCKED-BLOCKED"),
        "evidence.title": "Blocked visual package",
        "evidence.storage_ref": "ticket://visual-blocked",
      },
    );
    const timelineRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALEVIDENCEBLOCKED-TIMELINE"),
        "timeline.activity_synopsis_text": "Visual evidence badge row",
      },
    );

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
      "evidence-grid-blocked-preview",
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
      "evidence-grid-timeline-evidence-badge",
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

test.describe("browser.evidence-workflow visual readiness", () => {
  test("Capture evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures.", async ({
    page,
  }, testInfo) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALEVIDENCEWORKFLOW"),
      "browser.evidence-workflow visual evidence affordance",
    );
    const requested = await createVisualEvidenceRow(page, incidentId, {
      lifecycleState: "requested",
      requestedAt: "2026-05-01T10:00:00Z",
      storageRef: "case://evidence-workflow/requested",
      title: "01 requested evidence",
      txnPrefix: "VISUALEVIDENCEWORKFLOW-REQUESTED",
    });
    const pending = await createVisualEvidenceRow(page, incidentId, {
      lifecycleState: "pending_receipt",
      requestedAt: "2026-05-01T10:05:00Z",
      storageRef: "case://evidence-workflow/pending",
      title: "02 pending evidence",
      txnPrefix: "VISUALEVIDENCEWORKFLOW-PENDING",
    });
    const blocked = await createVisualEvidenceRow(page, incidentId, {
      lifecycleState: "quarantined",
      requestedAt: "2026-05-01T10:10:00Z",
      storageRef: "case://evidence-workflow/quarantined",
      title: "03 quarantined evidence",
      txnPrefix: "VISUALEVIDENCEWORKFLOW-BLOCKED",
    });
    const availablePreview = await createUploadedVisualEvidence(
      page,
      incidentId,
      {
        body: Buffer.from(
          "visual.evidence-workflow preview visual evidence\n",
          "utf8",
        ),
        contentType: "text/plain",
        filename: "evidence-preview.txt",
        requestedAt: "2026-05-01T10:15:00Z",
        title: "04 available preview evidence",
        txnPrefix: "VISUALEVIDENCEWORKFLOW-PREVIEW",
      },
    );
    const downloadHandle = await createUploadedVisualEvidence(
      page,
      incidentId,
      {
        body: Buffer.from(
          "visual.evidence-workflow download handle visual evidence\n",
          "utf8",
        ),
        contentType: "text/plain",
        filename: "evidence-download-handle.txt",
        requestedAt: "2026-05-01T10:20:00Z",
        title: "05 download handle evidence",
        txnPrefix: "VISUALEVIDENCEWORKFLOW-DOWNLOAD",
      },
    );
    const previewBlocked = await createUploadedVisualEvidence(
      page,
      incidentId,
      {
        body: Buffer.from(
          "<!doctype html><title>visual.evidence-workflow unsupported preview</title>",
          "utf8",
        ),
        contentType: "text/html",
        filename: "evidence-preview-blocked.html",
        requestedAt: "2026-05-01T10:25:00Z",
        title: "06 preview blocked evidence",
        txnPrefix: "VISUALEVIDENCEWORKFLOW-PREVIEW-BLOCKED",
      },
    );
    const failedHandle = await createUploadedVisualEvidence(page, incidentId, {
      body: Buffer.from(
        "visual.evidence-workflow failed handle visual evidence\n",
        "utf8",
      ),
      contentType: "text/plain",
      filename: "evidence-failed-handle.txt",
      requestedAt: "2026-05-01T10:30:00Z",
      title: "07 failed handle evidence",
      txnPrefix: "VISUALEVIDENCEWORKFLOW-FAILED",
    });
    const inconsistentHandle = await createUploadedVisualEvidence(
      page,
      incidentId,
      {
        body: Buffer.from(
          "visual.evidence-workflow inconsistent handle visual evidence\n",
          "utf8",
        ),
        contentType: "text/plain",
        filename: "evidence-inconsistent-handle.txt",
        requestedAt: "2026-05-01T10:35:00Z",
        title: "08 inconsistent handle evidence",
        txnPrefix: "VISUALEVIDENCEWORKFLOW-INCONSISTENT",
      },
    );
    const timelineRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALEVIDENCEWORKFLOW-TIMELINE"),
        "timeline.activity_utc_text": "2026-05-01T11:00:00Z",
        "timeline.activity_synopsis_text":
          "browser.evidence-workflow timeline evidence count",
      },
    );

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
    expect(download.suggestedFilename()).toBe("evidence-download-handle.txt");
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
      "evidence-affordance-states",
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
        name: "timeline-evidence.png",
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
      "evidence-timeline-evidence-count",
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

test.describe("browser.collaboration workbook visual readiness", () => {
  test("Capture deterministic presence at the workbook header, row gutter, and cell with exact overflow counts.", async ({
    browser,
    page,
    sessionTracker,
  }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALCOLLABORATION"),
      "browser.collaboration visual collaboration states",
    );
    const remoteActors = await Promise.all(
      (
        [
          ["Alpha Analyst", "AA"],
          ["Bravo Analyst", "BA"],
          ["Charlie Analyst", "CA"],
          ["Delta Analyst", "DA"],
          ["Echo Analyst", "EA"],
          ["Foxtrot Analyst", "FA"],
        ] as const
      ).map(async ([displayName, actorText], index) => ({
        actorText,
        user: await createIncidentMemberUser(page, incidentId, {
          display_name: displayName,
          email: uniqueEmail(`collaboration-remote-${index + 1}`),
          initial_password: "FeVP7RemotePass!",
          role: "editor",
          is_deployment_admin: false,
          mfa_required: false,
        }),
      })),
    );
    const presenceRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOLLABORATION-PRESENCE"),
        "timeline.activity_synopsis_text": "Presence visual row",
      },
    );
    const primarySocket = installIncidentSocketMonitor(page, incidentId);

    const remotePages: Page[] = [];
    try {
      await page.goto(`/?incident_id=${incidentId}`);
      await primarySocket.waitForAcceptedSocket();
      await maskIncidentIdentity(page, incidentId);

      for (const [index, remoteActor] of remoteActors.entries()) {
        const remoteSession = await openIncidentAsTrackedUserReady(
          browser,
          sessionTracker,
          {
            createdBy: `visual.collaboration.presence-${index + 1}`,
            email: remoteActor.user.email,
            incidentId,
            password: remoteActor.user.initial_password,
            purpose: "browser.collaboration visual presence overflow analyst",
            readyRecordId: presenceRow.record_id,
            userId: remoteActor.user.user_id,
          },
        );
        remotePages.push(remoteSession.page);
        await focusRemoteTimelineCellAndWaitForPresence({
          fieldKey: "timeline.activity_synopsis_text",
          primaryPage: page,
          recordId: presenceRow.record_id,
          remotePage: remoteSession.page,
          socketMonitor: primarySocket,
          ...(index === 0 ? { actorText: remoteActor.actorText } : {}),
        });
      }
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
      await expect(
        page.getByTestId(workbookPresenceSummaryTestId()),
      ).toContainText("+1");
      await expect(
        page.getByTestId(rowPresenceMarkerTestId(presenceRow.record_id)),
      ).toContainText("+3");
      await expect(
        page.getByTestId(
          cellPresenceMarkerTestId(
            presenceRow.record_id,
            "timeline.activity_synopsis_text",
          ),
        ),
      ).toContainText("+4");
      await assertWorkbookGridVisualRegression(
        page,
        "collaboration-presence-markers",
        timelineViewSchemaId,
        { scroll: { top: 0, left: "left" } },
      );
    } finally {
      await Promise.all(
        remotePages.map((remotePage) => remotePage.context().close()),
      );
    }
  });

  test("Capture same-field conflict resolver.", async ({ page }) => {
    const fixture = await prepareFeP7ConflictVisual(page, {
      incidentKeyPrefix: "VISUALCOLLABORATIONRESOLVE",
      title: "browser.collaboration visual conflict resolver",
    });
    try {
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: gridRowTestId(
          timelineViewSchemaId,
          fixture.conflictRow.record_id,
        ),
      });
      await expect(
        page.getByTestId(
          timelineScalarEditorTestId({
            fieldKey: "timeline.activity_synopsis_text",
            recordId: fixture.conflictRow.record_id,
            surface: "grid",
          }),
        ),
      ).toHaveValue("Conflict visual local");
      await stabilizeConflictResolverVisual(page);
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "right" },
      });
      const conflictFixture = page.getByTestId(workbookShellReadyTestId());
      await conflictFixture.evaluate((element) => {
        element.setAttribute("data-design-fixture", "conflict");
      });
      await assertVisualRegression(
        page,
        "collaboration-conflict-resolver",
        conflictFixture,
      );
    } finally {
      await fixture.patchController.dispose();
    }
  });

  test("Capture conflict save-state strip.", async ({ page }) => {
    const fixture = await prepareFeP7ConflictVisual(page, {
      incidentKeyPrefix: "VISUALCOLLABORATIONCONFLICTSTRIP",
      title: "browser.collaboration visual conflict strip",
    });
    try {
      await page.getByTestId(workbookConflictControlTestId("close")).click();
      await expect(
        page.getByTestId(workbookConflictResolverTestId()),
      ).toHaveCount(0);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      await expect(
        page.getByTestId(workbookShellSlotTestId("status-strip")),
      ).toContainText("Conflict");
      await assertStatusStripVisualFixture(
        page,
        "collaboration-conflict-strip",
      );
    } finally {
      await fixture.patchController.dispose();
    }
  });

  test("Capture recovered saved-state strip.", async ({ page }) => {
    const fixture = await prepareFeP7ConflictVisual(page, {
      incidentKeyPrefix: "VISUALCOLLABORATIONRECOVERED",
      title: "browser.collaboration visual recovered strip",
    });
    try {
      await page
        .getByTestId(workbookConflictControlTestId("keep-saved"))
        .click();
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
      await assertStatusStripVisualFixture(
        page,
        "collaboration-recovered-saved-strip",
      );
    } finally {
      await fixture.patchController.dispose();
    }
  });
});

test.describe("browser.saved-view-query workbook visual readiness", () => {
  test("Capture saved-view selector, active chips, grouped result, group row, default/startup state indicator, and empty successful query fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const longSavedViewName =
      "browser.saved-view-query visual layout resilience with a deliberately long selected saved-view name";
    const longTagToken =
      "visual-unbroken-tag-0123456789-abcdefghijklmnopqrstuvwxyz";
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALSAVEDVIEW"),
      "browser.saved-view-query visual saved view query controls",
    );
    const reviewedRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALSAVEDVIEW-REVIEWED"),
        "timeline.activity_utc_text": "2026-06-08T12:00:00Z",
        "timeline.activity_synopsis_text":
          "browser.saved-view-query reviewed saved-view visual row",
        "timeline.tags": tagActionsPayload([longTagToken]),
      },
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALSAVEDVIEW-ROUGH"),
      "timeline.activity_utc_text": "2026-06-08T12:05:00Z",
      "timeline.activity_synopsis_text":
        "browser.saved-view-query rough grouped visual row",
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

    const sortTrigger = page.getByTestId(
      workbookSortMenuTriggerTestId(timelineViewSchemaId),
    );
    await sortTrigger.click();
    const sortMenu = page.getByTestId(
      workbookSortMenuTestId(timelineViewSchemaId),
    );
    await expect(sortMenu).toBeVisible();
    for (const fieldKey of [
      "timeline.activity_sort_ts",
      "timeline.date_entered_sort_day",
      "timeline.activity_synopsis_text",
      "timeline.analyst_text",
      "timeline.mitre_stage_text",
      "timeline.device_object_text",
      "timeline.ip_address_text",
      "timeline.capture_state",
    ]) {
      const option = page.getByTestId(
        workbookSortOptionTestId(timelineViewSchemaId, fieldKey),
      );
      await option.click();
      await expect(option).toHaveAttribute("aria-checked", "true");
    }
    await sortMenu.press("Escape");
    await expect(sortMenu).toHaveCount(0);
    await expect(sortTrigger).toBeFocused();
    await expect(
      page.locator('[data-grid-data-state="refreshing"]'),
    ).toHaveCount(0);
    await expect(
      page.locator(
        '[data-grid-data-state="stale_error"], [data-grid-data-state="unavailable"]',
      ),
    ).toHaveCount(0);

    await setSavedViewDraftName(page, timelineViewSchemaId, longSavedViewName);
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

    await applyFilterChip(
      page,
      timelineViewSchemaId,
      "timeline.tags",
      longTagToken,
    );
    await expect(
      page.getByTestId(savedViewModifiedTestId(timelineViewSchemaId)),
    ).toHaveText("Modified");
    await expect(
      page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
    ).toHaveAttribute("title", longSavedViewName);
    const queryControls = page.getByTestId(
      workbookViewBarQueryControlsTestId(timelineViewSchemaId),
    );
    await expect(queryControls).toHaveAttribute(
      "data-query-chip-capacity",
      "8",
    );
    await expect(queryControls).toHaveAttribute(
      "data-hidden-query-chip-count",
      "3",
    );
    await expect(
      queryControls
        .getByRole("toolbar", { name: "Active query chips" })
        .locator("button[title]"),
    ).toHaveCount(8);
    await expect(
      page.getByTestId(
        workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      ),
    ).toHaveAttribute("aria-label", "Filters, 3 hidden");
    await expect(
      page.getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(
      page.getByTestId(workbookAddRowButtonTestId(timelineViewSchemaId)),
    ).toBeVisible();

    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    await assertViewportVisualRegression(
      page,
      "workbook-query-saved-view-query-controls",
    );

    const emptyIncidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALSAVEDVIEWEMPTY"),
      "browser.saved-view-query empty successful Timeline query",
    );
    await page.setViewportSize({ width: 1280, height: 720 });
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
    await page
      .getByTestId(gridShellTestId(timelineViewSchemaId))
      .evaluate((element) => {
        element.setAttribute("data-design-fixture", "empty-state");
      });
    await assertWorkbookGridVisualRegression(
      page,
      "workbook-query-empty-successful-query",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("browser.inspector-history workbook visual readiness", () => {
  test("Capture inspector Details, Relationships, Evidence, History, rollback preview, destructive confirmation, and public error fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALINSPECTORHISTORY"),
      "browser.inspector-history visual inspector actions",
    );
    const evidence = await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALINSPECTORHISTORY-EVIDENCE"),
        "evidence.collector_party_text":
          "browser.inspector-history visual collector",
        "evidence.title": "browser.inspector-history visual attached evidence",
      },
    );
    const target = await createViewRow(page, incidentId, timelineViewSchemaId, {
      [hostRefsFieldKey]: collectionActionsPayload([
        "browser.inspector-history visual host",
      ]),
      client_txn_id: uniqueTxn("VISUALINSPECTORHISTORY-TARGET"),
      "timeline.raw_activity_text":
        "browser.inspector-history visual inspector details",
      "timeline.activity_synopsis_text":
        "browser.inspector-history visual inspector target",
    });
    const linkedTarget = await patchRecord(page, target.record_id, {
      base_row_version: target.row_version,
      changes: [
        {
          action_payload: feP9VisualAttachedEvidencePayload(evidence.record_id),
          field_key: "timeline.attached_evidence_ids",
        },
      ],
      client_txn_id: uniqueTxn("VISUALINSPECTORHISTORY-LINK"),
      view_schema_id: timelineViewSchemaId,
    });
    const hostItem = requireItemByRawText(
      collectionItems(linkedTarget, hostRefsFieldKey),
      "browser.inspector-history visual host",
    );
    const history = await fetchRecordHistory(page, target.record_id);
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
    ).toHaveValue("browser.inspector-history visual inspector details");
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
      "workbook-inspector-relationships",
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
    await assertViewportVisualRegression(page, "workbook-inspector-history");

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
    await assertViewportVisualRegression(
      page,
      "workbook-inspector-rollback-preview",
    );

    await page
      .getByTestId(rowHistoryRollbackCancelButtonTestId(rollbackAnchor))
      .click();
    await page.getByTestId(rowHistoryDeleteButtonTestId()).click();
    await expect(
      page.getByTestId(
        rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
      ),
    ).toContainText(target.record_id);
    const destructiveActions = page.getByTestId(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
    );
    await destructiveActions.evaluate((element) => {
      element.setAttribute("data-design-fixture", "destructive-actions");
    });
    await destructiveActions.scrollIntoViewIfNeeded();
    await assertVisualRegression(
      page,
      "workbook-inspector-destructive-confirmation",
      destructiveActions,
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
              status: 409,
              code: "row_version_conflict",
              message: "row_version_conflict",
              request_id: "visual-inspector-history-conflict",
              retryable: false,
              details: {
                reason_code: "stale_row_version",
              },
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
    await assertViewportVisualRegression(
      page,
      "workbook-inspector-public-error",
    );
  });
});

function feP9VisualAttachedEvidencePayload(
  recordId: string,
): CollectionActionsV1 {
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
  item: RecordHistoryItem,
  action: RecordHistoryItem["available_rollback_actions"][number],
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

function feP9VisualRollbackPreviewAnchor(
  item: RecordHistoryItem,
  action: RecordHistoryItem["available_rollback_actions"][number],
) {
  return {
    action,
    historyItemRef: item.history_item_ref,
  };
}

function requireFeP9VisualHistoryEntryAction(history: RecordHistoryData) {
  const item =
    history.items.find(
      (candidate) =>
        candidate.available_rollback_actions.includes("history_entry") &&
        typeof candidate.history_entry_ref === "string" &&
        candidate.history_entry_ref.length > 0,
    ) ?? null;
  if (item === null) {
    throw new Error(
      "missing browser.inspector-history visual history_entry rollback item",
    );
  }
  return item;
}

async function prepareFeP7ConflictVisual(
  page: Page,
  options: {
    incidentKeyPrefix: string;
    title: string;
  },
) {
  await page.setViewportSize({ width: 1280, height: 720 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(options.incidentKeyPrefix),
    options.title,
  );
  const conflictRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.incidentKeyPrefix}-CONFLICT`),
      "timeline.activity_utc_text": "2025-03-07T10:00:00Z",
      "timeline.activity_synopsis_text": "Conflict visual base",
    },
  );
  const patchController = await installPatchController(page);

  await page.goto(`/?incident_id=${incidentId}`);
  await maskIncidentIdentity(page, incidentId);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await driveRealTimelineSummaryConflict({
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

test.describe("workbook visual evidence", () => {
  test("regresses collaboration row-gutter and same-cell presence markers", async ({
    browser,
    page,
    sessionTracker,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALCOLLABORATIONPRESENCE"),
      "Collaboration visual presence markers",
    );
    const remote = await createIncidentMemberUser(page, incidentId, {
      display_name: "Visual Analyst",
      email: uniqueEmail("collaboration-v6grid01-remote"),
      initial_password: "CollaborationV6Grid01!",
      role: "editor",
      is_deployment_admin: false,
      mfa_required: false,
    });
    const timelineRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOLLABORATIONPRESENCE-ROW"),
        "timeline.activity_synopsis_text": "Presence visual row",
      },
    );
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
      ).toHaveText("Presence visual row");

      const remoteSession = await openIncidentAsTrackedUserReady(
        browser,
        sessionTracker,
        {
          createdBy: "collaboration-visual",
          email: remote.email,
          incidentId,
          password: remote.initial_password,
          purpose: "Collaboration visual presence analyst",
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
        "collaboration-grid-presence-markers",
        timelineViewSchemaId,
        { scroll: { top: 0, left: "left" } },
      );
    } finally {
      await remotePage?.context().close();
    }
  });

  test("regresses collaboration same-field conflict marker resolver and Conflict strip", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALCOLLABORATIONCONFLICT"),
      "Collaboration visual conflict resolver",
    );
    const timelineRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOLLABORATIONCONFLICT-ROW"),
        "timeline.activity_synopsis_text": "Conflict visual base",
      },
    );

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    const patchController = await installPatchController(page);
    try {
      const localValue = `Conflict_visual_local_${"L".repeat(96)}`;
      const remoteValue = `Conflict_visual_saved_${"S".repeat(96)}`;
      await driveRealTimelineSummaryConflict({
        baseRowVersion: timelineRow.row_version,
        localValue,
        page,
        patchController,
        recordId: timelineRow.record_id,
        remoteValue,
        txnPrefix: "visual-collaboration-conflict",
      });
      await stabilizeConflictResolverVisual(page);
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "right" },
      });

      await assertViewportVisualRegression(
        page,
        "collaboration-grid-conflict-resolver",
      );

      await page.setViewportSize({ width: 1024, height: 720 });
      await stabilizeConflictResolverVisual(page);
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "right" },
      });
      await assertViewportVisualRegression(
        page,
        "collaboration-grid-conflict-resolver-narrow",
      );

      await page.setViewportSize({ width: 768, height: 640 });
      await stabilizeConflictResolverVisual(page);
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "right" },
      });
      await assertViewportVisualRegression(
        page,
        "collaboration-grid-conflict-resolver-compact",
      );
    } finally {
      await patchController.dispose();
    }
  });

  test("regresses collaboration pending-queue save-state transitions", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALCOLLABORATIONSAVE"),
      "Collaboration visual pending queue",
    );
    const syncRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOLLABORATIONSAVE-ROW"),
        "timeline.activity_utc_text": "2025-03-06T10:00:00Z",
        "timeline.activity_synopsis_text": "Pending visual base",
      },
    );
    const conflictRow = await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOLLABORATIONSAVE-CONFLICT-ROW"),
        "timeline.activity_utc_text": "2025-03-06T10:05:00Z",
        "timeline.activity_synopsis_text": "Pending conflict visual base",
      },
    );
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
      await summaryInput.click();
      const summaryEditor = page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: syncRow.record_id,
          surface: "grid",
        }),
      );
      await expect(summaryEditor).toBeFocused();
      await summaryEditor.fill("Pending visual syncing");
      await summaryEditor.press("Enter");
      await hold.waitForHit;
      await expect(saveState).toHaveText("Syncing");
      await assertStatusStripVisualRegression(
        page,
        "collaboration-grid-syncing-strip",
      );
      await hold.release();
      await hold.waitForCompletion;
      await expect(saveState).toHaveText("Saved");
      await assertStatusStripVisualRegression(
        page,
        "collaboration-grid-saved-strip",
      );

      await driveRealTimelineSummaryConflict({
        baseRowVersion: conflictRow.row_version,
        localValue: "Pending visual blocked",
        page,
        patchController,
        recordId: conflictRow.record_id,
        remoteValue: "Pending visual server",
        txnPrefix: "visual-collaboration-pending-conflict",
      });
      await page
        .getByTestId(workbookConflictResolverTestId())
        .scrollIntoViewIfNeeded();
      await stabilizeConflictResolverVisual(page);
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "right" },
      });
      await assertViewportVisualRegression(
        page,
        "collaboration-grid-blocked-conflict",
      );

      await page
        .getByTestId(workbookConflictControlTestId("keep-saved"))
        .click();
      await expect(saveState).toHaveText("Saved");
    } finally {
      await patchController.dispose();
    }

    await expect(saveState).toHaveText("Saved");
    await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    await assertStatusStripVisualRegression(
      page,
      "collaboration-grid-recovered-saved-strip",
    );
  });
});

test.describe("browser.coordination-review workbook visual readiness", () => {
  test("Capture Task Requests or Decisions, Parties link state, Communications Log, Handoff, Status Review, Lesson, and keyboard focus fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.evaluate(() => {
      document.documentElement.style.zoom = "100%";
    });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALCOORDINATIONREVIEW"),
      "browser.coordination-review coordination visual readiness",
    );
    const owner = await createIncidentMemberUser(page, incidentId, {
      display_name: "browser.coordination-review visual owner",
      email: uniqueEmail("coordination-review-visual-owner"),
      initial_password: "BackupRestoreVisual1!",
      role: "editor",
      is_deployment_admin: false,
      mfa_required: false,
    });
    const party = await createViewRow(page, incidentId, partiesViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-PARTY"),
      "party.display_name": "browser.coordination-review Visual Party",
      "party.party_kind": "team",
    });
    const taskRow = await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOORDINATION-TASK"),
        "task.task_kind": "collection",
        "task.title": "Visual task request",
      },
    );
    const decision = await createViewRow(
      page,
      incidentId,
      decisionsViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-DECISION"),
        "decision.decision_type": "containment",
        "decision.rationale": "browser.coordination-review visual rationale",
        "decision.summary": "browser.coordination-review visual decision",
      },
    );
    const comm = await createViewRow(page, incidentId, commLogViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-COMM"),
      "comm_log.audience": "browser.coordination-review visual responders",
      "comm_log.channel_or_meeting":
        "browser.coordination-review visual bridge",
      "comm_log.comm_type": "briefing",
      "comm_log.decision_ids": {
        actions: [
          { linked_record_id: decision.record_id, op: "add_record_ref" },
        ],
        kind: "collection_actions_v1",
      },
      "comm_log.summary": "browser.coordination-review visual communication",
    });
    const handoff = await createViewRow(page, incidentId, handoffViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-HANDOFF"),
      "handoff.current_state_summary":
        "browser.coordination-review visual handoff state",
      "handoff.incoming_owner_user_id": owner.user_id,
    });
    const status = await createViewRow(
      page,
      incidentId,
      statusReviewViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-STATUS"),
        "status_review.current_state_summary":
          "browser.coordination-review visual status review state",
      },
    );
    const lesson = await createViewRow(page, incidentId, lessonViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-LESSON"),
      "lesson.summary": "browser.coordination-review visual lesson",
    });

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
      "record-relationships-task-requests",
      taskRequestsViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    const linkedTask = await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("VISUALCOORDINATIONREVIEW-LINKED-TASK"),
        "task.requester_party_id": party.record_id,
        "task.requester_party_text": "browser.coordination-review requester",
        "task.task_kind": "follow_up",
        "task.title": "browser.coordination-review party-linked task",
      },
    );

    const surfaceExpectations = [
      {
        expected: "browser.coordination-review party-linked task",
        fieldKey: "task.title",
        recordId: linkedTask.record_id,
        surface: taskRequestsViewSchemaId,
      },
      {
        expected: "browser.coordination-review Visual Party",
        fieldKey: "party.display_name",
        recordId: party.record_id,
        surface: partiesViewSchemaId,
      },
      {
        expected: "browser.coordination-review visual communication",
        fieldKey: "comm_log.summary",
        recordId: comm.record_id,
        surface: commLogViewSchemaId,
      },
      {
        expected: "browser.coordination-review visual handoff state",
        fieldKey: "handoff.current_state_summary",
        recordId: handoff.record_id,
        surface: handoffViewSchemaId,
      },
      {
        expected: "browser.coordination-review visual status review state",
        fieldKey: "status_review.current_state_summary",
        recordId: status.record_id,
        surface: statusReviewViewSchemaId,
      },
      {
        expected: "browser.coordination-review visual lesson",
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
    await test.info().attach("coordination-visual-fixture-matrix.json", {
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
              "visual.fixture.edit_cell readonly editor fixture captured in the grid-adapter fixture.",
            fixture_ids: [
              "visual.fixture.task_requests_or_decisions",
              "visual.fixture.frozen_column",
              "visual.fixture.resize_handle",
              "visual.fixture.drag_fill_handle",
              "visual.fixture.edit_cell",
              "visual.fixture.tree_group_row",
            ],
            focus_state:
              "Each browser.coordination-review coordination/review surface focuses its row cell after deterministic query state is visible.",
            screenshot_scopes: [
              "task requests workbook grid",
              "grid-adapter design fixture",
            ],
            seed_id: "coordination-visual-deterministic-seed",
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

test.describe("browser.design-readiness visual readiness", () => {
  test("Run the owned-stack Playwright visual suite with deterministic seed data, viewport, zoom, fixture ordering, dynamic masks, scroll anchors, focus/editor state, inspector state, and post-scroll settle behavior.", async ({
    browserName: _browserName,
  }, testInfo) => {
    const registry = loadFrontendVisualFixtureRegistry();
    expect(registry.schema_id).toBe(
      "cartulary.frontend_visual_fixture_registry.v5",
    );
    expect(registry.owner_id).toBe("harness.visual");
    expect(registry.verification_id).toBe(
      "harness.visual.verification.stable_fixture_identity",
    );
    expect(registry.fixtures.map((fixture) => fixture.fixture_id)).toEqual(
      expectedFrontendVisualFixtureIds,
    );
    for (const fixture of registry.fixtures) {
      expect(fixture.status).toBe("current");
      expectCurrentFrontendVisualFixtureMetadata(fixture);
    }
    expect(
      registry.fixtures
        .map((fixture) => fixture.design_contract_id)
        .filter((designContractId): designContractId is string =>
          Boolean(designContractId),
        )
        .sort(),
    ).toEqual(expectedDesignContractIds);

    await testInfo.attach("owned-stack-visual-suite.json", {
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

  test("Ensure the visual fixture matrix maps every design contract exactly once and retains implementation-only workbook, collaboration, evidence, entity, coordination, and grid-adapter fixtures.", async ({
    browserName: _browserName,
  }, testInfo) => {
    const registry = loadFrontendVisualFixtureRegistry();
    const fixturesById = new Map(
      registry.fixtures.map((fixture) => [fixture.fixture_id, fixture]),
    );
    expect([...fixturesById.keys()]).toEqual(expectedFrontendVisualFixtureIds);

    const defaultShell = fixturesById.get(
      "visual.fixture.default_timeline_workbook_shell",
    );
    expect(defaultShell?.fixture_title).toBe("Default Timeline workbook shell");
    expect(defaultShell?.capture_scope.kind).toBe("full_viewport");
    expect(defaultShell?.catalog_row_ids).not.toContain(
      "web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c",
    );

    const componentStates = fixturesById.get(
      "visual.fixture.component_state_matrix",
    );
    expect(componentStates?.fixture_title).toBe("Component state matrix");
    expect(componentStates?.design_contract_id).toBe("D-VFIX-009");
    expect(componentStates?.catalog_row_ids).toEqual([
      "web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c",
    ]);
    expect(componentStates?.capture_scope).toEqual({
      kind: "selector",
      selector: "[data-design-fixture='components']",
    });
    expect(componentStates?.scroll_normalization.kind).toBe("not_applicable");
    expect(componentStates?.viewport_css_px).toBe("1280x720");
    expect(componentStates?.dynamic_masks).toEqual([]);
    expect(componentStates?.no_dynamic_regions).toBe(true);

    const designMappings = registry.fixtures
      .filter(
        (
          fixture,
        ): fixture is FrontendVisualFixture & { design_contract_id: string } =>
          fixture.design_contract_id !== undefined,
      )
      .map((fixture) => ({
        capture_scope: fixture.capture_scope,
        design_contract_id: fixture.design_contract_id,
        fixture_id: fixture.fixture_id,
        viewport_css_px: fixture.viewport_css_px,
      }))
      .sort((left, right) =>
        left.design_contract_id.localeCompare(right.design_contract_id),
      );
    expect(designMappings.map((mapping) => mapping.design_contract_id)).toEqual(
      expectedDesignContractIds,
    );

    await testInfo.attach("visual-fixture-matrix.json", {
      body: Buffer.from(
        JSON.stringify(
          {
            fixture_ids: expectedFrontendVisualFixtureIds,
            matrix_titles: expectedFrontendVisualFixtureIds.map(
              (fixtureId) => fixturesById.get(fixtureId)?.fixture_title,
            ),
            non_claim_boundaries: [
              "component-state evidence does not satisfy shell geometry fixtures",
              "current fixture metadata does not close browser.design-readiness without row accounting",
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

  test("Capture exposed dark_graphite token and theme states with deterministic density, color, component, focus, and semantic-state samples.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("VISUALDESIGNREADINESS"),
      "browser.design-readiness exposed theme visual fixture",
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("VISUALDESIGNREADINESS-ROW"),
      "timeline.activity_utc_text": "2026-05-31T11:00:00Z",
      "timeline.activity_synopsis_text":
        "browser.design-readiness exposed theme fixture row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.locator("main.cartulary-shell").first()).toHaveAttribute(
      "data-cartulary-theme",
      cartularyDefaultThemeId,
    );
    await assertExposedThemeCssVariables(page);
    await injectExposedThemeVisualFixture(page);

    const fixture = page.locator("[data-design-fixture='components']");
    await expect(fixture).toBeVisible();
    await assertVisualRegression(page, "design-exposed-theme-states", fixture);

    await injectDelayedLoadingVisualFixture(page);
    const delayedLoading = page.locator(
      "[data-design-fixture='delayed-loading']",
    );
    await expect(delayedLoading).toContainText("Still loading this surface");
    await expect(delayedLoading.getByRole("button")).toHaveCount(0);
    await assertVisualRegression(
      page,
      "design-delayed-initial-loading",
      delayedLoading,
    );

    await injectErrorPresentationVisualFixture(page);
    const errorPresentation = page.locator(
      "[data-design-fixture='error-presentation']",
    );
    await expect(errorPresentation).toBeVisible();
    await expect(errorPresentation.locator("[data-error-locus]")).toHaveCount(
      5,
    );
    await assertVisualRegression(
      page,
      "design-error-presentation-loci",
      errorPresentation,
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
    lifecycleState: NonNullable<
      EvidenceCreateRequest["evidence.lifecycle_state"]
    >;
    requestedAt: string;
    storageRef: string;
    title: string;
    txnPrefix: string;
  },
): Promise<ViewRow> {
  return createEvidenceFixtureRow(page, incidentId, {
    collectorPartyText: "browser.evidence-workflow visual fixture",
    ...options,
  });
}

async function createUploadedVisualEvidence(
  page: Page,
  incidentId: string,
  options: EvidenceUploadOptions,
): Promise<ViewRow> {
  return createUploadedEvidenceFixture(page, incidentId, {
    collectorPartyText: "browser.evidence-workflow visual fixture",
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
  const response = await createEnvironmentTestControlClient(page.request, {
    endpointOrigin: apiBase,
  }).request({
    body: {
      code: "evidence_access_unavailable",
      consume_once: true,
      details: {
        reason_code: options.reasonCode,
      },
      message:
        "Evidence access failed for browser.evidence-workflow visual fixture.",
      method: "POST",
      path: options.path,
      retryable: false,
      status: 409,
    },
    method: "POST",
    path: "/api/v1/test/runtime/public-error-faults",
  });
  expect(response.status).toBe(201);
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
  await emitVisualCaptureIntent(
    page,
    name,
    "apps/web/e2e/workbook.visual.spec.ts#assertVisualRegression",
  );
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
  await emitVisualCaptureIntent(
    page,
    name,
    "apps/web/e2e/workbook.visual.spec.ts#assertViewportVisualRegression",
  );
  await expect(page).toHaveScreenshot(`${name}.png`, {
    animations: "disabled",
    caret: "hide",
    fullPage: false,
  });
}

async function emitVisualCaptureIntent(
  page: Page,
  captureIntent: string,
  screenshotAssertionLocation: string,
) {
  const testInfo = test.info();
  const viewport = page.viewportSize();
  if (viewport === null) {
    throw new Error(`visual capture ${captureIntent} requires a viewport`);
  }
  const browserProfile = await page.evaluate(() => {
    const zoom = Number.parseFloat(document.documentElement.style.zoom);
    return {
      browser_zoom_percent: Number.isFinite(zoom) ? Math.round(zoom) : 100,
      color_scheme: window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light",
      density_id:
        document.documentElement.getAttribute("data-cartulary-density") ?? "",
      device_scale_factor: window.devicePixelRatio,
      reduced_motion: window.matchMedia("(prefers-reduced-motion: reduce)")
        .matches,
      theme_id:
        document.documentElement.getAttribute("data-cartulary-theme") ?? "",
    };
  });
  const expectedGoldenPath = path
    .relative(findRepoRoot(), testInfo.snapshotPath(`${captureIntent}.png`))
    .replaceAll(path.sep, "/");
  const captureId = `visual.capture.${createHash("sha256")
    .update(
      JSON.stringify([
        testInfo.project.name,
        testInfo.title,
        expectedGoldenPath,
      ]),
    )
    .digest("hex")
    .slice(0, 20)}`;
  await testInfo.attach(`cartulary-visual-capture-intent-${captureId}.json`, {
    body: Buffer.from(
      `${JSON.stringify({
        schema_id: "cartulary.frontend_visual_capture_intent.v1",
        capture_id: captureId,
        capture_intent: captureIntent,
        expected_golden_path: expectedGoldenPath,
        project_id: testInfo.project.name,
        screenshot_assertion_location: screenshotAssertionLocation,
        capture_profile: {
          ...browserProfile,
          project_id: testInfo.project.name,
          snapshot_path_template:
            "{snapshotDir}/{testFileDir}/{testFileName}-snapshots/{arg}{-snapshotSuffix}{ext}",
          snapshot_suffix: testInfo.snapshotSuffix,
          viewport_css_px: `${viewport.width}x${viewport.height}`,
        },
        test_file: path
          .relative(findRepoRoot(), testInfo.file)
          .replaceAll(path.sep, "/"),
        test_title: testInfo.title,
      })}\n`,
      "utf8",
    ),
    contentType: "application/json",
  });
}

async function scrollVisualAnchorToScrollContainerTop(
  page: Page,
  locator: Locator,
  { clipTopPixels = 0 }: { clipTopPixels?: number } = {},
) {
  await locator.evaluate(
    (element, visualAnchorOptions) => {
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
        window.scrollBy({
          top: elementRect.top + visualAnchorOptions.clipTopPixels,
          left: 0,
          behavior: "instant",
        });
        return;
      }

      const containerRect = container.getBoundingClientRect();
      container.scrollTop +=
        elementRect.top - containerRect.top + visualAnchorOptions.clipTopPixels;
    },
    { clipTopPixels },
  );
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
    .getByTestId(authTestId("login-username"))
    .fill("visual.operator@example.test");
  await page.getByTestId(authTestId("login-password")).fill("VisualPass1!");
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
      clone.setAttribute("data-design-fixture", "status-strip");
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
  const focusAnchor = statusStrip.getByTestId(workbookFocusAnchorTestId());
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
    ariaLabel: "browser.grid-interaction grid adapter visual fixtures",
    fixtureName: "grid-interaction-grid-adapter",
    missingMainMessage:
      "Expected workbook shell main before browser.grid-interaction fixture",
    styleText: `
      [data-design-fixture='grid-interaction-grid-adapter'] {
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

      [data-design-fixture='grid-interaction-grid-adapter'] * {
        box-sizing: border-box;
      }

      .grid-interaction-grid-fixture-table {
        min-width: 0;
        overflow: hidden;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        background: var(--ct-colors-surface-1);
      }

      .grid-interaction-grid-fixture-row {
        display: grid;
        grid-template-columns: 10rem 16rem 12rem 14rem 14rem;
        min-width: 66rem;
      }

      .grid-interaction-grid-fixture-head,
      .grid-interaction-grid-fixture-cell {
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

      .grid-interaction-grid-fixture-head {
        min-height: 3rem;
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
        font-weight: var(--ct-typography-metadata-fontWeight);
      }

      .grid-interaction-grid-fixture-frozen {
        position: sticky;
        left: 0;
        z-index: 2;
        background: var(--ct-colors-surface-2);
        box-shadow: 0.75rem 0 1rem rgba(0, 0, 0, 0.28);
      }

      .grid-interaction-grid-fixture-resize-handle {
        position: absolute;
        inset-block: 0.45rem;
        inset-inline-end: 0.2rem;
        width: 0.25rem;
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-hairline-focus);
      }

      .grid-interaction-grid-fixture-active {
        outline: var(--ct-component-focus-ring-border);
        outline-offset: -0.2rem;
        background: var(--ct-colors-surface-3);
      }

      .grid-interaction-grid-fixture-editor {
        width: 100%;
        min-width: 0;
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-sm);
        padding: 0.45rem 0.55rem;
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
        font: inherit;
      }

      .grid-interaction-grid-fixture-fill {
        position: absolute;
        right: 0.2rem;
        bottom: 0.2rem;
        width: 0.65rem;
        height: 0.65rem;
        border: 0.15rem solid var(--ct-colors-hairline-focus);
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-surface-1);
      }

      .grid-interaction-grid-fixture-group {
        grid-column: 1 / -1;
        min-height: 3.5rem;
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink-muted);
        font-weight: 600;
      }

      .grid-interaction-grid-fixture-tree-toggle {
        display: inline-grid;
        place-items: center;
        width: 1.35rem;
        height: 1.35rem;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
      }

      .grid-interaction-grid-fixture-side {
        display: grid;
        grid-template-rows: auto 1fr;
        gap: var(--ct-spacing-sm);
        min-width: 0;
      }

      .grid-interaction-grid-fixture-caption {
        margin: 0;
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
      }

      .grid-interaction-grid-fixture-empty {
        display: grid;
        place-items: center;
        min-height: 16rem;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink-muted);
        text-align: center;
      }

      .grid-interaction-grid-fixture-empty strong {
        display: block;
        margin-block-end: var(--ct-spacing-xs);
        color: var(--ct-colors-ink);
      }
    `,
    html: `
      <div class="grid-interaction-grid-fixture-table" role="grid" aria-label="Adapter fixture grid">
        <div class="grid-interaction-grid-fixture-row" role="row">
          <div class="grid-interaction-grid-fixture-head grid-interaction-grid-fixture-frozen" role="columnheader" data-fixture-id="visual.fixture.frozen_column">Record</div>
          <div class="grid-interaction-grid-fixture-head" role="columnheader" data-fixture-id="visual.fixture.resize_handle">Summary<span class="grid-interaction-grid-fixture-resize-handle" aria-hidden="true"></span></div>
          <div class="grid-interaction-grid-fixture-head" role="columnheader">State</div>
          <div class="grid-interaction-grid-fixture-head" role="columnheader">Assignee</div>
          <div class="grid-interaction-grid-fixture-head" role="columnheader">Last edit</div>
        </div>
        <div class="grid-interaction-grid-fixture-row" role="row">
          <div class="grid-interaction-grid-fixture-cell grid-interaction-grid-fixture-group" role="rowheader" data-fixture-id="visual.fixture.tree_group_row"><span class="grid-interaction-grid-fixture-tree-toggle" aria-hidden="true">v</span> reviewed group, 2 rows</div>
        </div>
        <div class="grid-interaction-grid-fixture-row" role="row">
          <div class="grid-interaction-grid-fixture-cell grid-interaction-grid-fixture-frozen" role="rowheader">record-1</div>
          <div class="grid-interaction-grid-fixture-cell grid-interaction-grid-fixture-active" role="gridcell" data-fixture-id="visual.fixture.edit_cell"><input class="grid-interaction-grid-fixture-editor" value="Edit cell adapter" aria-label="Summary editor" readonly><span class="grid-interaction-grid-fixture-fill" data-fixture-id="visual.fixture.drag_fill_handle" aria-hidden="true"></span></div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">reviewed</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">Analyst</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">saved</div>
        </div>
        <div class="grid-interaction-grid-fixture-row" role="row">
          <div class="grid-interaction-grid-fixture-cell grid-interaction-grid-fixture-frozen" role="rowheader">record-2</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">Frozen column remains pinned</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">rough</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">Unassigned</div>
          <div class="grid-interaction-grid-fixture-cell" role="gridcell">clean</div>
        </div>
      </div>
      <aside class="grid-interaction-grid-fixture-side" aria-label="Empty successful query fixture">
        <p class="grid-interaction-grid-fixture-caption">Adapter-owned visual states only. Row-gutter presence remains browser.collaboration and grouped-result query ownership remains browser.saved-view-query.</p>
        <div class="grid-interaction-grid-fixture-empty" data-fixture-id="visual.fixture.empty_successful_query">
          <span><strong>No rows match this query</strong>Successful empty result</span>
        </div>
      </aside>
    `,
  });
}

async function injectExposedThemeVisualFixture(page: Page) {
  await injectDesignFixture(page, {
    ariaLabel: "Exposed theme token state fixture",
    fixtureName: "components",
    missingMainMessage: "Expected workbook shell main before theme fixture",
    styleText: `
      [data-design-fixture='components'] {
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

      [data-design-fixture='components'] * {
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

async function injectDelayedLoadingVisualFixture(page: Page) {
  await injectDesignFixture(page, {
    ariaLabel: "Delayed initial-loading state",
    fixtureName: "delayed-loading",
    missingMainMessage: "Expected workbook shell main before loading fixture",
    styleText: `
      [data-design-fixture='delayed-loading'] {
        position: fixed;
        inset: var(--ct-spacing-xl);
        box-sizing: border-box;
        display: grid;
        place-items: center;
        background: var(--ct-colors-canvas);
        color: var(--ct-colors-ink);
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-lg);
        box-shadow: var(--ct-elevation-panel);
        font-family: var(--ct-typography-ui-fontFamily);
        z-index: 1001;
      }

      .delayed-loading-card {
        display: grid;
        justify-items: center;
        gap: var(--ct-spacing-sm);
        min-inline-size: 24rem;
        padding: var(--ct-spacing-xl);
        background: var(--ct-colors-surface-1);
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
      }

      .delayed-loading-spinner {
        inline-size: var(--ct-component-icon-inline-size);
        block-size: var(--ct-component-icon-inline-size);
        border: var(--ct-border-strong);
        border-inline-start-color: var(--ct-colors-accent);
        border-radius: var(--ct-rounded-pill);
      }

      .delayed-loading-card strong {
        font-family: var(--ct-typography-surface-title-fontFamily);
        font-size: var(--ct-typography-surface-title-fontSize);
      }

      .delayed-loading-card span {
        color: var(--ct-colors-ink-muted);
      }
    `,
    html: `
      <div class="delayed-loading-card" role="status" aria-live="polite">
        <span class="delayed-loading-spinner" aria-hidden="true"></span>
        <strong>Still loading this surface</strong>
        <span>Timeline remains busy for this request generation.</span>
      </div>
    `,
  });
}

async function injectErrorPresentationVisualFixture(page: Page) {
  await injectDesignFixture(page, {
    ariaLabel: "Representative error-presentation loci",
    fixtureName: "error-presentation",
    missingMainMessage: "Expected workbook shell main before error fixture",
    styleText: `
      [data-design-fixture='error-presentation'] {
        position: fixed;
        inset: var(--ct-spacing-xl);
        box-sizing: border-box;
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: var(--ct-spacing-md);
        overflow: hidden;
        background: var(--ct-colors-canvas);
        color: var(--ct-colors-ink);
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-lg);
        padding: var(--ct-spacing-lg);
        box-shadow: var(--ct-elevation-panel);
        font-family: var(--ct-typography-ui-fontFamily);
        z-index: 1002;
      }

      .error-locus-title {
        grid-column: 1 / -1;
        margin: 0;
        font-family: var(--ct-typography-surface-title-fontFamily);
        font-size: var(--ct-typography-surface-title-fontSize);
      }

      [data-error-locus] {
        display: grid;
        align-content: start;
        gap: var(--ct-spacing-xs);
        min-width: 0;
        padding: var(--ct-spacing-md);
        background: var(--ct-colors-surface-1);
        border: var(--ct-border-hairline);
        border-inline-start: var(--ct-border-strong);
        border-inline-start-color: var(--ct-colors-semantic-conflict);
        border-radius: var(--ct-rounded-md);
      }

      [data-error-locus='permission-loss'] {
        grid-column: 1 / -1;
        border-inline-start-color: var(--ct-colors-semantic-destructive);
      }

      [data-error-locus] strong,
      [data-error-locus] p {
        margin: 0;
      }

      [data-error-locus] p {
        color: var(--ct-colors-ink-muted);
      }

      .error-locus-actions {
        display: flex;
        gap: var(--ct-spacing-xs);
      }

      .error-locus-actions button {
        border: var(--ct-component-button-secondary-border);
        border-radius: var(--ct-component-button-secondary-rounded);
        background: var(--ct-component-button-secondary-backgroundColor);
        color: var(--ct-component-button-secondary-textColor);
        padding: var(--ct-component-button-secondary-padding);
      }
    `,
    html: `
      <h2 class="error-locus-title">Typed error families at their recovery loci</h2>
      <section data-error-locus="local-validation" role="alert">
        <strong>Cell validation</strong>
        <p>Committed value retained · local draft “09:7x” retained</p>
        <div class="error-locus-actions"><button type="button">Correct value</button><button type="button">Cancel draft</button></div>
      </section>
      <section data-error-locus="client-transaction-conflict" role="alert">
        <strong>Queued edit needs recovery</strong>
        <p>Blocked edit and later FIFO edits retained</p>
        <div class="error-locus-actions"><button type="button">Retry with a new request ID</button><button type="button">Discard blocked edit</button></div>
      </section>
      <section data-error-locus="stale-refresh" role="alert">
        <strong>Refresh paused</strong>
        <p>Previously authorized rows, selection, and focus retained</p>
        <div class="error-locus-actions"><button type="button">Retry</button></div>
      </section>
      <section data-error-locus="evidence-preview-blocked" role="status">
        <strong>Evidence preview blocked</strong>
        <p>Authorized metadata retained · preview bytes unavailable</p>
        <div class="error-locus-actions"><button type="button">Download</button></div>
      </section>
      <section data-error-locus="permission-loss" role="alert">
        <strong>Incident access changed</strong>
        <p>Protected workbook materialization cleared · authenticated root focused</p>
      </section>
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
    const evidenceFixture = page.getByTestId(
      gridShellTestId(evidenceViewSchemaId),
    );
    await evidenceFixture.evaluate((element) => {
      element.setAttribute("data-design-fixture", "evidence");
    });
    await assertVisualRegression(page, name, evidenceFixture, {
      renderSurface: evidenceViewSchemaId,
    });
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
    const styleId = "evidence-workflow-evidence-access-visual-style";
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
  const responsiveBand = page.getByTestId(workbookResponsiveBandTestId());
  if ((await responsiveBand.count()) > 0) {
    const viewportWidth = page.viewportSize()?.width ?? 0;
    const expectedBand =
      viewportWidth >= 1280
        ? "base"
        : viewportWidth >= 1024
          ? "narrow_desktop"
          : viewportWidth >= 768
            ? "compact_desktop"
            : "below_supported_minimum";
    await expect(responsiveBand).toHaveAttribute(
      "data-workbook-responsive-band",
      expectedBand,
    );
  }
  await page.evaluate(() => {
    document.documentElement.dataset.visualSnapshot = "true";
  });
  await waitForVendoredFonts(page);
  await attachFontManifestDigest();
  await maskVisualDynamicText(page);
}

async function parkVisualPointer(page: Page) {
  await page.mouse.move(720, 360);
  await waitForVisualLayoutFrame(page);
}

async function stabilizeConflictResolverVisual(page: Page) {
  const gridEditorIsActive = await page.evaluate(
    () => document.activeElement?.closest('[role="gridcell"]') !== null,
  );
  if (gridEditorIsActive) {
    await page.keyboard.press("Escape");
  }
  await expect(
    page.getByTestId(workbookConflictResolverTestId()),
  ).toBeVisible();
  await expect(
    page.getByText("Resolve the existing field conflict before editing.", {
      exact: true,
    }),
  ).toHaveCount(0);
  await blurActiveElement(page);
  await parkVisualPointer(page);
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
      [/gpres_[0-9a-f]+…[0-9a-f]+/gi, "gpres_VISUAL…RESULT"],
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

if (
  (process.env.CARTULARY_BROWSER_RUNTIME_PROFILE_ID ?? "default") ===
  "network_flow_claimed"
) {
  test("Capture claimed Network Analysis accepted inspector, rejected diagnostics, and graph contributor drawer at the deterministic desktop viewport.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    await openClaimedNetworkAnalysis(page, "NETWORKFLOWVISUAL");
    const fixture = readFileSync(networkFlowMinimalCSV, "utf8");
    const lines = fixture.trimEnd().split("\n");
    const invalidRow = lines.at(-1)?.replace("192.0.2.10", "not-an-ip") ?? "";
    await importNetworkFlowCSV(page, {
      displayName: "visual-flow",
      file: {
        name: "visual-flow.csv",
        mimeType: "text/csv",
        buffer: Buffer.from(`${fixture.trimEnd()}\n${invalidRow}\n`),
      },
    });
    await page
      .getByRole("gridcell", { name: /Source IP:/u })
      .first()
      .click();
    await expect(
      page.getByTestId(networkAnalysisTestId("inspector")),
    ).toBeVisible();

    await assertViewportVisualRegression(
      page,
      "network-flow-analysis-accepted-inspector",
    );
    await page.getByTestId(networkAnalysisTestId("mode-rejected")).click();
    await assertViewportVisualRegression(
      page,
      "network-flow-analysis-rejected-diagnostics",
    );
    await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
    const edge = page.getByTestId(/^network-flow-edge-/).first();
    await expect(edge).toBeVisible();
    await edge.getByRole("button", { name: "Select edge" }).click();
    const contributorDrawer = page.getByTestId(
      networkAnalysisTestId("contributor-drawer"),
    );
    await expect(contributorDrawer).toBeVisible();
    await expect(contributorDrawer).toContainText("visual-flow");
    await assertViewportVisualRegression(
      page,
      "network-flow-analysis-graph-contributors",
    );
    await page.getByTestId(networkAnalysisTestId("contributor-close")).click();
    await page.getByRole("button", { name: "Saved graphs" }).click();
    await page.getByTestId(networkAnalysisTestId("saved-graph-create")).click();
    await page
      .getByTestId(networkAnalysisTestId("saved-graph-name"))
      .fill("Visual saved graph");
    await page.getByRole("button", { name: "Save graph" }).click();
    const savedGraphs = page.getByTestId(networkAnalysisTestId("saved-graphs"));
    await expect(
      savedGraphs.getByText("Materialization succeeded.", { exact: true }),
    ).toBeVisible({ timeout: 15_000 });
    await expect(
      page.getByTestId(/^network-flow-saved-graph-edge-/u).first(),
    ).toBeVisible();
    await assertViewportVisualRegression(
      page,
      "network-flow-analysis-saved-graph-result",
    );
  });
}
