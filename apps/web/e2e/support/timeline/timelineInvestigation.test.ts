import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import {
  seedTimelineInvestigation,
  timelineExpectations,
  timelineRecipe,
  timelineRows,
  validateTimelineRecipe,
} from "../../../../../tools/harness/fixtures/timeline-investigation/index.mjs";
import {
  requireTimelineDataProfile,
  timelineDataProfiles,
} from "../visual/timelineDataProfiles";
import { timelineScenarioFields } from "./timelineScenarioFields";

const requiredRow = (recipe: typeof timelineRecipe, index: number) => {
  const row = recipe.rows[index];
  if (!row) throw new Error("missing authored test row");
  return row;
};
const unexpectedOperation = async () => {
  throw new Error("unexpected owner operation");
};

describe("representative Timeline recipe", () => {
  it("covers all 24 fields and shares the authored core across both continuation sizes", () => {
    validateTimelineRecipe();
    expect(Object.values(timelineExpectations.fields).flat()).toHaveLength(24);
    const visual = timelineRows(timelineRecipe, 36);
    const review = timelineRows(timelineRecipe, 54);
    expect(visual).toHaveLength(48);
    expect(review).toHaveLength(66);
    expect(review.slice(0, 48)).toEqual(visual);
    expect(
      requiredRow({ ...timelineRecipe, rows: visual }, 0).fields[
        "timeline.raw_activity_text"
      ],
    ).toContain("123e4567-e89b-42d3-a456-426614174000");
    expect(
      requiredRow({ ...timelineRecipe, rows: visual }, 0).fields[
        "timeline.activity_utc_text"
      ],
    ).not.toBe(
      requiredRow({ ...timelineRecipe, rows: visual }, 1).fields[
        "timeline.activity_utc_text"
      ],
    );
  });
  it("rejects unsupported data before acquiring an incident", async () => {
    const corruptions = [
      (r: typeof timelineRecipe) => {
        r.schema_id = "unsupported";
      },
      (r: typeof timelineRecipe) => {
        requiredRow(r, 1).key = requiredRow(r, 0).key;
      },
      (r: typeof timelineRecipe) => {
        requiredRow(r, 0).fields["timeline.capture_state"] = "reviewed";
      },
      (r: typeof timelineRecipe) => {
        Object.assign(r, { credentials: "forbidden" });
      },
      (r: typeof timelineRecipe) => {
        Object.assign(r.relationships[1] ?? {}, { target: "missing" });
      },
      (r: typeof timelineRecipe) => {
        Object.assign(r.operations[0] ?? {}, { kind: "write_projection" });
      },
      (r: typeof timelineRecipe) => {
        Object.assign(r.operations[0] ?? {}, { row: "missing" });
      },
      (r: typeof timelineRecipe) => {
        r.operations.pop();
        r.operations.push({ kind: "relationship", relationship: 0 });
      },
      (r: typeof timelineRecipe) => {
        requiredRow(r, 0).fields["timeline.analyst_text"] = "";
      },
    ];
    let mutations = 0;
    for (const corrupt of corruptions) {
      const recipe = structuredClone(timelineRecipe);
      corrupt(recipe);
      await expect(
        seedTimelineInvestigation({
          recipe,
          continuationCount: 36,
          incident: {},
          port: {
            createIncident: async () => {
              mutations++;
              return "isolated";
            },
            call: unexpectedOperation,
            uploadEvidence: unexpectedOperation,
          },
        }),
      ).rejects.toThrow();
    }
    expect(mutations).toBe(0);
  });
  it("does not retry or substitute success after an owner operation fails", async () => {
    let calls = 0;
    await expect(
      seedTimelineInvestigation({
        continuationCount: 36,
        incident: {},
        port: {
          createIncident: async () => "isolated",
          uploadEvidence: unexpectedOperation,
          call: async () => {
            calls++;
            throw new Error("owner operation rejected");
          },
        },
      }),
    ).rejects.toThrow("owner operation rejected");
    expect(calls).toBe(1);
  });
});

describe("Timeline visual scenario admission", () => {
  it("composes complete mutable specimens without changing the shared core", () => {
    const before = JSON.stringify(timelineRecipe);
    const raw = "2026-04-18T14:12:34Z\n123e4567-e89b-42d3-a456-426614174000";
    const fields = timelineScenarioFields(
      {
        client_txn_id: "scenario",
        "timeline.raw_activity_text": raw,
        "timeline.activity_utc_text": "2025-02-17T11:05:00Z",
      },
      0,
    );
    expect(
      timelineExpectations.fields.source.every(
        (key) => typeof fields[key] === "string" && String(fields[key]).trim(),
      ),
    ).toBe(true);
    expect(fields["timeline.raw_activity_text"]).toBe(raw);
    expect(fields["timeline.activity_local_text"]).toBe("2025-02-17 07:05:00");
    expect(fields["timeline.date_entered_text"]).toBe("2025-02-17");
    expect(JSON.stringify(timelineRecipe)).toBe(before);
    expect(timelineRows(timelineRecipe, 0)).toHaveLength(12);
  });
  it("rejects missing source specimens and owner-managed seed fields", () => {
    for (const value of [null, "", "   "]) {
      expect(() =>
        timelineScenarioFields(
          { client_txn_id: "bad", "timeline.analyst_text": value },
          0,
        ),
      ).toThrow("Incomplete Timeline visual source");
      expect(() =>
        timelineScenarioFields(
          {
            client_txn_id: "bad",
            "timeline.activity_utc_text": "2025-02-17T11:05:00Z",
            "timeline.activity_local_text": value,
          },
          0,
        ),
      ).toThrow("Incomplete Timeline visual source");
    }
    for (const field of [
      "timeline.capture_state",
      "timeline.unknown",
      "password",
    ]) {
      expect(() =>
        timelineScenarioFields(
          { client_txn_id: "bad", [field]: "invented" },
          0,
        ),
      ).toThrow("Unsupported Timeline scenario field");
    }
  });
  it("requires exact capture profiles with explicit empty and sparse exceptions", () => {
    expect(() =>
      requireTimelineDataProfile("undeclared-timeline-capture"),
    ).toThrow("Undeclared Timeline capture");
    const normalization = JSON.parse(
      readFileSync(
        new URL(
          "../../../../../tools/frontend_visual_normalization.json",
          import.meta.url,
        ),
        "utf8",
      ),
    );
    for (const [name, profile] of Object.entries(
      timelineDataProfiles.captures,
    )) {
      expect(profile.capture_id).toBe(normalization.bindings[name]?.capture_id);
      expect(profile.reason.trim()).not.toBe("");
    }
    expect(
      requireTimelineDataProfile("timeline-grid-timeline-default").kind,
    ).toBe("rich");
    expect(requireTimelineDataProfile("timeline-grid-grouped-grid").kind).toBe(
      "rich",
    );
    expect(
      requireTimelineDataProfile("timeline-mutation-empty-timeline-query").kind,
    ).toBe("empty");
    expect(requireTimelineDataProfile("workbook-inspector-details").kind).toBe(
      "sparse",
    );
  });
});
