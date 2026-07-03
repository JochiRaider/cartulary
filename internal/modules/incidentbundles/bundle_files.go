package incidentbundles

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type bundleFileStore interface {
	stageBundle(fileSHA string, data []byte) (string, error)
	persistBundle(bundleID string, data []byte) (string, error)
	remove(path string)
}

type filesystemBundleFileStore struct {
	temporaryRoot string
	exportRoot    string
}

func newBundleFileStore(temporaryRoot string, exportRoot string) bundleFileStore {
	return filesystemBundleFileStore{temporaryRoot: temporaryRoot, exportRoot: exportRoot}
}

func (s filesystemBundleFileStore) stageBundle(fileSHA string, data []byte) (string, error) {
	root := s.temporaryRoot
	if strings.TrimSpace(root) == "" {
		root = os.TempDir()
	}
	bundleDir := filepath.Join(root, "incident-bundles", "imports")
	if err := os.MkdirAll(bundleDir, 0o700); err != nil {
		return "", err
	}
	name := fileSHA
	if strings.TrimSpace(name) == "" {
		name = uuid.NewString()
	}
	path := filepath.Join(bundleDir, name+"-"+uuid.NewString()+".bundle")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func (s filesystemBundleFileStore) persistBundle(bundleID string, data []byte) (string, error) {
	root := s.exportRoot
	if strings.TrimSpace(root) == "" {
		root = os.TempDir()
	}
	bundleDir := filepath.Join(root, "incident-bundles")
	if err := os.MkdirAll(bundleDir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(bundleDir, bundleID+".zip")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

func (s filesystemBundleFileStore) remove(path string) {
	_ = os.Remove(path)
}
