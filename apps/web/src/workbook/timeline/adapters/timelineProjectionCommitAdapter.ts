import { flushSync } from "react-dom";

/**
 * Single Timeline boundary for commits that must be visible before a caller
 * restores focus or resolves a continuity obligation.
 */
export function commitTimelineProjection(
  commit: () => void,
  synchronous: boolean,
) {
  if (!synchronous) {
    commit();
    return;
  }
  flushSync(commit);
}
