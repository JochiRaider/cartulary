import {
  scrollGridTargetIntoView,
  scrollGridToBottom,
} from "@cartulary/test-utils/grid";
import {
  mentionCreateEntityButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  rowCellTestId,
  timelineCollectionInputTestId,
  timelineInspectorTestId,
  workbookInspectorCloseButtonTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { openIncidentAsTrackedUserReady } from "./support/collaboration/replay";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  collectionItems,
  findRow,
  hostRefsFieldKey,
  identityRefsFieldKey,
  readMentionAction,
  readMentionActionRequest,
  requireItemByRawText,
  waitForMentionAction,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createTimelineFillers } from "./support/timeline/fixtures";
import {
  expectCollectionControlPainted,
  showTimelineCollectionColumns,
} from "./support/workbook/collections";
import {
  createViewRow,
  queryViewRows,
  waitForViewRow,
} from "./support/workbook/query";
import {
  ensureTimelineGridTargetVisible,
  expectNoPendingQueueAuthPause,
  expectTimelineMutationContinuity,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

test("resolves and creates entities from Timeline mentions in the inspector", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("MENTION-RESOLUTION"),
    "Record relationships entity-resolution",
  );
  const existingHost = await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e401-host"),
      "host.display_name": "WS-023",
      "host.hostname": "ws-023.corp.example.test",
    },
  );

  await createTimelineFillers(page, incidentId, "entity-resolution filler", 12);
  const siblingRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e401-sibling"),
      "timeline.activity_synopsis_text": "entity-resolution sibling unresolved",
      [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
    },
  );
  const mainRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e401-main"),
    "timeline.activity_synopsis_text": "entity-resolution workbook row",
  });
  const identitiesBefore = await queryViewRows(
    page,
    incidentId,
    identitiesViewSchemaId,
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const mainSummaryTestId = rowCellTestId(
    mainRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await ensureTimelineGridTargetVisible(page, mainSummaryTestId);
  await expect(page.getByTestId(mainSummaryTestId)).toHaveText(
    "entity-resolution workbook row",
  );

  const hostAddEnvelope = await addRelationshipTokenViaUI(
    page,
    mainRow.record_id,
    "hostRefs",
    "WS-023?",
  );
  const identityAddEnvelope = await addRelationshipTokenViaUI(
    page,
    mainRow.record_id,
    "identityRefs",
    "vpn.user@example.test",
  );
  const hostMention = requireItemByRawText(
    collectionItems(hostAddEnvelope.data.row, hostRefsFieldKey),
    "WS-023?",
  );
  const identityMention = requireItemByRawText(
    collectionItems(identityAddEnvelope.data.row, identityRefsFieldKey),
    "vpn.user@example.test",
  );

  await expect(
    page
      .getByTestId(relationshipItemsTestId(mainRow.record_id, hostRefsFieldKey))
      .getByLabel("Unresolved host mention: WS-023?"),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(mainRow.record_id, identityRefsFieldKey),
      )
      .getByLabel("Unresolved identity mention: vpn.user@example.test"),
  ).toBeVisible();

  await openTimelineInspector(page, mainRow.record_id);
  await page
    .getByTestId(mentionItemTestId(String(hostMention.item_ref)))
    .click();
  await expect(
    page.getByTestId(mentionResolveTargetSelectTestId()),
  ).toBeVisible();
  await expect(
    page.getByTestId(mentionResolveExistingButtonTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    "WS-023?",
  );

  const resolveScroll = await scrollGridToBottom(page, timelineViewSchemaId);
  await expectNoPendingQueueAuthPause(page, "before resolving host mention");
  const resolveResponsePromise = waitForMentionAction(
    page,
    hostMention.item_ref,
  );
  await page
    .getByTestId(mentionResolveTargetSelectTestId())
    .selectOption(existingHost.record_id);
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  const resolveResponse = await resolveResponsePromise;
  const resolveEnvelope = await readMentionAction(
    resolveResponse,
    mainRow.record_id,
  );
  const resolveBody = readMentionActionRequest(resolveResponse);

  await expect(
    page
      .getByTestId(relationshipItemsTestId(mainRow.record_id, hostRefsFieldKey))
      .getByLabel("Resolved host: WS-023"),
  ).toBeVisible();
  await expectTimelineMutationContinuity(
    page,
    mainRow.record_id,
    resolveEnvelope.data.source_record.row_version,
    resolveScroll,
  );

  await page
    .getByTestId(mentionItemTestId(String(identityMention.item_ref)))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    "vpn.user@example.test",
  );

  const createScroll = await scrollGridToBottom(page, timelineViewSchemaId);
  await expectNoPendingQueueAuthPause(page, "before creating identity mention");
  const createResponsePromise = waitForMentionAction(
    page,
    identityMention.item_ref,
  );
  await page.getByTestId(mentionCreateEntityButtonTestId("identity")).click();
  const createResponse = await createResponsePromise;
  const createEnvelope = await readMentionAction(
    createResponse,
    mainRow.record_id,
  );
  const createdIdentityRecordId = String(
    createEnvelope.data.entity_mention.resolved_record_id,
  );
  expect(createdIdentityRecordId).not.toBe("null");
  expect(createdIdentityRecordId).not.toBe("");
  const createBody = readMentionActionRequest(createResponse);

  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(mainRow.record_id, identityRefsFieldKey),
      )
      .getByLabel("Resolved identity: vpn.user@example.test"),
  ).toBeVisible();
  await expectTimelineMutationContinuity(
    page,
    mainRow.record_id,
    createEnvelope.data.source_record.row_version,
    createScroll,
  );

  const timelineRows = await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  const mainRowAfter = findRow(timelineRows, mainRow.record_id);
  const siblingRowAfter = findRow(timelineRows, siblingRow.record_id);
  const mainHostAfter = requireItemByRawText(
    collectionItems(mainRowAfter, hostRefsFieldKey),
    "WS-023?",
  );
  const mainIdentityAfter = requireItemByRawText(
    collectionItems(mainRowAfter, identityRefsFieldKey),
    "vpn.user@example.test",
  );
  const siblingHostAfter = requireItemByRawText(
    collectionItems(siblingRowAfter, hostRefsFieldKey),
    "WS-023?",
  );
  const createdIdentityRow = await waitForViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    createdIdentityRecordId,
  );

  expect(identitiesBefore).toHaveLength(0);
  expect(resolveBody).toMatchObject({
    base_mention_row_version: hostMention.mention_row_version,
    action: "resolve_item",
    resolved_record_id: existingHost.record_id,
  });
  expect(createBody).toMatchObject({
    base_mention_row_version: identityMention.mention_row_version,
    action: "resolve_item",
    resolved_record_id: createdIdentityRecordId,
  });
  expect(typeof resolveBody.client_txn_id).toBe("string");
  expect(resolveEnvelope.data.entity_mention.resolved_record_id).toBe(
    existingHost.record_id,
  );
  expect(String(mainHostAfter.item_kind)).toBe("resolved_ref");
  expect(String(mainHostAfter.raw_text)).toBe("WS-023?");
  expect(String(mainHostAfter.resolved_record_id)).toBe(existingHost.record_id);
  expect(String(mainIdentityAfter.item_kind)).toBe("resolved_ref");
  expect(String(mainIdentityAfter.raw_text)).toBe("vpn.user@example.test");
  expect(String(mainIdentityAfter.resolved_record_id)).toBe(
    createdIdentityRecordId,
  );
  expect(String(siblingHostAfter.item_kind)).toBe("unresolved_mention");
  expect(siblingHostAfter.resolved_record_id).toBeUndefined();
  expect(createdIdentityRow.record_id).toBe(createdIdentityRecordId);
});

