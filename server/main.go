package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
)

const serverPort = ":8080"
const delimiter byte = '\n'
const ackMessage string = "ACK" + string(delimiter)
const maxConnections = 5

func getListener(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, fmt.Errorf("could not listen on port %q: %w", port, err)
	}

	return listener, nil

}

func listenForMessage(conn net.Conn) {
	for {
		clientMessage, err := bufio.NewReader(conn).ReadString(delimiter)
		if err != nil {
			fmt.Printf("could not read message: %v\n", err)
			return
		}

		fmt.Printf("Client sent: %q\n", strings.TrimSpace(clientMessage))

		_, err = fmt.Fprint(conn, ackMessage+string(delimiter))
		if err != nil {
			fmt.Printf("could not send message acknowledgement: %v\n", err)
			return
		}

		fmt.Println("Response sent")
	}
}

func terminateIfError(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func handleIncomingConns(listener net.Listener, connCh chan<- net.Conn, numConns *atomic.Int32) {
	for {
		for numConns.Load() < maxConnections {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("could not create connection: %v\n", err)
			}

			connCh <- conn
			numConns.Add(1)
			fmt.Println("ADDED", numConns)
		}
	}
}

func createConnResponders(connCh <-chan net.Conn, numConns *atomic.Int32) {
	for conn := range connCh {
		go func(conn net.Conn) {
			defer conn.Close()
			listenForMessage(conn)
			numConns.Add(-1)
			fmt.Println("REMOVED", numConns)
		}(conn)
	}
}

func main() {
	listener, err := getListener(serverPort)
	terminateIfError(err)

	var numConns atomic.Int32
	connCh := make(chan net.Conn)

	go createConnResponders(connCh, &numConns)
	go handleIncomingConns(listener, connCh, &numConns)

	select {}

}
