package linker

import (
	"fmt"
	"github.com/rcrowley/go-metrics"
	"linker/pkg/system"
	"log"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"
)

var (
	addr = "0.0.0.0:8086"
)

func TestReactor(t *testing.T) {
	var connected int
	go func() {
		for {
			time.Sleep(time.Second * 5)

			fmt.Printf("current connection:%d, memory usage: %.3f MB\n", connected, float64(system.GetMem())/1024/1024)
		}

	}()
	system.SetLimit()
	reactor := NewReactor()

	reactor.OnConnect(func(conn Conn) {
		connected++
		log.Println("connected")
	})

	reactor.OnDisconnect(func(conn Conn) {
		connected--
		log.Println("disconnected", conn.FD())
	})
	reactor.OnRequest(func(ctx Context) {
		//log.Println("receive:", string(ctx.Body()))
		ctx.Conn().Push([]byte(time.Now().String()))
	})

	if err := reactor.Run("tcp", addr); err != nil {
		t.Fatal(err)
	}
}

func TestTCPServer(t *testing.T) {
	var connected int
	go func() {
		for {
			time.Sleep(time.Second * 5)

			fmt.Printf("current connection:%d, memory usage: %.3f MB\n", connected, float64(system.GetMem())/1024/1024)
		}

	}()
	system.SetLimit()
	reactor := NewTCPServer()

	reactor.OnConnect(func(conn Conn) {
		connected++
		log.Println("connected")
	})

	reactor.OnDisconnect(func(conn Conn) {
		connected--
		log.Println("disconnected", conn.FD())
	})
	reactor.OnRequest(func(ctx Context) {
		//log.Println("receive:", string(ctx.Body()))
		ctx.Conn().Push([]byte(time.Now().String()))
	})

	if err := reactor.Run("tcp", addr); err != nil {
		t.Fatal(err)
	}
	select {}
}

func TestClient(t *testing.T) {
	c, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			bytes := make([]byte, 100)
			if _, err := c.Read(bytes); err != nil {
				panic(err)
			}

			log.Println("receive:", string(bytes))
		}

	}()

	for {
		time.Sleep(time.Second * 1)
		if _, err := c.Write([]byte("hello,world\n")); err != nil {
			panic(err)
		}
	}
}

func BenchmarkClient(b *testing.B) {
	system.SetLimit()
	opsRate := metrics.NewRegisteredTimer("ops", nil)

	n := 20000
	connections := make([]net.Conn, 0, 1024)
	for i := 0; i < n; i++ {
		c, err := net.DialTimeout("tcp", addr, 10*time.Second)
		if err != nil {
			i--
		}
		connections = append(connections, c)
	}

	go func() {
		metrics.Log(metrics.DefaultRegistry, 5*time.Second, log.New(os.Stderr, "metrics: ", log.Lmicroseconds))
	}()

	b.ResetTimer()
	for i := 0; i < 100; i++ {
		go func() {
			for {
				n := rand.Intn(len(connections) - 1)
				c := connections[n]
				before := time.Now()
				if _, err := c.Write([]byte("hello\n")); err != nil {
					_ = c.Close()
					//log.Println(err)
				} else {
					opsRate.Update(time.Now().Sub(before))
				}
			}

		}()
	}
	select {}
}
