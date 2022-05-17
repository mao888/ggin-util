package gginutil

import (
	"context"
	"github.com/gin-gonic/gin"
	"gitlab.ftsview.com/fotoable-go/glog"
	"gitlab.ftsview.com/fotoable-go/gsecret"
	"gitlab.ftsview.com/fotoable-go/gutil"
	"net/http"
)

type AesKey string

const (
	aesContextKey AesKey = "aes_key"
)

func Sign2C(encryption bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := &Header2C{}
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
			glog.Debugf(c.Request.Context(), "aes_key: %s", aesKey)
			c.Request = c.Request.WithContext(context.WithValue(
				c.Request.Context(), aesContextKey, aesKey))
		}
	}
}

func GetAESKey(ctx context.Context) string {
	aesKey, ok := ctx.Value(aesContextKey).(string)
	if ok {
		return aesKey
	}
	return EmptyString
}
