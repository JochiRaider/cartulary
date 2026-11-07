package producer

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"golang.org/x/sys/unix"
)

type Publication struct {
	SchemaID    string `json:"schema_id"`
	Repository  string `json:"trust_repository_id"`
	Key         string `json:"pack_key"`
	Version     string `json:"pack_version"`
	Sequence    int64  `json:"pack_release_sequence"`
	InputSHA256 string `json:"input_sha256"`
}

// allocatePublication serializes reservations and publishes immutable records.
// A failed build leaves its reservation occupied; replay must bind identical
// explicit inputs. There is no overwrite or deployment-identity allocator.
func allocatePublication(directory, version, inputSHA string) (Publication, error) {
	if directory == "" || version == "" || !digestPattern.MatchString(inputSHA) {
		return Publication{}, errors.New("invalid fixture publication request")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return Publication{}, err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return Publication{}, err
	}
	defer root.Close()
	lock, err := root.OpenFile(".lock", os.O_CREATE|os.O_RDWR|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return Publication{}, err
	}
	defer lock.Close()
	if err := unix.Flock(int(lock.Fd()), unix.LOCK_EX); err != nil {
		return Publication{}, err
	}
	defer unix.Flock(int(lock.Fd()), unix.LOCK_UN) //nolint:errcheck // closing also releases the lock
	entries, err := os.ReadDir(directory)
	if err != nil {
		return Publication{}, err
	}
	slices.SortFunc(entries, func(a, b os.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
	var highest int64
	var replay *Publication
	for _, entry := range entries {
		if entry.Name() == ".lock" {
			continue
		}
		if entry.Type() != 0 || !strings.HasSuffix(entry.Name(), ".json") {
			return Publication{}, errors.New("invalid publication ledger member")
		}
		b, err := root.ReadFile(entry.Name())
		if err != nil {
			return Publication{}, err
		}
		var p Publication
		if err := DecodeClosed(b, &p); err != nil {
			return Publication{}, err
		}
		canonical, err := canonicaljson.Marshal(p)
		if err != nil || !bytes.Equal(b, canonical) || p.SchemaID != "cartulary.attack_fixture_publication.v1" || p.Repository != FixtureRepository || p.Key != "framework.attack" || p.Sequence != highest+1 || entry.Name() != fmt.Sprintf("%016d.json", p.Sequence) || !digestPattern.MatchString(p.InputSHA256) {
			return Publication{}, errors.New("invalid publication ledger chain")
		}
		highest = p.Sequence
		if p.Version == version {
			if replay != nil || p.InputSHA256 != inputSHA {
				return Publication{}, errors.New("fixture publication identity collision")
			}
			copy := p
			replay = &copy
		}
	}
	if replay != nil {
		return *replay, nil
	}
	if highest >= 9007199254740991 {
		return Publication{}, errors.New("fixture sequence exhausted")
	}
	p := Publication{"cartulary.attack_fixture_publication.v1", FixtureRepository, "framework.attack", version, highest + 1, inputSHA}
	b, err := canonicaljson.Marshal(p)
	if err != nil {
		return Publication{}, err
	}
	if err := Retain(root, fmt.Sprintf("%016d.json", p.Sequence), b, 0600); err != nil {
		return Publication{}, err
	}
	return p, nil
}

func readRegular(path string) ([]byte, error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Lstat(filepath.Base(path))
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("regular producer input required")
	}
	return root.ReadFile(filepath.Base(path))
}
