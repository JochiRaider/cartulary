package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

// Template limits are an owner projection, not deployment settings. Keep units,
// defaults, hard bounds and failure identities together so admission and every
// consuming stage use the same meaning.
type renderLimitRule struct {
	defaultValue, hardMax int64
	unit                  string
	failure, reason       string
}

var renderLimitRules = map[string]renderLimitRule{
	"timeout.export_model_materialization": {120, 300, "seconds", "", ""},
	"timeout.redaction_tokenization":       {120, 300, "seconds", "", ""},
	"timeout.mermaid_per_diagram":          {60, 180, "seconds", "", ""},
	"timeout.slidev_export":                {600, 900, "seconds", "", ""},
	// Zero is an internal inheritance marker, never an admitted scalar.
	"timeout.determinism_rerender":     {0, 900, "seconds", "", ""},
	"timeout.total_attempt":            {900, 1800, "seconds", "", ""},
	"export_model.canonical_bytes":     {64 << 20, 128 << 20, "bytes", "export_model_resource_limit_exceeded", "export_model_canonical_bytes_exceeded"},
	"sections.count":                   {100, 500, "count", "export_model_resource_limit_exceeded", "sections_count_exceeded"},
	"blocks.count":                     {10000, 50000, "count", "export_model_resource_limit_exceeded", "blocks_count_exceeded"},
	"fields.count":                     {100000, 250000, "count", "export_model_resource_limit_exceeded", "fields_count_exceeded"},
	"records.count":                    {50000, 100000, "count", "export_model_resource_limit_exceeded", "records_count_exceeded"},
	"relationships.count":              {100000, 250000, "count", "export_model_resource_limit_exceeded", "relationships_count_exceeded"},
	"diagrams.count":                   {50, 100, "count", "export_model_resource_limit_exceeded", "diagrams_count_exceeded"},
	"subjects.count":                   {25000, 50000, "count", "token_manifest_invalid", "subject_limit_exceeded"},
	"support_refs.count":               {100000, 250000, "count", "support_ref_limit_exceeded", "support_ref_limit_exceeded"},
	"composition.authored_title_chars": {120, 120, "count", "composition_invalid", "authored_title_limit_exceeded"},
	"composition.authored_text_chars":  {2000, 5000, "count", "composition_invalid", "authored_text_limit_exceeded"},
	"slides.count":                     {120, 240, "count", "slidev_source_invalid", "slide_count_exceeded"},
	"blocks_per_slide":                 {30, 60, "count", "slidev_source_invalid", "slide_block_limit_exceeded"},
	"speaker_notes_chars_per_slide":    {10000, 25000, "count", "slidev_source_invalid", "speaker_notes_limit_exceeded"},
	"click_steps_per_slide":            {20, 40, "count", "slidev_source_invalid", "click_step_limit_exceeded"},
	"local_assets.count":               {250, 500, "count", "asset_limit_exceeded", "local_asset_count_exceeded"},
	"local_asset.byte_size":            {25 << 20, 50 << 20, "bytes", "asset_limit_exceeded", "local_asset_bytes_exceeded"},
	"local_assets.total_bytes":         {100 << 20, 250 << 20, "bytes", "asset_limit_exceeded", "local_assets_total_bytes_exceeded"},
	"rendered_pdf.byte_size":           {250 << 20, 500 << 20, "bytes", "bundle_resource_limit_exceeded", "rendered_pdf_bytes_exceeded"},
}

type renderLimits map[string]int64

