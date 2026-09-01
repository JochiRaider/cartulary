import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import type { TimelineRelatedRecordPort } from "../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { useInspectorCreateRelatedWorkflow } from "./useInspectorCreateRelatedWorkflow";

const timeline = requireViewContract("cartulary.view.timeline.v2");
const taskRequests = requireViewContract("cartulary.view.task_requests.v1");
const createTaskRequest = timeline.inspectorConfig.featureGroups.find(
  (feature) => feature.featureGroupKey === "create_related.task_request",
);
const initialSubject = {
  cells: { "timeline.activity_synopsis_text": { value: "Investigate" } },
  recordId: "10000000-0000-4000-8000-000000000001",
  rowVersion: 4,
  viewSchemaId: timeline.viewSchemaId,
};

describe("useInspectorCreateRelatedWorkflow", () => {
  it("ignores a late accepted result after cancel and reopen while retaining the owner effect", async () => {
    expect(createTaskRequest).toBeDefined();
    if (createTaskRequest === undefined) return;
    const pending =
      deferred<
        Awaited<ReturnType<TimelineRelatedRecordPort["createRelatedRecord"]>>
      >();
    const mutationCommands = relatedPort(() => pending.promise);
    const onCreated = vi.fn(async () => undefined);
    const onMessage = vi.fn();
    const { result } = renderHook(() =>
      useInspectorCreateRelatedWorkflow({
        currentUserId: null,
        mutationCommands,
        onCreated,
        onMessage,
        selectedSubject: initialSubject,
      }),
    );

    act(() => result.current.commands.begin(createTaskRequest));
    let completion: Promise<void> | undefined;
    await act(async () => {
      completion = result.current.commands.submit();
      await Promise.resolve();
    });
    act(() => {
      result.current.commands.cancel();
      result.current.commands.begin(createTaskRequest);
    });
    const reopenedWorkflowId = result.current.snapshot.workflow?.workflowId;
    const messageCallCount = onMessage.mock.calls.length;

    await act(async () => {
      pending.resolve({
        kind: "accepted",
        value: {
          changeSetId: "20000000-0000-4000-8000-000000000001",
          recordId: "20000000-0000-4000-8000-000000000002",
          viewSchemaId: taskRequests.viewSchemaId,
        },
      });
      await completion;
    });

    expect(onCreated).toHaveBeenCalledOnce();
    expect(onMessage).toHaveBeenCalledTimes(messageCallCount);
    expect(result.current.snapshot.workflow).toMatchObject({
      phase: "editing",
      subjectKey: {
        recordId: initialSubject.recordId,
        rowVersion: initialSubject.rowVersion,
      },
      workflowId: reopenedWorkflowId,
    });
  });

  it("ignores a late rejection after a row-version retarget", async () => {
    expect(createTaskRequest).toBeDefined();
    if (createTaskRequest === undefined) return;
    const pending =
      deferred<
        Awaited<ReturnType<TimelineRelatedRecordPort["createRelatedRecord"]>>
      >();
    const { result, rerender } = renderHook(
      ({ subject }) =>
        useInspectorCreateRelatedWorkflow({
          currentUserId: null,
          mutationCommands: relatedPort(() => pending.promise),
          onCreated: vi.fn(),
          onMessage: vi.fn(),
          selectedSubject: subject,
        }),
      { initialProps: { subject: initialSubject } },
    );

    act(() => result.current.commands.begin(createTaskRequest));
    let completion: Promise<void> | undefined;
    await act(async () => {
      completion = result.current.commands.submit();
      await Promise.resolve();
    });
    rerender({
      subject: { ...initialSubject, rowVersion: initialSubject.rowVersion + 1 },
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
    linkCreatedEvidence: vi.fn(async () => ({
      kind: "rejected" as const,
      failure: retryableFailure("Unexpected Evidence link"),
    })),
  };
}

function retryableFailure(message: string): WorkbookOperationFailure {
  return { kind: "retryable", message };
}
