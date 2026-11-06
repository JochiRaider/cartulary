package incidentbundles

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

// stageBundleArchive accepts one current container without guessing another
// decoder after a recognized container fails. Logical bundle-version admission
// follows catalog verification and never enables a legacy format reader.
func stageBundleArchive(ctx context.Context, source io.ReaderAt, size int64, members *memberStore) error {
	if source == nil || size < 1 {
		return &verificationError{ReasonCode: "missing_required_file"}
	}
	var magic [4]byte
	n, _ := source.ReadAt(magic[:], 0)
	switch {
	case n >= 4 && bytes.Equal(magic[:2], []byte("PK")):
		return stageZIP(ctx, source, size, members)
	case n >= 2 && magic[0] == 0x1f && magic[1] == 0x8b:
		input := bufio.NewReader(io.NewSectionReader(source, 0, size))
		gz, err := gzip.NewReader(input)
		if err != nil {
			return &verificationError{ReasonCode: "unsupported_member_type"}
		}
		defer gz.Close()
		gz.Multistream(false)
		if err := stageTAR(ctx, gz, size, members); err != nil {
			return err
		}
		if _, err := input.Peek(1); err != io.EOF {
			return &verificationError{ReasonCode: "unsupported_member_type"}
		}
		return nil
	default:
		return stageTAR(ctx, io.NewSectionReader(source, 0, size), size, members)
	}
}

func stageZIP(ctx context.Context, source io.ReaderAt, size int64, members *memberStore) error {
	if err := admitZIPDirectory(source, size, members.limits); err != nil {
		return err
	}
	archive, err := zip.NewReader(source, size)
	if err != nil {
		return &verificationError{ReasonCode: "unsupported_member_type"}
	}
	if err := checkMemberCount(len(archive.File), members.limits); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, file := range archive.File {
		if err := ctx.Err(); err != nil {
			return err
		}
		kind, err := classifyZipMember(file)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(file.Name, "/")
		if seen[name] {
			return &verificationError{ReasonCode: "malformed_manifest"}
		}
		seen[name] = true
		if kind == archiveMemberDirectory {
			if file.UncompressedSize64 != 0 {
				return &verificationError{ReasonCode: "unsupported_member_type"}
			}
			continue
		}
		if file.UncompressedSize64 > uint64(members.maximumBytes()-members.total) {
			return &verificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
		}
		if err := checkCompressionRatio(members.total+int64(file.UncompressedSize64), size, members.limits); err != nil {
			return err
		}
		input, err := file.Open()
		if err != nil {
			return &verificationError{ReasonCode: "checksum_mismatch"}
		}
		sealed, writeErr := members.WriteMember(ctx, file.Name, input)
		closeErr := input.Close()
		if errors.Is(writeErr, zip.ErrChecksum) {
			return &verificationError{ReasonCode: "checksum_mismatch"}
		}
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		if sealed.ByteSize != int64(file.UncompressedSize64) {
			return &verificationError{ReasonCode: "checksum_mismatch"}
		}
	}
	return nil
}

