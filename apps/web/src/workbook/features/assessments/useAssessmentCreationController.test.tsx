import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type {
  AssessmentCreateOutcome,
  AssessmentMutationCommandPort,
} from "../../mutations/workbookMutationCommandPorts";
import { useAssessmentCreationController } from "./useAssessmentCreationController";

const subjectRecordIds = {
  host: ["00000000-0000-4000-8000-000000008100"],
  identity: ["00000000-0000-4000-8000-000000008101"],
} as const;

function port({
  canCreate = true,
  create,
}: {
  readonly canCreate?: boolean;
  readonly create: AssessmentMutationCommandPort["create"];
}): AssessmentMutationCommandPort {
  return { canCreate: () => canCreate, create };
}

describe("useAssessmentCreationController", () => {
  it("publishes accepted standalone creation as polite neutral feedback", async () => {
    const { result } = renderHook(() =>
      useAssessmentCreationController({
        beginMutation: () => vi.fn(),
        lifecycleResetKey: "authorized",
        mutationCommands: port({
          create: async () => ({
            kind: "accepted",
            value: {
              changeSetId: "00000000-0000-4000-8000-000000008300",
              row: {
                cells: {},
                record_id: "00000000-0000-4000-8000-000000008301",
                row_version: 1,
              },
              viewSchemaId: "cartulary.view.assessments.v1",
            },
          }),
        }),
        onRefreshAssessmentRows: async () => undefined,
        subjectRecordIds,
      }),
    );
    act(() => {
      result.current.commands.openStandalone(subjectRecordIds.host[0]);
      result.current.commands.updateDraft((current) => ({
        ...current,
        rationale: "Confirmed assessment",
      }));
    });

    await act(async () => result.current.commands.submit(true));

    expect(result.current.snapshot.feedback).toEqual({
      announcement: "polite",
      kind: "message",
      message: "Assessment created.",
    });
  });

  it("retains the append-only draft and support selection after rejection", async () => {
    const create = vi.fn(async () => ({
      kind: "rejected" as const,
      failure: { kind: "retryable" as const, message: "Try again." },
    }));
    const finishMutation = vi.fn();
    const { result } = renderHook(() =>
      useAssessmentCreationController({
        beginMutation: () => finishMutation,
        lifecycleResetKey: "authorized",
        mutationCommands: port({ create }),
        onRefreshAssessmentRows: async () => undefined,
        subjectRecordIds,
      }),
    );
    act(() => {
      result.current.commands.openStandalone(subjectRecordIds.host[0]);
      result.current.commands.updateDraft((current) => ({
        ...current,
        assessmentState: "suspected",
        rationale: "Preserve this draft",
        supportRecordIds: ["00000000-0000-4000-8000-000000008200"],
      }));
    });

    await act(async () => result.current.commands.submit(true));

    expect(create).toHaveBeenCalledTimes(1);
    expect(finishMutation).toHaveBeenCalledTimes(1);
    expect(result.current.snapshot.feedback).toEqual({
      error: { primaryMessage: "Try again.", technicalFields: [] },
      kind: "error",
    });
    expect(result.current.snapshot.draft.rationale).toBe("Preserve this draft");
    expect(result.current.snapshot.draft.supportRecordIds).toEqual([
      "00000000-0000-4000-8000-000000008200",
    ]);
  });

  it("rejects invalid drafts locally and ignores late accepted lifecycle completions", async () => {
    let resolveCreate: ((outcome: AssessmentCreateOutcome) => void) | undefined;
    const create = vi.fn(
      () =>
        new Promise<AssessmentCreateOutcome>((resolve) => {
          resolveCreate = resolve;
        }),
    );
    const canCreate = vi.fn(() => false);
    const onRefreshAssessmentRows = vi.fn(async () => undefined);
    const { result, rerender } = renderHook(
      ({ lifecycleResetKey }: { lifecycleResetKey: string }) =>
        useAssessmentCreationController({
          beginMutation: () => vi.fn(),
          lifecycleResetKey,
          mutationCommands: { canCreate, create },
          onRefreshAssessmentRows,
          subjectRecordIds,
        }),
      { initialProps: { lifecycleResetKey: "authorized" } },
    );

    await act(async () => result.current.commands.submit(true));
    expect(create).not.toHaveBeenCalled();
    expect(result.current.snapshot.feedback).toEqual({
      error: {
        primaryMessage: "Complete the required assessment fields.",
        technicalFields: [],
      },
      kind: "error",
    });

    canCreate.mockReturnValue(true);
    act(() => {
      result.current.commands.updateDraft((current) => ({
        ...current,
        rationale: "Pending assessment",
      }));
    });
    let completion: Promise<void> | undefined;
    await act(async () => {
      completion = result.current.commands.submit(true);
      await Promise.resolve();
    });
    rerender({ lifecycleResetKey: "authorization-lost" });
    await act(async () => {
      resolveCreate?.({
        kind: "accepted",
        value: {
          changeSetId: "00000000-0000-4000-8000-000000008300",
          row: {
            cells: {},
            record_id: "00000000-0000-4000-8000-000000008301",
            row_version: 1,
          },
          viewSchemaId: "cartulary.view.assessments.v1",
        },
      });
      await completion;
    });

    expect(onRefreshAssessmentRows).not.toHaveBeenCalled();
    expect(result.current.snapshot.feedback).toBeNull();
    expect(result.current.snapshot.isSubmitting).toBe(false);
  });
});
