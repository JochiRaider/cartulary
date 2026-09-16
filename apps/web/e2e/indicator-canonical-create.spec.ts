import type {
  CreateManualIndicatorObservationResponse,
  CreateViewRowResponse,
} from "@cartulary/protocol-ts/http";
import {
  indicatorCreateTestId,
  indicatorObservationTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  indicatorsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { csrfHeaders, loginLocalSession } from "./support/auth/browserSession";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncidentMembership } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import { uniqueEmail, uniqueTxn } from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { fetchFullRecordHistory } from "./support/workbook/history";
import {
  createCanonicalObservationFixture,
  openCanonicalProposal,
  submitCanonicalProposal,
} from "./support/workbook/indicatorCanonicalCreate";
import {
  listSourceObservations,
  observationServiceSnapshot,
  openObservationEditor,
} from "./support/workbook/indicatorObservations";
import { createViewRow } from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";

test("Canonical Indicator creation and resolution replay independently after response loss with two committed operations", async ({
  page,
}, testInfo) => {
  const fixture = await createCanonicalObservationFixture(page);
  const creates: string[] = [],
    resolves: string[] = [],
    receipts: CreateViewRowResponse["data"][] = [],
    links: CreateManualIndicatorObservationResponse["data"][] = [];
  await page.route(`**/views/${indicatorsViewSchemaId}/rows`, async (route) => {
    creates.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.status()).toBe(creates.length === 1 ? 201 : 200);
    receipts.push((await response.json()).data);
    if (creates.length === 1) await route.abort("failed");
    else await route.fulfill({ response });
  });
  await page.route(
    `**/indicator-observations/${fixture.observation.observation_id}/resolve`,
    async (route) => {
      resolves.push(route.request().postData() ?? "");
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      links.push((await response.json()).data);
      if (resolves.length === 1) await route.abort("failed");
      else await route.fulfill({ response });
    },
  );
  const before = await observationServiceSnapshot(page, fixture);
  await page.goto(`/?incident_id=${fixture.incidentId}`);
  const { form } = await openCanonicalProposal(page, fixture.source.record_id);
  await submitCanonicalProposal(form);
  await expect(
    page.getByText(/Canonical create outcome unknown/),
  ).toBeVisible();
  expect(resolves).toHaveLength(0);
  expect(
    await page
      .getByRole("button", { name: "Resolve observation to this Indicator" })
      .count(),
  ).toBe(0);
  const created = await observationServiceSnapshot(page, fixture);
  expect(created.observations).toEqual(before.observations);
  expect(created.source).toEqual(before.source);
  expect(created.indicators).toHaveLength(before.indicators.length + 1);
  const target = receipts[0]?.row;
  if (!target) throw new Error("Create receipt missing");
  const targetHistory = await fetchFullRecordHistory(page, target.record_id);
  expect(targetHistory.items).toHaveLength(1);
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await openRecoveryItem(page, /^Canonical Indicator creation ·/);
  const recovery = page.getByTestId(indicatorCreateTestId("recovery"));
  await recovery
    .getByRole("button", { name: "Replay original canonical create" })
    .click();
  await expect(
    recovery.getByText("Indicator available: new.example", { exact: true }),
  ).toBeVisible();
  await expect(
    recovery.getByText("Indicator and history refreshed.", { exact: true }),
  ).toBeVisible();
  expect(creates).toHaveLength(2);
  expect(creates[1]).toBe(creates[0]);
  expect(receipts[1]).toEqual(receipts[0]);
  expect(await observationServiceSnapshot(page, fixture)).toEqual(created);
  expect(await fetchFullRecordHistory(page, target.record_id)).toEqual(
    targetHistory,
  );
  expect(resolves).toHaveLength(0);
  await page.getByRole("button", { name: "Close recovery" }).click();
  const editor = await openObservationEditor(page, fixture.source.record_id);
  await expect(
    editor.getByText("Not linked to this Indicator.", { exact: true }),
  ).toBeVisible();
  await editor
    .getByRole("button", {
      name: "Resolve observation to this Indicator",
      exact: true,
    })
    .click();
  await expect(editor.getByText(/Link outcome unknown/)).toBeVisible();
  const linked = await observationServiceSnapshot(page, fixture);
  const linkedHistory = await fetchFullRecordHistory(page, target.record_id);
  expect(linked.observations[0]).toEqual({
    ...fixture.observation,
    ...links[0]?.observation,
  });
  expect(linked.observations[0]?.resolved_indicator_record_id).toBe(
    target.record_id,
  );
  expect(linked.observations[0]?.observed_text).toBe(
    fixture.observation.observed_text,
  );
  expect(linked.source?.cells).toEqual(before.source?.cells);
  expect(linkedHistory.items).toHaveLength(2);
  expect(links[0]?.change_set_id).not.toBe(receipts[0]?.change_set_id);
  await editor
    .getByRole("button", { name: "Replay original observation request" })
    .click();
  await expect(
    editor.getByText("Observation linked to this Indicator.", { exact: true }),
  ).toBeVisible();
  expect(resolves).toHaveLength(2);
  expect(resolves[1]).toBe(resolves[0]);
  expect(JSON.parse(resolves[0] ?? "{}").base_row_version).toBe(
    fixture.observation.row_version,
  );
  expect(links[1]).toEqual({ ...links[0], replayed: true });
  expect(creates).toHaveLength(2);
  expect(await observationServiceSnapshot(page, fixture)).toEqual(linked);
  expect(await fetchFullRecordHistory(page, target.record_id)).toEqual(
    linkedHistory,
  );
  await testInfo.attach("canonical-and-resolution-independent-receipts", {
    body: JSON.stringify({
      creates,
      resolves,
      receipts,
      links,
      sourceBefore: before.source,
      sourceAfter: linked.source,
    }),
    contentType: "application/json",
  });
});

