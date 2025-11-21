package iowrapper

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
)

const (
	defaultFileCount    = 5
	defaultLinesPerFile = 1_000_000
	randomDataLength    = 32
)

func createFakeLogFiles(t testing.TB, fileCount, linesPerFile int) (string, []string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "pipeline-logs-")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	files := make([]string, 0, fileCount)
	for i := 0; i < fileCount; i++ {
		filename := fmt.Sprintf("file_%02d.log", i)
		path := filepath.Join(dir, filename)

		if err := writeFakeLogFile(path, filename, linesPerFile); err != nil {
			os.RemoveAll(dir)
			t.Fatalf("failed to write fake log file %s: %v", filename, err)
		}

		files = append(files, path)
	}

	cleanup := func() {
		_ = os.RemoveAll(dir)
	}
	return dir, files, cleanup
}

func writeFakeLogFile(path, filename string, lines int) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := bufio.NewWriterSize(f, 256*1024)
	defer writer.Flush()

	rng := rand.New(rand.NewSource(int64(len(filename) + lines)))

	for i := 0; i < lines; i++ {
		if _, err := fmt.Fprintf(writer, "file=%s index=%d data=%s\n", filename, i, randomData(rng, randomDataLength)); err != nil {
			return err
		}
	}

	return nil
}

func randomData(rng *rand.Rand, n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}

func TestCreateFakeLogFilesLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping large file generation in short test mode")
	}
	if os.Getenv("GENERATE_LARGE_FIXTURES") != "1" {
		t.Skip("set GENERATE_LARGE_FIXTURES=1 to generate 5x1M-line files")
	}

	dir, files, cleanup := createFakeLogFiles(t, defaultFileCount, defaultLinesPerFile)
	defer cleanup()

	t.Logf("generated %d files in %s", len(files), dir)
}

type FakeLogRecord struct {
	File  string
	Index int
	Data  string
}

func parseChunkToFakeLogRecords(chunk FileChunk) ([]FakeLogRecord, error) {
	base := filepath.Base(chunk.File)
	lines := bytes.Split(chunk.Chunk.Data, []byte{'\n'})
	records := make([]FakeLogRecord, 0, len(lines))

	for _, raw := range lines {
		if len(raw) == 0 {
			continue
		}
		line := string(raw)
		parts := strings.Fields(line)
		if len(parts) < 3 {
			return nil, fmt.Errorf("malformed line %q", line)
		}

		if !strings.HasPrefix(parts[0], "file=") || !strings.HasPrefix(parts[1], "index=") || !strings.HasPrefix(parts[2], "data=") {
			return nil, fmt.Errorf("unexpected line format %q", line)
		}

		fileName := strings.TrimPrefix(parts[0], "file=")
		if fileName != base {
			return nil, fmt.Errorf("chunk file %s does not match line %s", base, fileName)
		}

		idxStr := strings.TrimPrefix(parts[1], "index=")
		idx, err := strconv.Atoi(idxStr)
		if err != nil {
			return nil, fmt.Errorf("invalid index %q: %w", idxStr, err)
		}

		data := strings.TrimPrefix(parts[2], "data=")
		records = append(records, FakeLogRecord{
			File:  chunk.File,
			Index: idx,
			Data:  data,
		})
	}

	return records, nil
}

func TestPipelineEndToEnd(t *testing.T) {
	fileCount := 10
	linesPerFile := 1000000

	_, files, cleanup := createFakeLogFiles(t, fileCount, linesPerFile)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileCh := StartFileProducer(ctx, files)
	chunkCh, errCh := StartChunkWorkers(ctx, 2, 4*1024*1024, fileCh)
	processedCh := StartChunkProcessors[FakeLogRecord](ctx, 8, chunkCh, parseChunkToFakeLogRecords)

	var mu sync.Mutex
	collected := make(map[string][]FakeLogRecord)

	sink := func(file string, items []FakeLogRecord) error {
		mu.Lock()
		defer mu.Unlock()
		cpy := append([]FakeLogRecord(nil), items...)
		collected[file] = append(collected[file], cpy...)
		// for _, item := range items {
		// 	fmt.Printf("%s %d %s\n", item.File, item.Index, item.Data)
		// }
		return nil
	}

	if err := StreamOrdered(ctx, processedCh, sink); err != nil {
		t.Fatalf("StreamOrdered returned error: %v", err)
	}

	cancel()

	for err := range errCh {
		if err != nil {
			t.Fatalf("chunk worker error: %v", err)
		}
	}

	for _, path := range files {
		records, ok := collected[path]
		if !ok {
			t.Fatalf("missing records for %s", path)
		}
		if len(records) != linesPerFile {
			t.Fatalf("file %s expected %d records, got %d", path, linesPerFile, len(records))
		}
		for idx, rec := range records {
			if rec.File != path {
				t.Fatalf("record file mismatch: got %s want %s", rec.File, path)
			}
			if rec.Index != idx {
				t.Fatalf("record index mismatch for %s: got %d want %d", path, rec.Index, idx)
			}
			if len(rec.Data) != randomDataLength {
				t.Fatalf("unexpected data length for file %s index %d: got %d", path, rec.Index, len(rec.Data))
			}
		}
	}
}
