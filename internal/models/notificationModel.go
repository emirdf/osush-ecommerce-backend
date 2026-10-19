package models

import "time"

type Notification struct {
	ID                int
	TitleDictionaryID int
	TitleTm           string
	TitleRu           string
	BodyDictionaryID  int
	BodyTm            string
	BodyRu            string
	Life              int
	CreatedAt         time.Time

	IsRead bool
}
