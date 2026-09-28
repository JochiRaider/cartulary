import {
  boundedRead,
  ObservationStopped,
} from "../../../services/asyncObservation";
import { evidenceAccessFailureIsDefinitiveBlocker } from "../../evidence/evidenceAccessPresentation";
import type {
  EvidenceCapabilityPort,
  EvidenceHandleOutcome,
} from "../../mutations/workbookMutationCommandPorts";

type PreviewProbeResult = "previewable" | "blocked" | "indeterminate";

function classifyPreviewProbe(
  outcome: EvidenceHandleOutcome,
): PreviewProbeResult {
  if (outcome.kind === "accepted") return "previewable";
  return evidenceAccessFailureIsDefinitiveBlocker(outcome.failure)
    ? "blocked"
    : "indeterminate";
}

export type SolePreviewResult =
  | { readonly kind: "sole"; readonly recordId: string }
  | { readonly kind: "none" | "multiple" | "indeterminate" };

/** Only an explicit Space intent calls this bounded, one-at-a-time probe. */
export async function resolveSolePreviewableEvidence(
  recordIds: readonly string[],
  port: EvidenceCapabilityPort,
  signal: AbortSignal,
  isCurrent: () => boolean,
): Promise<SolePreviewResult> {
  try {
    return await boundedRead(
      async (boundedSignal): Promise<SolePreviewResult> => {
        let sole: string | null = null;
        for (const recordId of recordIds) {
          if (boundedSignal.aborted || !isCurrent())
            return { kind: "indeterminate" };
          const outcome = await port.issueHandle({
            evidenceRecordId: recordId,
            kind: "preview",
            signal: boundedSignal,
          });
          if (boundedSignal.aborted || !isCurrent())
            return { kind: "indeterminate" };
          const result = classifyPreviewProbe(outcome);
          if (result === "indeterminate") return { kind: "indeterminate" };
          if (result === "previewable") {
            if (sole !== null) return { kind: "multiple" };
            sole = recordId;
          }
        }
        return sole === null
          ? { kind: "none" }
          : { kind: "sole", recordId: sole };
      },
      signal,
    );
  } catch (error) {
    if (error instanceof ObservationStopped) return { kind: "indeterminate" };
    return { kind: "indeterminate" };
  }
}
