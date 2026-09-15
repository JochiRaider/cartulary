import { describe, expect, it } from "vitest";
import { timelineRow } from "../../testing/timelineWorkbookTestSupport";
import { WorkbookGridDraftStore } from "./WorkbookGridDraftStore";

const identity = {
  viewSchemaId: "cartulary.view.timeline.v2",
  recordId: "20000000-0000-4000-8000-000000000001",
  fieldKey: "timeline.activity_synopsis_text",
};
const authority = {
  actorId: "actor",
  sessionIdentity: "session",
  incidentId: "incident",
  role: "editor" as const,
  closed: false,
};
const row = (version = 1, value = "Saved") => ({
  ...timelineRow({
    captureState: "rough",
    recordId: identity.recordId,
    rowVersion: version,
  }),
  cells: { [identity.fieldKey]: { value } },
});
function fixture() {
  const store = new WorkbookGridDraftStore();
  store.setAuthority(authority);
  return store;
}

describe("Retained grid authoring", () => {
  it("retains exact raw input and explicit clear independently by semantic target", () => {
    const store = fixture();
    store.update(identity, row(), "  unfinished\r\ntext  ");
    expect(store.read(identity)?.value).toBe("  unfinished\r\ntext  ");
    expect(
      store.read({ ...identity, fieldKey: "timeline.analyst_text" }),
    ).toBeNull();
    expect(
      store.read({ ...identity, viewSchemaId: "cartulary.view.notes.v1" }),
    ).toBeNull();
    store.update(identity, row(), null);
    expect(store.read(identity)?.value).toBeNull();
    expect(store.capture(identity)).not.toBeNull();
    store.discard(identity);
    expect(store.read(identity)).toBeNull();
  });
  it("uses authoring revisions so an older equal-valued attempt cannot clear newer work", () => {
    const store = fixture();
    store.update(identity, row(), "A");
    const a = store.capture(identity);
    store.setValidation(a, "Invalid A");
    expect(store.read(identity)?.validation).toBe("Invalid A");
    store.update(identity, row(), "B");
    store.update(identity, row(), "A");
    store.setValidation(a, "Obsolete rejection");
    expect(store.read(identity)?.validation).toBeNull();
    store.acknowledge(a);
    expect(store.read(identity)?.value).toBe("A");
    store.acknowledge(store.capture(identity));
    expect(store.read(identity)).toBeNull();
  });
  it("requires review for relevant committed changes without retargeting the draft", () => {
    const store = fixture();
    store.update(identity, row(), "Raw");
    expect(store.staleFields(identity, row(2))).toEqual([]);
    expect(store.staleFields(identity, row(3, "Remote"))).toEqual([
      identity.fieldKey,
    ]);
    store.review(identity, row(3, "Remote"));
    expect(store.staleFields(identity, row(3, "Remote"))).toEqual([]);
    expect(store.read(identity)?.value).toBe("Raw");
    store.update(identity, { ...row(), record_id: "other" }, "Wrong");
    expect(store.read(identity)?.value).toBe("Raw");
  });
  it("conceals suspended authoring and restores only the same account", () => {
    const store = fixture();
    store.update(identity, row(), "Private");
    store.setAuthority(null);
    expect(store.read(identity)).toBeNull();
    expect(store.list(identity.viewSchemaId)).toEqual([]);
    store.update(identity, row(), "Denied");
    store.setAuthority({
      ...authority,
      sessionIdentity: "replacement-session",
    });
    expect(store.read(identity)?.value).toBe("Private");
    store.setAuthority({ ...authority, actorId: "another-account" });
    expect(store.read(identity)).toBeNull();
    store.setAuthority(authority);
    expect(store.read(identity)).toBeNull();
  });
  it("keeps readable rejected work on closure and role loss without granting mutation authority", () => {
    const store = fixture();
    store.update(identity, row(), "Retained");
    store.setAuthority({ ...authority, closed: true });
    expect(store.canAuthor()).toBe(false);
    expect(store.read(identity)?.value).toBe("Retained");
    store.setAuthority({ ...authority, role: "viewer" });
    store.update(identity, row(), "Forbidden");
    expect(store.read(identity)?.value).toBe("Retained");
    expect(store.list(identity.viewSchemaId)).toHaveLength(1);
    store.retire();
    expect(store.list(identity.viewSchemaId)).toEqual([]);
    expect(store.capture(identity)).toBeNull();
  });
});
