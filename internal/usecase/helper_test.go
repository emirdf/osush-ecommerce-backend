package usecase

import (
	"mime/multipart"
	"regexp"
	"strings"
	"testing"
)

func TestGenerateFileName(t *testing.T) {
	uuidName := regexp.MustCompile(`^main_img_product_[0-9a-f-]{36}`)

	tests := []struct {
		name     string
		filename string
		wantExt  string
	}{
		{"keeps a normal extension", "photo.JPG", ".jpg"},
		{"drops a previously generated prefix", "main_img_product_1794319044556996210_main_img_product_1794318986783738211_phone.webp", ".webp"},
		{"drops unsafe characters", "71GY6PCVxHL._AC_UF1000,1000_QL80_.jpg", ".jpg"},
		{"drops a missing extension", "image", ""},
		{"drops a strange extension", "image.tar,gz", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateFileName(&multipart.FileHeader{Filename: tt.filename}, "main_img_product")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !uuidName.MatchString(*got) {
				t.Errorf("name %q does not start with the type and a uuid", *got)
			}

			if !strings.HasSuffix(*got, tt.wantExt) || len(*got) != len("main_img_product_")+36+len(tt.wantExt) {
				t.Errorf("name %q, want a uuid name ending in %q", *got, tt.wantExt)
			}
		})
	}

	// Re-uploading a stored name must not make it longer.
	first, _ := GenerateFileName(&multipart.FileHeader{Filename: "phone.jpg"}, "main_img_product")
	second, _ := GenerateFileName(&multipart.FileHeader{Filename: *first}, "main_img_product")
	if len(*second) != len(*first) {
		t.Errorf("name grew on re-upload: %d -> %d", len(*first), len(*second))
	}
}
