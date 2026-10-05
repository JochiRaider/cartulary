package reference_data

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"slices"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

// One attempt owns this state. The typed registry, not construction order or
// transport orchestration, invokes every check. Dependent owner programs are
// initialized only after all their prerequisite ranks have succeeded.
type containerVerificationProgram struct {
	attempt                               verificationAttempt
	storage                               VerificationStorage
	source                                *checkedContainerReader
	size                                  int64
	workspace                             VerificationWorkspace
	extract                               func(packformat.Destination) (packformat.Inventory, error)
	inventory                             packformat.Inventory
	paths                                 []string
	manifestBytes                         []byte
	manifestSize                          int64
	manifest                              packformat.Manifest
	manifestResult                        func() packformat.Manifest
	key, version, repository              string
	manifestSHA, payloadSHA, containerSHA string
	trust                                 packformat.TrustSnapshot
	proposal                              func() packformat.TrustProposal
	trustPreparationErr                   error
	sequenceErr                           error
}

func (p *containerVerificationProgram) checks() (map[string]packformat.CheckFunc, error) {
	parts := []map[string]packformat.CheckFunc{{
		"retained_payload":     p.retained,
		"repository_selection": p.repositorySelection,
		"contract":             p.contract,
		"sequence_rollback":    p.sequenceRollback,
		"sequence_collision": func(_ context.Context, emit packformat.FindingSink) error {
			return identityFindings(p.sequenceErr, "sequence_collision", emit)
		},
		"member_missing": func(_ context.Context, emit packformat.FindingSink) error {
			return packformat.InventoryFindings("member_missing", p.manifest, p.inventory, true, emit)
		},
		"member_extra": func(_ context.Context, emit packformat.FindingSink) error {
			return packformat.InventoryFindings("member_extra", p.manifest, p.inventory, true, emit)
		},
		"member_length": p.memberLength,
		"member_hash":   p.memberHash,
		"logical_collision": func(ctx context.Context, emit packformat.FindingSink) error {
			return identityFindings(p.attempt.Identity.checkLogicalVersion(ctx, p.manifest, p.manifestSHA, p.payloadSHA), "logical_collision", emit)
		},
		"target_binding": func(_ context.Context, emit packformat.FindingSink) error {
			return packformat.TrustBindingFindings(p.manifest, p.manifestSHA, p.payloadSHA, p.proposal(), emit)
		},
		"content_schema": func(ctx context.Context, emit packformat.FindingSink) error {
			return packformat.ContentSchemaFindings(ctx, p.manifest, p.workspace, emit)
		},
		"disallowed_content": func(_ context.Context, emit packformat.FindingSink) error {
			return packformat.DisallowedContentFindings(p.manifest, emit)
		},
		"content_semantics": p.semantics,
		"runtime_compatibility": func(_ context.Context, emit packformat.FindingSink) error {
			return packformat.ManifestCompatibilityFindings(p.manifest, true, emit)
		},
		"type_registry": func(ctx context.Context, emit packformat.FindingSink) error {
			return packformat.RegistryCompatibilityFindings(ctx, p.manifest, p.workspace, emit)
		},
	}}
	for _, part := range []struct {
		first, last string
		build       func(context.Context) (map[string]packformat.CheckFunc, error)
	}{
		{"container_format", "archive_member_type", p.archiveChecks},
		{"bundle_shape", "bundle_canonical", p.hintChecks},
		{"tuf_schema", "target_hash", p.trustChecks},
		{"manifest_encoding", "manifest_canonical", p.manifestChecks},
		{"dependencies", "conflicts", func(context.Context) (map[string]packformat.CheckFunc, error) {
			return packformat.DependencyChecks(p.manifest, p.attempt.Identity.resolveDependency)
		}},
	} {
		program, err := packformat.DeferredChecks(part.first, part.last, part.build)
		if err != nil {
			return nil, err
		}
		parts = append(parts, program)
	}
	return packformat.ComposeChecks(parts...)
}

func (p *containerVerificationProgram) retained(ctx context.Context, emit packformat.FindingSink) error {
	retainedErr := p.attempt.Identity.checkRetainedPresence(ctx)
	if err := emitRetainedFindings(retainedErr, "retained_payload", emit); err != nil {
		return err
	}
	var source ContainerReader
	var err error
	if p.attempt.Staged != nil {
		source, p.size, err = p.storage.OpenStaged(ctx, *p.attempt.Staged)
	} else {
		source, p.size, err = p.storage.OpenPublished(ctx, *p.attempt.Retained)
	}
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if errors.Is(err, fs.ErrNotExist) || errors.Is(err, ErrArtifactUnavailable) {
			return emit(packformat.Finding{Path: "$.container_reference"})
		}
		return err
	}
	if source == nil {
		return errors.New("reference pack: invalid storage container result")
	}
	if p.size < 0 {
		return errors.Join(errors.New("reference pack: invalid storage container size"), source.Close())
	}
	p.source = &checkedContainerReader{ContainerReader: source}
	return nil
}

