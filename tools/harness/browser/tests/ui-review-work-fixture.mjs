import { randomBytes } from "node:crypto";
import path from "node:path";
import { ArtifactStore } from "../ui-review/artifact-store.mjs";
import { toolProfile } from "../ui-review/toolchain.mjs";
import { captureCapabilitySnapshot, resourceCapacities } from "../../scheduler/work-graph/capability.mjs";
import { inheritedHostLease } from "../../runtime/host-admission.mjs";
import { executeWork } from "../ui-review/executor.mjs";
import { pageSource } from "../ui-review/source.mjs";
import { execute as importCapture } from "../ui-review/capture.mjs";
import { repoRoot } from "../ui-review/policy.mjs";

export function work(session, command, request, operationID, options = {}) {
  return executeWork({ store: session.store, identity: { sessionID: session.sessionID, profile: session.profile }, signal: session.abort.signal,
    capacities: Object.fromEntries(resourceCapacities(captureCapabilitySnapshot({ root: repoRoot }))), parentLease: session.hostLease?.token ?? inheritedHostLease(),
    command, request, operationID, observe: (stage) => session.browser.capture(request, { identity: { sessionID: session.sessionID, profile: session.profile }, source: pageSource({ mode: session.mode, profile: session.input.REVIEW_PROFILE, workspaceDigest: session.workspaceDigest, runID: session.runID, prepared: session.seeded, browserVersion: session.browser.version() }, request.binding) }, operationID, stage), ...options });
}
export function capture(session, request, operationID) {
  return request.source === "page" ? work(session, "ui-capture", request, operationID) : importCapture(session, request, operationID);
}

export function sessionAt(root) {
  const session = { input: { UI_MODE: "artifacts" }, mode: "artifacts", profile: toolProfile(),
    sessionID: `uireview-${randomBytes(16).toString("hex")}`, abort: new AbortController(),
    workspaceDigest: "a".repeat(64), runtime: { privatePath: (...parts) => path.join(root, ...parts) } };
  session.store = new ArtifactStore({ sessionID: session.sessionID, profile: session.profile,
    signal: session.abort.signal, privatePath: session.runtime.privatePath });
  return session;
}
