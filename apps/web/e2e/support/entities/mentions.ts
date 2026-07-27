import {
  pendingQueueNoticeTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  timelineCollectionInputTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import { expect, type Page, type Response } from "@playwright/test";

import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "../contracts/workbookSurfaces";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow, type ViewApiRow } from "../workbook/query";
import {
  openTimelineInspector,
  readTimelineMutation,
  waitForTimelinePatch,
} from "../workbook/rowMutations";

export const hostRefsFieldKey = "timeline.host_refs";
export const identityRefsFieldKey = "timeline.identity_refs";

export type ViewRow = ViewApiRow;

export type MentionActionEnvelope = {
  data: {
    incident_id: string;
    entity_mention: {
      entity_mention_id: string;
      source_record_id: string;
      source_field_key: string;
      entity_type: "host" | "identity" | string;
      raw_text: string;
      resolution_status: "unresolved" | "resolved" | "dismissed" | string;
      resolved_record_id: string | null;
      row_version: number;
      resolution_method: string | null;
    };
    source_record: {
      record_id: string;
      row_version: number;
    };
    change_set_id: string;
  };
};

type CollectionItem = Record<string, unknown>;

type TimelinePatchRequestPayload = {
  base_row_version?: unknown;
};

export async function addRelationshipTokenViaUI(
  page: Page,
  recordId: string,
  draftKey: "hostRefs" | "identityRefs",
  rawText: string,
  options: {
    onPatchRequest?: (payload: TimelinePatchRequestPayload) => void;
    requireVisibleChip?: boolean;
  } = {},
) {
  const fieldKey =
    draftKey === "identityRefs" ? identityRefsFieldKey : hostRefsFieldKey;
  const inputTestId = timelineCollectionInputTestId(recordId, fieldKey);
  await openTimelineInspector(page, recordId);
  const input = page.getByTestId(inputTestId);
  await input.focus();
  await expect(input).toBeVisible();
  const responsePromise = waitForTimelinePatch(page, recordId);
  await input.fill(rawText);
  await input.press("Enter");
  const response = await responsePromise;
  const requestPayload = readRequestPayload(response);
  const envelope = await readTimelineMutation(response);
  options.onPatchRequest?.(requestPayload);
  const item = requireItemByRawText(
    collectionItems(envelope.data.row, fieldKey),
    rawText,
  );
  if (options.requireVisibleChip === true) {
    await expect(
      page
        .getByTestId(relationshipItemsTestId(recordId, fieldKey))
        .getByTestId(relationshipChipTestId(String(item.item_ref))),
    ).toBeVisible();
  }
  await expect
    .poll(
      async () => ({
        inputValue: await input.inputValue().catch((error: unknown) => {
          return `<<failed to read input value: ${String(error)}>>`;
        }),
        pendingQueueNoticeCount: await page
          .getByTestId(pendingQueueNoticeTestId())
          .count(),
        renderedRowVersion: await page
          .getByTestId(timelineRowVersionTestId(recordId))
          .textContent()
          .catch((error: unknown) => {
            return `<<failed to read row version: ${String(error)}>>`;
          }),
      }),
      {
        message: [
          "relationship token commit did not converge",
          `record_id=${recordId}`,
          `draft_key=${draftKey}`,
          `raw_text=${JSON.stringify(rawText)}`,
          `request_payload=${JSON.stringify(requestPayload)}`,
          `response_row_version=${envelope.data.row.row_version}`,
        ].join("\n"),
      },
    )
    .toEqual({
      inputValue: "",
      pendingQueueNoticeCount: 0,
      renderedRowVersion: String(envelope.data.row.row_version),
    });
  return envelope;
}

function readRequestPayload(response: Response): TimelinePatchRequestPayload {
  const postData = response.request().postData();
  if (!postData) return {};
  try {
    const parsed = JSON.parse(postData) as unknown;
    return parsed && typeof parsed === "object" && !Array.isArray(parsed)
      ? (parsed as TimelinePatchRequestPayload)
      : {};
  } catch {
    return {};
  }
}

export function collectionActionsPayload(rawTexts: string[]) {
  return {
    kind: "collection_actions_v1",
    actions: rawTexts.map((rawText) => ({
      op: "add_token",
      raw_text: rawText,
    })),
  };
}

export function aliasCollectionActionsPayload(aliasTexts: string[]) {
  return {
    kind: "collection_actions_v1",
    actions: aliasTexts.map((aliasText) => ({
      op: "add_alias",
      alias_text: aliasText,
    })),
  };
}

export function resolvedRefPayload(rawText: string, resolvedRecordId: string) {
  return {
    kind: "collection_actions_v1",
    actions: [
      {
        op: "add_resolved_ref",
        raw_text: rawText,
        resolved_record_id: resolvedRecordId,
      },
    ],
  };
}

export function entityMentionIdFromItemRef(itemRef: unknown) {
  const value = String(itemRef);
  expect(value.startsWith("entity_mention:")).toBe(true);
  const mentionId = value.slice("entity_mention:".length);
  expect(mentionId).not.toBe("");
  return mentionId;
}

