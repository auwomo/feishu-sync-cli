package export

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/your-org/feishu-sync/internal/config"
	"github.com/your-org/feishu-sync/internal/feishu"
	"github.com/your-org/feishu-sync/internal/manifest"
)

type ErrorEntry struct {
	Time   string `json:"time"`
	Scope  string `json:"scope"`
	Token  string `json:"token"`
	Type   string `json:"type"`
	Path   string `json:"path"`
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

type Puller struct {
	// Counters for human-friendly summaries.
	driveExportedMu sync.Mutex
	driveExported   int
	wikiExportedMu  sync.Mutex
	wikiExported    int
	unsupportedMu   sync.Mutex
	unsupported     int
	errorsCountMu   sync.Mutex
	errorsCount     int
	Client *feishu.Client
	Token  string
	Cfg    *config.Config

	OutDir string // absolute output dir

	sem chan struct{}
	qps chan struct{}

	errorsMu sync.Mutex
	errorsW  *os.File

	manifest manifest.PullManifest
}

func NewPuller(client *feishu.Client, accessToken string, cfg *config.Config, outAbs string, errorsPath string) (*Puller, error) {
	if client == nil {
		return nil, errors.New("client required")
	}
	if accessToken == "" {
		return nil, errors.New("access token required")
	}
	if cfg == nil {
		return nil, errors.New("config required")
	}

	if err := os.MkdirAll(filepath.Join(outAbs, "_meta"), 0o755); err != nil {
		return nil, err
	}

	ew, err := os.OpenFile(errorsPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}

	p := &Puller{Client: client, Token: accessToken, Cfg: cfg, OutDir: outAbs, sem: make(chan struct{}, max(1, cfg.Runtime.Concurrency)), qps: make(chan struct{}, max(1, cfg.Runtime.RateLimitQPS)), errorsW: ew}
	if cfg.Runtime.RateLimitQPS > 0 {
		interval := time.Second / time.Duration(max(1, cfg.Runtime.RateLimitQPS))
		ticker := time.NewTicker(interval)
		go func() {
			for range ticker.C {
				select {
				case p.qps <- struct{}{}:
				default:
				}
			}
		}()
	} else {
		// unlimited
		close(p.qps)
	}

	p.manifest = manifest.PullManifest{OutputDir: outAbs, Mode: cfg.Scope.Mode}
	p.manifest.Drive.Folders = map[string][]manifest.DriveItem{}

	return p, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (p *Puller) Close() error {
	if p.errorsW != nil {
		return p.errorsW.Close()
	}
	return nil
}

func (p *Puller) logError(e ErrorEntry) {
	p.errorsMu.Lock()
	defer p.errorsMu.Unlock()
	_ = json.NewEncoder(p.errorsW).Encode(e)

	p.errorsCountMu.Lock()
	p.errorsCount++
	p.errorsCountMu.Unlock()
}

func (p *Puller) DriveExportedCount() int {
	p.driveExportedMu.Lock()
	defer p.driveExportedMu.Unlock()
	return p.driveExported
}

func (p *Puller) WikiExportedCount() int {
	p.wikiExportedMu.Lock()
	defer p.wikiExportedMu.Unlock()
	return p.wikiExported
}

func (p *Puller) UnsupportedCount() int {
	p.unsupportedMu.Lock()
	defer p.unsupportedMu.Unlock()
	return p.unsupported
}

func (p *Puller) ErrorCount() int {
	p.errorsCountMu.Lock()
	defer p.errorsCountMu.Unlock()
	return p.errorsCount
}

func safeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, string(os.PathSeparator), "_")
	if name == "" {
		name = "untitled"
	}
	return name
}

func shortHash(s string) string {
	h := sha1.Sum([]byte(s))
	return hex.EncodeToString(h[:])[:8]
}

func (p *Puller) driveOutPath(item manifest.DriveItem) (mdOrFilePath string, assetsDir string) {
	relDir := filepath.Join("drive", filepath.FromSlash(filepath.Dir(item.Path)))
	base := safeName(item.Name) + "__" + item.Token
	ext := ".md"
	if item.Type == "file" {
		ext = ""
	}
	mdOrFilePath = filepath.Join(p.OutDir, relDir, base+ext)
	assetsDir = filepath.Join(p.OutDir, relDir, base+".assets")
	return
}

