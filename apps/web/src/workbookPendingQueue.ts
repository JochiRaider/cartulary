import { publicErrorCode, publicErrorStatusText } from "./publicError";
import { parseSameFieldConflictFields } from "./timelineConflictModel";

export const pendingReplayCapacity = 64;

export type PendingReplayKind = "create" | "patch";
export type PendingReplayMethod = "POST" | "PATCH";
export type PendingReplaySource = "autosave" | "paste";
export type PendingReplayStatus = "queued" | "in_flight";
export type PendingReplayPayloadIntent = Record<string, unknown>;
export type PendingReplayPrimarySaveState = "Conflict" | "Saved" | "Syncing";
export type WorkbookSaveStatePrimaryLabel = PendingReplayPrimarySaveState;
export type WorkbookSaveStateSecondaryKind =
  | "auth_paused"
  | "overflow"
  | "queued"
  | "replay_halted"
  | "same_field_conflict";

export type PendingReplayScope = {
  incidentId: string;
  clientInstanceId: string;
};

export type PendingReplayPresentationHint = {
  label?: string;
  recordType?: string;
  sortRank?: number;
  visibleRowIndex?: number;
};

export type PendingReplayOperationClass =
  | "conflict_resolution"
  | "destructive"
  | "hot_path"
  | "non_hot_path";

export type PendingReplayVisibleEdit = {
  rowKey: string;
  fieldKey: string;
  value: unknown;
};

export type PendingReplayUnitBase = {
  id: string;
  kind: PendingReplayKind;
  source: PendingReplaySource;
  incidentId: string;
  clientInstanceId: string;
  viewSchemaId: string;
  rowKey: string;
  recordId: string | null;
  method: PendingReplayMethod;
  path: string;
  payloadIntent: PendingReplayPayloadIntent;
  clientTxnId: string;
  coalesceKey: string;
  enqueueOrder: number;
  presentationHint?: PendingReplayPresentationHint;
};

export type PendingReplayUnitInput = PendingReplayUnitBase & {
  mutationSignature?: string;
  operationClass?: PendingReplayOperationClass;
  status?: PendingReplayStatus;
  visibleEdit?: PendingReplayVisibleEdit;
};

export type PendingReplayCanonicalChange = Record<string, unknown> & {
  field_key: string;
};

export type PendingReplayCreateIdentity = {
  kind: "create";
  method: "POST";
  route_scope: {
    incident_id: string;
    path: string;
    view_schema_id: string;
  };
  client_txn_id: string;
};

export type PendingReplayPatchIdentity = {
  kind: "patch";
  method: "PATCH";
  route_scope: {
    path: string;
    record_id: string;
  };
  record_id: string;
  client_txn_id: string;
  view_schema_id: string;
  base_row_version: number | null;
  changes: PendingReplayCanonicalChange[];
};

export type PendingReplayMutationIdentity =
  | PendingReplayCreateIdentity
  | PendingReplayPatchIdentity;

export type PendingReplayUnitState = PendingReplayUnitBase & {
  mutationSignature: string;
  operationClass: PendingReplayOperationClass;
  status: PendingReplayStatus;
  identity: PendingReplayMutationIdentity;
};

export type PendingReplayPublicError = {
  code: string;
  message?: string;
  retryable?: boolean;
  conflict?: PendingReplayPublicSameFieldConflict;
  details?: Record<string, unknown>;
};

export type PendingReplayPublicSameFieldConflict = Record<string, unknown> & {
  base_row_version: number;
  conflict_resolution_class: string;
  conflict_token: string;
  current_row_version: number;
  field_key: string;
  record_id: string;
};

export type PendingReplayPublicResult =
  | {
      ok: true;
      row: {
        record_id: string;
        row_version: number;
      };
      change_set_id?: string;
    }
  | {
      ok: false;
      status: number;
      error: PendingReplayPublicError;
    };

export type PendingReplayPublicAnchor =
  | {
      kind: "cell";
      record_id: string | null;
      field_key: string;
    }
  | {
      kind: "mutation";
      client_txn_id: string;
      route_scope: PendingReplayMutationIdentity;
    }
  | {
      kind: "record";
      record_id: string | null;
    }
  | {
      kind: "surface";
    };

export type PendingReplaySameFieldConflict = {
  key: string;
  conflict_token: string;
  record_id: string;
  field_key: string;
  conflict_resolution_class: string;
  base_row_version: number;
  current_row_version: number;
};

export type WorkbookSaveStateConflictAnchor = {
  record_id: string;
  field_key: string;
  base_row_version: number;
  current_row_version?: number;
};

export function sameFieldConflictQueueKey(
  anchor: Pick<WorkbookSaveStateConflictAnchor, "field_key" | "record_id">,
): string {
  return `${anchor.record_id}:${anchor.field_key}`;
}