func stageTAR(ctx context.Context, source io.Reader, containerSize int64, members *memberStore) error {
	// Bound expanded framing too, including hidden PAX headers and trailing zero
	// blocks. Those bytes do not evade the normal extracted-member ceiling.
	maxMembers := members.limits.Archives.MaxMembers
	if maxMembers <= 0 {
		maxMembers = defaultArchiveMaxMembers
	}
	framing := &io.LimitedReader{R: source, N: members.maximumBytes() + maxMembers*1024 + PortabilityParticipantByteLimit + 1}
	archive := tar.NewReader(framing)
	seen := map[string]bool{}
	count := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		header, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return &verificationError{ReasonCode: "unsupported_member_type"}
		}
		count++
		if err := checkMemberCount(count, members.limits); err != nil {
			return err
		}
		kind, err := classifyTarMember(header)
		if err != nil {
			return err
		}
		name := strings.TrimSuffix(header.Name, "/")
		if seen[name] {
			return &verificationError{ReasonCode: "malformed_manifest"}
		}
		seen[name] = true
		if kind == archiveMemberDirectory {
			if header.Size != 0 {
				return &verificationError{ReasonCode: "unsupported_member_type"}
			}
			continue
		}
		if header.Size < 0 || header.Size > members.maximumBytes()-members.total {
			return &verificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
		}
		if err := checkCompressionRatio(members.total+header.Size, containerSize, members.limits); err != nil {
			return err
		}
		sealed, err := members.WriteMember(ctx, header.Name, archive)
		if err != nil {
			return err
		}
		if sealed.ByteSize != header.Size {
			return &verificationError{ReasonCode: "checksum_mismatch"}
		}
	}
	if count == 0 {
		return &verificationError{ReasonCode: "missing_required_file"}
	}
	var buf [32 * 1024]byte
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := framing.Read(buf[:])
		for _, b := range buf[:n] {
			if b != 0 {
				return &verificationError{ReasonCode: "unsupported_member_type"}
			}
		}
		if framing.N == 0 {
			return &verificationError{ReasonCode: "archive_extracted_bytes_exceeded"}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return &verificationError{ReasonCode: "checksum_mismatch"}
		}
	}
}

// Go's ZIP reader builds central-directory metadata before exposing members.
// Bound the declared directory count and bytes first, including ZIP64 archives.
func admitZIPDirectory(source io.ReaderAt, size int64, limits Limits) error {
	if size < 22 {
		return &verificationError{ReasonCode: "unsupported_member_type"}
	}
	tail := make([]byte, min(size, 65535+22))
	if _, err := source.ReadAt(tail, size-int64(len(tail))); err != nil {
		return err
	}
	index := -1
	for i := len(tail) - 22; i >= 0; i-- {
		if binary.LittleEndian.Uint32(tail[i:]) == 0x06054b50 && i+22+int(binary.LittleEndian.Uint16(tail[i+20:])) == len(tail) {
			index = i
			break
		}
	}
	if index < 0 {
		return &verificationError{ReasonCode: "unsupported_member_type"}
	}
	end := tail[index:]
	offset := size - int64(len(tail)) + int64(index)
	if binary.LittleEndian.Uint16(end[4:]) != 0 || binary.LittleEndian.Uint16(end[6:]) != 0 {
		return &verificationError{ReasonCode: "unsupported_member_type"}
	}
	count := uint64(binary.LittleEndian.Uint16(end[10:]))
	directoryBytes := uint64(binary.LittleEndian.Uint32(end[12:]))
	directoryOffset := uint64(binary.LittleEndian.Uint32(end[16:]))
	if count == 65535 || directoryBytes == 0xffffffff || directoryOffset == 0xffffffff {
		var locator [20]byte
		if offset < 20 {
			return &verificationError{ReasonCode: "unsupported_member_type"}
		}
		if _, err := source.ReadAt(locator[:], offset-20); err != nil {
			return err
		}
		if binary.LittleEndian.Uint32(locator[:]) != 0x07064b50 || binary.LittleEndian.Uint32(locator[4:]) != 0 || binary.LittleEndian.Uint32(locator[16:]) != 1 {
			return &verificationError{ReasonCode: "unsupported_member_type"}
		}
		position := binary.LittleEndian.Uint64(locator[8:])
		if position > uint64(offset-20) || uint64(offset-20)-position < 56 {
			return &verificationError{ReasonCode: "unsupported_member_type"}
		}
		var record [56]byte
		if _, err := source.ReadAt(record[:], int64(position)); err != nil {
			return err
		}
		if binary.LittleEndian.Uint32(record[:]) != 0x06064b50 || binary.LittleEndian.Uint64(record[4:]) < 44 || binary.LittleEndian.Uint32(record[16:]) != 0 || binary.LittleEndian.Uint32(record[20:]) != 0 {
			return &verificationError{ReasonCode: "unsupported_member_type"}
		}
		count = binary.LittleEndian.Uint64(record[32:])
		directoryBytes = binary.LittleEndian.Uint64(record[40:])
		directoryOffset = binary.LittleEndian.Uint64(record[48:])
		if binary.LittleEndian.Uint64(record[24:]) != count {
			return &verificationError{ReasonCode: "unsupported_member_type"}
		}
	} else if uint64(binary.LittleEndian.Uint16(end[8:])) != count {
		return &verificationError{ReasonCode: "unsupported_member_type"}
	}
	maximum := limits.Archives.MaxMembers
	if maximum <= 0 {
		maximum = defaultArchiveMaxMembers
	}
	if count > uint64(maximum) {
		return &verificationError{ReasonCode: "archive_member_count_exceeded"}
	}
	if directoryBytes > PortabilityParticipantByteLimit || directoryOffset > uint64(size) || directoryBytes > uint64(size)-directoryOffset {
		return &verificationError{ReasonCode: "malformed_manifest"}
	}
	// Do not trust the EOCD count to bound allocation: count the actual records
	// with fixed scratch storage before archive/zip constructs its file slice.
	directory := io.NewSectionReader(source, int64(directoryOffset), int64(directoryBytes))
	var actual uint64
	for {
		var header [46]byte
		n, err := io.ReadFull(directory, header[:])
		if err == io.EOF && n == 0 {
			break
		}
		if err != nil || binary.LittleEndian.Uint32(header[:]) != 0x02014b50 {
			return &verificationError{ReasonCode: "malformed_manifest"}
		}
		actual++
		if actual > uint64(maximum) {
			return &verificationError{ReasonCode: "archive_member_count_exceeded"}
		}
		skip := int64(binary.LittleEndian.Uint16(header[28:])) + int64(binary.LittleEndian.Uint16(header[30:])) + int64(binary.LittleEndian.Uint16(header[32:]))
		if _, err := io.CopyN(io.Discard, directory, skip); err != nil {
			return &verificationError{ReasonCode: "malformed_manifest"}
		}
	}
	if actual != count {
		return &verificationError{ReasonCode: "malformed_manifest"}
	}
	return nil
}

