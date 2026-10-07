import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import { useWorkbookCommand } from "../commands/WorkbookCommands";
import { workbookFormInputStyle } from "./workbookFormStyles";

const presets = cartularyDesignPresentation.workbookWorkbench.presets;
export function WorkbookTimelinePresets({
  onApply,
}: {
  readonly onApply: (id: string) => void;
}) {
  return (
    <>
      <label>
        Timeline presets{" "}
        <select
          style={workbookFormInputStyle}
          aria-label="Timeline query preset"
          value=""
          onChange={(event) => {
            if (event.target.value) onApply(event.target.value);
          }}
        >
          <option value="" disabled>
            Choose preset
          </option>
          {presets.map((preset) => (
            <option key={preset.id} value={preset.id}>
              {preset.label}
            </option>
          ))}
        </select>
      </label>
    </>
  );
}
export function WorkbookTimelinePresetCommands({
  onApply,
  surface,
}: {
  readonly onApply: (id: string) => void;
  readonly surface: string;
}) {
  return (
    <>
      {presets.map((preset) => (
        <PresetCommand
          key={preset.id}
          preset={preset}
          onApply={onApply}
          surface={surface}
        />
      ))}
    </>
  );
}
function PresetCommand({
  preset,
  onApply,
  surface,
}: {
  readonly preset: (typeof presets)[number];
  readonly onApply: (id: string) => void;
  readonly surface: string;
}) {
  useWorkbookCommand({
    id: `view.preset.${preset.id}`,
    family: "View",
    label: preset.label,
    terms: ["lens", "preset", "timeline"],
    targetKind: "surface",
    availability: (target) =>
      target.kind === "surface" && target.viewSchemaId === surface
        ? null
        : "This preset is available on Timeline.",
    invoke: (target) => {
      if (target.kind !== "surface" || target.viewSchemaId !== surface)
        return false;
      onApply(preset.id);
      return true;
    },
  });
  return null;
}
