package linker

import (
	"log"
	"testing"
	"time"
)

func TestReactor(t *testing.T) {
	reactor := NewReactor()

	reactor.OnConnect(func(conn Conn) {
		log.Println("connected")
	})

	reactor.OnDisconnect(func(conn Conn) {
		log.Println("disconnected")
	})
	reactor.OnRequest(func(ctx Context) {
		log.Println("receive:", string(ctx.Body()))
		ctx.Conn().Push([]byte(time.Now().String()))
	})

	if err := reactor.Listen("tcp", "0.0.0.0:8086"); err != nil {
		t.Fatal(err)
	}
}
