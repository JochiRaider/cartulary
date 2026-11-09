package publication

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
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

// ReserveRelease serializes append-only logical release sequence reservations.
// Aborted builds occupy their reservation. Only identical explicit inputs replay.
func ReserveRelease(directory, version, inputSHA string, scope Scope) (Publication, error) {
	schema, repository, key := "cartulary.reference_pack_release_reservation.v1", scope.repository, scope.key
	if !scope.valid() {
		return Publication{}, errors.New("invalid publication scope")
	}
	if directory == "" || version == "" || !digestPattern.MatchString(inputSHA) {
		return Publication{}, errors.New("invalid publication request")
	}
	root, err := openLedgerNamespace(directory, "logical")
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
	entries, err := readRootDirectory(root)
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
		b, err := readRootRegular(root, entry.Name(), 65536)
		if err != nil {
			return Publication{}, err
		}
		var p Publication
		if err := DecodeClosed(b, &p); err != nil {
			return Publication{}, err
		}
		canonical, err := canonicaljson.Marshal(p)
		if err != nil || !bytes.Equal(b, canonical) || p.SchemaID != schema || p.Repository != repository || p.Key != key || p.Sequence != highest+1 || entry.Name() != fmt.Sprintf("%016d.json", p.Sequence) || !digestPattern.MatchString(p.InputSHA256) {
			return Publication{}, errors.New("invalid publication ledger chain")
		}
		highest = p.Sequence
		if p.Version == version {
			if replay != nil || p.InputSHA256 != inputSHA {
				return Publication{}, errors.New("publication identity collision")
			}
			copy := p
			replay = &copy
		}
	}
	if replay != nil {
		return *replay, nil
	}
	if highest >= 9007199254740991 {
		return Publication{}, errors.New("release sequence exhausted")
	}
	p := Publication{schema, repository, key, version, highest + 1, inputSHA}
	b, err := canonicaljson.Marshal(p)
	if err != nil {
		return Publication{}, err
	}
	if err := Retain(root, fmt.Sprintf("%016d.json", p.Sequence), b, 0600); err != nil {
		return Publication{}, err
	}
	return p, nil
}

// Logical release sequences and envelope metadata sequences have different
// keys and lifetimes. Their closed namespaces share one publication root.
// A flat predecessor ledger is rejected, never interpreted as an empty ledger.
func openLedgerNamespace(directory, name string) (*os.Root, error) {
	if directory == "" || name != "logical" && name != "envelopes" {
		return nil, errors.New("explicit publication ledger namespace required")
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	entries, err := readRootDirectory(root)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() != "logical" && entry.Name() != "envelopes" {
			return nil, errors.New("incompatible publication ledger layout")
		}
	}
	if err = root.MkdirAll(name, 0700); err != nil {
		return nil, err
	}
	return root.OpenRoot(name)
}
