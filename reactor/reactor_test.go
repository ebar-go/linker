package reactor

import (
	"log"
	"testing"
)

func TestReactor(t *testing.T) {
	reactor := NewReactor()
	if err := reactor.Listen("tcp", "0.0.0.0:8086"); err != nil {
		t.Fatal(err)
	}

	reactor.ev.OnConnect(func(conn Conn) {
		log.Println("connected")
	})

	reactor.ev.OnDisconnect(func(conn Conn) {
		log.Println("disconnected")
	})
	reactor.ev.OnRequest(func(ctx Context) {
		log.Println("receive:", string(ctx.Body()))
	})

	reactor.Start()
}
