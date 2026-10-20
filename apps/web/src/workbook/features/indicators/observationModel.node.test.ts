import { expect, it } from "vitest";
import {
  observationSource,
  observationTargetId,
  testObservation,
} from "../../../testing/observationTestSupport";
import { ObservationDraftStore } from "./ObservationDraftStore";
import {
  observationCanTransition,
  observationOrder,
  observationSelection,
  observationSourceFields,
  observationTextMap,
  observationTimeKey,
} from "./observationModel";

it("Observation spans preserve ASCII Unicode repeated occurrences whitespace and line endings", () => {
  expect(observationTimeKey("2026-09-11T18:50:43.56884-04:00")).toBe(
    observationTimeKey("2026-09-11T22:50:43.568840000Z"),
  );
  expect(observationTimeKey("2026-02-30T18:50:43-04:00")).toBeNull();
  expect(
    observationOrder(
      { ...testObservation, created_at: "2026-09-11T18:50:43.568840001-04:00" },
      { ...testObservation, created_at: "2026-09-11T22:50:43.568840000Z" },
    ),
  ).toBe(true);
  const text = observationSource.text;
  expect(observationTextMap(text).display).toBe(
    "  α.example\nα.example 😀 e\u0301  ",
  );
  expect(observationSelection(text, 2, 11)).toEqual({
    startByte: 2,
    endByte: 12,
    text: "α.example",
  });
  expect(observationSelection(text, 12, 21)).toEqual({
    startByte: 14,
    endByte: 24,
    text: "α.example",
  });
  expect(observationSelection(text, 11, 12)).toEqual({
    startByte: 12,
    endByte: 14,
    text: "\r\n",
  });
  expect(observationSelection(text, 22, 24)).toEqual({
    startByte: 25,
    endByte: 29,
    text: "😀",
  });
  expect(observationSelection(text, 22, 23)).toBeNull();
  expect(observationSelection(text, 23, 24)).toBeNull();
  expect(observationSelection(text, 25, 27)).toEqual({
    startByte: 30,
    endByte: 33,
    text: "e\u0301",
  });
  expect(observationSelection("a\rb\r\nc\nd", 0, 7)).toEqual({
    startByte: 0,
    endByte: 8,
    text: "a\rb\r\nc\nd",
  });
  expect(observationSelection(" \t ", 0, 3)?.text).toBe(" \t ");
  for (const [start, end] of [
    [0, 0],
    [-1, 2],
    [2, 1],
    [0, 999],
    [0.5, 2],
  ])
    expect(observationSelection(text, start ?? 0, end ?? 0)).toBeNull();
});

it("Observation field discovery follows source text capability rather than editability", () => {
  const fields = observationSourceFields(observationSource.viewSchemaId, {
    record_id: observationSource.recordId,
    row_version: 4,
    cells: {
      "timeline.raw_activity_text": { value: "literal" },
      "timeline.capture_state": { value: "rough" },
      "timeline.replacement_record_id": { value: null },
      "timeline.tags": { value: "not scalar" },
      unknown: { value: "hidden" },
    },
  });
  expect(fields.map((field) => field.fieldKey)).toEqual([
    "timeline.raw_activity_text",
    "timeline.capture_state",
  ]);
});

it("Observation drafts isolate source fields and observation targets without content dedupe", () => {
  const store = new ObservationDraftStore();
  const a = store.ensure("source:one:field"),
    b = store.ensure("source:two:field");
  store.update(a.key, {
    source: observationSource,
    selection: observationSelection(observationSource.text, 2, 11),
    target: {
      recordId: observationTargetId,
      label: "Target",
      type: "domain_name",
    },
  });
  expect(store.get(b.key)).toBe(b);
  expect(store.ensure("observation:one").target).toBeNull();
  expect(store.ensure("observation:two").target).toBeNull();
});

it("Observation actions enforce resolve reassignment dismiss and unresolved restore", () => {
  const resolved = {
    ...testObservation,
    resolution_status: "resolved" as const,
    resolved_indicator_record_id: observationTargetId,
  };
  const dismissed = {
    ...testObservation,
    resolution_status: "dismissed" as const,
  };
  expect(
    observationCanTransition(testObservation, "resolve", observationTargetId),
  ).toBe(true);
  expect(
    observationCanTransition(resolved, "resolve", observationTargetId),
  ).toBe(false);
  expect(
    observationCanTransition(
      resolved,
      "resolve",
      "50000000-0000-4000-8000-000000000002",
    ),
  ).toBe(true);
  expect(observationCanTransition(resolved, "dismiss")).toBe(true);
  expect(observationCanTransition(testObservation, "restore")).toBe(false);
  expect(observationCanTransition(dismissed, "restore")).toBe(true);
  expect(observationCanTransition(dismissed, "dismiss")).toBe(false);
  expect(
    observationCanTransition(dismissed, "resolve", observationTargetId),
  ).toBe(false);
});
