package main

import (
	"flag"
	"log"
	"net"
	"time"
)

var (
	host = flag.String("ip", "127.0.0.1:8086", "server IP")
)

func main() {
	flag.Parse()
	addr := *host

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
		if _, err := c.Write([]byte("hello,world")); err != nil {
			panic(err)
		}
		//time.Sleep(time.Hour)
	}
}
