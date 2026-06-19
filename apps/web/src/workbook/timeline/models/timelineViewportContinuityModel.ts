export type TimelineContinuityEntityType = "host" | "identity";

export type TimelineContinuityEntity = {
  readonly entityType: TimelineContinuityEntityType | string;
  readonly recordId: string | null | undefined;
  readonly rowVersion: number | null | undefined;
};

export type TimelineEntityRefreshExpectation = {
  readonly entityType: TimelineContinuityEntityType;
  readonly recordId: string;
};

export type TimelineEntityCatalogInput = {
  readonly hostEntities: readonly TimelineContinuityEntity[];
  readonly identityEntities: readonly TimelineContinuityEntity[];
};

export type TimelineEntityRefreshSettleState = "complete" | "terminal";

export type TimelineViewportContinuityBarrier = {
  readonly kind: "entity-refresh";
  readonly baselineCatalogKey: string;
  readonly refreshState: "pending" | TimelineEntityRefreshSettleState;
  readonly expectedEntity?: TimelineEntityRefreshExpectation | undefined;
} | null;

export type TimelineMentionEntityRefreshRow = {
  readonly collectionValues: {
    readonly hostRefs: readonly TimelineMentionEntityRefreshItem[];
    readonly identityRefs: readonly TimelineMentionEntityRefreshItem[];
  };
};

type TimelineMentionEntityRefreshItem = {
  readonly entityType: TimelineContinuityEntityType | string;
  readonly itemRef: string;
  readonly resolvedRecordId: string | null;
};

function entityCatalogToken(entity: TimelineContinuityEntity) {
  if (entity.recordId === null || entity.recordId === undefined) {
    return null;
  }
  const rowVersion =
    typeof entity.rowVersion === "number" ? String(entity.rowVersion) : "none";
  return `${entity.entityType}:${entity.recordId}:${rowVersion}`;
}

export function timelineEntityCatalogKey(input: TimelineEntityCatalogInput) {
  return [...input.hostEntities, ...input.identityEntities]
    .map(entityCatalogToken)
    .filter((token): token is string => token !== null)
    .sort()
    .join("|");
}

export function beginTimelineEntityRefreshBarrier(
  input: TimelineEntityCatalogInput,
): TimelineViewportContinuityBarrier {
  return {
    kind: "entity-refresh",
    baselineCatalogKey: timelineEntityCatalogKey(input),
    refreshState: "pending",
  };
}

export function settleTimelineViewportContinuityBarrier(
  barrier: TimelineViewportContinuityBarrier,
  refreshState: TimelineEntityRefreshSettleState,
): TimelineViewportContinuityBarrier {
  if (barrier === null) {
    return null;
  }
  return {
    ...barrier,
    refreshState,
  };
}

export function withTimelineEntityRefreshExpectation(
  barrier: TimelineViewportContinuityBarrier,
  expectedEntity: TimelineEntityRefreshExpectation | null,
): TimelineViewportContinuityBarrier {
  if (barrier === null || expectedEntity === null) {
    return barrier;
  }
  return {
    ...barrier,
    expectedEntity,
  };
}

export function timelineEntityRefreshExpectationForMention(
  row: TimelineMentionEntityRefreshRow,
  itemRef: string,
): TimelineEntityRefreshExpectation | null {
  const item = [
    ...row.collectionValues.hostRefs,
    ...row.collectionValues.identityRefs,
  ].find((candidate) => candidate.itemRef === itemRef);
  if (
    item?.resolvedRecordId === null ||
    item?.resolvedRecordId === undefined ||
    !(item.entityType === "host" || item.entityType === "identity")
  ) {
    return null;
  }
  return {
    entityType: item.entityType,
    recordId: item.resolvedRecordId,
  };
}

function entityCatalogHasExpectedEntity(
  input: TimelineEntityCatalogInput,
  expectedEntity: TimelineEntityRefreshExpectation | undefined,
) {
  if (expectedEntity === undefined) {
    return false;
  }
  const entities =
    expectedEntity.entityType === "host"
      ? input.hostEntities
      : input.identityEntities;
  return entities.some(
    (entity) =>
      entity.entityType === expectedEntity.entityType &&
      entity.recordId === expectedEntity.recordId,
  );
}

export function timelineViewportContinuityBarrierSatisfied(
  barrier: TimelineViewportContinuityBarrier,
  input: TimelineEntityCatalogInput,
) {
  if (barrier === null) {
    return true;
  }
  if (barrier.refreshState === "pending") {
    return false;
  }
  if (barrier.refreshState === "terminal") {
    return true;
  }
  return (
    entityCatalogHasExpectedEntity(input, barrier.expectedEntity) ||
    timelineEntityCatalogKey(input) !== barrier.baselineCatalogKey
  );
}
