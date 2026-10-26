import {
  autoResolutionNoticeTestId,
  autoResolutionUndoButtonTestId,
  dataTestIdSelector,
  gridScrollportSelector,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  timelineCollectionInputTestId,
  timelineInspectorTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectVisibleSemanticGridCellFocus } from "./support/accessibility/grid";
import {
  expectKeyboardFocusReachesTestId,
  expectNoFocusTrap,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectAndRecordContrast } from "./support/accessibility/records";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
  identityRefsFieldKey,
  inspectorMentionGeometry,
  placeInspectorMentionNearBottom,
  requireItemByRawText,
  seedHostMentionStateFixture,
} from "./support/entities/mentions";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  expectCollectionControlPainted,
  showTimelineCollectionColumns,
} from "./support/workbook/collections";
import { mountedGridCell } from "./support/workbook/keyboardGrid";
import { createViewRow } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

async function expectDismissedMentionTextContrast(
  page: Page,
  itemRef: string,
  rawText: string,
  context: string,
) {
  const item = page
    .getByTestId(timelineInspectorTestId())
    .getByTestId(mentionItemTestId(itemRef));
  await expect(item).toHaveAccessibleName(`Dismissed mention: ${rawText}`);
  await item.scrollIntoViewIfNeeded();
  const chip = item.getByTestId(relationshipChipTestId(itemRef));
  const selector = `${dataTestIdSelector(timelineInspectorTestId())} ${dataTestIdSelector(mentionItemTestId(itemRef))} ${dataTestIdSelector(relationshipChipTestId(itemRef))}`;
  await expectAndRecordContrast(
    page,
    [
      {
        id: `${context}-label`,
        selector: `${selector} span:not([aria-hidden])`,
        required: true,
        text: rawText,
      },
      {
        id: `${context}-marker`,
        selector: `${selector} span[aria-hidden='true']`,
        required: true,
        text: "dismissed",
      },
    ],
    context,
  );
  await expect(chip).toHaveCount(1);
  await expectCollectionControlPainted(
    chip.getByText("dismissed", { exact: true }),
  );
  await expectCollectionControlPainted(
    chip.getByText(rawText, { exact: true }),
  );
}

async function expectMentionChipContrast(
  page: Page,
  itemRef: string,
  context: string,
) {
  await expectAndRecordContrast(
    page,
    [
      {
        id: context,
        selector: `${dataTestIdSelector(timelineInspectorTestId())} ${dataTestIdSelector(mentionItemTestId(itemRef))} ${dataTestIdSelector(relationshipChipTestId(itemRef))}`,
      },
    ],
    context,
  );
}

