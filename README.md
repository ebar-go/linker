# linker
长链接的网关层,支持tcp,udp,websocket等协议

## 安装
```
go get github.com/ebar-go/linker
```

## 示例
```go
package main

import (
	"linker"
	"linker/utils/system"
	"log"
)

func main() {
	server := linker.NewGroupServer()
	server.Register(linker.NewTCPServer([]string{"127.0.0.1:7086"}, linker.WithDebug()))

	// 主逻辑
	server.SetOnReceive(func(ctx linker.IContext) {
		log.Println("receive:", string(ctx.Request().Body()))
		ctx.Output(ctx.Request().Body())
	})

	if err := server.Start(); err != nil {
		panic(err)
	}

	system.Shutdown(func() {
		log.Println("server shutdown")
	})
}

```
## 特性
- 支持创建tcp,websocket服务以及服务组
- 支持Connect,Disconnect,Receive事件回调
- 支持上下文自定义参数
- 支持中间件

## 协议报文
```
--------------
| len | data |
--------------
```
其中：

- len: 协议包的总长度（字节数）。(设置len长度主要是为了解决粘包问题,默认为0)
- data: 数据体，使用 JSON 或者 protobuf 编码（也可自定义其他编码方式）