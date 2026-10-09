import {
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { sheetRefsEqual } from "../../shared/sheetRef";
import {
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  emptyWorkbookQueryState,
  workbookLayoutStateFromSavedViewLayoutJson,
  workbookQueryStateFromSavedViewQueryJson,
} from "../models/workbookQuery";
import type { SavedViewResource } from "../models/workbookSavedViews";
import { savedViewJSONEqual } from "../models/workbookSavedViews";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import type { WorkbookQueryAcceptance } from "../query/WorkbookQueryBrowser";
import { useWorkbookBrowsingRegistry } from "../query/WorkbookQueryBrowsingContext";
import type { WorkbookRecordLocatorPort } from "../query/WorkbookRecordLocatorPort";
import type {
  WorkbookViewQueryAccepted,
  WorkbookViewQueryPort,
} from "../query/WorkbookViewQueryPort";
import type { WorkbookNavigationHost } from "./WorkbookNavigationHost";
import { WorkbookNavigationPresentation } from "./WorkbookNavigationPresentation";
import type {
  WorkbookReturnOrigin,
  WorkbookSessionNavigation,
  WorkbookSessionPin,
} from "./WorkbookSessionNavigation";
import type {
  WorkbookInspectValue,
  WorkbookNavigationTarget,
  WorkbookWorkbench,
} from "./WorkbookWorkbenchContext";

type Notice = {
  message: string;
  retry: (() => void) | null;
  openBase: (() => void) | null;
};
export function useWorkbookWorkbench(options: {
  readonly session: WorkbookSessionNavigation;
  readonly admitsPage: (page: WorkbookViewQueryAccepted) => boolean;
  readonly incidentId: string;
  readonly actorId: string | null;
  readonly readable: boolean;
  readonly host: WorkbookNavigationHost;
  readonly query: WorkbookViewQueryPort;
  readonly locator: WorkbookRecordLocatorPort;
  readonly extensionAvailable: (
    target: Extract<
      WorkbookNavigationTarget["sheetRef"],
      { kind: "extension_workspace" }
    >,
  ) => boolean;
  readonly authorityFailure: () => void;
}): WorkbookWorkbench {
  const { incidentId, query, locator, readable } = options;
  const registry = useWorkbookBrowsingRegistry();
  const session = options.session;
  const navigation = useSyncExternalStore(
    session.subscribe,
    session.getSnapshot,
  );
  useSyncExternalStore(registry.subscribe, registry.getSnapshot);
  const presentation = useMemo(
    () => new WorkbookNavigationPresentation(registry),
    [registry],
  );
  useLayoutEffect(() => presentation.connect(), [presentation]);
  useLayoutEffect(() => presentation.resume());
  const navigationReady =
    navigation.outcome === "succeeded" && presentation.isCurrent();
  const [inspectValue, setInspectValue] = useState<WorkbookInspectValue | null>(
    null,
  );
  const inspectRevision = useRef(0);
  const [notice, setNotice] = useState<Notice | null>(null);
  const current = useRef(options);
  current.current = options;
  const actor = useRef(options.actorId);
  useLayoutEffect(() => {
    // A transient unknown actor conceals retained state. Confirmed replacement retires it.
    if (options.actorId && actor.current && actor.current !== options.actorId) {
      session.clear();
      presentation.cancel();
      registry.clearNavigation();
      setNotice(null);
      setInspectValue(null);
    }
    if (options.actorId) actor.current = options.actorId;
    session.setReadable(readable);
    if (!readable) {
      presentation.cancel();
      registry.clearNavigation();
      setNotice(null);
      setInspectValue(null);
    }
  }, [session, registry, presentation, options.actorId, readable]);
  useEffect(() => () => session.setReadable(false), [session]);
  const origin = (): WorkbookReturnOrigin => {
    const { snapshot, commands } = current.current.host;
    if (snapshot.startupSheetRef.kind === "extension_workspace")
      return {
        incidentId,
        sheetRef: snapshot.startupSheetRef,
        invoker: "view",
      };
    const browser = registry.find(snapshot.surface);
    const grid = registry.grid(snapshot.surface);
    const anchor = grid?.getActiveCell?.();
    const query =
      browser?.getSnapshot().authored ??
      commands.currentQueryStateForSurface(snapshot.surface);
    if (anchor?.rowIdentity.kind === "core_record")
      browser?.rememberAnchor(anchor.rowIdentity.recordId);
    const selected = snapshot.selectedSavedView;
    return {
      incidentId,
      sheetRef: snapshot.startupSheetRef,
      viewSchemaId: snapshot.surface,
      query,
      layout: buildSavedViewLayoutJson(
        snapshot.activeContract,
        commands.currentLayoutStateForSurface(snapshot.surface),
      ),
      ...(anchor?.rowIdentity.kind === "core_record"
        ? { recordId: anchor.rowIdentity.recordId, fieldKey: anchor.fieldKey }
        : {}),
      ...(selected ? { savedViewVersion: selected.saved_view_version } : {}),
      invoker: anchor ? "grid" : "view",
    };
  };
  const navigate = (
    target: WorkbookNavigationTarget,
    returning?: WorkbookReturnOrigin,
    pin?: WorkbookSessionPin,
    baseFallback: "record" | "surface" | false = false,
    inspect = false,
  ) => {
    const from = origin();
    const intent = {
      target,
      inspect,
      entry:
        baseFallback === "record"
          ? ("base_record" as const)
          : baseFallback === "surface"
            ? ("base_surface" as const)
            : returning
              ? ("return" as const)
              : pin
                ? ("pin" as const)
                : ("open" as const),
      ...(returning ? { returning } : {}),
    };
    const retry = () => navigate(target, returning, pin, baseFallback, inspect);
    setNotice(null);
    void session.navigate(
      intent,
      from,
      async (signal) => {
        const { host, extensionAvailable } = current.current;
        const { commands } = host;
        const admitted = () => !signal.aborted && current.current.readable;
        const attempt = session.getSnapshot().attemptId;
        const present = (
          destination: WorkbookNavigationTarget,
          view: string,
          mode: "record" | "entry" | "extension",
          acceptance?: WorkbookQueryAcceptance,
        ) => {
          presentation.start({
            view,
            mode,
            inspect: inspect && !!destination.recordId,
            ...(destination.recordId ? { recordId: destination.recordId } : {}),
            ...(destination.fieldKey ? { fieldKey: destination.fieldKey } : {}),
            ...(acceptance ? { acceptance } : {}),
            signal,
            admitted: () =>
              session.getSnapshot().attemptId === attempt &&
              session.getSnapshot().outcome === "admitted",
            committed: () =>
              admitted() &&
              sheetRefsEqual(
                current.current.host.snapshot.startupSheetRef,
                destination.sheetRef,
              ) &&
              (destination.sheetRef.kind === "extension_workspace" ||
                current.current.host.snapshot.surface === view),
            entryReady: () =>
              current.current.host.snapshot.gridEntryFocusRequest.kind ===
              "idle",
            settled: (outcome) => {
              session.completePresentation(attempt, outcome);
              if (outcome === "failed" && admitted())
                setNotice({
                  message:
                    "This destination could not be presented. Retry navigation.",
                  retry,
                  openBase: null,
                });
            },
          });
        };
        const fail = (
          message: string,
          base: "record" | "surface" | false = false,
        ) => {
          if (admitted())
            setNotice({
              message,
              retry,
              openBase: base
                ? () => {
                    const view =
                      returning?.viewSchemaId ??
                      (target.sheetRef.kind === "view_schema"
                        ? target.sheetRef.id
                        : undefined);
                    if (view)
                      navigate(
                        {
                          sheetRef: { kind: "view_schema", id: view },
                          ...(base === "record" && target.recordId
                            ? {
                                recordId: target.recordId,
                                ...(target.fieldKey
                                  ? { fieldKey: target.fieldKey }
                                  : {}),
                              }
                            : {}),
                        },
                        returning,
                        undefined,
                        base,
                      );
                  }
                : null,
            });
          return "failed" as const;
        };
        const failedRead = (failure: WorkbookOperationFailure) => {
          if (
            admitted() &&
            workbookFailureLifecycle(failure).kind === "authority_unavailable"
          )
            current.current.authorityFailure();
          return fail(
            "Navigation was not applied. Your previous view is retained.",
          );
        };
        if (target.sheetRef.kind === "extension_workspace") {
          if (!extensionAvailable(target.sheetRef))
            return fail("This workspace is unavailable.");
          if (!admitted()) return "failed";
          registry.detachPresentation(host.snapshot.surface);
          present(target, host.snapshot.surface, "extension");
          if (sheetRefsEqual(from.sheetRef, target.sheetRef)) return "same";
          registry.grid(host.snapshot.surface)?.detachEdit?.();
          commands.selectExtensionWorkspace(target.sheetRef);
          return "changed";
        }
        // Ordinary surface selection retains its established creation-first entry and
        // bounded browsing checkpoint. Record-directed pivots, pins and Return below
        // must instead stage an authorized destination read before replacing context.
        if (
          target.sheetRef.kind === "view_schema" &&
          !target.recordId &&
          !returning &&
          !pin &&
          !baseFallback
        ) {
          if (!admitted()) return "failed";
          registry.detachPresentation(host.snapshot.surface);
          registry.grid(host.snapshot.surface)?.detachEdit?.();
          present(target, target.sheetRef.id, "entry");
          commands.selectWorkbookSurface(target.sheetRef.id, {
            focusFirstGridTarget: true,
          });
          return sheetRefsEqual(from.sheetRef, target.sheetRef)
            ? "same"
            : "changed";
        }
        let resource: SavedViewResource | null = null;
        let view = target.sheetRef.id;
        const observation =
          target.sheetRef.kind === "saved_view"
            ? host.savedViews.observe(target.sheetRef.id, signal)
            : null;
        try {
          if (observation) {
            const result = await observation.result;
            if (!admitted()) return "failed";
            if (result.kind === "aborted") {
              session.cancel();
              return "failed";
            }
            if (result.kind === "rejected") {
              const unavailable = result.failure.kind === "unavailable_target";
              if (pin && unavailable) session.concealPin(pin);
              if (
                result.failure.kind === "authentication_required" ||
                result.failure.kind === "authorization_denied"
              )
                current.current.authorityFailure();
              return fail(
                "This saved view could not be opened.",
                unavailable && returning?.viewSchemaId ? "surface" : false,
              );
            }
            resource = result.value;
            view = resource.view_schema_id;
          }
          const contract = workbookContractForViewSchemaId(view);
          const queryState =
            baseFallback || (pin && target.sheetRef.kind === "view_schema")
              ? emptyWorkbookQueryState()
              : (returning?.query ??
                (resource
                  ? workbookQueryStateFromSavedViewQueryJson(
                      contract,
                      resource.query_json,
                    )
                  : commands.currentQueryStateForSurface(view)));
          const layoutJson = baseFallback
            ? undefined
            : (returning?.layout ?? resource?.layout_json);
          const layout = layoutJson
            ? workbookLayoutStateFromSavedViewLayoutJson(contract, layoutJson)
            : commands.currentLayoutStateForSurface(view);
          if (layout === null)
            return fail(
              "The captured view configuration is no longer compatible.",
              "surface",
            );
          if (
            returning?.query &&
            !baseFallback &&
            !savedViewJSONEqual(
              returning.query,
              workbookQueryStateFromSavedViewQueryJson(
                contract,
                buildSavedViewQueryJson(contract, returning.query),
              ),
            )
          )
            return fail(
              "The captured query is no longer compatible.",
              "surface",
            );
          const anchor =
            baseFallback === "surface" ? undefined : target.recordId;
          let page: WorkbookViewQueryAccepted | null = null;
          // Return shares two automatic recovery attempts: one invalid-cursor restart and one locator.
          const checkpoint =
            returning && !baseFallback
              ? registry.checkpoint(view, queryState)
              : null;
          if (!anchor || checkpoint) {
            let read = await query.query({
              contract,
              queryState,
              signal,
              limit: 100,
              ...(checkpoint
                ? {
                    cursorToken: checkpoint.request.cursorToken,
                    expectedCanonicalQuery: checkpoint.canonical,
                  }
                : {}),
            });
            if (!admitted() || read.kind === "aborted") return "failed";
            if (
              read.kind === "rejected" &&
              checkpoint?.request.cursorToken &&
              read.failure.publicReason === "invalid_cursor_token"
            )
              read = await query.query({
                contract,
                queryState,
                signal,
                limit: 100,
              });
            if (!admitted() || read.kind === "aborted") return "failed";
            if (read.kind === "rejected") return failedRead(read.failure);
            page = read.value;
          }
          if (anchor && !page?.rows.some((row) => row.record_id === anchor)) {
            const located = await locator.locate({
              contract,
              recordId: anchor,
              queryState,
              signal,
            });
            if (!admitted() || located.kind === "aborted") return "failed";
            if (located.kind === "rejected") return failedRead(located.failure);
            if (located.value.outcome !== "located") {
              if (located.value.outcome === "unavailable" && pin)
                session.concealPin(pin);
              return fail(
                located.value.outcome === "outside_query"
                  ? "This record is outside the current query. Open its base view explicitly to continue."
                  : "This record is unavailable.",
                located.value.outcome === "outside_query"
                  ? "record"
                  : "surface",
              );
            }
            page = located.value.page;
          }
          if (!page || !admitted()) return "failed";
          if (!current.current.admitsPage(page))
            return fail(
              "The destination changed during navigation. Retry to read current records.",
            );
          const same =
            sheetRefsEqual(from.sheetRef, target.sheetRef) &&
            from.recordId === anchor &&
            savedViewJSONEqual(from.query, queryState) &&
            savedViewJSONEqual(
              from.layout,
              buildSavedViewLayoutJson(contract, layout),
            );
          registry.detachPresentation(host.snapshot.surface);
          registry.grid(host.snapshot.surface)?.detachEdit?.();
          commands.cancelGridEntryFocus();
          const acceptance = registry.stageNavigation(
            page,
            query.query,
            signal,
          );
          present(
            baseFallback
              ? {
                  sheetRef: { kind: "view_schema", id: view },
                  ...(anchor
                    ? {
                        recordId: anchor,
                        ...(target.fieldKey
                          ? { fieldKey: target.fieldKey }
                          : {}),
                      }
                    : {}),
                }
              : target,
            view,
            anchor || returning ? "record" : "entry",
            acceptance,
          );
          if (resource) host.savedViews.acceptResource(resource);
          // Applying this authorized destination is a fresh requested intent.
          // Its read owner consumes the staged page; a sheet reload would reset
          // that accepted window and the inspector handoff a second time.
          commands.applyQueryStateForSurface(view, structuredClone(queryState));
          commands.applyLayoutStateForSurface(view, layout);
          commands.applyWorkbookIdentity(
            {
              sheetRef: baseFallback
                ? { kind: "view_schema", id: view }
                : target.sheetRef,
              viewSchemaId: view,
            },
            { focusFirstGridTarget: !anchor && !returning },
          );
          if (
            returning?.savedViewVersion &&
            resource &&
            returning.savedViewVersion !== resource.saved_view_version
          )
            setNotice({
              message:
                "The saved view changed. Your captured query and layout were restored; Modified reflects their current difference.",
              retry: null,
              openBase: null,
            });
          return same ? "same" : "changed";
        } finally {
          observation?.release();
        }
      },
      !!returning,
    );
  };
  return {
    navigationReady,
    inspectValue,
    acknowledgeInspectValue: (revision) =>
      setInspectValue((current) =>
        current?.revision === revision ? null : current,
      ),
    requestInspectValue: () => {
      const viewSchemaId = current.current.host.snapshot.surface;
      const anchor = registry.grid(viewSchemaId)?.getActiveCell?.();
      if (
        !current.current.readable ||
        anchor?.rowIdentity.kind !== "core_record"
      )
        return false;
      if (
        !registry.selectLoadedRecord(viewSchemaId, anchor.rowIdentity.recordId)
      )
        return false;
      setInspectValue({
        revision: ++inspectRevision.current,
        viewSchemaId,
        recordId: anchor.rowIdentity.recordId,
        fieldKey: anchor.fieldKey,
      });
      return true;
    },
    session,
    open: (target, inspect = false) =>
      navigate(target, undefined, undefined, false, inspect),
    openPin: (pin) => navigate(pin, undefined, pin),
    pinCurrentView: () => {
      const { snapshot } = current.current.host;
      session.pin({
        incidentId,
        sheetRef: snapshot.startupSheetRef,
        label:
          snapshot.selectedSavedView?.display_name ??
          (snapshot.startupSheetRef.kind === "extension_workspace"
            ? "Network Analysis"
            : snapshot.activeContract.title),
      });
    },
    pinRecord: (viewSchemaId, recordId, label, fieldKey) =>
      session.pin({
        incidentId,
        sheetRef: { kind: "view_schema", id: viewSchemaId },
        recordId,
        label,
        ...(fieldKey ? { fieldKey } : {}),
      }),
    registerInspector: presentation.registerInspector,
    registerInspectorFocus: presentation.registerInspectorFocus,
    pinViewLabel:
      current.current.host.snapshot.startupSheetRef.kind === "view_schema" &&
      !savedViewJSONEqual(
        current.current.host.commands.currentQueryStateForSurface(
          current.current.host.snapshot.surface,
        ),
        emptyWorkbookQueryState(),
      )
        ? "Pin base surface"
        : "Pin view",
    cancelNavigation: () => {
      session.cancel();
      presentation.cancel();
      registry.clearNavigation();
      setNotice(null);
    },
    returnToOrigin: (inspect = false) => {
      const entry = session.getSnapshot().trail.at(-1);
      if (entry) navigate(entry, entry, undefined, false, inspect);
    },
    message: notice?.message ?? null,
    retry: notice?.retry ?? null,
    openBase: notice?.openBase ?? null,
  };
}
