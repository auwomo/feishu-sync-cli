package export

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/your-org/feishu-sync/internal/feishu"
)

// Minimal block types we care about (matches Feishu docx).
const (
	blockPage   = 1
	blockText   = 2
	blockH1     = 3
	blockH2     = 4
	blockH3     = 5
	blockH4     = 6
	blockH5     = 7
	blockH6     = 8
	blockBullet = 12
	blockOrder  = 13
	blockCode   = 14
	blockFile   = 23
	blockImage  = 28
	blockQuote  = 34
)

type AssetSink interface {
	// AddAsset downloads token and writes into assetsDir, returning relative path from md file.
	AddAsset(kind string, token string, suggestedName string) (rel string, err error)
}

func elementsToMD(e []feishu.DocxElement) string {
	var b strings.Builder
	for _, el := range e {
		switch {
		case el.TextRun != nil:
			ct := el.TextRun.Content
			sty := el.TextRun.Style
			if strings.TrimSpace(ct) == "" {
				continue
			}
			if sty.Link != nil && sty.Link.URL != "" {
				u := sty.Link.URL
				if du, err := url.QueryUnescape(u); err == nil {
					u = du
				}
				ct = fmt.Sprintf("[%s](%s)", ct, u)
			}
			if sty.Strikethrough {
				ct = "~~" + strings.TrimSpace(ct) + "~~"
			}
			if sty.Bold {
				ct = "**" + strings.TrimSpace(ct) + "**"
			}
			if sty.Italic {
				ct = "*" + strings.TrimSpace(ct) + "*"
			}
			if sty.InlineCode {
				ct = "`" + ct + "`"
			}
			if sty.Underline {
				ct = "<u>" + ct + "</u>"
			}
			b.WriteString(ct)
		case el.Equation != nil:
			b.WriteString("$")
			b.WriteString(strings.TrimSpace(el.Equation.Content))
			b.WriteString("$")
		case el.MentionDoc != nil:
			u := el.MentionDoc.URL
			if du, err := url.QueryUnescape(u); err == nil {
				u = du
			}
			b.WriteString(fmt.Sprintf("[%s](%s)", el.MentionDoc.Title, u))
		}
	}
	return b.String()
}

func blocksByParent(blocks []feishu.DocxBlock) map[string][]feishu.DocxBlock {
	m := map[string][]feishu.DocxBlock{}
	for _, b := range blocks {
		m[b.ParentID] = append(m[b.ParentID], b)
	}
	return m
}

func RenderDocxToMarkdown(rootID string, blocks []feishu.DocxBlock, sink AssetSink) (string, error) {
	byParent := blocksByParent(blocks)
	// Some docs have a synthetic root. We'll render from rootID's children.
	var out strings.Builder

	var render func(parent string, indent string, orderedIndex *int) error
	render = func(parent string, indent string, orderedIndex *int) error {
		kids := byParent[parent]
		contType := 0
		localOrder := orderedIndex
		for _, blk := range kids {
			if contType != 0 && contType != blk.BlockType {
				out.WriteString("\n")
				contType = 0
			}
			switch blk.BlockType {
			case blockPage:
				if blk.Page != nil {
					out.WriteString("# ")
					out.WriteString(elementsToMD(blk.Page.Elements))
					out.WriteString("\n\n")
				}
				if err := render(blk.BlockID, indent, nil); err != nil {
					return err
				}
			case blockText:
				contType = blockText
				if blk.Text != nil {
					out.WriteString(indent)
					out.WriteString(elementsToMD(blk.Text.Elements))
					out.WriteString("\n\n")
				}
				if err := render(blk.BlockID, indent, nil); err != nil {
					return err
				}
			case blockH1, blockH2, blockH3, blockH4, blockH5, blockH6:
				level := map[int]string{blockH1: "#", blockH2: "##", blockH3: "###", blockH4: "####", blockH5: "#####", blockH6: "######"}[blk.BlockType]
				var els []feishu.DocxElement
				switch blk.BlockType {
				case blockH1:
					if blk.Heading1 != nil {
						els = blk.Heading1.Elements
					}
				case blockH2:
					if blk.Heading2 != nil {
						els = blk.Heading2.Elements
					}
				case blockH3:
					if blk.Heading3 != nil {
						els = blk.Heading3.Elements
					}
				case blockH4:
					if blk.Heading4 != nil {
						els = blk.Heading4.Elements
					}
				case blockH5:
					if blk.Heading5 != nil {
						els = blk.Heading5.Elements
					}
				case blockH6:
					if blk.Heading6 != nil {
						els = blk.Heading6.Elements
					}
				}
				out.WriteString(level + " " + elementsToMD(els) + "\n\n")
				if err := render(blk.BlockID, indent, nil); err != nil {
					return err
				}
			case blockBullet:
				contType = blockBullet
				out.WriteString(indent + "* ")
				if blk.Bullet != nil {
					out.WriteString(elementsToMD(blk.Bullet.Elements))
				}
				out.WriteString("\n")
				if err := render(blk.BlockID, indent+"  ", nil); err != nil {
					return err
				}
			case blockOrder:
				contType = blockOrder
				if localOrder == nil {
					i := 1
					localOrder = &i
				}
				out.WriteString(fmt.Sprintf("%s%d. ", indent, *localOrder))
				if blk.Ordered != nil {
					out.WriteString(elementsToMD(blk.Ordered.Elements))
				}
				out.WriteString("\n")
				*localOrder = *localOrder + 1
				if err := render(blk.BlockID, indent+"    ", nil); err != nil {
					return err
				}
			case blockCode:
				if blk.Code != nil {
					out.WriteString("\n```\n")
					out.WriteString(elementsToMD(blk.Code.Elements))
					out.WriteString("\n```\n\n")
				}
				if err := render(blk.BlockID, indent, nil); err != nil {
					return err
				}
			case blockQuote:
				// Best effort: render children with > prefix
				out.WriteString("\n")
				beforeLen := out.Len()
				if err := render(blk.BlockID, indent+"> ", nil); err != nil {
					return err
				}
				if out.Len() == beforeLen {
					out.WriteString(indent + ">\n")
				}
				out.WriteString("\n")
			case blockImage:
				if blk.Image != nil {
					rel, err := sink.AddAsset("image", blk.Image.Token, "")
					if err != nil {
						return err
					}
					out.WriteString("![](" + filepath.ToSlash(rel) + ")\n\n")
				}
			case blockFile:
				if blk.File != nil {
					rel, err := sink.AddAsset("file", blk.File.Token, blk.File.Name)
					if err != nil {
						return err
					}
					out.WriteString("[" + blk.File.Name + "](" + filepath.ToSlash(rel) + ")\n\n")
				}
			default:
				// ignore unsupported blocks
			}
		}
		return nil
	}

	if err := render(rootID, "", nil); err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()) + "\n", nil
}
