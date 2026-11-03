package extensioncodec

import (
	"bufio"
	"bytes"
	"container/heap"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"sort"
	"strconv"
)

const (
	proofChunkEntries = 4096
	proofChunkBytes   = 4 << 20
	proofMergeFanIn   = 32
)

type Proof struct {
	ItemCount     int64
	ContentBytes  int64
	BindingDigest string
}

type Entries func(yield func(Entry) error) error

// ComputeProof consumes one entry at a time and sorts only bounded chunks of
// identity/hash/length records. Its confined spool contains no content bytes.
// The item and identity bounds cap disk use; merging opens at most 32 inputs
// and temporarily uses at most twice the complete metadata spool size.
// scratch is a borrowed capability; this function removes only its private
// child directory and does not close the caller's root.
func ComputeProof(ctx context.Context, bindingID, storageKind string, limits Limits, scratch *os.Root, entries Entries) (proof Proof, resultErr error) {
	if ctx == nil || scratch == nil || entries == nil || !validLogicalID(bindingID) || !limits.valid() ||
		(storageKind != "postgres" && storageKind != "object_store" && storageKind != "filesystem") {
		return Proof{}, ErrInvalidContract
	}
	if err := ctx.Err(); err != nil {
		return Proof{}, err
	}
	directory := "extension-proof-" + rand.Text()
	if err := scratch.Mkdir(directory, 0700); err != nil {
		return Proof{}, err
	}
	work, err := scratch.OpenRoot(directory)
	if err != nil {
		_ = scratch.Remove(directory)
		return Proof{}, err
	}
	defer func() {
		cleanupErr := errors.Join(work.Close(), scratch.RemoveAll(directory))
		if cleanupErr != nil {
			proof = Proof{}
			resultErr = errors.Join(resultErr, cleanupErr)
		}
	}()
	parts := []string{}
	chunk := make([]proofRecord, 0, min(int(limits.MaxItems), proofChunkEntries))
	chunkBytes, sequence := 0, 0
	var itemCount, contentBytes int64
	writePart := func(emit func(func(proofRecord) error) error) (string, error) {
		name := fmt.Sprintf("part-%08d", sequence)
		sequence++
		file, err := work.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if err != nil {
			return "", err
		}
		writer := bufio.NewWriterSize(file, 65536)
		emitErr := emit(func(record proofRecord) error { return writeProofRecord(writer, record) })
		if err := errors.Join(emitErr, writer.Flush(), file.Close()); err != nil {
			return "", err
		}
		return name, nil
	}
	flush := func() error {
		if len(chunk) == 0 {
			return nil
		}
		sort.Slice(chunk, func(i, j int) bool { return bytes.Compare(chunk[i].identity, chunk[j].identity) < 0 })
		name, err := writePart(func(yield func(proofRecord) error) error {
			for i, record := range chunk {
				if err := ctx.Err(); err != nil {
					return err
				}
				if i > 0 && bytes.Equal(chunk[i-1].identity, record.identity) {
					return ErrDuplicateIdentity
				}
				if err := yield(record); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
		parts = append(parts, name)
		clear(chunk)
		chunk = chunk[:0]
		chunkBytes = 0
		return nil
	}
	admit := func(entry Entry) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(entry.Identity) == 0 {
			return ErrInvalidEntry
		}
		length := int64(len(entry.Content))
		if itemCount >= limits.MaxItems || len(entry.Identity) > 4096 || length > limits.MaxEntryBytes || length > limits.MaxBindingBytes-contentBytes {
			return ErrLimitExceeded
		}
		itemCount++
		contentBytes += length
		chunk = append(chunk, proofRecord{identity: append([]byte(nil), entry.Identity...), digest: sha256.Sum256(entry.Content), length: length})
		chunkBytes += len(entry.Identity) + 48
		if len(chunk) >= proofChunkEntries || chunkBytes >= proofChunkBytes {
			return flush()
		}
		return nil
	}
	var admissionErr error
	streamErr := entries(func(entry Entry) error {
		if admissionErr == nil {
			admissionErr = admit(entry)
		}
		return admissionErr
	})
	if err := errors.Join(streamErr, admissionErr); err != nil {
		return Proof{}, err
	}
	if err := flush(); err != nil {
		return Proof{}, err
	}
	for len(parts) > proofMergeFanIn {
		next := make([]string, 0, (len(parts)+proofMergeFanIn-1)/proofMergeFanIn)
		for start := 0; start < len(parts); start += proofMergeFanIn {
			group := parts[start:min(start+proofMergeFanIn, len(parts))]
			name, err := writePart(func(yield func(proofRecord) error) error { return mergeProofRecords(ctx, work, group, yield) })
			if err != nil {
				return Proof{}, err
			}
			next = append(next, name)
			for _, old := range group {
				if err := work.Remove(old); err != nil {
					return Proof{}, err
				}
			}
		}
		parts = next
	}
	digest := sha256.New()
	_, _ = digest.Write([]byte("cartulary.extension.binding.v1\x00"))
	hashProofFrame(digest, []byte(bindingID))
	hashProofFrame(digest, []byte(storageKind))
	hashProofFrame(digest, []byte(strconv.FormatInt(itemCount, 10)))
	var checkedItems, checkedBytes int64
	if err := mergeProofRecords(ctx, work, parts, func(record proofRecord) error {
		if record.length > limits.MaxEntryBytes || record.length > limits.MaxBindingBytes-checkedBytes {
			return ErrLimitExceeded
		}
		checkedItems++
		checkedBytes += record.length
		hashProofFrame(digest, record.identity)
		hashProofFrame(digest, record.digest[:])
		hashProofFrame(digest, []byte(strconv.FormatInt(record.length, 10)))
		return nil
	}); err != nil {
		return Proof{}, err
	}
	if err := ctx.Err(); err != nil {
		return Proof{}, err
	}
	if checkedItems != itemCount || checkedBytes != contentBytes {
		return Proof{}, ErrInvalidEntry
	}
	return Proof{ItemCount: itemCount, ContentBytes: contentBytes, BindingDigest: hex.EncodeToString(digest.Sum(nil))}, nil
}

func hashProofFrame(digest hash.Hash, value []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(value)))
	_, _ = digest.Write(length[:])
	_, _ = digest.Write(value)
}

