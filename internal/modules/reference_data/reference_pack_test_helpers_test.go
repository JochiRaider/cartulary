package reference_data_test

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func consumerIntegrityOptions(t testing.TB) reference_data.IntegrityOptions {
	t.Helper()
	finalizer, err := extensionassembly.NewReferencePackMutationFinalizer(func(err error) { t.Errorf("unexpected indeterminate integrity mutation: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	return reference_data.IntegrityOptions{Finalizer: finalizer, Limits: reference_data.DefaultLimits()}
}

type bundleOptions struct {
	PackKey           string
	PackKind          string
	PackVersion       string
	ContractVersion   string
	PayloadPath       string
	BadPayloadSHA     bool
	BadSignature      bool
	ExtraPath         string
	OmitPayload       bool
	PayloadTransform  func([]byte) []byte
	ObjectsTransform  func([]byte) []byte
	ManifestTransform func(map[string]any)
	MetadataVersion   int
	RootVersion       int
}

// The integration producer uses a deterministic test-only key and bounded
// relative expiries. Independent cryptographic expected values live in the
// separately authored signed-container and TUF vector contracts.
var integrationSigningDay = time.Now().UTC().Truncate(24 * time.Hour)

func integrationCanonical(t testing.TB, v any) []byte {
	t.Helper()
	b, err := canonicaljson.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
func integrationDigest(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
func integrationSigner(t testing.TB, ordinal int) (ed25519.PrivateKey, string, map[string]any) {
	seed := sha256.Sum256([]byte("cartulary reference pack integration test key only " + strconv.Itoa(ordinal)))
	key := ed25519.NewKeyFromSeed(seed[:])
	public := map[string]any{"keytype": "ed25519", "scheme": "ed25519", "keyval": map[string]any{"public": hex.EncodeToString(key.Public().(ed25519.PublicKey))}}
	return key, integrationDigest(integrationCanonical(t, public)), public
}
func integrationSigned(t testing.TB, signed map[string]any) []byte {
	count := 1
	if signed["_type"] == "root" {
		count = 3
	}
	signatures := []map[string]any{}
	for i := 0; i < count; i++ {
		key, id, _ := integrationSigner(t, i)
		signatures = append(signatures, map[string]any{"keyid": id, "sig": hex.EncodeToString(ed25519.Sign(key, integrationCanonical(t, signed)))})
	}
	slices.SortFunc(signatures, func(a, b map[string]any) int { return strings.Compare(a["keyid"].(string), b["keyid"].(string)) })
	return integrationCanonical(t, map[string]any{"signed": signed, "signatures": signatures})
}
func integrationRoot(t testing.TB) []byte {
	keys := map[string]any{}
	ids := []string{}
	for i := 0; i < 3; i++ {
		_, id, public := integrationSigner(t, i)
		keys[id] = public
		ids = append(ids, id)
	}
	_, online, _ := integrationSigner(t, 0)
	slices.Sort(ids)
	roles := map[string]any{"root": map[string]any{"keyids": ids, "threshold": 2}}
	for _, name := range []string{"targets", "snapshot", "timestamp"} {
		roles[name] = map[string]any{"keyids": []string{online}, "threshold": 1}
	}
	return integrationSigned(t, map[string]any{"_type": "root", "spec_version": "1.0.35", "version": 1, "expires": integrationSigningDay.Add(365 * 24 * time.Hour).Format(time.RFC3339), "consistent_snapshot": false, "keys": keys, "roles": roles, "cartulary": map[string]any{"schema_id": "cartulary.reference_pack_tuf_root_binding.v1", "trust_repository_id": "integration.repo"}})
}
func startReferencePackServer(t *testing.T, runtime *appsupport.Runtime, prefix string) *appsupport.ServerHarness {
	return startReferencePackServerWithEnv(t, runtime, prefix, nil)
}
func startReferencePackServerWithEnv(t *testing.T, runtime *appsupport.Runtime, prefix string, env map[string]string, barriers ...*appsupport.ReferencePackVerificationBarrier) *appsupport.ServerHarness {
	return startReferencePackServerConfigured(t, runtime, appsupport.ServerOptions{Prefix: prefix, TestRouteMode: httptestx.TestRouteModeDisabled, Env: env, ConfigureRuntime: func(o *server.Options) {
		if len(barriers) > 0 {
			o.ReferenceDataComposition = barriers[0]
		}
	}})
}
func startReferencePackServerConfigured(t *testing.T, runtime *appsupport.Runtime, options appsupport.ServerOptions) *appsupport.ServerHarness {
	env := options.Env
	root := filepath.Join(t.TempDir(), "bootstrap.json")
	b := integrationCanonical(t, map[string]any{"schema_id": "cartulary.reference_pack_trust_bootstrap.v1", "repositories": []any{map[string]any{"repository_id": "integration.repo", "trusted_root": json.RawMessage(integrationRoot(t)), "trusted_root_sha256": integrationDigest(integrationRoot(t))}}})
	if err := os.WriteFile(root, b, 0600); err != nil {
		t.Fatal(err)
	}
	if env == nil {
		env = map[string]string{}
	}
	env["CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH"] = root
	env["CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED"] = "true"
	options.Env = env
	return runtime.StartServer(t, options)
}
func referencePackBundle(t testing.TB, options bundleOptions) []byte {
	t.Helper()
	var release struct {
		Packs []struct {
			Binding struct {
				Key string `json:"pack_key"`
			} `json:"binding"`
			Members map[string]string `json:"members"`
		} `json:"packs"`
	}
	b, err := os.ReadFile("../../../contracts/reference-packs/builtins/release.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(b, &release); err != nil {
		t.Fatal(err)
	}
	members := map[string][]byte{}
	selected := release.Packs[0]
	for _, p := range release.Packs {
		if p.Binding.Key == options.PackKey {
			selected = p
		}
	}
	for path, data := range selected.Members {
		members[path] = []byte(data)
	}
	var manifest map[string]any
	if err = json.Unmarshal(members["manifest.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	var profiles struct {
		Cases []struct {
			Key     string            `json:"pack_key"`
			Members map[string]string `json:"members"`
		} `json:"cases"`
	}
	if !strings.HasPrefix(options.PackKey, "type_registry.") {
		b, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/profiles.v1.json")
		if err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(b, &profiles); err != nil {
			t.Fatal(err)
		}
		for path := range members {
			if strings.HasPrefix(path, "payload/") {
				delete(members, path)
			}
		}
		found := false
		for _, p := range profiles.Cases {
			if p.Key == options.PackKey {
				found = true
				for path, data := range p.Members {
					members[path] = []byte(data)
				}
			}
		}
		if !found {
			t.Fatalf("no canonical profile fixture for %s", options.PackKey)
		}
	}
	manifest["pack_key"] = options.PackKey
	manifest["pack_kind"] = strings.Split(options.PackKey, ".")[0]
	manifest["pack_version"] = options.PackVersion
	sequence, err := strconv.Atoi(options.PackVersion)
	if err != nil {
		sum := sha256.Sum256([]byte(options.PackVersion))
		sequence = int(sum[0]) + 100
	}
	manifest["pack_release_sequence"] = sequence + 1
	manifest["trust_repository_id"] = "integration.repo"
	manifest["source_as_of"] = integrationSigningDay.Format(time.RFC3339)
	manifest["built_at"] = integrationSigningDay.Format(time.RFC3339)
	manifest["content_profile_id"] = "cartulary.reference_pack." + options.PackKey + ".v1"
	if options.ContractVersion != "" {
		manifest["pack_contract_version"] = options.ContractVersion
	}
	files := []packformat.File{}
	if options.PayloadTransform != nil {
		members["payload/entries.ndjson"] = options.PayloadTransform(members["payload/entries.ndjson"])
	}
	if options.ObjectsTransform != nil {
		members["payload/objects.ndjson"] = options.ObjectsTransform(members["payload/objects.ndjson"])
	}
	counts := map[string]int{}
	paths := []string{}
	for path := range members {
		if path != "manifest.json" {
			paths = append(paths, path)
		}
	}
	slices.Sort(paths)
	for _, path := range paths {
		data := members[path]
		role, media := "payload", "application/x-ndjson"
		if strings.HasPrefix(path, "notices/") {
			role, media = "notice", "text/plain"
		} else {
			counts[path] = bytes.Count(data, []byte("\n"))
		}
		files = append(files, packformat.File{Path: path, Role: role, MediaType: media, Size: int64(len(data)), SHA256: integrationDigest(data)})
	}
	if strings.HasPrefix(options.PackKey, "framework.") {
		manifest["content_summary"] = map[string]any{"kind": "objects_relationships", "object_count": counts["payload/objects.ndjson"], "relationship_count": counts["payload/relationships.ndjson"]}
	} else {
		manifest["content_summary"] = map[string]any{"kind": "entries", "entry_count": counts["payload/entries.ndjson"]}
	}
	manifest["files"] = files
	if options.ManifestTransform != nil {
		options.ManifestTransform(manifest)
	}
	members["manifest.json"] = integrationCanonical(t, manifest)
	if options.ContractVersion == "" && options.ManifestTransform == nil {
		if _, err := packformat.DecodeManifest(members["manifest.json"], true); err != nil {
			t.Fatalf("invalid canonical producer manifest: %v", err)
		}
	}
	members["bundle.json"] = integrationCanonical(t, map[string]any{"schema_id": "cartulary.reference_pack_bundle_hint.v1", "trust_repository_id": "integration.repo"})
	payloadDigest, err := packformat.PayloadDigest(files)
	if err != nil {
		t.Fatal(err)
	}
	metadataVersion := options.MetadataVersion
	if metadataVersion == 0 {
		metadataVersion = 1
	}
	targets := map[string]any{}
	for path, data := range members {
		targets[path] = map[string]any{"length": len(data), "hashes": map[string]any{"sha256": integrationDigest(data)}}
	}
	binding := map[string]any{"schema_id": "cartulary.reference_pack_tuf_targets_binding.v1", "trust_repository_id": "integration.repo", "pack_key": options.PackKey, "pack_version": options.PackVersion, "pack_release_sequence": manifest["pack_release_sequence"], "manifest_sha256": integrationDigest(members["manifest.json"]), "payload_sha256": payloadDigest}
	members["metadata/targets.json"] = integrationSigned(t, map[string]any{"_type": "targets", "spec_version": "1.0.35", "version": metadataVersion, "expires": integrationSigningDay.Add(30 * 24 * time.Hour).Format(time.RFC3339), "targets": targets, "cartulary": binding})
	if options.BadSignature {
		var role map[string]any
		json.Unmarshal(members["metadata/targets.json"], &role)
		role["signatures"].([]any)[0].(map[string]any)["sig"] = strings.Repeat("0", 128)
		members["metadata/targets.json"] = integrationCanonical(t, role)
	}
	for _, link := range []struct {
		role, child string
		days        int
	}{{"snapshot", "targets", 14}, {"timestamp", "snapshot", 7}} {
		data := members["metadata/"+link.child+".json"]
		members["metadata/"+link.role+".json"] = integrationSigned(t, map[string]any{"_type": link.role, "spec_version": "1.0.35", "version": metadataVersion, "expires": integrationSigningDay.Add(time.Duration(link.days) * 24 * time.Hour).Format(time.RFC3339), "meta": map[string]any{link.child + ".json": map[string]any{"version": metadataVersion, "length": len(data), "hashes": map[string]any{"sha256": integrationDigest(data)}}}})
	}
	if options.RootVersion != 0 {
		if options.RootVersion != 2 {
			t.Fatal("integration producer supports only the consecutive root-2 rotation")
		}
		var root struct {
			Signed map[string]any `json:"signed"`
		}
		if err := json.Unmarshal(integrationRoot(t), &root); err != nil {
			t.Fatal(err)
		}
		root.Signed["version"] = 2
		members["metadata/2.root.json"] = integrationSigned(t, root.Signed)
	}
	if options.BadPayloadSHA {
		members["payload/entries.ndjson"] = []byte("{}\n")
	}
	if options.OmitPayload {
		delete(members, "payload/entries.ndjson")
	}
	if options.ExtraPath != "" {
		members[options.ExtraPath] = []byte("{}")
	}
	if options.PayloadPath != "" {
		members[options.PayloadPath] = members["payload/entries.ndjson"]
		delete(members, "payload/entries.ndjson")
	}
	paths = nil
	for path := range members {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, path := range paths {
		addZipFile(t, writer, path, members[path])
	}
	if err = writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func addZipFile(t testing.TB, writer *zip.Writer, name string, data []byte) {
	t.Helper()
	file, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create zip member %s: %v", name, err)
	}
	if _, err := file.Write(data); err != nil {
		t.Fatalf("write zip member %s: %v", name, err)
	}
}

func postReferencePackUpload(t testing.TB, baseURL string, login flowtest.LoginResult, metadata string, bundle []byte, filename string, contentType string, streaming ...bool) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	metadataHeader := textproto.MIMEHeader{}
	metadataHeader.Set("Content-Disposition", `form-data; name="metadata"`)
	metadataHeader.Set("Content-Type", "application/json")
	metadataPart, err := writer.CreatePart(metadataHeader)
	if err != nil {
		t.Fatalf("create metadata part: %v", err)
	}
	if _, err := io.WriteString(metadataPart, metadata); err != nil {
		t.Fatalf("write metadata part: %v", err)
	}
	fileHeader := textproto.MIMEHeader{}
	fileHeader.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	fileHeader.Set("Content-Type", contentType)
	filePart, err := writer.CreatePart(fileHeader)
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := filePart.Write(bundle); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	var source io.Reader = &body
	if len(streaming) > 0 && streaming[0] {
		source = struct{ io.Reader }{&body}
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/reference-packs/import", source)
	if err != nil {
		t.Fatalf("new upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)
	return httptestx.Do(t, http.DefaultClient, req)
}

func requireJob(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var job map[string]any
	for {
		job = requireJobNow(t, harness, login, jobID)
		switch job["status"] {
		case "succeeded", "failed", "canceled":
			return job
		}
		if time.Now().After(deadline) {
			t.Fatalf("job %s did not reach terminal state: %#v", jobID, job)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func requireJobNow(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}
