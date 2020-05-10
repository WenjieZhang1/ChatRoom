package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"runtime"
)

type Client struct {
	Name        string
	Addr        string
	MessageChan chan interface{}
}

var userList = make(map[string]Client)

var message = make(chan interface{})

func getBytes(msg interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(msg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func handleConnect(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()

	client := Client{
		Name:        clientAddr,
		Addr:        clientAddr,
		MessageChan: make(chan interface{}),
	}

	userList[clientAddr] = client

	go func() {
		for {
			msg := <-client.MessageChan

			content, err := getBytes(msg)
			if err != nil {
				fmt.Printf("Error when converting interface to bytes array: %v", err)
				continue
			}
			_, err = conn.Write(content)
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
		for _, user := range userList {
			user.MessageChan <- msg
		}
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
