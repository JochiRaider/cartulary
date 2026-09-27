import { evidenceAccessFailureIsDefinitiveBlocker } from "../../evidence/evidenceAccessPresentation";
import type {
  EvidenceCapabilityPort,
  EvidenceHandleOutcome,
} from "../../mutations/workbookMutationCommandPorts";

export type PreviewProbeResult = "previewable" | "blocked" | "indeterminate";

export function classifyPreviewProbe(
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
  let sole: string | null = null;
  for (const recordId of recordIds) {
    if (signal.aborted || !isCurrent()) return { kind: "indeterminate" };
    let outcome: EvidenceHandleOutcome;
    try {
      outcome = await port.issueHandle({
        evidenceRecordId: recordId,
        kind: "preview",
        signal,
      });
    } catch {
      return { kind: "indeterminate" };
    }
    if (signal.aborted || !isCurrent()) return { kind: "indeterminate" };
    const result = classifyPreviewProbe(outcome);
    if (result === "indeterminate") return { kind: "indeterminate" };
    if (result === "previewable") {
      if (sole !== null) return { kind: "multiple" };
      sole = recordId;
    }
  }
  return sole === null ? { kind: "none" } : { kind: "sole", recordId: sole };
}
