package networkflow

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"testing"
)

func TestTableLifecycleNameControlPrecedence_Unit(t *testing.T) {
	for _, value := range []string{"\t", "\n", "\u0085", " \t "} {
		_, err := normalizeExplicitDisplayName(value)
		var invalid *invalidDisplayNameError
		if !errors.As(err, &invalid) || invalid.ReasonCode != "forbidden_control" {
			t.Errorf("name %q: got %v, want forbidden_control", value, err)
		}
	}
}

func TestTableLifecycleMutationDigest_Unit(t *testing.T) {
	id := "nft_11111111111111111111111111111111"
	request := tableRenameRequest{ClientTxnID: "one", BaseTableVersion: 3, DisplayName: "\u00a0Cafe\u0301\u3000"}
	want := sha256.Sum256([]byte("cartulary.network_flow.mutation_request_digest.v1\x00nf.tables.patch\x00network_flow_table_id:" + id + "\x00{\"base_table_version\":3,\"display_name\":\"Café\"}\x00"))
	if got := tableRenameRequestHash(id, request); !bytes.Equal(got, want[:]) {
		t.Errorf("rename digest differs from normalized route/path transcript")
	}
	request.ClientTxnID = "two"
	request.DisplayName = "Café"
	if got := tableRenameRequestHash(id, request); !bytes.Equal(got, want[:]) {
		t.Errorf("canonical equivalents or transaction ID changed semantic digest")
	}
	deleted := sha256.Sum256([]byte("cartulary.network_flow.mutation_request_digest.v1\x00nf.tables.delete\x00network_flow_table_id:" + id + "\x00{\"base_table_version\":3}\x00"))
	if got := tableSoftDeleteRequestHash(id, tableSoftDeleteRequest{ClientTxnID: "delete", BaseTableVersion: 3}); !bytes.Equal(got, deleted[:]) {
		t.Errorf("delete digest differs from route/path transcript")
	}
}
