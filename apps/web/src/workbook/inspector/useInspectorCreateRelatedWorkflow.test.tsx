import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import type { TimelineRelatedRecordPort } from "../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { useInspectorCreateRelatedWorkflow } from "./useInspectorCreateRelatedWorkflow";

const timeline = requireViewContract("cartulary.view.assessments.v1");
const commLog = requireViewContract("cartulary.view.assessments.v1");
const createCommLog = timeline.inspectorConfig.featureGroups.find(
  (feature) => feature.featureGroupKey === "create_related.assessment",
);
const initialSubject = {
  cells: { "timeline.activity_synopsis_text": { value: "Investigate" } },
  subject: {
    kind: "live" as const,
    label: "Timeline row",
    recordId: "10000000-0000-4000-8000-000000000001",
    rowVersion: 4,
    surfaceLabel: timeline.title,
    viewSchemaId: timeline.viewSchemaId,
  },
};

describe("useInspectorCreateRelatedWorkflow", () => {
  it("publishes related success as neutral non-announced feedback", async () => {
    expect(createCommLog).toBeDefined();
    if (createCommLog === undefined) return;
    const onFeedback = vi.fn();
    const finishReport = vi.fn();
    const beginMutation = vi.fn(() => finishReport);
    const { result } = renderHook(() =>
      useInspectorCreateRelatedWorkflow({
        beginMutation,
        currentUserId: null,
        mutationCommands: relatedPort(async () => ({
          kind: "accepted",
          value: {
            changeSetId: "20000000-0000-4000-8000-000000000001",
            recordId: "20000000-0000-4000-8000-000000000002",
            viewSchemaId: commLog.viewSchemaId,
          },
        })),
        onCreated: vi.fn(),
        onFeedback,
        selectedSubject: initialSubject,
      }),
    );

    act(() => result.current.commands.begin(createCommLog));
    await act(async () => result.current.commands.submit());
    expect(beginMutation).toHaveBeenCalledOnce();
    expect(finishReport).toHaveBeenCalledOnce();

    expect(onFeedback).toHaveBeenLastCalledWith({
      announcement: "none",
      kind: "message",
      sourceRecordId: initialSubject.subject.recordId,
      destination: {
        kind: "feature",
        panel: createCommLog.panelId,
        featureGroupKey: createCommLog.featureGroupKey,
      },
      message: `Created ${commLog.title} record 20000000-0000-4000-8000-000000000002.`,
    });
  });

  it("ignores a late accepted result after cancel and reopen while retaining the owner effect", async () => {
    expect(createCommLog).toBeDefined();
    if (createCommLog === undefined) return;
    const pending =
      deferred<
        Awaited<ReturnType<TimelineRelatedRecordPort["createRelatedRecord"]>>
      >();
    const mutationCommands = relatedPort(() => pending.promise);
    const onCreated = vi.fn(async () => undefined);
    const onFeedback = vi.fn();
    const { result } = renderHook(() =>
      useInspectorCreateRelatedWorkflow({
        beginMutation: () => vi.fn(),
        currentUserId: null,
        mutationCommands,
        onCreated,
        onFeedback,
        selectedSubject: initialSubject,
      }),
    );

    act(() => result.current.commands.begin(createCommLog));
    let completion: Promise<void> | undefined;
    await act(async () => {
      completion = result.current.commands.submit();
      await Promise.resolve();
    });
    act(() => {
      result.current.commands.cancel();
      result.current.commands.begin(createCommLog);
    });
    const reopenedWorkflowId = result.current.snapshot.workflow?.workflowId;
    const feedbackCallCount = onFeedback.mock.calls.length;

    await act(async () => {
      pending.resolve({
        kind: "accepted",
        value: {
          changeSetId: "20000000-0000-4000-8000-000000000001",
          recordId: "20000000-0000-4000-8000-000000000002",
          viewSchemaId: commLog.viewSchemaId,
        },
      });
      await completion;
    });

    expect(onCreated).toHaveBeenCalledOnce();
    expect(onFeedback).toHaveBeenCalledTimes(feedbackCallCount);
    expect(result.current.snapshot.workflow).toMatchObject({
      phase: "editing",
      subject: {
        recordId: initialSubject.subject.recordId,
        rowVersion: initialSubject.subject.rowVersion,
      },
      workflowId: reopenedWorkflowId,
    });
  });

  it("ignores a late rejection after a row-version retarget", async () => {
    expect(createCommLog).toBeDefined();
    if (createCommLog === undefined) return;
    const pending =
      deferred<
        Awaited<ReturnType<TimelineRelatedRecordPort["createRelatedRecord"]>>
      >();
    const { result, rerender } = renderHook(
      ({ subject }) =>
        useInspectorCreateRelatedWorkflow({
          beginMutation: () => vi.fn(),
          currentUserId: null,
          mutationCommands: relatedPort(() => pending.promise),
          onCreated: vi.fn(),
          onFeedback: vi.fn(),
          selectedSubject: subject,
        }),
      { initialProps: { subject: initialSubject } },
    );

    act(() => result.current.commands.begin(createCommLog));
    let completion: Promise<void> | undefined;
    await act(async () => {
      completion = result.current.commands.submit();
      await Promise.resolve();
    });
    rerender({
      subject: {
        ...initialSubject,
        subject: {
          ...initialSubject.subject,
          rowVersion: initialSubject.subject.rowVersion + 1,
        },
      },
    });
    await waitFor(() => expect(result.current.snapshot.workflow).toBeNull());
    await act(async () => {
      pending.resolve({
        kind: "rejected",
        failure: retryableFailure("Late rejection"),
      });
      await completion;
    });

    expect(result.current.snapshot.workflow).toBeNull();
  });
});

function relatedPort(
  createRelatedRecord: TimelineRelatedRecordPort["createRelatedRecord"],
): TimelineRelatedRecordPort {
  return {
    createRelatedRecord,
  };
}

function retryableFailure(message: string): WorkbookOperationFailure {
  return { kind: "retryable", message };
}
