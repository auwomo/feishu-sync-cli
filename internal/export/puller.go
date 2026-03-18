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
	"github.com/your-org/feishu-sync/internal/meta"
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

	ledger *meta.Ledger
	runID  string

	manifest manifest.PullManifest
}

func NewPuller(client *feishu.Client, accessToken string, cfg *config.Config, outAbs string, errorsPath string, ledger *meta.Ledger, runID string) (*Puller, error) {
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

	p := &Puller{Client: client, Token: accessToken, Cfg: cfg, OutDir: outAbs, sem: make(chan struct{}, max(1, cfg.Runtime.Concurrency)), qps: make(chan struct{}, max(1, cfg.Runtime.RateLimitQPS)), errorsW: ew, ledger: ledger, runID: runID}
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
		_ = p.errorsW.Close()
	}
	if p.ledger != nil {
		_ = p.ledger.Close()
	}
	return nil
}

func (p *Puller) ledgerWrite(e meta.Entry) {
	if p.ledger == nil {
		return
	}
	p.ledger.Write(e)
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

	tmr := meta.StartTimer()
	err := s.p.withLimits(func() error {
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
	})
	st, et, dur := tmr.Done()
	if err != nil {
		s.p.ledgerWrite(meta.Entry{RunID: s.p.runID, ResourceType: "drive", ResourceToken: token, Action: "download_asset", Status: "error", StartedAt: st, EndedAt: et, DurationMS: dur, ErrorMessage: meta.Trunc(err.Error(), 500)})
		return "", err
	}
	s.p.ledgerWrite(meta.Entry{RunID: s.p.runID, ResourceType: "drive", ResourceToken: token, Action: "download_asset", Status: "ok", StartedAt: st, EndedAt: et, DurationMS: dur})
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
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "export", Status: "skipped", StartedAt: meta.NowRFC3339(), EndedAt: meta.NowRFC3339(), DurationMS: 0, ErrorCode: "unsupported", ErrorMessage: "unsupported: export not implemented"})
	default:
		p.unsupportedMu.Lock()
		p.unsupported++
		p.unsupportedMu.Unlock()
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "unsupported type"})
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "export", Status: "skipped", StartedAt: meta.NowRFC3339(), EndedAt: meta.NowRFC3339(), DurationMS: 0, ErrorCode: "unsupported", ErrorMessage: "unsupported type"})
	}
}

func (p *Puller) exportDocx(ctx context.Context, it manifest.DriveItem) {
	outPath, assetsDir := p.driveOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)

	// discovery/export key path: download blocks -> render -> write
	{
		tmr := meta.StartTimer()
		var blocks []feishu.DocxBlock
		err := p.withLimits(func() error {
			b, err := p.Client.DocxAllBlocks(ctx, p.Token, it.Token)
			if err != nil {
				return err
			}
			blocks = b
			return nil
		})
		st, et, dur := tmr.Done()
		if err != nil {
			p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "docx blocks: " + err.Error()})
			p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "download_blocks", Status: "error", StartedAt: st, EndedAt: et, DurationMS: dur, ErrorMessage: meta.Trunc(err.Error(), 500)})
			return
		}
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "download_blocks", Status: "ok", StartedAt: st, EndedAt: et, DurationMS: dur})

		// render
		tmr2 := meta.StartTimer()
		md, err := RenderDocxToMarkdown(it.Token, blocks, &assetSink{p: p, assetsDir: assetsDir})
		st2, et2, dur2 := tmr2.Done()
		if err != nil {
			p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "docx render: " + err.Error()})
			p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "convert", Status: "error", StartedAt: st2, EndedAt: et2, DurationMS: dur2, ErrorMessage: meta.Trunc(err.Error(), 500)})
			return
		}
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "convert", Status: "ok", StartedAt: st2, EndedAt: et2, DurationMS: dur2, Bytes: int64(len(md))})

		// write
		tmr3 := meta.StartTimer()
		err = os.WriteFile(outPath, []byte(md), 0o644)
		st3, et3, dur3 := tmr3.Done()
		if err != nil {
			p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "write md: " + err.Error()})
			p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "write", Status: "error", StartedAt: st3, EndedAt: et3, DurationMS: dur3, ErrorMessage: meta.Trunc(err.Error(), 500)})
			return
		}
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "write", Status: "ok", StartedAt: st3, EndedAt: et3, DurationMS: dur3, Bytes: fileSize(outPath)})

		p.driveExportedMu.Lock()
		p.driveExported++
		p.driveExportedMu.Unlock()
	}
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

	p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "export", Status: "skipped", StartedAt: meta.NowRFC3339(), EndedAt: meta.NowRFC3339(), DurationMS: 0, ErrorCode: "unimplemented", ErrorMessage: "doc(v1) export not implemented"})
	_ = ctx
}

func (p *Puller) exportFile(ctx context.Context, it manifest.DriveItem) {
	outPath, _ := p.driveOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)

	// Save binary as <name>__<token><ext-from-name>
	if ext := filepath.Ext(it.Name); ext != "" {
		outPath = outPath + ext
	}

	tmr := meta.StartTimer()
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
	st, et, dur := tmr.Done()
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "drive", Token: it.Token, Type: it.Type, Path: it.Path, Name: it.Name, Reason: "file download: " + err.Error()})
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "download", Status: "error", StartedAt: st, EndedAt: et, DurationMS: dur, ErrorMessage: meta.Trunc(err.Error(), 500)})
		return
	}
	p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "drive", ResourceToken: it.Token, Action: "download", Status: "ok", StartedAt: st, EndedAt: et, DurationMS: dur, Bytes: fileSize(outPath)})

	p.driveExportedMu.Lock()
	p.driveExported++
	p.driveExportedMu.Unlock()
}

func fileSize(path string) int64 {
	st, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return st.Size()
}

func (p *Puller) WriteManifest(path string, m manifest.PullManifest) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
