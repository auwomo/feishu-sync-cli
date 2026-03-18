package export

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/your-org/feishu-sync/internal/feishu"
	"github.com/your-org/feishu-sync/internal/manifest"
	"github.com/your-org/feishu-sync/internal/meta"
)

func (p *Puller) wikiOutPath(it manifest.WikiItem) (mdOrFilePath string, assetsDir string) {
	spaceDir := safeName(it.SpaceName)
	if spaceDir == "" || spaceDir == "untitled" {
		spaceDir = it.SpaceID
	}
	relDir := filepath.Join("wiki", spaceDir, filepath.FromSlash(it.Path))
	id := it.NodeToken
	if id == "" {
		id = it.ObjToken
	}
	base := safeName(it.Title) + "__" + id
	ext := ".md"
	if it.ObjType == "file" {
		ext = ""
	}
	mdOrFilePath = filepath.Join(p.OutDir, relDir, base+ext)
	assetsDir = filepath.Join(p.OutDir, relDir, base+".assets")
	return
}

func (p *Puller) ExportWikiItems(ctx context.Context, items []manifest.WikiItem) {
	for _, it := range items {
		it := it
		p.exportWikiOne(ctx, it)
	}
}

func (p *Puller) exportWikiOne(ctx context.Context, it manifest.WikiItem) {
	switch it.ObjType {
	case "docx":
		p.exportWikiDocx(ctx, it)
	case "file":
		p.exportWikiFile(ctx, it)
	case "doc":
		p.exportWikiDoc(ctx, it)
	case "sheet", "bitable":
		p.unsupportedMu.Lock()
		p.unsupported++
		p.unsupportedMu.Unlock()
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "unsupported: export not implemented"})
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "export", Status: "skipped", StartedAt: meta.NowRFC3339(), EndedAt: meta.NowRFC3339(), DurationMS: 0, ErrorCode: "unsupported", ErrorMessage: "unsupported: export not implemented"})
	default:
		p.unsupportedMu.Lock()
		p.unsupported++
		p.unsupportedMu.Unlock()
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "unsupported type"})
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "export", Status: "skipped", StartedAt: meta.NowRFC3339(), EndedAt: meta.NowRFC3339(), DurationMS: 0, ErrorCode: "unsupported", ErrorMessage: "unsupported type"})
	}
}

func (p *Puller) exportWikiDocx(ctx context.Context, it manifest.WikiItem) {
	outPath, assetsDir := p.wikiOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)

	// download blocks
	var blocks []feishu.DocxBlock
	tmr := meta.StartTimer()
	err := p.withLimits(func() error {
		b, err := p.Client.DocxAllBlocks(ctx, p.Token, it.ObjToken)
		if err != nil {
			return err
		}
		blocks = b
		return nil
	})
	st, et, dur := tmr.Done()
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "docx blocks: " + err.Error()})
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "download_blocks", Status: "error", StartedAt: st, EndedAt: et, DurationMS: dur, ErrorMessage: meta.Trunc(err.Error(), 500)})
		return
	}
	p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "download_blocks", Status: "ok", StartedAt: st, EndedAt: et, DurationMS: dur})

	// render
	tmr2 := meta.StartTimer()
	md, err := RenderDocxToMarkdown(it.ObjToken, blocks, &assetSink{p: p, assetsDir: assetsDir})
	st2, et2, dur2 := tmr2.Done()
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "docx render: " + err.Error()})
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "convert", Status: "error", StartedAt: st2, EndedAt: et2, DurationMS: dur2, ErrorMessage: meta.Trunc(err.Error(), 500)})
		return
	}
	p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "convert", Status: "ok", StartedAt: st2, EndedAt: et2, DurationMS: dur2, Bytes: int64(len(md))})

	// write
	tmr3 := meta.StartTimer()
	err = os.WriteFile(outPath, []byte(md), 0o644)
	st3, et3, dur3 := tmr3.Done()
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "write md: " + err.Error()})
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "write", Status: "error", StartedAt: st3, EndedAt: et3, DurationMS: dur3, ErrorMessage: meta.Trunc(err.Error(), 500)})
		return
	}
	p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "write", Status: "ok", StartedAt: st3, EndedAt: et3, DurationMS: dur3, Bytes: fileSize(outPath)})

	p.wikiExportedMu.Lock()
	p.wikiExported++
	p.wikiExportedMu.Unlock()
}

func (p *Puller) exportWikiDoc(ctx context.Context, it manifest.WikiItem) {
	outPath, _ := p.wikiOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)
	rawPath := strings.TrimSuffix(outPath, ".md") + ".raw.json"
	placeholder := fmt.Sprintf("# %s\n\n> Feishu doc(v1) export is not implemented yet. Raw JSON saved next to this file: %s\n", it.Title, filepath.Base(rawPath))
	_ = os.WriteFile(outPath, []byte(placeholder), 0o644)
	_ = os.WriteFile(rawPath, []byte("{}\n"), 0o644)
	p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "export", Status: "skipped", StartedAt: meta.NowRFC3339(), EndedAt: meta.NowRFC3339(), DurationMS: 0, ErrorCode: "unimplemented", ErrorMessage: "doc(v1) export not implemented"})
	_ = ctx
}

func (p *Puller) exportWikiFile(ctx context.Context, it manifest.WikiItem) {
	outPath, _ := p.wikiOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)
	if ext := filepath.Ext(it.Title); ext != "" {
		outPath = outPath + ext
	}

	tmr := meta.StartTimer()
	err := p.withLimits(func() error {
		_, _, body, err := p.Client.DriveDownload(ctx, p.Token, it.ObjToken)
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
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "file download: " + err.Error()})
		p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "download", Status: "error", StartedAt: st, EndedAt: et, DurationMS: dur, ErrorMessage: meta.Trunc(err.Error(), 500)})
		return
	}
	p.ledgerWrite(meta.Entry{RunID: p.runID, ResourceType: "wiki", ResourceToken: it.NodeToken, Action: "download", Status: "ok", StartedAt: st, EndedAt: et, DurationMS: dur, Bytes: fileSize(outPath)})

	p.wikiExportedMu.Lock()
	p.wikiExported++
	p.wikiExportedMu.Unlock()
}
