# RPC Chat System (Go)

## Overview

This project implements a real-time RPC-based chat system in Go. It supports multiple clients and provides the following features:

* Real-time message broadcasting using goroutines and channels
* No self-echo (messages are not sent back to the sender)
* User join and leave notifications
* Full chat history delivered to new clients
* Thread-safe client management using a Mutex

---

## Project Structure

```
rpc-chat-system/
├── server/
│   └── main.go
├── client/
│   └── main.go
├── go.mod
└── README.md
```

---

## How to Run

### 1. Run the Server

Open a terminal and execute:

```bash
cd server
go run main.go
```

The server will start listening on port `12345`.

### 2. Run Clients

Open a separate terminal for each client.

**Client 1:**

```bash
cd client
go run main.go user1 :9001
```

**Client 2:**

```bash
cd client
go run main.go user2 :9002
```

> Each client must use a unique callback port.

---

## Client Commands

* Type a message and press Enter → broadcast it to all other clients
* `/leave` → exit the chat gracefully
* `Ctrl + C` → exit the chat and notify others
