import {
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { expect, it } from "vitest";
import { taskAuthority, taskRow } from "../../testing/taskWorkbookTestSupport";
import { prepareWorkbookInspectorChange } from "./prepareWorkbookInspectorChange";
import { WorkbookInspectorDraftStore } from "./WorkbookInspectorDraftStore";

const identity = {
  viewSchemaId: "cartulary.view.task_requests.v1",
  recordId: taskRow().record_id,
  fieldKey: "task.title",
  action: "value",
};
function fixture() {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  return store;
}
it("retains original raw intent and review dependencies independently from saved row objects", () => {
  const store = fixture(),
    row = taskRow();
  store.update(identity, row, "  unfinished\n", "first");
  const replaced = structuredClone(row);
  replaced.row_version++;
  replaced.cells["task.status"] = { value: "blocked" };
  expect(store.staleFields(identity, replaced, ["task.title"])).toEqual([]);
  replaced.cells["task.title"] = { value: "Concurrent title" };
  expect(store.staleFields(identity, replaced, ["task.title"])).toEqual([
    "task.title",
  ]);
  expect(store.read(identity)?.baseline.row_version).toBe(7);
  expect(store.read(identity)?.value).toBe("  unfinished\n");
  store.review(identity, replaced, "task.title", true);
  expect(store.staleFields(identity, replaced, ["task.title"])).toEqual([]);
  expect(store.read(identity)?.value).toBe("  unfinished\n");
});
it("requires explicit resumption and never lets another record or attachment consume raw work", () => {
  const store = fixture(),
    row = taskRow();
  store.update(identity, row, "first", "first");
  store.detach("first");
  store.update(identity, row, "accidental", "second");
  expect(store.read(identity)?.value).toBe("first");
  store.resume(identity, "second");
  store.update(
    identity,
    { ...row, record_id: "different" },
    "wrong target",
    "second",
  );
  expect(store.read(identity)?.value).toBe("first");
  store.update(identity, row, null, "second");
  expect(store.read(identity)?.value).toBeNull();
});
it("clears only the dispatched authoring revision and preserves later typing", () => {
  const store = fixture(),
    row = taskRow();
  store.update(identity, row, "sent", "form");
  const captured = store.capture(identity);
  store.update(identity, row, "newer", "form");
  store.acknowledge(captured);
  expect(store.read(identity)?.value).toBe("newer");
  store.acknowledge(store.capture(identity));
  expect(store.read(identity)).toBeNull();
});
it("conceals suspended drafts and retires them across incident or account replacement", () => {
  for (const replacement of [
    { ...taskAuthority, actorId: "other" },
    { ...taskAuthority, incidentId: "other" },
  ]) {
    const store = fixture();
    store.update(identity, taskRow(), "protected", "form");
    store.setAuthority(null);
    expect(store.read(identity)).toBeNull();
    store.setAuthority(taskAuthority);
    expect(store.read(identity)?.value).toBe("protected");
    expect(store.read(identity)?.attachment).toBeNull();
    store.setAuthority({ ...taskAuthority, closed: true });
    expect(store.canAuthor()).toBe(false);
    store.setAuthority(replacement);
    store.setAuthority(taskAuthority);
    expect(store.read(identity)).toBeNull();
  }
});
it("admits only the frozen existing-record capabilities while preserving create-only contracts", () => {
  const schemas = listViewContracts().filter(
    (contract) =>
      ![
        "cartulary.view.timeline.v2",
        "cartulary.view.assessments.v1",
        "cartulary.view.indicators.v1",
      ].includes(contract.viewSchemaId),
  );
  expect(
    schemas.flatMap((contract) =>
      contract.fields.filter(
        (field) => field.patchWritable && field.writeKind === "direct_value",
      ),
    ),
  ).toHaveLength(87);
  expect(
    schemas.flatMap((contract) =>
      contract.fields.filter(
        (field) => field.patchWritable && field.writeKind === "action_payload",
      ),
    ),
  ).toHaveLength(20);
  for (const contract of listViewContracts())
    for (const field of contract.fields)
      if (!field.patchWritable)
        expect(
          prepareWorkbookInspectorChange(
            field,
            "value",
            "add",
            contract.viewSchemaId,
          ).error,
        ).toBeDefined();
  const host = requireViewContract("cartulary.view.hosts.v1");
  expect(host.fieldMap["host.fqdn"]?.createWritable).toBe(true);
  expect(host.fieldMap["host.fqdn"]?.patchWritable).toBe(false);
});
it("validates scalar values and explicit clearing without mutating rough authoring", () => {
  const host = requireViewContract("cartulary.view.hosts.v1"),
    field = host.fieldMap["host.location"];
  if (!field) throw new Error("Missing location contract");
  expect(
    prepareWorkbookInspectorChange(field, null, "add", host.viewSchemaId)
      .change,
  ).toEqual({ field_key: field.fieldKey, value: null });
  expect(
    prepareWorkbookInspectorChange(
      field,
      "  Office  ",
      "add",
      host.viewSchemaId,
    ).change,
  ).toEqual({ field_key: field.fieldKey, value: "Office" });
  const timestamp = requireViewContract("cartulary.view.task_requests.v1")
    .fieldMap["task.completed_at"];
  if (!timestamp) throw new Error("Missing timestamp contract");
  expect(
    prepareWorkbookInspectorChange(
      timestamp,
      "2026-02-31T12:00:00Z",
      "add",
      identity.viewSchemaId,
    ).error,
  ).toBeDefined();
});
