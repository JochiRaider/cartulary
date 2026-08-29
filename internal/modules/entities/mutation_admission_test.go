package entities

import (
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mutationadmission"
)

func TestMutationAdmissionContractAndHTTPTranslation(t *testing.T) {
	t.Parallel()

	for index := range reflect.TypeOf(mutationadmission.Failure{}).NumField() {
		field := reflect.TypeOf(mutationadmission.Failure{}).Field(index)
		if field.PkgPath == "" {
			t.Fatalf("mutation admission state field %q is exported", field.Name)
		}
	}

	limit := mutationadmission.NewLimit(
		"changes",
		mutationadmission.ReasonChangeCountExceeded,
		33,
		32,
		"host.aliases",
	)
	apiErr := mutationAdmissionAPIError(limit)
	if apiErr.Status != 400 || apiErr.Code != "invalid_mutation_payload" ||
		apiErr.Message != "invalid mutation payload" ||
		apiErr.Details["field"] != "changes" ||
		apiErr.Details["reason_code"] != "change_count_exceeded" ||
		apiErr.Details["requested_count"] != 33 ||
		apiErr.Details["max_count"] != 32 ||
		apiErr.Details["field_key"] != "host.aliases" {
		t.Fatalf("translated admission failure = %#v", apiErr)
	}

	validConflictClaims := hostidentity.WorkbookConflictClaims{
		RecordID:          uuid.MustParse("00000000-0000-4000-8000-000000000001"),
		ViewSchemaID:      entitycontract.HostsViewSchemaID,
		FieldKey:          "host.display_name",
		CurrentRowVersion: 1,
	}
	decoders := []struct {
		name   string
		decode func(string) *mutationadmission.Failure
	}{
		{name: "create", decode: func(body string) *mutationadmission.Failure {
			_, failure := hostidentity.DecodeCreateRequest(entitycontract.HostsViewSchemaID, strings.NewReader(body))
			return failure
		}},
		{name: "patch", decode: func(body string) *mutationadmission.Failure {
			_, failure := hostidentity.DecodePatchRequest(strings.NewReader(body))
			return failure
		}},
		{name: "clipboard", decode: func(body string) *mutationadmission.Failure {
			_, failure := hostidentity.DecodeClipboardPasteRequest(strings.NewReader(body), entitycontract.HostsViewSchemaID)
			return failure
		}},
		{name: "conflict", decode: func(body string) *mutationadmission.Failure {
			_, failure := hostidentity.DecodeWorkbookConflictResolveRequest(
				strings.NewReader(body),
				"opaque",
				validConflictClaims,
			)
			return failure
		}},
		{name: "merge", decode: func(body string) *mutationadmission.Failure {
			_, failure := merge.DecodeMergeRequest(strings.NewReader(body))
			return failure
		}},
		{name: "mention", decode: func(body string) *mutationadmission.Failure {
			_, failure := mentions.DecodeMentionActionRequest(strings.NewReader(body))
			return failure
		}},
	}
	ambiguousBodies := map[string]string{
		"malformed":        `{`,
		"scalar":           `7`,
		"null":             `null`,
		"duplicate member": `{"member":1,"member":2}`,
		"nested duplicate": `{"member":{"nested":1,"nested":2}}`,
		"trailing value":   `{} {}`,
	}
	for _, decoder := range decoders {
		decoder := decoder
		t.Run(decoder.name, func(t *testing.T) {
			t.Parallel()
			for name, body := range ambiguousBodies {
				failure := decoder.decode(body)
				if failure == nil || failure.ReasonCode() != mutationadmission.ReasonRequestNotObject {
					t.Fatalf("%s body failure = %#v", name, failure)
				}
				if _, ok := failure.Field(); ok {
					t.Fatalf("%s body unexpectedly attributed a field", name)
				}
			}
		})
	}

	defer func() {
		if recover() == nil {
			t.Fatal("unregistered reason did not fail closed")
		}
	}()
	mutationadmission.New("payload", mutationadmission.ReasonCode("unregistered_reason"))
}