export function workbookSaveStateConflictAnchorIdentity(
  anchor: WorkbookSaveStateConflictAnchor,
): string {
  return `${anchor.record_id}\u0000${anchor.field_key}\u0000${anchor.base_row_version}`;
}

export type WorkbookSaveStateDerivationInput = {
  queuedCount: number;
  inFlightCount: number;
  authPaused?: boolean;
  refreshPaused?: boolean;
  pendingMutationCount?: number;
  halted?: PendingReplayHalt | null;
  overflow?: PendingReplayOverflow | null;
  sameFieldConflicts?: readonly PendingReplaySameFieldConflict[];
  localDraftConflicts?: readonly WorkbookSaveStateConflictAnchor[];
};

export type WorkbookSaveStatePresentation = {
  primaryLabel: WorkbookSaveStatePrimaryLabel;
  secondaryKind: WorkbookSaveStateSecondaryKind | null;
  secondaryMessage: string | null;
  conflictAnchors: WorkbookSaveStateConflictAnchor[];
};

export type PendingReplayHalt = {
  unit_id: string;
  error_code: string;
  message: string;
  anchor: PendingReplayPublicAnchor;
};

export type PendingReplayOverflow = {
  message: string;
  refused_unit_id: string;
  preserve_visible_edit_as_unsaved: true;
  visible_edit: PendingReplayVisibleEdit | null;
};

export type PendingQueueSnapshot = {
  scope: PendingReplayScope;
  capacity: number;
  units: PendingReplayUnitState[];
  queuedCount: number;
  inFlightCount: number;
  halted: PendingReplayHalt | null;
  authPaused: boolean;
  overflow: PendingReplayOverflow | null;
  sameFieldConflicts: PendingReplaySameFieldConflict[];
  primarySaveStateInput: PendingReplayPrimarySaveState;
  saveStatePresentation: WorkbookSaveStatePresentation;
};

export type PendingQueueAdmissionResult =
  | {
      accepted: true;
      status: "accepted";
      unit: PendingReplayUnitState;
      snapshot: PendingQueueSnapshot;
    }
  | {
      accepted: true;
      status: "coalesced";
      coalescedIntoUnitId: string;
      unit: PendingReplayUnitState;
      snapshot: PendingQueueSnapshot;
    }
  | {
      accepted: false;
      status: "duplicate";
      refusedReason: "duplicate";
      unit: PendingReplayUnitState;
      snapshot: PendingQueueSnapshot;
    }
  | {
      accepted: false;
      status: "refused";
      refusedReason: "capacity" | "scope_mismatch";
      refusedUnit: PendingReplayUnitState;
      preserveVisibleEditAsUnsaved: boolean;
      primarySaveStateInput: PendingReplayPrimarySaveState;
      overflowMessage: string | null;
      snapshot: PendingQueueSnapshot;
    };

export type PendingReplayDispatch = {
  unit: PendingReplayUnitState;
  identity: PendingReplayMutationIdentity;
  payloadIntent: PendingReplayPayloadIntent;
  snapshot: PendingQueueSnapshot;
};

export type PendingReplaySettlement =
  | {
      outcome: "auth_paused";
      unit: PendingReplayUnitState;
      snapshot: PendingQueueSnapshot;
    }
  | {
      outcome: "halted";
      unit: PendingReplayUnitState;
      halt: PendingReplayHalt;
      snapshot: PendingQueueSnapshot;
    }
  | {
      outcome: "no_dispatched_unit";
      snapshot: PendingQueueSnapshot;
    }
  | {
      outcome: "retryable_failure";
      unit: PendingReplayUnitState;
      snapshot: PendingQueueSnapshot;
    }
  | {
      outcome: "same_field_conflict";
      unit: PendingReplayUnitState;
      conflict: PendingReplaySameFieldConflict;
      snapshot: PendingQueueSnapshot;
    }
  | {
      outcome: "success";
      unit: PendingReplayUnitState;
      row: {
        record_id: string;
        row_version: number;
      };
      snapshot: PendingQueueSnapshot;
    };

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function publicErrorMessageFromPayload(payload: unknown): string {
  if (!isRecord(payload) || !isRecord(payload.error)) {
    return "Request failed.";
  }
  const error = payload.error;
  const status = typeof error.status === "number" ? error.status : null;
  const message = publicErrorStatusText(
    {
      code: typeof error.code === "string" ? error.code : undefined,
      details: isRecord(error.details) ? error.details : undefined,
      message: typeof error.message === "string" ? error.message : undefined,
      status: status ?? undefined,
    },
    status,
  );
  if (message !== "Request failed.") {
    return message;
  }
  if (typeof error.code === "string" && error.code.trim() !== "") {
    const details = error.details;
    if (
      isRecord(details) &&
      typeof details.reason_code === "string" &&
      details.reason_code.trim() !== ""
    ) {
      return `${error.code}: ${details.reason_code}`;
    }
    return error.code;
  }
  return "Request failed.";
}

