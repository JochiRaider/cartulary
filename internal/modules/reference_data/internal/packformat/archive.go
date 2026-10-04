// Package packformat admits canonical Reference Pack bytes. It has no database,
// network, activation, or incident dependencies.
package packformat

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash/crc32"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

type Failure struct {
	Code    string
	CheckID string
	Summary *ValidationSummary
	Details SafeDetails
}

func (e *Failure) Error() string { return "reference pack: " + e.Code }
func fail(code string) error     { return &Failure{Code: code} }
func limitFailure(code, id string, maximum, observed uint64) error {
	return &Failure{Code: code, Details: unsignedLimitFinding("$", id, maximum, observed).Details}
}

var ErrStorage = errors.New("reference pack: extraction storage unavailable")

type ArchiveLimits struct {
	ContainerBytes   int64
	ExtractedBytes   int64
	CompressionRatio int64
	Members          int64
}

func DefaultArchiveLimits() ArchiveLimits {
	return ArchiveLimits{ContainerBytes: 536870912, ExtractedBytes: 536870912, CompressionRatio: 100, Members: 10000}
}

// Destination is a fresh, private, root-confined extraction workspace. The
// owner removes the entire workspace on failure; no member is published here.
type Destination interface {
	Write(context.Context, string, func(io.Writer) error) error
}
type Member struct {
	Path   string
	Size   int64
	SHA256 string
}
type Inventory map[string]Member

type archiveAdmission struct {
	ctx         context.Context
	limits      ArchiveLimits
	sourceBytes int64
	compressed  bool
	extracted   int64
	directories int64
	records     []archiveRecord
	expected    []archiveRecord
	regulars    int64
	members     Inventory
	destination Destination
}

// scan admits framing and bounded decompression only. Original names are
// retained as inert data for the ordered path/type checks; preflight writes no
// filesystem member. A second scan can extract only the exact admitted list.
func (a *archiveAdmission) scan(source io.ReaderAt, size int64, format string) error {
	var err error
	switch format {
	case "zip":
		a.compressed = true
		err = a.zip(source, size)
	case "gzip":
		a.compressed = true
		input := bufio.NewReader(io.NewSectionReader(source, 0, size))
		var stream *gzip.Reader
		stream, err = gzip.NewReader(input)
		if err == nil {
			stream.Multistream(false)
			err = a.tar(stream)
			closeErr := stream.Close()
			if err == nil {
				err = closeErr
			}
			if err == nil {
				if _, end := input.ReadByte(); !errors.Is(end, io.EOF) {
					err = fail("archive_structure_invalid")
				}
			}
		}
	case "tar":
		err = a.tar(io.NewSectionReader(source, 0, size))
	default:
		return errors.New("reference pack: unadmitted archive format")
	}
	if err != nil {
		if canceled := a.ctx.Err(); canceled != nil {
			return canceled
		}
		var failure *Failure
		if errors.As(err, &failure) || errors.Is(err, ErrStorage) {
			return err
		}
		return fail("archive_structure_invalid")
	}
	if a.expected != nil && len(a.records) != len(a.expected) {
		return ErrStorage
	}
	return nil
}

// ValidatePath never cleans, case-folds, or otherwise repairs an admitted name.
func ValidatePath(name string, directory bool) error {
	if len(name) == 0 || len(name) > 1024 {
		return fail("archive_structure_invalid")
	}
	for _, b := range []byte(name) {
		if b == 0 || b > 127 || b == '\\' {
			return fail("path_traversal")
		}
	}
	if directory {
		if !strings.HasSuffix(name, "/") {
			return fail("archive_structure_invalid")
		}
		name = strings.TrimSuffix(name, "/")
	} else if strings.HasSuffix(name, "/") {
		return fail("archive_structure_invalid")
	}
	for _, segment := range strings.Split(name, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fail("path_traversal")
		}
		if len(segment) > 255 {
			return fail("archive_structure_invalid")
		}
	}
	return nil
}

