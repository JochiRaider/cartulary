import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { createAssessmentAppendTransport } from "../../adapters/createAssessmentAppendTransport";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type {
  AssessmentAppendAttempt,
  AssessmentAppendOutcome,
  AssessmentAppendReceipt,
} from "./assessmentOperation";
import { useAssessmentCreationController } from "./useAssessmentCreationController";
import { WorkbookAssessmentAuthoringOwner } from "./WorkbookAssessmentAuthoringOwner";

const authority: WorkbookMutationAuthority = {
  actorId: "actor",
  sessionIdentity: "session",
  incidentId: "incident",
  role: "editor",
  closed: false,
};
const sheetRef = {
  kind: "view_schema",
  id: "cartulary.view.assessments.v1",
} as const;
const receipt: AssessmentAppendReceipt = {
  data: {
    change_set_id: "change-set",
    row: { cells: {}, record_id: "assessment", row_version: 1 },
    view_schema_id: sheetRef.id,
  },
  meta: { request_id: "request" },
};
const accepted = { kind: "accepted", receipt } as const;
function setup(
  send = vi.fn(
    async (
      _attempt: AssessmentAppendAttempt,
      _signal: AbortSignal,
    ): Promise<AssessmentAppendOutcome> => accepted,
  ),
  refresh = vi.fn(async () => {}),
) {
  const ids = {
    create: vi.fn(() => `assessment-${ids.create.mock.calls.length}`),
  };
  const effect = vi.fn();
  const owner = new WorkbookAssessmentAuthoringOwner(
    authority.incidentId,
    ids,
    { accepted: effect, refresh },
  );
  const transport = { ...createAssessmentAppendTransport(undefined), send };
  const readAuthority = vi.fn(async () => authority);
  owner.configure(transport, readAuthority);
  owner.setAuthority(authority);
  const prepare = () => {
    owner.openStandalone();
    owner.updateDraft((draft) => ({
      ...draft,
      subjectRecordId: "host",
      rationale: "Judgment",
      supportRecordIds: ["support"],
    }));
  };
  return {
    owner,
    transport,
    send,
    ids,
    effect,
    refresh,
    readAuthority,
    prepare,
  };
}
function entryAt(owner: WorkbookAssessmentAuthoringOwner, index = 0) {
  const entry = owner.getSnapshot().entries[index];
  if (!entry) throw new Error("Expected retained append attempt.");
  return entry;
}
const binding = { sheetRef, isCurrent: () => true };