export function waitForMentionAction(page: Page, itemRef: unknown) {
  const mentionId = entityMentionIdFromItemRef(itemRef);
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/entity-mentions/${mentionId}/resolve`),
  );
}

export async function readMentionAction(
  response: Response,
  expectedSourceRecordId: string,
) {
  expect(response.ok()).toBeTruthy();
  const envelope = (await response.json()) as MentionActionEnvelope;
  expect(envelope.data.source_record.record_id).toBe(expectedSourceRecordId);
  expect(
    Number.isSafeInteger(envelope.data.source_record.row_version),
  ).toBeTruthy();
  expect(envelope.data.source_record.row_version).toBeGreaterThan(0);
  expect(envelope.data.entity_mention.source_record_id).toBe(
    expectedSourceRecordId,
  );
  return envelope;
}

export function readMentionActionRequest(response: Response) {
  return JSON.parse(response.request().postData() ?? "{}") as Record<
    string,
    unknown
  >;
}

export function collectionItems(
  row: { cells: Record<string, unknown> },
  fieldKey: string,
) {
  const cell = row.cells[fieldKey];
  const cellValue =
    cell !== null && typeof cell === "object" && "value" in cell
      ? cell.value
      : undefined;
  if (
    !cellValue ||
    typeof cellValue !== "object" ||
    Array.isArray(cellValue) ||
    !("items" in cellValue)
  ) {
    return [] as CollectionItem[];
  }
  const items = (cellValue as { items?: unknown }).items;
  if (!Array.isArray(items)) {
    return [] as CollectionItem[];
  }
  return items.filter(
    (item): item is CollectionItem =>
      item !== null && typeof item === "object" && !Array.isArray(item),
  );
}

export function findRow(rows: ViewRow[], recordId: string) {
  const row = rows.find((candidate) => candidate.record_id === recordId);
  if (!row) throw new Error(`missing row ${recordId}`);
  return row;
}

export function requireItemByRawText(items: CollectionItem[], rawText: string) {
  const item = items.find((candidate) => candidate.raw_text === rawText);
  if (!item) throw new Error(`missing collection item raw_text=${rawText}`);
  return item;
}

export async function seedHostMentionStateFixture(
  page: Page,
  incidentId: string,
  options: {
    displayPrefix: string;
    hostnamePrefix: string;
    occurredAt: {
      auto: string;
      dismissed: string;
      manual: string;
      resolved: string;
      unresolved: string;
    };
    rawTextPrefix: string;
    summary: {
      auto: string;
      dismissed: string;
      manual: string;
      resolved: string;
      unresolved: string;
    };
    txnPrefix: string;
  },
) {
  const resolvedTarget = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-resolved-target`),
      "host.display_name": `${options.displayPrefix} Resolved Target`,
      "host.hostname": `${options.hostnamePrefix}-resolved-target.example.test`,
    },
  )) as ViewRow;
  const manualTarget = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-manual-target`),
      "host.display_name": `${options.displayPrefix} Manual Target`,
      "host.hostname": `${options.hostnamePrefix}-manual-target.example.test`,
    },
  )) as ViewRow;
  await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn(`${options.txnPrefix}-auto-target`),
    "host.display_name": `${options.displayPrefix} Auto Target`,
    "host.hostname": `${options.hostnamePrefix}-auto-target.example.test`,
    "host.aliases": aliasCollectionActionsPayload([
      `${options.rawTextPrefix} Auto Alias`,
    ]),
  });
  const unresolvedRawText = `${options.rawTextPrefix} Unresolved?`;
  const resolvedRawText = `${options.rawTextPrefix} Resolved Raw`;
  const manualRawText = `${options.rawTextPrefix} Manual Raw`;
  const autoRawText = `${options.rawTextPrefix} Auto Alias`;
  const dismissedRawText = `${options.rawTextPrefix} Dismissed Raw`;
  const createTimelineMention = async (
    suffix: string,
    occurredAt: string,
    summary: string,
    refs?: unknown,
  ) =>
    (await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-${suffix}-row`),
      "timeline.activity_utc_text": occurredAt,
      "timeline.activity_synopsis_text": summary,
      ...(refs === undefined ? {} : { [hostRefsFieldKey]: refs }),
    })) as ViewRow;
  const unresolvedRow = await createTimelineMention(
    "unresolved",
    options.occurredAt.unresolved,
    options.summary.unresolved,
    collectionActionsPayload([unresolvedRawText]),
  );
  const resolvedRow = await createTimelineMention(
    "resolved",
    options.occurredAt.resolved,
    options.summary.resolved,
    resolvedRefPayload(resolvedRawText, resolvedTarget.record_id),
  );
  const manualRow = await createTimelineMention(
    "manual",
    options.occurredAt.manual,
    options.summary.manual,
    collectionActionsPayload([manualRawText]),
  );
  const autoRow = await createTimelineMention(
    "auto",
    options.occurredAt.auto,
    options.summary.auto,
  );
  const dismissedRow = await createTimelineMention(
    "dismissed",
    options.occurredAt.dismissed,
    options.summary.dismissed,
    collectionActionsPayload([dismissedRawText]),
  );
  return {
    autoRawText,
    autoRow,
    dismissedMention: requireItemByRawText(
      collectionItems(dismissedRow, hostRefsFieldKey),
      dismissedRawText,
    ),
    dismissedRawText,
    dismissedRow,
    manualMention: requireItemByRawText(
      collectionItems(manualRow, hostRefsFieldKey),
      manualRawText,
    ),
    manualRawText,
    manualRow,
    manualTarget,
    resolvedMention: requireItemByRawText(
      collectionItems(resolvedRow, hostRefsFieldKey),
      resolvedRawText,
    ),
    resolvedRawText,
    resolvedRow,
    resolvedTarget,
    unresolvedMention: requireItemByRawText(
      collectionItems(unresolvedRow, hostRefsFieldKey),
      unresolvedRawText,
    ),
    unresolvedRawText,
    unresolvedRow,
  };
}
