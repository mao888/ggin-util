package gginutil

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

type User struct {
	AppID  string `json:"app_id"`
	UserID string `json:"user_id"`
}

type UserHandler struct {
	BaseHandler
}

func TestName(t *testing.T) {
	handler := &UserHandler{}
	engine := gin.New()
	engine.Use(SignToC(true))
	engine.POST("/info", Encryption(handler.Info, AllEncryption))
	engine.Run(":8001")
}

func (u *UserHandler) Info(c *gin.Context) {
	user := &User{}
	if err := c.Bind(user); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	resp := map[string]interface{}{"name": "zhangsan", "age": "30"}
	u.Success(c, resp)
}
