package reporting

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type retainedReportArtifact struct {
	ID          string  `json:"artifact_id"`
	Kind        string  `json:"kind"`
	Filename    string  `json:"filename"`
	ByteSize    int64   `json:"byte_size"`
	SHA256      string  `json:"sha256"`
	SnapshotID  *string `json:"snapshot_id"`
	ReleaseID   *string `json:"release_id"`
	ContentHref string  `json:"content_href"`
}

func retainedArtifactResource(incident uuid.UUID, member incidentportability.Member) (retainedReportArtifact, error) {
	id := fmt.Sprintf("%x", sha256.Sum256([]byte("cartulary.retained_report_artifact.v1\x00"+incident.String()+"\x00"+member.Path)))
	result := retainedReportArtifact{ID: id, Filename: path.Base(member.Path), ByteSize: member.ByteSize, SHA256: member.SHA256, ContentHref: "/api/v1/incidents/" + incident.String() + "/retained-report-artifacts/" + id + "/content"}
	parts := strings.Split(member.Path, "/")
	switch {
	case member.Path == portableArtifactCatalogPath:
		result.Kind = "catalog"
	case len(parts) == 4 && parts[0] == "ext" && parts[1] == "snapshots" && canonicalUUID(parts[2]) && parts[3] == "source-capture.json":
		result.Kind = "source_capture"
		result.SnapshotID = &parts[2]
	case len(parts) >= 5 && parts[0] == "ext" && parts[1] == "snapshots" && parts[2] == "releases" && canonicalUUID(parts[3]):
		result.ReleaseID = &parts[3]
		if len(parts) == 5 && parts[4] == "manifest.json" {
			result.Kind = "release_manifest"
		} else if len(parts) >= 6 && parts[4] == "files" {
			result.Kind = "release_file"
		} else {
			return result, ErrNotFound
		}
	default:
		return result, ErrNotFound
	}
	return result, nil
}

// All reads are bound to the same authorization and immutable member snapshot.
func (s *ApplicationService) beginRetainedArtifactRead(ctx context.Context, incident, actor uuid.UUID) (pgx.Tx, *retainedArtifactMembers, *httpapi.APIError) {
	tx, err := s.store.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, nil, internalAPIError(err)
	}
	grant, err := s.incidentAccess.CheckReadTx(ctx, tx, incident, actor, admission.Requirement{AllowedRoles: admission.RolesEditorReviewerAdmin, Lifecycle: admission.LifecycleAny})
	if _, apiErr := reportingAdmissionResult(grant, err, "editor|reviewer|admin"); apiErr != nil {
		rollbackSnapshotTx(tx)
		return nil, nil, apiErr
	}
	members, err := loadImportedArtifactMembers(ctx, tx, incident)
	if err != nil {
		rollbackSnapshotTx(tx)
		return nil, nil, internalAPIError(err)
	}
	return tx, members, nil
}

func (s *Service) handleRetainedReportArtifacts(w http.ResponseWriter, r *http.Request) {
	s.serveRetainedReportArtifacts(w, r, "")
}

func (s *Service) handleRetainedReportArtifactContent(w http.ResponseWriter, r *http.Request) {
	artifactID, identified := httpapi.PathString(r, "artifact_id")
	if !identified {
		writeAPIError(w, r, internalAPIError(errors.New("reporting route was not admitted centrally")))
		return
	}
	s.serveRetainedReportArtifacts(w, r, artifactID)
}

// serveRetainedReportArtifacts lists the incident's retained artifacts, or
// streams the requested one. Central admission has authenticated the caller
// and admitted incident_id and artifact_id (Core 01 REQ-01-707).
func (s *Service) serveRetainedReportArtifacts(w http.ResponseWriter, r *http.Request, requested string) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incident, apiErr := admittedPathUUID(r, "incident_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	tx, members, apiErr := s.app.beginRetainedArtifactRead(r.Context(), incident, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	defer rollbackSnapshotTx(tx)
	// REQ-01-707 step 5: the singleton-read query check is route handling,
	// after the incident's visibility and role gate.
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	artifacts := []retainedReportArtifact{}
	for _, member := range members.Members() {
		resource, err := retainedArtifactResource(incident, member)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if requested == "" {
			artifacts = append(artifacts, resource)
			continue
		}
		if resource.ID != requested {
			continue
		}
		reader, err := incidentportability.OpenVerifiedMember(r.Context(), members, member)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		defer reader.Close()
		if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": resource.Filename}))
		w.Header().Set("Content-Length", strconv.FormatInt(member.ByteSize, 10))
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if _, err := io.CopyBuffer(w, reader, make([]byte, 32<<10)); err != nil {
			panic(http.ErrAbortHandler)
		}
		return
	}
	if requested != "" {
		writeAPIError(w, r, httpapi.NotFound("report_artifact_not_found"))
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{"schema_id": "cartulary.retained_report_artifacts.v1", "incident_id": incident.String(), "artifacts": artifacts})
}
