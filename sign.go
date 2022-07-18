package gginutil

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"

	"gitlab.ftsview.com/fotoable-go/glog"
	"gitlab.ftsview.com/fotoable-go/gsecret"
	"gitlab.ftsview.com/fotoable-go/gutil"

	"github.com/gin-gonic/gin"
)

type (
	AesKey         string
	EncryptionType int
)

const (
	aesContextKey AesKey = "aes_key"
)

const (
	ReqEncryption  EncryptionType = 1 << iota
	RespEncryption EncryptionType = 1 << iota
	AllEncryption  EncryptionType = ReqEncryption | RespEncryption
)

//responseBodyWriter 拦截返回的Body
type responseBodyWriter struct {
	gin.ResponseWriter
	//body *bytes.Buffer
	ctx context.Context
}

//Write 缓存返回数据
func (r responseBodyWriter) Write(b []byte) (int, error) {
	//r.body.Write(b)
	resp := &EncryptionResponse{}
	if err := gutil.JSON2ObjectE(b, resp); err != nil {
		glog.Error(r.ctx, "response json2Obj error. error: %s", err.Error())
	}
	resp.Encryption = true
	resp.Data = base64.StdEncoding.EncodeToString(
		gutil.AESEncrypt(gutil.Object2JSONByte(resp.Data), []byte(GetAESKey(r.ctx))))
	return r.ResponseWriter.Write(gutil.Object2JSONByte(resp))
}

func SignToC(encryption bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := &HeaderToC{}
		if err := c.ShouldBindHeader(header); err != nil {
			c.Header(HeaderError, err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//获取客户端秘钥
		client, server := gsecret.GetAuthSign(gsecret.GetGameID(header.AppID))
		if client == EmptyString || server == EmptyString {
			c.Header(HeaderError, "client/server token not find.app_id: "+header.AppID)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//验证签名
		if header.Sign == EmptyString {
			c.Header(HeaderError, "sign is nil")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		serverSign := gutil.PramSign([]string{client,
			header.AppID, header.SdkID,
			header.TimeStamp, header.Random})
		if header.Sign != serverSign {
			glog.Debugf(c.Request.Context(), "sign error,client: %s,server: %s", header.Sign, serverSign)
			c.Header(HeaderError, "check sign error")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if encryption {
			aesKey := gutil.PramSign([]string{server, header.AppID, header.SdkID})
			c.Request = c.Request.WithContext(context.WithValue(
				c.Request.Context(), aesContextKey, aesKey))
		}
	}
}

func Encryption(h gin.HandlerFunc, t EncryptionType) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		defer func() {
			if err := recover(); err != nil {
				glog.Errorf(ctx, "recover error: %+v", err)
			}
		}()
		//判断是否请求加密
		if t&ReqEncryption > 0 {
			req := &EncryptionRequest{}
			if err := c.Bind(req); err != nil {
				glog.Errorf(ctx, "request encryption data bind error: %s", err.Error())
				c.Status(http.StatusBadRequest)
				return
			}
			decodeString, err := base64.StdEncoding.DecodeString(req.Data)
			if err != nil {
				glog.Errorf(ctx, "request base64 error: %s", err.Error())
				c.Status(http.StatusBadRequest)
				return
			}
			body := gutil.AESDecrypt(decodeString, []byte(GetAESKey(ctx)))
			if len(body) == 0 {
				glog.Errorf(ctx, "request decrypt data error: %s")
				c.Status(http.StatusBadRequest)
				return
			}
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		//判断响应是否加密
		if t&RespEncryption > 0 {
			c.Writer = &responseBodyWriter{ResponseWriter: c.Writer, ctx: ctx}
		}
		h(c)
	}
}

func GetAESKey(ctx context.Context) string {
	aesKey, ok := ctx.Value(aesContextKey).(string)
	if ok {
		return aesKey
	}
	return EmptyString
}