func (a *archiveAdmission) record(name string, directory, regular bool) error {
	if err := a.ctx.Err(); err != nil {
		return err
	}
	// No conforming stream can exceed its regular-file allowance plus its
	// independent directory allowance. Also bounds inventories of forbidden
	// headers without admitting them as regular files.
	if int64(len(a.records)) >= 10000 && int64(len(a.records))-10000 >= a.limits.Members {
		return fail("archive_structure_invalid")
	}
	if directory && a.directories >= 10000 {
		return limitFailure("archive_structure_invalid", "max_directory_markers", 10000, uint64(a.directories)+1)
	}
	if a.expected != nil {
		if len(a.records) >= len(a.expected) {
			return ErrStorage
		}
		expected := a.expected[len(a.records)]
		if name != expected.name || directory != expected.directory || regular != expected.regular {
			return ErrStorage
		}
	}
	a.records = append(a.records, archiveRecord{name: name, directory: directory, regular: regular})
	if directory {
		a.directories++
	}
	return nil
}

func (a *archiveAdmission) reserve(size int64) error {
	if size < 0 {
		return fail("archive_structure_invalid")
	}
	if size > 268435456 {
		return limitFailure("archive_structure_invalid", "max_member_bytes", 268435456, uint64(size))
	}
	if a.regulars >= a.limits.Members {
		return limitFailure("archive_member_count_exceeded", "max_members", uint64(a.limits.Members), uint64(a.regulars)+1)
	}
	if size > a.limits.ExtractedBytes-a.extracted {
		return limitFailure("archive_extracted_bytes_exceeded", "max_extracted_bytes", uint64(a.limits.ExtractedBytes), uint64(a.extracted)+uint64(size))
	}
	if a.compressed && a.sourceBytes <= math.MaxInt64/a.limits.CompressionRatio && a.extracted+size > a.sourceBytes*a.limits.CompressionRatio {
		return limitFailure("archive_compression_ratio_exceeded", "max_compression_ratio", uint64(a.sourceBytes)*uint64(a.limits.CompressionRatio), uint64(a.extracted)+uint64(size))
	}
	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.reader.Read(p)
}

type storageWriter struct{ writer io.Writer }

func (w storageWriter) Write(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if err != nil || n != len(p) {
		return n, ErrStorage
	}
	return n, nil
}

func (a *archiveAdmission) member(name string, size int64, source io.Reader) error {
	if err := a.reserve(size); err != nil {
		return err
	}
	index := len(a.records) - 1
	if index < 0 {
		return errors.New("reference pack: member without header")
	}
	if a.expected != nil && size != a.expected[index].member.Size {
		return ErrStorage
	}
	hash := sha256.New()
	var written int64
	var copyErr error
	copyTo := func(file io.Writer) error {
		written, copyErr = io.CopyBuffer(io.MultiWriter(storageWriter{file}, hash), contextReader{a.ctx, io.LimitReader(source, size)}, make([]byte, 32768))
		return copyErr
	}
	if a.expected == nil {
		if err := copyTo(io.Discard); err != nil {
			return err
		}
	} else if err := a.destination.Write(a.ctx, name, copyTo); err != nil {
		if copyErr != nil {
			return copyErr
		}
		return ErrStorage
	}
	if written != size {
		return fail("archive_structure_invalid")
	}
	a.extracted += written
	a.regulars++
	member := Member{Path: name, Size: written, SHA256: hex.EncodeToString(hash.Sum(nil))}
	a.records[index].member = member
	if a.expected != nil {
		if member != a.expected[index].member {
			return ErrStorage
		}
		a.members[name] = member
	}
	return nil
}

