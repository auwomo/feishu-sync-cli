package manifest

type DriveItem struct {
	Token string `json:"token"`
	Name  string `json:"name"`
	Type  string `json:"type"`
}

type DiscoveryError struct {
	Scope   string `json:"scope"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

type PullManifest struct {
	WorkspaceRoot string `json:"workspace_root"`
	OutputDir     string `json:"output_dir"`
	Mode          string `json:"mode"`

	Drive struct {
		// Roots are the starting folder tokens used for discovery.
		Roots []string `json:"roots"`

		Folders map[string][]DriveItem `json:"folders"`

		Summary struct {
			FolderCount int `json:"folder_count"`
			ItemCount   int `json:"item_count"`
		} `json:"summary"`

		Errors []DiscoveryError `json:"errors"`
	} `json:"drive"`
}
