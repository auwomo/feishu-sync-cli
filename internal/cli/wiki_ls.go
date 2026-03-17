package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/your-org/feishu-sync/internal/discovery"
)

type wikiLsOptions struct {
	SpaceID string
	Nodes   bool
}

func runWikiLs(ctx context.Context, chdir, configPath string, opt wikiLsOptions, w io.Writer) error {
	ws, cfg, err := loadWorkspaceAndConfig(chdir, configPath)
	if err != nil {
		return err
	}
	secret, err := resolveAppSecret(ws, cfg)
	if err != nil {
		return err
	}
	token, _, err := resolveAccessToken(ctx, ws.Path(), cfg, secret)
	if err != nil {
		return err
	}
	client := feishuNewClient()

	spaces, err := discovery.DiscoverWikiSpaces(ctx, client, token)
	if err != nil {
		return err
	}
	for _, sp := range spaces {
		if opt.SpaceID != "" && sp.SpaceID != opt.SpaceID {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\n", sp.SpaceID, sp.Name)
		if opt.Nodes {
			items, _ := discovery.DiscoverWikiTree(ctx, client, token, sp.SpaceID)
			for _, it := range items {
				fmt.Fprintf(w, "  %s/%s (%s:%s)\n", it.Path, it.Title, it.ObjType, it.ObjToken)
			}
		}
	}
	return nil
}

func parseWikiLsFlags(args []string) (wikiLsOptions, []string, error) {
	fs := flag.NewFlagSet("wiki ls", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	space := fs.String("space", "", "wiki space id")
	nodes := fs.Bool("nodes", false, "also list nodes")
	if err := fs.Parse(args); err != nil {
		return wikiLsOptions{}, nil, err
	}
	return wikiLsOptions{SpaceID: *space, Nodes: *nodes}, fs.Args(), nil
}
