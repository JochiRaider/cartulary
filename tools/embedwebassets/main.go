package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var stableZipTime = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type archiveEntry struct {
	absPath string
	name    string
	isDir   bool
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run(args []string) error {
	var sourceDir string
	var sourceIndex string
	var output string
	var tempDir string

	flags := flag.NewFlagSet("embedwebassets", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&sourceDir, "source-dir", "", "frontend dist directory")
	flags.StringVar(&sourceIndex, "source-index", "", "frontend dist index.html")
	flags.StringVar(&output, "output", "", "output zip path")
	flags.StringVar(&tempDir, "temp-dir", "", "temporary archive directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if sourceDir == "" {
		return errors.New("--source-dir is required")
	}
	if sourceIndex == "" {
		return errors.New("--source-index is required")
	}
	if output == "" {
		return errors.New("--output is required")
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	return writeArchive(sourceDir, sourceIndex, output, tempDir)
}

func writeArchive(sourceDir, sourceIndex, output string, tempDir string) error {
	sourceRoot, err := filepath.Abs(sourceDir)
	if err != nil {
		return fmt.Errorf("resolve source dir: %w", err)
	}
	sourceInfo, err := os.Lstat(sourceRoot)
	if err != nil {
		return fmt.Errorf("stat source dir: %w", err)
	}
	if !sourceInfo.IsDir() || sourceInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source dir must be a real directory: %s", sourceDir)
	}

	indexPath, err := filepath.Abs(sourceIndex)
	if err != nil {
		return fmt.Errorf("resolve source index: %w", err)
	}
	indexRel, err := filepath.Rel(sourceRoot, indexPath)
	if err != nil || indexRel == "." || strings.HasPrefix(indexRel, ".."+string(filepath.Separator)) || filepath.IsAbs(indexRel) {
		return fmt.Errorf("source index must be inside source dir: %s", sourceIndex)
	}
	indexInfo, err := os.Lstat(indexPath)
	if err != nil {
		return fmt.Errorf("stat source index: %w", err)
	}
	if !indexInfo.Mode().IsRegular() || indexInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("source index must be a regular file: %s", sourceIndex)
	}

	entries, err := collectEntries(sourceRoot)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("source dir contains no embeddable files: %s", sourceDir)
	}

	outputPath, err := filepath.Abs(output)
	if err != nil {
		return fmt.Errorf("resolve output: %w", err)
	}
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}
	tempRoot := outputDir
	if tempDir != "" {
		tempRoot, err = filepath.Abs(tempDir)
		if err != nil {
			return fmt.Errorf("resolve temp dir: %w", err)
		}
	}
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	tempFile, err := os.CreateTemp(tempRoot, "."+filepath.Base(outputPath)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp archive: %w", err)
	}
	tempPath := tempFile.Name()
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			_ = os.Remove(tempPath)
		}
	}()

	if err := writeZip(tempFile, entries); err != nil {
		_ = tempFile.Close()
		return err
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp archive: %w", err)
	}
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("publish archive: %w", err)
	}
	cleanupTemp = false
	return nil
}

func collectEntries(sourceRoot string) ([]archiveEntry, error) {
	var entries []archiveEntry
	if err := filepath.WalkDir(sourceRoot, func(current string, dirEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if current == sourceRoot {
			return nil
		}
		info, err := dirEntry.Info()
		if err != nil {
			return fmt.Errorf("stat source entry %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("source entry must not be a symlink: %s", current)
		}
		if !info.IsDir() && !info.Mode().IsRegular() {
			return fmt.Errorf("source entry must be a regular file or directory: %s", current)
		}
		name, err := archiveName(sourceRoot, current, info.IsDir())
		if err != nil {
			return err
		}
		entries = append(entries, archiveEntry{
			absPath: current,
			name:    name,
			isDir:   info.IsDir(),
		})
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].name < entries[right].name
	})
	return entries, nil
}

func archiveName(sourceRoot, current string, isDir bool) (string, error) {
	rel, err := filepath.Rel(sourceRoot, current)
	if err != nil {
		return "", fmt.Errorf("resolve source-relative path: %w", err)
	}
	if rel == "." || rel == "" {
		return "", fmt.Errorf("invalid source-relative path: %s", current)
	}
	name := filepath.ToSlash(rel)
	clean := path.Clean(name)
	if clean == "." || clean == "" || strings.HasPrefix(clean, "../") || path.IsAbs(clean) || clean != name {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	if isDir {
		clean += "/"
	}
	return clean, nil
}

func writeZip(writer io.Writer, entries []archiveEntry) error {
	zipWriter := zip.NewWriter(writer)
	for _, entry := range entries {
		header := &zip.FileHeader{
			Name:     entry.name,
			Method:   zip.Deflate,
			Modified: stableZipTime,
		}
		if entry.isDir {
			header.Method = zip.Store
			header.SetMode(0o755 | fs.ModeDir)
		} else {
			header.SetMode(0o644)
		}
		zipEntry, err := zipWriter.CreateHeader(header)
		if err != nil {
			_ = zipWriter.Close()
			return fmt.Errorf("create zip entry %s: %w", entry.name, err)
		}
		if entry.isDir {
			continue
		}
		file, err := os.Open(entry.absPath)
		if err != nil {
			_ = zipWriter.Close()
			return fmt.Errorf("open source file %s: %w", entry.absPath, err)
		}
		if _, err := io.Copy(zipEntry, file); err != nil {
			_ = file.Close()
			_ = zipWriter.Close()
			return fmt.Errorf("copy source file %s: %w", entry.absPath, err)
		}
		if err := file.Close(); err != nil {
			_ = zipWriter.Close()
			return fmt.Errorf("close source file %s: %w", entry.absPath, err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("close zip archive: %w", err)
	}
	return nil
}
