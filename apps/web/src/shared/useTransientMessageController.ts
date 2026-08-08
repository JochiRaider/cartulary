import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import {
  type FocusEvent,
  type MouseEvent,
  useEffect,
  useRef,
  useState,
} from "react";

export type TransientMessageController = {
  readonly onBlur: (event: FocusEvent<HTMLElement>) => void;
  readonly onFocus: () => void;
  readonly onMouseEnter: (event: MouseEvent<HTMLElement>) => void;
  readonly onMouseLeave: (event: MouseEvent<HTMLElement>) => void;
};

export function useTransientMessageController({
  actionAvailable,
  enabled,
  messageKey,
  onDismiss,
}: {
  readonly actionAvailable: boolean;
  readonly enabled: boolean;
  readonly messageKey: string;
  readonly onDismiss: () => void;
}): TransientMessageController {
  const duration =
    cartularyDesignPresentation.transientConfirmation.visibleUnpausedMs;
  const [pointerPaused, setPointerPaused] = useState(false);
  const [focusPaused, setFocusPaused] = useState(false);
  const [documentPaused, setDocumentPaused] = useState(
    () => typeof document !== "undefined" && document.hidden,
  );
  const remainingRef = useRef<number>(duration);
  const runningSinceRef = useRef<number | null>(null);
  const messageKeyRef = useRef(messageKey);
  const onDismissRef = useRef(onDismiss);
  onDismissRef.current = onDismiss;

  useEffect(() => {
    const onVisibilityChange = () => setDocumentPaused(document.hidden);
    document.addEventListener("visibilitychange", onVisibilityChange);
    return () =>
      document.removeEventListener("visibilitychange", onVisibilityChange);
  }, []);

  const paused = pointerPaused || focusPaused || documentPaused;
  useEffect(() => {
    if (messageKeyRef.current !== messageKey) {
      messageKeyRef.current = messageKey;
      remainingRef.current = duration;
      runningSinceRef.current = null;
    }
    if (!enabled || actionAvailable || paused) {
      return;
    }
    const startedAt = performance.now();
    runningSinceRef.current = startedAt;
    const timeout = window.setTimeout(() => {
      runningSinceRef.current = null;
      remainingRef.current = 0;
      onDismissRef.current();
    }, remainingRef.current);
    return () => {
      window.clearTimeout(timeout);
      if (runningSinceRef.current === startedAt) {
        remainingRef.current = Math.max(
          0,
          remainingRef.current - (performance.now() - startedAt),
        );
        runningSinceRef.current = null;
      }
    };
  }, [actionAvailable, duration, enabled, messageKey, paused]);

  return {
    onBlur: (event) => {
      if (
        event.relatedTarget instanceof Node &&
        event.currentTarget.contains(event.relatedTarget)
      ) {
        return;
      }
      setFocusPaused(false);
    },
    onFocus: () => setFocusPaused(true),
    onMouseEnter: () => setPointerPaused(true),
    onMouseLeave: () => setPointerPaused(false),
  };
}
