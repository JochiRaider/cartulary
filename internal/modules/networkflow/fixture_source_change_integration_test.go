package networkflow_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type publicationSourceChangeCase struct {
	Kind     string `json:"kind"`
	Pass     int    `json:"pass"`
	Source   string `json:"source,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// This adapter changes only the existing owner-facing ports, after the real
// approval. Core owns the original source capability, frozen plan and transaction.
type publicationSourceChange struct {
	ImportSourcePort
	mu             sync.Mutex
	active         *publicationSourceChangeCase
	replacement    []byte
	opens          int
	requestMutated bool
}

func (s *publicationSourceChange) arm(c *publicationSourceChangeCase, replacement []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active = c
	s.replacement = replacement
}
func (s *publicationSourceChange) OpenSourceStreamTx(ctx context.Context, tx pgx.Tx, ref string) (imports.ImportSourceStream, error) {
	stream, err := s.ImportSourcePort.OpenSourceStreamTx(ctx, tx, ref)
	if err != nil {
		return stream, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active == nil {
		return stream, nil
	}
	s.opens++
	if s.opens != s.active.Pass {
		return stream, nil
	}
	switch s.active.Kind {
	case "stream_ref":
		stream.SourceStreamRef += "changed"
	case "session_id":
		stream.ImportSessionID = uuid.MustParse("00000000-0000-4000-8000-000000000099")
	case "unit_id":
		stream.ImportUnitID = uuid.MustParse("00000000-0000-4000-8000-000000000099")
	case "capability_digest":
		stream.SourceContentSHA256 = strings.Repeat("a", 64)
	case "media_type":
		stream.SourceMediaType = "application/octet-stream"
	case "byte_size":
		stream.SourceByteSize++
	case "header", "content":
		if err := stream.Reader.Close(); err != nil {
			return stream, err
		}
		stream.Reader = io.NopCloser(bytes.NewReader(s.replacement))
	case "filename":
		stream.OriginalFilename = s.active.Filename
	}
	return stream, nil
}
func (s *publicationSourceChange) outcome() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{"kind": s.active.Kind, "pass": s.active.Pass, "apply_open_count": s.opens, "request_mutated": s.requestMutated}
}

type publicationChangedFacade struct {
	imports.ExtensionImportFacade
	source *publicationSourceChange
}

func (f publicationChangedFacade) ApplyImportUnitTx(ctx context.Context, tx pgx.Tx, request imports.ExtensionImportApplyRequest) (imports.ExtensionImportApplyResult, error) {
	f.source.mu.Lock()
	if c := f.source.active; c != nil {
		switch c.Kind {
		case "expected_digest":
			request.ExpectedSourceContentSHA256 = strings.Repeat("a", 64)
			f.source.requestMutated = true
		case "mapping_fingerprint":
			request.MappingFingerprint = strings.Repeat("a", 64)
			f.source.requestMutated = true
		}
	}
	f.source.mu.Unlock()
	return f.ExtensionImportFacade.ApplyImportUnitTx(ctx, tx, request)
}
func (c publicationComposition) WrapImportFacade(f imports.ExtensionImportFacade) imports.ExtensionImportFacade {
	return publicationChangedFacade{c.Controls.WrapImportFacade(f), c.source}
}
func (r *publicationFixtureRuntime) armSourceChange(bundle *networkflowfixtures.Bundle, c *publicationSourceChangeCase) error {
	var replacement []byte
	if c.Source != "" {
		var err error
		replacement, err = bundle.Read(c.Source)
		if err != nil {
			return err
		}
	}
	r.source.arm(c, replacement)
	return nil
}
func (r *publicationFixtureRuntime) sourceChangePrivacy(c *publicationSourceChangeCase) error {
	if c.Filename == "" {
		return nil
	}
	encoded, err := json.Marshal(r.transport)
	if err != nil {
		return err
	}
	forbidden, err := json.Marshal(c.Filename)
	if err != nil {
		return err
	}
	if bytes.Contains(encoded, forbidden) {
		return fmt.Errorf("source replacement filename leaked through a public result")
	}
	return nil
}
func TestNetworkFlowSourceChangedImportFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-023-import-facade-source-change", 3, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return publicationFixtureOutcomes(t, bundle)
	})
}
