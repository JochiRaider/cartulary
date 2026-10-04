package packformat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// ManifestChecks separates every registry rank without reconstructing malformed
// objects. Size -1 means the member is absent: its dependent checks are
// inapplicable until the required-member check. Oversized bytes are never read.
func ManifestChecks(data []byte, size int64, operator bool) (map[string]CheckFunc, func() Manifest) {
	var value any
	var strictErr error
	var manifest Manifest
	readable := size >= 0 && size <= 1048576
	return map[string]CheckFunc{
		"manifest_encoding": func(_ context.Context, emit FindingSink) error {
			if readable && (canonicaljson.ValidateUnicode(data) != nil || bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf})) {
				return emit(Finding{Path: "manifest.json"})
			}
			return nil
		},
		"manifest_json": func(_ context.Context, emit FindingSink) error {
			if !readable {
				return nil
			}
			value, strictErr = canonicaljson.DecodeStrict(data)
			if errors.Is(strictErr, canonicaljson.ErrDuplicateMember) {
				if bytes.TrimSpace(data)[0] != '{' {
					return emit(Finding{Path: "manifest.json"})
				}
				return nil
			}
			if _, ok := value.(map[string]any); strictErr != nil || !ok {
				return emit(Finding{Path: "manifest.json"})
			}
			return nil
		},
		"manifest_duplicate": func(ctx context.Context, emit FindingSink) error {
			if readable && errors.Is(strictErr, canonicaljson.ErrDuplicateMember) {
				return duplicateManifestFindings(ctx, data, emit)
			}
			return nil
		},
		"manifest_schema": func(_ context.Context, emit FindingSink) error {
			if size < 0 {
				return nil
			}
			if size > 1048576 {
				return emit(LimitFinding("manifest.json", "max_manifest_json_bytes", 1048576, size))
			}
			invalid := false
			report := func(f Finding) error { invalid = true; return emit(f) }
			if err := manifestShape.walk(value, "$", report); err != nil {
				return err
			}
			if !validPackNumbers(value) {
				if err := report(Finding{Path: "$"}); err != nil {
					return err
				}
			}
			if invalid {
				return nil
			}
			if err := json.Unmarshal(data, &manifest); err != nil {
				return errors.New("reference pack: admitted manifest decode failed")
			}
			return manifestFieldFindings(manifest, operator, report)
		},
		"manifest_canonical": func(_ context.Context, emit FindingSink) error {
			if size < 0 {
				return nil
			}
			encoded, err := canonicaljson.Marshal(value)
			if err != nil {
				return err
			}
			if !bytes.Equal(encoded, data) {
				return emit(Finding{Path: "manifest.json"})
			}
			return nil
		},
	}, func() Manifest { return manifest }
}

func HintChecks(data []byte, size int64) (map[string]CheckFunc, func() string) {
	var value any
	return map[string]CheckFunc{
		"bundle_shape": func(_ context.Context, emit FindingSink) error {
			if size < 0 {
				return emit(Finding{Path: "bundle.json"})
			}
			if size > 16384 {
				return emit(LimitFinding("bundle.json", "max_bundle_json_bytes", 16384, size))
			}
			var err error
			value, err = canonicaljson.DecodeStrict(data)
			if err != nil {
				return emit(Finding{Path: "bundle.json"})
			}
			if !validPackNumbers(value) {
				if err := emit(Finding{Path: "$"}); err != nil {
					return err
				}
			}
			return hintShape.walk(value, "$", emit)
		},
		"bundle_canonical": func(_ context.Context, emit FindingSink) error {
			encoded, err := canonicaljson.Marshal(value)
			if err != nil {
				return err
			}
			if !bytes.Equal(encoded, data) {
				return emit(Finding{Path: "bundle.json"})
			}
			return nil
		},
	}, func() string { return value.(map[string]any)["trust_repository_id"].(string) }
}

// All declared profiles admit data only. Earlier closed-schema and inventory
// checks reject active roles/members before this rank is reachable. Reassert
// the complete role/media allowlist here; profile strings remain inert, never
// scanned for guessed executable or credential syntax.
func DisallowedContentFindings(manifest Manifest, emit FindingSink) error {
	if emit == nil {
		return errors.New("reference pack: missing content policy findings sink")
	}
	for i, file := range manifest.Files {
		allowed := file.Role == "payload" && file.MediaType == "application/x-ndjson" || file.Role == "notice" && file.MediaType == "text/plain"
		if !allowed {
			if err := emit(Finding{Path: "$.files[" + strconv.Itoa(i) + "]"}); err != nil {
				return err
			}
		}
	}
	return nil
}
