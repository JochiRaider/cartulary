package reference_data_test

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func testOversizedUploadAdmission(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	h := startReferencePackServerWithEnv(t, runtime, "reference-pack-stream-limit", map[string]string{"CARTULARY__LIMITS__REFERENCE_PACKS__MAX_CONTAINER_BYTES": "128"})
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, h.Server.HTTP.URL)
	tables := []string{"jobs", "reference_pack_operations", "reference_pack_candidates", "reference_pack_events", "reference_pack_attempt_members", "reference_pack_objects", "reference_pack_indexes"}
	before := map[string]int{}
	for _, table := range tables {
		before[table] = queryCount(t, h.DB, "SELECT count(*) FROM "+table)
	}
	for _, streaming := range []bool{false, true} {
		t.Run(fmt.Sprintf("streaming=%v", streaming), func(t *testing.T) {
			response := postReferencePackUpload(t, h.Server.HTTP.URL, admin, fmt.Sprintf(`{"client_txn_id":"oversized-%v"}`, streaming), []byte(strings.Repeat("x", 1024)), "bounded.zip", reference_data.MediaTypeZip, streaming)
			envelope := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "reference_pack_verification_failed")
			details := envelope["error"].(map[string]any)["details"].(map[string]any)
			if len(details) != 3 || details["reason_code"] != "container_bytes_exceeded" || details["check_id"] != "container_bytes" {
				t.Fatal(details)
			}
			summary := details["validation_summary"].(map[string]any)
			// Literal SHA-256 independently calculated over the specified issue preimage.
			const issueID = "rpi_66fc39b6c5aebe40032bce0a8403bc7c6bfca8006258b564eaa15b9d352271eb"
			if summary["primary_issue_id"] != issueID || summary["total_issue_count"] != float64(1) || summary["retained_issue_count"] != float64(1) || summary["issues_truncated"] != false || summary["result"] != "failed" {
				t.Fatal(summary)
			}
			issues := summary["issues"].([]any)
			if len(issues) != 1 {
				t.Fatal(issues)
			}
			issue := issues[0].(map[string]any)
			if issue["issue_id"] != issueID || issue["path"] != "$" || issue["phase"] != "archive_preflight" || issue["code"] != "container_bytes_exceeded" || issue["reason_code"] != "container_bytes_exceeded" || issue["entry_id"] != nil {
				t.Fatal(issue)
			}
			safe := issue["safe_details"].(map[string]any)
			if safe["limit_id"] != "max_container_bytes" || safe["expected_token"] != "128" || safe["actual_token"] != "129" || safe["related_pack_key"] != nil || safe["related_pack_version"] != nil {
				t.Fatal(safe)
			}
		})
	}
	for _, table := range tables {
		if got := queryCount(t, h.DB, "SELECT count(*) FROM "+table); got != before[table] {
			t.Errorf("%s changed from %d to %d", table, before[table], got)
		}
	}
	root := filepath.Join(h.Server.Config.Roots.TemporaryWork.Path, "reference-packs/imports")
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			t.Errorf("staging file retained: %s", path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
