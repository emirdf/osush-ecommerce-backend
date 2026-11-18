package models

import "time"

type Product struct {
	ID                      int
	Name                    string
	SubcategoryID           int
	SubcategoryName         string
	CategoryID              int
	CategoryName            string
	DescriptionDictionaryID *int
	DescriptionTm           string
	DescriptionRu           string
	Price                   float64
	DiscountPrice           *float64
	FeaturesDictionaryID    *int
	FeaturesTm              string
	FeaturesRu              string
	IsPaused                bool
	CreatedAt               time.Time
	UpdatedAt               time.Time
	Images                  []ProductImage
}

type ProductImage struct {
	ID        int
	ProductID int
	ImagePath string
	IsMain    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type IndexProduct struct {
	ID            int
	Name          string
	Price         float64
	DiscountPrice *float64
	MainImage     string
	IsPaused      bool
}
