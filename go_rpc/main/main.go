package main

import (
	"encoding/json"
	"fmt"
	"go_rpc"
	"go_rpc/codec"
	"log"
	"net"
	"time"
)

func startServer(addr chan string) {
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	go_rpc.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	// in fact, following code is like a simple go rpc client
	conn, _ := net.Dial("tcp", <-addr)
	close(addr)
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)
	// send options
	_ = json.NewEncoder(conn).Encode(go_rpc.DefaultOption)
	cc := codec.NewJsonCodec(conn)
	// send request & receive response
	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("client go rpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("client get reply:", reply)
	}

}