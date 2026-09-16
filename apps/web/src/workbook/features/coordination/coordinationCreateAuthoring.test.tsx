import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  renderHook,
  screen,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { WorkbookAuthoringReferenceControl } from "../../components/WorkbookAuthoringReferenceControl";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";
import { CoordinationCreateContext } from "./CoordinationCreateContext";
import { CoordinationCreateForm } from "./CoordinationCreateForm";
import {
  type CoordinationVariant,
  coordinationFeature,
  coordinationReferenceView,
  coordinationSourceViews,
  coordinationVariants,
  prepareCoordination,
} from "./coordinationCreateModel";
import { useCoordinationCreateAttachment } from "./useCoordinationCreateAttachment";
import { WorkbookCoordinationCreateOwner } from "./WorkbookCoordinationCreateOwner";

afterEach(cleanup);
const sourceId = "00000000-0000-4000-8000-000000000001",
  memberId = "00000000-0000-4000-8000-000000000002";
const authority = {
  actorId: memberId,
  incidentId: "incident",
  sessionIdentity: "session",
  role: "editor" as const,
  closed: false,
};
const token = Symbol("test");
function fixture(
  variant: CoordinationVariant = "comm_log",
  view = coordinationSourceViews(variant)[0] as string,
) {
  const owner = new WorkbookCoordinationCreateOwner("incident");
  owner.setAuthority(authority);
  const reader: WorkbookAuthoringReadPort = {
    verify: vi.fn(async () => {}),
    availableViews: vi.fn(async () => ({
      kind: "accepted" as const,
      value: [...coordinationSourceViews(variant), "cartulary.view.parties.v1"],
    })),
    page: vi.fn(async (input) => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          {
            recordId:
              input.viewSchemaId === "incident_members" ? memberId : sourceId,
            displayText: "Available",
            viewSchemaId: input.viewSchemaId,
            row: { record_id: sourceId, row_version: 1, cells: {} },
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    })),
  };
  owner.configure(reader, async () => authority);
  const subject = {
    subject: {
      kind: "live" as const,
      recordId: sourceId,
      viewSchemaId: view,
      rowVersion: 1,
      label: "Original",
      surfaceLabel: requireViewContract(view).title,
    },
    cells: {},
  };
  const feature = required(coordinationFeature(view, variant));
  const sheet = { kind: "view_schema" as const, id: view };
  expect(owner.begin(subject, feature, sheet, token)).toBe(true);
  return { owner, reader, subject, feature, sheet };
}
const minima: Record<CoordinationVariant, Record<string, string>> = {
  comm_log: {
    "comm_log.comm_type": "briefing",
    "comm_log.audience": "Team",
    "comm_log.channel_or_meeting": "Call",
    "comm_log.summary": "Summary",
  },
  handoff: {
    "handoff.incoming_owner_user_id": memberId,
    "handoff.current_state_summary": "Current state",
  },
  status_review: { "status_review.current_state_summary": "Current state" },
  lesson: { "lesson.summary": "Lesson" },
};
function fill(
  owner: WorkbookCoordinationCreateOwner,
  variant: CoordinationVariant,
) {
  for (const [key, value] of Object.entries(minima[variant]))
    owner.update(key, value);
}
describe("Coordination authoring", () => {
  it("covers all twelve actions with independent source context and target minima", () => {
    let count = 0;
    for (const variant of coordinationVariants)
      for (const view of coordinationSourceViews(variant)) {
        count++;
        const { owner } = fixture(variant, view);
        expect(owner.getSnapshot().draft?.values).toEqual({});
        expect(owner.validate()).toBe(false);
        fill(owner, variant);
        expect(owner.validate()).toBe(true);
        const result = prepareCoordination(
          required(owner.getSnapshot().draft),
          "txn",
        );
        expect(result.request).toEqual({
          client_txn_id: "txn",
          "coordination.source_record_id": sourceId,
          ...minima[variant],
        });
        for (const key of Object.keys(minima[variant])) {
          owner.update(key, " \u2003");
          expect(owner.validate()).toBe(false);
          owner.update(key, minima[variant][key] as string);
        }
      }
    expect(count).toBe(12);
  });
  it("retains raw authoring through versions navigation detach replacement and clearing", () => {
    const { owner, subject, feature, sheet } = fixture();
    owner.discard();
    const hook = renderHook(
      ({ subject, open }) => useCoordinationCreateAttachment(subject, open),
      {
        initialProps: { subject, open: true },
        wrapper: ({ children }) => (
          <CoordinationCreateContext.Provider
            value={{ owner, sheetRef: sheet }}
          >
            {children}
          </CoordinationCreateContext.Provider>
        ),
      },
    );
    act(() => {
      hook.result.current.begin(feature);
      owner.update("comm_log.summary", " unfinished\r\n ");
    });
    hook.rerender({
      subject: { ...subject, subject: { ...subject.subject, rowVersion: 2 } },
      open: true,
    });
    expect(owner.getSnapshot().needsReview).toBe(true);
    expect(owner.getSnapshot().draft?.source?.rowVersion).toBe(1);
    hook.rerender({
      subject: {
        ...subject,
        subject: { ...subject.subject, recordId: memberId },
      },
      open: true,
    });
    expect(hook.result.current.workflow).toBeNull();
    expect(owner.getSnapshot().draft?.source?.recordId).toBe(sourceId);
    act(() => {
      expect(owner.begin(subject, feature, sheet, Symbol())).toBe(false);
      owner.resume(token);
      owner.changeSource({
        recordId: memberId,
        viewSchemaId: coordinationSourceViews("comm_log")[1] as string,
        rowVersion: 3,
        label: "Replacement",
      });
      owner.changeSource(null);
    });
    expect(owner.getSnapshot().draft?.values).toEqual({
      "comm_log.summary": " unfinished\r\n ",
    });
    expect(owner.getSnapshot().draft?.source).toBeNull();
    hook.unmount();
    owner.discard();
    expect(owner.begin(subject, feature, sheet, token)).toBe(true);
  });
  it("prepares every optional field and keeps defaults nulls risks and record references distinct", () => {
    for (const variant of coordinationVariants) {
      const { owner } = fixture(variant);
      fill(owner, variant);
      const target = required(owner.getSnapshot().draft).target;
      for (const field of target.fields.filter(
        (f) => f.createWritable && !Object.hasOwn(minima[variant], f.fieldKey),
      )) {
        const ref = coordinationReferenceView(field);
        owner.update(
          field.fieldKey,
          ref
            ? memberId
            : field.enumValues
              ? required(field.enumValues[0])
              : field.directScalarContractId
                ? "2026-09-13T14:00:00Z"
                : field.fieldKey === "handoff.open_risk_refs"
                  ? " Risk one\nRisk two "
                  : " Optional ",
        );
      }
      expect(owner.validate()).toBe(true);
      const request = required(
        prepareCoordination(required(owner.getSnapshot().draft), "txn").request,
      );
      for (const field of target.fields.filter((f) => f.createWritable))
        expect(request).toHaveProperty(field.fieldKey);
      if (variant === "handoff")
        expect(request).toHaveProperty("handoff.open_risk_refs", {
          kind: "collection_actions_v1",
          actions: [
            { op: "add_risk_ref", risk_ref_text: "Risk one" },
            { op: "add_risk_ref", risk_ref_text: "Risk two" },
          ],
        });
      for (const field of target.fields.filter(
        (f) => f.clearable && f.createWritable,
      )) {
        owner.update(field.fieldKey, null);
        expect(
          prepareCoordination(required(owner.getSnapshot().draft), "txn")
            .request,
        ).toHaveProperty(field.fieldKey, null);
        owner.omit(field.fieldKey);
        expect(
          prepareCoordination(required(owner.getSnapshot().draft), "txn")
            .request,
        ).not.toHaveProperty(field.fieldKey);
      }
      owner.changeSource(null);
      expect(
        prepareCoordination(required(owner.getSnapshot().draft), "txn").request,
      ).toHaveProperty("coordination.source_record_id", null);
    }
  });
  it("normalizes only prepared values and rejects controls overflow invalid references and timestamps", () => {
    const { owner } = fixture("status_review");
    owner.update(
      "status_review.current_state_summary",
      " \r\nCafe\u0301\rbody\t\u2003",
    );
    expect(
      prepareCoordination(required(owner.getSnapshot().draft), "txn").request,
    ).toHaveProperty("status_review.current_state_summary", "Café\nbody");
    expect(
      owner.getSnapshot().draft?.values["status_review.current_state_summary"],
    ).toContain("\r");
    for (const raw of ["bad\u0000", "x".repeat(16385), "\ud800"]) {
      owner.update("status_review.current_state_summary", raw);
      expect(owner.validate()).toBe(false);
      expect(
        owner.getSnapshot().draft?.values[
          "status_review.current_state_summary"
        ],
      ).toBe(raw);
    }
    fill(owner, "status_review");
    owner.update("status_review.blocked_task_ids", ` ${sourceId}`);
    expect(owner.validate()).toBe(false);
    owner.update("status_review.blocked_task_ids", `${sourceId}\n${sourceId}`);
    expect(owner.validate()).toBe(false);
    owner.omit("status_review.blocked_task_ids");
    for (const raw of [
      "tomorrow",
      "2026-02-30T10:00:00Z",
      "2026-09-13T24:00:00Z",
    ]) {
      owner.update("status_review.next_report_at", raw);
      expect(owner.validate()).toBe(false);
    }
    const comm = fixture().owner;
    fill(comm, "comm_log");
    comm.update("comm_log.audience", "");
    comm.update("comm_log.audience_party_ids", memberId);
    expect(comm.validate()).toBe(false);
  });
  it("reviews source changes and off-page references without treating page absence as deletion", async () => {
    const { owner, reader } = fixture("lesson", "cartulary.view.timeline.v2");
    fill(owner, "lesson");
    owner.update("lesson.follow_up_task_ids", memberId);
    owner.observe(sourceId, 2);
    vi.mocked(reader.page).mockImplementation(async (input) => ({
      kind: "accepted",
      value: {
        candidates:
          input.viewSchemaId.includes("task_requests") && !input.cursor
            ? []
            : [
                {
                  recordId: input.viewSchemaId.includes("task_requests")
                    ? memberId
                    : sourceId,
                  displayText: "Available",
                  viewSchemaId: input.viewSchemaId,
                  row: { record_id: sourceId, row_version: 2, cells: {} },
                },
              ],
        hasMore: input.viewSchemaId.includes("task_requests") && !input.cursor,
        nextCursor: input.cursor ? null : "next",
      },
    }));
    await owner.review();
    expect(owner.getSnapshot().needsReview).toBe(false);
    expect(owner.getSnapshot().draft?.source?.rowVersion).toBe(2);
    expect(reader.page).toHaveBeenCalledWith(
      expect.objectContaining({ cursor: "next" }),
    );
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    owner.observe();
    await owner.review();
    expect(owner.getSnapshot().needsReview).toBe(true);
    expect(owner.getSnapshot().message).toContain("Source is unavailable");
    expect(owner.getSnapshot().draft?.values["lesson.follow_up_task_ids"]).toBe(
      memberId,
    );
  });
  it("conceals on scoped suspension and retires on account replacement", () => {
    const { owner } = fixture();
    owner.update("comm_log.summary", "Private");
    owner.suspend();
    expect(owner.getSnapshot().draft).toBeNull();
    owner.update("comm_log.summary", "Obsolete");
    owner.setAuthority({ ...authority, sessionIdentity: "renewed" });
    expect(owner.getSnapshot().draft?.values["comm_log.summary"]).toBe(
      "Private",
    );
    owner.closeIncident();
    expect(owner.canSubmit()).toBe(false);
    owner.setAuthority(authority);
    expect(owner.getSnapshot().needsReview).toBe(true);
    owner.setAuthority({ ...authority, actorId: "replacement" });
    expect(owner.getSnapshot().draft).toBeNull();
  });
  it("fences obsolete presentation callbacks through attachment replacement and authority recovery", () => {
    const { owner, subject, feature, sheet } = fixture("lesson");
    const old = owner.captureDraftActions(token);
    old.update("lesson.summary", "Original");
    const recovery = Symbol("recovery");
    owner.resume(recovery);
    old.update("lesson.summary", "Detached callback");
    old.changeSource(null);
    old.discard();
    expect(owner.getSnapshot().draft?.values["lesson.summary"]).toBe(
      "Original",
    );
    expect(owner.getSnapshot().draft?.source?.recordId).toBe(sourceId);
    const suspended = owner.captureDraftActions(recovery);
    owner.suspend();
    owner.setAuthority({ ...authority, sessionIdentity: "renewed" });
    expect(owner.getSnapshot().attachment).toBeNull();
    owner.resume(recovery);
    suspended.discard();
    suspended.update("lesson.summary", "Old session");
    expect(owner.getSnapshot().draft?.values["lesson.summary"]).toBe(
      "Original",
    );
    const replaced = owner.captureDraftActions(recovery);
    owner.discard();
    owner.begin(subject, feature, sheet, recovery);
    replaced.changeSource(null);
    replaced.update("lesson.summary", "Old draft");
    replaced.discard();
    expect(owner.getSnapshot().draft?.values).toEqual({});
    expect(owner.getSnapshot().draft?.source?.recordId).toBe(sourceId);
    owner.setAuthority({ ...authority, role: "viewer" });
    expect(owner.canSubmit()).toBe(false);
    expect(owner.canReplay()).toBe(false);
    expect(owner.getSnapshot().draft).not.toBeNull();
    owner.setAuthority(authority);
    expect(owner.getSnapshot().needsReview).toBe(true);
  });
  it("shows field-local errors and restores focus while retaining a closed draft", () => {
    const { owner } = fixture("lesson");
    render(
      <>
        <button type="button">Trigger</button>
        <CoordinationCreateForm
          owner={owner}
          attachment={token}
          onSubmit={vi.fn()}
        />
      </>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Create Lesson" }));
    const field = screen.getByLabelText("Summary (required)");
    expect(field.getAttribute("aria-invalid")).toBe("true");
    expect(document.activeElement).toBe(field);
    fireEvent.change(field, { target: { value: "Draft" } });
    fireEvent.click(screen.getByRole("button", { name: "Close draft" }));
    expect(owner.getSnapshot().draft?.values["lesson.summary"]).toBe("Draft");
    expect(owner.getSnapshot().attachment).toBeNull();
  });
  it("stages paged target references with retained selection and keyboard cancel", async () => {
    const { reader } = fixture();
    const onApply = vi.fn();
    vi.mocked(reader.page).mockImplementation(async (input) => ({
      kind: "accepted",
      value: {
        candidates: [
          {
            recordId: input.cursor ? memberId : sourceId,
            displayText: input.cursor ? "Second" : "First",
            viewSchemaId: input.viewSchemaId,
          },
        ],
        hasMore: !input.cursor,
        nextCursor: input.cursor ? null : "next",
      },
    }));
    render(
      <WorkbookAuthoringReferenceControl
        targetKey="coordination-test"
        maximum={64}
        label="Audience Parties"
        testId="party-picker"
        views={["cartulary.view.parties.v1"]}
        multiple
        selected={[
          {
            recordId: "retained",
            displayText: "Retained",
            viewSchemaId: "cartulary.view.parties.v1",
          },
        ]}
        reader={reader}
        revision={0}
        disabled={false}
        onApply={onApply}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Choose audience parties" }),
    );
    await screen.findByRole("option", { name: "First" });
    expect(
      screen.getByRole("button", {
        name: "Remove selected Audience Parties Retained",
      }),
    ).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Next candidates" }));
    await screen.findByRole("option", { name: "Second" });
    fireEvent.keyDown(screen.getByLabelText("Audience Parties"), {
      key: "Escape",
    });
    expect(onApply).not.toHaveBeenCalled();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Choose audience parties" }),
    );
  });
  it("distinguishes failed discovery empty results and unavailable surfaces", async () => {
    const { reader } = fixture();
    vi.mocked(reader.availableViews).mockRejectedValueOnce(
      new Error("offline"),
    );
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    render(
      <WorkbookAuthoringReferenceControl
        targetKey="coordination-test"
        maximum={64}
        label="Owner"
        testId="owner-picker"
        views={["cartulary.view.parties.v1"]}
        multiple={false}
        selected={[]}
        reader={reader}
        revision={0}
        disabled={false}
        onApply={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Choose owner" }));
    await screen.findByRole("button", { name: "Retry surfaces" });
    expect(screen.queryByText("No candidates match this query.")).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Retry surfaces" }));
    await screen.findByText("No candidates match this query.");
    cleanup();
    vi.mocked(reader.availableViews).mockResolvedValue({
      kind: "accepted",
      value: [],
    });
    render(
      <WorkbookAuthoringReferenceControl
        targetKey="coordination-test"
        maximum={64}
        label="Parties"
        testId="unavailable-parties"
        views={["cartulary.view.parties.v1"]}
        multiple
        selected={[
          {
            recordId: sourceId,
            displayText: "Retained party",
            viewSchemaId: "cartulary.view.parties.v1",
          },
        ]}
        reader={reader}
        revision={1}
        disabled={false}
        onApply={vi.fn()}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Choose parties" }));
    await screen.findByText(
      "This reference surface is unavailable. Your selection is retained.",
    );
    expect(
      screen.getByRole("button", {
        name: "Remove selected Parties Retained party",
      }),
    ).toBeTruthy();
    expect(
      screen
        .getByRole("button", { name: "Apply references" })
        .hasAttribute("disabled"),
    ).toBe(false);
  });
});
function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Required fixture");
  return value;
}
