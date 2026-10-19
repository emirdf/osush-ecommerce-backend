package models

import "time"

type Banner struct {
	ID                      int
	TitleDictionaryID       int
	TitleTm                 string
	TitleRu                 string
	URL                     *string
	DescriptionDictionaryID *int
	DescriptionTm           string
	DescriptionRu           string
	PhoneNumber             *string
	DisplayFrom             time.Time
	DisplayTo               time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
	Images                  []BannerImage
}

type BannerImage struct {
	ID        int
	BannerID  int
	ImagePath string
	IsMain    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
