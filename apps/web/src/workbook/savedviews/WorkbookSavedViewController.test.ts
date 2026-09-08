import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import { observeAccountOperation } from "../../app/accountOperation";
import { deferred } from "../../testing/fetchMockTestSupport";
import { buildSavedViewLayoutJson } from "../models/workbookQuery";
import type { SavedViewResource } from "../models/workbookSavedViews";
import type {
  SavedViewResult,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";
import type {
  SavedViewAuthority,
  SavedViewBinding,
  SavedViewControllerPorts,
} from "./savedViewOperationModel";
import { WorkbookSavedViewController } from "./WorkbookSavedViewController";

const schema = "cartulary.view.timeline.v2";
const authority: SavedViewAuthority = {
  incidentId: "incident-1",
  actorId: "user-1",
  lifetime: "session-1",
  role: "admin",
};
function resource(
  overrides: Partial<SavedViewResource> = {},
): SavedViewResource {
  return {
    incident_id: authority.incidentId,
    saved_view_id: "view-1",
    view_schema_id: schema,
    display_name: "Saved timeline",
    scope: "private",
    owner_user_id: authority.actorId,
    saved_view_version: 1,
    created_at: "2026-08-01T00:00:00Z",
    updated_at: "2026-08-01T00:00:00Z",
    query_json: { sort: [], filters: [] },
    layout_json: buildSavedViewLayoutJson(requireViewContract(schema)),
    ...overrides,
  };
}
async function settle() {
  for (let i = 0; i < 30; i++) await Promise.resolve();
}
async function setup(
  overrides: Partial<WorkbookSavedViewPort> = {},
  options: Partial<SavedViewControllerPorts> = {},
) {
  let resources = [resource()];
  const ports: WorkbookSavedViewPort = {
    listPage: vi.fn<WorkbookSavedViewPort["listPage"]>(async () => ({
      kind: "accepted",
      value: { nextCursor: null, savedViews: resources },
    })),
    create: vi.fn<WorkbookSavedViewPort["create"]>(async ({ definition }) => {
      const created = resource({
        saved_view_id: "view-new",
        display_name: definition.displayName,
        scope: definition.scope,
        query_json: definition.queryJson,
        layout_json: definition.layoutJson,
      });
      resources = [...resources, created];
      return { kind: "accepted", value: created };
    }),
    patch: vi.fn<WorkbookSavedViewPort["patch"]>(async ({ base, changes }) => {
      const updated = {
        ...base,
        display_name: changes.displayName ?? base.display_name,
        query_json: changes.queryJson ?? base.query_json,
        layout_json: changes.layoutJson ?? base.layout_json,
        scope: changes.scope ?? base.scope,
        saved_view_version: base.saved_view_version + 1,
        updated_at: "2026-08-01T00:01:00Z",
      };
      resources = [updated];
      return { kind: "accepted", value: updated };
    }),
    delete: vi.fn<WorkbookSavedViewPort["delete"]>(async () => {
      resources = [];
      return { kind: "accepted", value: undefined };
    }),
    ...overrides,
  };
  const lost = vi.fn();
  const recover = vi.fn<SavedViewControllerPorts["recover"]>(async (a) => ({
    kind: "authorized",
    userId: a.actorId,
    role: a.role,
  }));
  let active = authority;
  const controller = new WorkbookSavedViewController({
    observe: observeAccountOperation,
    port: () => ports,
    recover,
    lost,
    isCurrent: (a) =>
      a.incidentId === active.incidentId &&
      a.lifetime === active.lifetime &&
      a.actorId === active.actorId,
    ...options,
  });
  controller.setAuthority(active);
  await settle();
  const binding: SavedViewBinding = {
    incidentId: authority.incidentId,
    subject: {
      viewSchemaId: schema,
      savedViewId: null,
      savedViewVersion: null,
    },
    sheetRef: { kind: "view_schema", id: schema },
    selectionGeneration: 1,
    workingGeneration: 1,
    queryJson: { sort: [], filters: [] },
    layoutJson: buildSavedViewLayoutJson(requireViewContract(schema)),
    applyConfiguration: vi.fn(),
    select: vi.fn(),
    deleted: vi.fn(),
    authorizationRecovered: vi.fn(),
  };
  controller.setBinding(binding);
  return {
    controller,
    ports,
    binding,
    recover,
    lost,
    resources: (next: SavedViewResource[]) => {
      resources = next;
    },
    authority: (next: SavedViewAuthority) => {
      active = next;
      controller.setAuthority(next);
    },
  };
}
function selectSaved(h: Awaited<ReturnType<typeof setup>>, base = resource()) {
  const binding = {
    ...h.binding,
    subject: {
      viewSchemaId: base.view_schema_id,
      savedViewId: base.saved_view_id,
      savedViewVersion: base.saved_view_version,
    },
    sheetRef: { kind: "saved_view" as const, id: base.saved_view_id },
  };
  h.controller.setBinding(binding);
  return binding;
}
afterEach(() => vi.useRealTimers());

describe("Saved-view operation owner", () => {
  it("captures one exact intent and admits one synchronous write", async () => {
    const pending = deferred<SavedViewResult<SavedViewResource>>();
    const create = vi.fn(() => pending.promise);
    const h = await setup({ create });
    h.controller.changeDraft(h.binding.subject, { displayName: "Submitted" });
    h.controller.run({ kind: "create" }, h.binding.subject);
    h.controller.run({ kind: "create" }, h.binding.subject);
    await settle();
    expect(create).toHaveBeenCalledTimes(1);
    h.controller.changeDraft(h.binding.subject, { displayName: "Newer draft" });
    pending.resolve({
      kind: "accepted",
      value: resource({ saved_view_id: "new", display_name: "Submitted" }),
    });
    await settle();
    expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(h.controller.draftFor(h.binding.subject).displayName).toBe(
      "Newer draft",
    );
    expect(h.binding.select).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("retains a receipt after presentation detaches without navigation", async () => {
    const pending = deferred<SavedViewResult<SavedViewResource>>();
    const h = await setup({ create: () => pending.promise });
    h.controller.run({ kind: "create" }, h.binding.subject);
    await settle();
    h.controller.setBinding(null);
    pending.resolve({
      kind: "accepted",
      value: resource({ saved_view_id: "new" }),
    });
    await settle();
    expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(h.binding.select).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("does not overwrite working edits or navigate after surface selection changes", async () => {
    for (const change of ["working", "selection", "surface"] as const) {
      const pending = deferred<SavedViewResult<SavedViewResource>>();
      const h = await setup({ create: () => pending.promise });
      h.controller.run({ kind: "create" }, h.binding.subject);
      await settle();
      h.controller.setBinding({
        ...h.binding,
        workingGeneration: change === "working" ? 2 : 1,
        selectionGeneration: change !== "working" ? 2 : 1,
        ...(change === "surface"
          ? {
              subject: {
                viewSchemaId: "cartulary.view.evidence.v1",
                savedViewId: null,
                savedViewVersion: null,
              },
            }
          : {}),
      });
      pending.resolve({
        kind: "accepted",
        value: resource({ saved_view_id: "new" }),
      });
      await settle();
      expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
      expect(h.binding.select).not.toHaveBeenCalled();
      h.controller.dispose();
    }
  });
  it("fences late callbacks and sensitive state across incident actor and session replacement", async () => {
    for (const replacement of [
      { ...authority, incidentId: "incident-2" },
      { ...authority, actorId: "user-2" },
      { ...authority, lifetime: "session-2" },
    ]) {
      const pending = deferred<SavedViewResult<SavedViewResource>>();
      const h = await setup({ create: () => pending.promise });
      h.controller.run({ kind: "create" }, h.binding.subject);
      await settle();
      h.authority(replacement);
      pending.resolve({
        kind: "accepted",
        value: resource({ saved_view_id: "new" }),
      });
      await settle();
      expect(h.controller.getSnapshot().operation.kind).toBe("idle");
      expect(h.binding.select).not.toHaveBeenCalled();
      expect(h.controller.getSnapshot().drafts.size).toBe(0);
      h.controller.dispose();
    }
  });
  it("bounds observation keeps the transport lock and accepts an exact late receipt", async () => {
    vi.useFakeTimers();
    const pending = deferred<SavedViewResult<SavedViewResource>>();
    const create = vi.fn(() => pending.promise);
    const h = await setup({ create });
    h.controller.run({ kind: "create" }, h.binding.subject);
    await settle();
    await vi.advanceTimersByTimeAsync(30_001);
    expect(h.controller.getSnapshot()).toMatchObject({
      transportPending: true,
      operation: { kind: "uncertain" },
    });
    h.controller.run({ kind: "create" }, h.binding.subject);
    await settle();
    expect(create).toHaveBeenCalledTimes(1);
    pending.resolve({
      kind: "accepted",
      value: resource({ saved_view_id: "new" }),
    });
    await settle();
    expect(h.controller.getSnapshot()).toMatchObject({
      transportPending: false,
      operation: { kind: "confirmed" },
    });
    h.controller.dispose();
  });
  it("does not infer a create receipt or replay an uncertain create", async () => {
    const create = vi.fn(
      async (): Promise<SavedViewResult<SavedViewResource>> => ({
        kind: "uncertain",
        failure: { kind: "transport", message: "lost" },
      }),
    );
    const h = await setup({ create });
    h.controller.changeDraft(h.binding.subject, {
      displayName: resource().display_name,
    });
    h.controller.run({ kind: "create" }, h.binding.subject);
    await settle();
    expect(h.controller.getSnapshot().operation.kind).toBe("uncertain");
    expect(create).toHaveBeenCalledTimes(1);
    const snapshot = h.controller.getSnapshot();
    if (snapshot.operation.kind === "idle") throw new Error("Missing attempt");
    expect(
      h.controller.canReview(
        snapshot.operation.attempt.id,
        snapshot.observation,
      ),
    ).toBe(true);
    h.controller.review(
      snapshot.operation.attempt.id,
      snapshot.observation,
      "create",
    );
    await settle();
    expect(create).toHaveBeenCalledTimes(2);
    h.controller.dispose();
  });
  it("retains submitted conflict changes and requires current explicit review", async () => {
    const patch = vi
      .fn<WorkbookSavedViewPort["patch"]>()
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: {
          kind: "conflict",
          message: "changed",
          conflict: {
            savedViewId: "view-1",
            baseVersion: 1,
            currentVersion: 2,
          },
        },
      });
    const h = await setup({ patch });
    const binding = selectSaved(h);
    h.controller.changeDraft(binding.subject, {
      displayName: "Submitted name",
    });
    h.resources([
      resource({
        saved_view_version: 2,
        scope: "shared",
        display_name: "Concurrent name",
      }),
    ]);
    h.controller.run({ kind: "update" }, binding.subject);
    await settle();
    const old = h.controller.getSnapshot();
    if (old.operation.kind !== "conflict") throw new Error("Missing conflict");
    expect(old.operation.attempt.definition.displayName).toBe("Submitted name");
    expect(patch).toHaveBeenCalledTimes(1);
    await h.controller.refresh();
    expect(
      h.controller.canReview(old.operation.attempt.id, old.observation),
    ).toBe(false);
    const snapshot = h.controller.getSnapshot();
    patch.mockResolvedValueOnce({
      kind: "accepted",
      value: resource({
        saved_view_version: 3,
        scope: "shared",
        display_name: "Submitted name",
      }),
    });
    h.controller.review(
      old.operation.attempt.id,
      snapshot.observation,
      "apply",
    );
    await settle();
    expect(patch).toHaveBeenLastCalledWith(
      expect.objectContaining({
        base: expect.objectContaining({
          saved_view_version: 2,
          scope: "shared",
        }),
        changes: { displayName: "Submitted name" },
      }),
    );
    expect(h.binding.applyConfiguration).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("keeps accepted writes confirmed when list materialization fails", async () => {
    const h = await setup();
    vi.mocked(h.ports.listPage).mockResolvedValue({
      kind: "rejected",
      failure: { kind: "transport", message: "list failed" },
    });
    h.controller.run({ kind: "create" }, h.binding.subject);
    await settle();
    expect(h.controller.getSnapshot()).toMatchObject({
      operation: { kind: "confirmed" },
      list: "ready",
      listProblem: { kind: "transport" },
    });
    expect(
      h.controller
        .getSnapshot()
        .resources.some((r) => r.saved_view_id === "view-new"),
    ).toBe(true);
    h.controller.dispose();
  });
  it("prevents an older list from overwriting a receipt or resurrecting a deletion", async () => {
    for (const kind of ["update", "delete"] as const) {
      const stale =
        deferred<Awaited<ReturnType<WorkbookSavedViewPort["listPage"]>>>();
      const h = await setup();
      const binding = selectSaved(h);
      vi.mocked(h.ports.listPage).mockReturnValueOnce(stale.promise);
      const oldRead = h.controller.refresh();
      h.controller.changeDraft(binding.subject, { displayName: "Updated" });
      h.controller.run({ kind }, binding.subject);
      await settle();
      stale.resolve({
        kind: "accepted",
        value: { nextCursor: null, savedViews: [resource()] },
      });
      await oldRead;
      await settle();
      expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
      if (kind === "delete")
        expect(h.controller.getSnapshot().resources).toEqual([]);
      else
        expect(h.controller.getSnapshot().resources[0]?.display_name).toBe(
          "Updated",
        );
      h.controller.dispose();
    }
  });
  it("duplicates the canonical source and resets only local configuration", async () => {
    const h = await setup();
    const binding = selectSaved(h);
    h.controller.setBinding({
      ...binding,
      queryJson: { sort: [], filters: [], group_by: "timeline.capture_state" },
      workingGeneration: 2,
    });
    h.controller.run({ kind: "duplicate" }, binding.subject);
    await settle();
    expect(h.ports.create).toHaveBeenCalledWith(
      expect.objectContaining({
        definition: expect.objectContaining({
          displayName: "Saved timeline Copy",
          scope: "private",
          queryJson: resource().query_json,
          layoutJson: resource().layout_json,
        }),
      }),
    );
    h.controller.run({ kind: "reset" }, binding.subject);
    expect(h.binding.applyConfiguration).toHaveBeenCalledWith(
      schema,
      resource().query_json,
      resource().layout_json,
    );
    expect(h.ports.patch).not.toHaveBeenCalled();
    expect(h.ports.delete).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("permits owner viewers and visible system duplication while denying system mutation", async () => {
    const h = await setup();
    h.authority({ ...authority, role: "viewer" });
    await settle();
    const binding = selectSaved(h);
    expect(
      h.controller.unavailableReason("update", binding.subject),
    ).toBeNull();
    h.resources([resource({ scope: "system", owner_user_id: null })]);
    await h.controller.refresh();
    expect(h.controller.unavailableReason("update", binding.subject)).toContain(
      "immutable",
    );
    expect(h.controller.unavailableReason("delete", binding.subject)).toContain(
      "immutable",
    );
    expect(
      h.controller.unavailableReason("duplicate", binding.subject),
    ).toBeNull();
    h.controller.dispose();
  });
  it("rechecks role before dispatch and clears newly hidden private recovery", async () => {
    for (const kind of ["create", "duplicate", "update", "delete"] as const) {
      const h = await setup();
      const foreign = resource({ owner_user_id: "user-2" });
      h.resources([foreign]);
      await h.controller.refresh();
      const binding = selectSaved(h, foreign);
      h.recover.mockResolvedValue({
        kind: "authorized",
        userId: authority.actorId,
        role: "viewer",
      });
      h.controller.run({ kind }, binding.subject);
      await settle();
      expect(h.ports.create).not.toHaveBeenCalled();
      expect(h.ports.patch).not.toHaveBeenCalled();
      expect(h.ports.delete).not.toHaveBeenCalled();
      expect(h.controller.getSnapshot().resources).toEqual([]);
      expect(h.controller.getSnapshot().operation.kind).toBe("idle");
      h.controller.dispose();
    }
  });
  it("falls back after confirmed deletion only when its target remains selected while preserving newer edits", async () => {
    for (const change of ["working", "form", "selection"] as const) {
      const pending = deferred<SavedViewResult<undefined>>();
      const h = await setup({ delete: () => pending.promise });
      const binding = selectSaved(h);
      h.controller.run({ kind: "delete" }, binding.subject);
      await settle();
      if (change === "working")
        h.controller.setBinding({
          ...binding,
          workingGeneration: 2,
          queryJson: {
            sort: [],
            filters: [],
            group_by: "timeline.capture_state",
          },
        });
      if (change === "form")
        h.controller.changeDraft(binding.subject, {
          displayName: "Newer form",
        });
      if (change === "selection")
        h.controller.setBinding({ ...h.binding, selectionGeneration: 2 });
      h.resources([]);
      pending.resolve({ kind: "accepted", value: undefined });
      await settle();
      expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
      expect(binding.deleted).toHaveBeenCalledTimes(
        change === "selection" ? 0 : 1,
      );
      expect(binding.applyConfiguration).not.toHaveBeenCalled();
      expect(binding.select).not.toHaveBeenCalled();
      if (change === "form")
        expect(h.controller.draftFor(binding.subject).displayName).toBe(
          "Newer form",
        );
      h.controller.dispose();
    }
  });
  it("treats saved-view denial as a resource problem and revalidates incident access", async () => {
    for (const kind of [
      "authorization_denied",
      "unavailable_target",
    ] as const) {
      const h = await setup({
        patch: async () => ({
          kind: "rejected",
          failure: { kind, message: "resource unavailable" },
        }),
      });
      const binding = selectSaved(h);
      h.controller.run({ kind: "update" }, binding.subject);
      await settle();
      expect(h.lost).not.toHaveBeenCalled();
      expect(h.recover).toHaveBeenCalledTimes(2);
      expect(h.controller.getSnapshot().authority).not.toBeNull();
      h.controller.dispose();
    }
  });
  it("uses authoritative access loss and never publishes a stale access recovery", async () => {
    const h = await setup();
    h.recover.mockResolvedValue({ kind: "access_lost" });
    h.controller.run({ kind: "create" }, h.binding.subject);
    await settle();
    expect(h.lost).toHaveBeenCalledWith("incident", authority);
    expect(h.controller.getSnapshot().resources).toEqual([]);
    expect(h.ports.create).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("keeps invalid duplicate names editable through explicit recovery", async () => {
    const h = await setup();
    const long = resource({ display_name: "x".repeat(256) });
    h.resources([long]);
    await h.controller.refresh();
    const binding = selectSaved(h, long);
    h.controller.run({ kind: "duplicate" }, binding.subject);
    await settle();
    const snapshot = h.controller.getSnapshot();
    expect(snapshot.operation).toMatchObject({
      kind: "rejected",
      problem: { field: "display_name" },
    });
    expect(h.ports.create).not.toHaveBeenCalled();
    if (snapshot.operation.kind === "idle") throw new Error("Missing attempt");
    h.controller.review(
      snapshot.operation.attempt.id,
      snapshot.observation,
      "create",
      "Short copy",
    );
    await settle();
    expect(h.ports.create).toHaveBeenCalledWith(
      expect.objectContaining({
        definition: expect.objectContaining({ displayName: "Short copy" }),
      }),
    );
    h.controller.dispose();
  });
  it("observes uncertain updates and deletes without inferring receipts or silently repeating writes", async () => {
    for (const kind of ["update", "delete"] as const) {
      const h = await setup();
      const binding = selectSaved(h);
      const failure = {
        kind: "uncertain" as const,
        failure: { kind: "transport" as const, message: "Response lost" },
      };
      if (kind === "update")
        vi.mocked(h.ports.patch).mockResolvedValueOnce(failure);
      else vi.mocked(h.ports.delete).mockResolvedValueOnce(failure);
      h.controller.changeDraft(binding.subject, {
        displayName: "Retained draft",
      });
      h.controller.run({ kind }, binding.subject);
      await settle();
      const attempted = h.controller.getSnapshot().operation;
      if (attempted.kind !== "uncertain")
        throw new Error("Expected uncertain outcome");
      h.resources(
        kind === "delete"
          ? []
          : [
              resource({
                display_name: "Retained draft",
                saved_view_version: 2,
              }),
            ],
      );
      await h.controller.refresh();
      const state = h.controller.getSnapshot();
      expect(state.operation.kind).toBe("uncertain");
      expect(
        h.ports[kind === "update" ? "patch" : "delete"],
      ).toHaveBeenCalledTimes(1);
      if (kind === "delete") {
        h.controller.review(attempted.attempt.id, state.observation, "apply");
        await settle();
        expect(h.ports.delete).toHaveBeenCalledTimes(1);
        h.controller.review(attempted.attempt.id, state.observation, "keep");
        expect(h.controller.getSnapshot().operation.kind).toBe("reviewed");
      } else {
        h.controller.review(attempted.attempt.id, state.observation, "apply");
        await settle();
        expect(h.ports.patch).toHaveBeenCalledTimes(2);
        expect(h.ports.patch).toHaveBeenLastCalledWith(
          expect.objectContaining({
            base: expect.objectContaining({ saved_view_version: 2 }),
            changes: {},
          }),
        );
      }
      expect(h.binding.select).not.toHaveBeenCalled();
      expect(h.binding.applyConfiguration).not.toHaveBeenCalled();
      h.controller.dispose();
    }
  });
  it("enforces the complete visible scope ownership and role action matrix", async () => {
    for (const role of ["viewer", "editor", "reviewer", "admin"] as const) {
      for (const scope of ["private", "shared", "system"] as const) {
        for (const own of [false, true]) {
          const h = await setup();
          h.authority({ ...authority, role });
          const source = resource({
            scope,
            owner_user_id:
              scope === "system"
                ? null
                : own
                  ? authority.actorId
                  : "another-user",
          });
          h.resources([source]);
          await h.controller.refresh();
          const binding = selectSaved(h, source);
          const visible = scope !== "private" || own || role === "admin";
          for (const action of [
            "create",
            "duplicate",
            "reset",
            "update",
            "delete",
          ] as const) {
            const allowed =
              visible &&
              ((action !== "update" && action !== "delete") ||
                (scope !== "system" && (own || role === "admin")));
            expect(
              h.controller.unavailableReason(action, binding.subject) === null,
              `${role}/${scope}/${own}/${action}`,
            ).toBe(allowed);
          }
          h.controller.dispose();
        }
      }
    }
  });
  it("bounds access revalidation after a denied list without looping or declaring incident loss", async () => {
    const h = await setup();
    vi.mocked(h.ports.listPage).mockResolvedValue({
      kind: "rejected",
      failure: { kind: "authorization_denied", message: "List denied" },
    });
    await h.controller.refresh();
    await settle();
    expect(h.recover).toHaveBeenCalledTimes(1);
    expect(h.lost).not.toHaveBeenCalled();
    expect(h.controller.getSnapshot().access).toBe("unavailable");
    expect(h.ports.listPage).toHaveBeenCalledTimes(3);
    h.controller.dispose();
  });
});
