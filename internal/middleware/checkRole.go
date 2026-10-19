package middleware

import (
	"osush/internal/config"
	customError "osush/internal/errors"
	"osush/internal/repository"

	"github.com/gin-gonic/gin"
)

func AdminOnly(
	cfg *config.Config,
	clientRepo *repository.ClientRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {

		adminID, role, err := ExtractUserIDFromToken(c, cfg, clientRepo)
		if err != nil {
			customError.HandleError(c, err)
			c.Abort()
			return
		}

		if role == nil || *role != "ADMIN" {
			customError.HandleError(c, customError.ErrForbidden)
			c.Abort()
			return
		}

		c.Set("adminId", adminID)

		c.Next()
	}
}
