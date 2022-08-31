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

	HeaderXRequestID = "X-Request-ID"
	HeaderUserID     = "X-USER-ID"
	HeaderUserName   = "X-USER-Name"
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

//SignToC 验证ToC服务的签名及生成AES秘钥
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

//Encryption 对应用的Handler进行数据的加解密
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

//GetAESKey 获取AES的秘钥，因为TOC服务所有秘钥生产规则一致
func GetAESKey(ctx context.Context) string {
	aesKey, ok := ctx.Value(aesContextKey).(string)
	if ok {
		return aesKey
	}
	return EmptyString
}

//SignToB 验证ToB服务的签名及生成追踪ID和设置请求头信息
func SignToB(whiteList ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Request.Header.Get(HeaderUserID)
		userName := c.Request.Header.Get(HeaderUserName)

		if !gutil.IsExactExist(whiteList, c.Request.URL.Path) && userID == EmptyString {
			c.Header(HeaderError, "userID is nil")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		trackID := c.GetHeader(HeaderXRequestID)
		if len(trackID) == 0 {
			trackID = gutil.UUID()
		}
		//设置请求头参数
		ctx := context.WithValue(c.Request.Context(), glog.TrackKey, trackID)
		ctx = context.WithValue(ctx, gutil.HeaderUserID, userID)
		ctx = context.WithValue(ctx, gutil.HeaderUserName, userName)
		c.Request = c.Request.WithContext(ctx)
	}
}

//GetUserID 获取用户信息
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value(gutil.HeaderUserID).(string)
	if ok {
		return userID
	}
	return EmptyString
}

//GetUserName 获取用户名称
func GetUserName(ctx context.Context) string {
	userName, ok := ctx.Value(gutil.HeaderUserName).(string)
	if ok {
		return userName
	}
	return EmptyString
}
