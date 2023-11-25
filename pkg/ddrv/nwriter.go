package ddrv

import (
	"bytes"
	"io"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/forscht/ddrv/pkg/breader"
)

// NWriter implements io.WriteCloser.
// It streams data in chunks to Discord server channels using webhook
// NWriter buffers bytes into memory and writes data to discord in parallel
// at the cost of high-memory usage.
// expected memory usage - (chunkSize * number of webhooks) + 20% bytes
type NWriter struct {
	rest      *Rest // Manager where Writer writes data
	chunkSize int   // The maximum Size of a chunk
	onChunk   func(chunk *Chunk)

	mu sync.Mutex
	wg sync.WaitGroup

	closed       bool // Whether the Writer has been closed
	err          error
	chunks       []*Chunk
	pwriter      *io.PipeWriter // PipeWriter for writing the current chunk
	chunkCounter int64
}

func NewNWriter(onChunk func(chunk *Chunk), chunkSize int, rest *Rest) io.WriteCloser {
	reader, writer := io.Pipe()
	w := &NWriter{
		rest:      rest,
		onChunk:   onChunk,
		chunkSize: chunkSize,
		pwriter:   writer,
	}
	go w.startWorkers(breader.New(reader))

	return w
}

func (w *NWriter) Write(p []byte) (int, error) {
	if w.closed {
		return 0, ErrClosed
	}
	if w.err != nil {
		return 0, w.err
	}
	return w.pwriter.Write(p)
}

func (w *NWriter) Close() error {
	if w.closed {
		return ErrAlreadyClosed
	}
	w.closed = true
	if w.pwriter != nil {
		if err := w.pwriter.Close(); err != nil {
			return err
		}
	}
	w.wg.Wait()
	if w.onChunk != nil {
		sort.SliceStable(w.chunks, func(i, j int) bool {
			return w.chunks[i].Start < w.chunks[j].Start
		})
		for _, chunk := range w.chunks {
			w.onChunk(chunk)
		}
	}
	return nil
}

func (w *NWriter) startWorkers(reader io.Reader) {
	concurrency := len(w.rest.channels)
	w.wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer w.wg.Done()
			buff := make([]byte, w.chunkSize)
			for {
				if w.err != nil {
					return
				}
				n, err := reader.Read(buff)
				if n > 0 {
					cIdx := atomic.AddInt64(&w.chunkCounter, 1)
					attachment, werr := w.rest.CreateAttachment(bytes.NewReader(buff[:n]))
					if werr != nil {
						w.err = werr
						return
					}
					w.mu.Lock()
					attachment.Start = cIdx
					w.chunks = append(w.chunks, attachment)
					w.mu.Unlock()
				}
				if err != nil {
					if err != io.EOF {
						w.err = err
					}
					return
				}
			}
		}()
	}
}