function parsePendingReplayPublicConflict(
  value: unknown,
): PendingReplayPublicSameFieldConflict | undefined {
  const conflict = parseSameFieldConflictFields(value);
  if (conflict === null) {
    return undefined;
  }
  return {
    ...conflict,
    conflict_token: conflict.conflict_token,
    record_id: conflict.record_id,
    field_key: conflict.field_key,
    conflict_resolution_class: conflict.conflict_resolution_class,
    base_row_version: conflict.base_row_version,
    current_row_version: conflict.current_row_version,
  };
}

export function parsePendingReplayPublicError(
  payload: unknown,
): PendingReplayPublicError {
  if (!isRecord(payload) || !isRecord(payload.error)) {
    return {
      code: "public_error",
      message: publicErrorMessageFromPayload(payload),
    };
  }

  const error = payload.error;
  const status = typeof error.status === "number" ? error.status : null;
  const details = isRecord(error.details) ? error.details : undefined;
  const parsed: PendingReplayPublicError = {
    code: publicErrorCode({
      code: typeof error.code === "string" ? error.code : undefined,
    }),
  };
  if (typeof error.message === "string" && error.message.trim() !== "") {
    parsed.message = publicErrorStatusText(
      {
        code: parsed.code,
        details,
        message: error.message,
        status: status ?? undefined,
      },
      status,
    );
  }
  if (typeof error.retryable === "boolean") {
    parsed.retryable = error.retryable;
  }
  if (details !== undefined) {
    parsed.details = details;
  }
  const conflict = parsePendingReplayPublicConflict(error.conflict);
  if (conflict !== undefined) {
    parsed.conflict = conflict;
  }
  return parsed;
}

function stableJSONValue(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map((entry) => stableJSONValue(entry));
  }
  if (!isRecord(value)) {
    return value;
  }
  const output: Record<string, unknown> = {};
  for (const key of Object.keys(value).sort()) {
    output[key] = stableJSONValue(value[key]);
  }
  return output;
}

function cloneJSONRecord(
  value: Record<string, unknown>,
): Record<string, unknown> {
  return stableJSONValue(value) as Record<string, unknown>;
}

function cloneVisibleEdit(
  value: PendingReplayVisibleEdit,
): PendingReplayVisibleEdit {
  return {
    fieldKey: value.fieldKey,
    rowKey: value.rowKey,
    value: stableJSONValue(value.value),
  };
}

function clonePresentationHint(
  value: PendingReplayPresentationHint,
): PendingReplayPresentationHint {
  return { ...value };
}

function stableStringify(value: unknown): string {
  return JSON.stringify(stableJSONValue(value));
}

// Dedup queued autosaves by the logical mutation payload, not per-request metadata.
export function buildStableMutationSignature(payload: unknown): string {
  if (!isRecord(payload)) {
    return JSON.stringify(payload);
  }

  const {
    client_txn_id: _clientTxnID,
    base_row_version: _baseRowVersion,
    ...stablePayload
  } = payload;
  return stableStringify(stablePayload);
}

export function mergePendingCollectionActionPayload(
  existing: unknown,
  next: unknown,
): unknown {
  if (
    isRecord(existing) &&
    isRecord(next) &&
    existing.kind === "collection_actions_v1" &&
    next.kind === "collection_actions_v1" &&
    Array.isArray(existing.actions) &&
    Array.isArray(next.actions)
  ) {
    return {
      kind: "collection_actions_v1",
      actions: [...existing.actions, ...next.actions],
    };
  }
  return next;
}

