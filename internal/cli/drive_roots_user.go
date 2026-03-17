package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/your-org/feishu-sync/internal/discovery"
)

func runDriveRootsUser(ctx context.Context, chdir, configPath string, out io.Writer) error {
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
	roots, err := discovery.DiscoverUserDriveRoots(ctx, client, token)
	if err != nil {
		return err
	}
	fmt.Fprintln(out, "Feishu Drive roots (user mode)")
	for _, r := range roots {
		fmt.Fprintln(out, r)
	}
	if len(roots) == 0 {
		fmt.Fprintln(out, "(no roots found)")
	}
	return nil
}
