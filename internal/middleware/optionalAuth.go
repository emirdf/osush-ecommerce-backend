package middleware

import (
	"errors"
	"osush/internal/config"
	customError "osush/internal/errors"
	"osush/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func OptionalAuthMiddleware(
	cfg *config.Config,
	clientRepo *repository.ClientRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {

		clientID, _, err := ExtractOptionalUserIDFromToken(
			c,
			cfg,
			clientRepo,
		)

		if err != nil {
			customError.HandleError(c, err)
			c.Abort()
			return
		}

		c.Set("clientId", clientID)

		c.Next()
	}
}

func ExtractOptionalUserIDFromToken(
	c *gin.Context,
	cfg *config.Config,
	clientRepo *repository.ClientRepository,
) (int, *string, error) {

	ctx := c.Request.Context()

	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {
		return -1, nil, nil
	}

	if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
		return -1, nil, customError.ErrBadToken
	}

	tokenString := authHeader[7:]

	if tokenString == "" {
		return -1, nil, customError.ErrBadToken
	}

	exist, err := clientRepo.CheckInBlacklist(ctx, tokenString)
	if err != nil {
		return -1, nil, errors.New("failed to check client")
	}

	if exist {
		return -1, nil, customError.ErrTokenInBlacklist
	}

	token, err := ParseToken(cfg, tokenString)
	if err != nil {
		return -1, nil, customError.ErrBadToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return -1, nil, errors.New("unable to parse claims")
	}

	clientID, ok := claims["id"].(float64)
	if !ok {
		return -1, nil, errors.New("missing 'id' claim in token")
	}

	clientRole, ok := claims["role"].(string)
	if !ok {
		return -1, nil, errors.New("missing 'role' claim in token")
	}

	return int(clientID), &clientRole, nil
}
