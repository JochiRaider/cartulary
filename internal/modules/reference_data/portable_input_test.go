package reference_data

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

type portableInputVector struct {
	Container  string                         `json:"container_base64"`
	Descriptor packformat.PortableContainer   `json:"descriptor"`
	Reference  IncidentBundleVersionReference `json:"reference"`
}

type portableCloseFault struct {
	io.Reader
	err    error
	closed bool
}

func (r *portableCloseFault) Close() error { r.closed = true; return r.err }

type portableReadFault struct{ err error }

func (r portableReadFault) Read([]byte) (int, error) { return 0, r.err }

type portableCleanupFault struct {
	ArtifactStorage
	staged, published error
}

func (s portableCleanupFault) RemoveStaged(r StagingRef) error {
	_ = s.ArtifactStorage.RemoveStaged(r)
	return s.staged
}
func (s portableCleanupFault) RemovePublished(r StorageRef) error {
	_ = s.ArtifactStorage.RemovePublished(r)
	return s.published
}

func testPortableInputAdmission(t *testing.T) {
	t.Run("dependency scheduling", testPortableVerificationOrder)
	t.Run("shared parent execution deadline", testPortableExecutionDeadline)
	t.Helper()
	raw, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/portable-input.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Allowed, Restricted portableInputVector }
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	decode := func(v portableInputVector) []byte {
		t.Helper()
		data, err := base64.StdEncoding.DecodeString(v.Container)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	body := decode(fixture.Allowed)
	memory := &coordinatorMemoryStorage{objects: map[string][]byte{}}
	c := &verificationService{referenceDependencies: &referenceDependencies{storage: memory, limits: DefaultLimits()}}
	r := &incidentReferences{storage: memory, verifier: c}
	request := func(data []byte) IncidentReferenceImportRequest {
		return IncidentReferenceImportRequest{OpenContainer: func(_ context.Context, path string) (io.ReadCloser, error) {
			if path != "ext/reference_packs/containers/"+fixture.Allowed.Reference.ManifestSHA256+".pack" {
				t.Fatal("transport path did not derive from exact manifest", path)
			}
			return io.NopCloser(bytes.NewReader(data)), nil
		}}
	}
	initial := request(body)
	input, err := r.preparePortableContainer(context.Background(), initial, fixture.Allowed.Descriptor, fixture.Allowed.Reference)
	if err != nil || input == nil || input.identity.Key != fixture.Allowed.Reference.Key || input.identity.Manifest != nil || input.object.Digest != fixture.Allowed.Descriptor.ContainerSHA256 || input.object.Size != int64(len(body)) {
		t.Fatal("eligible private input", input, err)
	}
	if len(memory.objects) != 1 || len(memory.staged) != 0 {
		t.Fatal("private input retention", len(memory.objects), len(memory.staged))
	}
	_ = memory.RemovePublished(input.object.Reference)
	checkEmpty := func() {
		t.Helper()
		if len(memory.objects) != 0 || len(memory.staged) != 0 {
			t.Fatal("transport rejection leaked unretained bytes")
		}
	}
	for _, name := range []string{"short stream", "long stream", "digest mismatch", "tuple mismatch", "profile mismatch", "source mismatch"} {
		t.Run(name, func(t *testing.T) {
			descriptor, reference, data := fixture.Allowed.Descriptor, fixture.Allowed.Reference, body
			switch name {
			case "short stream":
				data = data[:len(data)-1]
			case "long stream":
				data = append(append([]byte{}, data...), 0)
			case "digest mismatch":
				descriptor.ContainerSHA256 = strings.Repeat("0", 64)
			case "tuple mismatch":
				reference.Version = "other.1"
			case "profile mismatch":
				reference.ProfileVersion = "2"
			case "source mismatch":
				reference.SourceProfileSHA256 = strings.Repeat("0", 64)
			}
			input, err := r.preparePortableContainer(context.Background(), request(data), descriptor, reference)
			var invalid *IncidentBundleReferenceValidationError
			if input != nil || !errors.As(err, &invalid) || invalid.InvariantID != IncidentBundleReferenceIdentityInvariant {
				t.Fatal("transport contradiction", input, err)
			}
			checkEmpty()
		})
	}
	t.Run("unattributable optional bytes confer no candidate or verdict", func(t *testing.T) {
		data := []byte("not an archive")
		descriptor := fixture.Allowed.Descriptor
		descriptor.SizeBytes = int64(len(data))
		descriptor.ContainerSHA256 = "6bbf954ab0045bc546f16a6db16c95afef820dccd807348411ea924dabb972e9"
		// Independent fixed SHA-256 of the literal source, checked by admission.
		input, err := r.preparePortableContainer(context.Background(), request(data), descriptor, fixture.Allowed.Reference)
		if input != nil || err != nil {
			t.Fatal("unattributable transport", input, err)
		}
		checkEmpty()
	})
	t.Run("restricted bytes are denied before trust", func(t *testing.T) {
		opened := false
		req := IncidentReferenceImportRequest{OpenContainer: func(context.Context, string) (io.ReadCloser, error) {
			opened = true
			return io.NopCloser(bytes.NewReader(decode(fixture.Restricted))), nil
		}}
		input, err := r.preparePortableContainer(context.Background(), req, fixture.Restricted.Descriptor, fixture.Restricted.Reference)
		if input != nil || err != nil || !opened {
			t.Fatal("redistribution transport eligibility", input, err)
		}
		checkEmpty()
	})
	t.Run("declared oversized container is not opened", func(t *testing.T) {
		descriptor := fixture.Allowed.Descriptor
		descriptor.SizeBytes = c.limits.ReferencePacks.MaxContainerBytes + 1
		req := IncidentReferenceImportRequest{OpenContainer: func(context.Context, string) (io.ReadCloser, error) {
			t.Fatal("oversized source opened")
			return nil, nil
		}}
		if input, err := r.preparePortableContainer(context.Background(), req, descriptor, fixture.Allowed.Reference); input != nil || err != nil {
			t.Fatal(input, err)
		}
	})
	sentinel := errors.New("source unavailable")
	for _, name := range []string{"close", "read", "staged cleanup", "published cleanup"} {
		t.Run(name, func(t *testing.T) {
			defer func() { r.storage = memory; c.storage = memory }()
			reader := &portableCloseFault{Reader: bytes.NewReader(body)}
			descriptor := fixture.Allowed.Descriptor
			switch name {
			case "close":
				reader.err = sentinel
			case "read":
				reader.Reader = io.MultiReader(bytes.NewReader(body), portableReadFault{sentinel})
			case "staged cleanup":
				r.storage = portableCleanupFault{ArtifactStorage: memory, staged: sentinel}
				c.storage = r.storage
			case "published cleanup":
				r.storage = portableCleanupFault{ArtifactStorage: memory, published: sentinel}
				c.storage = r.storage
				descriptor.ContainerSHA256 = strings.Repeat("0", 64)
			}
			req := IncidentReferenceImportRequest{OpenContainer: func(context.Context, string) (io.ReadCloser, error) { return reader, nil }}
			input, err := r.preparePortableContainer(context.Background(), req, descriptor, fixture.Allowed.Reference)
			var invalid *IncidentBundleReferenceValidationError
			if input != nil || !errors.Is(err, sentinel) || errors.As(err, &invalid) || !reader.closed {
				t.Fatal("I/O failure became a transport or content verdict", input, err, reader.closed)
			}
			checkEmpty()
		})
	}
}