func (p *containerVerificationProgram) archiveChecks(ctx context.Context) (map[string]packformat.CheckFunc, error) {
	var err error
	p.workspace, err = p.storage.NewWorkspace(ctx)
	if err != nil {
		return nil, err
	}
	var checks map[string]packformat.CheckFunc
	checks, p.extract, err = packformat.ArchiveChecks(ctx, p.source, p.size, p.attempt.Limits)
	return checks, err
}

func (p *containerVerificationProgram) boundedMember(ctx context.Context, path string, maximum int64) ([]byte, int64, error) {
	member, ok := p.inventory[path]
	if !ok {
		return nil, -1, nil
	}
	if member.Size > maximum {
		return nil, member.Size, nil
	}
	data, err := readAttemptMember(ctx, p.workspace, p.inventory, path, maximum)
	return data, member.Size, err
}

func (p *containerVerificationProgram) hintChecks(ctx context.Context) (map[string]packformat.CheckFunc, error) {
	var err error
	p.inventory, err = p.extract(p.workspace)
	if err != nil {
		return nil, err
	}
	p.manifestBytes, p.manifestSize, err = p.boundedMember(ctx, "manifest.json", 1048576)
	if err != nil {
		return nil, err
	}
	p.key, p.version = packformat.SafeManifestIdentity(p.manifestBytes)
	if p.key != "" && p.attempt.IdentityAdmitted != nil {
		if err := p.attempt.IdentityAdmitted(ctx, p.key, p.version); err != nil {
			return nil, err
		}
	}
	hint, size, err := p.boundedMember(ctx, "bundle.json", 16384)
	if err != nil {
		return nil, err
	}
	checks, result := packformat.HintChecks(hint, size)
	canonical := checks["bundle_canonical"]
	checks["bundle_canonical"] = func(ctx context.Context, emit packformat.FindingSink) error {
		if err := canonical(ctx, emit); err != nil {
			return err
		}
		p.repository = result()
		return nil
	}
	return checks, nil
}

func (p *containerVerificationProgram) repositorySelection(ctx context.Context, emit packformat.FindingSink) error {
	for path := range p.inventory {
		p.paths = append(p.paths, path)
	}
	slices.Sort(p.paths)
	var ok bool
	p.trust, ok = p.attempt.Repositories[p.repository]
	if p.attempt.ResolveTrust != nil {
		versions := []int64{}
		for _, path := range p.paths {
			if strings.HasPrefix(path, "metadata/") && strings.HasSuffix(path, ".root.json") {
				v, err := strconv.ParseInt(strings.TrimSuffix(strings.TrimPrefix(path, "metadata/"), ".root.json"), 10, 64)
				if err == nil && v > 0 {
					versions = append(versions, v)
				}
			}
		}
		slices.Sort(versions)
		var err error
		p.trust, ok, err = p.attempt.ResolveTrust(ctx, p.repository, versions)
		if err != nil {
			var rejected *ContentRejection
			if errors.As(err, &rejected) && rejected.CheckID == "tuf_schema" {
				// Retained-root resource guards belong to metadata admission,
				// after repository selection. They never escape the registry.
				p.trustPreparationErr = err
				return nil
			}
			return err
		}
	}
	if !ok {
		return emit(packformat.Finding{Path: "$.trust_repository_id"})
	}
	return nil
}

func (p *containerVerificationProgram) trustChecks(ctx context.Context) (map[string]packformat.CheckFunc, error) {
	if p.trustPreparationErr != nil {
		checks, result := packformat.TrustChecks(nil, p.inventory, p.repository, p.attempt.Start.UTC(), p.trust)
		p.proposal = result
		checks["tuf_schema"] = func(_ context.Context, emit packformat.FindingSink) error {
			return identityFindings(p.trustPreparationErr, "tuf_schema", emit)
		}
		return checks, nil
	}
	metadata := map[string][]byte{}
	total, ordinal := int64(0), 0
	var guard *packformat.Finding
	for _, path := range p.paths {
		if !strings.HasPrefix(path, "metadata/") {
			continue
		}
		member := p.inventory[path]
		if member.Size > 2097152 || member.Size > 8388608-total {
			limit, maximum, observed := "max_metadata_total_bytes", int64(8388608), total+member.Size
			if member.Size > 2097152 {
				limit, maximum, observed = "max_metadata_file_bytes", 2097152, member.Size
			}
			finding := packformat.LimitFinding("$.metadata["+strconv.Itoa(ordinal)+"]", limit, maximum, observed)
			guard = &finding
			break
		}
		total += member.Size
		var err error
		metadata[path], err = readAttemptMember(ctx, p.workspace, p.inventory, path, 2097152)
		if err != nil {
			return nil, err
		}
		ordinal++
	}
	checks, result := packformat.TrustChecks(metadata, p.inventory, p.repository, p.attempt.Start.UTC(), p.trust)
	p.proposal = result
	if guard != nil {
		// The guard forbids further parsing; later hypothetical findings are
		// inapplicable. It still executes at the registered schema rank.
		checks["tuf_schema"] = func(_ context.Context, emit packformat.FindingSink) error { return emit(*guard) }
	}
	return checks, nil
}

