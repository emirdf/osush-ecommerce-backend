package request

import (
	"errors"
	"net/url"
	"strings"

	"osush/internal/version"
)

type AppVersionRequest struct {
	Platform      string  `json:"platform"`
	MinVersion    string  `json:"min_version"`
	LatestVersion string  `json:"latest_version"`
	StoreURL      *string `json:"store_url"`
}

func (r *AppVersionRequest) Validate() error {
	r.Platform = strings.ToLower(strings.TrimSpace(r.Platform))
	r.MinVersion = strings.TrimSpace(r.MinVersion)
	r.LatestVersion = strings.TrimSpace(r.LatestVersion)

	if !version.IsValidPlatform(r.Platform) {
		return errors.New("platform must be android or ios")
	}
	if !version.IsValid(r.MinVersion) {
		return errors.New("min_version: expected like 1.2.3 or 1.2.3+7")
	}
	if !version.IsValid(r.LatestVersion) {
		return errors.New("latest_version: expected like 1.2.3 or 1.2.3+7")
	}
	if version.IsLower(r.LatestVersion, r.MinVersion) {
		return errors.New("latest_version must not be lower than min_version")
	}

	if r.StoreURL != nil {
		s := strings.TrimSpace(*r.StoreURL)
		if s == "" {
			r.StoreURL = nil
		} else {
			u, err := url.ParseRequestURI(s)
			if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(s) > 500 {
				return errors.New("store_url must be a valid http(s) link")
			}
			r.StoreURL = &s
		}
	}

	return nil
}