export function mergePendingReplayPayload(
  existing: PendingReplayPayloadIntent,
  next: PendingReplayPayloadIntent,
  kind: PendingReplayKind,
): PendingReplayPayloadIntent {
  if (kind === "create") {
    const merged: PendingReplayPayloadIntent = { ...existing };
    for (const [key, value] of Object.entries(next)) {
      if (key === "client_txn_id") {
        continue;
      }
      merged[key] = mergePendingCollectionActionPayload(merged[key], value);
    }
    return merged;
  }

  const existingChanges = Array.isArray(existing.changes)
    ? existing.changes.filter(isRecord)
    : [];
  const nextChanges = Array.isArray(next.changes)
    ? next.changes.filter(isRecord)
    : [];
  const mergedByField = new Map<string, Record<string, unknown>>();
  for (const change of existingChanges) {
    const fieldKey = change.field_key;
    if (typeof fieldKey === "string") {
      mergedByField.set(fieldKey, { ...change });
    }
  }
  for (const change of nextChanges) {
    const fieldKey = change.field_key;
    if (typeof fieldKey !== "string") {
      continue;
    }
    const existingChange = mergedByField.get(fieldKey);
    if (
      existingChange &&
      "action_payload" in existingChange &&
      "action_payload" in change
    ) {
      mergedByField.set(fieldKey, {
        ...existingChange,
        action_payload: mergePendingCollectionActionPayload(
          existingChange.action_payload,
          change.action_payload,
        ),
      });
      continue;
    }
    mergedByField.set(fieldKey, { ...change });
  }
  return {
    ...existing,
    changes: Array.from(mergedByField.values()).sort((left, right) =>
      String(left.field_key).localeCompare(String(right.field_key)),
    ),
  };
}

export function canonicalizePendingReplayChanges(
  payloadIntent: PendingReplayPayloadIntent,
): PendingReplayCanonicalChange[] {
  const rawChanges = Array.isArray(payloadIntent.changes)
    ? payloadIntent.changes.filter(isRecord)
    : [];
  return rawChanges
    .map((change) => {
      const canonical = stableJSONValue(change) as Record<string, unknown>;
      return {
        ...canonical,
        field_key: typeof change.field_key === "string" ? change.field_key : "",
      };
    })
    .sort((left, right) => left.field_key.localeCompare(right.field_key));
}

export function buildPendingReplayMutationIdentity(
  unit: Pick<
    PendingReplayUnitState,
    | "clientTxnId"
    | "incidentId"
    | "kind"
    | "method"
    | "path"
    | "payloadIntent"
    | "recordId"
    | "viewSchemaId"
  >,
): PendingReplayMutationIdentity {
  if (unit.kind === "create") {
    return {
      kind: "create",
      method: "POST",
      route_scope: {
        incident_id: unit.incidentId,
        path: unit.path,
        view_schema_id: unit.viewSchemaId,
      },
      client_txn_id: unit.clientTxnId,
    };
  }

  const baseRowVersion = unit.payloadIntent.base_row_version;
  return {
    kind: "patch",
    method: "PATCH",
    route_scope: {
      path: unit.path,
      record_id: unit.recordId ?? "",
    },
    record_id: unit.recordId ?? "",
    client_txn_id: unit.clientTxnId,
    view_schema_id: unit.viewSchemaId,
    base_row_version:
      typeof baseRowVersion === "number" ? baseRowVersion : null,
    changes: canonicalizePendingReplayChanges(unit.payloadIntent),
  };
}

function cloneIdentity(
  identity: PendingReplayMutationIdentity,
): PendingReplayMutationIdentity {
  if (identity.kind === "create") {
    return {
      kind: "create",
      method: "POST",
      route_scope: { ...identity.route_scope },
      client_txn_id: identity.client_txn_id,
    };
  }
  return {
    kind: "patch",
    method: "PATCH",
    route_scope: { ...identity.route_scope },
    record_id: identity.record_id,
    client_txn_id: identity.client_txn_id,
    view_schema_id: identity.view_schema_id,
    base_row_version: identity.base_row_version,
    changes: identity.changes.map((change) => ({
      ...cloneJSONRecord(change),
      field_key: change.field_key,
    })),
  };
}

function cloneUnit(unit: PendingReplayUnitState): PendingReplayUnitState {
  const output: PendingReplayUnitState = {
    id: unit.id,
    kind: unit.kind,
    source: unit.source,
    incidentId: unit.incidentId,
    clientInstanceId: unit.clientInstanceId,
    viewSchemaId: unit.viewSchemaId,
    rowKey: unit.rowKey,
    recordId: unit.recordId,
    method: unit.method,
    path: unit.path,
    payloadIntent: cloneJSONRecord(unit.payloadIntent),
    clientTxnId: unit.clientTxnId,
    mutationSignature: unit.mutationSignature,
    coalesceKey: unit.coalesceKey,
    enqueueOrder: unit.enqueueOrder,
    operationClass: unit.operationClass,
    status: unit.status,
    identity: cloneIdentity(unit.identity),
  };
  if (unit.presentationHint !== undefined) {
    output.presentationHint = clonePresentationHint(unit.presentationHint);
  }
  return output;
}