test("collection chips disclose exact members without edits across keyboard and display profiles", async ({
  page,
  browser,
  sessionTracker,
}) => {
  test.setTimeout(180000);
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COLLECTIONINSPECTION"),
    "Collection inspection",
  );
  const rawHidden = "  Hidden host Ω 東京  ";
  const longTag = "Long tag Ω 東京 ".repeat(3).trim();
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("collection-inspection-row"),
    "timeline.activity_synopsis_text": "Inspect complete collections",
    [hostRefsFieldKey]: collectionActionsPayload(["first-host?", rawHidden]),
    [identityRefsFieldKey]: collectionActionsPayload([
      "first-identity?",
      "identity Ω 東京",
    ]),
    "timeline.tags": {
      kind: "collection_actions_v1",
      actions: [
        { op: "add_tag", tag_name: "triage" },
        { op: "add_tag", tag_name: "hidden tag Ω 東京" },
        { op: "add_tag", tag_name: longTag },
      ],
    },
  });
  const hiddenMention = requireItemByRawText(
    collectionItems(row, hostRefsFieldKey),
    rawHidden,
  );
  const mutations: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() !== "GET" &&
      /\/api\/v1\/(?:records|entity-mentions|timeline-records)(?:\/|$)/u.test(
        new URL(request.url()).pathname,
      )
    )
      mutations.push(request.url());
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page, ["Tags", "Hosts", "Identities"]);
  const tagOverflow = page.getByTestId(
    relationshipOverflowButtonTestId(row.record_id, "timeline.tags"),
  );
  const close = page.getByTestId(
    workbookInspectorCloseButtonTestId(timelineViewSchemaId),
  );
  const hiddenTag = page.getByRole("note", {
    name: `Tag: ${String(collectionItems(row, "timeline.tags")[1]?.display_text)}`,
    exact: true,
  });
  for (const key of ["Enter", "Space"]) {
    await expectCollectionControlPainted(tagOverflow);
    await tagOverflow.focus();
    await tagOverflow.press("F2");
    await expect(tagOverflow).toBeFocused();
    await tagOverflow.press(key);
    await expect(hiddenTag).toBeFocused();
    await expect(
      page
        .getByTestId(relationshipItemsTestId(row.record_id, "timeline.tags"))
        .getByRole("note"),
    ).toHaveCount(3);
    const completeLongTag = page.getByRole("note", {
      name: `Tag: ${longTag}`,
      exact: true,
    });
    await completeLongTag.scrollIntoViewIfNeeded();
    await expectCollectionControlPainted(completeLongTag);
    await hiddenTag.focus();
    await hiddenTag.press("Escape");
    await expect(close).toHaveCount(0);
    await expect(tagOverflow).toBeFocused();
  }
  const draftHosts = page.getByRole("textbox", {
    name: "Hosts draft row",
    exact: true,
  });
  await draftHosts.fill("recordless local Ω");
  await tagOverflow.click();
  await expect(hiddenTag).toBeFocused();
  await expect(draftHosts).toHaveValue("recordless local Ω");
  expect(mutations).toEqual([]);
  await close.click();
  await expect(tagOverflow).toBeFocused();
  await draftHosts.fill("");
  const hostCell = page
    .getByRole("group", { name: "Hosts collection cell", exact: true })
    .filter({
      has: page.getByTestId(
        relationshipItemsTestId(row.record_id, hostRefsFieldKey, "grid"),
      ),
    });
  await expect(
    hostCell.getByRole("button", { name: "Add hosts token" }),
  ).toBeVisible();
  await hostCell.getByRole("button", { name: "Add hosts token" }).click();
  const gridInput = page.getByTestId(
    timelineCollectionInputTestId(row.record_id, hostRefsFieldKey, "grid"),
  );
  await gridInput.fill("pending local Ω");
  await tagOverflow.click();
  await expect(hiddenTag).toBeFocused();
  await close.click();
  await expect(gridInput).toHaveValue("pending local Ω");
  expect(mutations).toEqual([]);

  const hostOverflow = page.getByTestId(
    relationshipOverflowButtonTestId(row.record_id, hostRefsFieldKey),
  );
  await hostOverflow.click();
  const hiddenHost = page.getByTestId(
    mentionItemTestId(String(hiddenMention.item_ref)),
  );
  await expect(hiddenHost).toBeFocused();
  await expect(hiddenHost).toHaveAttribute(
    "aria-label",
    `Unresolved host mention: ${rawHidden}`,
  );
  await expect(
    page.getByTestId(
      timelineCollectionInputTestId(row.record_id, hostRefsFieldKey),
    ),
  ).toHaveValue("pending local Ω");
  await hiddenHost.press("Escape");
  await expect(hostOverflow).toBeFocused();
  await expect(gridInput).toHaveValue("pending local Ω");
  expect(mutations).toEqual([]);

  const preferences = (
    await (
      await page.request.get(`${apiBase}/api/v1/account/preferences`)
    ).json()
  ).data;
  const setDensity = async (density: string) => {
    const current = (
      await (
        await page.request.get(`${apiBase}/api/v1/account/preferences`)
      ).json()
    ).data;
    const response = await page.request.put(
      `${apiBase}/api/v1/account/preferences`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_preferences_version: current.preferences_version,
          client_txn_id: uniqueTxn("collection-density"),
          density_mode: density,
        },
      },
    );
    expect(response.ok()).toBeTruthy();
    await page.reload();
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await showTimelineCollectionColumns(page, ["Tags", "Hosts", "Identities"]);
  };
  try {
    for (const density of ["compact", "default", "comfortable"]) {
      await setDensity(density);
      await expectCollectionControlPainted(tagOverflow);
      await expect(tagOverflow).toHaveText("+2");
      const spacing = await page.addStyleTag({
        content:
          "* { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; } p { margin-block-end: 2em !important; }",
      });
      await expectCollectionControlPainted(tagOverflow);
      await tagOverflow.click();
      await expect(hiddenTag).toBeFocused();
      await close.click();
      await expect(tagOverflow).toBeFocused();
      await spacing.evaluate((element) =>
        element.parentNode?.removeChild(element),
      );
    }
    await page.emulateMedia({ reducedMotion: "reduce" });
    for (const width of [1024, 768]) {
      await page.setViewportSize({ width, height: 720 });
      await tagOverflow.scrollIntoViewIfNeeded();
      await expectCollectionControlPainted(tagOverflow);
      await tagOverflow.press("Enter");
      await expect(hiddenTag).toBeFocused();
      await close.click();
      await expect(tagOverflow).toBeFocused();
    }
    await page.setViewportSize({ width: 1440, height: 900 });
    await page.evaluate(() => {
      document.documentElement.style.zoom = "200%";
    });
    await tagOverflow.scrollIntoViewIfNeeded();
    await expectCollectionControlPainted(tagOverflow);
    await tagOverflow.click();
    await expect(hiddenTag).toBeFocused();
    await close.click();
    await page.evaluate(() => {
      document.documentElement.style.zoom = "";
    });
    const viewerPassword = "CollectionViewer1!";
    const viewer = await createIncidentMemberUser(page, incidentId, {
      display_name: "Collection Viewer",
      email: uniqueEmail("collection-viewer"),
      initial_password: viewerPassword,
      role: "viewer",
      is_deployment_admin: false,
      mfa_required: false,
    });
    const viewerSession = await openIncidentAsTrackedUserReady(
      browser,
      sessionTracker,
      {
        createdBy: "collection inspection",
        email: viewer.email,
        incidentId,
        password: viewerPassword,
        purpose: "Read-only complete collection inspection",
        readyRecordId: row.record_id,
        userId: viewer.user_id,
      },
    );
    try {
      const viewerPage = viewerSession.page;
      await viewerPage.setViewportSize({ width: 1440, height: 900 });
      await showTimelineCollectionColumns(viewerPage, [
        "Tags",
        "Hosts",
        "Identities",
      ]);
      await scrollGridTargetIntoView({
        page: viewerPage,
        surface: timelineViewSchemaId,
        targetTestId: relationshipOverflowButtonTestId(
          row.record_id,
          "timeline.tags",
        ),
      });
      await viewerPage
        .getByTestId(
          relationshipOverflowButtonTestId(row.record_id, "timeline.tags"),
        )
        .press("Enter");
      await expect(
        viewerPage
          .getByTestId(relationshipItemsTestId(row.record_id, "timeline.tags"))
          .getByRole("note"),
      ).toHaveCount(3);
      await viewerPage
        .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
        .click();
      await viewerPage
        .getByTestId(
          relationshipOverflowButtonTestId(row.record_id, hostRefsFieldKey),
        )
        .press("Space");
      await expect(
        viewerPage.getByTestId(
          mentionItemTestId(String(hiddenMention.item_ref)),
        ),
      ).toBeFocused();
      await expect(
        viewerPage.getByTestId(mentionResolveExistingButtonTestId()),
      ).toHaveCount(0);
      await expect(
        viewerPage.getByTestId(mentionCreateEntityButtonTestId("host")),
      ).toHaveCount(0);
    } finally {
      await viewerSession.page.context().close();
    }
    const incident = (
      await (
        await page.request.get(`${apiBase}/api/v1/incidents/${incidentId}`)
      ).json()
    ).data;
    const closed = await page.request.post(
      `${apiBase}/api/v1/incidents/${incidentId}/close`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_incident_version: incident.incident_version,
          client_txn_id: uniqueTxn("collection-close"),
          reason: "Inspect closed collections",
        },
      },
    );
    expect(closed.ok()).toBeTruthy();
    await page.reload();
    await showTimelineCollectionColumns(page, ["Tags", "Hosts", "Identities"]);
    await tagOverflow.click();
    await expect(hiddenTag).toBeFocused();
    await expect(
      page.getByTestId(
        timelineCollectionInputTestId(row.record_id, "timeline.tags"),
      ),
    ).toHaveAttribute("readonly", "");
    await close.click();
    await hostOverflow.click();
    await expect(hiddenHost).toBeFocused();
    await expect(
      page.getByTestId(mentionResolveExistingButtonTestId()),
    ).toHaveCount(0);
    await expect(
      page.getByTestId(mentionCreateEntityButtonTestId("host")),
    ).toHaveCount(0);
    expect(mutations).toEqual([]);
  } finally {
    await page.evaluate(() => {
      document.documentElement.style.zoom = "";
    });
    await setDensity(preferences.density_mode);
  }
});
