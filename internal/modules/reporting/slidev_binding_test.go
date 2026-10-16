package reporting

import (
	"context"
	"encoding/json"
	"testing"
)

func assertFrozenSlidevSourceBinding(t *testing.T, payload renderInput, rendered RenderedRelease) {
	t.Helper()
	for name, mutate := range map[string]func(*RenderedRelease){
		"altered deck": func(r *RenderedRelease) {
			for i := range r.RenderBundle.Files {
				if r.RenderBundle.Files[i].Role != renderBundleRoleDeckModel {
					continue
				}
				var deck reportingDeckModel
				if err := json.Unmarshal(r.RenderBundle.Files[i].Bytes, &deck); err != nil {
					t.Fatal(err)
				}
				deck.Slides[0].Title = "Substituted presentation"
				r.RenderBundle.Files[i].Bytes, _ = canonicalJSON(deck)
			}
		},
		"retired deck shape": func(r *RenderedRelease) {
			for i := range r.RenderBundle.Files {
				if r.RenderBundle.Files[i].Role == renderBundleRoleDeckModel {
					r.RenderBundle.Files[i].Bytes = []byte(`{"schema_id":"cartulary.reporting_slide_deck.v2","slides":[{"ordinal":1,"layout":"default","blocks":[{"text":"flat"}]}]}`)
				}
			}
		},
		"altered source": func(r *RenderedRelease) {
			for i := range r.RenderBundle.Files {
				if r.RenderBundle.Files[i].Role == renderBundleRoleSourceSlidev {
					r.RenderBundle.Files[i].Bytes = append(append([]byte{}, r.RenderBundle.Files[i].Bytes...), []byte("\nSubstituted presentation.\n")...)
				}
			}
		},
		"missing deck": func(r *RenderedRelease) {
			for i, file := range r.RenderBundle.Files {
				if file.Role == renderBundleRoleDeckModel {
					r.RenderBundle.Files = append(r.RenderBundle.Files[:i], r.RenderBundle.Files[i+1:]...)
					break
				}
			}
		},
		"renamed deck": func(r *RenderedRelease) {
			for i := range r.RenderBundle.Files {
				if r.RenderBundle.Files[i].Role == renderBundleRoleDeckModel {
					r.RenderBundle.Files[i].Path = "intermediate/other-deck.json"
				}
			}
		},
		"extra source": func(r *RenderedRelease) {
			r.RenderBundle.Files = append(r.RenderBundle.Files, newBundleFile("extra.md", renderBundleRoleSourceSlidev, "text/markdown; charset=utf-8", []byte("Extra content"), true))
		},
		"unbound diagram": func(r *RenderedRelease) {
			r.RenderBundle.Files = append(r.RenderBundle.Files, newBundleFile("diagrams/unbound.svg", renderBundleRoleRenderedSVG, "image/svg+xml", []byte(`<svg xmlns="http://www.w3.org/2000/svg"/>`), true))
		},
		"missing requested PDF": func(r *RenderedRelease) {
			for i, file := range r.RenderBundle.Files {
				if file.Role == renderBundleRoleRenderedPDF {
					r.RenderBundle.Files = append(r.RenderBundle.Files[:i], r.RenderBundle.Files[i+1:]...)
					rebindFixtureBrowserUse(t, r, false)
					return
				}
			}
			t.Fatal("required test PDF missing")
		},
	} {
		t.Run("frozen presentation "+name, func(t *testing.T) {
			candidate := rendered
			candidate.RenderBundle.Files = append([]RenderBundleFile{}, rendered.RenderBundle.Files...)
			mutate(&candidate)
			rehashPresentationFixture(t, &candidate)
			boundaryErr := validateShareableBundle(candidate.RenderBundle)
			if name == "renamed deck" || name == "extra source" {
				if boundaryErr == nil {
					t.Fatal("role/path mismatch passed the shareable boundary")
				}
			} else if boundaryErr != nil {
				t.Fatal("tampered fixture must remain internally consistent", boundaryErr)
			}
			if validateRenderCandidateBinding(payload, candidate) == nil {
				t.Fatal("rehashing replaced admitted presentation")
			}
		})
	}
	payload.OutputOptions = fixtureOutputOptions(OutputKindSlidev, payload.ReleaseScope)
	var options map[string]any
	if err := json.Unmarshal(payload.OutputOptions, &options); err != nil {
		t.Fatal(err)
	}
	options["source_only"], options["pdf"], options["rendered_diagrams"] = true, false, false
	payload.OutputOptions, _ = canonicalJSON(options)
	onlySource, _, err := fixtureRenderer().Render(context.Background(), payload)
	if err != nil || validateRenderCandidateBinding(payload, onlySource) != nil {
		t.Fatal("source-only fixture rejected", err)
	}
	onlySource.RenderBundle.Files = append(onlySource.RenderBundle.Files, newBundleFile("deck.pdf", renderBundleRoleRenderedPDF, "application/pdf", []byte("unrequested rendered bytes"), true))
	rebindFixtureBrowserUse(t, &onlySource, true)
	rehashPresentationFixture(t, &onlySource)
	if validateShareableBundle(onlySource.RenderBundle) != nil || validateRenderCandidateBinding(payload, onlySource) == nil {
		t.Fatal("source-only admission accepted unrequested rendered output")
	}
}

