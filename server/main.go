package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Message struct {
	From string
	Body string
	Time time.Time
}

type Client struct {
	ID       string
	Callback string
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

type ChatServer struct {
	mu       sync.Mutex
	clients  map[string]*Client
	history  []Message
	incoming chan Message
}

func NewChatServer() *ChatServer {
	s := &ChatServer{
		clients:  make(map[string]*Client),
		history:  []Message{},
		incoming: make(chan Message, 100),
	}
	go s.broadcaster()
	return s
}

func (s *ChatServer) Join(args JoinArgs, reply *JoinReply) error {
	if args.ID == "" || args.Callback == "" {
		return errors.New("ID and Callback required")
	}

	// register client
	s.mu.Lock()
	s.clients[args.ID] = &Client{ID: args.ID, Callback: args.Callback}

	// return history copy
	h := make([]Message, len(s.history))
	copy(h, s.history)
	s.mu.Unlock()

	// broadcast join message
	notice := Message{
		From: "system",
		Body: fmt.Sprintf("User %s joined", args.ID),
		Time: time.Now(),
	}

	s.mu.Lock()
	s.history = append(s.history, notice)
	s.mu.Unlock()

	s.incoming <- Message{From: args.ID, Body: notice.Body, Time: notice.Time}

	reply.History = h
	fmt.Println("User", args.ID, "joined. Total:", len(s.clients))
	return nil
}

func (s *ChatServer) Send(msg Message, ack *struct{}) error {
	if msg.From == "" || msg.Body == "" {
		return errors.New("invalid message")
	}
	msg.Time = time.Now()

	s.mu.Lock()
	s.history = append(s.history, msg)
	s.mu.Unlock()

	s.incoming <- msg
	return nil
}

func (s *ChatServer) Leave(args LeaveArgs, ack *struct{}) error {
	s.mu.Lock()
	delete(s.clients, args.ID)
	s.mu.Unlock()

	notice := Message{
		From: "system",
		Body: fmt.Sprintf("User %s left", args.ID),
		Time: time.Now(),
	}

	s.mu.Lock()
	s.history = append(s.history, notice)
	s.mu.Unlock()

	s.incoming <- Message{From: args.ID, Body: notice.Body, Time: notice.Time}
	fmt.Println("User", args.ID, "left")
	return nil
}

func (s *ChatServer) broadcaster() {
	for msg := range s.incoming {

		s.mu.Lock()
		clients := make([]*Client, 0, len(s.clients))
		for _, c := range s.clients {
			clients = append(clients, c)
		}
		s.mu.Unlock()

		var wg sync.WaitGroup
		for _, c := range clients {
			if c.ID == msg.From {
				continue
			}
			wg.Add(1)
			go func(cl *Client) {
				defer wg.Done()

				client, err := rpc.Dial("tcp", cl.Callback)
				if err != nil {
					s.mu.Lock()
					delete(s.clients, cl.ID)
					s.mu.Unlock()
					return
				}
				defer client.Close()

				var resp struct{}
				client.Call("Client.Receive", msg, &resp)
			}(c)
		}
		wg.Wait()
	}
}

func main() {
	chat := NewChatServer()
	rpc.Register(chat)

	l, err := net.Listen("tcp", ":12345")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Chat server running on :12345")

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeConn(conn)
	}
}