test("a11y.mention resolution and undo preserve names and focus", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YENTITYLINKING"),
    "a11y.entity-linking.row-01 mention states",
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
    displayPrefix: "a11y.entity-linking",
    hostnamePrefix: "a11y.entity-linking",
    occurredAt: {
      auto: "2026-06-06T16:15:00Z",
      dismissed: "2026-06-06T16:20:00Z",
      manual: "2026-06-06T16:10:00Z",
      resolved: "2026-06-06T16:05:00Z",
      unresolved: "2026-06-06T16:00:00Z",
    },
    rawTextPrefix: "A11YENTITYLINKING",
    summary: {
      auto: "a11y.entity-linking auto chip",
      dismissed: "a11y.entity-linking dismissed chip",
      manual: "a11y.entity-linking manual chip",
      resolved: "a11y.entity-linking resolved chip",
      unresolved: "a11y.entity-linking unresolved chip",
    },
    txnPrefix: "a11y.entity-linking",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  for (const recordId of [
    unresolvedRow.record_id,
    resolvedRow.record_id,
    manualRow.record_id,
    autoRow.record_id,
    dismissedRow.record_id,
  ]) {
    await expectVisibleSemanticGridCellFocus(
      await mountedGridCell(
        page,
        timelineViewSchemaId,
        recordId,
        "timeline.activity_synopsis_text",
      ),
    );
  }
  await showTimelineCollectionColumns(page);
  await openTimelineInspector(page, unresolvedRow.record_id);
  const unresolvedChip = page.getByTestId(
    mentionItemTestId(String(unresolvedMention.item_ref)),
  );
  await expect(unresolvedChip).toHaveAttribute(
    "aria-label",
    `Unresolved host mention: ${unresolvedRawText}`,
  );
  await unresolvedChip.scrollIntoViewIfNeeded();
  await expectCollectionControlPainted(
    unresolvedChip.getByText("?", { exact: true }),
  );
  await expectVisibleFocus(unresolvedChip);
  await expectMentionChipContrast(
    page,
    String(unresolvedMention.item_ref),
    "unresolved-mention",
  );
  await unresolvedChip.click();
  const unresolvedDismiss = page.getByTestId(mentionDismissButtonTestId());
  await expect(unresolvedDismiss).toBeEnabled();
  await expectVisibleFocus(unresolvedDismiss);
  const createHost = page.getByRole("button", {
    name: "Create host",
    exact: true,
  });
  await expectVisibleFocus(createHost);
  await createHost.press("Enter");
  const createForm = page.getByRole("form", {
    name: "Create host from mention",
  });
  await expect(
    createForm.getByRole("textbox", { name: "Hostname value", exact: true }),
  ).toHaveValue("");
  await expectVisibleFocus(
    createForm.getByRole("textbox", {
      name: "Display Name value",
      exact: true,
    }),
  );
  await expectVisibleFocus(
    createForm.getByRole("button", {
      name: "Create host and resolve",
      exact: true,
    }),
  );
  await expectAllInteractiveControlsNamed(page);
  await createForm.getByRole("button", { name: "Cancel creation" }).click();
  await openTimelineInspector(page, resolvedRow.record_id);
  const resolvedChip = page.getByTestId(
    mentionItemTestId(String(resolvedMention.item_ref)),
  );
  await expect(resolvedChip).toHaveAttribute(
    "aria-label",
    /^Resolved host: a11y.entity-linking Resolved Target$/u,
  );
  await expect(resolvedChip).toContainText("Resolved");
  await expectVisibleFocus(resolvedChip);
  await expectMentionChipContrast(
    page,
    String(resolvedMention.item_ref),
    "resolved-mention",
  );
  await openTimelineInspector(page, manualRow.record_id);
  const manualMentionItem = page.getByTestId(
    mentionItemTestId(String(manualMention.item_ref)),
  );
  await expectVisibleFocus(manualMentionItem);
  await manualMentionItem.click();
  await manualMentionItem.focus();
  const resolveSelect = page.getByTestId(mentionResolveTargetSelectTestId());
  await expect(resolveSelect).toHaveAccessibleName("Search targets");
  await expectKeyboardFocusReachesTestId(
    page,
    mentionResolveTargetSelectTestId(),
  );
  await expectVisibleFocus(resolveSelect);
  await expect(resolveSelect).toBeEnabled();
  await resolveSelect.fill("a11y.entity-linking Manual Target");
  await page
    .getByRole("option", {
      name: "a11y.entity-linking Manual Target",
      exact: true,
    })
    .waitFor();
  await resolveSelect.press("ArrowDown");
  await expect(resolveSelect).toHaveAttribute(
    "aria-activedescendant",
    new RegExp(manualTarget.record_id),
  );
  await resolveSelect.press("Enter");
  const resolveButton = page.getByTestId(mentionResolveExistingButtonTestId());
  await expectVisibleFocus(resolveButton);
  await resolveButton.click();
  const manualChip = page.getByTestId(
    mentionItemTestId(String(manualMention.item_ref)),
  );
  await expect(manualChip).toHaveAttribute(
    "aria-label",
    /^Resolved host: a11y.entity-linking Manual Target$/u,
  );
  await expect(manualChip).toContainText("Resolved");
  await expectVisibleFocus(manualChip);
  await expectMentionChipContrast(
    page,
    String(manualMention.item_ref),
    "manual-resolved-mention",
  );
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
  const autoChip = page.getByTestId(
    mentionItemTestId(String(autoItem.item_ref)),
  );
  await expect(autoChip).toHaveAttribute(
    "aria-label",
    /^Auto-resolved .*matched/u,
  );
  await autoChip.scrollIntoViewIfNeeded();
  await expectCollectionControlPainted(
    autoChip.getByText("auto", { exact: true }),
  );
  await expectVisibleFocus(autoChip);
  await expectMentionChipContrast(
    page,
    String(autoItem.item_ref),
    "auto-resolved-mention",
  );
  const autoNotice = page.getByTestId(
    autoResolutionNoticeTestId(String(autoItem.item_ref)),
  );
  await expect(autoNotice).toContainText("a11y.entity-linking Auto Target");
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
    .fill("a11y.entity-linking Manual Target");
  await page
    .getByRole("option", {
      name: "a11y.entity-linking Manual Target",
      exact: true,
    })
    .click();
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  await expect(
    page.getByTestId(mentionItemTestId(String(dismissedMention.item_ref))),
  ).toHaveAccessibleName("Resolved host: a11y.entity-linking Manual Target");
  await page.getByTestId(mentionDismissButtonTestId()).click();
  const dismissedMentionItem = page
    .getByTestId(timelineInspectorTestId())
    .getByRole("button", { name: `Dismissed mention: ${dismissedRawText}` });
  await expect(dismissedMentionItem).toBeVisible();
  await dismissedMentionItem.scrollIntoViewIfNeeded();
  await expectCollectionControlPainted(
    dismissedMentionItem.getByText("dismissed", { exact: true }),
  );
  await expectVisibleFocus(dismissedMentionItem);
  await expectVisibleFocus(
    page.getByTestId(mentionRestoreUnresolvedButtonTestId()),
  );
  await expectDismissedMentionTextContrast(
    page,
    String(dismissedMention.item_ref),
    dismissedRawText,
    "dismissed-host-selected-1440x900",
  );
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await expectAndRecordContrast(page, [mentionRestoreUnresolvedButtonTestId()]);
});

