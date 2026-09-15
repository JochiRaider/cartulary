import type { ReferenceIdentityKind } from "@cartulary/view-contracts";
import type { WorkbookQueryState } from "../models/workbookQuery";
import type {
  WorkbookCanonicalQuery,
  WorkbookQueryPaging,
} from "../query/WorkbookViewQueryPort";
import type { WorkbookPortResult } from "./WorkbookPortResult";

/** Identity and presentation are separate; neither establishes current admission. */
export type WorkbookReference = Readonly<{
  identity: Readonly<{ kind: ReferenceIdentityKind; id: string }>;
  viewSchemaId: string;
  displayText: string;
  presentation: "observed" | "retained" | "unresolved";
}>;
export type WorkbookReferenceRequest = Readonly<{
  identityKind: ReferenceIdentityKind;
  viewSchemaId: string;
  queryState: WorkbookQueryState;
  cursorToken?: string;
  expectedCanonicalQuery?: WorkbookCanonicalQuery;
}>;
export type WorkbookReferencePage = Readonly<{
  candidates: readonly WorkbookReference[];
  canonicalQuery: WorkbookCanonicalQuery | null;
  paging: WorkbookQueryPaging;
  producingRequest: WorkbookReferenceRequest;
}>;
export interface WorkbookReferenceReadPort {
  page(
    input: WorkbookReferenceRequest,
    signal: AbortSignal,
  ): Promise<WorkbookPortResult<WorkbookReferencePage>>;
}
export type WorkbookReferenceMemberPage = Readonly<{
  members: readonly { readonly userId: string; readonly displayName: string }[];
  paging: WorkbookQueryPaging;
}>;

export function workbookReferenceKey(reference: WorkbookReference): string {
  return `${reference.identity.kind}:${reference.identity.id}`;
}