func (a *archiveAdmission) tar(source io.Reader) error {
	r := contextReader{a.ctx, source}
	for {
		var block [512]byte
		if _, err := io.ReadFull(r, block[:]); err != nil {
			return fail("archive_structure_invalid")
		}
		if block == [512]byte{} {
			// Bound structural padding separately from regular-file bytes. A
			// compressed trailer must not expand indefinitely without consuming
			// either the file count or payload byte budgets.
			if _, err := io.ReadFull(r, block[:]); err != nil || block != [512]byte{} {
				return fail("archive_structure_invalid")
			}
			for paddingBlocks := 2; ; paddingBlocks++ {
				n, err := io.ReadFull(r, block[:])
				if n == 0 && errors.Is(err, io.EOF) {
					return nil
				}
				if paddingBlocks >= 20 {
					return limitFailure("archive_structure_invalid", "max_tar_terminating_zero_blocks", 20, uint64(paddingBlocks)+1)
				}
				if err != nil || block != [512]byte{} {
					return fail("archive_structure_invalid")
				}
			}
		}
		if !bytes.Equal(block[257:265], []byte("ustar\x0000")) || !validUSTARChecksum(block[:]) || !canonicalTARNameField(block[:100]) || !canonicalTARNameField(block[345:500]) {
			return fail("archive_structure_invalid")
		}
		kind := block[156]
		// Both ASCII '0' and the original NUL marker denote a regular ustar file.
		regular := kind == tar.TypeReg || kind == 0
		// Extension headers are never passed to archive/tar: it would consume
		// their attacker-declared bodies while parsing Next. Plain link and
		// special headers have no body and can be diagnosed after path checks.
		if kind == tar.TypeXHeader || kind == tar.TypeXGlobalHeader || kind == tar.TypeGNULongName || kind == tar.TypeGNULongLink || kind == tar.TypeGNUSparse {
			name := string(bytes.TrimRight(block[:100], "\x00"))
			if prefix := string(bytes.TrimRight(block[345:500], "\x00")); prefix != "" {
				name = prefix + "/" + name
			}
			if err := a.record(name, false, false); err != nil {
				return err
			}
			return fail("disallowed_member_type")
		}
		// Parse only a previously classified plain ustar header. GNU/PAX readers
		// never receive an extension header and cannot allocate its claimed body.
		header, err := tar.NewReader(bytes.NewReader(block[:])).Next()
		if err != nil || header.Format != tar.FormatUSTAR {
			return fail("archive_structure_invalid")
		}
		directory := kind == tar.TypeDir
		if err := a.record(header.Name, directory, regular); err != nil {
			return err
		}
		if !directory && !regular {
			if header.Size != 0 {
				return fail("disallowed_member_type")
			}
			continue
		}
		if directory {
			if header.Size != 0 {
				return fail("archive_structure_invalid")
			}
			continue
		}
		if err := a.member(header.Name, header.Size, r); err != nil {
			return err
		}
		padding := (512 - header.Size%512) % 512
		if _, err := io.CopyN(io.Discard, r, padding); err != nil {
			return fail("archive_structure_invalid")
		}
	}
}

