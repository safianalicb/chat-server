package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const SERVER_ADDRESS = "127.0.0.1:8080"

func startConnection(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not start connection: %s", err)
	}

	return conn, nil

}

func sendMessage(conn net.Conn, message string) (string, error) {
	_, err := fmt.Fprint(conn, message)
	if err != nil {
		return "", fmt.Errorf("could not send message: %s", err)
	}

	response, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("could not send message: %s", err)
	}

	return response, nil

}

func main() {

	conn, err := startConnection(SERVER_ADDRESS)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	response, err := sendMessage(conn, "HELLO\n")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(response)

}
