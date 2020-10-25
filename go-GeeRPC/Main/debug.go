package main

import (
	"context"
	"go-GeeRPC/client"
	"go-GeeRPC/server"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

func startServer2(addrCh chan string) {
	var foo Foo
	l, _ := net.Listen("tcp", ":9999")
	_ = server.Register(&foo)
	server.HandleHTTP()
	print(l.Addr().String())
	addrCh <- l.Addr().String()
	_ = http.Serve(l, nil)
}
func call(addrCh chan string) {
	client, _ := client.DialHTTP("tcp", <-addrCh)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	ch := make(chan string)
	go call(ch)
	startServer2(ch)
}
