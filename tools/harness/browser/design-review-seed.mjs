import { createHmac, randomBytes, randomUUID } from "node:crypto";
import { readFileSync, writeFileSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import { writeReviewSamples } from "./design-review-samples.mjs";

export function reviewTotp(secret, now = Date.now()) {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
  const bits = [...secret.replace(/=+$/u, "").toUpperCase()].map((char) => {
    const value = alphabet.indexOf(char);
    if (value < 0) throw new Error("Invalid review authenticator key");
    return value.toString(2).padStart(5, "0");
  }).join("");
  const bytes = Buffer.from((bits.match(/.{8}/gu) ?? []).map((byte) => Number.parseInt(byte, 2)));
  const counter = Buffer.alloc(8);
  counter.writeBigUInt64BE(BigInt(Math.floor(now / 30000)));
  const digest = createHmac("sha1", bytes).update(counter).digest();
  return String((digest.readUInt32BE(digest.at(-1) & 15) & 0x7fffffff) % 1000000).padStart(6, "0");
}

export async function seedDesignReview({ root, environment, privateDirectory, runRoot, profile, signal, verifySamples, registerSecret }) {
  const apiOrigin = environment.CARTULARY_WEB_E2E_API_ORIGIN;
  const publicOrigin = environment.CARTULARY_WEB_E2E_PUBLIC_ORIGIN;
  const specification = JSON.parse(readFileSync(path.join(root, "contracts/openapi/cartulary.openapi.yaml"), "utf8"));
  const operations = new Map(Object.entries(specification.paths).flatMap(([route, methods]) => Object.entries(methods).filter(([, value]) => value.operationId).map(([method, value]) => [value.operationId, { route, method }])));
  const cookies = new Map();
  async function call(operation, body, parameters = {}, extraHeaders = {}, expectedError = false) {
    signal?.throwIfAborted();
    const definition = operations.get(operation);
    if (!definition) throw new Error(`Unknown public review operation ${operation}`);
    const route = definition.route.replace(/\{([^}]+)\}/gu, (_, name) => {
      if (!parameters[name]) throw new Error(`Missing ${operation} parameter ${name}`);
      return encodeURIComponent(parameters[name]);
    });
    const response = await fetch(new URL(route, apiOrigin), {
      method: definition.method.toUpperCase(), signal: AbortSignal.any([...(signal ? [signal] : []), AbortSignal.timeout(30000)]),
      headers: { "Content-Type": "application/json", Cookie: [...cookies].map(([key, value]) => `${key}=${value}`).join("; "), ...(cookies.has("cartulary_csrf") ? { "X-CSRF-Token": cookies.get("cartulary_csrf") } : {}), ...extraHeaders },
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
    for (const header of response.headers.getSetCookie()) {
      const pair = header.split(";", 1)[0];
      const index = pair.indexOf("=");
      cookies.set(pair.slice(0, index), pair.slice(index + 1));
      registerSecret(pair.slice(index + 1));
    }
    const result = await response.json();
    if (!response.ok && !expectedError) throw new Error(`${operation}: HTTP ${response.status} (${result.error?.code ?? "invalid response"}; ${result.error?.details?.reason_code ?? "no reason"}; field=${result.error?.details?.field ?? "none"}; view=${parameters.view_schema_id ?? "none"})`);
    return expectedError ? result : result.data;
  }
  const txn = () => randomUUID();
  const bootstrap = JSON.parse(readFileSync(path.join(root, "configs/dev/bootstrap-admin.json"), "utf8"));
  const login = { username: bootstrap.email, password: bootstrap.initial_password };
  const challenge = await call("loginLocalUser", login, {}, {}, true);
  if (challenge.error?.code !== "mfa_setup_required") throw new Error("Fresh review admin must require authenticator enrollment");
  const bootstrapToken = challenge.error.details.bootstrap_token;
  registerSecret(bootstrapToken);
  const auth = { Authorization: `Bearer ${bootstrapToken}` };
  const enrollment = await call("beginTOTPEnrollment", { client_txn_id: txn() }, {}, auth);
  const secret = enrollment.totp_setup.secret_base32;
  registerSecret(secret);
  await call("completeTOTPEnrollment", { client_txn_id: txn(), enrollment_id: enrollment.enrollment_id, code: reviewTotp(secret) }, {}, auth);
  await call("loginLocalUser", { ...login, second_factor: { kind: "totp", assertion: { code: reviewTotp(secret) } } });
  const samples = writeReviewSamples(root, path.join(runRoot, "samples"));
  const views = JSON.parse(readFileSync(path.join(root, "contracts/view-schemas/index.json"), "utf8")).view_schemas;
  const view = (recordType) => {
    const artifactViews = { comm_log: "cartulary.view.comm_log.v1", handoff: "cartulary.view.handoff.v1", status_review: "cartulary.view.status_review.v1", lesson: "cartulary.view.lesson.v1" };
    const found = views.find((entry) => artifactViews[recordType]
      ? entry.view_schema_id === artifactViews[recordType]
      : entry.source_record_types.includes(recordType));
    if (!found) throw new Error(`No current view for ${recordType}`);
    return found.view_schema_id;
  };
  const populated = (await call("createIncident", { client_txn_id: txn(), incident_key: "BROWSER-REVIEW", title: "Browser review investigation" })).incident_id;
  const empty = (await call("createIncident", { client_txn_id: txn(), incident_key: "BROWSER-EMPTY", title: "Empty review investigation" })).incident_id;
  const accounts = [{ email: bootstrap.email, password: bootstrap.initial_password, role: "deployment admin and incident admin", totp_setup_key: secret }];
  for (const [name, role, incidents] of [["empty", "viewer", []], ["viewer", "viewer", [populated]], ["editor", "editor", [populated, empty]]]) {
    const password = `Review1!${randomBytes(18).toString("base64url")}`;
    registerSecret(password);
    const email = `review-${name}@example.test`;
    const user = await call("createDeploymentUser", { client_txn_id: txn(), email, display_name: `Review ${name}`, auth_kind: "local", initial_password: password, is_deployment_admin: false, mfa_required: false });
    for (const incident of incidents) await call("createIncidentMembership", { client_txn_id: txn(), email, role }, { incident_id: incident });
    accounts.push({ email, password, role, user_id: user.user_id, visible_incident_count: incidents.length });
  }
  writeFileSync(path.join(privateDirectory, "access.json"), `${JSON.stringify({ origin: publicOrigin, instructions: "Use separate browser profiles for concurrent roles. Add the admin's generated setup key to a TOTP authenticator. Credentials expire with this disposable environment.", accounts }, null, 2)}\n`, { mode: 0o600 });
  const row = async (recordType, fields) => (await call("createViewRow", { client_txn_id: txn(), ...fields }, { incident_id: populated, view_schema_id: view(recordType) })).row;
  const first = await row("timeline_event", { "timeline.activity_synopsis_text": "Review initial triage and supporting evidence", "timeline.activity_utc_text": "2026-04-10T10:00:00Z" });
  const revised = await call("patchRecord", { client_txn_id: txn(), view_schema_id: view("timeline_event"), base_row_version: first.row_version, changes: [{ field_key: "timeline.activity_synopsis_text", value: "Review initial triage — revised after evidence collection" }] }, { record_id: first.record_id });
  for (let index = 1; index <= 65; index++) await row("timeline_event", { "timeline.activity_synopsis_text": `Review activity ${String(index).padStart(2, "0")}: ${index % 3 === 0 ? "follow-up investigation" : "observed activity"}`, "timeline.activity_utc_text": new Date(Date.UTC(2026, 3, 10, 10, index)).toISOString() });
  await row("host", { "host.hostname": "review-workstation.example.test" });
  await row("party", { "party.display_name": "Response coordination team", "party.party_kind": "team" });
  await row("task_request", { "task.title": "Collect endpoint evidence", "task.task_kind": "follow_up" });
  await row("decision", { "decision.summary": "Isolate affected workstation", "decision.decision_type": "containment", "decision.rationale": "Synthetic review decision with retained reasoning." });
  for (const [type, fields] of [
    ["comm_log", { "comm_log.summary": "Shift briefing", "comm_log.comm_type": "briefing", "comm_log.audience": "Response team", "comm_log.channel_or_meeting": "Incident bridge", "comm_log.timestamp_utc": "2026-04-10T12:00:00Z" }],
    ["handoff", { "handoff.current_state_summary": "Evidence collection in progress", "handoff.incoming_owner_user_id": accounts.find((account) => account.role === "editor").user_id, "handoff.timestamp_utc": "2026-04-10T12:00:00Z" }],
    ["status_review", { "status_review.current_state_summary": "Containment under review", "status_review.timestamp_utc": "2026-04-10T12:00:00Z" }],
    ["lesson", { "lesson.summary": "Keep investigation context beside the grid", "lesson.timestamp_utc": "2026-04-10T12:00:00Z" }],
  ]) await row(type, { ...fields, "coordination.source_record_id": first.record_id });
  const evidence = await row("evidence", { "evidence.title": "Synthetic acquisition notes", "evidence.collector_party_text": "Response coordination team", "evidence.requested_at": "2026-04-10T10:00:00Z" });
  const bytes = readFileSync(path.join(samples, "evidence.txt"));
  const blob = await call("createObjectBlobSlot", { client_txn_id: txn(), incident_id: populated, byte_size: bytes.length, filename_hint: "evidence.txt", content_type_hint: "text/plain" });
  const upload = blob.upload_target;
  if (upload.method !== "PUT" || !/^\/api\/v1\/object-uploads\/[^/?#]+$/u.test(upload.href)) throw new Error("Unexpected review upload target");
  const uploaded = await fetch(new URL(upload.href, apiOrigin), { method: "PUT", signal, body: bytes, headers: { ...upload.headers, Cookie: [...cookies].map(([key, value]) => `${key}=${value}`).join("; "), "X-CSRF-Token": cookies.get("cartulary_csrf") } });
  if (!uploaded.ok) throw new Error(`Review evidence upload failed: ${uploaded.status}`);
  const attached = await call("attachBlobToEvidenceRecord", { client_txn_id: txn(), base_row_version: evidence.row_version, object_blob_id: blob.object_blob_id }, { record_id: evidence.record_id });
  await call("patchRecord", { client_txn_id: txn(), view_schema_id: view("evidence"), base_row_version: attached.row.row_version, changes: [{ field_key: "evidence.lifecycle_state", value: "available" }] }, { record_id: evidence.record_id });
  await call("patchRecord", { client_txn_id: txn(), view_schema_id: view("timeline_event"), base_row_version: revised.row.row_version, changes: [{ field_key: "timeline.attached_evidence_ids", action_payload: { kind: "collection_actions_v1", actions: [{ op: "add_record_ref", linked_record_id: evidence.record_id }] } }] }, { record_id: first.record_id });
  await call("createIncidentSavedView", { display_name: "Shared review Timeline", scope: "shared", view_schema_id: view("timeline_event"), query_json: {}, layout_json: {} }, { incident_id: populated });

  // A real attached browser exercises the production surface while preparing
  // the sample. These screenshots are observations, never golden inputs.
  const require = createRequire(path.join(root, "apps/web/package.json"));
  const { chromium, expect } = require("@playwright/test");
  const browser = await chromium.launch({ headless: true });
  let page;
  const cancelBrowser = () => { void browser.close(); };
  signal?.addEventListener("abort", cancelBrowser, { once: true });
  try {
    const context = await browser.newContext({ viewport: { width: 1440, height: 900 } });
    await context.addCookies([...cookies].map(([name, value]) => ({ name, value, url: publicOrigin, httpOnly: name === "cartulary_session", sameSite: "Lax" })));
    page = await context.newPage();
    await page.goto(`${publicOrigin}/?incident_id=${populated}`);
    await expect(page.getByTestId("workbook-shell-ready")).toBeVisible({ timeout: 30000 });
    writeFileSync(path.join(runRoot, "review-workbook.png"), await page.screenshot(), { mode: 0o600 });
    if (profile === "network_flow_claimed") {
      await page.getByTestId("network-flow-analysis-tab").click();
      await page.getByTestId("network-flow-analysis-import-input").setInputFiles(path.join(samples, "network-flow.csv"));
      await page.getByTestId("network-flow-analysis-mapping-display-name").fill("Review network flows");
      await page.getByTestId("network-flow-analysis-mapping-preview").click();
      await expect(page.getByTestId("network-flow-analysis-mapping-apply")).toBeEnabled({ timeout: 30000 });
      await page.getByTestId("network-flow-analysis-mapping-apply").click();
      await expect(page.getByRole("tab", { name: /Review network flows/u })).toBeVisible({ timeout: 30000 });
      writeFileSync(path.join(runRoot, "review-network-analysis.png"), await page.screenshot(), { mode: 0o600 });
    } else await expect(page.getByTestId("network-flow-analysis-tab")).toHaveCount(0);
    if (verifySamples) {
      const handle = await call("issueEvidenceDownloadHandle", {}, { record_id: evidence.record_id });
      const downloadURL = new URL(handle.href, publicOrigin);
      if (downloadURL.origin !== publicOrigin) throw new Error("Review download escaped application origin");
      const download = await page.request.get(downloadURL.href);
      expect(download.ok()).toBe(true);
      expect(await download.body()).toEqual(bytes);
      await page.goto(`${publicOrigin}/?incident_id=${populated}`);
      await expect(page.getByTestId("workbook-shell-ready")).toBeVisible();
      await page.getByLabel("Account and application navigation").click();
      await page.getByTestId("incident-controls-trigger").click();
      await page.getByTestId("incident-controls-menu-item-import-assistant").click();
      const assistant = page.getByTestId("workbook-import-assistant");
      await assistant.getByLabel("Source workbook").setInputFiles(path.join(samples, "timeline.csv"));
      await assistant.getByRole("button", { name: "Upload and discover" }).click();
      await expect(assistant.getByRole("button", { name: "Approve mapping and select" })).toBeEnabled({ timeout: 30000 });
      await assistant.getByRole("button", { name: "Approve mapping and select" }).click();
      await assistant.getByRole("button", { name: "Apply 1 selected unit" }).click();
      await expect(assistant.getByRole("status")).toContainText("Import completed", { timeout: 30000 });
      await page.goto(`${publicOrigin}/deployment-administration`);
      await page.getByTestId("landing-admin-menu-item-reference-packs").click();
      await page.getByLabel("Reference pack bundle").setInputFiles(path.join(samples, "reference-pack.tar"));
      await page.getByRole("button", { name: "Import", exact: true }).click();
      await expect(page.getByText("type_registry.browser_review", { exact: true })).toBeVisible({ timeout: 30000 });
      await page.getByTestId("landing-admin-menu-item-incident-import").click();
      await page.getByLabel("Incident bundle file").setInputFiles(path.join(samples, "incident.tar"));
      await page.getByRole("button", { name: "Start import", exact: true }).click();
      await expect(page.getByRole("button", { name: "Open imported incident", exact: true })).toBeEnabled({ timeout: 30000 });
      await page.getByRole("button", { name: "Open imported incident", exact: true }).click();
      await expect(page.getByTestId("workbook-shell-ready")).toBeVisible();
      for (const account of accounts.slice(1)) {
        cookies.clear();
        await call("loginLocalUser", { username: account.email, password: account.password });
        const listing = await call("listVisibleIncidents");
        expect(listing.incidents).toHaveLength(account.visible_incident_count);
        const roleContext = await browser.newContext();
        try {
          await roleContext.addCookies([...cookies].map(([name, value]) => ({ name, value, url: publicOrigin, httpOnly: name === "cartulary_session", sameSite: "Lax" })));
          const rolePage = await roleContext.newPage();
          await rolePage.goto(publicOrigin);
          await expect(rolePage.getByTestId("incident-landing")).toBeVisible({ timeout: 30000 });
          expect(new URL(rolePage.url()).search).toBe("");
          if (account.visible_incident_count > 0) {
            await rolePage.goto(`${publicOrigin}/?incident_id=${populated}`);
            await expect(rolePage.getByTestId("workbook-shell-ready")).toBeVisible();
            await rolePage.getByLabel("Account and application navigation").click();
            await expect(rolePage.getByTestId("current-incident-role")).toContainText(account.role);
            await expect(rolePage.getByRole("menuitem", { name: "Deployment administration", exact: true })).toHaveCount(0);
            await rolePage.getByTestId("incident-controls-trigger").click();
            const importEntry = rolePage.getByTestId("incident-controls-menu-item-import-assistant");
            await importEntry.click();
            const sourceInput = rolePage.getByTestId("workbook-import-assistant").getByLabel("Source workbook");
            if (account.role === "viewer") await expect(sourceInput).toBeDisabled();
            else await expect(sourceInput).toBeEnabled();
          }
        } finally { await roleContext.close(); }
      }
      writeFileSync(path.join(runRoot, "review-smoke.json"), `${JSON.stringify({ purpose: "review_setup_observations", status: "pass", profile, checks: ["v7_attachment", "workbook", "evidence_upload_download", "csv_import", "reference_pack_import", "incident_import_handoff", "zero_one_many_roles", ...(profile === "network_flow_claimed" ? ["network_flow_import"] : ["network_flow_omitted"])] }, null, 2)}\n`, { mode: 0o600 });
    }
  } catch (error) {
    if (page && !page.isClosed()) {
      try { writeFileSync(path.join(runRoot, "review-failure.png"), await page.screenshot(), { mode: 0o600 }); } catch { /* Preserve the original failure if capture is unavailable. */ }
    }
    throw error;
  } finally { signal?.removeEventListener("abort", cancelBrowser); await browser.close(); }
  return { samples, scenarios: [
    { name: "Populated workbook, Evidence, history, coordination and saved view", url: `${publicOrigin}/?incident_id=${populated}` },
    { name: "Empty workbook", url: `${publicOrigin}/?incident_id=${empty}` },
    { name: "Incident directory: zero/one/many with the supplied roles", url: publicOrigin },
    { name: "Deployment administration: users, reference packs, incident import and audit", url: `${publicOrigin}/deployment-administration` },
  ] };
}
