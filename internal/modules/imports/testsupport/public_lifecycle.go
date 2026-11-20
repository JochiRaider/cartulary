package testsupport

import (
	"archive/zip"
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// The helpers below drive the public import lifecycle of an already running
// application over HTTP. They never start services, seed storage or bypass a
// route; every resource they return was admitted by its public operation.

// XLSXMediaType is the media type an XLSX upload declares.
const XLSXMediaType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// PostUpload submits createImportSession's multipart envelope. fileFirst
// reverses the part order to exercise envelope admission.
func PostUpload(t testing.TB, serverURL string, login authcookietest.AuthCookies, metadata string, file []byte, filename string, contentType string, fileFirst bool) *http.Response {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if fileFirst {
		writeFilePart(t, writer, filename, contentType, file)
		writeMetadataPart(t, writer, metadata)
	} else {
		writeMetadataPart(t, writer, metadata)
		writeFilePart(t, writer, filename, contentType, file)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, serverURL+"/api/v1/import-sessions", &body)
	if err != nil {
		t.Fatalf("build import request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	req.AddCookie(login.Session)
	req.AddCookie(login.CSRF)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post import upload: %v", err)
	}
	return resp
}

// DoJSON sends one authenticated, CSRF-bearing JSON request to path.
func DoJSON(t testing.TB, serverURL string, login authcookietest.AuthCookies, method string, path string, body any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(
		t, http.DefaultClient,
		method,
		serverURL+path,
		body,
		httptestx.WithCookies(login.Session, login.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value),
	)
}

// WaitJobTerminal observes an import Job until it is succeeded, failed or
// canceled, and returns its public projection.
func WaitJobTerminal(t testing.TB, serverURL string, login authcookietest.AuthCookies, jobID string) map[string]any {
	t.Helper()
	return httptestx.PollJob(t, serverURL, jobID, httptestx.PollOptions{
		Timeout: 5 * time.Second, Interval: 10 * time.Millisecond, Request: []func(*http.Request){httptestx.WithCookies(login.Session)},
	}, "succeeded", "failed", "canceled")
}

// StartCSVSession uploads one CSV, waits for discovery to succeed and returns
// the session and its only unit.
func StartCSVSession(t testing.TB, serverURL string, login authcookietest.AuthCookies, incidentID string, clientTxnID string, csv string, filename string) (string, string) {
	t.Helper()
	sessionID, unitIDs := startDiscoveredSession(t, serverURL, login, incidentID, clientTxnID, []byte(csv), filename, "text/csv")
	if len(unitIDs) != 1 {
		t.Fatalf("expected one import unit, got %d", len(unitIDs))
	}
	return sessionID, unitIDs[0]
}

// StartXLSXSession uploads MultipleSheetXLSX, waits for discovery to succeed
// and returns the session and its units in discovery (sheet) order.
func StartXLSXSession(t testing.TB, serverURL string, login authcookietest.AuthCookies, incidentID string, clientTxnID string, filename string) (string, []string) {
	t.Helper()
	return startDiscoveredSession(t, serverURL, login, incidentID, clientTxnID, MultipleSheetXLSX(t), filename, XLSXMediaType)
}

// PrepareReadyTimelineImport starts a one-column CSV session, approves a
// Timeline mapping and selects its unit, leaving the session ready to apply.
func PrepareReadyTimelineImport(t testing.TB, serverURL string, login authcookietest.AuthCookies, incidentID string, clientPrefix string) (string, string) {
	t.Helper()
	sessionID, unitID := StartCSVSession(
		t,
		serverURL,
		login,
		incidentID,
		"txn-"+clientPrefix+"-upload",
		"summary\nrecover exactly once\n",
		"recovery.csv",
	)
	mappingResp := DoJSON(
		t,
		serverURL,
		login,
		http.MethodPut,
		UnitRoute(sessionID, unitID, "mapping"),
		map[string]any{
			"client_txn_id":         "txn-" + clientPrefix + "-mapping",
			"target_view_schema_id": "cartulary.view.timeline.v2",
			"header_row_ref":        1,
			"data_start_row_ref":    2,
			"unknown_column_policy": "preserve_raw_capture",
			"source_columns": []map[string]any{{
				"source_column_ordinal": 1,
				"source_header_text":    "summary",
				"field_key":             "timeline.activity_synopsis_text",
				"entity_binding_mode":   nil,
				"transform_id":          nil,
				"transform_options":     map[string]any{},
				"empty_value_policy":    "omit_field",
			}},
		},
	)
	httptestx.RequireSuccessEnvelope(t, mappingResp, http.StatusOK)
	selectResp := DoJSON(
		t,
		serverURL,
		login,
		http.MethodPost,
		UnitRoute(sessionID, unitID, "select"),
		map[string]any{"client_txn_id": "txn-" + clientPrefix + "-unit-selection"},
	)
	httptestx.RequireSuccessEnvelope(t, selectResp, http.StatusOK)
	return sessionID, unitID
}

// UnitRoute addresses one import unit's action route, such as "mapping" or
// "select". Building it from separate segments keeps a route segment that is
// also an SQL keyword out of any composed string, so the boundary checker's
// SQL inference never reads a request path as SQL text.
func UnitRoute(sessionID string, unitID string, action string) string {
	return "/api/v1/import-sessions/" + sessionID + "/units/" + unitID + "/" + action
}

func startDiscoveredSession(t testing.TB, serverURL string, login authcookietest.AuthCookies, incidentID string, clientTxnID string, file []byte, filename string, contentType string) (string, []string) {
	t.Helper()
	metadata := `{"client_txn_id":"` + clientTxnID + `","incident_id":"` + incidentID + `"}`
	uploadResp := PostUpload(t, serverURL, login, metadata, file, filename, contentType, false)
	uploadJob := httptestx.RequireSuccessEnvelope(t, uploadResp, http.StatusAccepted)["data"].(map[string]any)
	job := WaitJobTerminal(t, serverURL, login, uploadJob["job_id"].(string))
	if job["status"] != "succeeded" {
		t.Fatalf("unexpected discovery job: %#v", job)
	}
	sessionID := job["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
	unitsResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, serverURL+"/api/v1/import-sessions/"+sessionID+"/units", nil, httptestx.WithCookies(login.Session))
	units := httptestx.RequireSuccessEnvelope(t, unitsResp, http.StatusOK)["data"].(map[string]any)["import_units"].([]any)
	unitIDs := make([]string, 0, len(units))
	for _, unit := range units {
		unitIDs = append(unitIDs, unit.(map[string]any)["import_unit_id"].(string))
	}
	return sessionID, unitIDs
}

func writeMetadataPart(t testing.TB, writer *multipart.Writer, metadata string) {
	t.Helper()
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="metadata"`},
		"Content-Type":        []string{"application/json"},
	})
	if err != nil {
		t.Fatalf("create metadata part: %v", err)
	}
	if _, err := io.WriteString(part, metadata); err != nil {
		t.Fatalf("write metadata part: %v", err)
	}
}

func writeFilePart(t testing.TB, writer *multipart.Writer, filename string, contentType string, file []byte) {
	t.Helper()
	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="file"; filename="` + filename + `"`},
		"Content-Type":        []string{contentType},
	})
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write(file); err != nil {
		t.Fatalf("write file part: %v", err)
	}
}

