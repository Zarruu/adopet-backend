package middleware

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"adopet-backend/utils"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authDB *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusUnauthorized, "Token autentikasi diperlukan")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Error(c, http.StatusUnauthorized, "Format token tidak valid. Gunakan: Bearer <token>")
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "Token tidak valid atau sudah kadaluarsa")
			c.Abort()
			return
		}

		tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(tokenString)))
		var sessionID int
		err = authDB.QueryRow(
			"SELECT id FROM sessions WHERE token_hash = ? AND user_id = ? AND expires_at > ?",
			tokenHash, claims.UserID, time.Now(),
		).Scan(&sessionID)

		if err != nil {
			utils.Error(c, http.StatusUnauthorized, "Sesi tidak ditemukan atau sudah berakhir. Silakan login kembali")
			c.Abort()
			return
		}

		var isActive bool
		err = authDB.QueryRow("SELECT is_active FROM users WHERE id = ?", claims.UserID).Scan(&isActive)
		if err != nil || !isActive {
			utils.Error(c, http.StatusForbidden, "Akun Anda telah dinonaktifkan")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("token", tokenString)
		c.Next()
	}
}
