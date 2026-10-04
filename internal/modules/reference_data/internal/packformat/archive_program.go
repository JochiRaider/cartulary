package packformat

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strconv"
	"strings"
)

type archiveRecord struct {
	name               string
	directory, regular bool
	member             Member
}

type archiveProgram struct {
	ctx       context.Context
	source    io.ReaderAt
	size      int64
	limits    ArchiveLimits
	format    string
	admission *archiveAdmission
	blocked   *Failure
	ordered   []archiveRecord
}

func Extract(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits, destination Destination) (Inventory, error) {
	return extract(ctx, source, size, limits, destination, nil, false)
}

func ExtractWithDiagnostics(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits, destination Destination, scratch DiagnosticScratch) (Inventory, error) {
	return extract(ctx, source, size, limits, destination, scratch, true)
}

// Framing and decompression are checked without writing any archive path. The
// registry then validates all original names and collisions before extraction.
// Extraction rechecks each header and content digest against that frozen scan.
func extract(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits, destination Destination, scratch DiagnosticScratch, diagnostics bool) (Inventory, error) {
	if source == nil || destination == nil || limits.ContainerBytes < 1 || limits.ExtractedBytes < 1 || limits.CompressionRatio < 1 || limits.CompressionRatio > 1000 || limits.Members < 1 {
		return nil, errors.New("reference pack: invalid extraction configuration")
	}
	program, extract, err := ArchiveChecks(ctx, source, size, limits)
	if err != nil {
		return nil, err
	}
	if err := runCheckRange(ctx, "container_format", "archive_member_type", scratch, program, diagnostics); err != nil {
		return nil, err
	}
	return extract(destination)
}

// ArchiveChecks owns every archive rank. Extraction is only available to its
// coordinator after this complete program succeeds.
func ArchiveChecks(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits) (map[string]CheckFunc, func(Destination) (Inventory, error), error) {
	if source == nil || limits.ContainerBytes < 1 || limits.ExtractedBytes < 1 || limits.CompressionRatio < 1 || limits.CompressionRatio > 1000 || limits.Members < 1 {
		return nil, nil, errors.New("reference pack: invalid extraction configuration")
	}
	p := &archiveProgram{ctx: ctx, source: source, size: size, limits: limits}
	p.admission = &archiveAdmission{ctx: ctx, limits: limits, sourceBytes: size}
	program := map[string]CheckFunc{
		"container_format": p.containerFormat, "container_bytes": p.containerBytes,
		"archive_structure": p.structure, "archive_member_count": p.memberCount,
		"archive_extracted_bytes": p.extractedBytes, "archive_compression_ratio": p.ratio,
		"archive_path": p.paths, "archive_collision": p.collisions, "archive_member_type": p.types,
	}
	return program, func(destination Destination) (Inventory, error) {
		if destination == nil {
			return nil, errors.New("reference pack: missing extraction destination")
		}
		extraction := &archiveAdmission{ctx: ctx, limits: limits, sourceBytes: size, expected: p.admission.records, members: Inventory{}, destination: destination}
		// A non-nil empty inventory still denotes extraction of an admitted empty
		// archive; it must never become a second permissive preflight.
		if extraction.expected == nil {
			extraction.expected = []archiveRecord{}
		}
		if err := extraction.scan(source, size, p.format); err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, ErrStorage
		}
		return extraction.members, nil
	}, nil
}