// Resolve before canonicalizing a manifest: JSON exponent and fraction spellings
// are not integer-second declarations even when their mathematical value is an
// integer. Sort keys so rejected multi-defect declarations have a stable result.
func resolveRenderLimits(declared map[string]json.RawMessage) (renderLimits, error) {
	limits := make(renderLimits, len(renderLimitRules))
	for key, rule := range renderLimitRules {
		limits[key] = rule.defaultValue
	}
	keys := make([]string, 0, len(declared))
	for key := range declared {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		rule, known := renderLimitRules[key]
		reason := "invalid_scalar_value"
		if strings.HasPrefix(key, "timeout.") {
			reason = "invalid_template_timeout_key"
		}
		if !known {
			if !strings.HasPrefix(key, "timeout.") {
				reason = "template_manifest_invalid"
			}
			return nil, newRenderValidationError("invalid_release_request", reason)
		}
		raw := bytes.TrimSpace(declared[key])
		value, multiplier := string(raw), int64(1)
		if rule.unit == "bytes" && len(raw) > 0 && raw[0] == '"' {
			if json.Unmarshal(raw, &value) != nil || !strings.HasSuffix(value, " MiB") {
				return nil, newRenderValidationError("invalid_release_request", reason)
			}
			value, multiplier = strings.TrimSuffix(value, " MiB"), 1<<20
		}
		if value == "" || value[0] < '1' || value[0] > '9' {
			return nil, newRenderValidationError("invalid_release_request", reason)
		}
		for _, c := range value {
			if c < '0' || c > '9' {
				return nil, newRenderValidationError("invalid_release_request", reason)
			}
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil || parsed > rule.hardMax/multiplier {
			return nil, newRenderValidationError("invalid_release_request", "template_limit_exceeds_hard_limit")
		}
		limits[key] = parsed * multiplier
	}
	return limits, nil
}

func (limits renderLimits) timeout(stage string) time.Duration {
	return time.Duration(limits["timeout."+stage]) * time.Second
}

func (limits renderLimits) check(key string, count int64) error {
	if count <= limits[key] {
		return nil
	}
	rule := renderLimitRules[key]
	return &renderValidationError{FailureCode: rule.failure, ReasonCode: rule.reason, SafeDetails: map[string]any{"limit_key": key, "limit": limits[key], "actual": count}}
}

// A parent deadline at the same instant wins, preserving total-attempt
// precedence. Context cancellation still belongs to the caller; no work is
// detached to outlive a timed-out render.
func withRenderDeadline(ctx context.Context, duration time.Duration, stage, key string) (context.Context, context.CancelFunc) {
	return withRenderDeadlineFrom(ctx, time.Now(), duration, stage, key)
}

func withRenderDeadlineFrom(ctx context.Context, started time.Time, duration time.Duration, stage, key string) (context.Context, context.CancelFunc) {
	kind := "stage"
	if key == "timeout.total_attempt" {
		kind = "total"
	}
	failure := &renderValidationError{FailureCode: "render_timeout", ReasonCode: "render_timeout", Stage: stage, SafeDetails: map[string]any{"stage": stage, "timeout_kind": kind, "limit_key": key, "limit": int64(duration / time.Second)}, cause: context.DeadlineExceeded}
	return context.WithDeadlineCause(ctx, started.Add(duration), failure)
}

func classifyRenderContextError(ctx context.Context, err error) *renderValidationError {
	// A cleanup defect must never be hidden by the timer that caused teardown.
	if !errors.Is(err, renderprocess.ErrCleanup) && errors.Is(err, context.DeadlineExceeded) {
		var failure *renderValidationError
		if errors.As(context.Cause(ctx), &failure) {
			// The context owns its cause. Callers may attach the stage's wrapped
			// error, so never return a shared mutable cause pointer.
			copy := *failure
			return &copy
		}
	}
	return classifyRenderRuntimeError(err)
}

func (limits renderLimits) rerenderTimeout(kind string) time.Duration {
	if explicit := limits.timeout("determinism_rerender"); explicit != 0 {
		return explicit
	}
	if kind == OutputKindSlidev {
		return limits.timeout("slidev_export")
	}
	return limits.timeout("mermaid_per_diagram")
}

func validateDeckLimits(limits renderLimits, deck reportingDeckModel) error {
	issues := renderIssueCollector{stage: "slidev_source"}
	issues.add(limits.check("slides.count", int64(len(deck.Slides))))
	for _, slide := range deck.Slides {
		issues.add(limits.check("blocks_per_slide", int64(len(slide.Blocks))))
		issues.add(limits.check("click_steps_per_slide", int64(len(slide.ClickSteps))))
		notes, err := slideSpeakerNotes(slide)
		issues.add(err)
		issues.add(limits.check("speaker_notes_chars_per_slide", int64(utf8.RuneCountInString(strings.Join(notes, "\n")))))
	}
	return issues.finish()
}

// Composition edits the unchunked slides. Chunk only after those edits, retaining
// block identity and order. The first emitted ordinal names the shared base;
// subsequent base slides cannot collide with a previous chunk's identity.
func chunkDeckSlides(slides []reportingDeckSlide, limits renderLimits) ([]reportingDeckSlide, error) {
	maxBlocks := int(limits["blocks_per_slide"])
	if maxBlocks <= 0 {
		return nil, newRenderValidationError("slidev_source_invalid", "slide_block_limit_exceeded")
	}
	result := []reportingDeckSlide{}
	for _, slide := range slides {
		// Click targets and state hashes belong to the final chunk identities.
		// Precomputed steps cannot be discarded or transferred to another slide.
		if len(slide.ClickSteps) != 0 {
			return nil, invalidSlidevBlock()
		}
		chunks := max(1, (len(slide.Blocks)+maxBlocks-1)/maxBlocks)
		if err := limits.check("slides.count", int64(len(result)+chunks)); err != nil {
			return nil, err
		}
		baseID := fmt.Sprintf("sld-%04d", len(result)+1)
		for chunk := 0; chunk < chunks; chunk++ {
			part := slide
			part.SlideID, part.Ordinal = baseID, len(result)+1
			start := chunk * maxBlocks
			blocks, err := copyDeckBlocks(slide.Blocks[start:min(start+maxBlocks, len(slide.Blocks))])
			if err != nil {
				return nil, err
			}
			part.Blocks = blocks
			part.SpeakerNotes = cloneStringPtr(slide.SpeakerNotes)
			part.titlePartitions = append([]string{}, slide.titlePartitions...)
			part.notePartitions = append([]string{}, slide.notePartitions...)
			part.ClickSteps = []reportingClickStep{}
			if chunks > 1 {
				part.SlideID = fmt.Sprintf("%s__chunk_%04d", baseID, chunk+1)
				part.Title = fmt.Sprintf("%s (%d/%d)", slide.Title, chunk+1, chunks)
				if chunk > 0 {
					part.SpeakerNotes = nil
					part.notePartitions = nil
					part.noteSourceRefs, part.noteSupportRefs = nil, nil
				}
			}
			refs := emptyContainerRefs()
			refs.include(part.titleSourceRefs, part.titleSupportRefs, part.titlePartitions)
			refs.include(part.noteSourceRefs, part.noteSupportRefs, part.notePartitions)
			for _, block := range part.Blocks {
				refs.include(block.SourceRefs, block.SupportRefs, block.DisclosurePartitionRefs)
			}
			refs = refs.normalized()
			part.SourceRefs, part.SupportRefs, part.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
			result = append(result, part)
		}
	}
	return result, nil
}
