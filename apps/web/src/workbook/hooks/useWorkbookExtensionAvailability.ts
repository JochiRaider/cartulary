import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ExtensionAvailabilityController,
  type ExtensionDiscoveryProfile,
} from "../../extensions/extensionAvailability";

type WorkbookExtensionAvailabilityOptions = {
  readonly clientInstanceId: string;
  readonly incidentId: string;
  readonly profiles: readonly ExtensionDiscoveryProfile[];
};

/** Owns the shell-lifetime extension controller and all of its mutations. */
export function useWorkbookExtensionAvailability({
  clientInstanceId,
  incidentId,
  profiles,
}: WorkbookExtensionAvailabilityOptions) {
  const controller = useMemo(
    () =>
      new ExtensionAvailabilityController({
        clientInstanceId,
        incidentId,
      }),
    [clientInstanceId, incidentId],
  );
  const [revision, setRevision] = useState(0);
  const publishChange = useCallback(() => {
    setRevision((current) => current + 1);
  }, []);

  useEffect(() => {
    if (controller.setDiscovery(profiles)) {
      publishChange();
    }
  }, [controller, profiles, publishChange]);

  const invalidate = useCallback(() => {
    controller.invalidate();
    publishChange();
  }, [controller, publishChange]);

  return {
    controller,
    invalidate,
    publishChange,
    revision,
  };
}

export function useWorkbookExtensionFallback({
  active,
  available,
  onFallback,
}: {
  readonly active: boolean;
  readonly available: boolean;
  readonly onFallback: () => void;
}) {
  useEffect(() => {
    if (active && !available) {
      onFallback();
    }
  }, [active, available, onFallback]);
}
