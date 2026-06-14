import { useState } from "react";

export function useTimelineEvidenceActions() {
  const [inspectorMessage, setInspectorMessage] = useState<string | null>(null);

  return {
    commands: {
      setInspectorMessage,
    },
    snapshot: {
      inspectorMessage,
    },
  };
}