describe("useAssessmentCreationController", () => {
  it("publishes accepted standalone creation as polite neutral feedback", async () => {
    const { owner } = setup();
    const { result } = renderHook(() =>
      useAssessmentCreationController({
        owner,
        sheetRef,
        lifecycleResetKey: "selected",
      }),
    );
    act(() => {
      result.current.commands.openStandalone();
      result.current.commands.updateDraft((draft) => ({
        ...draft,
        subjectRecordId: "host",
        rationale: "Judgment",
      }));
    });
    await act(async () => result.current.commands.submit(true));
    expect(result.current.snapshot.feedback).toEqual({
      announcement: "polite",
      kind: "message",
      message: "Assessment created.",
    });
    expect(result.current.snapshot.hasDraft).toBe(false);
  });
  it("retains the append-only draft and support selection after rejection", async () => {
    const { owner, prepare, send } = setup(
      vi.fn(async () => ({
        kind: "rejected",
        failure: { kind: "validation", message: "Support is unavailable." },
      })),
    );
    prepare();
    const before = owner.getSnapshot().draft;
    await owner.submit(binding);
    expect(send).toHaveBeenCalledTimes(1);
    expect(owner.getSnapshot().draft).toBe(before);
    expect(owner.getSnapshot().entries[0]?.phase).toBe("rejected");
  });
  it("retains drafts across selection and close and requires explicit resume or discard", () => {
    const { owner } = setup();
    const { result, rerender, unmount } = renderHook(
      ({ key }) =>
        useAssessmentCreationController({
          owner,
          sheetRef,
          lifecycleResetKey: key,
        }),
      { initialProps: { key: "original" } },
    );
    act(() => {
      result.current.commands.openStandalone();
      result.current.commands.updateDraft((draft) => ({
        ...draft,
        rationale: "Keep me",
      }));
    });
    rerender({ key: "elsewhere" });
    expect(result.current.snapshot.attached).toBe(false);
    act(() => result.current.commands.openStandalone());
    expect(result.current.snapshot.draft.rationale).toBe("Keep me");
    act(() => result.current.commands.resume());
    expect(result.current.snapshot.attached).toBe(true);
    act(() => result.current.commands.cancel());
    expect(result.current.snapshot.hasDraft).toBe(true);
    unmount();
    expect(owner.getSnapshot().draft?.values.rationale).toBe("Keep me");
    owner.discardDraft();
    expect(owner.getSnapshot().draft).toBeNull();
  });
  it("detaches inline feedback while retaining a late committed result", async () => {
    let resolve!: (value: AssessmentAppendOutcome) => void;
    const { owner, send } = setup(
      vi.fn(
        () =>
          new Promise((done) => {
            resolve = done;
          }),
      ),
    );
    const { result, rerender } = renderHook(
      ({ key }) =>
        useAssessmentCreationController({
          owner,
          sheetRef,
          lifecycleResetKey: key,
        }),
      { initialProps: { key: "original" } },
    );
    act(() => {
      result.current.commands.openStandalone();
      result.current.commands.updateDraft((draft) => ({
        ...draft,
        subjectRecordId: "host",
        rationale: "Reviewed",
      }));
    });
    let pending!: Promise<void>;
    await act(async () => {
      pending = result.current.commands.submit(true);
    });
    await waitFor(() => expect(send).toHaveBeenCalledOnce());
    rerender({ key: "another selection" });
    await act(async () => {
      resolve(accepted);
      await pending;
    });
    expect(result.current.snapshot.feedback).toBeNull();
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
  });
});

