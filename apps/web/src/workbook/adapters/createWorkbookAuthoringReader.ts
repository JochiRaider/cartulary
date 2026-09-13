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
      if (result.kind !== "accepted")
        throw new Error("Reference surfaces are unavailable.");
      return result.value.data.view_schemas
        .filter((view) => getViewContract(view.view_schema_id))
        .map((view) => view.view_schema_id);
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
        !schema.create_capable ||
        schema.view_schema_id !== draft.target.viewSchemaId ||
        schema.inline_create.permits_zero_field_create !==
          draft.target.permitsZeroFieldCreate ||
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
          ) ||
        JSON.stringify(schema.inline_create.minimum_create_field_sets) !==
          JSON.stringify(draft.target.minimumCreateFieldSets) ||
        schema.fields.length !== draft.target.fields.length ||
        schema.fields.some(
          (field) =>
            field.create_writable !==
              draft.target.fieldMap[field.field_key]?.createWritable ||
            field.read_kind !==
              draft.target.fieldMap[field.field_key]?.readKind ||
            field.write_kind !==
              draft.target.fieldMap[field.field_key]?.writeKind ||
            field.clearable !==
              draft.target.fieldMap[field.field_key]?.clearable ||
            field.string_contract_id !==
              draft.target.fieldMap[field.field_key]?.stringContractId ||
            field.direct_scalar_contract_id !==
              draft.target.fieldMap[field.field_key]?.directScalarContractId ||
            JSON.stringify(field.enum_values) !==
              JSON.stringify(
                draft.target.fieldMap[field.field_key]?.enumValues,
              ) ||
            field.direct_reference_contract_id !==
              draft.target.fieldMap[field.field_key]?.directReferenceContractId,
        ) ||
        JSON.stringify(
          schema.create_inputs.map((input) => ({
            inputKey: input.input_key,
            nullable: input.nullable,
            required: input.required,
            valueContractId: input.value_contract_id,
          })),
        ) !== JSON.stringify(draft.target.createInputs)
      )
        throw new Error(
          "Creation capability changed. Review the retained draft.",
        );
    },
    async page(input) {
      try {
        let candidates: WorkbookAuthoringCandidate[];
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
          if (result.kind !== "accepted") {
            options.recheckAuthority();
            return result;
          }
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
          if (result.kind !== "accepted") {
            options.recheckAuthority();
            return result;
          }
          if (
            result.value.data.incident_id !== options.incidentId ||
            result.value.data.view_schema_id !== input.viewSchemaId
          )
            throw new Error("Invalid candidate scope.");
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
                kind: "retryable",
                message: "References could not be verified. Retry the read.",
              },
            };
      }
    },
  };
}
