package gginutil

import (
	"net/http"

	"gitlab.ftsview.com/fotoable-go/glog"

	"github.com/gin-gonic/gin"
)

const (
	HeaderError = "Fotoable-Error"
	EmptyString = ""
)

type BaseHandler struct {
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

type Header2C struct {
	AppID      string `header:"Fotoable-App-ID"`
	AppVersion string `header:"Fotoable-App-Version"`
	SdkID      string `header:"Fotoable-Sdk-ID"`
	SdkVersion string `header:"Fotoable-Sdk-Version"`
	Sign       string `header:"sign"`
	TimeStamp  string `header:"timestamp"`
	Random     string `header:"random"`
}

func (h *BaseHandler) bind(c *gin.Context, vo interface{}, bindHeader bool) bool {
	if bindHeader {
		if err := c.ShouldBindHeader(vo); err != nil {
			glog.Errorf(c.Request.Context(), "bind header error: %s", err.Error())
			c.Status(http.StatusBadRequest)
			c.Header(HeaderError, err.Error())
			return false
		}
	}

	if err := c.ShouldBind(vo); err != nil {
		glog.Errorf(c.Request.Context(), "bind body error: %s", err.Error())
		c.Status(http.StatusBadRequest)
		c.Header(HeaderError, err.Error())
		return false
	}
	return true
}

func (h *BaseHandler) success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Data: data})
}

func (h *BaseHandler) fail(c *gin.Context, code int, msg string) {
	if code == 0 {
		code = -1
	}
	if msg == "" {
		msg = "internal server error"
	}
	c.Header(HeaderError, msg)
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg})
}
