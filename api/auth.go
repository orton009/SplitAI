package apiServer

import (
	"errors"
	"splitExpense/expense"
	"strings"

	"github.com/gin-gonic/gin"
)

const CtxUserId = "user_id"
const CtxUser = "user"

func CtxGetUserId(c *gin.Context) (string, error) {
	value, exists := c.Get(CtxUserId)
	if !exists {
		return "", errors.New("user Id not present in gin context")
	}

	userId, ok := value.(string)
	if !ok {
		return "", errors.New("user Id in gin context is not a string")
	}

	return userId, nil
}

func CtxGetUser(c *gin.Context) (*expense.User, error) {
	value, exists := c.Get(CtxUserId)
	if !exists {
		return nil, errors.New("user not present in gin context")
	}

	user, ok := value.(expense.User)
	if !ok {
		return nil, errors.New("value inside gin context is not a user")
	}

	return &user, nil
}

func Authenticate(c *gin.Context) {
	tokenStr, err := c.Cookie("token")
	if err != nil {
		// Try Authorization header
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			c.AbortWithError(401, err)
			return
		}
	}
	claims, err := expense.ParseToken(tokenStr)
	if err != nil {
		c.AbortWithError(401, err)
	}
	user := expense.User{ID: claims.UserID, Name: claims.Name, Email: claims.Email, IsVerified: claims.IsVerified}
	if expense.ShouldRefreshToken(claims) {
		token, _ := expense.GenerateToken(user)
		c.SetCookie("token", token, 3600, "/", "", false, true)
	}
	c.Set(CtxUser, user)
	c.Set(CtxUserId, user.ID)
	c.Next()
}
