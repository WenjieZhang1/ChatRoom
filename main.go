package main

import (
	"fmt"
	"log"
	"net"
	"runtime"
	"sync"
)

type Client struct {
	Name        string
	Addr        string
	MessageChan chan string
}

type userList struct {
	users map[string]Client
	mux   sync.Mutex
}

func addUser(users *userList, client Client) {
	users.mux.Lock()
	users.users[client.Addr] = client
	users.mux.Unlock()
}

func broadCastUsers(users *userList, msg string) {
	users.mux.Lock()
	defer users.mux.Unlock()
	for _, user := range users.users {
		user.MessageChan <- msg
	}
}

var message = make(chan string)

var onlineUsers = userList{
	users: make(map[string]Client),
	mux:   sync.Mutex{},
}

func handleConnect(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()

	client := Client{
		Name:        clientAddr,
		Addr:        clientAddr,
		MessageChan: make(chan string),
	}

	addUser(&onlineUsers, client)

	go func() {
		for {
			msg := <-client.MessageChan
			_, err := conn.Write([]byte(msg + "\n"))
			if err != nil {
				fmt.Printf("Error when writing message to client: %v", err)
			}
		}
	}()

	msg := "[" + clientAddr + "]" + client.Name + ": get online!"
	message <- msg

	for {
		runtime.GC()
	}
}

func Manager() {
	for {
		msg := <-message
		broadCastUsers(&onlineUsers, msg)
	}
}

func main() {
	// Start the server
	listener, err := net.Listen("tcp", "127.0.0.1:8800")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	// Create message manager
	go Manager()
	// Listen client side queries
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept error: ", err)
			continue
		}
		go handleConnect(conn)
	}
}
