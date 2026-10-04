package reference_data

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

// VerificationAttempt is frozen application input. Start is the successful
// verification instant, never admission or completion time. The coordinator
// supplies the same instant and deadline to every member of a refresh cohort.
type VerificationAttempt struct {
	Observer         OperationObserver
	Staged           *StagingRef
	Retained         *StorageRef
	ContainerSHA256  string
	Start            time.Time
	ClockTrusted     bool
	Limits           packformat.ArchiveLimits
	Repositories     map[string]packformat.TrustSnapshot
	ResolveTrust     func(context.Context, string, []int64) (packformat.TrustSnapshot, bool, error)
	IdentityAdmitted func(context.Context, string, string) error
	Identity         verificationIdentityHistory
}

// A live attempt checks retained availability, lengths, hashes and persisted
// logical identity at their distinct registry ranks. Historical restore
// supplies the exact retained envelope binding after validating its objects.
// Omission is an implementation error, never permission to skip these checks.
type verificationIdentityHistory interface {
	checkRetainedPresence(context.Context) error
	checkRetainedLength(context.Context) error
	checkRetainedIntegrity(context.Context) error
	checkReleaseSequence(context.Context, packformat.Manifest, string, string) error
	checkLogicalVersion(context.Context, packformat.Manifest, string, string) error
	resolveDependency(context.Context, packformat.Dependency) (packformat.Manifest, bool, error)
}

// VerifiedContent owns its private workspace until the coordinator has copied
// its immutable members and index into unpublished durable storage. Close must
// be called on both a committed result and a proven uncommitted result.
type VerifiedContent struct {
	Manifest        packformat.Manifest
	ManifestBytes   []byte
	ManifestSHA256  string
	PayloadSHA256   string
	ContainerSHA256 string
	ContainerBytes  int64
	Inventory       packformat.Inventory
	Trust           packformat.TrustProposal
	TrustSnapshot   packformat.TrustSnapshot
	VerifiedAt      time.Time
	workspace       VerificationWorkspace
}

func (v *VerifiedContent) Close() error {
	if v == nil || v.workspace == nil {
		return nil
	}
	return v.workspace.Close()
}

type OperationRejection struct{ Reason string }

func (e *OperationRejection) Error() string { return "reference pack operation rejected: " + e.Reason }

type ContentRejection struct {
	Code             string
	CheckID          string
	CandidateKey     string
	CandidateVersion string
	Summary          *packformat.ValidationSummary
}

func (e *ContentRejection) Error() string { return "reference pack verification failed: " + e.Code }

// VerifyCanonicalContainer is the application verification boundary shared by
// imports, renewal, reverify, refresh and destination portability. It performs
// no trust mutation or consumer publication and makes no network request.
func VerifyCanonicalContainer(ctx context.Context, storage VerificationStorage, attempt VerificationAttempt, index func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error)) (result *VerifiedContent, resultErr error) {
	ctx, end := observeReferenceOperation(ctx, attempt.Observer, "reference_pack.verify")
	defer func() { end(referenceOutcome(resultErr)) }()
	defer func() {
		if resultErr != nil && result != nil {
			_ = result.Close()
			result = nil
		}
	}()
	if storage == nil || index == nil || attempt.Identity == nil || attempt.Start.IsZero() || (attempt.Staged == nil) == (attempt.Retained == nil) {
		return nil, errors.New("reference pack: incomplete verification attempt")
	}
	if !attempt.ClockTrusted {
		return nil, &OperationRejection{Reason: "clock_untrusted"}
	}
	p := &containerVerificationProgram{attempt: attempt, storage: storage}
	defer func() {
		if p.source != nil {
			if err := errors.Join(p.source.fault, p.source.Close()); err != nil {
				resultErr = err
			}
		}
	}()
	keep := false
	defer func() {
		if !keep && p.workspace != nil {
			if err := p.workspace.Close(); err != nil {
				resultErr = err
			}
		}
	}()
	scratch := &verificationDiagnosticScratch{storage: storage}
	defer func() {
		if err := scratch.close(); err != nil {
			resultErr = err
		}
	}()
	program, err := p.checks()
	if err != nil {
		return nil, err
	}
	summary, err := packformat.RunChecks(ctx, scratch, program)
	if err != nil {
		var misplaced *ContentRejection
		if errors.As(err, &misplaced) {
			return nil, errors.New("reference pack: content verdict bypassed check registry")
		}
		return nil, err
	}
	if summary.Result == "failed" {
		issue := summary.Issues[0]
		return nil, &ContentRejection{Code: issue.Code, CheckID: issue.CheckID, CandidateKey: p.key, CandidateVersion: p.version, Summary: summary}
	}
	sink, err := index(ctx, p.manifest, p.manifestSHA, p.payloadSHA)
	if err != nil {
		return nil, err
	}
	if sink == nil {
		return nil, errors.New("reference pack: missing private index writer")
	}
	if err := packformat.ValidateContentSemantics(ctx, p.manifest, p.workspace, sink); err != nil {
		// Changed private bytes or an index write failure are operational.
		return nil, errors.New("reference pack: validated index construction failed")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	keep = true
	return &VerifiedContent{Manifest: p.manifest, ManifestBytes: p.manifestBytes, ManifestSHA256: p.manifestSHA, PayloadSHA256: p.payloadSHA, ContainerSHA256: p.containerSHA, ContainerBytes: p.size, Inventory: p.inventory, Trust: p.proposal(), TrustSnapshot: p.trust, VerifiedAt: attempt.Start.UTC(), workspace: p.workspace}, nil
}

type attemptReader struct {
	ctx    context.Context
	source io.Reader
}

func (r *attemptReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.source.Read(p)
}

func readAttemptMember(ctx context.Context, source VerificationWorkspace, inventory packformat.Inventory, path string, maximum int64) ([]byte, error) {
	member, ok := inventory[path]
	if !ok {
		return nil, &packformat.Failure{Code: "required_member_missing"}
	}
	if member.Size > maximum {
		code := "archive_structure_invalid"
		switch {
		case path == "manifest.json":
			code = "manifest_schema_invalid"
		case path == "bundle.json":
			code = "bundle_hint_invalid"
		case strings.HasPrefix(path, "metadata/"):
			code = "tuf_metadata_invalid"
		}
		return nil, &packformat.Failure{Code: code}
	}
	reader, err := source.Open(ctx, path)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(io.LimitReader(reader, maximum+1))
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil {
		return nil, errors.Join(readErr, closeErr)
	}
	if int64(len(data)) != member.Size || packformat.Digest(data) != member.SHA256 {
		return nil, errors.New("reference pack: private workspace integrity failure")
	}
	return data, nil
}
