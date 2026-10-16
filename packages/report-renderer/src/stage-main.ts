import { RenderFailure } from "./failure.ts";
import {
  decodeStagePort,
  decodeStageRequest,
  executeStage,
  stageRequestLimit,
} from "./stage.ts";

const diagnosticReasons = new Set([
  "stage_browser_library_missing",
  "stage_browser_crashpad_failed",
  "stage_browser_operation_denied",
  "stage_browser_launch_failed",
  "renderer_server_start_failed",
  "renderer_page_load_failed",
  "renderer_print_surface_missing",
  "renderer_fonts_or_paint_failed",
  "renderer_pdf_export_failed",
  "browser_render_error",
  "browser_module_export_missing",
  "browser_slide_import_failed",
  "renderer_resource_missing",
  "renderer_resource_transform_failed",
  "renderer_resource_load_failed",
  "renderer_svg_export_failed",
  "invalid_mermaid_construct",
  "click_export_page_count_mismatch",
  "outbound_request_observed",
  "loopback_not_declared",
  "external_reference_in_rendered_output",
]);

// Installation supplies the executable and cwd. The process protocol accepts
// only stage data on stdin, never command-line capabilities or inherited options.
try {
  if (process.argv.length !== 2) throw new Error("unexpected_arguments");
  const chunks: Buffer[] = [];
  let length = 0;
  for await (const chunk of process.stdin) {
    const bytes = Buffer.from(chunk);
    length += bytes.byteLength;
    if (length > stageRequestLimit) throw new Error("request_limit");
    chunks.push(bytes);
  }
  const input = decodeStageRequest(Buffer.concat(chunks, length));
  const result = await executeStage(
    input,
    process.cwd(),
    decodeStagePort(process.env.CARTULARY_RENDER_LOOPBACK_PORT),
  );
  process.stdout.write(`${JSON.stringify(result)}\n`);
} catch (error: unknown) {
  // The supervisor classifies exit and lifecycle failures. Vendor exceptions,
  // request bytes, local paths and raw destinations never cross this channel.
  const reason =
    error instanceof RenderFailure && diagnosticReasons.has(error.reasonCode)
      ? error.reasonCode
      : "stage_failed";
  process.stderr.write(`report-render-stage: ${reason}\n`);
  process.exitCode = 1;
}
