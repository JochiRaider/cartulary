import { Buffer } from "node:buffer";
import {
  authTestId,
  incidentAdministrationTestId,
  incidentLandingTestId,
  networkAnalysisTestId,
  saveStateTestId,
  surfaceTabTestId,
  workbookPreferenceTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";
import { openIncidentControls } from "./pages/deploymentAdministration";
import { revokeAllSessions } from "./support/auth/sessions";
import {
  createTimelineRow,
  editTimelineSummary,
  expectServerSummaries,
  installPatchController,
  openIncidentAsTrackedUserReady,
} from "./support/collaboration/replay";
import {
  expectNetworkFlowRuntimeProfile,
  importNetworkFlowCSV,
  openClaimedNetworkAnalysis,
  openNetworkFlowIncident,
} from "./support/extensions/network_flow_activity/workspace";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { responseBarrier } from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";
import { uniqueEmail } from "./support/runtime/fixtureIdentity";
import {
  awaitResponseBarrier,
  installOwnedRoute,
} from "./support/transport/requestInterception";

test("Verify extension availability bootstrap, no-store startup, lazy Network Analysis loading, and Base client identity continuity.", async ({
  page,
}) => {
  expectNetworkFlowRuntimeProfile("network_flow_claimed");
  const networkFlowChunks: string[] = [];
  page.on("request", (request) => {
    if (request.url().includes("/assets/NetworkFlowFeature-")) {
      networkFlowChunks.push(request.url());
    }
  });
  const startupResponsePromise = page.waitForResponse((response) =>
    response.url().includes("/workbook-startup"),
  );
  const incidentId = await openNetworkFlowIncident(page, "EXTAVAILABILITY");
  const startupResponse = await startupResponsePromise;

  expect(startupResponse.headers()["cache-control"]).toBe("no-store");
  const startup = (await startupResponse.json()) as {
    data: {
      extension_workspace_availability: {
        incident_id: string;
        schema_id: string;
        workspaces: Array<{
          extension_profile_id: string;
          workspace_key: string;
        }>;
      };
    };
  };
  expect(startup.data.extension_workspace_availability).toEqual({
    schema_id: "cartulary.extension_workspace_availability.v1",
    incident_id: incidentId,
    workspaces: [
      {
        extension_profile_id: "network_flow_activity",
        workspace_key: "network_analysis",
      },
    ],
  });
  const packagedRootResponse = await page.request.get(`${apiBase}/`);
  expect(packagedRootResponse.ok()).toBe(true);
  const packagedRoot = await packagedRootResponse.text();
  const supportMatch = packagedRoot.match(
    /<script id="cartulary-client-extension-support-registry" type="application\/json">([^<]+)<\/script>/u,
  );
  expect(supportMatch).not.toBeNull();
  expect(JSON.parse(supportMatch?.[1] ?? "null")).toMatchObject({
    asset_set_sha256: expect.stringMatching(/^[0-9a-f]{64}$/u),
    client_build_class: "standard",
    profiles: [
      {
        profile_id: "import",
        supported_contract_majors: [1],
        workspace_keys: [],
        capability_ids: [],
      },
      {
        profile_id: "network_flow_activity",
        supported_contract_majors: [10],
        workspace_keys: ["network_analysis"],
        capability_ids: [],
      },
      {
        profile_id: "snapshot_reporting",
        supported_contract_majors: [1],
        workspace_keys: [],
        capability_ids: [],
      },
    ],
  });
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  expect(networkFlowChunks).toEqual([]);

  const clientInstanceID = await page.evaluate(() =>
    window.sessionStorage.getItem("cartulary.client_instance_id"),
  );
  expect(clientInstanceID).toMatch(/^[0-9a-f-]{36}$/u);
  const row = await createTimelineRow(
    page,
    incidentId,
    "Pending Base work across extension navigation",
  );
  await page.reload();
  const patchController = await installPatchController(page);
  const held = patchController.holdNextPatch({ recordId: row.record_id });
  try {
    await editTimelineSummary(page, row.record_id, "Retained Base work", {
      outcome: "queued",
    });
    await held.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    await page.getByRole("button", { name: "More views", exact: true }).click();
    await page.getByTestId(networkAnalysisTestId("tab")).click();
    await expect(
      page.getByTestId(networkAnalysisTestId("workspace")),
    ).toBeVisible();
    expect(networkFlowChunks).toHaveLength(1);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

    await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
    await expect(
      page.getByTestId(networkAnalysisTestId("workspace")),
    ).toHaveCount(0);
    expect(
      await page.evaluate(() =>
        window.sessionStorage.getItem("cartulary.client_instance_id"),
      ),
    ).toBe(clientInstanceID);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    held.release();
    await held.waitForCompletion;
    try {
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    } catch (error) {
      await test.info().attach("save-settlement-timing.json", {
        body: JSON.stringify(
          await page.evaluate(() =>
            performance
              .getEntriesByType("mark")
              .filter((entry) =>
                entry.name.startsWith("cartulary.workbook.pending_"),
              )
              .map((entry) => ({
                ...entry.toJSON(),
                detail: (entry as PerformanceMark).detail,
              })),
          ),
        ),
        contentType: "application/json",
      });
      throw error;
    }

    held.release();

    await patchController.dispose();
  } finally {
    held.cancel();
    await held.waitForCompletion;
  }
});

test("Verify authorization invalidation retires stale extension responses and preserves Base pending mutations.", async ({
  page: admin,
  testActorRequest,
  browserScope,
  sessionTracker,
}) => {
  expectNetworkFlowRuntimeProfile("network_flow_claimed");
  const incidentId = await openClaimedNetworkAnalysis(admin, "EXTAUTHORITY");
  await importNetworkFlowCSV(admin, {
    displayName: "Protected extension source",
    file: {
      name: "authority.csv",
      mimeType: "text/csv",
      buffer: Buffer.from(
        "Start Time,End Time,Source IP,Destination IP,Source Port,Destination Port,Protocol,Bytes,Packets\n" +
          "2026-07-10T00:00:00Z,2026-07-10T00:01:00Z,192.0.2.10,192.0.2.20,443,443,6,42,1\n",
      ),
    },
  });
  const row = await createTimelineRow(
    admin,
    incidentId,
    "Base authority witness",
  );
  const actor = await createIncidentMemberUser(admin, incidentId, {
    display_name: "Extension authority witness",
    email: uniqueEmail("extension-authority"),
    initial_password: "ExtensionAuthority123!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const { page, socketMonitor } = await openIncidentAsTrackedUserReady(
    browserScope,
    sessionTracker,
    {
      createdBy: "extensions authorization witness",
      email: actor.email,
      incidentId,
      password: actor.initial_password,
      purpose: "observe live extension authorization invalidation",
      readyRecordId: row.record_id,
      userId: actor.user_id,
    },
  );
  const clientInstanceId = await page.evaluate(() =>
    sessionStorage.getItem("cartulary.client_instance_id"),
  );
  expect(clientInstanceId).toMatch(/^[0-9a-f-]{36}$/u);
  const patches = await installPatchController(page);
  const pending = patches.holdNextPatch({ recordId: row.record_id });
  const graphGate = responseBarrier(page);
  const graphHit = responseBarrier(page);
  let captured = false;
  const oldGraph = await installOwnedRoute(
    page,
    /\/network-flow\/graphs\/query$/u,
    async (route) => {
      if (captured) {
        await route.fallback();
        return;
      }
      captured = true;
      // Keep a real, successfully authorized response from the old authority.
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      graphHit.release();
      if (!(await awaitResponseBarrier(route, graphGate.promise))) return;
      await route.fulfill({ response });
    },
    () => {
      graphGate.cancel();
      graphHit.cancel();
    },
  );
  const membership = `/api/v1/incidents/${incidentId}/memberships/${actor.user_id}`;
  try {
    await editTimelineSummary(
      page,
      row.record_id,
      "Base work survives invalidation",
      {
        outcome: "queued",
      },
    );
    const originalPatch = await pending.waitForHit;
    expect(originalPatch).not.toBeNull();
    await page.getByRole("button", { name: "More views", exact: true }).click();
    await page.getByTestId(networkAnalysisTestId("tab")).click();
    await expect(
      page.getByRole("tab", { name: /Protected extension source/u }),
    ).toBeVisible();
    await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
    expect(await graphHit.promise).toBe("released");

    // Revocation supplies an actual current-authorization invalidation. A
    // read-authorized role downgrade alone does not revoke the read session.
    await revokeAllSessions(
      testActorRequest,
      actor.user_id,
      "extension authority witness",
    );
    const revoked = await socketMonitor.waitForMessage("session_revoked", {
      timeoutMs: 25_000,
    });
    await socketMonitor.waitForClose(revoked.socketIndex, 25_000);
    await expect(page.getByTestId(authTestId("shell"))).toHaveAttribute(
      "data-bootstrap-state",
      "revoked",
    );
    await expect(
      page.getByTestId(networkAnalysisTestId("workspace")),
    ).toHaveCount(0);
    graphGate.release();
    await oldGraph.dispose();
    await expect(
      page.getByTestId(networkAnalysisTestId("workspace")),
    ).toHaveCount(0);
    pending.release();
    await pending.waitForCompletion;
    const messageStart = socketMonitor.messageCount();
    await sessionTracker.loginTrackedUser(page, {
      createdBy: "extensions authorization witness",
      email: actor.email,
      password: actor.initial_password,
      purpose: "resume Base work after real authorization invalidation",
      userId: actor.user_id,
    });
    await socketMonitor.waitForAcceptedSocket({ startAt: messageStart });
    expect(
      await page.evaluate(() =>
        sessionStorage.getItem("cartulary.client_instance_id"),
      ),
    ).toBe(clientInstanceId);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerSummaries(admin, incidentId, {
      [row.record_id]: "Base work survives invalidation",
    });
    const mutationIds = patches.calls
      .filter((call) => call.recordId === row.record_id)
      .map((call) => call.body.client_txn_id);
    expect(mutationIds.length).toBeGreaterThan(0);
    expect(new Set(mutationIds)).toEqual(
      new Set([originalPatch?.body.client_txn_id]),
    );
    const handshakes = socketMonitor
      .sentMessages()
      .filter(
        (message) => message.type === "hello" || message.type === "resume",
      );
    expect(handshakes.length).toBeGreaterThan(0);
    expect(
      handshakes.every(
        (message) => message.payload.client_instance_id === clientInstanceId,
      ),
    ).toBe(true);

    // Reopening requires new authorized availability, never the retired cache.
    // Core authorization recovery does not imply an automatic startup retry.
    const freshStartup = page.waitForResponse((response) =>
      response.url().includes("/workbook-startup"),
    );
    await page.reload();
    expect((await freshStartup).ok()).toBe(true);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    // Full incident-access loss follows Core's landing behavior and must also
    // remove all loaded extension content after this authorized reopen.
    await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
    await page.getByRole("button", { name: "More views", exact: true }).click();
    await page.getByTestId(networkAnalysisTestId("tab")).click();
    await expect(
      page.getByRole("tab", { name: /Protected extension source/u }),
    ).toBeVisible();
    expect(
      (
        await testActorRequest.delete(membership, {
          data: { base_membership_version: 1 },
        })
      ).status(),
    ).toBe(204);
    await expect(
      page.getByTestId(incidentLandingTestId("shell")),
    ).toBeVisible();
    await expect(
      page.getByTestId(networkAnalysisTestId("workspace")),
    ).toHaveCount(0);
    await expect(
      page.getByRole("tab", { name: /Protected extension source/u }),
    ).toHaveCount(0);
  } finally {
    pending.cancel();
    graphGate.cancel();
    graphHit.cancel();
    await oldGraph.dispose();
    await patches.dispose();
  }
});

test("preferences store and clear the claimed authorized extension workspace without changing active identity", async ({
  page,
}) => {
  const incidentId = await openClaimedNetworkAnalysis(page, "WP-EXTENSION");
  const original = page.url();
  const requests: unknown[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PUT" &&
      request.url().includes("workbook-preferences")
    )
      requests.push(request.postDataJSON());
  });
  await openIncidentControls(page);
  for (const kind of ["home", "default"] as const) {
    await expect(
      page.getByTestId(workbookPreferenceTestId(kind, "set")),
    ).toHaveAttribute("aria-disabled", "false");
    await page.getByTestId(workbookPreferenceTestId(kind, "set")).click();
    await expect(
      page.getByTestId(workbookPreferenceTestId(kind, "outcome")),
    ).toHaveText(
      `${kind === "home" ? "Home" : "Incident default"} update confirmed.`,
    );
    await expect(
      page.getByTestId(
        incidentAdministrationTestId(
          kind === "home" ? "pref-home-sheet-ref" : "pref-default-sheet-ref",
        ),
      ),
    ).toContainText("network_flow_activity/network_analysis");
    await expect(
      page.getByTestId(workbookPreferenceTestId(kind, "clear")),
    ).toHaveAttribute("aria-disabled", "false");
    await page.getByTestId(workbookPreferenceTestId(kind, "clear")).click();
    await expect(
      page.getByTestId(workbookPreferenceTestId(kind, "outcome")),
    ).toHaveText(
      `${kind === "home" ? "Home" : "Incident default"} clear confirmed.`,
    );
  }
  expect(requests).toEqual([
    {
      home_sheet_ref: {
        kind: "extension_workspace",
        extension_profile_id: "network_flow_activity",
        workspace_key: "network_analysis",
      },
    },
    { home_sheet_ref: null },
    {
      default_sheet_ref: {
        kind: "extension_workspace",
        extension_profile_id: "network_flow_activity",
        workspace_key: "network_analysis",
      },
    },
    { default_sheet_ref: null },
  ]);
  expect(page.url()).toBe(original);
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toBeVisible();
  expect(incidentId).not.toBe("");
});
