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

		fmt.Fprintf(w, "Response sent to client %d\n", id)
	}
}

func terminateIfError(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

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

			newClientCh <- newClient
			numClientsConnected.Add(1)

			fmt.Printf("Client %d added, %d clients currently connected.\n", numClientsConnected.Load(), connsSoFar)
		}

		time.Sleep(25 * time.Millisecond)
	}
}

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
