package middleware

import (
	"errors"
	"fmt"

	"osush/internal/config"
	"osush/internal/repository"

	customError "osush/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthMiddleware(
	cfg *config.Config,
	clientRepo *repository.ClientRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {

		clientID, clientRole, err := ExtractUserIDFromToken(c, cfg, clientRepo)
		if err != nil {
			customError.HandleError(c, err)
			c.Abort()
			return
		}

		if *clientRole != "CLIENT" {
			customError.HandleError(c, customError.ErrUnauthorized)
			c.Abort()
			return
		}

		// Tokens outlive deleted accounts (on other devices, for example).
		exists, err := clientRepo.ExistsByID(c.Request.Context(), clientID)
		if err != nil {
			customError.HandleError(c, errors.New("failed to check client"))
			c.Abort()
			return
		}

		if !exists {
			customError.HandleError(c, customError.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set("clientId", clientID)

		c.Next()
	}
}

func ExtractUserIDFromToken(
	c *gin.Context,
	cfg *config.Config,
	clientRepo *repository.ClientRepository) (int, *string, error) {
	ctx := c.Request.Context()

	tokenString := c.GetHeader("Authorization")

	if tokenString == "" {
		return -1, nil, customError.ErrMissingToken
	}

	if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
		return 0, nil, customError.ErrBadToken
	}

	tokenString = tokenString[7:]

	exist, err := clientRepo.CheckInBlacklist(ctx, tokenString)

	if err != nil {
		return 0, nil, errors.New("failed to check client")
	}

	if exist {
		return 0, nil, customError.ErrTokenInBlacklist
	}

	token, err := ParseToken(cfg, tokenString)
	if err != nil {
		return 0, nil, customError.ErrBadToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, nil, errors.New("unable to parse claims")
	}

	clientID, ok := claims["id"].(float64)
	if !ok {
		return 0, nil, errors.New("missing 'id' claim in token")
	}

	clientRole, ok := claims["role"].(string)
	if !ok {
		return 0, nil, errors.New("missing 'id' claim in token")
	}

	return int(clientID), &clientRole, nil
}

func ParseToken(cfg *config.Config, jwtToken string) (*jwt.Token, error) {

	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf(
				"unexpected signing method: %v",
				token.Header["alg"],
			)
		}

		return []byte(cfg.JWTKey), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
