import { requireViewContract } from "@cartulary/view-contracts";
import { boundedRead } from "../../services/asyncObservation";
import { genericInspectorRowLabel } from "../models/genericWorkbookModel";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";
import type {
  WorkbookReferenceMemberPage,
  WorkbookReferencePage,
  WorkbookReferenceReadPort,
  WorkbookReferenceRequest,
} from "../ports/WorkbookReferenceReadPort";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";

type Result = WorkbookPortResult<WorkbookReferencePage>;
type Entry = {
  controller: AbortController;
  consumers: number;
  settled: boolean;
  promise: Promise<Result>;
};
export interface WorkbookReferenceReader extends WorkbookReferenceReadPort {
  dispose(): void;
}

/** One authority instance, in-flight sharing only. Disposed instances never publish. */
export function createWorkbookReferenceReader(options: {
  readonly authorityScope: string;
  readonly viewQuery: WorkbookViewQueryPort;
  readonly readMembers: (
    cursorToken: string | undefined,
    signal: AbortSignal,
  ) => Promise<WorkbookPortResult<WorkbookReferenceMemberPage>>;
}): WorkbookReferenceReader {
  const inFlight = new Map<string, Entry>();
  let disposed = false;
  const aborted: Result = { kind: "aborted" };
  const read = async (
    input: WorkbookReferenceRequest,
    signal: AbortSignal,
  ): Promise<Result> => {
    if (input.identityKind === "incident_member") {
      if (
        input.viewSchemaId !== "incident_members" ||
        input.queryState.filters.length ||
        input.queryState.sort.length ||
        input.queryState.groupBy
      ) {
        return {
          kind: "rejected",
          failure: {
            kind: "invalid_contract",
            message: "Member discovery does not support record filters.",
          },
        };
      }
      const result = await options.readMembers(input.cursorToken, signal);
      if (result.kind !== "accepted") return result;
      return {
        kind: "accepted",
        value: {
          candidates: result.value.members.map((member) => ({
            identity: { kind: "incident_member", id: member.userId },
            viewSchemaId: "incident_members",
            displayText: member.displayName,
            presentation: "observed",
          })),
          paging: result.value.paging,
          canonicalQuery: null,
          producingRequest: input,
        },
      };
    }
    const contract = requireViewContract(input.viewSchemaId);
    const result = await options.viewQuery.query({
      contract,
      queryState: input.queryState,
      limit: 100,
      signal,
      ...(input.cursorToken ? { cursorToken: input.cursorToken } : {}),
      ...(input.expectedCanonicalQuery
        ? { expectedCanonicalQuery: input.expectedCanonicalQuery }
        : {}),
    });
    if (result.kind !== "accepted") return result;
    return {
      kind: "accepted",
      value: {
        candidates: result.value.rows.map((row) => ({
          identity: { kind: input.identityKind, id: row.record_id },
          viewSchemaId: input.viewSchemaId,
          displayText: genericInspectorRowLabel(contract, row),
          presentation: "observed",
        })),
        canonicalQuery: result.value.canonicalQuery,
        paging: result.value.paging,
        producingRequest: input,
      },
    };
  };
  return {
    page(input, signal) {
      if (disposed || signal.aborted) return Promise.resolve(aborted);
      const key = JSON.stringify([
        options.authorityScope,
        input.identityKind,
        input.viewSchemaId,
        input.queryState,
        100,
        input.cursorToken ?? null,
        input.expectedCanonicalQuery ?? null,
      ]);
      let entry = inFlight.get(key);
      if (!entry || entry.controller.signal.aborted) {
        const controller = new AbortController();
        const current: Entry = {
          controller,
          consumers: 0,
          settled: false,
          promise: Promise.resolve(aborted),
        };
        entry = current;
        inFlight.set(key, current);
        current.promise = boundedRead(
          (readSignal) => read(input, readSignal),
          controller.signal,
        )
          .then(
            (result): Result =>
              disposed || controller.signal.aborted ? aborted : result,
          )
          .catch(
            (): Result =>
              disposed || controller.signal.aborted
                ? aborted
                : {
                    kind: "rejected",
                    failure: {
                      kind: "retryable",
                      message:
                        "Reference read did not complete. Retry the read.",
                    },
                  },
          )
          .finally(() => {
            current.settled = true;
            if (inFlight.get(key) === current) inFlight.delete(key);
          });
      }
      const current = entry;
      current.consumers += 1;
      return new Promise<Result>((resolve) => {
        let released = false;
        const finish = (result: Result) => {
          if (released) return;
          released = true;
          signal.removeEventListener("abort", cancel);
          current.consumers -= 1;
          if (!current.consumers && !current.settled) {
            current.controller.abort();
            if (inFlight.get(key) === current) inFlight.delete(key);
          }
          resolve(
            disposed || signal.aborted || current.controller.signal.aborted
              ? aborted
              : result,
          );
        };
        const cancel = () => finish(aborted);
        signal.addEventListener("abort", cancel, { once: true });
        void current.promise.then(finish);
        if (signal.aborted) cancel();
      });
    },
    dispose() {
      disposed = true;
      for (const entry of inFlight.values()) entry.controller.abort();
      inFlight.clear();
    },
  };
}
