package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const SERVER_PORT = ":8080"
const DELIMITER = '\n'

func startListening(port string) (net.Listener, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, fmt.Errorf("could not listen on %q: %s", port, err)
	}

	return listener, nil

}

func getConnFromListener(listener net.Listener) (net.Conn, error) {
	conn, err := listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("could not create connection: %s", err)
	}

	return conn, nil

}

func terminateIfError(err error) {
	if err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func main() {
	listener, err := startListening(SERVER_PORT)
	terminateIfError(err)

	conn, err := getConnFromListener(listener)
	terminateIfError(err)

	for {
		response, err := bufio.NewReader(conn).ReadString(DELIMITER)
		terminateIfError(err)

		fmt.Println(response)

	}

}
