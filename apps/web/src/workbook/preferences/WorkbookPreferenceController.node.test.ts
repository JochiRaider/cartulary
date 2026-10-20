import { afterEach, describe, expect, it, vi } from "vitest";
import {
  preferenceAccepted as accepted,
  preferenceAuthority as authority,
  defaultPreference,
  preferenceDeferred as deferred,
  preferenceFailed as failed,
  flushPreferences,
  homePreference,
  preferenceHosts as hosts,
  preferenceFixture,
  preferenceTimeline as timeline,
  preferenceUncertain as uncertain,
} from "../../testing/workbookPreferenceTestSupport";
import type {
  HomePreference,
  PreferenceResult,
} from "./workbookPreferenceModel";

describe("Workbook preference owner", () => {
  const controllers: ReturnType<typeof preferenceFixture>["controller"][] = [];
  const setup = () => {
    const fixture = preferenceFixture();
    controllers.push(fixture.controller);
    return fixture;
  };
  afterEach(() => {
    for (const c of controllers.splice(0)) c.dispose();
    vi.useRealTimers();
  });
  it("sets clears and repeats both resources without changing surface identity", async () => {
    const { controller: c, port } = setup();
    for (const kind of ["home", "default"] as const) {
      c.setCurrent(kind);
      await flushPreferences();
      expect(c.getSnapshot()[kind].operation.kind).toBe("confirmed");
      c.clear(kind);
      await flushPreferences();
      c.clear(kind);
      await flushPreferences();
      expect(c.getSnapshot()[kind].resource).toEqual(
        kind === "home" ? homePreference() : defaultPreference(),
      );
    }
    expect(port.setHomeSheet).toHaveBeenCalledTimes(3);
    expect(port.setDefaultSheet).toHaveBeenCalledTimes(3);
    expect(c.getSnapshot().surface?.sheetRef).toEqual(timeline);
  });
  it("admits home independently of a delayed or failed default read", async () => {
    const { controller: c, port } = setup();
    const read =
      deferred<PreferenceResult<ReturnType<typeof defaultPreference>>>();
    port.readDefault.mockImplementation(() => read.promise);
    c.setInspectionActive(true);
    await flushPreferences();
    expect(c.getSnapshot().home.read).toBe("ready");
    expect(c.getSnapshot().default.read).toBe("loading");
    read.resolve(failed);
    await flushPreferences();
    expect(c.getSnapshot().home.resource).toEqual(homePreference());
    expect(c.getSnapshot().default.read).toBe("failed");
  });
  it("locks each resource synchronously across entry points while allowing the other write", async () => {
    const { controller: c, port } = setup();
    const write = deferred<PreferenceResult<HomePreference>>();
    port.setHomeSheet.mockImplementation(() => write.promise);
    c.setCurrent("home");
    c.clear("home");
    c.setCurrent("default");
    await flushPreferences();
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
    expect(port.setDefaultSheet).toHaveBeenCalledTimes(1);
    c.setSurface({ sheetRef: hosts, label: "Hosts", available: true });
    c.setCurrent("home");
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
    write.resolve(accepted(homePreference(timeline)));
    await flushPreferences();
    expect(c.getSnapshot().home.operation.kind).toBe("confirmed");
  });
  it("captures the target before authorization and preserves it across a surface switch", async () => {
    const { controller: c, port, recover } = setup();
    const access = deferred<Awaited<ReturnType<typeof recover>>>();
    recover.mockImplementation(() => access.promise);
    c.setCurrent("home");
    c.setSurface({ sheetRef: hosts, label: "Hosts", available: true });
    access.resolve({
      kind: "authorized",
      userId: authority.actorId,
      role: "admin",
    });
    await flushPreferences();
    expect(port.setHomeSheet.mock.calls[0]?.[0].sheetRef).toEqual(timeline);
    expect(c.getSnapshot().surface?.sheetRef).toEqual(hosts);
  });
  it("ignores a delayed read that predates an acknowledged write", async () => {
    const { controller: c, port } = setup();
    const read = deferred<PreferenceResult<HomePreference>>();
    port.readHome.mockImplementationOnce(() => read.promise);
    c.refresh("home");
    c.setCurrent("home");
    await flushPreferences();
    read.resolve(accepted(homePreference(hosts)));
    await flushPreferences();
    expect(c.getSnapshot().home.resource?.home_sheet_ref).toEqual(timeline);
  });
  it("keeps acknowledgement and stale retained value after failed follow-up without resending", async () => {
    const { controller: c, port } = setup();
    port.readHome.mockResolvedValue(failed);
    c.setCurrent("home");
    await flushPreferences();
    expect(c.getSnapshot().home).toMatchObject({
      read: "failed",
      operation: { kind: "confirmed" },
      resource: homePreference(timeline),
    });
    c.refresh("home");
    await flushPreferences();
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
    expect(port.readHome).toHaveBeenCalledTimes(2);
  });
  it("retains uncertainty after a matching GET and requires an explicit new write", async () => {
    const { controller: c, port } = setup();
    port.setHomeSheet.mockResolvedValueOnce(uncertain);
    port.readHome.mockResolvedValue(accepted(homePreference(timeline)));
    c.setCurrent("home");
    await flushPreferences();
    const slot = c.getSnapshot().home;
    expect(slot.operation.kind).toBe("uncertain");
    c.setCurrent("home");
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
    if (slot.operation.kind !== "uncertain")
      throw new Error("Missing uncertainty");
    c.resolve("home", slot.operation.attempt.id, slot.observation, "write");
    await flushPreferences();
    expect(port.setHomeSheet).toHaveBeenCalledTimes(2);
    expect(c.getSnapshot().home.operation.kind).toBe("confirmed");
  });
  it("keeps a divergent observation without writing and rejects obsolete review decisions", async () => {
    const { controller: c, port } = setup();
    port.setHomeSheet.mockResolvedValueOnce(uncertain);
    c.setCurrent("home");
    await flushPreferences();
    const old = c.getSnapshot().home;
    c.refresh("home");
    await flushPreferences();
    if (old.operation.kind !== "uncertain")
      throw new Error("Missing uncertainty");
    c.resolve("home", old.operation.attempt.id, old.observation, "write");
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
    c.resolve(
      "home",
      old.operation.attempt.id,
      c.getSnapshot().home.observation,
      "keep",
    );
    expect(c.getSnapshot().home.operation.kind).toBe("reviewed");
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
  });
  it("bounds observation but retains the transport lock and accepts a late exact receipt", async () => {
    vi.useFakeTimers();
    const { controller: c, port } = setup();
    const write = deferred<PreferenceResult<HomePreference>>();
    port.setHomeSheet.mockImplementation(() => write.promise);
    c.setCurrent("home");
    await flushPreferences();
    await vi.advanceTimersByTimeAsync(30_001);
    expect(c.getSnapshot().home).toMatchObject({
      transportPending: true,
      operation: { kind: "uncertain" },
    });
    c.clear("home");
    c.refresh("home");
    await flushPreferences();
    expect(c.getSnapshot().home.observation).toBeNull();
    write.resolve(accepted(homePreference(timeline)));
    await flushPreferences();
    expect(c.getSnapshot().home).toMatchObject({
      transportPending: false,
      operation: { kind: "confirmed" },
    });
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
  });
  it("does not automatically repeat failed observations", async () => {
    vi.useFakeTimers();
    const { controller: c, port } = setup();
    port.setHomeSheet.mockResolvedValue(uncertain);
    port.readHome.mockResolvedValue(failed);
    c.setCurrent("home");
    await flushPreferences();
    await vi.advanceTimersByTimeAsync(120_000);
    expect(port.readHome).toHaveBeenCalledTimes(1);
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
  });
  it("rejects mismatched acknowledgement as uncertain and malformed reads as unavailable", async () => {
    const { controller: c, port } = setup();
    port.setHomeSheet.mockResolvedValue(accepted(homePreference(hosts)));
    port.readHome.mockResolvedValue(
      accepted({ ...homePreference(), user_id: "another-user" }),
    );
    c.setCurrent("home");
    await flushPreferences();
    expect(c.getSnapshot().home).toMatchObject({
      read: "failed",
      resource: null,
      operation: { kind: "uncertain" },
    });
  });
  it("admits personal viewer writes and denies default writes including preflight demotion", async () => {
    const { controller: c, port, setAuthority, recover } = setup();
    setAuthority({ ...authority, role: "viewer" });
    c.clear("home");
    c.clear("default");
    await flushPreferences();
    expect(port.setHomeSheet).toHaveBeenCalledTimes(1);
    expect(port.setDefaultSheet).not.toHaveBeenCalled();
    setAuthority(authority);
    recover.mockResolvedValue({
      kind: "authorized",
      userId: authority.actorId,
      role: "viewer",
    });
    c.clear("default");
    await flushPreferences();
    expect(port.setDefaultSheet).not.toHaveBeenCalled();
    expect(c.getSnapshot().default.operation.kind).toBe("rejected");
  });
  it("retains authorized recovery on role demotion but clears on incident actor or session retirement", async () => {
    const { controller: c, port, setAuthority } = setup();
    port.setDefaultSheet.mockResolvedValue(uncertain);
    c.setCurrent("default");
    await flushPreferences();
    setAuthority({ ...authority, role: "viewer" });
    expect(c.getSnapshot().default.operation.kind).toBe("uncertain");
    setAuthority({ ...authority, lifetime: "new-session" });
    expect(c.getSnapshot().default.resource).toBeNull();
    expect(c.hasDepartureWork()).toBe(false);
    c.setSurface({ sheetRef: hosts, available: true, label: "Hosts" });
    c.setCurrent("home");
    await flushPreferences();
    setAuthority({ ...authority, incidentId: "another-incident" });
    expect(c.getSnapshot().home.resource).toBeNull();
    setAuthority({ ...authority, actorId: "another-user" });
    expect(c.getSnapshot().surface).toBeNull();
  });
  it("does not publish late reads errors or writes into a replacement lifetime", async () => {
    const { controller: c, port, setAuthority, lost } = setup();
    const write = deferred<PreferenceResult<HomePreference>>();
    port.setHomeSheet.mockImplementation(() => write.promise);
    c.setCurrent("home");
    await flushPreferences();
    setAuthority({ ...authority, lifetime: "new-session" });
    write.resolve({
      kind: "rejected",
      status: 401,
      problem: { code: "authentication_required" },
    });
    await flushPreferences();
    expect(c.getSnapshot().home.operation.kind).toBe("idle");
    expect(lost).not.toHaveBeenCalled();
    expect(c.getSnapshot().announcement).toBeNull();
  });
  it("retains state during temporary access failure and retires only on confirmed loss", async () => {
    const { controller: c, recover, lost } = setup();
    c.setCurrent("home");
    await flushPreferences();
    recover.mockResolvedValue({ kind: "unavailable", failure: "transient" });
    c.clear("home");
    await flushPreferences();
    expect(c.getSnapshot().home.resource?.home_sheet_ref).toEqual(timeline);
    expect(lost).not.toHaveBeenCalled();
    recover.mockResolvedValue({ kind: "access_lost" });
    c.clear("home");
    await flushPreferences();
    expect(c.getSnapshot().home.resource).toBeNull();
    expect(lost).toHaveBeenCalledWith("incident", authority);
  });
  it("preserves unsettled work across panel closure and reviews explicit departure", async () => {
    const { controller: c, port } = setup();
    port.setHomeSheet.mockResolvedValue(uncertain);
    c.setCurrent("home");
    await flushPreferences();
    c.setInspectionActive(false);
    expect(c.hasDepartureWork()).toBe(true);
    const stay = c.requestLeave();
    c.resolveDeparture("stay");
    expect(await stay).toBe(false);
    expect(c.hasDepartureWork()).toBe(true);
    const leave = c.requestLeave();
    c.resolveDeparture("discard");
    expect(await leave).toBe(true);
    expect(c.hasDepartureWork()).toBe(false);
  });
  it("permits clear without an available surface and preserves extension and saved-view targets", async () => {
    const { controller: c, port } = setup();
    for (const sheetRef of [
      { kind: "saved_view", id: "saved-1" },
      {
        kind: "extension_workspace",
        extension_profile_id: "network_flow_activity",
        workspace_key: "network_analysis",
      },
    ] as const) {
      c.setSurface({ sheetRef, available: true, label: null });
      c.setCurrent("home");
      await flushPreferences();
      expect(c.getSnapshot().home.resource?.home_sheet_ref).toEqual(sheetRef);
    }
    c.setSurface(null);
    c.clear("home");
    await flushPreferences();
    expect(port.setHomeSheet.mock.calls.at(-1)?.[0].sheetRef).toBeNull();
  });
  it("allows read recovery after preflight failure interrupted an initial observation", async () => {
    const { controller: c, port, recover } = setup();
    const delayed = deferred<ReturnType<typeof homePreference>>();
    port.readHome.mockImplementationOnce(async () =>
      accepted(await delayed.promise),
    );
    c.setInspectionActive(true);
    recover.mockResolvedValueOnce({
      kind: "unavailable",
      failure: "transient",
    });
    c.clear("home");
    await flushPreferences();
    expect(c.getSnapshot().home.operation.kind).toBe("rejected");
    expect(c.getSnapshot().home.read).toBe("idle");
    c.refresh("home");
    await flushPreferences();
    expect(c.getSnapshot().home.read).toBe("ready");
    expect(port.setHomeSheet).not.toHaveBeenCalled();
    delayed.resolve(homePreference(hosts));
    await flushPreferences();
    expect(c.getSnapshot().home.resource?.home_sheet_ref).toBeNull();
  });
});
