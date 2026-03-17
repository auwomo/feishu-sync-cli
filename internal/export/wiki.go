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
	default:
		p.unsupportedMu.Lock()
		p.unsupported++
		p.unsupportedMu.Unlock()
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "unsupported type"})
	}
}

func (p *Puller) exportWikiDocx(ctx context.Context, it manifest.WikiItem) {
	outPath, assetsDir := p.wikiOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)

	var blocks []feishu.DocxBlock
	err := p.withLimits(func() error {
		b, err := p.Client.DocxAllBlocks(ctx, p.Token, it.ObjToken)
		if err != nil {
			return err
		}
		blocks = b
		return nil
	})
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "docx blocks: " + err.Error()})
		return
	}

	md, err := RenderDocxToMarkdown(it.ObjToken, blocks, &assetSink{p: p, assetsDir: assetsDir})
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "docx render: " + err.Error()})
		return
	}

	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "write md: " + err.Error()})
		return
	}
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
}

func (p *Puller) exportWikiFile(ctx context.Context, it manifest.WikiItem) {
	outPath, _ := p.wikiOutPath(it)
	_ = os.MkdirAll(filepath.Dir(outPath), 0o755)
	if ext := filepath.Ext(it.Title); ext != "" {
		outPath = outPath + ext
	}
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
	if err != nil {
		p.logError(ErrorEntry{Time: time.Now().Format(time.RFC3339), Scope: "wiki", Token: it.NodeToken, Type: it.ObjType, Path: it.Path, Name: it.Title, Reason: "file download: " + err.Error()})
		return
	}
	p.wikiExportedMu.Lock()
	p.wikiExported++
	p.wikiExportedMu.Unlock()
}
