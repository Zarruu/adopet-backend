package middleware

import (
	"net/http"

	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			utils.Error(c, http.StatusUnauthorized, "Autentikasi diperlukan")
			c.Abort()
			return
		}

		role, ok := userRole.(string)
		if !ok {
			utils.Error(c, http.StatusInternalServerError, "Kesalahan internal server")
			c.Abort()
			return
		}

		for _, allowedRole := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		utils.Error(c, http.StatusForbidden, "Anda tidak memiliki izin untuk mengakses resource ini")
		c.Abort()
	}
}
