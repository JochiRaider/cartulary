export type TimelineContinuitySemanticTarget =
  | { readonly kind: "row-inspect"; readonly recordId: string }
  | { readonly kind: "input"; readonly focusKey: string }
  | { readonly kind: "scroll-only" };

export type TimelineContinuityRequirementName =
  | "entity-refresh"
  | "row-projection";

export type TimelineContinuityRequirementState =
  | "pending"
  | "settled"
  | "terminal";

export type TimelineSourceRecordRequirement = {
  readonly recordId: string;
  readonly minimumRowVersion: number;
};

export type TimelineSourceRecordEvidence = {
  readonly recordId: string;
  readonly rowVersion: number;
};

export type TimelineContinuityLifecycleState =
  | "pending"
  | "completed"
  | "cancelled"
  | "failed";

export type TimelineContinuityLifecycle = {
  readonly semanticFocusTarget: TimelineContinuitySemanticTarget;
  readonly sourceRecordRequirement: TimelineSourceRecordRequirement | null;
  readonly followUpRequirements: Readonly<
    Partial<
      Record<
        TimelineContinuityRequirementName,
        TimelineContinuityRequirementState
      >
    >
  >;
  readonly renderGeneration: number;
  readonly userInterruptionGeneration: number;
  readonly state: TimelineContinuityLifecycleState;
};

export function beginTimelineContinuityLifecycle({
  semanticFocusTarget,
  userInterruptionGeneration,
  requirements = [],
}: {
  readonly semanticFocusTarget: TimelineContinuitySemanticTarget;
  readonly userInterruptionGeneration: number;
  readonly requirements?: readonly TimelineContinuityRequirementName[];
}): TimelineContinuityLifecycle {
  return {
    semanticFocusTarget,
    sourceRecordRequirement: null,
    followUpRequirements: Object.fromEntries(
      requirements.map((requirement) => [requirement, "pending"]),
    ),
    renderGeneration: 0,
    userInterruptionGeneration,
    state: "pending",
  };
}

export function requireTimelineSourceRecord(
  lifecycle: TimelineContinuityLifecycle,
  sourceRecord: TimelineSourceRecordRequirement,
): TimelineContinuityLifecycle {
  if (
    sourceRecord.recordId.trim() === "" ||
    !Number.isSafeInteger(sourceRecord.minimumRowVersion) ||
    sourceRecord.minimumRowVersion < 1
  ) {
    throw new Error("Timeline continuity source record evidence is invalid.");
  }
  if (
    lifecycle.semanticFocusTarget.kind === "row-inspect" &&
    lifecycle.semanticFocusTarget.recordId !== sourceRecord.recordId
  ) {
    throw new Error(
      `Timeline continuity source record ${sourceRecord.recordId} does not match target ${lifecycle.semanticFocusTarget.recordId}.`,
    );
  }
  return {
    ...lifecycle,
    sourceRecordRequirement: sourceRecord,
    followUpRequirements: {
      ...lifecycle.followUpRequirements,
      "row-projection": "pending",
    },
  };
}

export function timelineSourceRecordRequirementSatisfied(
  requirement: TimelineSourceRecordRequirement,
  evidence: TimelineSourceRecordEvidence,
) {
  return (
    evidence.recordId === requirement.recordId &&
    evidence.rowVersion >= requirement.minimumRowVersion
  );
}

export function advanceTimelineContinuityRender(
  lifecycle: TimelineContinuityLifecycle,
  options: {
    readonly sourceRecord?: TimelineSourceRecordEvidence | undefined;
  } = {},
): TimelineContinuityLifecycle {
  const sourceRecordRequirement = lifecycle.sourceRecordRequirement;
  const sourceRecordSatisfied =
    sourceRecordRequirement !== null &&
    options.sourceRecord !== undefined &&
    timelineSourceRecordRequirementSatisfied(
      sourceRecordRequirement,
      options.sourceRecord,
    );
  return {
    ...lifecycle,
    followUpRequirements: sourceRecordSatisfied
      ? {
          ...lifecycle.followUpRequirements,
          "row-projection": "settled",
        }
      : lifecycle.followUpRequirements,
    renderGeneration: lifecycle.renderGeneration + 1,
  };
}

export function settleTimelineContinuityRequirement(
  lifecycle: TimelineContinuityLifecycle,
  requirement: TimelineContinuityRequirementName,
  state: Exclude<TimelineContinuityRequirementState, "pending">,
): TimelineContinuityLifecycle {
  if (lifecycle.followUpRequirements[requirement] === undefined) {
    return lifecycle;
  }
  return {
    ...lifecycle,
    followUpRequirements: {
      ...lifecycle.followUpRequirements,
      [requirement]: state,
    },
    renderGeneration: lifecycle.renderGeneration + 1,
  };
}

export function timelineContinuityRequirementsSettled(
  lifecycle: TimelineContinuityLifecycle,
) {
  return Object.values(lifecycle.followUpRequirements).every(
    (requirement) => requirement !== "pending",
  );
}

export function transitionTimelineContinuity(
  lifecycle: TimelineContinuityLifecycle,
  state: Exclude<TimelineContinuityLifecycleState, "pending">,
): TimelineContinuityLifecycle {
  if (lifecycle.state !== "pending") {
    return lifecycle;
  }
  return {
    ...lifecycle,
    followUpRequirements:
      state === "failed"
        ? Object.fromEntries(
            Object.keys(lifecycle.followUpRequirements).map((requirement) => [
              requirement,
              "terminal",
            ]),
          )
        : lifecycle.followUpRequirements,
    renderGeneration: lifecycle.renderGeneration + 1,
    state,
  };
}
