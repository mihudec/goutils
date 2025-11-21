package iowrapper

import (
	"strings"
	"testing"
)

func collectChunks(t testing.TB, fn bytesChunkerFunc, input string, chunkSize int) []BytesChunk {
	t.Helper()

	r := strings.NewReader(input)
	ch := make(chan BytesChunk)

	go func() {
		fn(r, chunkSize, ch)
		close(ch)
	}()

	var chunks []BytesChunk
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func TestSliceToBytesChunksVariants(t *testing.T) {
	chunkers := []struct {
		name string
		fn   bytesChunkerFunc
	}{
		{name: "LineReader", fn: SliceToBytesChunks},
		{name: "Buffered", fn: SliceToBytesChunks2},
	}

	longLine := strings.Repeat("x", 20)

	tests := []struct {
		name      string
		input     string
		chunkSize int
		want      []string
	}{
		{
			name:      "BasicAggregation",
			input:     "line1\nline2\nline3\n",
			chunkSize: 15,
			want:      []string{"line1\nline2\n", "line3\n"},
		},
		{
			name:      "EveryLineOwnChunk",
			input:     "line1\nline2\nline3\n",
			chunkSize: 10,
			want:      []string{"line1\n", "line2\n", "line3\n"},
		},
		{
			name:      "OversizedLine",
			input:     longLine + "\nshort\n",
			chunkSize: 10,
			want:      []string{longLine + "\n", "short\n"},
		},
		{
			name:      "NoTrailingNewline",
			input:     "line1\nline2",
			chunkSize: 10,
			want:      []string{"line1\n", "line2"},
		},
		{
			name:      "NoNewlineAtAll",
			input:     strings.Repeat("abc", 5),
			chunkSize: 4,
			want:      []string{strings.Repeat("abc", 5)},
		},
	}

	for _, chunker := range chunkers {
		chunker := chunker
		t.Run(chunker.name, func(t *testing.T) {
			for _, tt := range tests {
				tt := tt
				t.Run(tt.name, func(t *testing.T) {
					got := collectChunks(t, chunker.fn, tt.input, tt.chunkSize)
					if len(got) != len(tt.want) {
						t.Fatalf("expected %d chunks, got %d (%v)", len(tt.want), len(got), got)
					}
					for i, chunk := range got {
						if chunk.Index != i {
							t.Fatalf("chunk %d index = %d, want %d", i, chunk.Index, i)
						}
						if string(chunk.Data) != tt.want[i] {
							t.Fatalf("chunk %d data = %q, want %q", i, string(chunk.Data), tt.want[i])
						}
					}
				})
			}
		})
	}
}

func benchmarkChunker(b *testing.B, fn bytesChunkerFunc) {
	line := strings.Repeat("x", 1000) + "\n"
	input := strings.Repeat(line, 1_000_000) // ~1 GB of data
	chunkSize := 4 * 1024 * 1024

	for i := 0; i < b.N; i++ {
		r := strings.NewReader(input)
		ch := make(chan BytesChunk, 16)

		b.StartTimer()
		go func() {
			fn(r, chunkSize, ch)
			close(ch)
		}()

		for range ch {
			// Drain channel
		}
		b.StopTimer()
	}
}

func BenchmarkSliceToBytesChunks(b *testing.B) {
	benchmarkChunker(b, SliceToBytesChunks)
}

func BenchmarkSliceToBytesChunks2(b *testing.B) {
	benchmarkChunker(b, SliceToBytesChunks2)
}