function normalizeUnit(input: PendingReplayUnitInput): PendingReplayUnitState {
  const unit: PendingReplayUnitState = {
    id: input.id,
    kind: input.kind,
    source: input.source,
    incidentId: input.incidentId,
    clientInstanceId: input.clientInstanceId,
    viewSchemaId: input.viewSchemaId,
    rowKey: input.rowKey,
    recordId: input.recordId,
    method: input.method,
    path: input.path,
    payloadIntent: cloneJSONRecord(input.payloadIntent),
    clientTxnId: input.clientTxnId,
    mutationSignature:
      input.mutationSignature ??
      buildStableMutationSignature(input.payloadIntent),
    coalesceKey: input.coalesceKey,
    enqueueOrder: input.enqueueOrder,
    operationClass: input.operationClass ?? "hot_path",
    status: input.status ?? "queued",
    identity: {
      kind: "create",
      method: "POST",
      route_scope: {
        incident_id: input.incidentId,
        path: input.path,
        view_schema_id: input.viewSchemaId,
      },
      client_txn_id: input.clientTxnId,
    },
  };
  if (input.presentationHint !== undefined) {
    unit.presentationHint = clonePresentationHint(input.presentationHint);
  }
  unit.identity = buildPendingReplayMutationIdentity(unit);
  return unit;
}

function cloneHalt(value: PendingReplayHalt): PendingReplayHalt {
  return {
    unit_id: value.unit_id,
    error_code: value.error_code,
    message: value.message,
    anchor: cloneAnchor(value.anchor),
  };
}

function cloneAnchor(
  anchor: PendingReplayPublicAnchor,
): PendingReplayPublicAnchor {
  if (anchor.kind === "cell") {
    return {
      kind: "cell",
      record_id: anchor.record_id,
      field_key: anchor.field_key,
    };
  }
  if (anchor.kind === "mutation") {
    return {
      kind: "mutation",
      client_txn_id: anchor.client_txn_id,
      route_scope: cloneIdentity(anchor.route_scope),
    };
  }
  if (anchor.kind === "record") {
    return {
      kind: "record",
      record_id: anchor.record_id,
    };
  }
  return { kind: "surface" };
}

function cloneConflict(
  value: PendingReplaySameFieldConflict,
): PendingReplaySameFieldConflict {
  return {
    key: value.key,
    conflict_token: value.conflict_token,
    record_id: value.record_id,
    field_key: value.field_key,
    conflict_resolution_class: value.conflict_resolution_class,
    base_row_version: value.base_row_version,
    current_row_version: value.current_row_version,
  };
}

function cloneOverflow(value: PendingReplayOverflow): PendingReplayOverflow {
  return {
    message: value.message,
    refused_unit_id: value.refused_unit_id,
    preserve_visible_edit_as_unsaved: true,
    visible_edit:
      value.visible_edit === null ? null : cloneVisibleEdit(value.visible_edit),
  };
}

function conflictAnchorFromSameFieldConflict(
  conflict: PendingReplaySameFieldConflict,
): WorkbookSaveStateConflictAnchor {
  return {
    record_id: conflict.record_id,
    field_key: conflict.field_key,
    base_row_version: conflict.base_row_version,
    current_row_version: conflict.current_row_version,
  };
}

function cloneSaveStateConflictAnchor(
  anchor: WorkbookSaveStateConflictAnchor,
): WorkbookSaveStateConflictAnchor {
  const output: WorkbookSaveStateConflictAnchor = {
    record_id: anchor.record_id,
    field_key: anchor.field_key,
    base_row_version: anchor.base_row_version,
  };
  if (anchor.current_row_version !== undefined) {
    output.current_row_version = anchor.current_row_version;
  }
  return output;
}

function normalizeSaveStateConflictAnchors(
  value: readonly WorkbookSaveStateConflictAnchor[] | undefined,
): WorkbookSaveStateConflictAnchor[] {
  return (value ?? []).map((anchor) => cloneSaveStateConflictAnchor(anchor));
}

function dedupeSaveStateConflictAnchors(
  value: readonly WorkbookSaveStateConflictAnchor[],
): WorkbookSaveStateConflictAnchor[] {
  const seen = new Set<string>();
  const output: WorkbookSaveStateConflictAnchor[] = [];
  for (const anchor of value) {
    const identity = workbookSaveStateConflictAnchorIdentity(anchor);
    if (seen.has(identity)) {
      continue;
    }
    seen.add(identity);
    output.push(anchor);
  }
  return output;
}

function sameFieldConflictSecondaryMessage(conflictCount: number): string {
  return conflictCount === 1
    ? "1 same-field conflict needs review."
    : `${conflictCount} same-field conflicts need review.`;
}

