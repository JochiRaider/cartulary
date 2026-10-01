import type { ViewRow } from "../../../../packages/protocol-ts/src/generated/core-http-types.js";

export type SourceRow = { key: string; fields: Record<string, string | null> };
export type EvidenceRecipe = { key: string; title: string; filename: string; content: string };
export type TimelineRecipe = {
  schema_id: string;
  dataset_id: string;
  content_revision: number;
  view_schema_id: string;
  rows: SourceRow[];
  entities: { key: string; view_schema_id: string; fields: Record<string, string>; aliases?: string[] }[];
  evidence: EvidenceRecipe[];
  relationships: { row: string; field: string; kind: string; text: string; target?: string }[];
  operations: ({ kind: "tags"; row: string; values: string[] } | { kind: "patch"; row: string; fields: Record<string, string | null> } | { kind: "evidence"; row: string; evidence: string } | { kind: "relationship"; relationship: number } | { kind: "review"; row: string } | { kind: "supersede"; row: string; replacement: string; reason: string })[];
  continuation: { start_utc: string; interval_minutes: number; local_offset_minutes: number; maximum_rows: number; source_fields: Record<string, string>; synopsis_template: string; raw_template: string };
};
export type TimelineExpectations = {
  schema_id: string; dataset_id: string; content_revision: number;
  fields: { source: string[]; collections: string[]; derived: string[] };
  rows: Record<string, { capture_state: string; evidence_count: number; has_unresolved_mentions: boolean; replacement: string | null }>;
};
export type InvestigationOperation = "queryWorkbookView" | "patchRecord" | "getTimelineTimeConversionProfile" | "createViewRow" | "resolveEntityMention" | "markTimelineRecordReviewed" | "supersedeRecord";
export type InvestigationOutcomes = { dataset_id: string; content_revision: number; core_rows: number; total_rows: number; semantic_sha256: string };
export const timelineRecipe: TimelineRecipe;
export const timelineExpectations: TimelineExpectations;
export function validateTimelineRecipe(recipe?: unknown, expectations?: unknown): TimelineRecipe;
export function timelineRows(recipe: TimelineRecipe, continuationCount: number): SourceRow[];
export function seedTimelineInvestigation<T>(input: {
  incident: T; continuationCount: number; recipe?: unknown;
  port: {
    createIncident(incident: T): Promise<string>;
    call(operation: InvestigationOperation, body: unknown, pathParameters: Record<string, string>): Promise<unknown>;
    uploadEvidence(incidentId: string, evidence: EvidenceRecipe): Promise<ViewRow>;
  };
}): Promise<{
  incidentId: string; mapping: Map<string, string>; recordId(key: string): string; rows: ViewRow[];
  dismissed: { row: string; field: string; text: string; id: string }[];
  outcomes: InvestigationOutcomes; readRows(): Promise<ViewRow[]>; verify(): Promise<InvestigationOutcomes>;
}>;
export function verifyTimelineInvestigation(input: { recipe?: TimelineRecipe; authored: SourceRow[]; rows: ViewRow[]; mapping: Map<string, string> }): InvestigationOutcomes;
