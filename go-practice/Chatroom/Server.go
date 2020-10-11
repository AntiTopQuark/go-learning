package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:18080")
	if err != nil {
		log.Fatal(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	out := make(chan string, 10)
	go clientWriter(conn, out)
	in := make(chan string)
	go clientReader(conn, in)

	var who string
	var nameTimer = time.NewTimer(timeout)

	out <- "Entry your name:"
	select {
	case name := <-in:
		who = name
	case <-nameTimer.C:
		conn.Close()
		return
	}
	cli := client{out, who}

	out <- "You are" + who
	messages <- who + "has arrived"
	entering <- cli

	idle := time.NewTimer(timeout)

Loop:
	for {
		select {
		case msg := <-in:
			messages <- who + ":" + msg
			idle.Reset(timeout)
		case <-idle.C:
			conn.Close()
			break Loop
		}
	}

}

func clientWriter(conn net.Conn, out chan string) {
	for msg := range out {
		fmt.Fprintln(conn, msg)
	}
}

func clientReader(conn net.Conn, ch chan<- string) {
	input := bufio.NewScanner(conn)
	for input.Scan() {
		ch <- input.Text()
	}
}

type client struct {
	Out  chan<- string
	Name string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
	timeout  = 10 * time.Second
)

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli.Out <- msg
			}
		case cli := <-entering:
			clients[cli] = true
			cli.Out <- "在线: "
			for c := range clients {
				cli.Out <- c.Name
			}
		case cli := <-leaving:
			delete(clients, cli)
			close(cli.Out)
		}
	}
}
