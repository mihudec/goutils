package iowrapper

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
)

// Scanner opens the source and returns a bufio.Scanner and an io.Closer.
// Caller must defer closer.Close().
func Scanner(source string) (*bufio.Scanner, io.Closer, error) {
	r, closer, err := Reader(source, 64*1024)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open source %s: %v", source, err)
	}
	scanner := bufio.NewScanner(r)
	return scanner, closer, err
}

// Reader opens the source and returns a bufio.Scanner and an io.Closer.
// Caller must defer closer.Close().
func Reader(source string, size int) (*bufio.Reader, io.Closer, error) {
	var r io.Reader
	var closer io.Closer

	if source == "-" {
		r = os.Stdin
		closer = io.NopCloser(nil) // No need to close stdin, but return a no-op closer
	} else {
		file, err := os.Open(source)
		if err != nil {
			return nil, nil, err
		}
		closer = file
		r = file

		// Handle compression
		if strings.HasSuffix(source, ".gz") {
			gr, err := gzip.NewReader(file)
			if err != nil {
				_ = file.Close()
				return nil, nil, err
			}
			r = gr
			closer = closerFunc(func() error {
				gr.Close()
				return file.Close()
			})
		} else if strings.HasSuffix(source, ".zst") || strings.HasSuffix(source, ".zstd") {
			zr, err := zstd.NewReader(file)
			if err != nil {
				_ = file.Close()
				return nil, nil, err
			}
			r = zr
			closer = closerFunc(func() error {
				zr.Close()
				return file.Close()
			})
		}
	}

	reader := bufio.NewReaderSize(r, size)
	return reader, closer, nil
}

func ReadLines(source string) ([]string, error) {
	scanner, closer, err := Scanner(source)
	if err != nil {
		return nil, err
	}
	defer closer.Close()

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// helper to wrap a Close() function
type closerFunc func() error

func (f closerFunc) Close() error {
	return f()
}

func WriteZstdFile(filename string, lines []string) {
	// Do not write empty files
	// if len(lines) == 0 {
	// 	fmt.Fprintf(os.Stderr, "No data to write to %s\n", filename)
	// 	return
	// }

	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	writer, err := zstd.NewWriter(file, zstd.WithZeroFrames(true))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating zstd writer: %v\n", err)
		return
	}
	defer writer.Close()

	linecount := 0
	for _, line := range lines {
		_, err := writer.Write([]byte(line + "\n"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing to %s: %v\n", filename, err)
			return
		}
		linecount++
	}
	fmt.Fprintf(os.Stderr, "Written %d lines to %s\n", linecount, filename)
}

func StdOutWriter() (*bufio.Writer, func()) {
	w := bufio.NewWriter(os.Stdout)
	flush := func() {
		_ = w.Flush() // ignore error, usually nil unless stdout is closed
	}
	return w, flush
}

type BytesChunk struct {
	Index int
	Data  []byte
}

type StringChunk struct {
	Index int
	Data  string
}

type StringsChunk struct {
	Index int
	Data  []string
}

func SliceToBytesChunks(r io.Reader, chunkSize int, chunkCh chan<- BytesChunk) {
	if chunkSize <= 0 {
		return
	}

	reader := bufio.NewReader(r)
	var (
		index  int
		buffer []byte
	)

	flush := func() {
		if len(buffer) == 0 {
			return
		}
		data := make([]byte, len(buffer))
		copy(data, buffer)
		chunkCh <- BytesChunk{Index: index, Data: data}
		index++
		buffer = buffer[:0]
	}

	for {
		line, err := reader.ReadBytes('\n')

		if len(line) > 0 {
			// If the current line alone exceeds chunkSize and we're not holding any data,
			// flush it as a single chunk to avoid splitting mid-line.
			if len(line) > chunkSize && len(buffer) == 0 {
				data := make([]byte, len(line))
				copy(data, line)
				chunkCh <- BytesChunk{Index: index, Data: data}
				index++
			} else {
				if len(buffer)+len(line) > chunkSize && len(buffer) > 0 {
					flush()
				}
				buffer = append(buffer, line...)
				if len(buffer) >= chunkSize {
					flush()
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			// On unexpected errors, flush what we have and stop.
			break
		}
	}

	if len(buffer) > 0 {
		flush()
	}
}