describe("Assessment retained append owner", () => {
  it("reserves same-frame activation before authority reads and sends once", async () => {
    const { owner, prepare, send, ids } = setup();
    prepare();
    await Promise.all([owner.submit(binding), owner.submit(binding)]);
    expect(send).toHaveBeenCalledTimes(1);
    expect(ids.create).toHaveBeenCalledTimes(1);
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
  });
  it("replays uncertainty with the exact immutable body route and key then permits a deliberate identical append", async () => {
    const { owner, prepare, send, ids } = setup(
      vi.fn(async () => ({ kind: "uncertain" })),
    );
    prepare();
    await owner.submit(binding);
    const attempt = entryAt(owner).attempt;
    expect(Object.isFrozen(attempt.review.draft.values.supportRecordIds)).toBe(
      true,
    );
    expect(attempt.body).not.toContain("assessment.assessed_at");
    expect(attempt.body).not.toContain("assessment.assessor");
    owner.updateDraft((draft) => ({ ...draft, rationale: "different" }));
    expect(owner.getSnapshot().draft?.values.rationale).toBe("Judgment");
    send.mockResolvedValue(accepted);
    await owner.replay(attempt.clientTxnId);
    expect(send.mock.calls[1]?.[0]).toBe(attempt);
    expect(ids.create).toHaveBeenCalledTimes(1);
    prepare();
    await owner.submit(binding);
    expect(ids.create).toHaveBeenCalledTimes(2);
    const next = entryAt(owner, 1).attempt;
    expect(next.clientTxnId).not.toBe(attempt.clientTxnId);
    expect(next.review.draft.values).toEqual(attempt.review.draft.values);
  });
  it("retains acceptance before failed refresh and recovers with reads only", async () => {
    const refresh = vi.fn(async (): Promise<void> => {
      throw new Error("Read failed");
    });
    const { owner, prepare, send } = setup(undefined, refresh);
    prepare();
    await owner.submit(binding);
    await waitFor(() =>
      expect(owner.getSnapshot().entries[0]?.refresh).toBe("required"),
    );
    const entry = entryAt(owner);
    expect(entry.receipt).toEqual(receipt);
    expect(owner.getSnapshot().draft).toBeNull();
    refresh.mockResolvedValue(undefined);
    await owner.retryRefresh(entry.attempt.clientTxnId);
    expect(owner.getSnapshot().entries[0]?.refresh).toBe("complete");
    await owner.replay(entry.attempt.clientTxnId);
    expect(send).toHaveBeenCalledTimes(1);
  });
  it("cancels preparation before dispatch without losing intent", async () => {
    const { owner, prepare, readAuthority, send } = setup();
    prepare();
    let resolve!: (value: WorkbookMutationAuthority) => void;
    readAuthority.mockImplementation(
      () =>
        new Promise((done) => {
          resolve = done;
        }),
    );
    let current = true;
    const pending = owner.submit({ sheetRef, isCurrent: () => current });
    current = false;
    resolve(authority);
    await pending;
    expect(send).not.toHaveBeenCalled();
    expect(owner.getSnapshot().draft).not.toBeNull();
    expect(owner.pendingCount).toBe(0);
  });
  it("retains late detached acceptance while hiding suspended protected state and retires on account change", async () => {
    let resolve!: (value: AssessmentAppendOutcome) => void;
    const { owner, prepare, send, effect } = setup(
      vi.fn(
        () =>
          new Promise((done) => {
            resolve = done;
          }),
      ),
    );
    prepare();
    const pending = owner.submit(binding);
    await waitFor(() => expect(send).toHaveBeenCalledOnce());
    owner.suspend();
    resolve(accepted);
    await pending;
    expect(owner.getSnapshot().entries).toEqual([]);
    expect(effect).toHaveBeenCalledOnce();
    owner.setAuthority(authority);
    expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    owner.setAuthority({ ...authority, actorId: "other" });
    expect(owner.getSnapshot().entries).toEqual([]);
  });
  it("preserves an uncertain result after replay rejection and prevents fresh writes after role or closure changes", async () => {
    const { owner, prepare, send, readAuthority } = setup(
      vi.fn(async () => ({ kind: "uncertain" })),
    );
    prepare();
    await owner.submit(binding);
    const id = entryAt(owner).attempt.clientTxnId;
    send.mockResolvedValue({
      kind: "rejected",
      failure: { kind: "authorization_lost", message: "Subject unavailable" },
    });
    await owner.replay(id);
    expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
    readAuthority.mockResolvedValue({ ...authority, role: "viewer" });
    await owner.replay(id);
    expect(send).toHaveBeenCalledTimes(2);
    expect(owner.canSubmit()).toBe(false);
    owner.setAuthority({ ...authority, closed: true });
    expect(owner.canSubmit()).toBe(false);
    expect(owner.canReplay()).toBe(true);
  });
  it("observes a deadline without losing a late authoritative receipt", async () => {
    vi.useFakeTimers();
    try {
      let resolve!: (value: AssessmentAppendOutcome) => void;
      const { owner, prepare, send } = setup(
        vi.fn(
          () =>
            new Promise((done) => {
              resolve = done;
            }),
        ),
      );
      prepare();
      const pending = owner.submit(binding);
      await vi.advanceTimersByTimeAsync(30_001);
      await pending;
      expect(send).toHaveBeenCalledOnce();
      expect(owner.getSnapshot().entries[0]?.phase).toBe("uncertain");
      expect(owner.getSnapshot().entries[0]?.transportPending).toBe(true);
      await owner.replay(entryAt(owner).attempt.clientTxnId);
      expect(send).toHaveBeenCalledOnce();
      resolve(accepted);
      await vi.advanceTimersByTimeAsync(0);
      expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    } finally {
      vi.useRealTimers();
    }
  });
  it("keeps invalid input and secure identity failure editable without dispatch", async () => {
    const { owner, prepare, ids, send } = setup();
    prepare();
    owner.updateDraft((draft) => ({ ...draft, assessedAt: "unfinished" }));
    await owner.submit(binding);
    expect(send).not.toHaveBeenCalled();
    expect(ids.create).not.toHaveBeenCalled();
    expect(owner.getSnapshot().draft?.values.assessedAt).toBe("unfinished");
    owner.updateDraft((draft) => ({ ...draft, assessedAt: "" }));
    ids.create.mockImplementation(() => {
      throw new Error("Secure transaction identity unavailable.");
    });
    await owner.submit(binding);
    expect(send).not.toHaveBeenCalled();
    expect(owner.getSnapshot().draft?.values.supportRecordIds).toEqual([
      "support",
    ]);
  });
  it("keeps immutable receipts separate from newer rows in both HTTP and socket orders", async () => {
    for (const order of ["http_first", "socket_first"]) {
      const { owner, prepare } = setup();
      prepare();
      const latest = {
        ...receipt.data.row,
        row_version: 3,
        cells: { "assessment.rationale": { value: "Newer observation" } },
      };
      if (order === "socket_first") owner.acceptRow(latest);
      await owner.submit(binding);
      if (order === "http_first") owner.acceptRow(latest);
      expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
      expect(owner.latestRow("assessment")).toEqual(latest);
      expect(owner.acceptRow(receipt.data.row)).toEqual(latest);
      owner.acceptVersion("assessment", 4, true);
      expect(owner.acceptRow(latest)).toBeNull();
      expect(owner.getSnapshot().entries[0]?.receipt).toEqual(receipt);
    }
  });
  it("invalidates review only for selected subjects supports or origin and never retargets intent", () => {
    const { owner, prepare } = setup();
    prepare();
    const draft = owner.getSnapshot().draft;
    const review = owner.getSnapshot().reviewRevision;
    owner.observeCandidates("unrelated");
    expect(owner.getSnapshot().reviewRevision).toBe(review);
    owner.observeCandidates("support");
    expect(owner.getSnapshot().reviewRevision).toBeGreaterThan(review);
    expect(owner.getSnapshot().draft).toBe(draft);
    owner.discardDraft();
    const original = {
      ...receipt.data.row,
      cells: {
        "assessment.subject_ref": { value: "host" },
        "assessment.subject_type": { value: "host" },
      },
    };
    expect(owner.openFollowOn(original)).toBe(true);
    const seeded = owner.getSnapshot().draft;
    owner.acceptVersion(original.record_id, 2, true);
    expect(owner.getSnapshot().draft).toBe(seeded);
    owner.discardDraft();
    expect(owner.openFollowOn(original)).toBe(false);
  });
  it("distinguishes a local forbidden support from confirmed incident access loss", async () => {
    const { owner, prepare, readAuthority } = setup(
      vi.fn(async () => ({
        kind: "rejected",
        failure: { kind: "authorization_lost", message: "Support unavailable" },
      })),
    );
    prepare();
    const original = owner.getSnapshot().draft;
    await owner.submit(binding);
    await waitFor(() =>
      expect(readAuthority.mock.calls.length).toBeGreaterThan(1),
    );
    expect(owner.getSnapshot().draft).toBe(original);
    expect(owner.canSubmit()).toBe(true);
    readAuthority.mockImplementation(async () => {
      owner.suspend();
      throw new Error("Confirmed incident access loss");
    });
    await owner.recheckAuthority();
    expect(owner.getSnapshot().draft).toBeNull();
    expect(owner.getSnapshot().entries).toEqual([]);
    owner.setAuthority(authority);
    expect(owner.getSnapshot().draft).toBe(original);
  });
  it("does not resurrect retired attempts when an old transport completes", async () => {
    let resolve!: (value: AssessmentAppendOutcome) => void;
    const { owner, prepare, send, effect } = setup(
      vi.fn(
        () =>
          new Promise((done) => {
            resolve = done;
          }),
      ),
    );
    prepare();
    const pending = owner.submit(binding);
    await waitFor(() => expect(send).toHaveBeenCalledOnce());
    owner.retire();
    owner.setAuthority(authority);
    resolve(accepted);
    await pending;
    expect(owner.getSnapshot().entries).toEqual([]);
    expect(effect).not.toHaveBeenCalled();
  });
});
