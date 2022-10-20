# ggin-util

#### 基于gin包装的工具包
> 注意：改包需要依赖gsecret
* 使用
```go
    router := gin.New()
    router.Use(
        gginutil.SignToC(true),
    )
```
> 参数true标识使用加密

* 使用加密时获取AES秘钥
```go
    gginutil.GetAESKey(ctx)
```
> ctx context.Context则是从gin.Context中通过Request.Context()获取的

* 增加多参数绑定
```go
type InfoBean struct {
	XUser string     `header:"x-user"` // 绑定请求头的信息 注意：需要传入bindHeader为true
	Query int32      `form:"query"`    // 绑定Query中的参数url，?后之后的参数，示例：localhost:8888/info/100?query=666，
	Time  int32      `uri:"time"`      // 绑定URL上的路径参数，定义路由为：/info/:time，localhost:8888/info/100?query=666
	Name  string     `json:"name"`
	Age   int32      `json:"age"`
	Class []*string  `json:"class"`
	Sub   []*SubBean `json:"sub"`
}
```

* 增加参数校验方法
```go
type InfoBean struct {
	XUser string     `header:"x-user"`
	Query int32      `form:"query" bt-validator:"required"`
	Time  int32      `uri:"time" bt-validator:"required"`
	Name  string     `json:"name"`
	Age   int32      `json:"age"`
	Class []*string  `json:"class" bt-validator:"required"`
	Sub   []*SubBean `json:"sub"`
}
```
> 添加bt-validator标签，启用验证，目前只支持必须输入验证