func (a *archiveAdmission) zip(source io.ReaderAt, size int64) error {
	endSize := min(size, int64(65557))
	tail := make([]byte, endSize)
	if _, err := source.ReadAt(tail, size-endSize); err != nil {
		return err
	}
	end := -1
	for i := len(tail) - 22; i >= 0; i-- {
		if bytes.Equal(tail[i:i+4], []byte{'P', 'K', 5, 6}) && i+22+int(binary.LittleEndian.Uint16(tail[i+20:i+22])) == len(tail) {
			end = i
			break
		}
	}
	if end < 0 {
		return fail("archive_structure_invalid")
	}
	e := tail[end:]
	u16 := binary.LittleEndian.Uint16
	u32 := binary.LittleEndian.Uint32
	if u16(e[4:6]) != 0 || u16(e[6:8]) != 0 || u16(e[8:10]) != u16(e[10:12]) || u16(e[10:12]) == 65535 || u32(e[12:16]) == math.MaxUint32 || u32(e[16:20]) == math.MaxUint32 {
		return fail("archive_structure_invalid")
	}
	central := int64(u32(e[16:20]))
	centralSize := int64(u32(e[12:16]))
	if central+centralSize != size-endSize+int64(end) {
		return fail("archive_structure_invalid")
	}
	offset := central
	// Do not construct archive/zip.Reader: it retains every untrusted name,
	// comment and extra field before any owner limit can be checked. Read one
	// bounded central record at a time; comments never enter working storage.
	intervals := make([][2]int64, 0)
	for range int(u16(e[10:12])) {
		if err := a.ctx.Err(); err != nil {
			return err
		}
		var record [46]byte
		if offset > central+centralSize-46 {
			return fail("archive_structure_invalid")
		}
		if _, err := source.ReadAt(record[:], offset); err != nil || !bytes.Equal(record[:4], []byte{'P', 'K', 1, 2}) {
			return fail("archive_structure_invalid")
		}
		nameLength := int64(u16(record[28:30]))
		extraLength := int64(u16(record[30:32]))
		commentLength := int64(u16(record[32:34]))
		local := int64(u32(record[42:46]))
		if nameLength < 1 || nameLength > 1024 || offset+46+nameLength+extraLength+commentLength > central+centralSize {
			return fail("archive_structure_invalid")
		}
		name := make([]byte, nameLength)
		if _, err := source.ReadAt(name, offset+46); err != nil {
			return fail("archive_structure_invalid")
		}
		file := zip.FileHeader{
			Name: string(name), CreatorVersion: u16(record[4:6]),
			Flags: u16(record[8:10]), Method: u16(record[10:12]), CRC32: u32(record[16:20]),
			CompressedSize64: uint64(u32(record[20:24])), UncompressedSize64: uint64(u32(record[24:28])),
			ExternalAttrs: u32(record[38:42]),
		}
		if u16(record[34:36]) != 0 || file.Flags&(1|64|8192) != 0 || (file.Method != zip.Store && file.Method != zip.Deflate) || file.CompressedSize64 >= math.MaxUint32 || file.UncompressedSize64 >= math.MaxUint32 {
			return fail("archive_structure_invalid")
		}
		extra := make([]byte, extraLength)
		if _, err := source.ReadAt(extra, offset+46+nameLength); err != nil || !validZIPExtra(extra) {
			return fail("archive_structure_invalid")
		}
		var localHeader [30]byte
		if local > central-30 {
			return fail("archive_structure_invalid")
		}
		if _, err := source.ReadAt(localHeader[:], local); err != nil || !bytes.Equal(localHeader[:4], []byte{'P', 'K', 3, 4}) || u16(localHeader[6:8]) != file.Flags || u16(localHeader[8:10]) != file.Method || int64(u16(localHeader[26:28])) != nameLength {
			return fail("archive_structure_invalid")
		}
		localName := make([]byte, nameLength)
		if _, err := source.ReadAt(localName, local+30); err != nil || string(localName) != file.Name {
			return fail("archive_structure_invalid")
		}
		localExtra := make([]byte, int(u16(localHeader[28:30])))
		if _, err := source.ReadAt(localExtra, local+30+nameLength); err != nil || !validZIPExtra(localExtra) {
			return fail("archive_structure_invalid")
		}
		dataOffset := local + 30 + nameLength + int64(len(localExtra))
		expectedLocal := dataOffset + int64(file.CompressedSize64)
		if expectedLocal > central {
			return fail("archive_structure_invalid")
		}
		if file.Flags&8 != 0 {
			var descriptor [16]byte
			if _, err := source.ReadAt(descriptor[:], expectedLocal); err != nil {
				return fail("archive_structure_invalid")
			}
			start := 0
			if u32(descriptor[:4]) == 0x08074b50 {
				start = 4
			}
			if u32(descriptor[start:start+4]) != file.CRC32 || uint64(u32(descriptor[start+4:start+8])) != file.CompressedSize64 || uint64(u32(descriptor[start+8:start+12])) != file.UncompressedSize64 {
				return fail("archive_structure_invalid")
			}
			expectedLocal += int64(start + 12)
		} else if u32(localHeader[14:18]) != file.CRC32 || uint64(u32(localHeader[18:22])) != file.CompressedSize64 || uint64(u32(localHeader[22:26])) != file.UncompressedSize64 {
			return fail("archive_structure_invalid")
		}
		if expectedLocal > central {
			return fail("archive_structure_invalid")
		}
		directory := file.FileInfo().IsDir()
		regular := !directory && file.Mode().IsRegular()
		if err := a.record(file.Name, directory, regular); err != nil {
			return err
		}
		if directory && file.UncompressedSize64 != 0 {
			return fail("archive_structure_invalid")
		}
		if !directory && !regular && file.UncompressedSize64 != 0 {
			return fail("disallowed_member_type")
		}
		if regular {
			if err := a.reserve(int64(file.UncompressedSize64)); err != nil {
				return err
			}
		}
		if err := a.zipMember(source, dataOffset, file, !regular); err != nil {
			return err
		}
		intervals = append(intervals, [2]int64{local, expectedLocal})
		offset += 46 + nameLength + extraLength + commentLength
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i][0] < intervals[j][0] })
	expectedLocal := int64(0)
	for _, interval := range intervals {
		if interval[0] != expectedLocal {
			return fail("archive_structure_invalid")
		}
		expectedLocal = interval[1]
	}
	if expectedLocal != central || offset != central+centralSize {
		return fail("archive_structure_invalid")
	}
	return nil
}

