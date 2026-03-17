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
		Folders map[string][]DriveItem `json:"folders"`
		Errors  []DiscoveryError       `json:"errors"`
	} `json:"drive"`
}