// MultipleSheetXLSX is a two-sheet workbook: Sheet1 holds a host/summary
// table with two data rows (A1:B3), Sheet2 an indicator/type table with one.
func MultipleSheetXLSX(t testing.TB) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	writeZipText(t, writer, "[Content_Types].xml", `<?xml version="1.0" encoding="UTF-8"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/worksheets/sheet2.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
</Types>`)
	writeZipText(t, writer, "xl/workbook.xml", `<?xml version="1.0" encoding="UTF-8"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets>
    <sheet name="Sheet1" sheetId="1" r:id="rId1"/>
    <sheet name="Sheet2" sheetId="2" r:id="rId2"/>
  </sheets>
</workbook>`)
	writeZipText(t, writer, "xl/_rels/workbook.xml.rels", `<?xml version="1.0" encoding="UTF-8"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet2.xml"/>
</Relationships>`)
	writeZipText(t, writer, "xl/worksheets/sheet1.xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>host</t></is></c><c r="B1" t="inlineStr"><is><t>summary</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>host-1</t></is></c><c r="B2" t="inlineStr"><is><t>Alpha summary</t></is></c></row>
    <row r="3"><c r="A3" t="inlineStr"><is><t>host-2</t></is></c><c r="B3" t="inlineStr"><is><t>Beta summary</t></is></c></row>
  </sheetData>
</worksheet>`)
	writeZipText(t, writer, "xl/worksheets/sheet2.xml", `<?xml version="1.0" encoding="UTF-8"?>
<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <sheetData>
    <row r="1"><c r="A1" t="inlineStr"><is><t>indicator</t></is></c><c r="B1" t="inlineStr"><is><t>type</t></is></c></row>
    <row r="2"><c r="A2" t="inlineStr"><is><t>203.0.113.42</t></is></c><c r="B2" t="inlineStr"><is><t>ipv4_addr</t></is></c></row>
  </sheetData>
</worksheet>`)
	if err := writer.Close(); err != nil {
		t.Fatalf("close XLSX zip: %v", err)
	}
	return buffer.Bytes()
}

func writeZipText(t testing.TB, writer *zip.Writer, name string, content string) {
	t.Helper()
	entry, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create XLSX entry %s: %v", name, err)
	}
	if _, err := io.WriteString(entry, content); err != nil {
		t.Fatalf("write XLSX entry %s: %v", name, err)
	}
}