// Make the hostile fixture agree about browser use so output-option rejection
// cannot be attributed to an incidental stale toolchain claim.
func rebindFixtureBrowserUse(t *testing.T, rendered *RenderedRelease, used bool) {
	t.Helper()
	for i := range rendered.RenderBundle.Files {
		file := &rendered.RenderBundle.Files[i]
		if file.Role != renderBundleRoleToolchainSnapshot {
			continue
		}
		var snapshot ToolchainSnapshot
		if err := json.Unmarshal(file.Bytes, &snapshot); err != nil {
			t.Fatal(err)
		}
		snapshot.ChromiumVersion, snapshot.ChromiumExecutablePath, snapshot.ChromiumExecutableSHA256 = nil, nil, nil
		snapshot.BrowserLaunchArgs = []string{}
		if used {
			fixture, _ := (fixtureRenderRuntime{}).Toolchain()
			snapshot.ChromiumVersion, snapshot.ChromiumExecutablePath, snapshot.ChromiumExecutableSHA256 = fixture.ChromiumVersion, fixture.ChromiumExecutablePath, fixture.ChromiumExecutableSHA256
		}
		var err error
		file.Bytes, err = canonicalJSON(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		digest := reportingObjectDigest(snapshot.SchemaID, file.Bytes)
		rendered.RenderBundle.Manifest.ToolchainSnapshotSHA256 = &digest
	}
}

func rehashPresentationFixture(t *testing.T, rendered *RenderedRelease) {
	t.Helper()
	sortBundleFiles(rendered.RenderBundle.Files)
	rendered.RenderBundle.Manifest.Files = []RenderBundleManifestFile{}
	for i := range rendered.RenderBundle.Files {
		file := &rendered.RenderBundle.Files[i]
		file.SizeBytes, file.SHA256 = int64(len(file.Bytes)), hashHex(file.Bytes)
		rendered.RenderBundle.Manifest.Files = append(rendered.RenderBundle.Manifest.Files, file.RenderBundleManifestFile)
	}
	var err error
	rendered.RenderBundle.ManifestJSON, err = canonicalJSON(rendered.RenderBundle.Manifest)
	if err != nil {
		t.Fatal(err)
	}
	rendered.RenderBundle.ManifestSHA256 = reportingObjectDigest(RenderBundleManifestSchemaID, rendered.RenderBundle.ManifestJSON)
	rendered.OutputSHA256 = rendered.RenderBundle.ManifestSHA256
}
