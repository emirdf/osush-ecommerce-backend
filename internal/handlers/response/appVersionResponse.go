package response

type AppVersionResponse struct {
	Platform      string  `json:"platform"`
	MinVersion    string  `json:"min_version"`
	LatestVersion string  `json:"latest_version"`
	StoreURL      *string `json:"store_url"`
	UpdatedAt     string  `json:"updated_at"`
	// Filled only when the app sends ?version=: "force" | "optional" | "none".
	Update string `json:"update,omitempty"`
}

type AppVersionListResponse struct {
	AppVersions []AppVersionResponse `json:"app_versions"`
}
