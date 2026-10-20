import { expect, it } from "vitest";
import {
  importActorID,
  importedIncidentID,
  importJob,
  importJobID,
} from "../testing/incidentImportTestSupport";
import { captureImport } from "./api/incidentImportClient";
import {
  initialImportState,
  openableImport,
  transitionImport,
} from "./incidentImportState";

function admitted() {
  const attempt = captureImport(new File(["bytes"], "bundle.tar"), "txn");
  let state = transitionImport(initialImportState(), {
    type: "admission_started",
    attempt,
  });
  state = transitionImport(state, {
    type: "admission_finished",
    attempt,
    actorId: importActorID,
    outcome: { kind: "accepted", job: importJob() },
  });
  return { state, attempt };
}

it("pure import transitions preserve immutable snapshots independently of read activity", () => {
  const { state } = admitted();
  const ready = transitionImport(state, {
    type: "read_finished",
    id: importJobID,
    outcome: { kind: "observed", job: importJob("succeeded") },
    active: true,
  });
  Object.freeze(ready);
  Object.freeze(ready.jobs);
  Object.freeze(ready.jobs[importJobID]);
  const reading = transitionImport(ready, {
    type: "read_started",
    id: importJobID,
  });
  expect(ready.jobs[importJobID]?.reading).toBe(false);
  expect(reading.jobs[importJobID]?.reading).toBe(true);
  expect(reading.jobs[importJobID]?.job).toBe(ready.jobs[importJobID]?.job);
  const entry = reading.jobs[importJobID];
  if (!entry) throw new Error("Expected admitted job");
  expect(openableImport(entry)).toBe(importedIncidentID);
  const failed = transitionImport(reading, {
    type: "read_finished",
    id: importJobID,
    outcome: { kind: "failed", problem: "transport" },
    active: true,
  });
  expect(failed.jobs[importJobID]?.job).toBe(ready.jobs[importJobID]?.job);
  expect(failed.jobs[importJobID]).toMatchObject({
    reading: false,
    observation: { kind: "failed" },
  });
});

it("pure import transitions reject conflicting terminal targets without replacing the last validated result", () => {
  let { state } = admitted();
  const terminal = importJob("succeeded");
  state = transitionImport(state, {
    type: "read_finished",
    id: importJobID,
    outcome: { kind: "observed", job: terminal },
    active: true,
  });
  if (!terminal.result_summary) throw new Error("Expected terminal result");
  const changed = importJob("succeeded", {
    result_summary: {
      ...terminal.result_summary,
      resource_refs: [
        {
          kind: "incident",
          id: importJobID,
          route: `/api/v1/incidents/${importJobID}`,
        },
      ],
    },
  });
  const rejected = transitionImport(state, {
    type: "read_finished",
    id: importJobID,
    outcome: { kind: "observed", job: changed },
    active: true,
  });
  expect(rejected.jobs[importJobID]?.job).toBe(terminal);
  expect(rejected.jobs[importJobID]?.observation).toEqual({
    kind: "failed",
    problem: "contract",
  });
});

it("pure import transitions retain exact unresolved attempts and clear protected data on reset", () => {
  const first = captureImport(new File(["first"], "first.tar"), "first");
  const later = captureImport(new File(["later"], "later.tar"), "later");
  const pending = transitionImport(initialImportState(), {
    type: "admission_started",
    attempt: first,
  });
  expect(
    transitionImport(pending, {
      type: "admission_finished",
      attempt: later,
      actorId: importActorID,
      outcome: { kind: "accepted", job: importJob() },
    }),
  ).toBe(pending);
  const uncertain = transitionImport(pending, {
    type: "admission_finished",
    attempt: first,
    actorId: importActorID,
    outcome: { kind: "uncertain", problem: "transport" },
  });
  expect(uncertain.admission).toEqual({ kind: "uncertain", attempt: first });
  expect(
    transitionImport(uncertain, {
      type: "file",
      file: new File(["other"], "other.tar"),
    }),
  ).toBe(uncertain);
  expect(transitionImport(admitted().state, { type: "reset" })).toEqual(
    initialImportState(),
  );
});

it("retains server-established unavailability through later read failures until a validated observation succeeds", () => {
  let { state } = admitted();
  for (const outcome of [
    { kind: "observed", job: importJob("succeeded") },
    { kind: "unavailable" },
    { kind: "failed", problem: "transport" },
  ] as const)
    state = transitionImport(state, {
      type: "read_finished",
      id: importJobID,
      outcome,
      active: true,
    });
  const unavailable = state.jobs[importJobID];
  if (!unavailable) throw new Error("Expected retained job");
  expect(openableImport(unavailable)).toBeNull();
  const recovered = transitionImport(state, {
    type: "read_finished",
    id: importJobID,
    outcome: { kind: "observed", job: importJob("succeeded") },
    active: true,
  });
  const entry = recovered.jobs[importJobID];
  if (!entry) throw new Error("Expected recovered job");
  expect(openableImport(entry)).toBe(importedIncidentID);
});