test("a11y.dismissed host and identity source text remains readable in active and closed incidents", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YENTITYLINKING"),
    "a11y.entity-linking.row-01 mention states",
  );
  const { dismissedMention, dismissedRawText, dismissedRow } =
    await seedHostMentionStateFixture(page, incidentId, {
      displayPrefix: "a11y.entity-linking",
      hostnamePrefix: "a11y.entity-linking",
      occurredAt: {
        auto: "2026-06-06T16:15:00Z",
        dismissed: "2026-06-06T16:20:00Z",
        manual: "2026-06-06T16:10:00Z",
        resolved: "2026-06-06T16:05:00Z",
        unresolved: "2026-06-06T16:00:00Z",
      },
      rawTextPrefix: "A11YENTITYLINKING",
      summary: {
        auto: "a11y.entity-linking auto chip",
        dismissed: "a11y.entity-linking dismissed chip",
        manual: "a11y.entity-linking manual chip",
        resolved: "a11y.entity-linking resolved chip",
        unresolved: "a11y.entity-linking unresolved chip",
      },
      txnPrefix: "a11y.entity-linking",
    });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await showTimelineCollectionColumns(page);
  await openTimelineInspector(page, dismissedRow.record_id);
  await page
    .getByTestId(mentionItemTestId(String(dismissedMention.item_ref)))
    .click();
  await page
    .getByTestId(mentionResolveTargetSelectTestId())
    .fill("a11y.entity-linking Manual Target");
  await page
    .getByRole("option", {
      name: "a11y.entity-linking Manual Target",
      exact: true,
    })
    .click();
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  await page.getByTestId(mentionDismissButtonTestId()).click();
  const dismissedMentionItem = page
    .getByTestId(timelineInspectorTestId())
    .getByRole("button", { name: `Dismissed mention: ${dismissedRawText}` });
  const identityRawText = `A11Y-DISMISSED-${"long-source-account-".repeat(6)}Ω@example.test?`;
  await createViewRow(page, incidentId, identitiesViewSchemaId, {
    client_txn_id: uniqueTxn("a11y-dismissed-identity-target"),
    "identity.display_name": "A11Y dismissed identity prior target",
    "identity.upn": "a11y-dismissed-identity@example.test",
  });
  const identityEnvelope = await addRelationshipTokenViaUI(
    page,
    dismissedRow.record_id,
    "identityRefs",
    identityRawText,
  );
  const identityMention = requireItemByRawText(
    collectionItems(identityEnvelope.data.row, identityRefsFieldKey),
    identityRawText,
  );
  const identityItem = page.getByTestId(
    mentionItemTestId(String(identityMention.item_ref)),
  );
  await identityItem.click();
  await page
    .getByTestId(mentionResolveTargetSelectTestId())
    .fill("A11Y dismissed identity prior target");
  await page
    .getByRole("option", {
      name: "A11Y dismissed identity prior target",
      exact: true,
    })
    .click();
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  await expect(identityItem).toHaveAccessibleName(
    `Resolved identity: ${identityRawText}`,
  );
  await page.getByTestId(mentionDismissButtonTestId()).click();
  await expect(identityItem).toHaveAccessibleName(
    `Dismissed mention: ${identityRawText}`,
  );
  await expect(
    page.getByTestId(
      relationshipItemsTestId(dismissedRow.record_id, identityRefsFieldKey),
    ),
  ).toContainText("No items");
  const dismissedCases = [
    {
      entityType: "host",
      item: dismissedMentionItem,
      itemRef: String(dismissedMention.item_ref),
      rawText: dismissedRawText,
      section: "Selected Hosts item",
    },
    {
      entityType: "identity",
      item: identityItem,
      itemRef: String(identityMention.item_ref),
      rawText: identityRawText,
      section: "Selected Identities item",
    },
  ];
  const checkDismissedStates = async (context: string, closed = false) => {
    for (const [index, mention] of dismissedCases.entries()) {
      await mention.item.click();
      await expect(mention.item).toHaveAttribute("aria-pressed", "true");
      await expectDismissedMentionTextContrast(
        page,
        mention.itemRef,
        mention.rawText,
        `${context}-${mention.entityType}-selected`,
      );
      const detail = page.getByRole("region", {
        name: mention.section,
        exact: true,
      });
      const disclosure = detail.getByText("Mention details", { exact: true });
      if (
        !(await disclosure.evaluate(
          (element) => element.closest("details")?.open,
        ))
      )
        await disclosure.click();
      const sourceText = detail.getByText(mention.rawText, { exact: true });
      await sourceText.scrollIntoViewIfNeeded();
      await expectCollectionControlPainted(sourceText);
      expect(await sourceText.textContent()).toBe(mention.rawText);
      if (closed)
        await expect(
          page.getByTestId(mentionRestoreUnresolvedButtonTestId()),
        ).toBeDisabled();
      await dismissedCases[index === 0 ? 1 : 0]?.item.click();
      await expect(mention.item).toHaveAttribute("aria-pressed", "false");
      await expectDismissedMentionTextContrast(
        page,
        mention.itemRef,
        mention.rawText,
        `${context}-${mention.entityType}-unselected`,
      );
    }
    const geometry = await page
      .getByTestId(timelineInspectorTestId())
      .evaluate((element) => ({
        clientWidth: element.clientWidth,
        scrollWidth: element.scrollWidth,
      }));
    expect(geometry.scrollWidth).toBeLessThanOrEqual(geometry.clientWidth + 1);
  };
  for (const layout of [
    { width: 1440, height: 900, zoom: 100, spacing: false },
    { width: 1024, height: 720, zoom: 100, spacing: false },
    { width: 1024, height: 720, zoom: 200, spacing: false },
    { width: 320, height: 720, zoom: 100, spacing: false },
    { width: 1024, height: 720, zoom: 100, spacing: true },
  ]) {
    await page.setViewportSize({
      width: layout.width,
      height: layout.height,
    });
    await withPresentation(page, { zoom: `${layout.zoom}%` }, async () => {
      await withPresentation(
        page,
        layout.spacing
          ? {
              textSpacing: `
        ${dataTestIdSelector(timelineInspectorTestId())} * {
          line-height: 1.5 !important; letter-spacing: .12em !important; word-spacing: .16em !important;
        }
        ${dataTestIdSelector(timelineInspectorTestId())} p { margin-block-end: 2em !important; }
      `,
            }
          : {},
        async () => {
          await checkDismissedStates(
            `dismissed-${layout.width}x${layout.height}-zoom${layout.zoom}-spacing${layout.spacing}`,
          );
        },
      );
    });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  const lifecycle = await currentLifecycle(page, incidentId);
  expect(
    (
      await lifecycleAction(page, incidentId, "closeIncident", {
        client_txn_id: uniqueTxn("a11y-dismissed-close"),
        base_incident_version: lifecycle.incident_version,
        reason: "Verify retained dismissed text in closed read-only mode",
      })
    ).ok,
  ).toBe(true);
  await expect(
    page.getByRole("status", { name: "Incident lifecycle", exact: true }),
  ).toHaveText("Closed, read-only");
  await openTimelineInspector(page, dismissedRow.record_id);
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 720 },
  ]) {
    await page.setViewportSize(viewport);
    await checkDismissedStates(
      `dismissed-closed-${viewport.width}x${viewport.height}`,
      true,
    );
  }
});

