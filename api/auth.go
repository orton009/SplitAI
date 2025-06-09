package apiServer

import (
	"splitExpense/expense"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("token")
		if err != nil {
			// Try Authorization header
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				c.AbortWithStatusJSON(401, gin.H{"error": "missing or invalid token"})
				return
			}
		}
		claims, err := expense.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or expired token"})
			return
		}
		if expense.ShouldRefreshToken(claims) {
			user := expense.User{ID: claims.UserID, Name: claims.Name, Email: claims.Email, IsVerified: claims.IsVerified}
			token, _ := expense.GenerateToken(user)
			c.SetCookie("token", token, 3600, "/", "", false, true)
		}
		c.Set("user", claims)
		c.Next()
	}
}
