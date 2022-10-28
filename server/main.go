package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const serverPort = ":8080"
const delimiter byte = '\n'
const ackMessage string = "ACK" + string(delimiter)
const maxConnections = 5

var allClients []client

type client struct {
	id   int
	conn net.Conn
}

func getListener(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, fmt.Errorf("could not listen on port %q: %w", port, err)
	}

	return listener, nil

}

func sendMessageToAllClients(userId int, message string) {
	for _, client := range allClients {
		if client.id != userId {
			_, _ = fmt.Fprintf(client.conn, "Client %d: %v", userId, message)
		}
	}
}

// This function will listen for input to conn, and respond with an acknowldgement message.
// The function will log to the writer w as it receives messages and responds to them.
//
// The function will run until either a read or write from conn will return an error, so the
// function should only be run within a goroutine.
func listenForMessage(conn io.ReadWriter, id int, w io.Writer) error {
	for {
		clientMessage, err := bufio.NewReader(conn).ReadString(delimiter)
		if err != nil {
			return fmt.Errorf("could not read message: %v", err)
		}

		fmt.Fprintf(w, "Client %d sent: %q\n", id, strings.TrimSpace(clientMessage))

		_, err = fmt.Fprint(conn, ackMessage+string(delimiter))
		if err != nil {
			return fmt.Errorf("could not send message acknowledgement: %v", err)
		}

		sendMessageToAllClients(id, clientMessage)

		fmt.Fprintf(w, "Response sent to client %d\n", id)
	}
}

func terminateIfError(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

// This function will listen for connections on listener, and create clients and pass them into the newClientCh channel.
// The function will update numClientsConnected as new connections come in.
//
// The function will run indefinitely, so it should only be run within a goroutine.
func handleIncomingConns(listener net.Listener, newClientCh chan<- client, numClientsConnected *atomic.Int32) {
	connsSoFar := 0
	for {
		for numClientsConnected.Load() < maxConnections {
			conn, err := listener.Accept()
			if err != nil {
				fmt.Printf("could not create connection: %v\n", err)
				continue
			}
			connsSoFar++
			newClient := client{id: connsSoFar, conn: conn}

			allClients = append(allClients, newClient)

			newClientCh <- newClient
			numClientsConnected.Add(1)

			fmt.Printf("Client %d added, %d clients currently connected.\n", numClientsConnected.Load(), connsSoFar)
		}

		// This sleep is required, otherwise, when maxConnections has been reached, then the condition in the for
		// loop will constantly be checked, which leads to excessive CPU usage.
		time.Sleep(25 * time.Millisecond)
	}
}

// This function listens to the channel for new clients sent to the channel, and creates seperate goroutines to handle them
func createConnResponders(newClientCh <-chan client, numClientsConnected *atomic.Int32) {
	for newClient := range newClientCh {
		go func(c client) {
			defer c.conn.Close()
			err := listenForMessage(c.conn, c.id, os.Stdin)

			numClientsConnected.Add(-1)
			fmt.Printf("Client %d removed due to error %q\n%d clients still connected.\n", c.id, err, numClientsConnected.Load())
		}(newClient)
	}
}

func main() {
	listener, err := getListener(serverPort)
	terminateIfError(err)

	var numClientsConnected atomic.Int32
	newClientCh := make(chan client)

	go createConnResponders(newClientCh, &numClientsConnected)
	go handleIncomingConns(listener, newClientCh, &numClientsConnected)

	select {}

}
