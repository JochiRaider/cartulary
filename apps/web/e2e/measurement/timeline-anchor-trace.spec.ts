import type { ViewRow } from "@cartulary/protocol-ts/http";
import {
  applyFilterChip,
  changeGrouping,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  cartularyLargeGridFixture,
  cartularyAnchorTraceContract as contract,
  dataTestIdSelector,
  gridFieldCellSelector,
  gridRowTestId,
  gridRowVersionAttribute,
  gridScrollportSelector,
  gridShellTestId,
  rowCellTestId,
  saveStateTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import type { Route } from "@playwright/test";
import {
  type Ac043TrafficDriver,
  prepareAc043Snapshot,
  startAc043SnapshotTraffic,
} from "../support/performance/ac043Snapshot";
import { uniqueTxn } from "../support/runtime/fixtureIdentity";
import { holdBrowserRequest } from "../support/transport/requestInterception";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "../support/workbook/query";
import {
  type AnchorTraceProof,
  anchorCheckpoint,
  attachAnchorTrace,
  installAnchorTrace,
} from "./anchorTraceSupport";
import { expect, test } from "./fixtures";

test.use({
  viewport: { width: 1920, height: 1080 },
  deviceScaleFactor: 1,
  actionTimeout: 10000,
  navigationTimeout: 30000,
});
test.describe.configure({ mode: "serial", timeout: 10 * 60 * 1000 });

test("traces anchored Timeline edits through large-grid refresh and live updates within AC-047", async ({
  browser,
  page,
  testActor,
  sessionTracker,
}, info) => {
  const view = contract.view_schema_id;
  let traffic: Ac043TrafficDriver | null = null;
  let proof: AnchorTraceProof | null = null;
  let attached = false;
  let requestRecordId: string | null = null;
  let committedVersion: number | null = null;
  let finalVersion: number | null = null;
  try {
    expect(info.retry).toBe(0);
    const snapshot = await prepareAc043Snapshot(page, testActor);
    const firstPage = [...snapshot.timelineRows]
      .sort((a, b) =>
        a.record_id < b.record_id ? -1 : a.record_id > b.record_id ? 1 : 0,
      )
      .slice(0, 100);
    const target = firstPage.find(
      (row) => row.cells[contract.filter.field_key]?.value === false,
    );
    if (!target) throw new Error("Anchor fixture lacks an unlinked target");
    const excluded = new Set(firstPage.map((row) => row.record_id));
    const id = target.record_id;
    // Keep the target in the sorted first window so the stale response carries
    // its older version. Sorting an edited summary could remove it from that
    // window and would exercise absence instead of stale-row reconciliation.
    const prepared = await patchRecord(page, id, {
      view_schema_id: view,
      base_row_version: target.row_version,
      client_txn_id: uniqueTxn("anchor-sort-preparation"),
      changes: [
        { field_key: contract.sort_field_key, value: contract.sort_value },
      ],
    });
    target.row_version = prepared.row_version;

    await page.goto(`/?incident_id=${snapshot.incidentId}`);
    await expect(
      page.getByTestId(timelineMutationSubstrateReadyTestId()),
    ).toBeVisible();
    traffic = await startAc043SnapshotTraffic(
      browser,
      sessionTracker,
      snapshot.incidentId,
      snapshot.timelineRows.filter((row) => !excluded.has(row.record_id)),
      snapshot.runtime.backgroundAccounts,
      contract.predicate_id,
    );
    expect(traffic.qualifiedBackgroundSessions).toBe(24);
    expect(traffic.sessionCount).toBe(25);
    expect(traffic.updatesPerSecond).toBe(4.8);
    await expect(
      page.getByTestId(workbookPresenceSummaryTestId()),
    ).toHaveAttribute("aria-label", /^24 collaborators present on this sheet/);
    const content = page.getByTestId(rowCellTestId(id, contract.field_key));
    const row = page.getByTestId(gridRowTestId(view, id));
    const editorId = timelineScalarEditorTestId({
      fieldKey: contract.field_key,
      recordId: id,
      surface: "grid",
    });
    const editor = page.getByTestId(editorId);
    const expectTargetValue = async () => {
      await expect
        .poll(() =>
          page.evaluate(
            ({ editorSelector, contentSelector }) => {
              const input = document.querySelector<
                HTMLInputElement | HTMLTextAreaElement
              >(editorSelector);
              return (
                input?.value ??
                document.querySelector(contentSelector)?.textContent?.trim() ??
                null
              );
            },
            {
              editorSelector: dataTestIdSelector(editorId),
              contentSelector: dataTestIdSelector(
                rowCellTestId(id, contract.field_key),
              ),
            },
          ),
        )
        .toBe(contract.draft_value);
    };
    const shell = page.getByTestId(gridShellTestId(view));
    const scrollport = shell.locator(gridScrollportSelector());
    const frame = page.locator(`[data-workbook-view-schema-id="${view}"]`);
    await scrollGridCellIntoView({
      page,
      surface: view,
      recordId: id,
      cellKey: contract.field_key,
    });
    await content.click();
    await editor.fill(contract.draft_value);
    await page.evaluate(installAnchorTrace, {
      recordId: id,
      initialVersion: target.row_version,
      gridSelector: `[data-workbook-view-schema-id="${view}"] [role="grid"], [data-workbook-view-schema-id="${view}"] [role="treegrid"]`,
      editorSelector: `[role="row"][data-grid-record-id] ${gridFieldCellSelector(contract.field_key)} :is(input, textarea)`,
      draftValue: contract.draft_value,
      totalRows: cartularyLargeGridFixture.timelineRows,
    });
    await anchorCheckpoint(page, "draft_open");
    await scrollport.hover();
    await page.mouse.wheel(0, 100000);
    await expect
      .poll(() => scrollport.evaluate((node) => node.scrollTop))
      .toBeGreaterThan(0);
    await expect(editor).toHaveValue(contract.draft_value);
    await anchorCheckpoint(page, "scroll_away");
    await page.mouse.wheel(0, -100000);
    await expect
      .poll(() => scrollport.evaluate((node) => node.scrollTop))
      .toBe(0);
    await expect(editor).toBeFocused();
    await anchorCheckpoint(page, "scroll_return");

    const patchPath = `/api/v1/records/${id}`;
    const held = await holdBrowserRequest(page, {
      method: "PATCH",
      path: patchPath,
    });
    let releaseQuery = () => {};
    let captureQuery = () => {};
    let finishDelivery = () => {};
    const queryGate = new Promise<void>((resolve) => {
      releaseQuery = resolve;
    });
    const queryCaptured = new Promise<void>((resolve) => {
      captureQuery = resolve;
    });
    const queryDelivered = new Promise<void>((resolve) => {
      finishDelivery = resolve;
    });
    const queryPattern = `**/api/v1/incidents/${snapshot.incidentId}/views/${view}/query`;
    let captured = false;
    const staleQuery = async (route: Route) => {
      if (captured) {
        await route.continue();
        return;
      }
      captured = true;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      const retained = (await response.json()) as { data: { rows: ViewRow[] } };
      const staleTarget = retained.data.rows.find(
        (row) => row.record_id === id,
      );
      expect(staleTarget?.row_version).toBe(target.row_version);
      expect(staleTarget?.cells[contract.field_key]?.value).not.toBe(
        contract.draft_value,
      );
      captureQuery();
      await queryGate;
      await route.fulfill({ response });
      finishDelivery();
    };
    try {
      await page.evaluate(() => {
        if (!window.cartularyAnchorTrace)
          throw new Error("Missing anchor trace");
        window.cartularyAnchorTrace.draft = false;
      });
      const request = page.waitForRequest(
        (request) =>
          request.method() === "PATCH" && request.url().endsWith(patchPath),
      );
      await editor.press("Enter");
      await held.waitForHit;
      const acceptedRequest = await request;
      requestRecordId =
        new URL(acceptedRequest.url()).pathname.split("/").at(-1) ?? null;
      const body = acceptedRequest.postDataJSON();
      expect(body.view_schema_id).toBe(view);
      expect(body.base_row_version).toBe(target.row_version);
      expect(body.changes).toEqual([
        { field_key: contract.field_key, value: contract.draft_value },
      ]);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
      await anchorCheckpoint(page, "pending_commit");
      await sortByHeader(page, view, contract.sort_field_key);
      await expect(frame).toHaveAttribute(
        "data-workbook-query-pending",
        "false",
      );
      await expect(frame).toHaveAttribute(
        "data-workbook-query-sort-count",
        "1",
      );
      await anchorCheckpoint(page, "sort_accepted");
      await applyFilterChip(
        page,
        view,
        contract.filter.field_key,
        contract.filter.value,
      );
      await expect(frame).toHaveAttribute(
        "data-workbook-query-pending",
        "false",
      );
      await expect(frame).toHaveAttribute(
        "data-workbook-query-filter-count",
        "1",
      );
      await anchorCheckpoint(page, "filter_accepted");
      await changeGrouping(page, view, contract.group_field_key);
      await expect(frame).toHaveAttribute(
        "data-workbook-query-pending",
        "false",
      );
      await expect(frame).toHaveAttribute(
        "data-workbook-query-group-field",
        contract.group_field_key,
      );
      // Query controls reveal their own columns. Return to the edited field
      // through ordinary scrolling before asserting the visible commit.
      await scrollGridTargetIntoView({
        page,
        surface: view,
        targetTestId:
          (await editor.count()) > 0
            ? editorId
            : rowCellTestId(id, contract.field_key),
      });
      await anchorCheckpoint(page, "group_accepted");

      // Retain a real server response taken before commit, then deliver it after
      // the acknowledged mutation. No response body, auth or storage is mocked.
      await page.route(queryPattern, staleQuery);
      await createViewRow(page, snapshot.incidentId, view, {
        client_txn_id: uniqueTxn("anchor-invalidation"),
        [contract.field_key]: "AC047 invalidation row",
      });
      await queryCaptured;
      expect(held.hitCount()).toBe(1);
      held.release();
      const committedResponse = await acceptedRequest.response();
      if (!committedResponse?.ok())
        throw new Error("Anchor commit was not accepted");
      const committed = (await committedResponse.json()) as {
        data: { row: ViewRow };
      };
      expect(committed.data.row.record_id).toBe(id);
      expect(committed.data.row.row_version).toBe(target.row_version + 1);
      expect(committed.data.row.cells[contract.field_key]?.value).toBe(
        contract.draft_value,
      );
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      committedVersion = target.row_version + 1;
      await expect(row).toHaveAttribute(
        gridRowVersionAttribute,
        String(committedVersion),
      );
      await expectTargetValue();
      await anchorCheckpoint(page, "commit_accepted");
      releaseQuery();
      await queryDelivered;
      await expect(
        page.locator('[data-grid-data-state="refreshing"]'),
      ).toHaveCount(0);
      await expect(frame).toHaveAttribute(
        "data-workbook-query-pending",
        "false",
      );
      await expect(row).toHaveAttribute(
        gridRowVersionAttribute,
        String(committedVersion),
      );
      await expectTargetValue();
      await anchorCheckpoint(page, "stale_refresh_released");
      expect(held.hitCount()).toBe(1);
    } finally {
      releaseQuery();
      await held.dispose();
      await page.unroute(queryPattern, staleQuery);
    }
    if (committedVersion === null) throw new Error("Anchor commit is missing");
    const updated = await patchRecord(page, id, {
      view_schema_id: view,
      base_row_version: committedVersion,
      client_txn_id: uniqueTxn("anchor-live-update"),
      changes: [
        { field_key: contract.live_field_key, value: contract.live_value },
      ],
    });
    finalVersion = updated.row_version;
    await expect(row).toHaveAttribute(
      gridRowVersionAttribute,
      String(finalVersion),
    );
    await expectTargetValue();
    await expect(
      page.getByTestId(rowCellTestId(id, contract.live_field_key)),
    ).toHaveText(contract.live_value);
    await anchorCheckpoint(page, "live_update_applied");
    const current = (
      await queryViewRows(page, snapshot.incidentId, view, {
        sort: [{ field_key: contract.sort_field_key, direction: "asc" }],
      })
    ).find((row) => row.record_id === id);
    expect(current?.cells[contract.field_key]?.value).toBe(
      contract.draft_value,
    );
    expect(current?.row_version).toBe(finalVersion);
    await anchorCheckpoint(page, "final_stable");
    proof = await page.evaluate(() => {
      const trace = window.cartularyAnchorTrace;
      if (!trace) throw new Error("Missing trace");
      trace.stop();
      return trace.read();
    });
    proof.request_record_id = requestRecordId;
    proof.committed_row_version = committedVersion;
    proof.final_row_version = finalVersion;
    await traffic.stop();
    traffic = null;
    const observation = await attachAnchorTrace(info, proof);
    attached = true;
    expect(proof.violations).toEqual([]);
    expect(observation.outcome).toBe("passed");
  } finally {
    if (!attached) {
      if (!page.isClosed())
        proof = await page.evaluate(() => {
          const trace = window.cartularyAnchorTrace;
          trace?.stop();
          return trace?.read() ?? null;
        });
      if (proof) {
        proof.request_record_id = requestRecordId;
        proof.committed_row_version = committedVersion;
        proof.final_row_version = finalVersion;
      }
      await attachAnchorTrace(info, proof, "anchor_trace_incomplete");
    }
    await traffic?.stop();
  }
});
