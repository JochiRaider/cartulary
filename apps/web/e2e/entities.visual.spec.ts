import {
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  relationshipOverflowButtonTestId,
  saveStateTestId,
  timelineInspectorTestId,
  workbookInspectorCloseButtonTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect } from "./fixtures";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
  requireItemByRawText,
  resolvedRefPayload,
  seedHostMentionStateFixture,
} from "./support/entities/mentions";
import { tagActionsPayload } from "./support/entities/tagActions";
import {
  navigateVisualApplication,
  reloadVisualApplication,
} from "./support/runtime/applicationReadiness";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  assertViewportVisualRegression,
  assertVisualRegression,
  settleVisualGeometry,
} from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import {
  createRichTimelineIncident,
  createVisualTimelineRow,
} from "./support/visual/timelineVisualFixture";
import { blurActiveElement } from "./support/visual/workbookCaptures";
import { normalizeWorkbookGridVisualState } from "./support/visual/workbookGeometry";
import {
  expectCollectionControlPainted,
  showTimelineCollectionColumns,
} from "./support/workbook/collections";
import { createViewRow } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test("Capture unresolved token, resolved chip, auto-resolved chip, dismissed mention, and manual resolution metadata fixtures.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const incidentId = await createRichTimelineIncident(
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
    resolvedMention,
    resolvedRow,
    unresolvedRawText,
    unresolvedRow,
  } = await seedHostMentionStateFixture(page, incidentId, {
    createTimelineRow: (key, fields) =>
      createVisualTimelineRow(page, incidentId, `mention-${key}`, fields),
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

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

  await showTimelineCollectionColumns(page, ["Hosts", "Identities", "Tags"]);
  await openTimelineInspector(page, unresolvedRow.record_id);
  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(unresolvedRow.record_id, hostRefsFieldKey),
      )
      .getByLabel(`Unresolved host mention: ${unresolvedRawText}`),
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
    .fill("visual.entity-linking Manual Target");
  await page
    .getByRole("option", {
      name: "visual.entity-linking Manual Target",
      exact: true,
    })
    .click();
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  await expect(
    page.getByTestId(mentionItemTestId(String(manualMention.item_ref))),
  ).toContainText("Resolved");

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
      .getByTestId(relationshipItemsTestId(autoRow.record_id, hostRefsFieldKey))
      .getByTestId(relationshipChipTestId(String(autoItem.item_ref))),
  ).toContainText("auto");

  await test.step("reload visual application", () =>
    reloadVisualApplication(page));

  await showTimelineCollectionColumns(page);
  await openTimelineInspector(page, dismissedRow.record_id);
  await page
    .getByTestId(mentionItemTestId(String(dismissedMention.item_ref)))
    .click();
  await page
    .getByTestId(mentionResolveTargetSelectTestId())
    .fill("visual.entity-linking Manual Target");
  await page
    .getByRole("option", {
      name: "visual.entity-linking Manual Target",
      exact: true,
    })
    .click();
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  await expect(
    page.getByTestId(mentionItemTestId(String(dismissedMention.item_ref))),
  ).toHaveAccessibleName("Resolved host: visual.entity-linking Manual Target");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await page.getByTestId(mentionDismissButtonTestId()).click();
  await expect(
    page.getByTestId(mentionItemTestId(String(dismissedMention.item_ref))),
  ).toHaveAccessibleName(`Dismissed mention: ${dismissedRawText}`);

  const dismissedMentionItem = page.getByTestId(
    mentionItemTestId(String(dismissedMention.item_ref)),
  );
  await dismissedMentionItem.click();
  await expect(
    page.getByTestId(mentionRestoreUnresolvedButtonTestId()),
  ).toBeVisible();
  await blurActiveElement(page);
  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await expectCollectionControlPainted(
    dismissedMentionItem.getByText("dismissed", { exact: true }),
  );
  await expectCollectionControlPainted(
    page
      .getByTestId(
        relationshipItemsTestId(
          unresolvedRow.record_id,
          hostRefsFieldKey,
          "grid",
        ),
      )
      .getByText("?", { exact: true }),
  );
  await expectCollectionControlPainted(
    page
      .getByTestId(
        relationshipItemsTestId(autoRow.record_id, hostRefsFieldKey, "grid"),
      )
      .getByText("auto", { exact: true }),
  );
  await expect(
    page
      .getByTestId(timelineInspectorTestId())
      .getByText("Manual", { exact: true }),
  ).toHaveCount(0);
  await blurActiveElement(page);
  const chipFixture = page.locator("main.cartulary-shell").first();
  await chipFixture.evaluate((element) => {
    element.setAttribute("data-design-fixture", "chips");
  });
  await assertVisualRegression(
    page,
    "entity-mention-chip-states",
    chipFixture,
    {
      prepareState: async () => {
        const scrollportSelector = `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`;
        await settleVisualGeometry(page, {
          locator: dismissedMentionItem,
          align: "start",
          scrollportSelector,
        });
        // Preserve the fixture's existing eight-pixel inset within the
        // declared body, after resetting all outer scroll containers.
        await page.locator(scrollportSelector).evaluate((element) => {
          element.scrollTop = Math.max(0, element.scrollTop - 8);
        });
      },
      verifyFraming: async () => {
        const inspector = page.getByTestId(timelineInspectorTestId());
        await expect(
          inspector.getByRole("heading", { level: 2 }),
        ).toBeInViewport({ ratio: 1 });
        await expect(
          inspector.getByRole("button", {
            name: "Relationships",
            exact: true,
          }),
        ).toHaveAttribute("aria-current", "location");
        await expect
          .poll(() =>
            dismissedMentionItem.evaluate((element) => {
              const body = element.closest("[data-inspector-scroll-body]");
              if (!body)
                throw new Error("Mention fixture requires its Inspector body");
              return Math.round(
                element.getBoundingClientRect().top -
                  body.getBoundingClientRect().top,
              );
            }),
          )
          .toBe(8);
      },
    },
  );
});

