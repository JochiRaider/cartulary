package workbookassembly

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestEntityAdmissionFailuresHaveOneWorkbookTranslation(t *testing.T) {
	t.Parallel()

	create, err := newEntityCreateProvider(
		entitycontract.HostsViewSchemaID,
		func(
			context.Context,
			authn.UserRecord,
			uuid.UUID,
			hostidentity.CreateRequest,
			[]byte,
			string,
			time.Time,
		) (hostidentity.MutationResult, error) {
			panic("strictly rejected input reached create")
		},
	)
	if err != nil {
		t.Fatalf("compose create provider: %v", err)
	}
	_, createFailure, err := create.DecodeCreate(strings.NewReader(`null`))
	if err != nil {
		t.Fatalf("decode create: %v", err)
	}
	assertEntityAdmissionFailure(t, createFailure)

	owner := &hostidentity.Store{}
	patch, err := newEntityPatchProvider("host", entitycontract.HostsViewSchemaID, owner)
	if err != nil {
		t.Fatalf("compose patch provider: %v", err)
	}
	_, patchFailure, err := patch.DecodePatch(strings.NewReader(`{} {}`))
	if err != nil {
		t.Fatalf("decode patch: %v", err)
	}
	assertEntityAdmissionFailure(t, patchFailure)

	clipboard, err := newEntityClipboardProvider(entitycontract.HostsViewSchemaID, owner)
	if err != nil {
		t.Fatalf("compose clipboard provider: %v", err)
	}
	_, clipboardFailure, err := clipboard.DecodeClipboard(strings.NewReader(`{"member":1,"member":2}`))
	if err != nil {
		t.Fatalf("decode clipboard: %v", err)
	}
	assertEntityAdmissionFailure(t, clipboardFailure)

	conflict, err := newEntityConflictProvider("host", entitycontract.HostsViewSchemaID, owner)
	if err != nil {
		t.Fatalf("compose conflict provider: %v", err)
	}
	_, conflictFailure, err := conflict.DecodeConflict(
		strings.NewReader(`7`),
		"opaque",
		workbook.ConflictClaims{
			RecordID:          uuid.MustParse("00000000-0000-4000-8000-000000000001"),
			ViewSchemaID:      entitycontract.HostsViewSchemaID,
			RouteKey:          workbookConflictResolveOperation,
			FieldKey:          "host.display_name",
			CurrentRowVersion: 1,
		},
	)
	if err != nil {
		t.Fatalf("decode conflict: %v", err)
	}
	assertEntityAdmissionFailure(t, conflictFailure)
}

func assertEntityAdmissionFailure(t testing.TB, failure *workbook.MutationFailure) {
	t.Helper()
	if failure == nil {
		t.Fatal("Workbook admission failure is nil")
	}
	field, reason, ok := failure.InvalidPayloadDetail()
	if !ok || field != "" || reason != "request_not_object" {
		t.Fatalf("Workbook admission failure = (%q, %q, %t)", field, reason, ok)
	}
}
