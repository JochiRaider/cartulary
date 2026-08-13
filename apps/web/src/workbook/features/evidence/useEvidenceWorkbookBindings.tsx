import {
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { evidenceAccessMessageLiveRegion } from "../../evidence/evidenceAccessPresentation";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import { workbookSurfaceOverlayPanelStyle } from "../../layout/WorkbookSurfaceLayout";
import { buildEvidenceLifecycleViewModel } from "../../models/evidenceLifecycleViewModel";
import type { EvidenceCapabilityPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { stringifyGridValue } from "../../utils/workbookValueFormat";

type EvidencePreviewState = {
  href: string;
  recordId: string;
  title: string;
  previewKind: string | null;
};

type MutationPorts = Pick<
  GenericSurfaceMutationController,
  "beginMutation" | "markMutationConflict" | "markMutationSaved"
>;

export function useEvidenceWorkbookBindings({
  mutationCommands,
  mutation,
  onRefresh,
  ownerBindings,
  resetKey,
}: {
  readonly mutationCommands: EvidenceCapabilityPort;
  readonly mutation: MutationPorts;
  readonly onRefresh: () => Promise<void> | void;
  readonly ownerBindings: readonly WorkbookOwnerBinding[];
  readonly resetKey: string;
}) {
  const active = ownerBindings.includes("evidence_lifecycle");
  const [messageByRecordId, setMessageByRecordId] = useState<
    Record<string, string>
  >({});
  const [preview, setPreview] = useState<EvidencePreviewState | null>(null);
  const generationRef = useRef(0);

  useEffect(() => {
    void resetKey;
    generationRef.current += 1;
    setMessageByRecordId({});
    setPreview(null);
  }, [resetKey]);

  useEffect(
    () => () => {
      generationRef.current += 1;
    },
    [],
  );

  const setMessage = useCallback((recordId: string, message: string | null) => {
    setMessageByRecordId((current) => {
      const next = { ...current };
      if (message === null) {
        delete next[recordId];
      } else {
        next[recordId] = message;
      }
      return next;
    });
  }, []);

  const issueHandle = useCallback(
    async (row: WorkbookQueryRow, kind: "preview" | "download") => {
      const generation = generationRef.current;
      setMessage(row.record_id, null);
      const handle = await mutationCommands.issueHandle({
        evidenceRecordId: row.record_id,
        kind,
      });
      if (generationRef.current !== generation) return;
      if (handle.kind === "rejected") {
        setMessage(row.record_id, handle.failure.message);
        return;
      }
      if (kind === "preview") {
        setPreview({
          href: handle.value.href,
          recordId: row.record_id,
          title:
            stringifyGridValue(row.cells["evidence.title"]?.value).trim() ||
            row.record_id,
          previewKind: handle.value.previewKind,
        });
        setMessage(row.record_id, "Preview loaded inline.");
        return;
      }
      const anchor = document.createElement("a");
      anchor.href = handle.value.href;
      anchor.download = handle.value.filename || "evidence";
      anchor.rel = "noopener";
      document.body.append(anchor);
      anchor.click();
      anchor.remove();
      setMessage(row.record_id, "Download handle issued.");
    },
    [mutationCommands, setMessage],
  );

  const attachFile = useCallback(
    async (row: WorkbookQueryRow, file: File) => {
      const generation = generationRef.current;
      if (file.size <= 0) {
        setMessage(row.record_id, "Evidence attach failed.");
        return;
      }
      setMessage(row.record_id, "Uploading evidence.");
      mutation.beginMutation();
      const outcome = await mutationCommands.attach({
        baseRowVersion: row.row_version,
        evidenceRecordId: row.record_id,
        file,
      });
      if (generationRef.current !== generation) return;
      if (outcome.kind === "rejected") {
        setMessage(row.record_id, outcome.failure.message);
        mutation.markMutationConflict();
        return;
      }
      setMessage(row.record_id, "Evidence attached.");
      await onRefresh();
      if (generationRef.current !== generation) return;
      mutation.markMutationSaved();
    },
    [mutation, mutationCommands, onRefresh, setMessage],
  );

  const renderRecordActions = useCallback(
    (row: WorkbookQueryRow) => {
      if (!active) {
        return null;
      }
      const access = buildEvidenceLifecycleViewModel({
        evidenceLifecycleState: row.cells["evidence.lifecycle_state"]?.value,
        objectBlobUploadState: row.cells["evidence.upload_state"]?.value,
      });
      const message = messageByRecordId[row.record_id] ?? access.message;
      const liveRegion =
        message === null
          ? null
          : evidenceAccessMessageLiveRegion(message, access);
      return (
        <div data-evidence-state-key={access.stateKey} style={actionStackStyle}>
          <div style={inlineButtonRowStyle}>
            <button
              data-testid={evidencePreviewButtonTestId(row.record_id)}
              disabled={!access.canPreview}
              style={actionButtonStyle}
              type="button"
              onClick={() => void issueHandle(row, "preview")}
            >
              Preview
            </button>
            <button
              data-testid={evidenceDownloadButtonTestId(row.record_id)}
              disabled={!access.canDownload}
              style={actionButtonStyle}
              type="button"
              onClick={() => void issueHandle(row, "download")}
            >
              Download
            </button>
          </div>
          <label style={labelStyle}>
            Attach file
            <input
              data-testid={evidenceAttachFileInputTestId(row.record_id)}
              style={inputStyle}
              type="file"
              accept="image/*,.txt,.pdf,text/plain,application/pdf"
              onChange={(event) => {
                const [file] = Array.from(event.currentTarget.files ?? []);
                event.currentTarget.value = "";
                if (file) {
                  void attachFile(row, file);
                }
              }}
            />
          </label>
          {message ? (
            <span
              aria-live={liveRegion?.ariaLive}
              data-testid={evidenceAccessMessageTestId(row.record_id)}
              role={liveRegion?.role}
              style={messageStyle}
            >
              {message}
            </span>
          ) : null}
        </div>
      );
    },
    [active, attachFile, issueHandle, messageByRecordId],
  );

  const overlay =
    active && preview ? (
      <section
        data-testid={evidencePreviewPanelTestId()}
        style={previewPanelStyle}
      >
        <div style={previewHeaderStyle}>
          <div>
            <p style={eyebrowStyle}>Preview</p>
            <h2 style={sectionTitleStyle}>{preview.title}</h2>
          </div>
          <button
            style={secondaryActionButtonStyle}
            type="button"
            onClick={() => setPreview(null)}
          >
            Close
          </button>
        </div>
        <iframe
          data-testid={evidencePreviewFrameTestId(preview.recordId)}
          src={preview.href}
          style={previewFrameStyle}
          title={`Evidence preview ${preview.title}`}
        />
        {preview.previewKind ? (
          <p style={messageStyle}>{preview.previewKind}</p>
        ) : null}
      </section>
    ) : null;

  return {
    actionsWidth: active ? 208 : 76,
    hasRecordActions: active,
    overlay,
    renderRecordActions,
  };
}

const actionStackStyle = { display: "grid", gap: "0.5rem" };
const inlineButtonRowStyle = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap" as const,
};
const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};
const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};
const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};
const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
};
const messageStyle = {
  margin: 0,
  fontSize: "0.85rem",
  color: "var(--ct-colors-ink-muted)",
};
const previewPanelStyle = {
  ...workbookSurfaceOverlayPanelStyle,
  display: "grid",
  gap: "0.75rem",
  padding: "1rem",
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
} satisfies CSSProperties;
const previewHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "start",
};
const previewFrameStyle = {
  width: "100%",
  blockSize: "min(28rem, 34vh)",
  minHeight: "12rem",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-2)",
};
const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};
const sectionTitleStyle = { margin: 0, fontSize: "1rem" };
