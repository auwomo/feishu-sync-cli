package cli

import "github.com/your-org/feishu-sync/internal/feishu"

var feishuNewClient = func() *feishu.Client { return feishu.NewClient(nil) }
