package gginutil

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
)

func Test_register(t *testing.T) {
	handler := &Handler{}
	engine := gin.New()
	engine.Any("/info/:time", handler.Info)
	engine.Run(":8888")
}

type InfoBean struct {
	XUser string     `header:"x-user"`
	Query int32      `form:"query" bt-validator:"required"`
	Time  int32      `uri:"time" bt-validator:"required"`
	Name  string     `json:"name"`
	Age   int32      `json:"age"`
	Class []*string  `json:"class" bt-validator:"required"`
	Sub   []*SubBean `json:"sub"`
}

type SubBean struct {
	ID    int      `json:"id" bt-validator:"required"`
	Games []string `json:"games" bt-validator:"required"`
}

type Handler struct {
	BaseHandler
}

func (h *Handler) Info(c *gin.Context) {
	info := &InfoBean{}
	if !h.Bind(c, info, true) {
		return
	}
	errMsg := h.Validator(info)

	fmt.Println(errMsg)
}