export function deriveWorkbookSaveState(
  input: WorkbookSaveStateDerivationInput,
): WorkbookSaveStatePresentation {
  const sameFieldConflictAnchors = (input.sameFieldConflicts ?? []).map(
    (conflict) => conflictAnchorFromSameFieldConflict(conflict),
  );
  const conflictAnchors = dedupeSaveStateConflictAnchors([
    ...sameFieldConflictAnchors,
    ...normalizeSaveStateConflictAnchors(input.localDraftConflicts),
  ]);

  if (conflictAnchors.length > 0) {
    return {
      primaryLabel: "Conflict",
      secondaryKind: "same_field_conflict",
      secondaryMessage: sameFieldConflictSecondaryMessage(
        conflictAnchors.length,
      ),
      conflictAnchors,
    };
  }

  if (input.overflow !== null && input.overflow !== undefined) {
    return {
      primaryLabel: "Conflict",
      secondaryKind: "overflow",
      secondaryMessage: input.overflow.message,
      conflictAnchors,
    };
  }

  if (input.halted !== null && input.halted !== undefined) {
    return {
      primaryLabel: "Conflict",
      secondaryKind: "replay_halted",
      secondaryMessage: input.halted.message,
      conflictAnchors,
    };
  }

  if (input.authPaused === true) {
    return {
      primaryLabel: "Syncing",
      secondaryKind: "auth_paused",
      secondaryMessage:
        "Authentication is required before queued edits can replay.",
      conflictAnchors,
    };
  }

  if (input.refreshPaused === true) {
    return {
      primaryLabel: "Syncing",
      secondaryKind: "queued",
      secondaryMessage: "Queued edits are waiting for workbook refresh.",
      conflictAnchors,
    };
  }

  const pendingMutationCount = input.pendingMutationCount ?? 0;
  const activeMutationCount =
    input.queuedCount + input.inFlightCount + pendingMutationCount;
  if (activeMutationCount > 0) {
    return {
      primaryLabel: "Syncing",
      secondaryKind: "queued",
      secondaryMessage:
        input.queuedCount + input.inFlightCount > 0
          ? "Queued edits are waiting to replay."
          : "Workbook edits are syncing.",
      conflictAnchors,
    };
  }

  return {
    primaryLabel: "Saved",
    secondaryKind: null,
    secondaryMessage: null,
    conflictAnchors,
  };
}

function stringDetail(
  details: Record<string, unknown> | undefined,
  keys: string[],
): string | null {
  if (details === undefined) {
    return null;
  }
  for (const key of keys) {
    const value = details[key];
    if (typeof value === "string" && value.trim() !== "") {
      return value;
    }
  }
  return null;
}

function failureAnchor(
  unit: PendingReplayUnitState,
  error: PendingReplayPublicError,
): PendingReplayPublicAnchor {
  const details = error.details;
  const fieldKey = stringDetail(details, ["field_key", "field", "fieldKey"]);
  const recordId =
    stringDetail(details, ["record_id", "recordId"]) ?? unit.recordId;

  if (fieldKey !== null) {
    return {
      kind: "cell",
      record_id: recordId,
      field_key: fieldKey,
    };
  }
  if (error.code === "row_version_conflict") {
    return {
      kind: "record",
      record_id: recordId,
    };
  }
  return {
    kind: "mutation",
    client_txn_id: unit.clientTxnId,
    route_scope: cloneIdentity(unit.identity),
  };
}

function sameFieldConflictAnchor(
  error: PendingReplayPublicError,
): PendingReplaySameFieldConflict | null {
  const conflict = parseSameFieldConflictFields(
    error.conflict,
  ) as PendingReplayPublicSameFieldConflict | null;
  if (conflict === null) {
    return null;
  }
  return {
    key: sameFieldConflictQueueKey(conflict),
    conflict_token: conflict.conflict_token,
    record_id: conflict.record_id,
    field_key: conflict.field_key,
    conflict_resolution_class: conflict.conflict_resolution_class,
    base_row_version: conflict.base_row_version,
    current_row_version: conflict.current_row_version,
  };
}

function isRetryableStatus(status: number): boolean {
  return status === 0 || status === 408 || status === 425 || status >= 500;
}

function isAuthFailure(status: number, code: string): boolean {
  return (
    status === 401 ||
    status === 403 ||
    code === "auth_required" ||
    code === "session_required" ||
    code === "session_revoked"
  );
}

export function shouldRetryPendingFailure(
  status: number,
  error: PendingReplayPublicError,
): boolean {
  if (isRetryableStatus(status)) {
    return true;
  }
  return (
    error.retryable === true &&
    ![
      "client_txn_conflict",
      "invalid_mutation_payload",
      "row_version_conflict",
      "same_field_conflict",
    ].includes(error.code)
  );
}

export type PendingReplayCoalescingUnit = {
  readonly status: PendingReplayStatus;
  readonly operationClass: PendingReplayOperationClass;
  readonly kind: PendingReplayKind;
  readonly incidentId: string;
  readonly clientInstanceId: string;
  readonly coalesceKey: string;
  readonly recordId: string | null;
};

