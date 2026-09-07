import type {
  ReferencePackAction,
  ReferencePackAttempt,
  ReferencePackCommand,
  ReferencePackJobResource,
  ReferencePackPaging,
  ReferencePackProblem,
  ReferencePackQuery,
  ReferencePackVersion,
} from "../services/referencePacks";
import { resolvePublicErrorPresentation } from "../shared/publicErrorPresentation";

export type ReferencePackAuthority = {
  readonly lifetime: string;
  readonly actorId: string;
};
export const terminalReferencePackJobStates = new Set<
  ReferencePackJobResource["status"]
>(["succeeded", "failed", "canceled"]);
export const defaultReferencePackPaging: ReferencePackPaging = {
  limit: 100,
  has_more: false,
  next_cursor: null,
};
export const emptyReferencePackQuery: ReferencePackQuery = {
  active: "",
  packVersionState: "",
  search: "",
  verificationResult: "",
};
export function normalizeReferencePackQuery(
  input: ReferencePackQuery,
): ReferencePackQuery {
  return {
    ...input,
    search: input.search
      .normalize("NFC")
      .replace(/^\p{White_Space}+|\p{White_Space}+$/gu, ""),
  };
}
export function sameReferencePackQuery(
  a: ReferencePackQuery,
  b: ReferencePackQuery,
) {
  return (
    a.search === b.search &&
    a.active === b.active &&
    a.packVersionState === b.packVersionState &&
    a.verificationResult === b.verificationResult
  );
}
export function referencePackIdentity(pack: {
  readonly pack_key: string;
  readonly pack_version: string;
}) {
  return JSON.stringify([pack.pack_key, pack.pack_version]);
}
export function referencePackEligible(
  pack: ReferencePackVersion,
  action: ReferencePackAction,
) {
  if (action === "activate")
    return pack.pack_version_state === "verified_available" && !pack.active;
  if (action === "disable")
    return pack.pack_version_state === "verified_available";
  return pack.pack_version_state !== "staged";
}
export type ReferencePackOperation = {
  readonly id: string;
  readonly command: ReferencePackCommand;
  readonly attempt: ReferencePackAttempt | null;
  readonly phase:
    | "checking"
    | "submitting"
    | "replaying"
    | "accepted"
    | "committed"
    | "rejected"
    | "uncertain"
    | "failed"
    | "canceled";
  readonly jobId: string | null;
  readonly problem: ReferencePackProblem | null;
};
export type ReferencePackKnownJob = {
  readonly operationId: string;
  readonly command: ReferencePackCommand;
  readonly snapshot: ReferencePackJobResource;
  readonly observation:
    | "idle"
    | "reading"
    | "paused"
    | "failed"
    | "unavailable";
  readonly problem: ReferencePackProblem | null;
  readonly cancellation: {
    readonly request: Readonly<{ client_txn_id: string }>;
    readonly phase:
      | "checking"
      | "submitting"
      | "uncertain"
      | "rejected"
      | "acknowledged";
    readonly problem: ReferencePackProblem | null;
  } | null;
};
export type ReferencePackCatalog = {
  readonly rows: readonly ReferencePackVersion[];
  readonly paging: ReferencePackPaging;
  readonly acceptedQuery: ReferencePackQuery | null;
  readonly generation: number;
  readonly invalidation: number;
  readonly pending: {
    readonly generation: number;
    readonly query: ReferencePackQuery;
    readonly append: boolean;
    readonly invalidation: number;
  } | null;
  readonly problem: ReferencePackProblem | null;
  readonly dirty: boolean;
};
export type ReferencePackAdminState = {
  readonly authority: ReferencePackAuthority | null;
  readonly active: boolean;
  readonly access: "ready" | "checking" | "unavailable" | "denied";
  readonly reconciling: boolean;
  readonly input: ReferencePackQuery;
  readonly query: ReferencePackQuery;
  readonly catalog: ReferencePackCatalog;
  readonly selectedKeys: readonly string[];
  readonly file: File | null;
  readonly operation: ReferencePackOperation | null;
  readonly jobs: Readonly<Record<string, ReferencePackKnownJob>>;
  readonly scrollTop: number;
  readonly announcement: { readonly serial: number; readonly text: string };
};
export function initialReferencePackState(): ReferencePackAdminState {
  return {
    authority: null,
    active: false,
    access: "denied",
    reconciling: false,
    input: { ...emptyReferencePackQuery },
    query: { ...emptyReferencePackQuery },
    catalog: {
      rows: [],
      paging: defaultReferencePackPaging,
      acceptedQuery: null,
      generation: 0,
      invalidation: 0,
      pending: null,
      problem: null,
      dirty: true,
    },
    selectedKeys: [],
    file: null,
    operation: null,
    jobs: {},
    scrollTop: 0,
    announcement: { serial: 0, text: "" },
  };
}
export function referencePackBusy(state: ReferencePackAdminState) {
  return (
    (state.operation !== null &&
      ["checking", "submitting", "replaying", "uncertain", "accepted"].includes(
        state.operation.phase,
      )) ||
    Object.values(state.jobs).some(
      (job) => !terminalReferencePackJobStates.has(job.snapshot.status),
    )
  );
}
export function referencePackListStatus(state: ReferencePackAdminState) {
  const catalog = state.catalog;
  if (catalog.pending) return "Searching reference packs";
  if (catalog.problem)
    return catalog.acceptedQuery === null
      ? "Reference packs unavailable"
      : "Previously loaded reference packs may be stale";
  if (catalog.acceptedQuery === null) return "Reference packs not loaded";
  if (!catalog.rows.length)
    return sameReferencePackQuery(
      catalog.acceptedQuery,
      emptyReferencePackQuery,
    )
      ? "No reference packs imported"
      : "No reference packs match these filters";
  return catalog.dirty
    ? "Reference pack catalog needs reloading"
    : "Reference packs loaded";
}
export function referencePackCommandLabel(command: ReferencePackCommand) {
  switch (command.kind) {
    case "import":
      return `Import ${command.filename}`;
    case "refresh_all":
      return "Refresh all reference packs";
    case "refresh_selected":
      return `Refresh ${command.packKeys.length} selected pack ${command.packKeys.length === 1 ? "key" : "keys"}`;
    default:
      return `${{ activate: "Activate", disable: "Disable", reverify: "Reverify" }[command.kind]} ${command.target.pack_key}@${command.target.pack_version}`;
  }
}
export function referencePackProblemText(problem: ReferencePackProblem) {
  switch (problem.kind) {
    case "state_conflict":
      return "The pack state changed. Review its current state before choosing another action.";
    case "activation_rejected":
      return problem.reason === "already_active"
        ? "This exact version is already active. Reload to review its current state."
        : "This version cannot be activated. Review its verification and availability.";
    case "verification_failed":
      return (
        (
          {
            contract_incompatible:
              "The bundle contract version is not supported by this deployment.",
            checksum_mismatch:
              "The bundle checksum did not match its contents.",
            signature_mismatch: "The bundle signature could not be verified.",
            payload_missing: "A required bundle payload is missing.",
            path_traversal: "The bundle contains an unsafe archive path.",
            disallowed_content: "The bundle contains unsupported content.",
          } as Record<string, string>
        )[problem.reason ?? ""] ??
        "Pack verification failed. Review the bundle and its integrity before another import or verification."
      );
    case "transaction_conflict":
      return "The transaction identifier conflicts with an existing request. Review current state; this request cannot be replayed with different content.";
    case "contract":
      return "The server response could not be validated.";
    case "transport":
      return "The server response was not received.";
    case "unavailable":
      return "The operation is currently unavailable. Its last observed state does not establish its current outcome.";
    default:
      return problem.code === "invalid_list_query"
        ? "The search or filters were rejected. Update the query and try again."
        : problem.code === "job_cancel_rejected"
          ? "Cancellation was rejected. Check the operation's current state."
          : resolvePublicErrorPresentation({
                code: problem.code,
                status: problem.status,
                operationFamily: "field_mutation",
                hasAuthorizedMaterialization: false,
              }).family === "local_validation"
            ? "The supplied values were rejected. Review the fields before trying again."
            : "The request was rejected. Review current state before trying again.";
  }
}
export function catalogStarted(
  state: ReferencePackAdminState,
  generation: number,
  append: boolean,
): ReferencePackAdminState {
  return {
    ...state,
    catalog: {
      ...state.catalog,
      generation,
      problem: null,
      paging: append ? state.catalog.paging : defaultReferencePackPaging,
      pending: {
        generation,
        append,
        query: { ...state.query },
        invalidation: state.catalog.invalidation,
      },
    },
  };
}
export function catalogAccepted(
  state: ReferencePackAdminState,
  generation: number,
  rows: readonly ReferencePackVersion[],
  paging: ReferencePackPaging,
): ReferencePackAdminState {
  const pending = state.catalog.pending;
  if (
    !pending ||
    pending.generation !== generation ||
    state.catalog.generation !== generation
  )
    return state;
  const merged = pending.append ? [...state.catalog.rows] : [];
  const positions = new Map(
    merged.map((row, index) => [referencePackIdentity(row), index]),
  );
  for (const row of rows) {
    const key = referencePackIdentity(row);
    const index = positions.get(key);
    if (index !== undefined) merged[index] = row;
    else {
      positions.set(key, merged.length);
      merged.push(row);
    }
  }
  return {
    ...state,
    catalog: {
      ...state.catalog,
      rows: merged,
      paging,
      acceptedQuery: pending.query,
      pending: null,
      problem: null,
      dirty: pending.invalidation !== state.catalog.invalidation,
    },
  };
}
export function catalogFailed(
  state: ReferencePackAdminState,
  generation: number,
  problem: ReferencePackProblem,
): ReferencePackAdminState {
  if (
    state.catalog.pending?.generation !== generation ||
    state.catalog.generation !== generation
  )
    return state;
  return {
    ...state,
    catalog: { ...state.catalog, pending: null, problem, dirty: true },
  };
}
export function catalogInvalidated(
  state: ReferencePackAdminState,
): ReferencePackAdminState {
  return {
    ...state,
    catalog: {
      ...state.catalog,
      dirty: true,
      invalidation: state.catalog.invalidation + 1,
      paging: defaultReferencePackPaging,
    },
  };
}
export function selectionChanged(
  state: ReferencePackAdminState,
  key: string,
  selected: boolean,
): ReferencePackAdminState {
  return {
    ...state,
    selectedKeys: selected
      ? [...new Set([...state.selectedKeys, key])].sort()
      : state.selectedKeys.filter((value) => value !== key),
  };
}
const reachableJobStates: Record<
  ReferencePackJobResource["status"],
  readonly ReferencePackJobResource["status"][]
