package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
	if email == "" {
		return nil
	}

	if len(email) > 255 {
		return errors.New("email must be less than 255 characters")
	}

	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

func ValidatePhone(phone string) (string, error) {
	if phone == "" {
		return "", nil
	}

	digitsOnly := strings.NewReplacer(
		" ", "", "-", "", "(", "", ")", "",
	).Replace(phone)

	pattern := `^\+993\d{8}$`
	matched, err := regexp.MatchString(pattern, digitsOnly)
	if err != nil {
		return "", err
	}

	if !matched {
		return "", errors.New("phone must be a valid number in the format (+993XXXXXXXX)")
	}

	return digitsOnly, nil
}

func ValidateUsername(username string, dateName string) error {
	if username == "" {
		return nil
	}

	if len(username) < 2 {
		return fmt.Errorf("%s must be at least 2 characters", dateName)
	}

	if len(username) > 255 {
		return fmt.Errorf("%s must be less than 255 characters", dateName)
	}

	pattern := `^[a-zA-Z0-9\-\._&#()]+$`
	matched, _ := regexp.MatchString(pattern, username)
	if !matched {
		return fmt.Errorf("%s contains invalid characters", dateName)
	}

	return nil
}

func ValidateFullname(fullname string) error {

	if fullname == "" {
		return nil
	}

	if len(fullname) < 2 {
		return errors.New("full name must be at least 2 characters")
	}

	if len(fullname) > 100 {
		return errors.New("full name must be less than 100 characters")
	}

	pattern := `^[a-zA-Z]+([ '\-][a-zA-Z]+)*$`
	matched, _ := regexp.MatchString(pattern, fullname)
	if !matched {
		return errors.New("full name contains invalid characters")
	}

	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return nil
	}
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if len(password) > 72 {
		return errors.New("password must be less than 72 characters")
	}

	if strings.Contains(password, " ") {
		return errors.New("password must not contain spaces")
	}

	return nil

}

func ValidateDate(dateStr string) error {
	if dateStr == "" {
		return nil
	}

	if len(dateStr) != 10 {
		return errors.New("date must be in DD-MM-YYYY format")
	}

	_, err := time.Parse("02-01-2006", dateStr)
	if err != nil {
		return errors.New("invalid date format")
	}

	return nil
}

func ValidatePagination(offset, limit string) (*int, *int, error) {

	offsetInt, err := strconv.Atoi(offset)
	if err != nil || offsetInt < 1 {
		offsetInt = 1
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil || limitInt < 1 {
		limitInt = 10
	}

	if offsetInt > 10000 {
		return nil, nil, errors.New("page must be less than 10000")
	}

	if limitInt > 100 {
		return nil, nil, errors.New("limit must be less than or equal to 100")
	}

	return &offsetInt, &limitInt, nil
}

func ValidateSearchQuery(search string) error {
	if search == "" {
		return nil
	}

	if len(search) < 2 {
		return errors.New("search query must be at least 2 characters")
	}

	if len(search) > 100 {
		return errors.New("search query must be less than 100 characters")
	}

	pattern := `^[a-zA-Z0-9\s\-\._@#()]+$`
	matched, _ := regexp.MatchString(pattern, search)
	if !matched {
		return errors.New("search query contains invalid characters")
	}

	return nil
}

func ValidateDateRange(startDate, endDate string) error {
	if err := ValidateDate(startDate); err != nil {
		return err
	}

	if err := ValidateDate(endDate); err != nil {
		return err
	}

	if startDate != "" && endDate != "" {
		start, _ := time.Parse("02-01-2006", startDate)
		end, _ := time.Parse("02-01-2006", endDate)

		if start.After(end) {
			return errors.New("startDate cannot be after endDate")
		}
	}

	return nil
}

func ValidateAddress(address string) error {

	if len(address) < 4 {
		return errors.New("address must be at least 4 characters")
	}

	if len(address) > 255 {
		return errors.New("address must be less than 255 characters")
	}

	return nil
}

func ValidateImageSize(size int64) error {
	const maxImageSize = 10 * 1024 * 1024 // 10 MB

	if size > maxImageSize {
		return errors.New("image size must not exceed 10 MB")
	}

	return nil
}

func ValidateImageFormat(contentType string) error {
	switch contentType {
	case "image/jpeg",
		"image/png",
		"image/webp",
		"image/heic",
		"image/heif",
		"image/avif":
		return nil

	default:
		return errors.New("unsupported image format")
	}
}

func ValidateTextSize(text string, object string) error {
	const maxTextSize = 250

	if text == "" {
		return nil
	}

	err := fmt.Sprintf("%s size must not exceed 250 characters", object)

	if utf8.RuneCountInString(text) > maxTextSize {
		return errors.New(err)
	}

	return nil
}

func ValidateMapJSONSize(data map[string]interface{}, fieldName string) error {
	if len(data) == 0 {
		return nil
	}

	maxSize := 250

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("invalid %s", fieldName)
	}

	size := utf8.RuneCount(jsonData)

	if size > maxSize {
		return fmt.Errorf("%s must not exceed %d characters", fieldName, maxSize)
	}

	return nil
}
