package reporting

import "bytes"

// Compare the pure materialization artifacts before a renderer sees bytes.
// Native output repetition alone cannot establish deterministic selection,
// redaction, expansion or token allocation.
func compareRenderMaterialization(first, second RedactionResult) error {
	for _, artifact := range []struct {
		first, second               []byte
		firstSHA, secondSHA, reason string
	}{
		{first.ModelJSON, second.ModelJSON, first.ModelSHA256, second.ModelSHA256, "export_model_mismatch"},
		{first.ManifestJSON, second.ManifestJSON, first.ManifestSHA256, second.ManifestSHA256, "redaction_manifest_mismatch"},
		{first.TokenManifestJSON, second.TokenManifestJSON, first.TokenManifestSHA256, second.TokenManifestSHA256, "token_manifest_mismatch"},
	} {
		if artifact.firstSHA != artifact.secondSHA || !bytes.Equal(artifact.first, artifact.second) {
			return newRenderValidationError("nondeterministic_render", artifact.reason)
		}
	}
	return nil
}
