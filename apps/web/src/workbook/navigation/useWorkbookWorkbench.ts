import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { sheetRefsEqual } from "../../shared/sheetRef";
import type { useWorkbookShellRuntime } from "../hooks/useWorkbookShellRuntime";
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
import { useWorkbookBrowsingRegistry } from "../query/WorkbookQueryBrowsingContext";
import type { WorkbookRecordLocatorPort } from "../query/WorkbookRecordLocatorPort";
import type {
  WorkbookViewQueryAccepted,
  WorkbookViewQueryPort,
} from "../query/WorkbookViewQueryPort";
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

type Runtime = ReturnType<typeof useWorkbookShellRuntime>;
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
  readonly runtime: Runtime;
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
  const [inspectValue, setInspectValue] = useState<WorkbookInspectValue | null>(
    null,
  );
  const inspectRevision = useRef(0);
  const inspectorBindings = useRef(new Map<string, () => void>());
  const inspectorFocus = useRef(
    new Map<string, { recordId: string; focus: () => boolean }>(),
  );
  const registerInspectorFocus = useCallback(
    (view: string, recordId: string, focus: () => boolean) => {
      const binding = { recordId, focus };
      inspectorFocus.current.set(view, binding);
      registry.resumeNavigation();
      return () => {
        if (inspectorFocus.current.get(view) === binding)
          inspectorFocus.current.delete(view);
      };
    },
    [registry],
  );
  const registerInspector = useCallback(
    (view: string, handler: () => void) => {
      inspectorBindings.current.set(view, handler);
      registry.resumeNavigation();
      return () => {
        if (inspectorBindings.current.get(view) === handler)
          inspectorBindings.current.delete(view);
      };
    },
    [registry],
  );
  const [notice, setNotice] = useState<Notice | null>(null);
  const current = useRef(options);
  current.current = options;
  const actor = useRef(options.actorId);
  useLayoutEffect(() => {
    // A transient unknown actor conceals retained state. Confirmed replacement retires it.
    if (options.actorId && actor.current && actor.current !== options.actorId) {
      session.clear();
      registry.clearNavigation();
      setNotice(null);
      setInspectValue(null);
    }
    if (options.actorId) actor.current = options.actorId;
    session.setReadable(readable);
    if (!readable) {
      registry.clearNavigation();
      setNotice(null);
      setInspectValue(null);
    }
  }, [session, registry, options.actorId, readable]);
  useEffect(() => () => session.setReadable(false), [session]);
  const origin = (): WorkbookReturnOrigin => {
    const { snapshot, commands } = current.current.runtime;
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
    const selected = snapshot.savedViewsResource.selectedSavedView;
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
    const key = JSON.stringify([target, returning, baseFallback, inspect]);
    const retry = () => navigate(target, returning, pin, baseFallback, inspect);
    setNotice(null);
    void session.navigate(
      key,
      from,
      async (signal) => {
        const { runtime, extensionAvailable } = current.current;
        const { commands } = runtime;
        const admitted = () => !signal.aborted && current.current.readable;
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
          registry.detachPresentation(runtime.snapshot.surface);
          if (sheetRefsEqual(from.sheetRef, target.sheetRef)) return "same";
          registry.grid(runtime.snapshot.surface)?.detachEdit?.();
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
          registry.detachPresentation(runtime.snapshot.surface);
          registry.grid(runtime.snapshot.surface)?.detachEdit?.();
          commands.selectWorkbookSurface(target.sheetRef.id, {
            focusFirstGridTarget: true,
          });
          return sheetRefsEqual(from.sheetRef, target.sheetRef)
            ? "same"
            : "changed";
        }
        let resource: SavedViewResource | null = null;
        let view = target.sheetRef.id;
        if (target.sheetRef.kind === "saved_view") {
          runtime.savedViewOwner.resources.retain(
            "navigation",
            target.sheetRef.id,
          );
          let unavailable = false;
          try {
            resource = await runtime.savedViewOwner.resources.read(
              target.sheetRef.id,
            );
            unavailable =
              runtime.savedViewOwner.resources.get(target.sheetRef.id)
                ?.status === "unavailable";
          } finally {
            runtime.savedViewOwner.resources.retain("navigation", null);
          }
          if (!admitted()) return "failed";
          if (!resource) {
            if (pin && unavailable) session.concealPin(pin);
            return fail(
              "This saved view could not be opened.",
              returning?.viewSchemaId ? "surface" : false,
            );
          }
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
          return fail("The captured query is no longer compatible.", "surface");
        const anchor = baseFallback === "surface" ? undefined : target.recordId;
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
              located.value.outcome === "outside_query" ? "record" : "surface",
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
        registry.detachPresentation(runtime.snapshot.surface);
        registry.grid(runtime.snapshot.surface)?.detachEdit?.();
        commands.cancelGridEntryFocus();
        let inspectorOpened = false;
        registry.stageNavigation(
          page,
          query.query,
          anchor
            ? {
                recordId: anchor,
                ...(target.fieldKey ? { fieldKey: target.fieldKey } : {}),
              }
            : null,
          inspect
            ? () => {
                if (!admitted()) return true;
                if (!inspectorOpened) {
                  const open = inspectorBindings.current.get(view);
                  if (!open) return false;
                  inspectorOpened = true;
                  open();
                  return false;
                }
                const binding = inspectorFocus.current.get(view);
                return (
                  !!binding && binding.recordId === anchor && binding.focus()
                );
              }
            : undefined,
          {
            signal,
            navigationOnly: !!returning,
            onUnavailable: () => {
              if (admitted())
                setNotice({
                  message: "This destination is unavailable.",
                  retry,
                  openBase: null,
                });
            },
          },
        );
        if (resource) runtime.savedViewOwner.acceptResource(resource);
        commands.applyQueryStateForSurface(view, queryState);
        commands.applyLayoutStateForSurface(view, layout);
        commands.applyWorkbookIdentity(
          {
            sheetRef: baseFallback
              ? { kind: "view_schema", id: view }
              : target.sheetRef,
            viewSchemaId: view,
          },
          { reloadSheet: true, focusFirstGridTarget: !anchor && !returning },
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
      },
      !!returning,
    );
  };
  return {
    inspectValue,
    acknowledgeInspectValue: (revision) =>
      setInspectValue((current) =>
        current?.revision === revision ? null : current,
      ),
    requestInspectValue: () => {
      const viewSchemaId = current.current.runtime.snapshot.surface;
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
      const { snapshot } = current.current.runtime;
      session.pin({
        incidentId,
        sheetRef: snapshot.startupSheetRef,
        label:
          snapshot.savedViewsResource.selectedSavedView?.display_name ??
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
    registerInspector,
    registerInspectorFocus,
    pinViewLabel:
      current.current.runtime.snapshot.startupSheetRef.kind === "view_schema" &&
      !savedViewJSONEqual(
        current.current.runtime.commands.currentQueryStateForSurface(
          current.current.runtime.snapshot.surface,
        ),
        emptyWorkbookQueryState(),
      )
        ? "Pin base surface"
        : "Pin view",
    cancelNavigation: () => {
      session.cancel();
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
