package cli

import (
	"fmt"
	"io"
)

func runDriveRoots(out io.Writer) error {
	fmt.Fprintln(out, "Feishu Drive roots")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "This CLI currently uses tenant_access_token (bot mode).")
	fmt.Fprintln(out, "In tenant-only mode, Feishu Drive does not provide a universal \"root folder\" that can be enumerated")
	fmt.Fprintln(out, "across all users. You must start discovery from one or more known folder tokens.")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "How to get a folder token:")
	fmt.Fprintln(out, "  1) In Feishu, open the folder in Cloud Space.")
	fmt.Fprintln(out, "  2) Copy the URL and extract the folder_token parameter.")
	fmt.Fprintln(out, "  3) Put it in .feishu-sync/config.yaml under scope.drive_folder_tokens.")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Then you can explore recursively:")
	fmt.Fprintln(out, "  feishu-sync drive ls --folder <folder_token> --depth 2")
	return nil
}
