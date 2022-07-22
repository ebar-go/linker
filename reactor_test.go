package linker

import (
	"log"
	"net"
	"testing"
	"time"
)

var (
	addr = "0.0.0.0:8086"
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

	if err := reactor.Listen("tcp", addr); err != nil {
		t.Fatal(err)
	}
}

func TestClient(t *testing.T) {
	c, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			bytes := make([]byte, 4096)
			if _, err := c.Read(bytes); err != nil {
				panic(err)
			}

			log.Println("receive:", string(bytes))
		}

	}()

	for {
		time.Sleep(time.Second * 3)
		if _, err := c.Write([]byte("hello,world\n")); err != nil {
			panic(err)
		}
	}
}
