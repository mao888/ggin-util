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

type EncryptionRequest struct {
	Data string `json:"data"`
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"message"`
	Data interface{} `json:"data"`
}

type EncryptionResponse struct {
	Code       int         `json:"code"`
	Msg        string      `json:"message"`
	Encryption bool        `json:"encryption"`
	Data       interface{} `json:"data"`
}

type HeaderToC struct {
	AppID      string `header:"Fotoable-App-ID"`
	AppVersion string `header:"Fotoable-App-Version"`
	SdkID      string `header:"Fotoable-Sdk-ID"`
	SdkVersion string `header:"Fotoable-Sdk-Version"`
	Sign       string `header:"sign"`
	TimeStamp  string `header:"timestamp"`
	Random     string `header:"random"`
}

type HeaderToB struct {
	Cookie        string `header:"Cookie"`
	UserID        int    `header:"USER-ID"`
	XToken        string `header:"X-Token"`
	Authorization string `header:"Authorization"`
}

func (h *BaseHandler) Bind(c *gin.Context, vo interface{}, bindHeader bool) bool {
	if bindHeader {
		if err := c.ShouldBindHeader(vo); err != nil {
			glog.Errorf(c.Request.Context(), "bind header error: %s", err.Error())
			c.Status(http.StatusBadRequest)
			c.Header(HeaderError, err.Error())
			return false
		}
	}
	if c.Request.ContentLength == 0 {
		return true
	}
	if err := c.ShouldBind(vo); err != nil {
		glog.Errorf(c.Request.Context(), "bind body error: %s", err.Error())
		c.Status(http.StatusBadRequest)
		c.Header(HeaderError, err.Error())
		return false
	}
	return true
}

func (h *BaseHandler) Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Data: data})
}

func (h *BaseHandler) Fail(c *gin.Context, code int, msg string) {
	if code == 0 {
		code = -1
	}
	if msg == "" {
		msg = "internal server error"
	}
	c.Header(HeaderError, msg)
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg})
}

func (h *BaseHandler) FailWithData(c *gin.Context, code int, msg string, data interface{}) {
	if code == 0 {
		code = -1
	}
	if msg == "" {
		msg = "internal server error"
	}
	c.Header(HeaderError, msg)
	c.JSON(http.StatusOK, Response{Code: code, Msg: msg, Data: data})
}

func (h *BaseHandler) SuccessEncryption(c *gin.Context, data interface{}, encryption bool) {
	c.JSON(http.StatusOK, EncryptionResponse{Data: data, Encryption: encryption})
}