test("captures Timeline unresolved and resolved mention chips in the workbook grid", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALENTITYLINKINGAUX"),
    "Entity linking visual mention chips",
  );
  const hostRow = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("VISUALENTITYLINKINGAUX-HOST"),
    "host.display_name": "WS-023",
    "host.hostname": "ws-023.visual.example.test",
  });
  const unresolvedRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualentitylinkingaux-unresolved",
    {
      client_txn_id: uniqueTxn("VISUALENTITYLINKINGAUX-UNRESOLVED"),
      "timeline.activity_utc_text": "2026-07-15T12:00:00Z",
      "timeline.activity_synopsis_text": "Unresolved mention visual row",
      [hostRefsFieldKey]: collectionActionsPayload([
        "WS-023?",
        "Hidden host Ω 東京",
      ]),
      "timeline.tags": tagActionsPayload([
        "triage",
        "Hidden tag Ω 東京",
        "Evidence follow-up",
      ]),
    },
  );
  const resolvedRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualentitylinkingaux-resolved",
    {
      client_txn_id: uniqueTxn("VISUALENTITYLINKINGAUX-RESOLVED"),
      "timeline.activity_utc_text": "2026-07-15T12:01:00Z",
      "timeline.activity_synopsis_text": "Resolved mention visual row",
      [hostRefsFieldKey]: resolvedRefPayload("WS-023", hostRow.record_id),
    },
  );

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  await showTimelineCollectionColumns(page, ["Hosts", "Identities", "Tags"]);
  await openTimelineInspector(page, unresolvedRow.record_id);
  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(unresolvedRow.record_id, hostRefsFieldKey),
      )
      .getByLabel("Unresolved host mention: WS-023?"),
  ).toBeVisible();
  await openTimelineInspector(page, resolvedRow.record_id);
  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(resolvedRow.record_id, hostRefsFieldKey),
      )
      .getByLabel(/^Resolved host: WS-023$/u),
  ).toBeVisible();

  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await expectCollectionControlPainted(
    page.getByTestId(
      relationshipOverflowButtonTestId(
        unresolvedRow.record_id,
        hostRefsFieldKey,
      ),
    ),
  );
  await expectCollectionControlPainted(
    page.getByTestId(
      relationshipOverflowButtonTestId(
        unresolvedRow.record_id,
        "timeline.tags",
      ),
    ),
  );
  await page
    .getByTestId(
      relationshipOverflowButtonTestId(
        unresolvedRow.record_id,
        "timeline.tags",
      ),
    )
    .click();
  const tags = page.getByTestId(
    relationshipItemsTestId(unresolvedRow.record_id, "timeline.tags"),
  );
  await expect(tags.getByRole("note")).toHaveCount(3);
  await tags.scrollIntoViewIfNeeded();
  for (const tag of await tags.getByRole("note").all())
    await expectCollectionControlPainted(tag);
  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });
  await blurActiveElement(page);
  await assertViewportVisualRegression(
    page,
    "record-relationships-mention-chips",
  );
});