func (a *archiveAdmission) zipMember(source io.ReaderAt, offset int64, file zip.FileHeader, directory bool) error {
	compressed := bufio.NewReader(contextReader{a.ctx, io.NewSectionReader(source, offset, int64(file.CompressedSize64))})
	stream := io.NopCloser(compressed)
	if file.Method == zip.Deflate {
		// ByteReader prevents flate from consuming bytes beyond its end marker.
		stream = flate.NewReader(compressed)
	} else if file.CompressedSize64 != file.UncompressedSize64 {
		return fail("archive_structure_invalid")
	}
	checksum := crc32.NewIEEE()
	decoded := io.TeeReader(stream, checksum)
	var copyErr error
	if !directory {
		copyErr = a.member(file.Name, int64(file.UncompressedSize64), decoded)
	}
	var one [1]byte
	n, endErr := decoded.Read(one[:])
	closeErr := stream.Close()
	if copyErr != nil {
		return copyErr
	}
	if n != 0 || !errors.Is(endErr, io.EOF) || closeErr != nil || checksum.Sum32() != file.CRC32 {
		return fail("archive_structure_invalid")
	}
	if _, err := compressed.ReadByte(); !errors.Is(err, io.EOF) {
		return fail("archive_structure_invalid")
	}
	return nil
}

func validZIPExtra(extra []byte) bool {
	for len(extra) > 0 {
		if len(extra) < 4 {
			return false
		}
		length := int(binary.LittleEndian.Uint16(extra[2:4]))
		if binary.LittleEndian.Uint16(extra[:2]) == 1 || length > len(extra)-4 {
			return false
		}
		extra = extra[4+length:]
	}
	return true
}

// Ustar has NUL-padded path fields. Reject hidden suffix bytes before the
// standard reader can truncate them into a different logical name.
func canonicalTARNameField(field []byte) bool {
	if at := bytes.IndexByte(field, 0); at >= 0 {
		for _, b := range field[at:] {
			if b != 0 {
				return false
			}
		}
	}
	return true
}
func validUSTARChecksum(header []byte) bool {
	token := strings.Trim(string(header[148:156]), " \x00")
	if token == "" {
		return false
	}
	for _, b := range token {
		if b < '0' || b > '7' {
			return false
		}
	}
	declared, err := strconv.ParseUint(token, 8, 64)
	if err != nil {
		return false
	}
	sum := uint64(0)
	for i, b := range header {
		if i >= 148 && i < 156 {
			sum += 32
		} else {
			sum += uint64(b)
		}
	}
	return sum == declared
}