test("Canonical reuse leaves metadata unchanged and retains its result after rejected resolution", async ({
  page,
}, testInfo) => {
  const fixture = await createCanonicalObservationFixture(page),
    before = await observationServiceSnapshot(page, fixture);
  let createCount = 0;
  await page.route(`**/views/${indicatorsViewSchemaId}/rows`, async (route) => {
    createCount++;
    const response = await route.fetch();
    expect(response.status()).toBe(201);
    await route.fulfill({ response });
  });
  await page.goto(`/?incident_id=${fixture.incidentId}`);
  const { form, editor } = await openCanonicalProposal(
    page,
    fixture.source.record_id,
    "ALPHA[.]EXAMPLE",
  );
  await form.getByText("Additional canonical details", { exact: true }).click();
  await form
    .getByLabel("Defanged presentation", { exact: true })
    .fill("must not enrich");
  await submitCanonicalProposal(form);
  await expect(
    editor.getByText("Indicator available: alpha.example", { exact: true }),
  ).toBeVisible();
  await expect(
    editor.getByText("Indicator and history refreshed.", { exact: true }),
  ).toBeVisible();
  const reused = await observationServiceSnapshot(page, fixture);
  expect(reused.indicators).toEqual(before.indicators);
  expect(reused.source).toEqual(before.source);
  expect(reused.observations).toEqual(before.observations);
  expect(reused.history[1]?.items).toHaveLength(
    (before.history[1]?.items.length ?? 0) + 1,
  );
  // A real stale child rejection after the owner's fresh read, before dispatch.
  let resolveCount = 0;
  await page.route(
    `**/indicator-observations/${fixture.observation.observation_id}/resolve`,
    async (route) => {
      resolveCount++;
      const dismissed = await publicHttpOperation({
        operationID: "dismissIndicatorObservation",
        pathParameters: { observation_id: fixture.observation.observation_id },
        headers: await csrfHeaders(page),
        request: atJsonOrigin(page.request, apiBase),
        body: {
          client_txn_id: uniqueTxn("concurrent-dismiss"),
          base_row_version: fixture.observation.row_version,
        },
      });
      expect(dismissed.ok).toBe(true);
      const response = await route.fetch();
      expect(response.status()).toBe(409);
      await route.fulfill({ response });
    },
  );
  await editor
    .getByRole("button", {
      name: "Resolve observation to this Indicator",
      exact: true,
    })
    .click();
  await expect(
    editor.getByText(
      "The observation change was not accepted. Your draft is retained.",
      { exact: true },
    ),
  ).toBeVisible();
  await expect(
    editor.getByText("Indicator available: alpha.example", { exact: true }),
  ).toBeVisible();
  expect(createCount).toBe(1);
  expect(resolveCount).toBe(1);
  const after = await listSourceObservations(page, fixture.source.record_id);
  expect(after.data.observations[0]?.resolution_status).toBe("dismissed");
  expect(after.data.observations[0]?.observed_text).toBe(
    fixture.observation.observed_text,
  );
  await testInfo.attach("canonical-reuse-stale-resolution", {
    body: JSON.stringify({ before, reused, after }),
    contentType: "application/json",
  });
});