type assetSink struct {
	p         *Puller
	assetsDir string
}

func (s *assetSink) AddAsset(kind, token, suggestedName string) (string, error) {
	// ensure assets dir
	if err := os.MkdirAll(s.assetsDir, 0o755); err != nil {
		return "", err
	}
	fname := token
	if suggestedName != "" {
		fname = token + "_" + safeName(suggestedName)
	}
	// best-effort keep extension from suggestedName
	rel := filepath.ToSlash(filepath.Join(filepath.Base(s.assetsDir), fname))

	// download
	if err := s.p.withLimits(func() error {
		_, _, body, err := s.p.Client.DriveDownload(context.Background(), s.p.Token, token)
		if err != nil {
			return err
		}
		defer body.Close()
		outPath := filepath.Join(s.assetsDir, fname)
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, body)
		return err
	}); err != nil {
		return "", err
	}
	return rel, nil
}

func (p *Puller) withLimits(fn func() error) error {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()
	// rate
	if p.qps != nil {
		if _, ok := <-p.qps; ok {
			// token consumed
		}
	}
	return fn()
}

func (p *Puller) ExportDriveItems(ctx context.Context, items []manifest.DriveItem) {
	var wg sync.WaitGroup
	for _, it := range items {
		it := it
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.exportOne(ctx, it)
		}()
	}
	wg.Wait()
}

func (p *Puller) exportOne(ctx context.Context, it manifest.DriveItem) {
	switch it.Type {
	case "docx":
		p.exportDocx(ctx, it)
	case "doc":
		p.exportDoc(ctx, it)
	case "file":
		p.exportFile(ctx, it)
	case "sheet", "bitable":
		p.unsupportedMu.Lock()
		p.unsupported++
		p.unsupportedMu.Unlock()
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "unsupported: export not implemented"})
	default:
		p.unsupportedMu.Lock()
		p.unsupported++
		p.unsupportedMu.Unlock()
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "unsupported type"})
	}
}

func (p *Puller) exportDocx(ctx context.Context, it manifest.DriveItem) {
	outPath, assetsDir := p.driveOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)

	var blocks []feishu.DocxBlock
	err := p.withLimits(func() error {
		b, err := p.Client.DocxAllBlocks(ctx, p.Token, it.Token)
		if err != nil {
			return err
		}
		blocks = b
		return nil
	})
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "docx blocks: " + err.Error()})
		return
	}

	md, err := RenderDocxToMarkdown(it.Token, blocks, &assetSink{p: p, assetsDir: assetsDir})
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "docx render: " + err.Error()})
		return
	}

	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "write md: " + err.Error()})
		return
	}
	p.driveExportedMu.Lock()
	p.driveExported++
	p.driveExportedMu.Unlock()
}

func (p *Puller) exportDoc(ctx context.Context, it manifest.DriveItem) {
	// V1 doc export: store placeholder md + raw json (best effort)
	outPath, _ := p.driveOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)

	rawPath := strings.TrimSuffix(outPath, ".md") + ".raw.json"
	placeholder := fmt.Sprintf("# %s\n\n> Feishu doc(v1) export is not implemented yet. Raw JSON saved next to this file: %s\n", it.Name, filepath.Base(rawPath))
	_ = os.WriteFile(outPath, []byte(placeholder), 0o644)
	// no API in this milestone; leave raw json empty marker
	_ = os.WriteFile(rawPath, []byte("{}\n"), 0o644)
}

func (p *Puller) exportFile(ctx context.Context, it manifest.DriveItem) {
	outPath, _ := p.driveOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)

	// Save binary as <name>__<token><ext-from-name>
	if ext := filepath.Ext(it.Name); ext != "" {
		outPath = outPath + ext
	}

	err := p.withLimits(func() error {
		_, _, body, err := p.Client.DriveDownload(ctx, p.Token, it.Token)
		if err != nil {
			return err
		}
		defer body.Close()
		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, body)
		return err
	})
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "file download: " + err.Error()})
	}
}

func (p *Puller) WriteManifest(path string, m manifest.PullManifest) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
