import { useEffect, useRef, useState } from "react";
import type {
  EvidenceMetadata,
  WorkbookEvidenceMetadataPort,
} from "../../evidence/WorkbookEvidenceMetadataPort";
import type { WorkbookReadScope } from "../../query/WorkbookQueryRow";
import { workbookRowIsAdmissible } from "../../query/workbookRowObservation";
import {
  readTimelineEvidenceLinks,
  timelineEvidenceLinkKey,
} from "../models/timelineEvidenceLinks";
import type { WorkbookRow } from "../models/timelineRowModel";

const windowSize = 5;
type Observation = {
  readonly key: string;
  readonly generation: number;
  readonly port: WorkbookEvidenceMetadataPort;
  readonly kind: "ready" | "failed";
  readonly items: readonly EvidenceMetadata[];
};

/** Presentation-scoped metadata: never survives a source, window or authority change. */
export function useTimelineEvidenceMetadata(input: {
  readonly port: WorkbookEvidenceMetadataPort;
  readonly row: WorkbookRow | null;
  readonly readScope: WorkbookReadScope | null;
  readonly canRead: boolean;
  readonly inspectorOpen: boolean;
  readonly scopeKey: string;
  readonly onAccessFailure: () => Promise<void> | void;
}) {
  const inputRef = useRef(input);
  inputRef.current = input;
  const sourceKey = input.row ? timelineEvidenceLinkKey(input.row) : null;
  const source = JSON.stringify([input.scopeKey, input.readScope, sourceKey]);
  const links = input.row ? readTimelineEvidenceLinks(input.row) : null;
  const [page, setPage] = useState({ source, index: 0 });
  const [attempt, setAttempt] = useState(0);
  const [observation, setObservation] = useState<Observation | null>(null);
  const total = links?.kind === "available" ? links.items.length : 0;
  const index =
    page.source === source
      ? Math.min(page.index, Math.max(0, Math.ceil(total / windowSize) - 1))
      : 0;
  const window =
    links?.kind === "available"
      ? links.items.slice(index * windowSize, (index + 1) * windowSize)
      : [];
  const enabled =
    input.canRead &&
    input.inspectorOpen &&
    !!input.row?.rawRow &&
    !!input.readScope &&
    workbookRowIsAdmissible(
      input.row.rawRow,
      input.row.recordId ?? "",
      input.readScope,
    );
  const key = JSON.stringify([source, index, enabled, attempt]);
  const request = useRef({ key, port: input.port, generation: 0 });
  if (request.current.key !== key || request.current.port !== input.port)
    request.current = {
      key,
      port: input.port,
      generation: request.current.generation + 1,
    };
  const generation = request.current.generation;
  const windowRef = useRef(window);
  windowRef.current = window;

  useEffect(() => {
    if (!enabled || windowRef.current.length === 0) return;
    const controller = new AbortController();
    const recordIds = windowRef.current.map((link) => link.evidenceRecordId);
    let live = true;
    const current = () =>
      live &&
      request.current.generation === generation &&
      !controller.signal.aborted;
    void input.port
      .query({ recordIds, signal: controller.signal })
      .then((result) => {
        if (!current()) return;
        if (result.kind === "accepted") {
          setObservation({
            key,
            generation,
            port: input.port,
            kind: "ready",
            items: result.items,
          });
          return;
        }
        setObservation({
          key,
          generation,
          port: input.port,
          kind: "failed",
          items: [],
        });
        if (
          result.kind === "rejected" &&
          (result.failure.kind === "authentication_required" ||
            result.failure.kind === "authorization_lost" ||
            result.failure.presentation?.family ===
              "permission_or_incident_access_loss")
        )
          void inputRef.current.onAccessFailure();
      })
      .catch(() => {
        if (!current()) return;
        setObservation({
          key,
          generation,
          port: input.port,
          kind: "failed",
          items: [],
        });
      });
    return () => {
      live = false;
      controller.abort();
    };
  }, [key, generation, enabled, input.port]);

  const current =
    enabled &&
    observation?.key === key &&
    observation.generation === generation &&
    observation.port === input.port
      ? observation
      : null;
  return {
    sourceKey,
    kind:
      !enabled || links?.kind === "unavailable" || !links
        ? ("unavailable" as const)
        : total === 0
          ? ("empty" as const)
          : (current?.kind ?? ("loading" as const)),
    items: current?.items ?? [],
    links: enabled ? window : [],
    total: enabled ? total : 0,
    start: index * windowSize,
    previous: index > 0 ? () => setPage({ source, index: index - 1 }) : null,
    next:
      (index + 1) * windowSize < total
        ? () => setPage({ source, index: index + 1 })
        : null,
    retry: () => setAttempt((value) => value + 1),
  };
}
