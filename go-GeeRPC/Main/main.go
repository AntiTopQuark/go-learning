package main

//
//import (
//	"context"
//	"github.com/arl/statsviz"
//	"go-GeeRPC/client"
//	"go-GeeRPC/server"
//	"log"
//	"net"
//	"net/http"
//	"sync"
//	"time"
//)

type Foo int

type Args struct {
	Num1, Num2 int
}

func (f *Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

//func startServer(addr chan string) {
//	var foo Foo
//	if err := server.Register(&foo); err != nil {
//		log.Fatal("register error:", err)
//	}
//	// pick a free port
//	l, err := net.Listen("tcp", ":0")
//	if err != nil {
//		log.Fatal("network error:", err)
//	}
//	log.Println("start rpc server on", l.Addr())
//	addr <- l.Addr().String()
//	server.Accept(l)
//}
//func startMonitor(){
//	// Register statsviz handlers on the default serve mux.
//	log.Println("Monitor Page is : http://localhost:6060/debug/statsviz/")
//	statsviz.RegisterDefault()
//	http.ListenAndServe(":6060", nil)
//}
//func main() {
//	log.SetFlags(0)
//	go startMonitor()
//
//	addr := make(chan string)
//	go startServer(addr)
//	client, _ := client.Dial("tcp", <-addr)
//	defer func() { _ = client.Close() }()
//
//	time.Sleep(time.Second*5)
//	// send request & receive response
//	var wg sync.WaitGroup
//	for i := 0; i < 5; i++ {
//		wg.Add(1)
//		go func(i int) {
//			defer wg.Done()
//			args := &Args{Num1: i, Num2: i * i}
//			ctx, _ := context.WithTimeout(context.Background(), time.Second)
//
//			var reply int
//			if err := client.Call(ctx,"Foo.Sum", args, &reply); err != nil {
//				log.Fatal("call Foo.Sum error:", err)
//			}
//			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
//		}(i)
//	}
//	time.Sleep(time.Second*5)
//	wg.Wait()
//}
