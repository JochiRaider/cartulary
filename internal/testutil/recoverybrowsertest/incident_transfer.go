package recoverybrowsertest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func portableNetworkFlowEnvironment(env map[string]string) {
	env["CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED"] = "true"
	env["CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH"] = fixtures.Path("network-flow", "key-rings.json")
	env["CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR"] = "AwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwM"
	env["CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST"] = "BAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ"
}

// importBrowserIncident streams the caller's exported archive through the real
// incident-import route. The file is confined to the private browser runtime.
func importBrowserIncident(ctx context.Context, client *http.Client, origin, csrf, runtimeRoot, bundleName string) (string, error) {
	root, err := os.OpenRoot(runtimeRoot)
	if err != nil {
		return "", err
	}
	defer root.Close()
	file, err := root.Open(bundleName)
	if err != nil {
		return "", err
	}
	defer file.Close()
	input, output := io.Pipe()
	writer := multipart.NewWriter(output)
	completed := make(chan error, 1)
	go func() {
		var writeErr error
		defer func() { _ = output.CloseWithError(writeErr); completed <- writeErr }()
		part, err := writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="metadata"`}, "Content-Type": {"application/json"}})
		if err != nil {
			writeErr = err
			return
		}
		if err := json.NewEncoder(part).Encode(map[string]string{"client_txn_id": "browser-incident-transfer"}); err != nil {
			writeErr = err
			return
		}
		part, err = writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="file"; filename="incident.zip"`}, "Content-Type": {"application/zip"}})
		if err != nil {
			writeErr = err
			return
		}
		if _, err := io.Copy(part, file); err != nil {
			writeErr = err
			return
		}
		writeErr = writer.Close()
	}()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, origin+"/api/v1/incident-bundles/import", input)
	if err != nil {
		_ = input.CloseWithError(err)
		return "", errors.Join(err, <-completed)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set(authn.CSRFHeaderName, csrf)
	response, err := client.Do(request)
	_ = input.CloseWithError(err)
	writeErr := <-completed
	if err != nil {
		return "", errors.Join(err, writeErr)
	}
	defer response.Body.Close()
	if writeErr != nil {
		return "", writeErr
	}
	if response.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("browser transfer admission returned %d", response.StatusCode)
	}
	var envelope struct {
		Data struct {
			JobID string `json:"job_id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return "", err
	}
	if envelope.Data.JobID == "" {
		return "", errors.New("browser transfer omitted its job")
	}
	for {
		payload, err := doSourceJSON(ctx, client, http.MethodGet, origin+"/api/v1/jobs/"+envelope.Data.JobID, "", nil)
		if err != nil {
			return "", err
		}
		status, err := stringAt(payload, "data", "status")
		if err != nil {
			return "", err
		}
		if status == "failed" || status == "canceled" {
			return "", fmt.Errorf("browser transfer ended as %s", status)
		}
		if status == "succeeded" {
			var job struct {
				Data struct {
					Summary struct {
						Refs []struct {
							Kind string `json:"kind"`
							ID   string `json:"id"`
						} `json:"resource_refs"`
					} `json:"result_summary"`
				} `json:"data"`
			}
			encoded, err := json.Marshal(payload)
			if err != nil {
				return "", err
			}
			if err := json.Unmarshal(encoded, &job); err != nil {
				return "", err
			}
			for _, ref := range job.Data.Summary.Refs {
				if ref.Kind == "incident" && ref.ID != "" {
					return ref.ID, nil
				}
			}
			return "", errors.New("browser transfer omitted incident result")
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}
