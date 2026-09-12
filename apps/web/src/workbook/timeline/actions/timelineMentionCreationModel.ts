import { requireViewContract } from "@cartulary/view-contracts";
import type { WorkbookProtocolCreateViewRowReceipt } from "../../adapters/workbookProtocolTypes";
import { buildGenericCreateRequest } from "../../features/generic/genericCreateRequestBuilder";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type {
  MentionAuthority,
  MentionSubject,
} from "./timelineMentionOperationModel";
export type MentionCreateReview = Readonly<{
  subject: MentionSubject;
  authority: MentionAuthority;
  draft: Readonly<Record<string, string>>;
}>;
export type MentionCreationReceipt =
  Readonly<WorkbookProtocolCreateViewRowReceipt>;
export type MentionCreationAttempt = Readonly<{
  id: string;
  review: MentionCreateReview;
  operationID: "createViewRow";
  method: "POST";
  path: string;
  apiBase: string | undefined;
  viewSchemaId: string;
  body: string;
}>;
export type MentionCreationOutcome =
  | Readonly<{ kind: "accepted"; receipt: MentionCreationReceipt }>
  | Readonly<{ kind: "uncertain" }>
  | Readonly<{ kind: "rejected"; failure: WorkbookOperationFailure }>;
export type MentionCreationOperation = Readonly<{
  key: number;
  attempt: MentionCreationAttempt;
  phase:
    | "preparing"
    | "preparation_failed"
    | "submitting"
    | "uncertain"
    | "rejected"
    | "accepted";
  receipt: MentionCreationReceipt | null;
  failure: WorkbookOperationFailure | null;
  transportPending: boolean;
  linkKey: number | null;
  refresh: "pending" | "refreshing" | "required" | "complete";
}>;
export function mentionEntityContract(type: MentionSubject["entityType"]) {
  return requireViewContract(
    type === "host" ? hostsViewSchemaId : identitiesViewSchemaId,
  );
}
export function initialMentionCreateDraft(
  subject: MentionSubject,
): Record<string, string> {
  const contract = mentionEntityContract(subject.entityType);
  const draft = Object.fromEntries(
    contract.fields
      .filter((field) => field.createWritable)
      .map((field) => [field.fieldKey, ""]),
  );
  draft[`${subject.entityType}.display_name`] = subject.rawText;
  return draft;
}
export function mentionCreateRequest(review: MentionCreateReview, id: string) {
  return review.subject.state === "unresolved"
    ? buildGenericCreateRequest(
        mentionEntityContract(review.subject.entityType),
        review.draft,
        id,
      )
    : null;
}
