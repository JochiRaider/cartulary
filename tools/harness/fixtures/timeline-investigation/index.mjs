import assert from "node:assert/strict";
import { randomUUID, createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import { validateSchemaSync } from "../../contract/index.mjs";

const read = (name) => JSON.parse(readFileSync(new URL(name, import.meta.url), "utf8"));
export const timelineRecipe = read("./recipe.json");
export const timelineExpectations = read("./expectations.json");
const view = read("../../../../contracts/view-schemas/cartulary.view.timeline.v2.json");
const value = (row, field) => row.cells[`timeline.${field}`]?.value;
const items = (row, field) => {
  const collection = row.cells[field]?.value;
  assert.ok(collection && Array.isArray(collection.items), `missing collection ${field}`);
  return collection.items;
};
const actions = (values) => ({ kind: "collection_actions_v1", actions: values });

export function validateTimelineRecipe(recipe = timelineRecipe, expectations = timelineExpectations) {
  validateSchemaSync("cartulary.timeline_investigation_recipe.v1", recipe);
  validateSchemaSync("cartulary.timeline_investigation_expectations.v1", expectations);
  assert.equal(recipe.dataset_id, expectations.dataset_id);
  assert.equal(recipe.content_revision, expectations.content_revision);
  const allKeys = [...recipe.rows, ...recipe.entities, ...recipe.evidence].map((entry) => entry.key);
  assert.equal(new Set(allKeys).size, allKeys.length, "duplicate semantic key");
  const rowKeys = new Set(recipe.rows.map((row) => row.key));
  assert.deepEqual(Object.keys(expectations.rows).sort(), [...rowKeys].sort(), "outcome keys");
  assert.deepEqual(expectations.fields.source, view.default_visible_fields, "source field contract");
  assert.deepEqual(Object.values(expectations.fields).flat().sort(), view.fields.map((field) => field.field_key).sort(), "24-field coverage");
  for (const [group, writeKind] of Object.entries({ source: "direct_value", collections: "action_payload", derived: "read_only" })) {
    assert.deepEqual([...expectations.fields[group]].sort(), view.fields.filter((field) => field.write_kind === writeKind).map((field) => field.field_key).sort(), `${group} field classification`);
  }
  for (const row of [...recipe.rows, { fields: recipe.continuation.source_fields }]) {
    assert.deepEqual(Object.keys(row.fields).sort(), [...view.default_visible_fields].sort(), "representative source fields");
    assert.ok(Object.values(row.fields).every((text) => typeof text === "string" && text.trim()), "meaningful representative values");
  }
  const times = recipe.rows.map((row) => Date.parse(row.fields["timeline.activity_utc_text"]));
  assert.ok(times.every((time, index) => Number.isFinite(time) && (index === 0 || time > times[index - 1])), "distinct ordered activity timestamps");
  assert.ok(Date.parse(recipe.continuation.start_utc) > times.at(-1), "continuations follow core");
  for (const entity of recipe.entities) {
    const prefix = entity.view_schema_id === "cartulary.view.hosts.v1" ? "host." : "identity.";
    assert.ok(Object.keys(entity.fields).every((key) => key.startsWith(prefix)), "entity field owner");
    if (entity.aliases) assert.equal(prefix, "host.", "alias owner");
  }
  for (const relationship of recipe.relationships) {
    assert.ok(rowKeys.has(relationship.row), "unknown relationship row");
    const resolved = ["resolved", "automatic"].includes(relationship.kind);
    assert.equal(relationship.target !== undefined, resolved, "relationship target requirement");
    if (resolved) {
      const target = recipe.entities.find((entity) => entity.key === relationship.target);
      assert.ok(target, "unknown relationship target");
      assert.equal(target.view_schema_id, relationship.field === "timeline.host_refs" ? "cartulary.view.hosts.v1" : "cartulary.view.identities.v1", "relationship target type");
    }
  }
  const relationshipOperations = [];
  for (const operation of recipe.operations) {
    if (operation.kind === "relationship") {
      assert.ok(recipe.relationships[operation.relationship], "unknown relationship operation");
      relationshipOperations.push(operation.relationship);
    } else {
      assert.ok(rowKeys.has(operation.row), "unknown operation row");
      if (operation.kind === "evidence") assert.ok(recipe.evidence.some((item) => item.key === operation.evidence), "unknown evidence key");
      if (operation.kind === "supersede") assert.ok(rowKeys.has(operation.replacement) && operation.row !== operation.replacement, "invalid replacement key");
    }
  }
  assert.deepEqual(relationshipOperations.sort((a, b) => a - b), recipe.relationships.map((_, i) => i), "each relationship executes exactly once");
  return recipe;
}

export function timelineRows(recipe, continuationCount) {
  validateTimelineRecipe(recipe);
  assert.ok(Number.isInteger(continuationCount) && continuationCount >= 0 && continuationCount <= recipe.continuation.maximum_rows, "invalid continuation count");
  const definition = recipe.continuation;
  const continuation = Array.from({ length: continuationCount }, (_, index) => {
    const timestamp = Date.parse(definition.start_utc) + index * definition.interval_minutes * 60_000;
    const sequence = String(index + 1).padStart(2, "0");
    return { key: `containment-verification-${sequence}`, fields: {
      ...definition.source_fields,
      "timeline.activity_utc_text": new Date(timestamp).toISOString().replace(".000Z", "Z"),
      "timeline.activity_local_text": new Date(timestamp + definition.local_offset_minutes * 60_000).toISOString().slice(0, 19).replace("T", " "),
      "timeline.activity_synopsis_text": definition.synopsis_template.replaceAll("{index}", sequence),
      "timeline.raw_activity_text": definition.raw_template.replaceAll("{index}", sequence),
    } };
  });
  return [...structuredClone(recipe.rows), ...continuation];
}

// This port is harness-only. The adapters own authenticated transport and cleanup;
// this finite recipe executor never owns a browser, credentials or a database.
export async function seedTimelineInvestigation({ port, incident, continuationCount, recipe = timelineRecipe }) {
  validateTimelineRecipe(recipe);
  const authored = timelineRows(recipe, continuationCount);
  const incidentId = await port.createIncident(incident);
  const mapping = new Map();
  const dismissed = [];
  const txn = () => randomUUID();
  const call = port.call;
  const query = async (schema = recipe.view_schema_id) => (await call("queryWorkbookView", { limit: 200 }, { incident_id: incidentId, view_schema_id: schema })).rows;
  const row = async (key) => {
    const result = (await query()).find((entry) => entry.record_id === mapping.get(key));
    assert.ok(result, `missing recipe row ${key}`);
    return result;
  };
  const patch = async (key, changes) => {
    const current = await row(key);
    return (await call("patchRecord", { client_txn_id: txn(), view_schema_id: recipe.view_schema_id, base_row_version: current.row_version, changes }, { record_id: current.record_id })).row;
  };
  const profile = await call("getTimelineTimeConversionProfile", undefined, { incident_id: incidentId });
  assert.equal(profile.enabled, false, "base investigation requires disabled conversion");
  for (const entity of recipe.entities) {
    const result = await call("createViewRow", { client_txn_id: txn(), ...entity.fields, ...(entity.aliases ? { "host.aliases": actions(entity.aliases.map((alias_text) => ({ op: "add_alias", alias_text }))) } : {}) }, { incident_id: incidentId, view_schema_id: entity.view_schema_id });
    mapping.set(entity.key, result.row.record_id);
  }
  for (const source of authored) {
    const result = await call("createViewRow", { client_txn_id: txn(), ...source.fields }, { incident_id: incidentId, view_schema_id: recipe.view_schema_id });
    mapping.set(source.key, result.row.record_id);
    assert.equal(value(result.row, "capture_state"), "rough", "populated creation must be rough");
    for (const [field, text] of Object.entries(source.fields)) assert.equal(result.row.cells[field]?.value, text, `creation source ${source.key}/${field}`);
  }
  for (const evidence of recipe.evidence) {
    const result = await port.uploadEvidence(incidentId, evidence);
    assert.equal(result.cells["evidence.upload_state"]?.value, "available", "evidence must be finalized");
    mapping.set(evidence.key, result.record_id);
  }
  for (const operation of recipe.operations) {
    switch (operation.kind) {
      case "tags": {
        const before = await row(operation.row);
        const after = await patch(operation.row, [{ field_key: "timeline.tags", action_payload: actions(operation.values.map((tag_name) => ({ op: "add_tag", tag_name }))) }]);
        assert.equal(value(after, "capture_state"), value(before, "capture_state"), "tag-only state");
        break;
      }
      case "patch":
        await patch(operation.row, Object.entries(operation.fields).map(([field_key, value]) => ({ field_key, value })));
        break;
      case "evidence":
        await patch(operation.row, [{ field_key: "timeline.attached_evidence_ids", action_payload: actions([{ op: "add_record_ref", linked_record_id: mapping.get(operation.evidence) }]) }]);
        break;
      case "relationship": {
        const relation = recipe.relationships[operation.relationship];
        const changed = await patch(relation.row, [{ field_key: relation.field, action_payload: actions([{ op: "add_token", raw_text: relation.text }]) }]);
        const mention = items(changed, relation.field).find((item) => item.raw_text === relation.text);
        assert.ok(mention, "created source-bound mention");
        if (["resolved", "dismissed"].includes(relation.kind)) {
          const resolution = await call("resolveEntityMention", { client_txn_id: txn(), base_mention_row_version: mention.mention_row_version, action: relation.kind === "resolved" ? "resolve_item" : "dismiss_item", ...(relation.target ? { resolved_record_id: mapping.get(relation.target) } : {}), reason: "Synthetic investigation source review" }, { entity_mention_id: mention.entity_mention_id });
          assert.equal(resolution.entity_mention.resolution_status, relation.kind);
          assert.equal(resolution.entity_mention.raw_text, relation.text);
          assert.equal(resolution.entity_mention.source_record_id, mapping.get(relation.row));
          assert.ok(resolution.entity_mention.row_version > mention.mention_row_version, "mention version advances");
          assert.ok(resolution.source_record.row_version > changed.row_version, "relationship source version advances");
          if (relation.kind === "dismissed") dismissed.push({ row: relation.row, field: relation.field, text: relation.text, id: mention.entity_mention_id });
        } else if (relation.kind === "automatic") {
          assert.equal(mention.item_kind, "resolved_ref", "automatic owner resolution");
          assert.equal(mention.auto_resolved, true, "automatic resolution provenance");
          assert.equal(mention.resolved_record_id, mapping.get(relation.target));
        }
        break;
      }
      case "review": {
        const current = await row(operation.row);
        const result = await call("markTimelineRecordReviewed", { client_txn_id: txn(), base_row_version: current.row_version }, { record_id: current.record_id });
        assert.equal(result.capture_state, "reviewed");
        break;
      }
      case "supersede": {
        const current = await row(operation.row);
        const result = await call("supersedeRecord", { client_txn_id: txn(), base_row_version: current.row_version, reason: operation.reason, replacement_record_id: mapping.get(operation.replacement) }, { record_id: current.record_id });
        assert.equal(result.capture_state, "superseded");
        break;
      }
      default: throw new Error("Unsupported recipe operation");
    }
  }
  const rows = await query();
  const outcomes = verifyTimelineInvestigation({ recipe, authored, rows, mapping });
  return { incidentId, mapping, recordId: (key) => { const id = mapping.get(key); assert.ok(id, `unknown semantic key ${key}`); return id; }, rows, dismissed, outcomes, readRows: query, verify: async () => verifyTimelineInvestigation({ recipe, authored, rows: await query(), mapping }) };
}

export function verifyTimelineInvestigation({ recipe = timelineRecipe, authored, rows, mapping }) {
  assert.equal(rows.length, authored.length, "Timeline row total");
  const semantic = {};
  const keysById = new Map([...mapping].map(([key, id]) => [id, key]));
  for (const source of authored) {
    const row = rows.find((entry) => entry.record_id === mapping.get(source.key));
    assert.ok(row, `missing ${source.key}`);
    assert.ok(Number.isInteger(row.row_version) && row.row_version > 0, "current row version");
    const expectedSource = { ...source.fields };
    for (const operation of recipe.operations) if (operation.kind === "patch" && operation.row === source.key) Object.assign(expectedSource, operation.fields);
    for (const [field, text] of Object.entries(expectedSource)) assert.equal(row.cells[field]?.value, text, `${source.key}/${field}`);
    const expected = timelineExpectations.rows[source.key] ?? { capture_state: "rough", evidence_count: 0, has_unresolved_mentions: false, replacement: null };
    for (const field of ["capture_state", "evidence_count", "has_unresolved_mentions"]) assert.equal(value(row, field), expected[field], `${source.key}/${field}`);
    assert.equal(value(row, "has_evidence"), expected.evidence_count > 0);
    assert.equal(value(row, "replacement_record_id"), expected.replacement === null ? null : mapping.get(expected.replacement));
    assert.equal(value(row, "activity_time_pair_state"), "disabled");
    assert.equal(Date.parse(value(row, "activity_sort_ts")), Date.parse(expectedSource["timeline.activity_utc_text"]));
    assert.equal(value(row, "date_entered_sort_day"), expectedSource["timeline.date_entered_text"]);
    for (const field of ["recorded_at", "edited_at"]) assert.ok(Number.isFinite(Date.parse(value(row, field))), `system ${field}`);
    assert.ok(Date.parse(value(row, "edited_at")) >= Date.parse(value(row, "recorded_at")), "metadata chronology");
    const relationships = {};
    for (const field of ["timeline.host_refs", "timeline.identity_refs"]) {
      const expectedItems = recipe.relationships.filter((item) => item.row === source.key && item.field === field && item.kind !== "dismissed");
      const actualItems = items(row, field);
      assert.equal(actualItems.length, expectedItems.length, `${source.key}/${field} count`);
      relationships[field] = actualItems.map((item, index) => {
        const expectedItem = expectedItems[index];
        assert.equal(item.raw_text, expectedItem.text);
        assert.equal(item.item_kind, expectedItem.kind === "unresolved" ? "unresolved_mention" : "resolved_ref");
        if (expectedItem.kind === "automatic") assert.equal(item.auto_resolved, true);
        if (expectedItem.kind === "resolved") assert.notEqual(item.auto_resolved, true);
        const target = item.resolved_record_id ? keysById.get(item.resolved_record_id) : null;
        assert.equal(target, expectedItem.target ?? null);
        return { text: item.raw_text, target, kind: item.item_kind };
      });
    }
    const evidence = recipe.operations.filter((op) => op.kind === "evidence" && op.row === source.key).map((op) => op.evidence).sort();
    const actualEvidence = items(row, "timeline.attached_evidence_ids").map((item) => keysById.get(item.linked_record_id)).sort();
    assert.deepEqual(actualEvidence, evidence, `${source.key}/attached evidence`);
    const tags = recipe.operations.filter((op) => op.kind === "tags" && op.row === source.key).flatMap((op) => op.values).sort();
    assert.deepEqual(items(row, "timeline.tags").map((item) => item.display_text).sort(), tags, `${source.key}/tags`);
    if (timelineExpectations.rows[source.key]) semantic[source.key] = { fields: expectedSource, relationships, evidence, tags, ...expected };
  }
  assert.deepEqual(rows.map((row) => row.record_id), authored.map((row) => mapping.get(row.key)), "default chronology");
  return { dataset_id: recipe.dataset_id, content_revision: recipe.content_revision, core_rows: recipe.rows.length, total_rows: rows.length, semantic_sha256: createHash("sha256").update(JSON.stringify(semantic)).digest("hex") };
}
