package usecase

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"osush/internal/config"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func GenerateTokenPair(cfg *config.Config, id int, role string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = id
	claims["role"] = role
	claims["time_now"] = time.Now()

	t, err := token.SignedString([]byte(cfg.JWTKey))
	if err != nil {
		return "", err
	}

	return t, nil
}

func CreateDir(fileType string) (*string, error) {
	uploadDir := filepath.Join(fmt.Sprintf("uploads/%s", fileType))

	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, errors.New("failed create category dir")
	}
	return &uploadDir, nil
}

// safeExtension matches the file extensions kept in generated object names.
var safeExtension = regexp.MustCompile(`^\.[a-z0-9]{1,5}$`)

// GenerateFileName returns a short, unique object name such as
// "product_<uuid>.jpg". Only the extension of the uploaded file is kept:
// clients re-upload stored images under their stored names, and building on
// those made names grow with every edit until storage rejected them.
func GenerateFileName(
	file *multipart.FileHeader,
	fileType string,
) (*string, error) {

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !safeExtension.MatchString(ext) {
		ext = ""
	}

	fileName := fmt.Sprintf("%s_%s%s", fileType, uuid.NewString(), ext)

	return &fileName, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
