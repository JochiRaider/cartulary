package packformat

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestOfflineTUFRootNestedAdmissionPrecedesCanonicalAndRotation_Unit(t *testing.T) {
	vector := trustVectors(t)[1]
	trust, err := AdmitBootstrap([]byte(vector.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	for _, boundary := range []string{"root", "binding", "keys", "key", "keyval", "roles", "role"} {
		value, err := canonicaljson.DecodeStrict([]byte(vector.Metadata["metadata/2.root.json"]))
		if err != nil {
			t.Fatal(err)
		}
		signed := value.(map[string]any)["signed"].(map[string]any)
		selectObject := func(signed map[string]any) map[string]any {
			keys := signed["keys"].(map[string]any)
			ids := make([]string, 0, len(keys))
			for id := range keys {
				ids = append(ids, id)
			}
			slices.Sort(ids)
			key := keys[ids[0]].(map[string]any)
			switch boundary {
			case "binding":
				return signed["cartulary"].(map[string]any)
			case "keys":
				return keys
			case "key":
				return key
			case "keyval":
				return key["keyval"].(map[string]any)
			case "roles":
				return signed["roles"].(map[string]any)
			case "role":
				return signed["roles"].(map[string]any)["root"].(map[string]any)
			default:
				return signed
			}
		}
		members := []string{}
		for name := range selectObject(signed) {
			members = append(members, name)
		}
		slices.Sort(members)
		if boundary == "root" {
			members = []string{"consistent_snapshot", "cartulary", "keys", "roles"}
		}
		for _, member := range members {
			// Dynamic key removal is a role-reference failure, not a schema
			// omission. Its value still has required closed object shape.
			mutations := []string{"null", "wrong type"}
			if boundary != "keys" {
				mutations = append(mutations, "omission")
			}
			for _, mutation := range mutations {
				t.Run(boundary+"/"+member+"/"+mutation, func(t *testing.T) {
					value, _ := canonicaljson.DecodeStrict([]byte(vector.Metadata["metadata/2.root.json"]))
					object := selectObject(value.(map[string]any)["signed"].(map[string]any))
					switch mutation {
					case "null":
						object[member] = nil
					case "wrong type":
						object[member] = []any{true}
					case "omission":
						delete(object, member)
					}
					data, _ := canonicaljson.Marshal(value)
					metadata := vectorMetadata(vector)
					delete(metadata, "metadata/2.root.json")
					// Both later faults are intentional. Shape must win.
					metadata["metadata/02.root.json"] = data
					metadata["metadata/targets.json"] = append(metadata["metadata/targets.json"], '\n')
					_, err := VerifyTrust(metadata, vector.Inventory, "test.repo", vector.VerificationTime, trust)
					var failure *Failure
					if !errors.As(err, &failure) || failure.Code != "tuf_metadata_invalid" {
						t.Fatalf("root shape precedence: %v", err)
					}
				})
			}
		}
		if boundary != "root" && boundary != "keys" {
			t.Run(boundary+"/unknown", func(t *testing.T) {
				object := selectObject(signed)
				object["hostile-name-must-not-escape"] = "payload-must-not-escape"
				data, _ := canonicaljson.Marshal(value)
				_, err := decodeMetadataShape(data, "root", 128)
				var failure *Failure
				if !errors.As(err, &failure) || failure.Code != "tuf_metadata_invalid" || strings.Contains(err.Error(), "must-not-escape") {
					t.Fatalf("closed root boundary: %v", err)
				}
			})
		}
	}
}