test("Canonical late response preserves retargeted drafts and exact recovery after incident closure", async ({
  page,
}, testInfo) => {
  const fixture = await createCanonicalObservationFixture(page);
  const other = await createViewRow(
    page,
    fixture.incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("other-source"),
      "timeline.raw_activity_text": "alpha.example",
      "timeline.activity_synopsis_text": "Other source",
    },
  );
  const observation = await publicHttpOperation({
    operationID: "createManualIndicatorObservation",
    pathParameters: { source_record_id: other.record_id },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
    body: {
      client_txn_id: uniqueTxn("other-observation"),
      base_row_version: other.row_version,
      source_field_key: "timeline.raw_activity_text",
      span_start_byte: 0,
      span_end_byte: 13,
    },
  });
  expect(observation.ok).toBe(true);
  let release!: () => void;
  const held = new Promise<void>((resolve) => {
    release = resolve;
  });
  const bodies: string[] = [];
  let committed = false;
  await page.route(`**/views/${indicatorsViewSchemaId}/rows`, async (route) => {
    bodies.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    if (bodies.length === 1) {
      committed = true;
      await held;
      const payload = await response.json();
      await route.fulfill({
        response,
        json: {
          ...payload,
          data: { ...payload.data, change_set_id: "invalid-receipt" },
        },
      });
    } else await route.fulfill({ response });
  });
  await page.goto(`/?incident_id=${fixture.incidentId}`);
  const { form } = await openCanonicalProposal(page, fixture.source.record_id);
  await submitCanonicalProposal(form);
  await expect.poll(() => committed).toBe(true);
  await openObservationEditor(page, other.record_id);
  const editor = page.getByTestId(indicatorObservationTestId("editor"));
  await editor
    .getByRole("button", { name: "Create canonical Indicator…", exact: true })
    .click();
  const otherValue = editor.getByLabel("Canonical value", { exact: true });
  await otherValue.fill("other retained proposal");
  await otherValue.focus();
  const received = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      r.url().endsWith(`/views/${indicatorsViewSchemaId}/rows`),
  );
  release();
  await received;
  await expect(otherValue).toBeFocused();
  await expect(otherValue).toHaveValue("other retained proposal");
  await expect(
    editor.getByText(/Canonical create outcome unknown/),
  ).toHaveCount(0);
  await page.getByTestId(systemViewSwitcherTriggerTestId()).click();
  await page
    .getByTestId(
      systemViewSwitcherOptionTestId(
        "scope-indicators",
        indicatorsViewSchemaId,
      ),
    )
    .click();
  await page.getByRole("button", { name: "Timeline", exact: true }).click();
  await openObservationEditor(page, other.record_id);
  await editor
    .getByRole("button", { name: "Create canonical Indicator…", exact: true })
    .click();
  await expect(otherValue).toHaveValue("other retained proposal");
  const lifecycle = await currentLifecycle(page, fixture.incidentId);
  expect(
    (
      await lifecycleAction(page, fixture.incidentId, "closeIncident", {
        client_txn_id: uniqueTxn("close-after-canonical"),
        base_incident_version: lifecycle.incident_version,
        reason: "Retained canonical recovery",
      })
    ).ok,
  ).toBe(true);
  await expect(editor).toHaveCount(0);
  await openRecoveryItem(page, /^Canonical Indicator creation ·/);
  const recovery = page.getByTestId(indicatorCreateTestId("recovery"));
  await recovery
    .getByRole("button", { name: "Replay original canonical create" })
    .click();
  await expect(
    recovery.getByText("Indicator available: new.example", { exact: true }),
  ).toBeVisible();
  await expect(
    recovery.getByText("Indicator and history refreshed.", { exact: true }),
  ).toBeVisible();
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  expect(
    (await listSourceObservations(page, fixture.source.record_id)).data
      .observations[0]?.resolution_status,
  ).toBe("unresolved");
  expect(
    (await listSourceObservations(page, other.record_id)).data.observations[0]
      ?.resolution_status,
  ).toBe("unresolved");
  await testInfo.attach("canonical-retarget-closed-replay", {
    body: JSON.stringify({ bodies }),
    contentType: "application/json",
  });
});