func (p *archiveProgram) containerFormat(ctx context.Context, emit FindingSink) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.size < 4 {
		return emit(Finding{Path: "$"})
	}
	var magic [4]byte
	if _, err := p.source.ReadAt(magic[:], 0); err != nil {
		return emit(Finding{Path: "$"})
	}
	switch {
	case magic == [4]byte{'P', 'K', 3, 4} || magic == [4]byte{'P', 'K', 5, 6}:
		p.format = "zip"
	case magic[0] == 0x1f && magic[1] == 0x8b:
		p.format = "gzip"
	default:
		if p.size < 1024 || p.size%512 != 0 {
			return emit(Finding{Path: "$"})
		}
		var header [512]byte
		if _, err := p.source.ReadAt(header[:], 0); err != nil {
			return emit(Finding{Path: "$"})
		}
		if header != [512]byte{} && !bytes.Equal(header[257:265], []byte("ustar\x0000")) {
			return emit(Finding{Path: "$"})
		}
		p.format = "tar"
	}
	return nil
}
func (p *archiveProgram) containerBytes(_ context.Context, emit FindingSink) error {
	if p.size > p.limits.ContainerBytes {
		return emit(LimitFinding("$", "max_container_bytes", p.limits.ContainerBytes, p.size))
	}
	return nil
}
func (p *archiveProgram) structure(ctx context.Context, emit FindingSink) error {
	err := p.admission.scan(p.source, p.size, p.format)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !errors.As(err, &p.blocked) {
			return err
		}
		if p.blocked.Code == "archive_structure_invalid" {
			return emit(Finding{Path: "$", Details: p.blocked.Details})
		}
		if p.blocked.Code != "archive_member_count_exceeded" && p.blocked.Code != "archive_extracted_bytes_exceeded" && p.blocked.Code != "archive_compression_ratio_exceeded" && p.blocked.Code != "disallowed_member_type" {
			return errors.New("reference pack: unexpected archive preflight check")
		}
	}
	p.ordered = slices.Clone(p.admission.records)
	slices.SortStableFunc(p.ordered, func(a, b archiveRecord) int { return strings.Compare(a.name, b.name) })
	for i, record := range p.ordered {
		if err := structuralPathFindings(record.name, record.directory, archiveDiagnosticPath(i), emit); err != nil {
			return err
		}
	}
	return nil
}
func (p *archiveProgram) guard(code string, emit FindingSink) error {
	if p.blocked != nil && p.blocked.Code == code {
		return emit(Finding{Path: "$", Details: p.blocked.Details})
	}
	return nil
}
func (p *archiveProgram) memberCount(_ context.Context, emit FindingSink) error {
	return p.guard("archive_member_count_exceeded", emit)
}
func (p *archiveProgram) extractedBytes(_ context.Context, emit FindingSink) error {
	return p.guard("archive_extracted_bytes_exceeded", emit)
}
func (p *archiveProgram) ratio(_ context.Context, emit FindingSink) error {
	return p.guard("archive_compression_ratio_exceeded", emit)
}
func archiveDiagnosticPath(i int) string { return "$.archive_members[" + strconv.Itoa(i) + "]" }
func (p *archiveProgram) paths(_ context.Context, emit FindingSink) error {
	for i, record := range p.ordered {
		if ValidatePath(record.name, record.directory) != nil {
			if err := emit(Finding{Path: archiveDiagnosticPath(i)}); err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *archiveProgram) collisions(ctx context.Context, emit FindingSink) error {
	groups := map[string][]int{}
	for i, r := range p.ordered {
		key := strings.ToLower(strings.TrimSuffix(r.name, "/"))
		groups[key] = append(groups[key], i)
	}
	conflicting := make([]bool, len(p.ordered))
	for key, indexes := range groups {
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(indexes) > 1 {
			for _, i := range indexes {
				conflicting[i] = true
			}
		}
		parts := strings.Split(key, "/")
		for j := 1; j < len(parts); j++ {
			for _, parent := range groups[strings.Join(parts[:j], "/")] {
				if !p.ordered[parent].directory {
					conflicting[parent] = true
					for _, i := range indexes {
						conflicting[i] = true
					}
				}
			}
		}
	}
	for i, conflict := range conflicting {
		if conflict {
			if err := emit(Finding{Path: archiveDiagnosticPath(i)}); err != nil {
				return err
			}
		}
	}
	return nil
}
func (p *archiveProgram) types(_ context.Context, emit FindingSink) error {
	found := false
	for i, r := range p.ordered {
		if !r.directory && !r.regular {
			found = true
			if err := emit(Finding{Path: archiveDiagnosticPath(i)}); err != nil {
				return err
			}
		}
	}
	if !found {
		return p.guard("disallowed_member_type", emit)
	}
	return nil
}

// Path byte/segment bounds and marker shape belong to archive_structure even
// when the same original name also has a traversal violation.
func structuralPathFindings(name string, directory bool, path string, emit FindingSink) error {
	if len(name) == 0 {
		return emit(Finding{Path: path})
	}
	if len(name) > 1024 {
		if err := emit(LimitFinding(path, "max_path_bytes", 1024, int64(len(name)))); err != nil {
			return err
		}
	}
	if directory {
		if !strings.HasSuffix(name, "/") {
			if err := emit(Finding{Path: path}); err != nil {
				return err
			}
		}
		name = strings.TrimSuffix(name, "/")
	} else if strings.HasSuffix(name, "/") {
		if err := emit(Finding{Path: path}); err != nil {
			return err
		}
	}
	for _, segment := range strings.Split(name, "/") {
		if len(segment) > 255 {
			if err := emit(LimitFinding(path, "max_segment_bytes", 255, int64(len(segment)))); err != nil {
				return err
			}
		}
	}
	return nil
}
