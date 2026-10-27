package convert

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Job is one source file in a batch.
type Job struct {
	ID     string
	Path   string
	Format Format // optional per-file override of Options.Format
}

// Result is the outcome of a Job.
type Result struct {
	ID   string
	Name string // output file name, e.g. "photo.webp"
	Data []byte
	Err  error
}

// OutputName derives the output file name from a source path.
func OutputName(src string, f Format, suffix string) string {
	base := filepath.Base(src)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	return base + suffix + f.Ext()
}

// Batch converts jobs on a worker pool and calls onResult (serialized) as each finishes.
func Batch(ctx context.Context, jobs []Job, o Options, suffix string, onResult func(Result)) {
	work := make(chan Job)
	var cb sync.Mutex
	var group sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := range work {
				jo := o
				if j.Format != "" {
					jo.Format = j.Format
				}
				r := Result{ID: j.ID, Name: OutputName(j.Path, jo.Format, suffix)}
				if ctx.Err() != nil {
					r.Err = ctx.Err()
				} else {
					r.Data, r.Err = File(j.Path, jo)
				}
				cb.Lock()
				onResult(r)
				cb.Unlock()
			}
		}()
	}
	for _, j := range jobs {
		work <- j
	}
	close(work)
	group.Wait()
}

// UniqueName returns name, or "name (n).ext" if taken reports a collision.
func UniqueName(name string, taken func(string) bool) string {
	if !taken(name) {
		return name
	}
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for n := 1; ; n++ {
		c := fmt.Sprintf("%s (%d)%s", stem, n, ext)
		if !taken(c) {
			return c
		}
	}
}

// WriteZip writes the successful results into a zip archive, de-duplicating names.
func WriteZip(w io.Writer, results []Result) error {
	zw := zip.NewWriter(w)
	used := map[string]bool{}
	for _, r := range results {
		if r.Err != nil {
			continue
		}
		name := UniqueName(r.Name, func(s string) bool { return used[s] })
		used[name] = true
		f, err := zw.Create(name)
		if err != nil {
			return err
		}
		if _, err := f.Write(r.Data); err != nil {
			return err
		}
	}
	return zw.Close()
}