test("Canonical results survive unavailable targets and current editor authority is checked before resolution", async ({
  page,
  workerAdminRequest,
  sessionTracker,
}, testInfo) => {
  const fixture = await createCanonicalObservationFixture(page),
    email = uniqueEmail("canonical-editor"),
    password = "CanonicalEditorPass!";
  const user = await createDeploymentUser(workerAdminRequest, {
    email,
    display_name: "Canonical editor",
    initial_password: password,
    is_deployment_admin: false,
    mfa_required: false,
  });
  await createIncidentMembership(page, fixture.incidentId, email, "editor");
  await loginLocalSession(page, email, password);
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "canonical observation workflow",
    email,
    purpose: "current editor authority",
    userId: user.user_id,
  });
  await page.goto(`/?incident_id=${fixture.incidentId}`);
  const { form, editor } = await openCanonicalProposal(
    page,
    fixture.source.record_id,
  );
  await submitCanonicalProposal(form);
  await expect(
    editor.getByText("Indicator and history refreshed.", { exact: true }),
  ).toBeVisible();
  const canonical = (
    await observationServiceSnapshot(page, fixture)
  ).indicators.find(
    (row) => row.cells["indicator.display_value"]?.value === "new.example",
  );
  if (!canonical) throw new Error("Canonical target missing");
  const mutateTarget = async (
    operationID: "deleteRecord" | "restoreRecord",
    version: number,
  ) =>
    publicHttpOperation({
      operationID,
      pathParameters: { record_id: canonical.record_id },
      request: workerAdminRequest,
      body: {
        client_txn_id: uniqueTxn(operationID),
        base_row_version: version,
      },
    });
  const removed = await mutateTarget("deleteRecord", canonical.row_version);
  expect(removed.ok).toBe(true);
  let resolveCount = 0;
  await page.route(
    `**/indicator-observations/${fixture.observation.observation_id}/resolve`,
    async (route) => {
      resolveCount++;
      const response = await workerAdminRequest.patch(
        `/api/v1/incidents/${fixture.incidentId}/memberships/${user.user_id}`,
        { data: { base_membership_version: 1, role: "viewer" } },
      );
      expect(response.ok()).toBe(true);
      const rejected = await route.fetch();
      expect(rejected.status()).toBe(403);
      await route.fulfill({ response: rejected });
    },
  );
  await editor
    .getByRole("button", {
      name: "Resolve observation to this Indicator",
      exact: true,
    })
    .click();
  await expect(
    editor.getByText(
      "The observation change was not accepted. Your draft is retained.",
      { exact: true },
    ),
  ).toBeVisible();
  await expect(
    editor.getByText("Indicator available: new.example", { exact: true }),
  ).toBeVisible();
  expect(resolveCount).toBe(0);
  const restored = await mutateTarget(
    "restoreRecord",
    canonical.row_version + 1,
  );
  expect(restored.ok).toBe(true);
  await editor
    .getByRole("button", {
      name: "Resolve observation to this Indicator",
      exact: true,
    })
    .click();
  await expect.poll(() => resolveCount).toBe(1);
  await expect(editor).toHaveCount(0);
  const observations = await publicHttpOperation({
    operationID: "listSourceRecordIndicatorObservations",
    pathParameters: { source_record_id: fixture.source.record_id },
    request: workerAdminRequest,
  });
  expect(observations.ok).toBe(true);
  if (observations.ok) {
    const actual = observations.payload.data.observations[0];
    expect(observations.payload.data.observations).toHaveLength(1);
    expect(actual).toEqual({
      ...fixture.observation,
      created_at: actual?.created_at,
    });
    expect(new Date(actual?.created_at ?? "").getTime()).toBe(
      new Date(fixture.observation.created_at).getTime(),
    );
  }
  const history = await publicHttpOperation({
    operationID: "getRecordHistory",
    pathParameters: { record_id: canonical.record_id },
    request: workerAdminRequest,
  });
  expect(history.ok).toBe(true);
  if (history.ok) expect(history.payload.data.items).toHaveLength(3);
  await testInfo.attach("canonical-unavailable-target-current-authority", {
    body: JSON.stringify({ canonical, observations, history }),
    contentType: "application/json",
  });
});
