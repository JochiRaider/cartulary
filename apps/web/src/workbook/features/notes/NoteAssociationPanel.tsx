import { workbookInspectorFeatureActionTestId } from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import { WorkbookAuthoringReferencePicker } from "../../components/WorkbookAuthoringReferencePicker";
import {
  workbookFormFieldsStyle,
  workbookFormMessageStyle,
} from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  type PresentInspectorRegion,
  type WorkbookInspectorPanelContentModel,
  WorkbookInspectorRegionContent,
  type WorkbookInspectorRegionModel,
} from "../../inspector/presentation/WorkbookInspectorPanelContent";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { NoteAssociationResult } from "./NoteAssociationRecovery";
import {
  type NoteAssociationKind,
  noteAssociationListKey,
  noteAssociationView,
} from "./noteAssociationOperation";
import { noteSourceViews } from "./noteCreateModel";
import type { WorkbookNoteAssociationOwner } from "./WorkbookNoteAssociationOwner";

const kindLabels = {
  source: "Sources",
  evidence: "Evidence",
  related_note: "Related notes",
} as const;
export function NoteAssociationPanel({
  owner,
  row,
  kind,
  sheetRef,
  label,
  onNavigateNote,
  present = (model) => <WorkbookInspectorRegionContent model={model} />,
}: {
  readonly owner: WorkbookNoteAssociationOwner;
  readonly row: WorkbookQueryRow;
  readonly kind: NoteAssociationKind;
  readonly sheetRef: SheetRef;
  readonly label: string;
  readonly onNavigateNote: (recordId: string) => Promise<void>;
  readonly present?: PresentInspectorRegion;
}) {
  const feature = requireViewContract(
    noteAssociationView,
  ).inspectorConfig.featureGroups.find(
    (feature) =>
      feature.mutates &&
      feature.routeBinding.owner === "note_associations_route" &&
      feature.routeBinding.actionKey === kind,
  );
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot),
    reasonId = useId();
  const [picker, setPicker] = useState(false),
    trigger = useRef<HTMLButtonElement>(null);
  const reader = owner.getReader(),
    key = noteAssociationListKey(row.record_id, kind),
    list = state.lists[key];
  const pickerReader = useMemo(
    () =>
      reader
        ? {
            availableViews: reader.availableViews,
            page: async (input: Parameters<typeof reader.page>[0]) => {
              const result = await reader.page(input);
              return result.kind === "accepted"
                ? {
                    ...result,
                    value: {
                      ...result.value,
                      candidates: result.value.candidates.filter(
                        (item) => item.recordId !== row.record_id,
                      ),
                    },
                  }
                : result;
            },
          }
        : null,
    [reader, row.record_id],
  );
  useEffect(() => {
    if (
      state.authority &&
      state.candidateRevision === owner.getSnapshot().candidateRevision
    )
      void owner.read(row.record_id, kind);
  }, [owner, row.record_id, kind, state.candidateRevision, state.authority]);
  const close = () => {
    setPicker(false);
    trigger.current?.focus({ preventScroll: true });
  };
  if (!state.authority || !feature) return null;
  const reason = !owner.canReplay()
    ? "An editor, reviewer, or admin role is required to manage associations."
    : state.authority.closed
      ? "This incident is closed. Associations are read-only."
      : owner.blocksRecord(row.record_id)
        ? "An association operation is pending. Recover it before making another change."
        : null;
  const page = list?.page,
    incoming =
      page?.items.filter(
        (item) => kind === "related_note" && item.direction === "incoming",
      ) ?? [],
    managed =
      page?.items.filter(
        (item) => kind !== "related_note" || item.direction === "outgoing",
      ) ?? [];
  const renderItems = (
    items: NonNullable<typeof page>["items"],
    readonly: boolean,
  ) => (
    <ul
      style={{
        ...workbookFormFieldsStyle,
        margin: 0,
        paddingInlineStart: "var(--ct-spacing-lg)",
      }}
    >
      {items.map((item) => (
        <li key={item.item_ref} style={{ overflowWrap: "anywhere" }}>
          {item.view_schema_id === noteAssociationView ? (
            <Button
              tone="secondary"
              onClick={() => void onNavigateNote(item.counterpart_record_id)}
            >
              {item.display_label}
            </Button>
          ) : (
            <span>{item.display_label}</span>
          )}
          {!readonly ? (
            <Button
              tone="secondary"
              aria-label={`Remove ${item.display_label}`}
              style={{ marginInlineStart: "var(--ct-spacing-xs)" }}
              disabled={reason !== null || !page}
              aria-describedby={reason ? reasonId : undefined}
              onClick={() =>
                void owner.submit({
                  row: {
                    ...row,
                    row_version: page?.row_version ?? row.row_version,
                  },
                  kind,
                  sheetRef,
                  label,
                  actions: [{ op: "remove", item_ref: item.item_ref }],
                })
              }
            >
              Remove
            </Button>
          ) : null}
        </li>
      ))}
    </ul>
  );
  const content: WorkbookInspectorPanelContentModel = page?.items.length
    ? {
        kind: "populated",
        content: (
          <>
            {managed.length ? (
              renderItems(managed, false)
            ) : (
              <p style={workbookFormMessageStyle}>No outgoing references.</p>
            )}
            {incoming.length ? (
              <section aria-label="Referenced by">
                <h4 style={workbookFormMessageStyle}>Referenced by</h4>
                {renderItems(incoming, true)}
              </section>
            ) : null}
          </>
        ),
      }
    : {
        kind: "empty",
        message: page?.next_cursor_token
          ? `No ${kindLabels[kind].toLowerCase()} in the loaded page. Load more to continue.`
          : `No ${kindLabels[kind].toLowerCase()} associated with this Note.`,
      };
  const readNotice = owner.readNotice(row.record_id, kind);
  const model: WorkbookInspectorRegionModel = {
    access: "readable",
    ...(readNotice
      ? {
          notice: {
            value: readNotice,
            consume: owner.inspectorNotices.consume,
          },
        }
      : {}),
    data: !list
      ? {
          state: "unavailable",
          cause: "not_requested",
          message: "Associations have not been loaded.",
        }
      : list.state === "initial_loading"
        ? { state: "initial_loading" }
        : list.state === "unavailable"
          ? {
              state: "unavailable",
              cause: "load_failed",
              message: list.message ?? "Could not load associations.",
            }
          : list.state === "stale_failure"
            ? {
                state: "stale_failure",
                content,
                message: list.message ?? "Could not refresh associations.",
              }
            : { state: list.state, content },
  };
  const entries = state.entries.filter(
    (entry) =>
      entry.attempt.review.row.record_id === row.record_id &&
      entry.attempt.review.kind === kind,
  );
  const latest = entries.at(-1);
  return (
    <section
      aria-label={`Note ${kindLabels[kind].toLowerCase()}`}
      style={workbookFormFieldsStyle}
    >
      <h4 style={workbookFormMessageStyle}>{kindLabels[kind]}</h4>
      {present({
        ...model,
        commands: (
          <>
            <div
              style={{
                display: "flex",
                flexWrap: "wrap",
                gap: "var(--ct-spacing-xs)",
              }}
            >
              <Button
                tone="secondary"
                disabled={
                  list?.state === "initial_loading" ||
                  list?.state === "refreshing"
                }
                onClick={() => void owner.read(row.record_id, kind)}
              >
                Refresh {kindLabels[kind].toLowerCase()}
              </Button>
              {page?.next_cursor_token ? (
                <Button
                  tone="secondary"
                  disabled={list?.state !== "ready"}
                  onClick={() => void owner.read(row.record_id, kind, true)}
                >
                  Load more {kindLabels[kind].toLowerCase()}
                </Button>
              ) : null}
              <Button
                tone="primary"
                data-testid={workbookInspectorFeatureActionTestId(
                  noteAssociationView,
                  feature.featureGroupKey,
                )}
                data-feature-group-key={feature.featureGroupKey}
                data-route-kind={feature.routeBinding.kind}
                data-route-owner={feature.routeBinding.owner}
                ref={trigger}
                aria-describedby={reason ? reasonId : undefined}
                disabled={reason !== null || !reader || !page}
                onClick={() => setPicker(true)}
              >
                {feature.label}
              </Button>
            </div>
            {reason ? (
              <p id={reasonId} style={workbookFormMessageStyle}>
                {reason}
              </p>
            ) : null}
            {picker && pickerReader && reason === null ? (
              <WorkbookAuthoringReferencePicker
                label={kindLabels[kind]}
                regionLabel={`Link Note ${kindLabels[kind].toLowerCase()}`}
                targetKey={`${key}:${state.generation}`}
                testId={`note-association-${kind}-picker`}
                views={
                  kind === "source"
                    ? noteSourceViews
                    : kind === "evidence"
                      ? ["cartulary.view.evidence.v1"]
                      : [noteAssociationView]
                }
                multiple
                maximum={64}
                selected={[]}
                reader={pickerReader}
                revision={state.candidateRevision}
                disabled={false}
                applyLabel="Link selected records"
                onCancel={close}
                onApply={(selected) => {
                  if (!selected.length) return;
                  close();
                  void owner.submit({
                    row: {
                      ...row,
                      row_version: page?.row_version ?? row.row_version,
                    },
                    kind,
                    sheetRef,
                    label,
                    actions: selected.map((item) => ({
                      op: "add",
                      counterpart_record_id: item.recordId,
                    })),
                  });
                }}
              />
            ) : null}
            {state.errors[key] ? <p role="alert">{state.errors[key]}</p> : null}
            {latest ? (
              <NoteAssociationResult owner={owner} entry={latest} />
            ) : null}
          </>
        ),
      })}
    </section>
  );
}
