import type { ChangeEvent } from "react";
import { useState } from "react";
import { importNetworkFlowCSV } from "./networkFlowClient";

export function useNetworkFlowImportController({
  apiBase,
  canImport,
  incidentId,
  onError,
  onImported,
  onMessage,
}: {
  readonly apiBase: string | undefined;
  readonly canImport: boolean;
  readonly incidentId: string;
  readonly onError: (message: string | null) => void;
  readonly onImported: () => Promise<void>;
  readonly onMessage: (message: string) => void;
}) {
  const [importing, setImporting] = useState(false);
  const handleImportChange = async (event: ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0] ?? null;
    event.target.value = "";
    if (file === null || importing || !canImport) {
      return;
    }
    setImporting(true);
    onError(null);
    try {
      await importNetworkFlowCSV({
        apiBase,
        incidentId,
        file,
        onProgress: onMessage,
      });
      onMessage("Import applied.");
      await onImported();
    } catch (caught) {
      onError(caught instanceof Error ? caught.message : "import_failed");
    } finally {
      setImporting(false);
    }
  };
  return { handleImportChange, importing };
}
