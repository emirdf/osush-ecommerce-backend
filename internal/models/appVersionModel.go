package models

import "time"

type AppVersion struct {
	Platform      string
	MinVersion    string
	LatestVersion string
	StoreURL      *string
	UpdatedAt     time.Time
}
