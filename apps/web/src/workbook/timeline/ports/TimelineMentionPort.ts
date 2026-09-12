import type {
  MentionCreateReview,
  MentionCreationAttempt,
  MentionCreationOutcome,
} from "../actions/timelineMentionCreationModel";
import type {
  MentionAttempt,
  MentionOutcome,
  MentionReview,
} from "../actions/timelineMentionOperationModel";

export interface TimelineMentionResolutionPort {
  capture(review: MentionReview, id: string): MentionAttempt;
  send(attempt: MentionAttempt, signal: AbortSignal): Promise<MentionOutcome>;
}
export interface TimelineMentionEntityCreationPort {
  capture(review: MentionCreateReview, id: string): MentionCreationAttempt;
  send(
    attempt: MentionCreationAttempt,
    signal: AbortSignal,
  ): Promise<MentionCreationOutcome>;
}
