package iowrapper

import (
	"context"
	"fmt"
	"sync"
)

// FileJob represents a single file that needs to be processed.
type FileJob struct {
	Path string
}

// FileChunk couples a BytesChunk with its originating file path.
type FileChunk struct {
	File  string
	Chunk BytesChunk
}

// ProcessedChunk carries the result of processing a chunk. Items must remain in the same order
// as they appeared in the input chunk.
type ProcessedChunk[T any] struct {
	File       string
	ChunkIndex int
	Items      []T
	Err        error
}

// ChunkProcessorFunc transforms a FileChunk into zero or more domain items.
type ChunkProcessorFunc[T any] func(FileChunk) ([]T, error)

// StartFileProducer pushes the provided file list into a channel and closes it when done.
func StartFileProducer(ctx context.Context, files []string) <-chan FileJob {
	out := make(chan FileJob)
	go func() {
		defer close(out)
		for _, file := range files {
			select {
			case <-ctx.Done():
				return
			case out <- FileJob{Path: file}:
			}
		}
	}()
	return out
}

// StartChunkWorkers spins up a worker pool that reads files, splits them into chunks and streams them.
func StartChunkWorkers(ctx context.Context, workerCount int, chunkSize int, files <-chan FileJob) (<-chan FileChunk, <-chan error) {
	out := make(chan FileChunk)
	errCh := make(chan error, workerCount)

	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-files:
					if !ok {
						return
					}
					if err := streamFileChunks(ctx, job.Path, chunkSize, out); err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
		close(errCh)
	}()

	return out, errCh
}

func streamFileChunks(ctx context.Context, path string, chunkSize int, out chan<- FileChunk) error {
	bufferSize := chunkSize
	if bufferSize <= 0 {
		bufferSize = 64 * 1024
	}

	reader, closer, err := Reader(path, bufferSize)
	if err != nil {
		return fmt.Errorf("open reader for %s: %w", path, err)
	}
	defer closer.Close()

	ch := make(chan BytesChunk)
	go func() {
		SliceToBytesChunks2(reader, chunkSize, ch)
		close(ch)
	}()

	for chunk := range ch {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case out <- FileChunk{File: path, Chunk: chunk}:
		}
	}
	return nil
}

// StartChunkProcessors consumes file chunks and hands them to lineProcessor. The processor function should
// convert each chunk into a slice of domain items. The resulting channel is closed when processing is complete.
func StartChunkProcessors[T any](ctx context.Context, workerCount int, chunks <-chan FileChunk, lineProcessor ChunkProcessorFunc[T]) <-chan ProcessedChunk[T] {
	out := make(chan ProcessedChunk[T])

	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			for chunk := range chunks {
				items, err := lineProcessor(chunk)
				select {
				case <-ctx.Done():
					return
				case out <- ProcessedChunk[T]{
					File:       chunk.File,
					ChunkIndex: chunk.Chunk.Index,
					Items:      items,
					Err:        err,
				}:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

// CollectOrdered drains processed chunk results, ensuring chunks are appended in order per file.
func CollectOrdered[T any](ctx context.Context, in <-chan ProcessedChunk[T]) (map[string][]T, error) {
	results := make(map[string][]T)
	nextIndex := make(map[string]int)
	buffers := make(map[string]map[int][]T)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res, ok := <-in:
			if !ok {
				return results, nil
			}
			if res.Err != nil {
				return nil, res.Err
			}

			if res.Items == nil {
				continue
			}

			if res.ChunkIndex == nextIndex[res.File] {
				results[res.File] = append(results[res.File], res.Items...)
				nextIndex[res.File]++
				drainBuffer(res.File, results, nextIndex, buffers)
				continue
			}

			if _, exists := buffers[res.File]; !exists {
				buffers[res.File] = make(map[int][]T)
			}
			buffers[res.File][res.ChunkIndex] = res.Items
		}
	}
}

func drainBuffer[T any](file string, results map[string][]T, nextIndex map[string]int, buffers map[string]map[int][]T) {
	for {
		if pendingChunks, ok := buffers[file]; ok {
			if entry, ok := pendingChunks[nextIndex[file]]; ok {
				results[file] = append(results[file], entry...)
				delete(pendingChunks, nextIndex[file])
				nextIndex[file]++
				continue
			}
			if len(pendingChunks) == 0 {
				delete(buffers, file)
			}
		}
		return
	}
}

// StreamOrdered invokes sink(file, items) for each chunk in order. It keeps only out-of-order chunks buffered.
func StreamOrdered[T any](ctx context.Context, in <-chan ProcessedChunk[T], sink func(string, []T) error) error {
	nextIndex := make(map[string]int)
	buffers := make(map[string]map[int][]T)

	flush := func(file string) error {
		for {
			if pending, ok := buffers[file]; ok {
				items, found := pending[nextIndex[file]]
				if !found {
					if len(pending) == 0 {
						delete(buffers, file)
					}
					return nil
				}
				if err := sink(file, items); err != nil {
					return err
				}
				delete(pending, nextIndex[file])
				nextIndex[file]++
				continue
			}
			return nil
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case res, ok := <-in:
			if !ok {
				return nil
			}
			if res.Err != nil {
				return res.Err
			}
			if res.Items == nil {
				continue
			}

			if res.ChunkIndex == nextIndex[res.File] {
				if err := sink(res.File, res.Items); err != nil {
					return err
				}
				nextIndex[res.File]++
				if err := flush(res.File); err != nil {
					return err
				}
				continue
			}

			if _, ok := buffers[res.File]; !ok {
				buffers[res.File] = make(map[int][]T)
			}
			buffers[res.File][res.ChunkIndex] = res.Items
		}
	}
}
