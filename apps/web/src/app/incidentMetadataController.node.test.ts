import { afterEach, describe, expect, it, vi } from "vitest";
import {
  metadataDeferred as deferred,
  metadataIncident as incident,
  metadataActorId,
  metadataIncidentId,
} from "../testing/incidentMetadataTestSupport";
import type { MetadataResult } from "./api/incidentMetadataClient";
import {
  IncidentMetadataController,
  type IncidentMetadataPorts,
} from "./incidentMetadataController";
import {
  buildIncidentMetadataPatch,
  metadataFields,
  metadataValues,
  newMetadataDraft,
} from "./incidentMetadataModel";

const authority = {
  incidentId: metadataIncidentId,
  actorId: metadataActorId,
  lifetime: "metadata-session",
  role: "reviewer" as const,
};
const controllers: IncidentMetadataController[] = [];
const settle = async () => {
  for (let n = 0; n < 24; ++n) await Promise.resolve();
};
function setup() {
  const read = vi.fn<IncidentMetadataPorts["read"]>(async () => ({
    ok: true,
    resource: incident(),
  }));
  const patch = vi.fn<IncidentMetadataPorts["patch"]>(async () => ({
    ok: true,
    resource: incident({ severity: "critical", incident_version: 2 }),
  }));
  const recover = vi.fn<IncidentMetadataPorts["recover"]>(async (a) => ({
    kind: "authorized",
    role: a.role,
    userId: a.actorId,
  }));
  const lost = vi.fn();
  const publishResource = vi.fn();
  const isCurrent = vi.fn(() => true);
  const controller = new IncidentMetadataController({
    read,
    patch,
    recover,
    lost,
    publishResource,
    isCurrent,
  });
  controllers.push(controller);
  controller.setAuthority(authority);
  controller.setActive(true);
  return { controller, read, patch, recover, lost, publishResource, isCurrent };
}
afterEach(() => {
  for (const c of controllers.splice(0)) c.dispose();
  vi.useRealTimers();
});
describe("Incident metadata ownership", () => {
  it("constructs each sparse field independently and distinguishes omission null and raw text", () => {
    for (const field of metadataFields) {
      const draft = newMetadataDraft(incident());
      const value = field === "tlp" ? "TLP:RED" : "  e\u0301\r\nexact  ";
      expect(
        buildIncidentMetadataPatch({
          ...draft,
          values: { ...draft.values, [field]: value },
        }),
      ).toEqual({ base_incident_version: 1, [field]: value });
      expect(
        buildIncidentMetadataPatch({
          ...draft,
          values: { ...draft.values, [field]: "" },
        }),
      ).toEqual({ base_incident_version: 1, [field]: null });
    }
    const draft = newMetadataDraft(incident());
    expect(buildIncidentMetadataPatch(draft)).toEqual({
      base_incident_version: 1,
    });
    expect(
      buildIncidentMetadataPatch({
        ...draft,
        values: { ...draft.values, severity: " \u0085\u2003 " },
      }).severity,
    ).toBe(" \u0085\u2003 ");
    expect(() =>
      buildIncidentMetadataPatch({
        ...draft,
        values: { ...draft.values, tlp: "amber" },
      }),
    ).toThrow();
  });
  it("suppresses untouched saves and locks admission before current authorization", async () => {
    const { controller, patch, recover } = setup();
    await settle();
    controller.save();
    expect(patch).not.toHaveBeenCalled();
    const access =
      deferred<Awaited<ReturnType<IncidentMetadataPorts["recover"]>>>();
    recover.mockReturnValue(access.promise);
    controller.change("severity", "critical");
    controller.save();
    controller.save();
    expect(controller.getSnapshot().operation.kind).toBe("pending");
    expect(patch).not.toHaveBeenCalled();
    access.resolve({
      kind: "authorized",
      role: "reviewer",
      userId: metadataActorId,
    });
    await settle();
    expect(patch).toHaveBeenCalledTimes(1);
    expect(patch.mock.calls[0]?.[0].payload).toEqual({
      base_incident_version: 1,
      severity: "critical",
    });
  });
  it("captures an immutable attempt and acknowledges only submitted input revisions", async () => {
    const { controller, patch } = setup();
    await settle();
    const response = deferred<MetadataResult>();
    patch.mockReturnValue(response.promise);
    controller.change("severity", "critical");
    controller.save();
    await settle();
    controller.change("severity", "  newer  ");
    controller.change("description", "new description\n");
    const payload = patch.mock.calls[0]?.[0].payload;
    expect(Object.isFrozen(payload)).toBe(true);
    response.resolve({
      ok: true,
      resource: incident({ severity: "critical", incident_version: 2 }),
    });
    await settle();
    expect(controller.getSnapshot().draft?.values).toMatchObject({
      severity: "  newer  ",
      description: "new description\n",
    });
    expect(controller.getSnapshot().draft?.base.incident_version).toBe(2);
    expect(controller.canSave()).toBe(true);
  });
  it("keeps acknowledgement confirmed when access or resource publication reads fail", async () => {
    const { controller, read, patch, recover } = setup();
    await settle();
    read.mockRejectedValue(new Error("private transport details"));
    controller.change("severity", "critical");
    controller.save();
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(controller.getSnapshot().resource?.severity).toBe("critical");
    expect(controller.getSnapshot().read).toBe("failed");
    recover.mockResolvedValue({ kind: "unavailable", failure: "transient" });
    controller.refresh();
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(patch).toHaveBeenCalledTimes(1);
  });
  it("preserves a draft during refresh and requires review of changed resources", async () => {
    const { controller, read, patch } = setup();
    await settle();
    const response = deferred<MetadataResult>();
    read.mockReturnValue(response.promise);
    controller.refresh();
    controller.change("description", "exact\n  draft  ");
    response.resolve({
      ok: true,
      resource: incident({ severity: "external", incident_version: 2 }),
    });
    await settle();
    expect(controller.getSnapshot().draft?.values.description).toBe(
      "exact\n  draft  ",
    );
    expect(controller.getSnapshot().draft?.base.incident_version).toBe(1);
    expect(controller.canSave()).toBe(false);
    controller.review();
    expect(patch).not.toHaveBeenCalled();
    expect(controller.getSnapshot().draft?.values.severity).toBe("external");
    expect(controller.canSave()).toBe(true);
  });
  it("preserves conflicts and requires explicit review followed by a distinct save", async () => {
    const { controller, read, patch } = setup();
    await settle();
    patch.mockResolvedValueOnce({
      ok: false,
      status: 409,
      problem: { code: "incident_version_conflict" },
    });
    read.mockResolvedValue({
      ok: true,
      resource: incident({ severity: "external", incident_version: 2 }),
    });
    controller.change("severity", " intended ");
    controller.save();
    await settle();
    controller.save();
    expect(controller.getSnapshot().operation.kind).toBe("conflicted");
    expect(controller.getSnapshot().draft?.values.severity).toBe(" intended ");
    expect(patch).toHaveBeenCalledTimes(1);
    controller.review();
    expect(patch).toHaveBeenCalledTimes(1);
    controller.save();
    await settle();
    expect(patch.mock.calls[1]?.[0].payload).toEqual({
      base_incident_version: 2,
      severity: " intended ",
    });
  });
  it("never treats matching observation as proof of an uncertain write", async () => {
    const { controller, read, patch } = setup();
    await settle();
    patch.mockRejectedValue(new Error("lost reply"));
    controller.change("severity", "critical");
    controller.save();
    await settle();
    read.mockResolvedValue({
      ok: true,
      resource: incident({ severity: "critical", incident_version: 2 }),
    });
    controller.refresh();
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    controller.save();
    expect(patch).toHaveBeenCalledTimes(1);
    controller.review();
    expect(controller.getSnapshot().reviewNotice).toContain(
      "remains unconfirmed",
    );
    expect(controller.canSave()).toBe(false);
  });
  it("bounds observation while retaining the transport lock and accepts a late acknowledgement", async () => {
    vi.useFakeTimers();
    const { controller, patch } = setup();
    await settle();
    const response = deferred<MetadataResult>();
    patch.mockReturnValue(response.promise);
    controller.change("severity", "critical");
    controller.save();
    await settle();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    controller.refresh();
    await settle();
    controller.review();
    controller.save();
    expect(patch).toHaveBeenCalledTimes(1);
    response.resolve({
      ok: true,
      resource: incident({ severity: "critical", incident_version: 2 }),
    });
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(controller.getSnapshot().transportPending).toBe(false);
  });
  it("does not regress a newer accepted resource when a late acknowledgement arrives", async () => {
    const { controller, patch } = setup();
    await settle();
    const response = deferred<MetadataResult>();
    patch.mockReturnValue(response.promise);
    controller.change("severity", "critical");
    controller.save();
    await settle();
    controller.change("description", "new draft");
    controller.acceptResource(
      incident({ severity: "external", incident_version: 3 }),
    );
    response.resolve({
      ok: true,
      resource: incident({ severity: "critical", incident_version: 2 }),
    });
    await settle();
    expect(controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(controller.getSnapshot().resource?.incident_version).toBe(3);
    expect(controller.getSnapshot().draft?.values.description).toBe(
      "new draft",
    );
    expect(controller.canSave()).toBe(false);
  });
  it("keeps field rejection attached only to the submitted field revision", async () => {
    const { controller, patch } = setup();
    await settle();
    const response = deferred<MetadataResult>();
    patch.mockReturnValue(response.promise);
    controller.change("severity", "bad");
    controller.save();
    await settle();
    controller.change("severity", "newer");
    response.resolve({
      ok: false,
      status: 400,
      problem: {
        code: "invalid_incident_patch",
        field: "severity",
        reason: "field_too_long",
      },
    });
    await settle();
    expect(controller.getSnapshot().fieldErrors).toEqual({});
    expect(controller.getSnapshot().draft?.values.severity).toBe("newer");
    patch.mockResolvedValue({
      ok: false,
      status: 400,
      problem: {
        code: "invalid_incident_patch",
        field: "severity",
        reason: "control_character_not_allowed",
      },
    });
    controller.save();
    await settle();
    expect(controller.getSnapshot().fieldErrors.severity?.message).toContain(
      "control",
    );
    patch.mockResolvedValue({
      ok: false,
      status: 400,
      problem: {
        code: "invalid_incident_patch",
        field: "description",
        reason: "invalid_value",
      },
    });
    controller.save();
    await settle();
    expect(controller.getSnapshot().fieldErrors).toEqual({});
  });
  it("classifies malformed and unknown mutation outcomes as uncertain without raw diagnostics", async () => {
    for (const code of [
      "invalid_public_contract_response",
      "unknown_public_error",
    ] as const) {
      const { controller, patch } = setup();
      await settle();
      patch.mockResolvedValue({ ok: false, status: 502, problem: { code } });
      controller.change("severity", "critical");
      controller.save();
      await settle();
      expect(controller.getSnapshot().operation.kind).toBe("uncertain");
    }
  });
  it("retains one draft through closure and section return but rereads current eligibility", async () => {
    const { controller, read } = setup();
    await settle();
    controller.change("description", "retained");
    controller.setActive(false);
    expect(controller.getSnapshot().draft?.values.description).toBe("retained");
    const calls = read.mock.calls.length;
    controller.setActive(true);
    expect(controller.canSave()).toBe(false);
    await settle();
    expect(read).toHaveBeenCalledTimes(calls + 1);
    expect(controller.canSave()).toBe(true);
  });
  it("revalidates a reopened drawer while transport remains pending without admitting another write", async () => {
    const { controller, patch, read } = setup();
    await settle();
    const response = deferred<MetadataResult>();
    patch.mockReturnValue(response.promise);
    controller.change("severity", "critical");
    controller.save();
    await settle();
    controller.setActive(false);
    const count = read.mock.calls.length;
    controller.setActive(true);
    await settle();
    expect(read).toHaveBeenCalledTimes(count + 1);
    expect(controller.getSnapshot().access).toBe("ready");
    controller.change("severity", "newer after return");
    controller.save();
    expect(patch).toHaveBeenCalledTimes(1);
    response.resolve({
      ok: true,
      resource: incident({ severity: "critical", incident_version: 2 }),
    });
    await settle();
    expect(controller.getSnapshot().draft?.values.severity).toBe(
      "newer after return",
    );
  });
  it("allows active reviewers and admins but requires review after role or lifecycle changes", async () => {
    const { controller } = setup();
    await settle();
    controller.change("severity", "critical");
    for (const role of ["viewer", "editor"] as const) {
      controller.setAuthority({ ...authority, role });
      controller.save();
      expect(controller.canSave()).toBe(false);
      expect(controller.getSnapshot().resource).not.toBeNull();
    }
    controller.setAuthority({ ...authority, role: "admin" });
    expect(controller.canSave()).toBe(false);
    controller.review();
    expect(controller.canSave()).toBe(true);
    controller.acceptResource(
      incident({ status: "closed", incident_version: 2 }),
    );
    expect(controller.canSave()).toBe(false);
    controller.acceptResource(
      incident({ status: "active", incident_version: 3 }),
    );
    expect(controller.canSave()).toBe(false);
    controller.review();
    expect(controller.canSave()).toBe(true);
  });
  it("prevents dispatch after role loss or section exit during admission", async () => {
    const { controller, recover, patch } = setup();
    await settle();
    const access =
      deferred<Awaited<ReturnType<IncidentMetadataPorts["recover"]>>>();
    recover.mockReturnValue(access.promise);
    controller.change("severity", "critical");
    controller.save();
    controller.setActive(false);
    access.resolve({
      kind: "authorized",
      role: "reviewer",
      userId: metadataActorId,
    });
    await settle();
    expect(patch).not.toHaveBeenCalled();
    controller.setActive(true);
    await settle();
    recover.mockResolvedValue({
      kind: "authorized",
      role: "viewer",
      userId: metadataActorId,
    });
    controller.save();
    await settle();
    expect(patch).not.toHaveBeenCalled();
  });
  it("clears protected work on access or session loss without conflating their destinations", async () => {
    for (const reason of ["access_lost", "session_lost"] as const) {
      const { controller, recover, lost } = setup();
      await settle();
      controller.change("severity", "draft");
      recover.mockResolvedValue({ kind: reason });
      controller.refresh();
      await settle();
      expect(controller.getSnapshot().resource).toBeNull();
      expect(controller.getSnapshot().draft).toBeNull();
      expect(lost).toHaveBeenCalledWith(
        reason === "access_lost" ? "incident" : "session",
        authority,
      );
    }
  });
  it("fences delayed reads errors and writes across incident and retired account lifetimes", async () => {
    const { controller, read, patch, lost, publishResource } = setup();
    await settle();
    const response = deferred<MetadataResult>();
    patch.mockReturnValue(response.promise);
    controller.change("severity", "critical");
    controller.save();
    await settle();
    controller.retire();
    controller.setAuthority({
      ...authority,
      lifetime: "replacement",
      incidentId: "00000000-0000-4000-8000-000000001002",
    });
    read.mockResolvedValue({
      ok: true,
      resource: incident({
        incident_id: "00000000-0000-4000-8000-000000001002",
      }),
    });
    controller.setActive(true);
    await settle();
    const published = publishResource.mock.calls.length;
    response.resolve({
      ok: false,
      status: 401,
      problem: { code: "authentication_required" },
    });
    await settle();
    expect(lost).not.toHaveBeenCalled();
    expect(publishResource).toHaveBeenCalledTimes(published);
    expect(controller.getSnapshot().operation.kind).toBe("idle");
    expect(controller.getSnapshot().draft?.values).toEqual(
      metadataValues(incident()),
    );
    const delayed = deferred<MetadataResult>();
    read.mockReturnValue(delayed.promise);
    controller.refresh();
    controller.retire();
    delayed.resolve({ ok: true, resource: incident() });
    await settle();
    expect(controller.getSnapshot().resource).toBeNull();
  });
  it("accepts acknowledgement after local discard without inventing a reversal and preserves typing after discard", async () => {
    for (const newer of [false, true]) {
      const { controller, patch } = setup();
      await settle();
      const response = deferred<MetadataResult>();
      patch.mockReturnValue(response.promise);
      controller.change("severity", "critical");
      controller.save();
      await settle();
      controller.discard();
      if (newer) controller.change("severity", "after discard");
      response.resolve({
        ok: true,
        resource: incident({ severity: "critical", incident_version: 2 }),
      });
      await settle();
      expect(controller.getSnapshot().draft?.values.severity).toBe(
        newer ? "after discard" : "critical",
      );
      expect(controller.hasDepartureWork()).toBe(newer);
    }
  });
  it("discards only local input and requires explicit departure to forget pending recovery", async () => {
    const { controller, patch, publishResource } = setup();
    await settle();
    const response = deferred<MetadataResult>();
    patch.mockReturnValue(response.promise);
    controller.change("severity", "critical");
    controller.save();
    await settle();
    controller.discard();
    expect(controller.getSnapshot().operation.kind).toBe("pending");
    expect(controller.hasDepartureWork()).toBe(true);
    let departure = controller.requestLeave();
    controller.resolveDeparture("stay");
    expect(await departure).toBe(false);
    departure = controller.requestLeave();
    controller.resolveDeparture("discard");
    expect(await departure).toBe(true);
    expect(controller.getSnapshot().draft).toBeNull();
    const published = publishResource.mock.calls.length;
    response.resolve({ ok: true, resource: incident({ incident_version: 2 }) });
    await settle();
    expect(publishResource).toHaveBeenCalledTimes(published);
  });
});
