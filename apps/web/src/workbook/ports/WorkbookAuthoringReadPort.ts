import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import type { WorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookPortResult } from "./WorkbookPortResult";

export type WorkbookAuthoringContext = Readonly<{
  source: { readonly viewSchemaId: string };
  target: ViewContract;
  feature: InspectorFeatureGroup;
}>;
export type WorkbookAuthoringCandidate = Readonly<{
  recordId: string;
  displayText: string;
  viewSchemaId: string;
  row?: WorkbookQueryRow;
}>;
export type WorkbookAuthoringPage = Readonly<{
  candidates: readonly WorkbookAuthoringCandidate[];
  hasMore: boolean;
  nextCursor: string | null;
}>;
export type WorkbookAuthoringQuery = Readonly<{
  viewSchemaId: string;
  cursor: string | null;
  queryState: WorkbookQueryState;
  signal: AbortSignal;
}>;
export interface WorkbookAuthoringReadPort {
  page(
    input: WorkbookAuthoringQuery,
  ): Promise<WorkbookPortResult<WorkbookAuthoringPage>>;
  availableViews(signal: AbortSignal): Promise<readonly string[]>;
  verify(context: WorkbookAuthoringContext, signal: AbortSignal): Promise<void>;
}
export type WorkbookAuthoringAuthorityReader = (
  baseline: WorkbookMutationAuthority,
  signal: AbortSignal,
) => Promise<WorkbookMutationAuthority>;
