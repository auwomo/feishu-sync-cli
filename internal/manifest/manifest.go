package manifest

type DriveItem struct {
	Token string `json:"token"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	// Path is the relative path under the discovered root folder, including this item's name.
	Path string `json:"path"`
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

	Wiki struct {
		Spaces []struct {
			SpaceID string `json:"space_id"`
			Name string `json:"name"`
		} `json:"spaces"`
		Items map[string][]WikiItem `json:"items"`
		Summary struct {
			SpaceCount int `json:"space_count"`
			NodeCount int `json:"node_count"`
			ItemCount int `json:"item_count"`
		} `json:"summary"`
		Errors []DiscoveryError `json:"errors"`
	} `json:"wiki"`
}
