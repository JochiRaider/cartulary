import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type {
  WorkbookCandidatePage,
  WorkbookCandidateQuery,
} from "./WorkbookCandidateReadPort";
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
/** Retained authoring selection; full rows remain confined to source verification. */
export type WorkbookAuthoringSelection = Readonly<{
  recordId: string;
  displayText: string;
  viewSchemaId: string;
  rowVersion?: number;
}>;
export type WorkbookAuthoringPage =
  WorkbookCandidatePage<WorkbookAuthoringCandidate>;
export type WorkbookAuthoringQuery = WorkbookCandidateQuery &
  Readonly<{ viewSchemaId: string }>;
export interface WorkbookAuthoringReadPort {
  page(
    input: WorkbookAuthoringQuery,
  ): Promise<WorkbookPortResult<WorkbookAuthoringPage>>;
  availableViews(
    signal: AbortSignal,
  ): Promise<WorkbookPortResult<readonly string[]>>;
  verify(context: WorkbookAuthoringContext, signal: AbortSignal): Promise<void>;
}
export type WorkbookAuthoringAuthorityReader = (
  baseline: WorkbookMutationAuthority,
  signal: AbortSignal,
) => Promise<WorkbookMutationAuthority>;
