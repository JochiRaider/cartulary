import path from "node:path";
import { readCanonicalUnitEvents } from "../evidence-accounting/index.mjs";
import { validateCanonicalRun } from "./canonical-evidence.mjs";

// Canonical lease events provide one interval per selected unit/lease. Shared
// resources and lease ownership are not inferred from overlapping durations.
export async function fixtureExplanation(root, target = "") {
  const run = await validateCanonicalRun(root);
  const selected = target ? run.targetSummaries.get(target)?.unit_ids : [...run.registrations.keys()];
  if (!selected) throw new Error("target is absent from the retained invocation");
  const units = new Set(selected), leases = new Map();
  for await (const event of readCanonicalUnitEvents(path.join(root, "unit-events.ndjson"))) {
    if (!units.has(event.unit_id) || !["fixture_acquired", "fixture_released"].includes(event.event)) continue;
    const key = JSON.stringify([event.unit_id, event.fixture_lease_id]);
    if (event.event === "fixture_acquired") {
      if (leases.has(key)) throw new Error("duplicate fixture acquisition");
      leases.set(key, { unit_id: event.unit_id, lease_id: event.fixture_lease_id, acquired_ms: event.monotonic_ms, held_ms: null, released: false });
    } else {
      const lease = leases.get(key);
      if (!lease || lease.released || event.monotonic_ms < lease.acquired_ms) throw new Error("unmatched fixture release");
      lease.held_ms = event.monotonic_ms - lease.acquired_ms;
      lease.released = true;
    }
  }
  return { run_id: run.manifest.run_id, target: target || run.manifest.target, leases: [...leases.values()] };
}