> = {
  queued: [
    "queued",
    "running",
    "cancel_requested",
    "succeeded",
    "failed",
    "canceled",
  ],
  running: ["running", "cancel_requested", "succeeded", "failed", "canceled"],
  cancel_requested: ["cancel_requested", "succeeded", "failed", "canceled"],
  succeeded: ["succeeded"],
  failed: ["failed"],
  canceled: ["canceled"],
};
export function acceptsReferencePackJob(
  prior: ReferencePackJobResource,
  next: ReferencePackJobResource,
) {
  if (
    prior.job_id !== next.job_id ||
    prior.submitted_by_user_id !== next.submitted_by_user_id ||
    prior.submitted_at !== next.submitted_at ||
    prior.status_route !== next.status_route ||
    JSON.stringify(prior.scope) !== JSON.stringify(next.scope) ||
    Date.parse(next.updated_at) < Date.parse(prior.updated_at) ||
    !reachableJobStates[prior.status].includes(next.status) ||
    next.progress.completed < prior.progress.completed ||
    (prior.progress.total !== null &&
      (next.progress.total === null ||
        next.progress.total < prior.progress.total))
  )
    return false;
  if (terminalReferencePackJobStates.has(prior.status))
    return JSON.stringify(prior) === JSON.stringify(next);
  return true;
}

/** Accept one authoritative observation without regressing another job or operation. */
export function jobObserved(
  state: ReferencePackAdminState,
  id: string,
  snapshot: ReferencePackJobResource,
): ReferencePackAdminState {
  const job = state.jobs[id];
  if (!job || !acceptsReferencePackJob(job.snapshot, snapshot)) return state;
  const terminal = terminalReferencePackJobStates.has(snapshot.status);
  return {
    ...state,
    jobs: {
      ...state.jobs,
      [id]: {
        ...job,
        snapshot,
        observation: "idle",
        problem: null,
        cancellation:
          terminal && job.cancellation?.phase === "uncertain"
            ? null
            : job.cancellation,
      },
    },
    operation:
      terminal && state.operation?.jobId === id
        ? {
            ...state.operation,
            phase:
              snapshot.status === "succeeded"
                ? "committed"
                : snapshot.status === "canceled"
                  ? "canceled"
                  : "failed",
          }
        : state.operation,
  };
}
