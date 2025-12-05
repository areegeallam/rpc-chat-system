package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Message struct {
	From string
	Body string
	Time time.Time
}

type ClientRPC struct {
	ID string
}

func (c *ClientRPC) Receive(msg Message, ack *struct{}) error {
	ts := msg.Time.Format("15:04:05")
	if msg.From == "system" {
		fmt.Printf("[%s] %s\n", ts, msg.Body)
	} else {
		fmt.Printf("[%s] %s: %s\n", ts, msg.From, msg.Body)
	}
	return nil
}

type JoinArgs struct {
	ID       string
	Callback string
}

type JoinReply struct {
	History []Message
}

type LeaveArgs struct {
	ID string
}

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: client <ID> <callbackPort>")
		return
	}

	id := os.Args[1]
	callback := os.Args[2] // example :9001

	// start local RPC server
	clientRPC := &ClientRPC{ID: id}
	rpc.Register(clientRPC)

	l, err := net.Listen("tcp", callback)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()

	// connect to server
	server, err := rpc.Dial("tcp", "127.0.0.1:12345")
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	// join
	var reply JoinReply
	err = server.Call("ChatServer.Join", JoinArgs{ID: id, Callback: callback}, &reply)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Chat History ===")
	for _, m := range reply.History {
		ts := m.Time.Format("15:04:05")
		if m.From == "system" {
			fmt.Printf("[%s] %s\n", ts, m.Body)
		} else {
			fmt.Printf("[%s] %s: %s\n", ts, m.From, m.Body)
		}
	}
	fmt.Println("====================")

	// handle exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		var ack struct{}
		server.Call("ChatServer.Leave", LeaveArgs{ID: id}, &ack)
		os.Exit(0)
	}()

	// send loop
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		text := strings.TrimSpace(sc.Text())

		if text == "/leave" {
			var ack struct{}
			server.Call("ChatServer.Leave", LeaveArgs{ID: id}, &ack)
			break
		}

		msg := Message{
			From: id,
			Body: text,
			Time: time.Now(),
		}
		var ack struct{}
		server.Call("ChatServer.Send", msg, &ack)
	}
}
