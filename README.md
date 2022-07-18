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