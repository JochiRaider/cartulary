/** Owner-issued work projection. File bytes, upload capabilities and requests stay private. */
export type EvidenceWorkAttention = Readonly<{
  workId: string;
  category:
    | "draft"
    | "review"
    | "failure"
    | "in_progress"
    | "uncertain"
    | "refresh";
  label: string;
  outcomeIdentity: string;
}>;