type proofRecord struct {
	identity []byte
	digest   [32]byte
	length   int64
}

func writeProofRecord(writer io.Writer, record proofRecord) error {
	var scalar [8]byte
	binary.BigEndian.PutUint64(scalar[:], uint64(len(record.identity)))
	for _, part := range [][]byte{scalar[:], record.identity, record.digest[:]} {
		if _, err := writer.Write(part); err != nil {
			return err
		}
	}
	binary.BigEndian.PutUint64(scalar[:], uint64(record.length))
	_, err := writer.Write(scalar[:])
	return err
}

func readProofRecord(reader io.Reader) (proofRecord, error) {
	var record proofRecord
	var scalar [8]byte
	if _, err := io.ReadFull(reader, scalar[:]); err != nil {
		return record, err
	}
	length := binary.BigEndian.Uint64(scalar[:])
	if length == 0 || length > 4096 {
		return record, ErrInvalidEntry
	}
	record.identity = make([]byte, int(length))
	for _, part := range [][]byte{record.identity, record.digest[:], scalar[:]} {
		if _, err := io.ReadFull(reader, part); err != nil {
			return proofRecord{}, err
		}
	}
	contentLength := binary.BigEndian.Uint64(scalar[:])
	if contentLength > 1<<63-1 {
		return proofRecord{}, ErrInvalidEntry
	}
	record.length = int64(contentLength)
	return record, nil
}

type proofCursor struct {
	reader *bufio.Reader
	record proofRecord
}

type proofHeap []*proofCursor

func (h proofHeap) Len() int { return len(h) }
func (h proofHeap) Less(i, j int) bool {
	return bytes.Compare(h[i].record.identity, h[j].record.identity) < 0
}
func (h proofHeap) Swap(i, j int)   { h[i], h[j] = h[j], h[i] }
func (h *proofHeap) Push(value any) { *h = append(*h, value.(*proofCursor)) }
func (h *proofHeap) Pop() any {
	old := *h
	value := old[len(old)-1]
	*h = old[:len(old)-1]
	return value
}

func mergeProofRecords(ctx context.Context, root *os.Root, names []string, yield func(proofRecord) error) error {
	if len(names) > proofMergeFanIn {
		return ErrLimitExceeded
	}
	files := make([]*os.File, 0, len(names))
	defer func() {
		for _, file := range files {
			_ = file.Close()
		}
	}()
	queue := proofHeap{}
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return err
		}
		file, err := root.Open(name)
		if err != nil {
			return err
		}
		files = append(files, file)
		cursor := &proofCursor{reader: bufio.NewReaderSize(file, 65536)}
		cursor.record, err = readProofRecord(cursor.reader)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			return err
		}
		heap.Push(&queue, cursor)
	}
	var previous []byte
	for queue.Len() > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		cursor := heap.Pop(&queue).(*proofCursor)
		if bytes.Equal(previous, cursor.record.identity) {
			return ErrDuplicateIdentity
		}
		previous = append(previous[:0], cursor.record.identity...)
		if err := yield(cursor.record); err != nil {
			return err
		}
		record, err := readProofRecord(cursor.reader)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			return err
		}
		cursor.record = record
		heap.Push(&queue, cursor)
	}
	return nil
}
