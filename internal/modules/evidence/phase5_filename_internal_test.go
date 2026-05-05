package evidence

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestPhase5_SanitizeFilenameRemovesNUL_U_5_07(t *testing.T) {
	recordID := uuid.New()
	got := sanitizeFilename("evil\x00name.txt", recordID, "text/plain")
	if got != "evilname.txt" {
		t.Fatalf("sanitizeFilename with NUL got %q want evilname.txt", got)
	}
	fallback := sanitizeFilename("..\\..", recordID, "application/pdf")
	want := "evidence-" + recordID.String() + ".pdf"
	if fallback != want {
		t.Fatalf("sanitizeFilename fallback got %q want %q", fallback, want)
	}
	header := formatContentDisposition("attachment", "résumé.txt")
	if !strings.Contains(header, "filename=") || !strings.Contains(header, "filename*=UTF-8''") {
		t.Fatalf("Content-Disposition did not include filename and filename*: %q", header)
	}
}
