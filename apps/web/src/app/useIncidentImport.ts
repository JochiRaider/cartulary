import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import {
  cancelImportJob,
  importIncidentBundle,
  readImportJob,
} from "./api/incidentImportClient";
import {
  admissionUnresolved,
  type ImportAuthority,
  IncidentImportController,
  type IncidentImportPorts,
} from "./incidentImportModel";

export function useIncidentImport(options: {
  authority: ImportAuthority | null;
  confirmAccess: IncidentImportPorts["confirmAccess"];
  active: boolean;
  isCurrent: IncidentImportPorts["isCurrent"];
  authorizationFailed: IncidentImportPorts["authorizationFailed"];
  openIncident: IncidentImportPorts["openIncident"];
}) {
  const current = useRef(options);
  current.current = options;
  const [controller] = useState(
    () =>
      new IncidentImportController({
        confirmAccess: (authority, signal, canAccept) =>
          current.current.confirmAccess(authority, signal, canAccept),
        admit: importIncidentBundle,
        read: readImportJob,
        cancel: cancelImportJob,
        isCurrent: (authority) => current.current.isCurrent(authority),
        authorizationFailed: (status) =>
          current.current.authorizationFailed(status),
        openIncident: (id, signal, canNavigate) =>
          current.current.openIncident(id, signal, canNavigate),
      }),
  );
  const mounted = useRef(false);
  const lifetime = options.authority?.lifetime ?? null;
  const actorId = options.authority?.actorId ?? null;
  useLayoutEffect(() => {
    controller.setAuthority(
      lifetime !== null && actorId !== null ? { lifetime, actorId } : null,
    );
  }, [controller, lifetime, actorId]);
  useLayoutEffect(() => {
    controller.setActive(
      options.active && document.visibilityState !== "hidden",
    );
  }, [controller, options.active]);
  useLayoutEffect(() => {
    mounted.current = true;
    const visibility = () =>
      controller.setActive(
        current.current.active && document.visibilityState !== "hidden",
      );
    const unload = (event: BeforeUnloadEvent) => {
      if (!admissionUnresolved(controller.getSnapshot())) return;
      event.preventDefault();
      event.returnValue = "";
    };
    document.addEventListener("visibilitychange", visibility);
    window.addEventListener("beforeunload", unload);
    return () => {
      mounted.current = false;
      controller.setActive(false);
      document.removeEventListener("visibilitychange", visibility);
      window.removeEventListener("beforeunload", unload);
      queueMicrotask(() => {
        if (!mounted.current) controller.dispose();
      });
    };
  }, [controller]);
  return useIncidentImportPresentation(controller, options.active);
}

/** DOM focus and native file-input clearing belong to the React binding. */
export function useIncidentImportPresentation(
  controller: IncidentImportController,
  active: boolean,
) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const fileInputRef = useRef<HTMLInputElement>(null);
  const jobHeadingRef = useRef<HTMLHeadingElement>(null);
  const retryAdmissionRef = useRef<HTMLButtonElement>(null);
  const recoveryOrigin = useRef<Element | null>(null);
  const retryAccessRef = useRef<HTMLButtonElement>(null);
  const accessOrigin = useRef<Element | null>(null);
  const cancellationOrigin = useRef<Element | null>(null);
  useLayoutEffect(() => {
    if (state.selectedFile === null && fileInputRef.current)
      fileInputRef.current.value = "";
    if (!active || document.visibilityState === "hidden") {
      accessOrigin.current = null;
      recoveryOrigin.current = null;
      cancellationOrigin.current = null;
      return;
    }
    if (state.access !== "checking" && accessOrigin.current !== null) {
      const origin = accessOrigin.current;
      accessOrigin.current = null;
      if (!origin.isConnected && document.activeElement === document.body)
        (state.access === "ready"
          ? fileInputRef.current
          : retryAccessRef.current
        )?.focus();
    }
    if (state.admission.kind !== "pending" && recoveryOrigin.current !== null) {
      const origin = recoveryOrigin.current;
      recoveryOrigin.current = null;
      if (!origin.isConnected && document.activeElement === document.body) {
        const target =
          state.admission.kind === "uncertain"
            ? retryAdmissionRef.current
            : state.admission.kind === "rejected"
              ? fileInputRef.current
              : jobHeadingRef.current;
        target?.focus();
      }
    }
    const selected = state.selectedJobId
      ? state.jobs[state.selectedJobId]
      : undefined;
    if (
      state.action.kind !== "checking" &&
      selected?.cancellation.kind !== "pending" &&
      cancellationOrigin.current !== null
    ) {
      const origin = cancellationOrigin.current;
      cancellationOrigin.current = null;
      if (!origin.isConnected && document.activeElement === document.body)
        jobHeadingRef.current?.focus();
    }
  }, [
    active,
    state.access,
    state.action.kind,
    state.admission.kind,
    state.selectedFile,
    state.jobs,
    state.selectedJobId,
  ]);
  return {
    controller,
    state,
    fileInputRef,
    jobHeadingRef,
    retryAdmissionRef,
    retryAccessRef,
    retryAccess: () => {
      accessOrigin.current = document.activeElement;
      controller.retryAccess();
    },
    retryAdmission: () => {
      recoveryOrigin.current = document.activeElement;
      controller.retryAdmission();
    },
    cancel: () => {
      cancellationOrigin.current = document.activeElement;
      controller.cancel();
    },
  };
}
export type IncidentImportPresentation = ReturnType<
  typeof useIncidentImportPresentation
>;
