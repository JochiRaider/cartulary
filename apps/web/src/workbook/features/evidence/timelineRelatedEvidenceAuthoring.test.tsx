import {
  evidenceViewSchemaId,
  partiesViewSchemaId,
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
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";
import { TimelineRelatedEvidenceForm } from "./TimelineRelatedEvidenceForm";
import {
  prepareTimelineRelatedEvidence,
  timelineRelatedEvidenceRequest,
} from "./timelineRelatedEvidenceModel";
import { WorkbookTimelineRelatedEvidenceOwner } from "./WorkbookTimelineRelatedEvidenceOwner";

afterEach(cleanup);
const authority: WorkbookMutationAuthority = {
  actorId: "10000000-0000-4000-8000-000000000001",
  incidentId: "incident",
  sessionIdentity: "session",
  role: "editor",
  closed: false,
};
const sourceId = "20000000-0000-4000-8000-000000000002",
  partyId = "30000000-0000-4000-8000-000000000003",
  token = Symbol("inspector");
function fixture() {
  const owner = new WorkbookTimelineRelatedEvidenceOwner(authority.incidentId);
  owner.setAuthority(authority);
  const feature = requireViewContract(
    timelineViewSchemaId,
  ).inspectorConfig.featureGroups.find(
    (item) => item.featureGroupKey === "create_related.evidence",
  );
  if (!feature) throw new Error("Missing feature");
  const subject = {
    subject: {
      kind: "live" as const,
      recordId: sourceId,
      rowVersion: 1,
      viewSchemaId: timelineViewSchemaId,
      label: "Raw original text",
      surfaceLabel: "Timeline",
    },
    cells: {},
  };
  const reader: WorkbookAuthoringReadPort = {
    verify: vi.fn(async () => {}),
    availableViews: async () => ({
      kind: "accepted" as const,
      value: [evidenceViewSchemaId, timelineViewSchemaId, partiesViewSchemaId],
    }),
    page: vi.fn(async ({ viewSchemaId }) => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          {
            recordId: sourceId,
            displayText: "Source",
            viewSchemaId,
            row: { record_id: sourceId, row_version: 1, cells: {} },
          },
        ],
        hasMore: false,
        nextCursor: null,
      },
    })),
  };
  owner.configure(reader, async () => authority);
  owner.begin(
    subject,
    feature,
    { kind: "view_schema", id: timelineViewSchemaId },
    token,
  );
  const draft = () => {
    const value = owner.getSnapshot().draft;
    if (!value) throw new Error("Missing draft");
    return value;
  };
  return { owner, feature, subject, reader, draft };
}
describe("Timeline related Evidence authoring", () => {
  it("retains original identity and exact input until explicit discard across presentation and authority changes", () => {
    const { owner, subject, feature, draft } = fixture();
    owner.update("evidence.title", "  retained title  ");
    owner.update("evidence.requested_at", "invalid timestamp");
    const original = draft();
    owner.detach(token);
    owner.observe(sourceId, 9);
    expect(
      owner.begin(
        { ...subject, subject: { ...subject.subject, recordId: partyId } },
        feature,
        { kind: "view_schema", id: evidenceViewSchemaId },
        Symbol(),
      ),
    ).toBe(false);
    owner.resume(token);
    expect(draft().values).toEqual(original.values);
    expect(draft().source).toEqual(original.source);
    expect(owner.getSnapshot().needsReview).toBe(true);
    owner.setAuthority({ ...authority, role: "viewer" });
    expect(owner.canSubmit()).toBe(false);
    expect(draft().values).toEqual(original.values);
    owner.suspend();
    expect(owner.getSnapshot().draft).toBeNull();
    owner.setAuthority(authority);
    expect(draft().values).toEqual(original.values);
    owner.discard();
    expect(owner.getSnapshot().draft).toBeNull();
    expect(
      owner.begin(
        subject,
        feature,
        { kind: "view_schema", id: timelineViewSchemaId },
        token,
      ),
    ).toBe(true);
    owner.setAuthority({ ...authority, actorId: partyId });
    expect(owner.getSnapshot().draft).toBeNull();
  });
  it("rejects blank default-only and Party-only input while accepting every non-title minimum with exact omission", () => {
    const { owner, draft } = fixture();
    expect(timelineRelatedEvidenceRequest(draft(), "txn")).toBeNull();
    owner.update("evidence.collector_party_id", partyId);
    expect(timelineRelatedEvidenceRequest(draft(), "txn")).toBeNull();
    const signals = {
      "evidence.storage_ref": "external://case/file",
      "evidence.collector_party_text": "Analyst",
      "evidence.source_party_text": "Source witness",
      "evidence.requested_at": "2026-09-12T14:30:00Z",
      "evidence.received_at": "2026-09-12T14:30:00Z",
      "evidence.lifecycle_state": "requested",
    };
    for (const [key, value] of Object.entries(signals)) {
      owner.update(key, value);
      expect(timelineRelatedEvidenceRequest(draft(), "txn")).toEqual({
        client_txn_id: "txn",
        "evidence.collector_party_id": partyId,
        [key]: value,
      });
      owner.update(key, "");
    }
    owner.update("evidence.title", " \u0085 ");
    expect(timelineRelatedEvidenceRequest(draft(), "txn")).toBeNull();
    owner.update("evidence.title", " Cafe\u0301  title ");
    expect(timelineRelatedEvidenceRequest(draft(), "txn")).toEqual({
      client_txn_id: "txn",
      "evidence.collector_party_id": partyId,
      "evidence.title": "Café  title",
    });
    owner.update("initial_object_blob_id", partyId);
    owner.update("evidence.blob_hash", "derived");
    expect(draft().values).not.toHaveProperty("initial_object_blob_id");
    expect(draft().values).not.toHaveProperty("evidence.blob_hash");
  });
  it("validates initial lifecycle timestamps strings and exact Party IDs without erasing invalid input", () => {
    const { owner, draft } = fixture();
    for (const state of [
      "requested",
      "pending_receipt",
      "received",
      "quarantined",
    ]) {
      owner.update("evidence.lifecycle_state", state);
      expect(timelineRelatedEvidenceRequest(draft(), "txn")).not.toBeNull();
    }
    for (const state of ["available", "released", "unknown"]) {
      owner.update("evidence.lifecycle_state", state);
      expect(prepareTimelineRelatedEvidence(draft()).errors).toHaveProperty(
        "evidence.lifecycle_state",
      );
      expect(draft().values["evidence.lifecycle_state"]).toBe(state);
    }
    owner.update("evidence.lifecycle_state", "");
    for (const stamp of [
      "2026-02-30T12:00:00Z",
      "2026-09-12",
      "2026-09-12T12:00:00",
      "2026-09-12T24:00:00Z",
    ]) {
      owner.update("evidence.requested_at", stamp);
      expect(prepareTimelineRelatedEvidence(draft()).errors).toHaveProperty(
        "evidence.requested_at",
      );
      expect(draft().values["evidence.requested_at"]).toBe(stamp);
    }
    owner.update("evidence.requested_at", "2026-09-12T14:30:00.123456-04:00");
    expect(timelineRelatedEvidenceRequest(draft(), "txn")).toEqual({
      client_txn_id: "txn",
      "evidence.requested_at": "2026-09-12T18:30:00.123456Z",
    });
    for (const [key, value] of [
      ["evidence.title", "x".repeat(513)],
      ["evidence.source_party_text", "bad\ntext"],
      ["evidence.storage_ref", `object://${partyId}`],
      ["evidence.collector_party_id", ` ${partyId}`],
    ]) {
      if (!key || !value) throw new Error("Missing case");
      owner.update(key, value);
      expect(prepareTimelineRelatedEvidence(draft()).errors).toHaveProperty(
        key,
      );
      owner.update(key, "");
    }
  });
  it("discovers Parties with staged paging cancellation off-page selection and independent text", async () => {
    const { owner, reader, draft } = fixture();
    owner.update(
      "evidence.collector_party_text",
      "Preserved collector spelling",
    );
    owner.update("evidence.collector_party_id", partyId, {
      [partyId]: "Earlier selection",
    });
    reader.page = vi.fn<WorkbookAuthoringReadPort["page"]>(
      async ({ cursor, viewSchemaId }) => ({
        kind: "accepted",
        value: {
          candidates: [
            {
              recordId: cursor ? sourceId : authority.actorId,
              displayText: cursor ? "Second page" : "First page",
              viewSchemaId,
            },
          ],
          hasMore: cursor === null,
          nextCursor: cursor === null ? "next" : null,
        },
      }),
    );
    render(
      <TimelineRelatedEvidenceForm
        owner={owner}
        attachment={token}
        onSubmit={() => {}}
        onReview={() => {}}
      />,
    );
    expect(
      (screen.getByLabelText("Lifecycle") as HTMLSelectElement).value,
    ).toBe("");
    expect(
      screen.getByRole("option", { name: "Use server default (Requested)" }),
    ).not.toBeNull();
    expect(screen.queryByRole("option", { name: "available" })).toBeNull();
    fireEvent.click(
      screen.getByRole("button", { name: "Choose Collector Party" }),
    );
    await screen.findByRole("option", { name: "First page" });
    expect(
      screen.getByRole("button", {
        name: "Remove selected Collector Party Earlier selection",
      }),
    ).not.toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Next candidates" }));
    await screen.findByRole("option", { name: "Second page" });
    fireEvent.change(screen.getByLabelText("Collector Party"), {
      target: { value: sourceId },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Cancel Party selection" }),
    );
    expect(draft().values["evidence.collector_party_id"]).toBe(partyId);
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Choose Collector Party" }),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Choose Collector Party" }),
    );
    await screen.findByRole("option", { name: "First page" });
    fireEvent.change(screen.getByLabelText("Collector Party"), {
      target: { value: authority.actorId },
    });
    fireEvent.click(screen.getByRole("button", { name: "Apply Party" }));
    expect(draft().values["evidence.collector_party_id"]).toBe(
      authority.actorId,
    );
    expect(draft().values["evidence.collector_party_text"]).toBe(
      "Preserved collector spelling",
    );
    expect(reader.page).toHaveBeenCalledWith(
      expect.objectContaining({
        viewSchemaId: partiesViewSchemaId,
        cursor: "next",
      }),
    );
  });
  it("exposes loading empty and failed discovery and rejects unavailable references on renewed review", async () => {
    const { owner, reader, draft } = fixture();
    reader.page = vi.fn(async () => {
      throw new Error("Party discovery failed");
    });
    render(
      <TimelineRelatedEvidenceForm
        owner={owner}
        attachment={token}
        onSubmit={() => {}}
        onReview={() => {}}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Choose Source Party" }),
    );
    expect(screen.getByText("Loading reference surfaces…")).not.toBeNull();
    await screen.findByRole("button", { name: "Retry candidates" });
    reader.page = vi.fn<WorkbookAuthoringReadPort["page"]>(async () => ({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    }));
    act(() => owner.observe());
    await screen.findByText("No candidates match this query.");
    fireEvent.click(
      screen.getByRole("button", { name: "Cancel Party selection" }),
    );
    reader.page = vi.fn<WorkbookAuthoringReadPort["page"]>(
      async ({ viewSchemaId }) => ({
        kind: "accepted",
        value: {
          candidates:
            viewSchemaId === timelineViewSchemaId
              ? [
                  {
                    recordId: sourceId,
                    viewSchemaId,
                    displayText: "source",
                    row: { record_id: sourceId, row_version: 2, cells: {} },
                  },
                ]
              : [],
          hasMore: false,
          nextCursor: null,
        },
      }),
    );
    act(() => {
      owner.update("evidence.source_party_text", "Witness");
      owner.update("evidence.source_party_id", partyId);
    });
    await act(async () => {
      expect(await owner.review()).toBe(false);
    });
    await waitFor(() =>
      expect(owner.getSnapshot().errors).toHaveProperty(
        "evidence.source_party_id",
      ),
    );
    expect(draft().values["evidence.source_party_id"]).toBe(partyId);
    expect(draft().values["evidence.source_party_text"]).toBe("Witness");
  });
});
