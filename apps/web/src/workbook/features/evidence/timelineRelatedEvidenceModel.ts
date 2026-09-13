import {
  evidenceViewSchemaId,
  getViewContract,
  type InspectorFeatureGroup,
  timelineViewSchemaId,
  type ViewContract,
} from "@cartulary/view-contracts";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookProtocolCreateViewRowRequest } from "../../adapters/workbookProtocolTypes";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import { workbookCreationAvailable } from "../../models/genericWorkbookModel";
import { decodeCreateViewRowRequest } from "../../models/workbookRequestDecoders";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import { freezeWorkbookValue } from "../../utils/freezeWorkbookValue";

export const relatedEvidenceFeature = "create_related.evidence";
export const metadataEvidenceStates = [
  "requested",
  "pending_receipt",
  "received",
  "quarantined",
] as const;
export type TimelineRelatedEvidenceDraft = Readonly<{
  id: number;
  revision: number;
  actorId: string;
  incidentId: string;
  source: {
    readonly recordId: string;
    readonly viewSchemaId: string;
    readonly rowVersion: number;
  };
  presentation: {
    readonly sheetRef: SheetRef;
    readonly label: string;
    readonly surfaceLabel: string;
  };
  feature: InspectorFeatureGroup;
  target: ViewContract;
  values: Readonly<Record<string, string>>;
  labels: Readonly<Record<string, string>>;
}>;

export function buildTimelineRelatedEvidenceDraft(
  id: number,
  authority: WorkbookMutationAuthority,
  subject: WorkbookInspectorLiveRowBinding,
  feature: InspectorFeatureGroup,
  sheetRef: SheetRef,
): TimelineRelatedEvidenceDraft | null {
  const source = getViewContract(timelineViewSchemaId),
    target = getViewContract(evidenceViewSchemaId);
  const declared = source?.inspectorConfig.featureGroups.find(
    (item) => item.featureGroupKey === relatedEvidenceFeature,
  );
  if (
    subject.subject.viewSchemaId !== timelineViewSchemaId ||
    !target ||
    !declared ||
    JSON.stringify(feature) !== JSON.stringify(declared) ||
    !workbookCreationAvailable(target) ||
    declared.routeBinding.targetViewSchemaId !== evidenceViewSchemaId ||
    declared.seedBindings.length
  )
    return null;
  return freezeWorkbookValue({
    id,
    revision: 0,
    actorId: authority.actorId,
    incidentId: authority.incidentId,
    source: {
      recordId: subject.subject.recordId,
      viewSchemaId: timelineViewSchemaId,
      rowVersion: subject.subject.rowVersion,
    },
    presentation: {
      sheetRef: structuredClone(sheetRef),
      label: subject.subject.label,
      surfaceLabel: subject.subject.surfaceLabel,
    },
    feature: declared,
    target,
    values: {},
    labels: {},
  });
}

const stableId =
  /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/u;
const reservedObject =
  /^object:\/\/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u;

/** Validate calendar components before Date.parse can normalize impossible dates. */
function timestamp(raw: string): string | null {
  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(\.\d+)?(Z|[+-]\d{2}:\d{2})$/u.exec(
      raw,
    );
  if (!match) return null;
  const year = Number(match[1]),
    month = Number(match[2]),
    day = Number(match[3]);
  const leap = year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  const days = [31, leap ? 29 : 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];
  if (
    month < 1 ||
    month > 12 ||
    day < 1 ||
    day > (days[month - 1] ?? 0) ||
    Number(match[4]) > 23 ||
    Number(match[5]) > 59 ||
    Number(match[6]) > 59
  )
    return null;
  const zone = match[8] ?? "";
  if (
    zone !== "Z" &&
    (Number(zone.slice(1, 3)) > 23 || Number(zone.slice(4)) > 59)
  )
    return null;
  const instant = Date.parse(raw);
  if (!Number.isFinite(instant)) return null;
  // Preserve supplied fractional precision while converting the whole second to UTC.
  return new Date(instant)
    .toISOString()
    .replace(/\.\d{3}Z$/u, `${match[7] ?? ""}Z`);
}

export function prepareTimelineRelatedEvidence(
  draft: TimelineRelatedEvidenceDraft,
) {
  const errors: Record<string, string> = {},
    values: Record<string, unknown> = {};
  let signal = false;
  for (const field of draft.target.fields.filter(
    (item) => item.createWritable,
  )) {
    const raw = draft.values[field.fieldKey] ?? "";
    if (raw === "") continue;
    if (field.directReferenceContractId === "same_incident_party_ref_v1") {
      if (!stableId.test(raw))
        errors[field.fieldKey] = "Choose an available Party reference.";
      else values[field.fieldKey] = raw;
      continue;
    }
    const normalized = raw
      .normalize("NFC")
      .replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");
    if (!normalized) continue;
    if (field.fieldKey === "evidence.lifecycle_state") {
      if (!metadataEvidenceStates.some((state) => state === normalized))
        errors[field.fieldKey] =
          "Choose Requested, Pending receipt, Received, or Quarantined for metadata creation.";
      else values[field.fieldKey] = normalized;
    } else if (field.directScalarContractId === "timestamp_instant_v1") {
      const value = timestamp(normalized);
      if (value === null)
        errors[field.fieldKey] =
          "Enter a valid timestamp with a timezone, for example 2026-09-12T14:30:00Z.";
      else values[field.fieldKey] = value;
    } else if (
      field.fieldKey === "evidence.storage_ref" &&
      reservedObject.test(normalized)
    ) {
      errors[field.fieldKey] =
        "Use an external storage reference. Object references are assigned by the server.";
    } else {
      const limit =
        field.stringContractId === "single_line_title_v1"
          ? 512
          : field.stringContractId === "party_text_v1"
            ? 256
            : field.stringContractId === "locator_text_v1"
              ? 1024
              : 0;
      if (
        !limit ||
        [...normalized].length > limit ||
        /[\p{Cc}\p{Cs}]/u.test(normalized)
      )
        errors[field.fieldKey] =
          `Use at most ${limit} characters without control characters for ${field.label.toLowerCase()}.`;
      else values[field.fieldKey] = normalized;
    }
    if (values[field.fieldKey] !== undefined) signal = true;
  }
  if (!signal)
    errors.payload =
      "Enter Evidence metadata or explicitly choose a lifecycle. Party references and Timeline context alone do not create Evidence.";
  return { errors, values };
}

export function timelineRelatedEvidenceRequest(
  draft: TimelineRelatedEvidenceDraft,
  clientTxnId: string,
): WorkbookProtocolCreateViewRowRequest | null {
  const prepared = prepareTimelineRelatedEvidence(draft);
  if (Object.keys(prepared.errors).length) return null;
  return decodeCreateViewRowRequest(draft.target, {
    ...prepared.values,
    client_txn_id: clientTxnId,
  });
}
