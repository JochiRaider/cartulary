# Artifact analysis and reports

Read only the section needed for the next import, image operation, or report.
Analysis and reports can use an existing live-review session. For an entirely
offline review, start `make ui-review UI_MODE=artifacts`, keep it running, and use
the exact printed session locator. This mode needs no browser installation or
Docker; pinned image dependencies still apply.

In command examples, bind `review_session` to that exact locator,
`review_request` to your private request file, and `review_bundle` to a returned
bundle ID. Do not derive IDs from directory order.

## Import an exact source

Read the selected source variant in the
[capture schema](../../../../tools/schemas/cartulary.ui_review_capture_request.v1.schema.json).
For an explicit PNG reference, replace the example path with the selected file:

```json
{"schema_id":"cartulary.ui_review_capture_request.v1","source":"image","path":"/private/selected-reference.png"}
```

For canonical diagnostics, supply the exact selected run root and capture ID:

```json
{"schema_id":"cartulary.ui_review_capture_request.v1","source":"canonical_visual","run_root":"/exact/selected/run","capture_id":"selected-capture-id"}
```

```bash
CARTULARY_OUTPUT_MODE=machine make ui-capture UI_SESSION="$review_session" UI_REQUEST="$review_request"
```

Obtain selection from the user's inputs or identified execution evidence. If
selection is missing or ambiguous, resolve it before importing. The harness owns
capture/test-result joins; do not select attachments by filename similarity,
rewrite unsupported metadata, or choose the newest run. Structurally valid failed
diagnostic runs can be useful inputs. Expected-only imports remain `no_actual`:
they can be reported, but cannot support operations requiring an actual/original
image. A standalone PNG supplies no DOM or accessibility observations.

Open the imported image and inspect the returned bundle's source and limitations.
Only load relevant fields of the
[bundle schema](../../../../tools/schemas/cartulary.ui_review_bundle.v1.schema.json)
when interpreting provenance or components.

## Derive only useful images

Read the [analysis schema](../../../../tools/schemas/cartulary.ui_review_analysis_request.v1.schema.json)
for the selected operation. Replace the example bundle ID with a returned one:

```json
{"schema_id":"cartulary.ui_review_analysis_request.v1","bundle_id":"bundle-1","operations":["contact_sheet"]}
```

```bash
CARTULARY_OUTPUT_MODE=machine make ui-analyze UI_SESSION="$review_session" UI_REQUEST="$review_request"
```

Choose operations by the review question:

| Operation | Useful for | Interpretation |
| --- | --- | --- |
| `contact_sheet` | Triaging available image channels and crops. | Inspect relevant originals/crops at native detail before concluding. |
| `crop` | Examining a specific region. | Use the prescribed source-pixel coordinates; no guessed CSS-to-image scaling. |
| `overlay` | Inspecting recorded element geometry. | Requires available observations and rectangles. |
| `exact_diff` | Locating exact RGBA changes between selected bundles. | Requires explicit comparison and equal dimensions; numerical difference is not a product assertion. |

For multiple operations, use the contract's sorted, unique operation list and
include only the associated fields. Analysis creates a new immutable bundle;
inspect its returned ID and image references while retaining parent provenance.

Use `matched_capture` only for compatible canonical capture metadata, including
scope and masks. Use an explicitly intended `reference` comparison for a visual
reference question; it carries a cross-source limitation. A failed matched
comparison is not permission to silently weaken it to a reference comparison.
Do not resize, align, introduce a tolerance, or refresh goldens to make a diff pass.

## Render a report when useful

```bash
CARTULARY_OUTPUT_MODE=machine make ui-review-report UI_SESSION="$review_session" UI_BUNDLE="$review_bundle"
```

Open the returned local HTML with an available local viewer while the session is
alive. The report is offline and private; do not start a hosted viewer or upload
it. If the available viewer cannot open local HTML, inspect returned images and
observations directly and state that the interactive report was not inspected.
Unavailable channels and expected-only evidence must remain visible in findings.

Finish through `SKILL.md` cleanup after all intended consumers have inspected the
temporary results. Report structural receipt evidence and concise conclusions;
do not export private bundle contents or promise durable report links.
