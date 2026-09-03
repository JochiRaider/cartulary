import { describe, expect, it } from "vitest";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import { entityRowFromApi } from "../models/entityWorkbookModel";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { planEntityLiveEventPatch } from "./entityLiveEventPatchPlanner";

function entityRow(
  recordId: string,
  rowVersion: number,
  entityType: "host" | "identity" = "host",
) {
  return entityRowFromApi(
    {
      cells: {
        [`${entityType}.display_name`]: { value: "Before" },
      },
      record_id: recordId,
      row_version: rowVersion,
    },
    entityType,
  );
}

function patchPayload(
  overrides: Partial<RecordChangedPayload> = {},
): RecordChangedPayload {
  return {
    actor_user_id: "user-2",
    affected_views: [
      {
        change_kind: "patch",
        patch_cells: {
          cells: { "host.display_name": { value: "After" } },
          record_id: "host-1",
          row_version: 2,
        },
        view_schema_id: hostsViewSchemaId,
      },
    ],
    change_set_id: "change-1",
    changed_field_keys: ["host.display_name"],
    client_txn_id: "txn-1",
    record_id: "host-1",
    row_version: 2,
    ...overrides,
  };
}

describe("Entity live-event patch planner", () => {
  it("applies an exact newer patch and preserves unrelated row references", () => {
    const target = entityRow("host-1", 1);
    const sibling = entityRow("host-2", 1);
    const plan = planEntityLiveEventPatch({
      hostRows: [target, sibling],
      identityRows: [],
      payload: patchPayload(),
      viewSchemaId: hostsViewSchemaId,
    });
    expect(plan.kind).toBe("apply");
    if (plan.kind !== "apply") throw new Error("expected apply plan");
    expect(plan.rows[0]).toMatchObject({ label: "After", rowVersion: 2 });
    expect(plan.rows[1]).toBe(sibling);
  });

  it("returns a reference-preserving stale no-op", () => {
    const rows = [entityRow("host-1", 2)];
    expect(
      planEntityLiveEventPatch({
        hostRows: rows,
        identityRows: [],
        payload: patchPayload(),
        viewSchemaId: hostsViewSchemaId,
      }),
    ).toEqual({ kind: "stale_noop" });
    expect(rows[0]?.label).toBe("Before");
  });

  it("requires exact surface, record, record type, change kind, schema, and version", () => {
    const hostRows = [entityRow("host-1", 1)];
    const inputs = [
      {
        payload: patchPayload(),
        viewSchemaId: "cartulary.view.timeline.v2",
      },
      {
        payload: patchPayload({ record_id: "host-2" }),
        viewSchemaId: hostsViewSchemaId,
      },
      {
        payload: patchPayload({ row_version: 3 }),
        viewSchemaId: hostsViewSchemaId,
      },
      {
        payload: patchPayload({
          affected_views: [
            { change_kind: "remove", view_schema_id: hostsViewSchemaId },
          ],
        }),
        viewSchemaId: hostsViewSchemaId,
      },
      {
        payload: patchPayload({
          affected_views: [
            {
              change_kind: "patch",
              patch_cells: {
                cells: { "identity.display_name": { value: "Wrong" } },
                record_id: "host-1",
                row_version: 2,
              },
              view_schema_id: hostsViewSchemaId,
            },
          ],
        }),
        viewSchemaId: hostsViewSchemaId,
      },
    ];
    for (const input of inputs) {
      expect(
        planEntityLiveEventPatch({ hostRows, identityRows: [], ...input }),
      ).toEqual({ kind: "refresh_required" });
    }
    expect(
      planEntityLiveEventPatch({
        hostRows: [entityRow("host-1", 1, "identity")],
        identityRows: [],
        payload: patchPayload(),
        viewSchemaId: hostsViewSchemaId,
      }),
    ).toEqual({ kind: "refresh_required" });
    expect(
      planEntityLiveEventPatch({
        hostRows,
        identityRows: [],
        payload: patchPayload(),
        viewSchemaId: identitiesViewSchemaId,
      }),
    ).toEqual({ kind: "refresh_required" });
  });
});
