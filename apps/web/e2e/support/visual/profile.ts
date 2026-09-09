import type { GetCurrentAccountPreferencesResponse } from "@cartulary/protocol-ts/http";
import { cartularyDefaultThemeId } from "@cartulary/ui-contracts";

export type VisualPresentation = {
  surface_kind: "workbook_shell" | "application_shell";
  density_id: GetCurrentAccountPreferencesResponse["data"]["density_mode"];
  theme_id: typeof cartularyDefaultThemeId;
};

export function assertVisualPresentation(
  declaration: VisualPresentation,
  observed: { density_id: unknown; theme_id: unknown },
) {
  const validDensity =
    declaration.surface_kind === "application_shell"
      ? declaration.density_id === null
      : ["compact", "default", "comfortable"].includes(
          declaration.density_id ?? "",
        );
  if (
    !validDensity ||
    declaration.theme_id !== cartularyDefaultThemeId ||
    observed.density_id !== declaration.density_id ||
    observed.theme_id !== declaration.theme_id
  ) {
    const error = new Error(
      "Rendered presentation must match the declared surface, density, and theme",
    );
    error.name = "CartularyVisualCaptureError";
    throw error;
  }
}
