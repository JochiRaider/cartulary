import {
  getViewContract,
  type InspectorFeatureGroup,
  notesViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookProtocolCreateLinkedNoteRequest } from "../../adapters/workbookProtocolTypes";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";
import { freezeWorkbookValue } from "../../utils/freezeWorkbookValue";

export const noteCreateFeature = "create_related.note";
export const noteCreateView = notesViewSchemaId;
export const noteSourceViews: readonly [string, string, string, string] = [
  "cartulary.view.timeline.v2",
  "cartulary.view.hosts.v1",
  "cartulary.view.identities.v1",
  "cartulary.view.evidence.v1",
];
export type NoteSource = Readonly<{
  recordId: string;
  viewSchemaId: string;
  rowVersion: number;
  label: string;
}>;
export type NoteDraft = Readonly<{
  id: number;
  revision: number;
  actorId: string;
  incidentId: string;
  values: Readonly<Record<string, string>>;
  source: NoteSource | null;
  origin: Readonly<{
    sheetRef: SheetRef;
    subject: WorkbookInspectorLiveRowBinding["subject"] | null;
    feature: InspectorFeatureGroup | null;
  }>;
}>;
export type NoteCreateReader = Pick<
  WorkbookAuthoringReadPort,
  "availableViews" | "page"
> & {
  verifyNote(draft: NoteDraft, signal: AbortSignal): Promise<void>;
};
export function noteFeature(view: string): InspectorFeatureGroup | null {
  if (!noteSourceViews.includes(view)) return null;
  const feature = getViewContract(view)?.inspectorConfig.featureGroups.find(
    (f) => f.featureGroupKey === noteCreateFeature,
  );
  return feature?.routeBinding.kind === "record_action" &&
    feature.routeBinding.owner === "record_linked_note_create_route" &&
    feature.routeBinding.actionKey === noteCreateFeature
    ? feature
    : null;
}
export function noteSourceFromSubject(
  subject: WorkbookInspectorLiveRowBinding,
): NoteSource {
  return freezeWorkbookValue({
    recordId: subject.subject.recordId,
    viewSchemaId: subject.subject.viewSchemaId,
    rowVersion: subject.subject.rowVersion,
    label: subject.subject.label,
  });
}
export function prepareNote(draft: NoteDraft, clientTxnId: string) {
  const errors: Record<string, string> = {};
  const request: WorkbookProtocolCreateLinkedNoteRequest = {
    client_txn_id: clientTxnId,
  };
  const normalized: Record<string, string> = {};
  for (const key of ["note.title", "note.body"] as const) {
    if (!Object.hasOwn(draft.values, key)) continue;
    const raw = draft.values[key] ?? "";
    const body = key === "note.body";
    const value = (body ? raw.replace(/\r\n?/gu, "\n") : raw)
      .normalize("NFC")
      .replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");
    // Validate controls before trimming: invalid raw input must not disappear.
    if (invalidControls(body ? raw.replace(/\r\n?/gu, "\n") : raw, body))
      errors[key] = "Remove unsupported control characters.";
    else if ([...value].length > (body ? 16384 : 512))
      errors[key] = `Use at most ${body ? 16384 : 512} characters.`;
    normalized[key] = value;
    request[key] = value;
  }
  if (!normalized["note.title"] && !normalized["note.body"])
    errors["note.title"] ??= "Enter a title or body before creating a Note.";
  const rawTags = (draft.values["note.tags"] ?? "").split(/\r\n?|\n/u);
  const tags = rawTags
    .map((s) =>
      s.normalize("NFC").replace(/^\p{White_Space}+|\p{White_Space}+$/gu, ""),
    )
    .filter(Boolean);
  if (rawTags.some((tag) => invalidControls(tag, false)))
    errors["note.tags"] = "Remove unsupported control characters from tags.";
  if (tags.some((tag) => [...tag].length > 64))
    errors["note.tags"] = "Use at most 64 characters per tag.";
  if (tags.length > 64) errors["note.tags"] = "Use at most 64 tag additions.";
  if (tags.length)
    request["note.tags"] = {
      kind: "collection_actions_v1",
      actions: tags.map((tag_name) => ({
        op: "add_tag" as const,
        tag_name,
      })) as [
        { op: "add_tag"; tag_name: string },
        ...{ op: "add_tag"; tag_name: string }[],
      ],
    };
  const contract = requireViewContract(noteCreateView);
  if (
    !contract.createCapable ||
    contract.createInputs.length ||
    Object.keys(draft.values).some(
      (key) => !contract.fieldMap[key]?.createWritable,
    )
  )
    errors["note.title"] =
      "Note creation capability changed. Review the draft.";
  return {
    errors,
    request: Object.keys(errors).length ? null : freezeWorkbookValue(request),
  };
}

function invalidControls(value: string, body: boolean) {
  return [...value].some((character) => {
    const code = character.codePointAt(0) ?? 0;
    return (
      (code < 32 && !(body && (code === 9 || code === 10))) ||
      (code >= 127 && code <= 159)
    );
  });
}