export function canCoalescePendingReplayUnits(
  existing: PendingReplayCoalescingUnit,
  next: PendingReplayCoalescingUnit,
): boolean {
  if (
    existing.status !== "queued" ||
    existing.operationClass !== "hot_path" ||
    next.operationClass !== "hot_path" ||
    existing.kind !== next.kind ||
    existing.incidentId !== next.incidentId ||
    existing.clientInstanceId !== next.clientInstanceId ||
    existing.coalesceKey !== next.coalesceKey
  ) {
    return false;
  }
  return (
    next.kind === "create" ||
    (next.recordId !== null && existing.recordId === next.recordId)
  );
}

export class WorkbookPendingQueueModel {
  readonly scope: PendingReplayScope;

  private units: PendingReplayUnitState[] = [];
  private halted: PendingReplayHalt | null = null;
  private authPaused = false;
  private overflow: PendingReplayOverflow | null = null;
  private readonly sameFieldConflicts: PendingReplaySameFieldConflict[] = [];

  constructor(scope: PendingReplayScope) {
    this.scope = { ...scope };
  }

  snapshot(): PendingQueueSnapshot {
    const queuedCount = this.units.filter(
      (unit) => unit.status === "queued",
    ).length;
    const inFlightCount = this.units.filter(
      (unit) => unit.status === "in_flight",
    ).length;
    const saveStatePresentation = deriveWorkbookSaveState({
      authPaused: this.authPaused,
      halted: this.halted,
      overflow: this.overflow,
      sameFieldConflicts: this.sameFieldConflicts,
      queuedCount,
      inFlightCount,
    });
    return {
      scope: { ...this.scope },
      capacity: pendingReplayCapacity,
      units: this.units.map((unit) => cloneUnit(unit)),
      queuedCount,
      inFlightCount,
      halted: this.halted === null ? null : cloneHalt(this.halted),
      authPaused: this.authPaused,
      overflow: this.overflow === null ? null : cloneOverflow(this.overflow),
      sameFieldConflicts: this.sameFieldConflicts.map((conflict) =>
        cloneConflict(conflict),
      ),
      primarySaveStateInput: saveStatePresentation.primaryLabel,
      saveStatePresentation: {
        primaryLabel: saveStatePresentation.primaryLabel,
        secondaryKind: saveStatePresentation.secondaryKind,
        secondaryMessage: saveStatePresentation.secondaryMessage,
        conflictAnchors: saveStatePresentation.conflictAnchors.map((anchor) =>
          cloneSaveStateConflictAnchor(anchor),
        ),
      },
    };
  }

  private isReplayBlocked(): boolean {
    return (
      this.authPaused ||
      this.halted !== null ||
      this.sameFieldConflicts.length > 0 ||
      this.units.some((unit) => unit.status === "in_flight")
    );
  }

  admit(input: PendingReplayUnitInput): PendingQueueAdmissionResult {
    this.overflow = null;
    const unit = normalizeUnit(input);
    if (
      unit.incidentId !== this.scope.incidentId ||
      unit.clientInstanceId !== this.scope.clientInstanceId
    ) {
      return {
        accepted: false,
        status: "refused",
        refusedReason: "scope_mismatch",
        refusedUnit: cloneUnit(unit),
        preserveVisibleEditAsUnsaved: false,
        primarySaveStateInput: this.snapshot().primarySaveStateInput,
        overflowMessage: null,
        snapshot: this.snapshot(),
      };
    }

    const duplicate = this.units.some(
      (candidate) =>
        candidate.rowKey === unit.rowKey &&
        candidate.mutationSignature === unit.mutationSignature,
    );
    if (duplicate) {
      return {
        accepted: false,
        status: "duplicate",
        refusedReason: "duplicate",
        unit: cloneUnit(unit),
        snapshot: this.snapshot(),
      };
    }

    const lastUnit = this.units[this.units.length - 1];
    if (
      lastUnit !== undefined &&
      canCoalescePendingReplayUnits(lastUnit, unit)
    ) {
      lastUnit.payloadIntent = mergePendingReplayPayload(
        lastUnit.payloadIntent,
        unit.payloadIntent,
        unit.kind,
      );
      lastUnit.mutationSignature = buildStableMutationSignature(
        lastUnit.payloadIntent,
      );
      lastUnit.identity = buildPendingReplayMutationIdentity(lastUnit);
      return {
        accepted: true,
        status: "coalesced",
        coalescedIntoUnitId: lastUnit.id,
        unit: cloneUnit(lastUnit),
        snapshot: this.snapshot(),
      };
    }

    if (this.units.length >= pendingReplayCapacity) {
      const visibleEdit =
        input.visibleEdit === undefined
          ? null
          : cloneVisibleEdit(input.visibleEdit);
      this.overflow = {
        message:
          "Local pending queue is full. The current edit remains unsaved local work.",
        refused_unit_id: unit.id,
        preserve_visible_edit_as_unsaved: true,
        visible_edit: visibleEdit,
      };
      return {
        accepted: false,
        status: "refused",
        refusedReason: "capacity",
        refusedUnit: cloneUnit(unit),
        preserveVisibleEditAsUnsaved: true,
        primarySaveStateInput: "Conflict",
        overflowMessage: this.overflow.message,
        snapshot: this.snapshot(),
      };
    }

    this.units.push(unit);
    return {
      accepted: true,
      status: "accepted",
      unit: cloneUnit(unit),
      snapshot: this.snapshot(),
    };
  }

