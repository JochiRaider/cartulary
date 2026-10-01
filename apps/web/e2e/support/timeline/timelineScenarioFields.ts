import type {
  CreateViewRowRequest,
  TimelineCreateRequest,
} from "@cartulary/protocol-ts/http";
import {
  timelineExpectations,
  timelineRecipe,
  timelineRows,
} from "../../../../../tools/harness/fixtures/timeline-investigation/index.mjs";

/** Explicit authored-data composition for additional mutable visual specimens only. */
export function timelineScenarioFields(
  payload: CreateViewRowRequest,
  index: number,
) {
  const source = payload as { client_txn_id: string } & Record<string, unknown>;
  const writable = new Set([
    ...timelineExpectations.fields.source,
    ...timelineExpectations.fields.collections,
  ]);
  for (const field of Object.keys(payload)) {
    if (field !== "client_txn_id" && !writable.has(field))
      throw new Error(`Unsupported Timeline scenario field: ${field}`);
  }
  const base = timelineRows(timelineRecipe, index + 1).at(-1);
  if (!base) throw new Error("Missing Timeline scenario source");
  const fields: { client_txn_id: string } & Record<string, unknown> = {
    ...base.fields,
    ...source,
  };
  // An authored UTC override needs its matching local/date strings in this fixed-offset fixture.
  // This is fixture authoring, never capture normalization or a product default.
  if (typeof source["timeline.activity_utc_text"] === "string") {
    const utc = source["timeline.activity_utc_text"];
    const milliseconds = Date.parse(utc);
    if (!Number.isFinite(milliseconds))
      throw new Error("Invalid authored scenario chronology");
    fields["timeline.activity_local_text"] = Object.hasOwn(
      payload,
      "timeline.activity_local_text",
    )
      ? source["timeline.activity_local_text"]
      : new Date(milliseconds - 4 * 60 * 60_000)
          .toISOString()
          .slice(0, 19)
          .replace("T", " ");
    fields["timeline.date_entered_text"] = Object.hasOwn(
      payload,
      "timeline.date_entered_text",
    )
      ? source["timeline.date_entered_text"]
      : utc.slice(0, 10);
  }
  for (const field of timelineExpectations.fields.source) {
    if (typeof fields[field] !== "string" || !(fields[field] as string).trim())
      throw new Error(`Incomplete Timeline visual source: ${field}`);
  }
  return fields as TimelineCreateRequest & Record<string, unknown>;
}
