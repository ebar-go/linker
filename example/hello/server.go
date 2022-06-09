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
