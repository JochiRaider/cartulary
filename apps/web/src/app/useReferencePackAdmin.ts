import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import {
  ReferencePackAdminController,
  type ReferencePackAdminPorts,
  referencePackTransportPorts,
} from "./referencePackAdminController";
import type { ReferencePackAuthority } from "./referencePackAdminModel";

export function useReferencePackAdmin(options: {
  readonly authority: ReferencePackAuthority | null;
  readonly active: boolean;
  readonly isCurrent: ReferencePackAdminPorts["isCurrent"];
  readonly confirmAccess: ReferencePackAdminPorts["confirmAccess"];
  readonly authorizationFailed: ReferencePackAdminPorts["authorizationFailed"];
}) {
  const current = useRef(options);
  current.current = options;
  const [controller] = useState(
    () =>
      new ReferencePackAdminController({
        ...referencePackTransportPorts,
        isCurrent: (authority) => current.current.isCurrent(authority),
        confirmAccess: (authority, signal, canAccept) =>
          current.current.confirmAccess(authority, signal, canAccept),
        authorizationFailed: (status) =>
          current.current.authorizationFailed(status),
      }),
  );
  const lifetime = options.authority?.lifetime ?? null;
  const actorId = options.authority?.actorId ?? null;
  useLayoutEffect(() => {
    controller.setAuthority(lifetime && actorId ? { lifetime, actorId } : null);
  }, [controller, lifetime, actorId]);
  useLayoutEffect(() => {
    controller.setActive(
      options.active && document.visibilityState !== "hidden",
    );
  }, [controller, options.active]);
  const mounted = useRef(false);
  useLayoutEffect(() => {
    mounted.current = true;
    const visibility = () =>
      controller.setActive(
        current.current.active && document.visibilityState !== "hidden",
      );
    document.addEventListener("visibilitychange", visibility);
    return () => {
      mounted.current = false;
      controller.setActive(false);
      controller.retire();
      document.removeEventListener("visibilitychange", visibility);
      queueMicrotask(() => {
        if (!mounted.current) controller.dispose();
      });
    };
  }, [controller]);
  return controller;
}

/** Native input, focus and scroll ownership are restricted to the visible panel. */
export function useReferencePackAdminPresentation(
  controller: ReferencePackAdminController,
  active: boolean,
) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const rootRef = useRef<HTMLElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const feedbackRef = useRef<HTMLElement>(null);
  const originRef = useRef<HTMLElement | null>(null);
  useLayoutEffect(() => {
    if (state.file === null && fileInputRef.current)
      fileInputRef.current.value = "";
    if (!active || document.visibilityState === "hidden") {
      originRef.current = null;
      return;
    }
    const origin = originRef.current;
    if (!origin) return;
    const focused = document.activeElement;
    if (focused !== origin && focused !== document.body && focused !== null) {
      originRef.current = null;
      return;
    }
    if (!origin.isConnected || origin.matches(":disabled")) {
      const recovery = rootRef.current?.querySelector<HTMLButtonElement>(
        "[data-rp-recovery]:not(:disabled)",
      );
      (recovery ?? feedbackRef.current)?.focus({ preventScroll: true });
      originRef.current = null;
    }
  }, [active, state]);
  useLayoutEffect(() => {
    const root = rootRef.current;
    if (!active || !root) return;
    const scrollPort = () => {
      for (
        let element: HTMLElement | null = root;
        element;
        element = element.parentElement
      ) {
        if (
          /(auto|scroll)/u.test(getComputedStyle(element).overflowY) &&
          element.scrollHeight > element.clientHeight
        )
          return element;
      }
      return document.scrollingElement;
    };
    const port = scrollPort();
    if (port) port.scrollTop = controller.getSnapshot().scrollTop;
    const remember = (event: Event) => {
      if (!controller.getSnapshot().active) return;
      const current = scrollPort();
      if (current && (event.target === current || event.target === document))
        controller.setScrollTop(current.scrollTop);
    };
    document.addEventListener("scroll", remember, true);
    return () => document.removeEventListener("scroll", remember, true);
  }, [active, controller]);
  return {
    state,
    rootRef,
    fileInputRef,
    feedbackRef,
    run: (origin: HTMLElement, action: () => unknown) => {
      originRef.current = origin;
      action();
    },
  };
}
