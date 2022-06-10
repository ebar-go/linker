package main

import (
	"linker"
	"linker/utils/system"
	"log"
)

func main() {
	system.SetLimit()

	server := linker.NewTCPServer([]string{"127.0.0.1:7086"})

	// 主逻辑
	server.SetOnReceive(func(ctx linker.IContext) {
		//log.Println("receive:", string(ctx.Request().Body()))

		// 将内容原封不动返回
		ctx.Output(ctx.Request().Body())
	})

	if err := server.Start(); err != nil {
		panic(err)
	}

	system.Shutdown(func() {
		log.Println("server shutdown")
	})
}
