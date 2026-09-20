import { useLayoutEffect, useRef } from "react";
import type { TimelinePresentationPorts } from "../mutations/WorkbookTimelineMutationOwner";
import { timelineMutationOwnerFor } from "../mutations/WorkbookTimelineMutationOwner";

/** A Timeline surface lends presentation effects to its retained mutation owner. */
export function useTimelineMutationDriver(ports: TimelinePresentationPorts) {
  const owner = timelineMutationOwnerFor(ports.mutationRuntime);
  const current = useRef(ports);
  current.current = ports;
  useLayoutEffect(() => owner.attach(() => current.current), [owner]);
  return {
    discardBlockedEdit: owner.discardBlockedEdit,
    enqueuePendingReplayUnit: owner.enqueuePendingReplayUnit,
    retryBlockedEdit: owner.retryBlockedEdit,
  };
}