test("a11y.entity-linking mention arrows reveal Inspector focus without activation", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("MENTIONFOCUS"),
    "Mention keyboard focus reveal",
  );
  await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("mention-focus-target"),
    "host.display_name": "Focus candidate",
    "host.hostname": "focus-candidate.example.test",
  });
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("mention-focus-source"),
    "timeline.activity_synopsis_text": "Mention focus source",
    [hostRefsFieldKey]: collectionActionsPayload([
      "focus-host-a?",
      "focus-host-b?",
      "focus-host-c?",
    ]),
    "timeline.identity_refs": collectionActionsPayload([
      "focus-identity-a?",
      "focus-identity-b?",
    ]),
  });
  const hostItems = collectionItems(row, hostRefsFieldKey);
  const first = page.getByTestId(
    mentionItemTestId(
      String(requireItemByRawText(hostItems, "focus-host-a?").item_ref),
    ),
  );
  const second = page.getByTestId(
    mentionItemTestId(
      String(requireItemByRawText(hostItems, "focus-host-b?").item_ref),
    ),
  );
  const third = page.getByTestId(
    mentionItemTestId(
      String(requireItemByRawText(hostItems, "focus-host-c?").item_ref),
    ),
  );
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openTimelineInspector(page, row.record_id);
  const inspector = page.getByTestId(timelineInspectorTestId());
  await inspector
    .getByRole("button", { name: "Relationships", exact: true })
    .click();
  await first.click();
  const details = inspector.getByText("Mention details", { exact: true });
  await details.click();
  await expect(
    inspector.getByText("1 target on this page.", { exact: true }),
  ).toBeVisible();
  await inspector.getByRole("combobox").click();
  await inspector
    .getByRole("option", { name: "Focus candidate", exact: true })
    .click();
  const candidate = inspector.getByText("Selected: Focus candidate", {
    exact: true,
  });
  await expect(candidate).toBeVisible();
  const draft = page.getByTestId(
    timelineCollectionInputTestId(row.record_id, hostRefsFieldKey),
  );
  const raw = "  retained Inspector host draft Ω  ";
  await draft.fill(raw);
  await draft.evaluate((node) =>
    (node as HTMLInputElement).setSelectionRange(4, 12),
  );
  const outsideScroll = () =>
    page.evaluate((selector) => {
      const grid = document.querySelector<HTMLElement>(selector);
      return {
        x: scrollX,
        y: scrollY,
        gridTop: grid?.scrollTop,
        gridLeft: grid?.scrollLeft,
      };
    }, gridScrollportSelector());
  const requests: string[] = [];
  page.on("request", (request) => {
    if (new URL(request.url()).pathname.startsWith("/api/v1/"))
      requests.push(`${request.method()} ${new URL(request.url()).pathname}`);
  });
  for (const layout of [
    { width: 1440, height: 900, zoom: 1 },
    { width: 1024, height: 720, zoom: 1 },
    { width: 1280, height: 720, zoom: 2 },
  ]) {
    await page.setViewportSize({ width: layout.width, height: layout.height });
    await withPresentation(page, { zoom: String(layout.zoom) }, async () => {
      await placeInspectorMentionNearBottom(first);
      const origin = await inspectorMentionGeometry(first);
      expect(origin.contained, JSON.stringify(origin)).toBe(true);
      expect((await inspectorMentionGeometry(second)).contained).toBe(false);
      const outside = await outsideScroll();
      requests.length = 0;
      // From this key until the assertions, every helper only observes.
      await page.keyboard.press("ArrowRight");
      await expect(second).toBeFocused();
      const forward = await inspectorMentionGeometry(second);
      await info.attach(`mention-forward-${layout.width}-${layout.zoom}`, {
        body: JSON.stringify(forward),
        contentType: "application/json",
      });
      await info.attach(
        `mention-forward-image-${layout.width}-${layout.zoom}`,
        {
          body: await page.screenshot(),
          contentType: "image/png",
        },
      );
      expect(forward.fits && forward.contained, JSON.stringify(forward)).toBe(
        true,
      );
      expect(await outsideScroll()).toEqual(outside);
      await expect(inspector).toHaveAttribute("data-record-id", row.record_id);
      await expect(candidate).toHaveText("Selected: Focus candidate");
      await expect(first).toHaveAttribute("aria-pressed", "true");
      await expect(second).toHaveAttribute("aria-pressed", "false");
      await expect(draft).toHaveValue(raw);
      expect(
        await draft.evaluate((node) => [
          (node as HTMLInputElement).selectionStart,
          (node as HTMLInputElement).selectionEnd,
        ]),
      ).toEqual([4, 12]);
      expect(
        await details.evaluate(
          (node) => (node.parentElement as HTMLDetailsElement).open,
        ),
      ).toBe(true);
      expect(requests).toEqual([]);
      // Arrange a separate reverse transition with the earlier item above the clip.
      await second.evaluate((node) => {
        const body = node.closest<HTMLElement>("[data-inspector-scroll-body]");
        if (!body) throw new Error("Missing Inspector body");
        const scale = body.getBoundingClientRect().height / body.offsetHeight;
        body.scrollTop +=
          (node.getBoundingClientRect().top -
            (body.getBoundingClientRect().top + 24 * scale)) /
          scale;
        (node as HTMLElement).focus({ preventScroll: true });
      });
      expect((await inspectorMentionGeometry(first)).contained).toBe(false);
      await page.keyboard.press("ArrowLeft");
      await expect(first).toBeFocused();
      const reverse = await inspectorMentionGeometry(first);
      expect(reverse.contained, JSON.stringify(reverse)).toBe(true);
      await info.attach(`mention-reverse-${layout.width}-${layout.zoom}`, {
        body: JSON.stringify(reverse),
        contentType: "application/json",
      });
      const boundaryTop = reverse.scrollTop;
      await page.keyboard.press("ArrowLeft");
      expect((await inspectorMentionGeometry(first)).scrollTop).toBe(
        boundaryTop,
      );
      await expect(first).toBeFocused();
      // Adjacent lower items fit together and require no reveal movement.
      await placeInspectorMentionNearBottom(third);
      await second.evaluate((node) =>
        (node as HTMLElement).focus({ preventScroll: true }),
      );
      expect((await inspectorMentionGeometry(second)).contained).toBe(true);
      const visibleTop = (await inspectorMentionGeometry(second)).scrollTop;
      await page.keyboard.press("ArrowRight");
      await expect(third).toBeFocused();
      expect((await inspectorMentionGeometry(third)).scrollTop).toBe(
        visibleTop,
      );
      await page.keyboard.press("ArrowRight");
      await expect(third).toBeFocused();
      expect((await inspectorMentionGeometry(third)).scrollTop).toBe(
        visibleTop,
      );
      expect(await outsideScroll()).toEqual(outside);
      expect(requests).toEqual([]);
    });
  }
  await page.keyboard.press("Enter");
  await expect(third).toHaveAttribute("aria-pressed", "true");
  await expect(first).toHaveAttribute("aria-pressed", "false");
  await expect(draft).toHaveValue(raw);
  await page.keyboard.press("ArrowLeft");
  await expect(second).toBeFocused();
  await page.keyboard.press("Space");
  await expect(second).toHaveAttribute("aria-pressed", "true");
  await expect(third).toHaveAttribute("aria-pressed", "false");
  await expect(inspector).toHaveAttribute("data-record-id", row.record_id);
  await expect(draft).toHaveValue(raw);
});