func writeZIPMembers(ctx context.Context, output io.Writer, members incidentportability.MemberReader) error {
	archive := zip.NewWriter(output)
	for _, member := range members.Members() {
		if err := ctx.Err(); err != nil {
			return err
		}
		header := &zip.FileHeader{Name: member.Path, Method: zip.Store}
		header.SetMode(0600)
		header.Modified = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
		destination, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		source, _, err := members.OpenMember(ctx, member.Path)
		if err != nil {
			return err
		}
		n, writeErr := io.CopyBuffer(destination, source, make([]byte, 32*1024))
		closeErr := source.Close()
		if writeErr != nil {
			return writeErr
		}
		if closeErr != nil {
			return closeErr
		}
		if n != member.ByteSize {
			return &verificationError{ReasonCode: "checksum_mismatch"}
		}
	}
	return archive.Close()
}

// Verify the exact admitted upload before container decoding. Archive member
// checksums do not bind an otherwise valid replacement archive to its receipt.
func verifyStagedArchiveDigest(ctx context.Context, source io.ReaderAt, size int64, expected string) error {
	if len(expected) != 64 || size < 1 || source == nil {
		return &verificationError{ReasonCode: "checksum_mismatch"}
	}
	input := io.NewSectionReader(source, 0, size)
	digest := sha256.New()
	var buffer [32 * 1024]byte
	var read int64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := input.Read(buffer[:])
		if n > 0 {
			_, _ = digest.Write(buffer[:n])
			read += int64(n)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return &verificationError{ReasonCode: "checksum_mismatch"}
		}
	}
	if read != size || hex.EncodeToString(digest.Sum(nil)) != expected {
		return &verificationError{ReasonCode: "checksum_mismatch"}
	}
	return nil
}
