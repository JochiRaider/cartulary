//go:build linux

package artifactstore

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// downloadAttempts bounds how often a release archive is requested before a
// network failure is reported; a digest mismatch is never retried. Each retry
// waits downloadRetryDelay times its attempt.
const downloadAttempts = 3

var downloadRetryDelay = 2 * time.Second

// verifiedArchive returns the archive name kept in directory when it has the
// committed digest, downloading it from url first when it is missing or
// altered. An archive with another digest is removed and never kept, so it
// can never be extracted.
func verifiedArchive(directory, name, url, digest string) (string, error) {
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}
	archive := filepath.Join(directory, name)
	if got, err := fileDigest(archive); err == nil && got == digest {
		return archive, nil
	} else if err == nil || !errors.Is(err, os.ErrNotExist) {
		if err := os.Remove(archive); err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}
	var err error
	for attempt := 1; attempt <= downloadAttempts; attempt++ {
		var retry bool
		if retry, err = download(directory, name, url, digest); err == nil || !retry {
			break
		}
		if attempt < downloadAttempts {
			time.Sleep(time.Duration(attempt) * downloadRetryDelay)
		}
	}
	if err != nil {
		return "", err
	}
	return archive, nil
}

// download writes url to directory/name when its bytes have digest. It
// reports whether a failure was the network's and may be retried.
func download(directory, name, url, digest string) (bool, error) {
	temporary, err := os.CreateTemp(directory, "."+name+".download-")
	if err != nil {
		return false, err
	}
	defer os.Remove(temporary.Name())
	client := http.Client{Timeout: 5 * time.Minute}
	response, err := client.Get(url)
	if err != nil {
		temporary.Close()
		return true, fmt.Errorf("download %s: %w", name, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		temporary.Close()
		return response.StatusCode >= 500, fmt.Errorf("download %s: %s", name, response.Status)
	}
	sum := sha256.New()
	_, err = io.Copy(io.MultiWriter(temporary, sum), response.Body)
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return true, fmt.Errorf("download %s: %w", name, err)
	}
	if got := hex.EncodeToString(sum.Sum(nil)); got != digest {
		return false, fmt.Errorf("archive %s has digest %s, not the committed %s; it was not kept", name, got, digest)
	}
	return false, os.Rename(temporary.Name(), filepath.Join(directory, name))
}