func (p *containerVerificationProgram) manifestChecks(context.Context) (map[string]packformat.CheckFunc, error) {
	checks, result := packformat.ManifestChecks(p.manifestBytes, p.manifestSize, true)
	p.manifestResult = result
	return checks, nil
}

func (p *containerVerificationProgram) contract(_ context.Context, emit packformat.FindingSink) error {
	if p.manifestSize < 0 {
		return nil
	}
	p.manifest = p.manifestResult()
	p.manifestSHA = packformat.Digest(p.manifestBytes)
	var err error
	p.payloadSHA, err = packformat.PayloadDigest(p.manifest.Files)
	if err != nil {
		return errors.New("reference pack: admitted manifest payload identity failed")
	}
	return packformat.ManifestCompatibilityFindings(p.manifest, false, emit)
}

func (p *containerVerificationProgram) sequenceRollback(ctx context.Context, emit packformat.FindingSink) error {
	if p.manifestSize < 0 {
		return nil
	}
	p.sequenceErr = p.attempt.Identity.checkReleaseSequence(ctx, p.manifest, p.manifestSHA, p.payloadSHA)
	var rejection *ContentRejection
	if errors.As(p.sequenceErr, &rejection) && rejection.CheckID == "sequence_collision" {
		return nil
	}
	return identityFindings(p.sequenceErr, "sequence_rollback", emit)
}

func identityFindings(err error, check string, emit packformat.FindingSink) error {
	if err == nil {
		return nil
	}
	var rejection *ContentRejection
	if !errors.As(err, &rejection) {
		return err
	}
	if rejection.CheckID != check {
		return errors.New("reference pack: identity verdict at wrong check")
	}
	if rejection.Summary != nil {
		return emitRetainedFindings(err, check, emit)
	}
	return emit(packformat.Finding{Path: "$"})
}

func (p *containerVerificationProgram) memberLength(ctx context.Context, emit packformat.FindingSink) error {
	if err := packformat.InventoryFindings("member_length", p.manifest, p.inventory, true, emit); err != nil {
		return err
	}
	return emitRetainedFindings(p.attempt.Identity.checkRetainedLength(ctx), "member_length", emit)
}

func (p *containerVerificationProgram) memberHash(ctx context.Context, emit packformat.FindingSink) error {
	if err := packformat.InventoryFindings("member_hash", p.manifest, p.inventory, true, emit); err != nil {
		return err
	}
	if err := emitRetainedFindings(p.attempt.Identity.checkRetainedIntegrity(ctx), "member_hash", emit); err != nil {
		return err
	}
	hash := sha256.New()
	if _, err := io.CopyBuffer(hash, &attemptReader{ctx: ctx, source: io.NewSectionReader(p.source, 0, p.size)}, make([]byte, 32768)); err != nil {
		return err
	}
	p.containerSHA = hex.EncodeToString(hash.Sum(nil))
	if p.attempt.ContainerSHA256 != "" && p.containerSHA != p.attempt.ContainerSHA256 {
		return emit(packformat.Finding{Path: "$.container_sha256"})
	}
	return nil
}

func (p *containerVerificationProgram) semantics(ctx context.Context, emit packformat.FindingSink) error {
	if err := packformat.ManifestLicenseFindings(p.manifest, emit); err != nil {
		return err
	}
	for i, file := range p.manifest.Files {
		if file.Role != "notice" {
			continue
		}
		reader, err := p.workspace.Open(ctx, file.Path)
		if err != nil {
			return err
		}
		validationErr := packformat.ValidateNoticeStream(ctx, reader)
		if err := reader.Close(); err != nil {
			return err
		}
		if validationErr != nil {
			var failure *packformat.Failure
			if !errors.As(validationErr, &failure) {
				return validationErr
			}
			if err := emit(packformat.Finding{Path: "$.files[" + strconv.Itoa(i) + "]"}); err != nil {
				return err
			}
		}
	}
	return packformat.ContentSemanticsFindings(ctx, p.manifest, p.workspace, emit)
}
