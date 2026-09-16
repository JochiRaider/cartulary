import {
  listViewContracts,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import { ContextualCreateForm } from "./ContextualCreateForm";
import {
  contextualCreateErrors,
  contextualCreateRequest,
  isContextualCreateFeature,
} from "./contextualCreateModel";
import type { ContextualCreateReader } from "./contextualCreateOperation";
import { WorkbookContextualTaskDecisionCreateOwner } from "./WorkbookContextualTaskDecisionCreateOwner";

afterEach(cleanup);
const actor = "10000000-0000-4000-8000-000000000001",
  sourceId = "20000000-0000-4000-8000-000000000002";
const authority: WorkbookMutationAuthority = {
  actorId: actor,
  incidentId: "incident",
  sessionIdentity: "session",
  role: "editor",
  closed: false,
};
const attachment = Symbol("test");
function fixture(
  view: string = timelineViewSchemaId,
  key = "create_related.task_request",
) {
  const owner = new WorkbookContextualTaskDecisionCreateOwner(
    authority.incidentId,
    { create: () => "txn" },
    {
      coordinate: async () => ({
        kind: "settled" as const,
        minimumRowVersion: 0,
      }),
      accepted: () => {},
      refresh: async () => {},
      observed: () => {},
    },
  );
  owner.setAuthority(authority);
  const contract = requireViewContract(view);
  const feature = contract.inspectorConfig.featureGroups.find(
    (feature) => feature.featureGroupKey === key,
  );
  if (!feature) throw new Error("Expected declared feature");
  const subject = {
    cells: {},
    subject: {
      kind: "live" as const,
      recordId: sourceId,
      viewSchemaId: view,
      rowVersion: 1,
      label: "Original source",
      surfaceLabel: contract.title,
    },
  };
  const reader: ContextualCreateReader = {
    availableViews: vi.fn(async () => ({
      kind: "accepted" as const,
      value: [view],
    })),
    verify: vi.fn(async () => {}),
    page: vi.fn(async () => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          {
            recordId: sourceId,
            displayText: "Source",
            viewSchemaId: contract.viewSchemaId,
            row: fullWorkbookViewRow(contract, sourceId, 1, {}),
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    })),
  };
  owner.configure(reader, async () => authority);
  owner.begin(subject, feature, { kind: "view_schema", id: view }, attachment);
  return { owner, feature, subject, reader };
}
describe("contextual Task and Decision authoring", () => {
  it("maps all 23 declared entry points across 15 public surfaces and excludes absent features", () => {
    const surfaces = listViewContracts().filter((contract) =>
      contract.inspectorConfig.featureGroups.some((feature) =>
        isContextualCreateFeature(feature.featureGroupKey),
      ),
    );
    expect(surfaces).toHaveLength(15);
    let count = 0;
    for (const contract of surfaces)
      for (const feature of contract.inspectorConfig.featureGroups.filter(
        (feature) => isContextualCreateFeature(feature.featureGroupKey),
      )) {
        const { owner } = fixture(
          contract.viewSchemaId,
          feature.featureGroupKey,
        );
        const draft = required(owner.getSnapshot().draft);
        expect(draft.target.viewSchemaId).toBe(
          feature.routeBinding.targetViewSchemaId,
        );
        expect(
          draft.values[
            required(required(feature.seedBindings[0]).targetFieldKey)
          ],
        ).toBe(sourceId);
        expect(contextualCreateRequest(draft, "txn")).toBeNull();
        count++;
      }
    expect(count).toBe(23);
    for (const id of [
      "cartulary.view.task_requests.v1",
      "cartulary.view.parties.v1",
    ])
      expect(
        requireViewContract(id).inspectorConfig.featureGroups.some((feature) =>
          isContextualCreateFeature(feature.featureGroupKey),
        ),
      ).toBe(false);
  });
  it("retains identity and edits across detachment, source versions and explicit replacement", async () => {
    const { owner, subject, feature } = fixture();
    owner.update("task.title", "Analyst draft");
    owner.update("task.linked_record_ids", "");
    owner.detach(attachment);
    owner.observe(sourceId, 2);
    expect(owner.getSnapshot()).toMatchObject({
      attachment: null,
      needsReview: true,
      draft: {
        source: { recordId: sourceId, rowVersion: 1 },
        values: { "task.title": "Analyst draft", "task.linked_record_ids": "" },
      },
    });
    expect(
      owner.begin(
        { ...subject, subject: { ...subject.subject, recordId: actor } },
        feature,
        { kind: "view_schema", id: timelineViewSchemaId },
        attachment,
      ),
    ).toBe(false);
    owner.resume(attachment);
    owner.update("task.task_kind", "investigation");
    owner.discard();
    expect(owner.getSnapshot().draft).toBeNull();
    expect(
      owner.begin(
        subject,
        feature,
        { kind: "view_schema", id: timelineViewSchemaId },
        attachment,
      ),
    ).toBe(true);
    for (const target of ["task_request", "decision"]) {
      const { owner: retained } = fixture(
        timelineViewSchemaId,
        `create_related.${target}`,
      );
      const field =
        target === "task_request" ? "task.title" : "decision.summary";
      retained.update(field, "Keep this exact draft");
      const before = required(retained.getSnapshot().draft);
      // A merge removes the original source projection. It does not authorize
      // rebinding its contextual draft to the surviving record.
      retained.observeSocket({
        type: "record_changed",
        incident_id: authority.incidentId,
        event_id: "source-removed-by-merge",
        emitted_at: "2026-09-12T20:00:00Z",
        stream_seq: 1,
        payload: {
          record_id: sourceId,
          row_version: 2,
          client_txn_id: "merge-operation",
          actor_user_id: actor,
          change_set_id: "merge-change-set",
          changed_field_keys: [],
          affected_views: [
            { view_schema_id: timelineViewSchemaId, change_kind: "remove" },
          ],
        },
      });
      expect(retained.getSnapshot().draft?.source).toEqual(before.source);
      expect(retained.getSnapshot().draft?.values).toEqual(before.values);
      expect(retained.getSnapshot().needsReview).toBe(true);
    }
  });
  it("requires target minima, omits unset optional fields, and guards initial lifecycle and exact references", () => {
    const { owner } = fixture();
    expect(owner.getSnapshot().draft?.values).toMatchObject({
      "task.status": "open",
      "task.priority": "normal",
      "task.owner_user_id": actor,
    });
    const kind = "follow_up";
    owner.update("task.title", "  Title  ");
    owner.update("task.task_kind", kind);
    expect(
      contextualCreateRequest(required(owner.getSnapshot().draft), "txn"),
    ).toMatchObject({
      "task.title": "Title",
      "task.task_kind": kind,
      client_txn_id: "txn",
    });
    expect(
      contextualCreateRequest(required(owner.getSnapshot().draft), "txn"),
    ).not.toHaveProperty("task.completed_at");
    owner.update("task.requester_party_id", ` ${actor}`);
    expect(
      contextualCreateErrors(required(owner.getSnapshot().draft)),
    ).toHaveProperty("task.requester_party_id");
    owner.update("task.requester_party_id", "");
    owner.update("task.status", "blocked");
    expect(
      contextualCreateErrors(required(owner.getSnapshot().draft)),
    ).toHaveProperty("task.blocked_reason");
    const { owner: decision } = fixture(
      timelineViewSchemaId,
      "create_related.decision",
    );
    expect(decision.getSnapshot().draft?.values["decision.status"]).toBe(
      "proposed",
    );
    decision.update("decision.status", "superseded");
    expect(
      contextualCreateErrors(required(decision.getSnapshot().draft)),
    ).toHaveProperty("decision.status");
    expect(
      contextualCreateErrors(required(decision.getSnapshot().draft)),
    ).toHaveProperty("decision.rationale");
  });
  it("stages reference changes, preserves off-page selections, pages and cancels without editing the draft", async () => {
    const { owner, reader } = fixture();
    vi.mocked(reader.page)
      .mockResolvedValueOnce({
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: actor,
              displayText: "Other record",
              viewSchemaId: timelineViewSchemaId,
            },
          ],
          hasMore: true,
          nextCursor: "page2",
        },
      })
      .mockResolvedValue({
        kind: "accepted",
        value: { candidates: [], hasMore: false, nextCursor: null },
      });
    render(
      <ContextualCreateForm
        owner={owner}
        attachment={attachment}
        onSubmit={() => {}}
      />,
    );
    const label = required(
      required(owner.getSnapshot().draft).target.fieldMap[
        "task.linked_record_ids"
      ],
    ).label;
    fireEvent.click(screen.getByRole("button", { name: `Choose ${label}` }));
    await waitFor(() =>
      expect(
        screen.getByRole("listbox", { name: label }).hasAttribute("disabled"),
      ).toBe(false),
    );
    expect(
      screen.getByRole("button", {
        name: `Remove selected ${label} Original source`,
      }),
    ).toBeDefined();
    fireEvent.click(screen.getByRole("button", { name: "Next candidates" }));
    await waitFor(() =>
      expect(reader.page).toHaveBeenCalledWith(
        expect.objectContaining({ cursor: "page2" }),
      ),
    );
    fireEvent.keyDown(screen.getByRole("region", { name: `Choose ${label}` }), {
      key: "Escape",
    });
    expect(owner.getSnapshot().draft?.values["task.linked_record_ids"]).toBe(
      sourceId,
    );
    fireEvent.click(
      screen.getByRole("button", { name: `Remove ${label} Original source` }),
    );
    expect(owner.getSnapshot().draft?.values["task.linked_record_ids"]).toBe(
      "",
    );
  });
  it("hides protected state on suspension and retires it on account replacement", () => {
    const { owner } = fixture();
    owner.update("task.title", "Retained");
    owner.suspend();
    expect(owner.getSnapshot().draft).toBeNull();
    owner.setAuthority(authority);
    expect(owner.getSnapshot().draft?.values["task.title"]).toBe("Retained");
    expect(owner.getSnapshot().draft?.labels).toEqual({});
    owner.setAuthority({ ...authority, actorId: sourceId });
    expect(owner.getSnapshot().draft).toBeNull();
  });
  it("retains both targets through stale failed reference pages and role changes without exposing old labels", async () => {
    for (const decision of [false, true]) {
      const { owner, reader } = fixture(
        timelineViewSchemaId,
        decision ? "create_related.decision" : "create_related.task_request",
      );
      const key = decision ? "decision.support_refs" : "task.linked_record_ids";
      const label = required(
        required(owner.getSnapshot().draft).target.fieldMap[key],
      ).label;
      const input = decision ? "decision.summary" : "task.title";
      owner.update(input, "Retained analyst input");
      vi.mocked(reader.page).mockResolvedValueOnce({
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: sourceId,
              displayText: "Original source",
              viewSchemaId: timelineViewSchemaId,
            },
          ],
          hasMore: true,
          nextCursor: "next",
        },
      });
      let failRead: (() => void) | undefined;
      vi.mocked(reader.page).mockImplementation(
        () =>
          new Promise((resolve) => {
            failRead = () =>
              resolve({
                kind: "rejected",
                failure: {
                  kind: "retryable",
                  message: "Reference read failed",
                },
              });
          }),
      );
      const rendered = render(
        <ContextualCreateForm
          owner={owner}
          attachment={attachment}
          onSubmit={() => {}}
        />,
      );
      fireEvent.click(screen.getByRole("button", { name: `Choose ${label}` }));
      await waitFor(() =>
        expect(
          screen.getByRole("button", {
            name: `Remove selected ${label} Original source`,
          }),
        ).toBeDefined(),
      );
      act(() => owner.observe(sourceId, 2));
      await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(2));
      expect(screen.queryByText("Original source", { exact: true })).toBeNull();
      expect(owner.getSnapshot().draft?.values[key]).toBe(sourceId);
      await act(async () => failRead?.());
      await waitFor(() =>
        expect(screen.getByRole("alert").textContent).toContain(
          "Reference read failed",
        ),
      );
      expect(owner.getSnapshot().needsReview).toBe(true);
      act(() => owner.setAuthority({ ...authority, role: "viewer" }));
      expect(rendered.container.querySelector("fieldset")?.disabled).toBe(true);
      expect(owner.getSnapshot().draft?.values[input]).toBe(
        "Retained analyst input",
      );
      act(() => owner.suspend());
      expect(rendered.container.textContent).toBe("");
      rendered.unmount();
    }
  });
});

function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Required test value missing");
  return value;
}