  peekNextQueued(): PendingReplayDispatch | null {
    if (this.isReplayBlocked()) {
      return null;
    }
    const unit = this.units.find((candidate) => candidate.status === "queued");
    if (unit === undefined) {
      return null;
    }
    return {
      unit: cloneUnit(unit),
      identity: cloneIdentity(unit.identity),
      payloadIntent: cloneJSONRecord(unit.payloadIntent),
      snapshot: this.snapshot(),
    };
  }

  markDispatched(unitId: string): PendingReplayDispatch | null {
    if (this.isReplayBlocked()) {
      return null;
    }
    const unit = this.units.find((candidate) => candidate.status === "queued");
    if (unit === undefined || unit.id !== unitId) {
      return null;
    }
    unit.status = "in_flight";
    return {
      unit: cloneUnit(unit),
      identity: cloneIdentity(unit.identity),
      payloadIntent: cloneJSONRecord(unit.payloadIntent),
      snapshot: this.snapshot(),
    };
  }

  dispatchNext(): PendingReplayDispatch | null {
    const next = this.peekNextQueued();
    return next === null ? null : this.markDispatched(next.unit.id);
  }

  settleDispatched(result: PendingReplayPublicResult): PendingReplaySettlement {
    const unit = this.units.find(
      (candidate) => candidate.status === "in_flight",
    );
    if (unit === undefined) {
      return {
        outcome: "no_dispatched_unit",
        snapshot: this.snapshot(),
      };
    }

    if (result.ok) {
      const completedUnit = cloneUnit(unit);
      this.units = this.units.filter((candidate) => candidate !== unit);
      this.authPaused = false;
      this.halted = null;
      return {
        outcome: "success",
        unit: completedUnit,
        row: {
          record_id: result.row.record_id,
          row_version: result.row.row_version,
        },
        snapshot: this.snapshot(),
      };
    }

    unit.status = "queued";
    if (isAuthFailure(result.status, result.error.code)) {
      this.authPaused = true;
      this.halted = null;
      return {
        outcome: "auth_paused",
        unit: cloneUnit(unit),
        snapshot: this.snapshot(),
      };
    }

    if (result.error.code === "same_field_conflict") {
      const conflict = sameFieldConflictAnchor(result.error);
      if (conflict !== null) {
        this.units = this.units.filter((candidate) => candidate !== unit);
        this.sameFieldConflicts.push(conflict);
        return {
          outcome: "same_field_conflict",
          unit: cloneUnit(unit),
          conflict: cloneConflict(conflict),
          snapshot: this.snapshot(),
        };
      }
    }

    if (shouldRetryPendingFailure(result.status, result.error)) {
      return {
        outcome: "retryable_failure",
        unit: cloneUnit(unit),
        snapshot: this.snapshot(),
      };
    }

    this.halted = {
      unit_id: unit.id,
      error_code: result.error.code,
      message: publicErrorStatusText(
        {
          code: result.error.code,
          details: result.error.details,
          message: result.error.message,
          status: result.status,
        },
        result.status,
      ),
      anchor: failureAnchor(unit, result.error),
    };
    return {
      outcome: "halted",
      unit: cloneUnit(unit),
      halt: cloneHalt(this.halted),
      snapshot: this.snapshot(),
    };
  }

  resumeAfterAuthRecovery(): PendingQueueSnapshot {
    this.authPaused = false;
    return this.snapshot();
  }

  pauseForAuthRecovery(): PendingQueueSnapshot {
    this.authPaused = true;
    this.halted = null;
    return this.snapshot();
  }

  clearSameFieldConflict(key: string): PendingQueueSnapshot {
    const conflictIndex = this.sameFieldConflicts.findIndex(
      (conflict) => conflict.key === key,
    );
    if (conflictIndex >= 0) {
      this.sameFieldConflicts.splice(conflictIndex, 1);
    }
    return this.snapshot();
  }
}
