package reporting

import "errors"

func validateFrozenRenderProfile(payload renderInput) error {
	digest, err := ValidateRedactionProfile(payload.RedactionProfile)
	if err != nil || !profileSupportsScope(RedactionProfileView{ProfileStatus: payload.RedactionProfile.ProfileStatus, AllowedReleaseScopes: payload.RedactionProfile.AllowedReleaseScopes}, payload.ReleaseScope) || digest != payload.RedactionProfileSHA256 || payload.RedactionProfile.ProfileID != payload.RedactionProfileID || payload.RedactionProfile.Version != payload.RedactionProfileVersion {
		return errors.New("reporting: invalid frozen render profile")
	}
	view, _, _, err := BuildRedactionProfileView(payload.RedactionProfile, digest)
	if err != nil {
		return err
	}
	paths, err := redactionFieldModelPaths(payload.ExportModel.SnapshotContent)
	if err != nil {
		return err
	}
	if err := addCompositionTextDecisionPaths(payload.CompositionJSON, paths); err != nil {
		return err
	}
	if err := addTemplateAssetDecisionPaths(payload.TemplateContract, paths); err != nil {
		return err
	}
	if err := validateRedactionPathBindings(paths, view, payload.ExportModel.SnapshotContent); err != nil {
		return err
	}
	return nil
}
