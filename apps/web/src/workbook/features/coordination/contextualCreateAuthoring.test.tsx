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
  within,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { WorkbookRecoveryBoundary } from "../../../shared/WorkbookRecoveryBoundary";
import { WorkbookRecoveryNavigation } from "../../../shared/workbookRecoveryNavigation";
import { deferred } from "../../../testing/fetchMockTestSupport";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { createContextualCreateTransport } from "../../adapters/createContextualCreateTransport";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import { ContextualCreateForm } from "./ContextualCreateForm";
import { ContextualCreateRecovery } from "./ContextualCreateRecovery";
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
  const retainedSourceJourney = async (target: "task_request" | "decision") => {
    const { owner, reader, subject, feature } = fixture(
      timelineViewSchemaId,
      `create_related.${target}`,
    );
    const transport = {
      ...createContextualCreateTransport(undefined),
      send: vi.fn(),
    };
    owner.configure(reader, async () => authority, transport);
    const scalar = target === "decision" ? "decision.summary" : "task.title";
    const references =
      target === "decision"
        ? ["decision.support_refs", "decision.affected_record_ids"]
        : ["task.linked_record_ids"];
    const other = "30000000-0000-4000-8000-000000000003";
    for (const key of references)
      owner.update(key, `${sourceId}\n${other}`, {
        [other]: "Other reference",
      });
    const original = required(owner.getSnapshot().draft);
    const navigation = new WorkbookRecoveryNavigation();
    const detailHost = document.createElement("div");
    document.body.append(detailHost);
    const rendered = render(
      <WorkbookRecoveryBoundary navigation={navigation} detailHost={detailHost}>
        <ContextualCreateForm
          owner={owner}
          attachment={attachment}
          onSubmit={() => void owner.submit(attachment)}
        />
        <ContextualCreateRecovery owner={owner} />
      </WorkbookRecoveryBoundary>,
    );
    try {
      for (const key of references) {
        const label = required(original.target.fieldMap[key]).label;
        const remove = screen.getByRole("button", {
          name: `Remove ${label} Original source`,
        });
        remove.focus();
        fireEvent.click(remove);
        expect(document.activeElement).toBe(
          screen.getByRole("button", {
            name: `Remove ${label} Other reference`,
          }),
        );
      }
      act(() => owner.update(scalar, "  Exact retained authoring  "));
      const values = {
        ...original.values,
        [scalar]: "  Exact retained authoring  ",
      };
      for (const key of references) values[key] = other;
      expect(owner.getSnapshot().draft?.source).toEqual(original.source);
      expect(owner.getSnapshot().draft?.values).toEqual(values);
      expect(Object.keys(required(owner.getSnapshot().draft).labels)).toEqual([
        other,
      ]);
      expect(screen.getByText(/^Create in .* Source:/).textContent).toContain(
        "Source: Original source (Timeline)",
      );
      fireEvent.click(
        screen.getByRole("button", { name: "Keep draft and close" }),
      );
      act(() => {
        expect(
          owner.begin(
            {
              ...subject,
              subject: {
                ...subject.subject,
                recordId: other,
                label: "Different source",
              },
            },
            feature,
            { kind: "view_schema", id: timelineViewSchemaId },
            attachment,
          ),
        ).toBe(false);
        navigation.openList();
        navigation.activate(required(navigation.getSnapshot().entries[0]).key);
      });
      expect(navigation.getSnapshot().entries[0]?.origin).toBe(
        "Original source",
      );
      fireEvent.click(
        screen.getByRole("button", { name: "Resume contextual draft" }),
      );
      expect(screen.getByText(/^Create in .* Source:/).textContent).toContain(
        "Source: Original source (Timeline)",
      );
      expect(owner.getSnapshot().draft?.source).toEqual(original.source);
      expect(owner.getSnapshot().draft?.values).toEqual(values);
      expect(transport.send).not.toHaveBeenCalled();
      expect(reader.verify).not.toHaveBeenCalled();
      expect(reader.page).not.toHaveBeenCalled();
    } finally {
      rendered.unmount();
      navigation.dispose();
      detailHost.remove();
    }
  };
  it("retains readable original source after removing every task_request target link and resuming through Recovery", () =>
    retainedSourceJourney("task_request"));
  it("retains readable original source after removing every decision target link and resuming through Recovery", () =>
    retainedSourceJourney("decision"));
  it("recovers original source presentation only through current source review independently of target labels and minima", async () => {
    for (const target of ["task_request", "decision"] as const) {
      const { owner, reader } = fixture(
        timelineViewSchemaId,
        `create_related.${target}`,
      );
      const field =
        target === "decision"
          ? "decision.support_refs"
          : "task.linked_record_ids";
      const other = "30000000-0000-4000-8000-000000000003";
      owner.update(field, other, { [other]: "Other reference" });
      const rendered = render(
        <ContextualCreateForm
          owner={owner}
          attachment={attachment}
          onSubmit={vi.fn()}
        />,
      );
      const sourceText = () =>
        screen.getByText(/^Create in .* Source:/).textContent;
      act(() => owner.observe(other, 2));
      expect(sourceText()).toContain("Source: Original source (Timeline)");
      expect(owner.getSnapshot().needsReview).toBe(true);
      act(() => owner.observe(sourceId, 2));
      expect(sourceText()).toContain(
        "Source: Original source needs review (Timeline)",
      );
      act(() =>
        owner.update(field, `${sourceId}\n${other}`, {
          [sourceId]: "Candidate presentation",
        }),
      );
      expect(sourceText()).not.toContain("Candidate presentation");
      expect(owner.getSnapshot().needsReview).toBe(true);
      const values = required(owner.getSnapshot().draft).values;
      vi.mocked(reader.page).mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "retryable", message: "Unavailable" },
      });
      await act(async () => {
        expect(await owner.review()).toBe(false);
      });
      expect(sourceText()).toContain("Original source needs review");
      vi.mocked(reader.page).mockResolvedValueOnce({
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: sourceId,
              viewSchemaId: timelineViewSchemaId,
              displayText: "Candidate hint is not the reviewed label",
              row: fullWorkbookViewRow(
                requireViewContract(timelineViewSchemaId),
                sourceId,
                2,
                {
                  "timeline.activity_synopsis_text": "Reviewed current source",
                },
              ),
            },
          ],
          hasMore: false,
          nextCursor: null,
        },
      });
      await act(async () => {
        expect(await owner.review()).toBe(false);
      });
      expect(sourceText()).toContain(
        "Source: Reviewed current source (Timeline)",
      );
      expect(owner.getSnapshot().draft?.source.rowVersion).toBe(2);
      expect(owner.getSnapshot().draft?.values).toEqual(values);
      expect(
        contextualCreateRequest(required(owner.getSnapshot().draft), "txn"),
      ).toBeNull();
      expect(Object.keys(owner.getSnapshot().errors).length).toBeGreaterThan(0);
      rendered.unmount();
    }
  });
  it("conceals original source labels during suspension and requires a current read after same-account recovery", async () => {
    for (const target of ["task_request", "decision"] as const) {
      const { owner, reader } = fixture(
        timelineViewSchemaId,
        `create_related.${target}`,
      );
      const scalar = target === "decision" ? "decision.summary" : "task.title";
      owner.update(scalar, "Retained raw input");
      const navigation = new WorkbookRecoveryNavigation();
      const rendered = render(
        <WorkbookRecoveryBoundary navigation={navigation} detailHost={null}>
          <ContextualCreateForm
            owner={owner}
            attachment={attachment}
            onSubmit={vi.fn()}
          />
          <ContextualCreateRecovery owner={owner} />
        </WorkbookRecoveryBoundary>,
      );
      expect(navigation.getSnapshot().entries[0]?.origin).toBe(
        "Original source",
      );
      act(() => owner.suspend());
      expect(rendered.container.textContent).toBe("");
      expect(navigation.getSnapshot().entries).toEqual([]);
      act(() => owner.setAuthority(authority));
      expect(navigation.getSnapshot().entries[0]?.origin).toBe(
        "Original source needs review",
      );
      expect(rendered.container.textContent).not.toContain(
        "Source: Original source (Timeline)",
      );
      expect(owner.getSnapshot().draft?.values[scalar]).toBe(
        "Retained raw input",
      );
      vi.mocked(reader.page).mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "retryable", message: "Read failed" },
      });
      await act(async () => {
        await owner.review();
      });
      expect(navigation.getSnapshot().entries[0]?.origin).toBe(
        "Original source needs review",
      );
      act(() => owner.setAuthority({ ...authority, role: "viewer" }));
      expect(owner.canSubmit()).toBe(false);
      expect(navigation.getSnapshot().entries[0]?.origin).toBe(
        "Original source needs review",
      );
      act(() => owner.setAuthority({ ...authority, actorId: sourceId }));
      expect(navigation.getSnapshot().entries).toEqual([]);
      expect(owner.getSnapshot().draft).toBeNull();
      rendered.unmount();
      navigation.dispose();
    }
  });
  it("ignores obsolete successful and failed source reviews after newer authoring or incident retirement", async () => {
    for (const result of ["accepted", "failed", "retired"] as const) {
      const { owner, reader } = fixture();
      owner.observe(sourceId, 2);
      const pending =
        deferred<Awaited<ReturnType<ContextualCreateReader["page"]>>>();
      vi.mocked(reader.page).mockReturnValueOnce(pending.promise);
      const reviewing = owner.review();
      await waitFor(() => expect(reader.page).toHaveBeenCalledOnce());
      if (result === "retired") owner.retire();
      else owner.update("task.title", "Newer exact authoring");
      const before = owner.getSnapshot();
      pending.resolve(
        result === "failed"
          ? {
              kind: "rejected",
              failure: { kind: "retryable", message: "Obsolete failure" },
            }
          : {
              kind: "accepted",
              value: {
                candidates: [
                  {
                    recordId: sourceId,
                    viewSchemaId: timelineViewSchemaId,
                    displayText: "Obsolete",
                    row: fullWorkbookViewRow(
                      requireViewContract(timelineViewSchemaId),
                      sourceId,
                      2,
                      {
                        "timeline.activity_synopsis_text":
                          "Obsolete source label",
                      },
                    ),
                  },
                ],
                hasMore: false,
                nextCursor: null,
              },
            },
      );
      expect(await reviewing).toBe(false);
      expect(owner.getSnapshot()).toBe(before);
    }
  });
  it("reconciles the default Owner only in staging until Apply and retains its readable label on resume", async () => {
    const { owner, reader } = fixture();
    owner.update("task.title", "Title remains exactly authored");
    const original = required(owner.getSnapshot().draft);
    const onSubmit = vi.fn();
    let accept!: (
      value: Awaited<ReturnType<ContextualCreateReader["page"]>>,
    ) => void;
    vi.mocked(reader.page).mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          accept = resolve;
        }),
    );
    const rendered = render(
      <ContextualCreateForm
        owner={owner}
        attachment={attachment}
        onSubmit={onSubmit}
      />,
    );
    expect(
      screen.getByText("Current actor", { exact: false }).textContent,
    ).toContain("Current actor");
    fireEvent.click(screen.getByRole("button", { name: "Choose Owner" }));
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(1));
    expect(owner.getSnapshot().draft).toBe(original);
    await act(async () =>
      accept({
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: actor,
              displayText: "Review editor",
              viewSchemaId: "incident_members",
            },
          ],
          hasMore: false,
          nextCursor: null,
        },
      }),
    );
    const picker = screen.getByRole("region", { name: "Choose Owner" });
    const selector = within(picker).getByRole("radio", {
      name: `Review editor (${actor})`,
    });
    expect(selector).toHaveProperty("checked", true);
    expect(within(picker).getByRole("list").textContent).toContain(
      "Review editor",
    );
    expect(
      within(picker)
        .getByRole("button", { name: /^Remove selected Owner / })
        .getAttribute("aria-label"),
    ).toBe("Remove selected Owner Review editor");
    expect(owner.getSnapshot().draft).toBe(original);
    fireEvent.keyDown(selector, { key: "Escape" });
    expect(owner.getSnapshot().draft).toBe(original);
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Choose Owner" }),
    );
    vi.mocked(reader.page).mockResolvedValueOnce({
      kind: "accepted",
      value: {
        candidates: [
          {
            recordId: actor,
            displayText: "Review editor",
            viewSchemaId: "incident_members",
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    });
    fireEvent.click(screen.getByRole("button", { name: "Choose Owner" }));
    await screen.findByRole("radio", { name: `Review editor (${actor})` });
    fireEvent.click(screen.getByRole("button", { name: "Apply references" }));
    expect(owner.getSnapshot().draft).toMatchObject({
      values: {
        "task.owner_user_id": actor,
        "task.title": "Title remains exactly authored",
      },
      labels: { [actor]: "Review editor" },
      source: original.source,
    });
    expect(
      screen.getByRole("button", { name: "Remove Owner Review editor" }),
    ).toBeTruthy();
    fireEvent.click(
      screen.getByRole("button", { name: "Keep draft and close" }),
    );
    const recovery = Symbol("recovery");
    act(() => owner.resume(recovery));
    rendered.rerender(
      <ContextualCreateForm
        owner={owner}
        attachment={recovery}
        onSubmit={onSubmit}
      />,
    );
    expect(
      screen.getByRole("button", { name: "Remove Owner Review editor" }),
    ).toBeTruthy();
    expect(owner.getSnapshot().draft?.values["task.title"]).toBe(
      "Title remains exactly authored",
    );
    expect(owner.getSnapshot().draft?.values["task.owner_user_id"]).toBe(actor);
    expect(onSubmit).not.toHaveBeenCalled();
    expect(reader.verify).not.toHaveBeenCalled();
  });

  it("recovers refreshed reference labels without restoring source readiness or replacing authored values", async () => {
    const { owner, reader } = fixture();
    owner.update("task.title", "Retained raw title");
    const original = required(owner.getSnapshot().draft);
    let accept!: (
      value: Awaited<ReturnType<ContextualCreateReader["page"]>>,
    ) => void;
    vi.mocked(reader.page)
      .mockResolvedValueOnce({
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: sourceId,
              displayText: "Original source",
              viewSchemaId: timelineViewSchemaId,
            },
          ],
          hasMore: false,
          nextCursor: null,
        },
      })
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            accept = resolve;
          }),
      );
    render(
      <ContextualCreateForm
        owner={owner}
        attachment={attachment}
        onSubmit={vi.fn()}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", {
        name: "Choose Linked Records",
      }),
    );
    await screen.findByRole("checkbox", {
      name: `Original source (${sourceId})`,
    });
    act(() => owner.observe(sourceId, 2));
    await waitFor(() => expect(reader.page).toHaveBeenCalledTimes(2));
    const picker = screen.getByRole("region", {
      name: "Choose Linked Records",
    });
    expect(within(picker).getByRole("list").textContent).not.toContain(
      "Original source",
    );
    expect(owner.getSnapshot().needsReview).toBe(true);
    await act(async () =>
      accept({
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: sourceId,
              displayText: "Refreshed source presentation",
              viewSchemaId: timelineViewSchemaId,
            },
          ],
          hasMore: false,
          nextCursor: null,
        },
      }),
    );
    expect(
      within(picker)
        .getByRole("button", { name: /^Remove selected Linked Records / })
        .getAttribute("aria-label"),
    ).toBe("Remove selected Linked Records Refreshed source presentation");
    expect(owner.getSnapshot().draft?.labels[sourceId]).toBeUndefined();
    fireEvent.click(
      within(picker).getByRole("button", { name: "Apply references" }),
    );
    expect(owner.getSnapshot()).toMatchObject({
      needsReview: true,
      draft: {
        source: original.source,
        values: {
          "task.title": "Retained raw title",
          "task.linked_record_ids": sourceId,
        },
        labels: { [sourceId]: "Refreshed source presentation" },
      },
    });
    expect(reader.verify).not.toHaveBeenCalled();
  });

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

  it("keeps equal-label requester choice separate from raw fields until Apply", async () => {
    const { owner, reader } = fixture();
    owner.update("task.title", "Retained raw title");
    owner.update("task.requester_party_text", "Source wording");
    vi.mocked(reader.availableViews).mockResolvedValue({
      kind: "accepted",
      value: ["cartulary.view.parties.v1"],
    });
    vi.mocked(reader.page).mockImplementation(async (input) => ({
      kind: "accepted",
      value: {
        candidates: ["party-a", "party-b"].map((recordId) => ({
          recordId,
          displayText: "Response coordination team",
          viewSchemaId: input.viewSchemaId,
        })),
        hasMore: false,
        nextCursor: null,
      },
    }));
    const submit = vi.fn();
    render(
      <ContextualCreateForm
        owner={owner}
        attachment={attachment}
        onSubmit={submit}
      />,
    );
    const before = owner.getSnapshot().draft;
    fireEvent.click(
      screen.getByRole("button", { name: "Choose Requester Party" }),
    );
    await screen.findByRole("radio", {
      name: "Response coordination team (party-b)",
    });
    fireEvent.click(
      screen.getByRole("radio", {
        name: "Response coordination team (party-b)",
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Cancel references" }));
    expect(owner.getSnapshot().draft).toEqual(before);
    fireEvent.click(
      screen.getByRole("button", { name: "Choose Requester Party" }),
    );
    await screen.findByRole("radio", {
      name: "Response coordination team (party-b)",
    });
    fireEvent.click(
      screen.getByRole("radio", {
        name: "Response coordination team (party-b)",
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Apply references" }));
    expect(owner.getSnapshot().draft?.values).toMatchObject({
      "task.title": "Retained raw title",
      "task.requester_party_text": "Source wording",
      "task.requester_party_id": "party-b",
    });
    expect(submit).not.toHaveBeenCalled();
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
        screen.getByRole("group", { name: label }).hasAttribute("disabled"),
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
    const removeSource = screen.getByRole("button", {
      name: `Remove ${label} Original source`,
    });
    removeSource.focus();
    fireEvent.click(removeSource);
    expect(owner.getSnapshot().draft?.values["task.linked_record_ids"]).toBe(
      "",
    );
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: `Choose ${label}` }),
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
