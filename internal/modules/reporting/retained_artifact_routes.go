package reporting

import (
	"context"
	"crypto/sha256"
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
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
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
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incident, err := uuid.Parse(r.PathValue("incident_id"))
	if err != nil || incident == uuid.Nil || incident.String() != r.PathValue("incident_id") {
		writeAPIError(w, r, &httpapi.APIError{Status: 404, Code: "incident_not_found", Details: map[string]any{}})
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	tx, members, apiErr := s.app.beginRetainedArtifactRead(r.Context(), incident, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	defer rollbackSnapshotTx(tx)
	requested := r.PathValue("artifact_id")
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
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
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
		writeAPIError(w, r, &httpapi.APIError{Status: 404, Code: "report_artifact_not_found", Details: map[string]any{}})
		return
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{"schema_id": "cartulary.retained_report_artifacts.v1", "incident_id": incident.String(), "artifacts": artifacts})
}
