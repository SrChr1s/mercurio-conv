package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"mercurio-conversor/internal/convert"
)

// FileInfo describes a queued source image for the frontend.
type FileInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	Format string `json:"format"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Thumb  string `json:"thumb"` // data URL
}

// Request is what the frontend sends to Convert.
type Request struct {
	IDs     []string          `json:"ids"`
	Options convert.Options   `json:"options"`
	Formats map[string]string `json:"formats"` // per-file format overrides, id -> format
	Output  string            `json:"output"`  // "folder" | "zip"
	OutDir  string            `json:"outDir"`
	Suffix  string            `json:"suffix"`
}

// Progress is emitted as "convert:progress" for each finished file.
type Progress struct {
	ID      string `json:"id"`
	Status  string `json:"status"` // "done" | "error"
	OutName string `json:"outName,omitempty"`
	OutSize int    `json:"outSize,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Summary is emitted as "convert:done".
type Summary struct {
	OK        int  `json:"ok"`
	Failed    int  `json:"failed"`
	Cancelled bool `json:"cancelled"`
}

// App holds the Wails-bound state.
type App struct {
	ctx context.Context

	mu      sync.Mutex
	files   map[string]string         // id -> source path
	results map[string]convert.Result // last zip-mode results
	cancel  context.CancelFunc
	nextID  int
}

func NewApp() *App {
	return &App{files: map[string]string{}, results: map[string]convert.Result{}}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

var imageExts = "*.jpg;*.jpeg;*.png;*.gif;*.webp;*.bmp;*.tif;*.tiff;*.svg"

// PickFiles opens a native multi-select dialog and queues the chosen images.
func (a *App) PickFiles() ([]FileInfo, error) {
	paths, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Añadir imágenes",
		Filters: []runtime.FileFilter{{DisplayName: "Imágenes", Pattern: imageExts}},
	})
	if err != nil {
		return nil, err
	}
	return a.AddPaths(paths), nil
}

// AddPaths queues files (skipping duplicates and anything that isn't a decodable image).
func (a *App) AddPaths(paths []string) []FileInfo {
	out := []FileInfo{}
	for _, p := range paths {
		st, err := os.Stat(p)
		if err != nil || st.IsDir() {
			continue
		}
		// Reserve the slot first: a drop can arrive twice (native path + DOM fallback).
		a.mu.Lock()
		dup := false
		for _, q := range a.files {
			if q == p {
				dup = true
				break
			}
		}
		if dup {
			a.mu.Unlock()
			continue
		}
		a.nextID++
		id := fmt.Sprintf("f%d", a.nextID)
		a.files[id] = p
		a.mu.Unlock()

		info, err := convert.Probe(p)
		if err != nil {
			a.Remove([]string{id})
			continue
		}
		fi := FileInfo{
			ID: id, Name: filepath.Base(p), Size: st.Size(),
			Format: string(info.Format), Width: info.Width, Height: info.Height,
		}
		if t, err := convert.File(p, convert.Options{Format: convert.JPEG, Quality: 70, MaxWidth: 96, MaxHeight: 96}); err == nil {
			fi.Thumb = "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(t)
		}
		out = append(out, fi)
	}
	return out
}

// Remove forgets queued files.
func (a *App) Remove(ids []string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, id := range ids {
		delete(a.files, id)
		delete(a.results, id)
	}
}

// PickOutputDir opens a native folder chooser.
func (a *App) PickOutputDir() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Elegir carpeta de destino"})
}

// Convert starts a batch in the background; progress arrives as events.
func (a *App) Convert(req Request) error {
	if req.Output == "folder" && req.OutDir == "" {
		return errors.New("elige primero una carpeta de destino")
	}
	a.mu.Lock()
	if a.cancel != nil {
		a.mu.Unlock()
		return errors.New("ya hay una conversión en curso")
	}
	var jobs []convert.Job
	for _, id := range req.IDs {
		if p, ok := a.files[id]; ok {
			job := convert.Job{ID: id, Path: p}
			if f, err := convert.ParseFormat(req.Formats[id]); err == nil && f != convert.SVG {
				job.Format = f
			}
			jobs = append(jobs, job)
		}
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.cancel = cancel
	a.results = map[string]convert.Result{}
	a.mu.Unlock()

	go func() {
		var sum Summary
		taken := map[string]bool{}
		convert.Batch(ctx, jobs, req.Options, req.Suffix, func(r convert.Result) {
			p := Progress{ID: r.ID, Status: "done", OutName: r.Name, OutSize: len(r.Data)}
			if r.Err == nil && req.Output == "folder" {
				r.Name = convert.UniqueName(r.Name, func(n string) bool {
					if taken[n] {
						return true
					}
					_, err := os.Stat(filepath.Join(req.OutDir, n))
					return err == nil
				})
				taken[r.Name] = true
				p.OutName = r.Name
				r.Err = os.WriteFile(filepath.Join(req.OutDir, r.Name), r.Data, 0o644)
			}
			if r.Err != nil {
				p.Status, p.Error, p.OutSize = "error", cleanErr(r.Err), 0
				sum.Failed++
			} else {
				sum.OK++
				if req.Output == "zip" {
					a.mu.Lock()
					a.results[r.ID] = r
					a.mu.Unlock()
				}
			}
			runtime.EventsEmit(a.ctx, "convert:progress", p)
		})
		sum.Cancelled = ctx.Err() != nil
		a.mu.Lock()
		a.cancel = nil
		a.mu.Unlock()
		cancel()
		runtime.EventsEmit(a.ctx, "convert:done", sum)
	}()
	return nil
}

// Cancel stops the running batch after in-flight files finish.
func (a *App) Cancel() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancel != nil {
		a.cancel()
	}
}

// SaveZip asks where to save, then zips the results of the last zip-mode batch.
// It returns the chosen path, or "" if the user cancelled.
func (a *App) SaveZip() (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Guardar ZIP",
		DefaultFilename: "converted-images.zip",
		Filters:         []runtime.FileFilter{{DisplayName: "Archivo ZIP", Pattern: "*.zip"}},
	})
	if err != nil || path == "" {
		return "", err
	}
	if !strings.HasSuffix(strings.ToLower(path), ".zip") {
		path += ".zip"
	}
	a.mu.Lock()
	results := make([]convert.Result, 0, len(a.results))
	for _, r := range a.results {
		results = append(results, r)
	}
	a.mu.Unlock()
	if len(results) == 0 {
		return "", errors.New("no hay nada que guardar")
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return path, convert.WriteZip(f, results)
}

func cleanErr(err error) string {
	if errors.Is(err, context.Canceled) {
		return "cancelado"
	}
	return err.Error()
}
