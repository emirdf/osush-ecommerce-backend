package usecase

import (
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"osush/internal/config"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
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

func GenerateFileName(
	file *multipart.FileHeader,
	fileType string,
) (*string, error) {

	cleanName := strings.ReplaceAll(filepath.Base(file.Filename), " ", "")
	fileName := fmt.Sprintf("%s_%d_%s", fileType, time.Now().UnixNano(), cleanName)

	return &fileName, nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
