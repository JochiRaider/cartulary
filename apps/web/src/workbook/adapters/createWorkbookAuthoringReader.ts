import type { HTTPOperationResponse } from "@cartulary/protocol-ts/http";
import {
  getViewContract,
  requireViewContract,
} from "@cartulary/view-contracts";
import { genericInspectorRowLabel } from "../models/genericWorkbookModel";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { buildQueryRequest } from "../models/workbookQuery";
import type {
  WorkbookAuthoringCandidate,
  WorkbookAuthoringReadPort,
} from "../ports/WorkbookAuthoringReadPort";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import type { WorkbookCanonicalQuery } from "../query/WorkbookViewQueryPort";
import { readWorkbookQueryMetadata } from "../query/workbookQueryMetadata";
import { workbookCreateCapabilityMatches } from "./workbookCreateCapability";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

/** Target form discovery; page membership never determines reference existence. */
export function createWorkbookAuthoringReader(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly recheckAuthority: () => void;
}): WorkbookAuthoringReadPort {
  const operations = createWorkbookOperationExecutor(options);
  return {
    async availableViews(signal) {
      const result = await operations.execute({
        operationID: "listViewSchemas",
        signal,
      });
      if (signal.aborted) return { kind: "aborted" };
      if (result.kind !== "accepted") return result;
      return {
        kind: "accepted",
        value: result.value.data.view_schemas
          .filter((view) => getViewContract(view.view_schema_id))
          .map((view) => view.view_schema_id),
      };
    },
    async verify(draft, signal) {
      const results = await Promise.all([
        operations.execute({
          operationID: "getViewSchema",
          pathParameters: { view_schema_id: draft.source.viewSchemaId },
          signal,
        }),
        operations.execute({
          operationID: "getViewSchema",
          pathParameters: { view_schema_id: draft.target.viewSchemaId },
          signal,
        }),
      ]);
      const [source, target] = results;
      if (source.kind !== "accepted" || target.kind !== "accepted")
        throw new Error("Creation capability could not be verified.");
      const feature = source.value.data.inspector_config.feature_groups.find(
        (item) => item.feature_group_key === draft.feature.featureGroupKey,
      );
      const route = feature?.route_binding;
      const schema = target.value.data;
      if (
        !feature ||
        route?.kind !== "view_row_create" ||
        route.owner !== "view_row_create_route" ||
        route.action_key !== draft.feature.routeBinding.actionKey ||
        route.target_view_schema_id !== draft.target.viewSchemaId ||
        source.value.data.view_schema_id !== draft.source.viewSchemaId ||
        feature.minimum_incident_role !== draft.feature.minimumIncidentRole ||
        feature.requires_confirmation !== draft.feature.requiresConfirmation ||
        feature.mutates !== draft.feature.mutates ||
        JSON.stringify(feature.disabled_when) !==
          JSON.stringify(draft.feature.disabledWhen) ||
        feature.success_result_behavior !==
          draft.feature.successResultBehavior ||
        feature.failure_result_behavior !==
          draft.feature.failureResultBehavior ||
        !workbookCreateCapabilityMatches(schema, draft.target) ||
        JSON.stringify(
          feature.seed_bindings.map((binding) => [
            binding.target_field_key ?? null,
            binding.target_input_key ?? null,
            binding.source.kind,
            binding.source.source_field_key ?? null,
            binding.source.value ?? null,
          ]),
        ) !==
          JSON.stringify(
            draft.feature.seedBindings.map((binding) => [
              binding.targetFieldKey ?? null,
              binding.targetInputKey ?? null,
              binding.source.kind,
              binding.source.sourceFieldKey ?? null,
              binding.source.value ?? null,
            ]),
          )
      )
        throw new Error(
          "Creation capability changed. Review the retained draft.",
        );
    },
    async page(input) {
      let responseAccepted = false;
      try {
        let candidates: WorkbookAuthoringCandidate[];
        let canonicalQuery: WorkbookCanonicalQuery | undefined;
        let paging: HTTPOperationResponse<"listIncidentMemberships">["meta"]["paging"];
        if (input.viewSchemaId === "incident_members") {
          const result = await operations.execute({
            operationID: "listIncidentMemberships",
            pathParameters: { incident_id: options.incidentId },
            query: {
              limit: 100,
              ...(input.cursor ? { cursor_token: input.cursor } : {}),
            },
            signal: input.signal,
          });
          if (input.signal.aborted || input.isCurrent?.() === false)
            return { kind: "aborted" };
          if (result.kind !== "accepted") {
            if (
              result.kind === "rejected" &&
              workbookFailureLifecycle(result.failure).kind ===
                "authority_unavailable"
            )
              (input.onAuthorityFailure ?? options.recheckAuthority)(
                result.failure,
              );
            return result;
          }
          responseAccepted = true;
          paging = result.value.meta.paging;
          candidates = result.value.data.memberships.map((member) => {
            if (member.incident_id !== options.incidentId || !member.user_id)
              throw new Error("Invalid membership.");
            return {
              recordId: member.user_id,
              displayText: member.display_name,
              viewSchemaId: input.viewSchemaId,
            };
          });
        } else {
          const contract = requireViewContract(input.viewSchemaId);
          const result = await operations.execute({
            operationID: "queryWorkbookView",
            pathParameters: {
              incident_id: options.incidentId,
              view_schema_id: input.viewSchemaId,
            },
            request: {
              ...buildQueryRequest(contract, input.queryState),
              limit: 100,
              ...(input.cursor ? { cursor_token: input.cursor } : {}),
            },
            signal: input.signal,
          });
          if (input.signal.aborted || input.isCurrent?.() === false)
            return { kind: "aborted" };
          if (result.kind !== "accepted") {
            if (
              result.kind === "rejected" &&
              workbookFailureLifecycle(result.failure).kind ===
                "authority_unavailable"
            )
              (input.onAuthorityFailure ?? options.recheckAuthority)(
                result.failure,
              );
            return result;
          }
          responseAccepted = true;
          if (
            result.value.data.incident_id !== options.incidentId ||
            result.value.data.view_schema_id !== input.viewSchemaId
          )
            throw new Error("Invalid candidate scope.");
          canonicalQuery = readWorkbookQueryMetadata(
            contract,
            result.value.meta,
            input.queryState,
            100,
            input.cursor ?? undefined,
            input.expectedCanonicalQuery,
          ).canonicalQuery;
          paging = result.value.meta.paging;
          candidates = normalizeWorkbookViewRows(
            contract,
            result.value.data.rows,
            "Contextual references",
          ).map((row) => {
            return {
              recordId: row.record_id,
              displayText: genericInspectorRowLabel(contract, row),
              viewSchemaId: contract.viewSchemaId,
              row,
            };
          });
        }
        if (
          candidates.length > 100 ||
          new Set(candidates.map((item) => item.recordId)).size !==
            candidates.length ||
          !paging ||
          paging.limit !== 100 ||
          paging.has_more !== (paging.next_cursor !== null) ||
          (paging.next_cursor !== null &&
            (!paging.next_cursor || paging.next_cursor === input.cursor))
        )
          throw new Error("Invalid reference paging.");
        return input.signal.aborted
          ? { kind: "aborted" }
          : {
              kind: "accepted",
              value: {
                candidates,
                ...(canonicalQuery ? { canonicalQuery } : {}),
                hasMore: paging.has_more,
                nextCursor: paging.next_cursor,
              },
            };
      } catch {
        return input.signal.aborted
          ? { kind: "aborted" }
          : {
              kind: "rejected",
              failure: {
                kind: responseAccepted ? "invalid_contract" : "retryable",
                message: responseAccepted
                  ? "References could not be verified. Restart from First."
                  : "References could not be loaded. Retry this read.",
              },
            };
      }
    },
  };
}
