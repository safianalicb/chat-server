package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/sync/errgroup"
)

const serverAddress = "127.0.0.1:8080"
const delimiter byte = '\n'

func startConnection(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("could not start connection: %w", err)
	}

	return conn, nil

}

func getUserInput() (string, error) {
	userReader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter text: ")

	userMessage, err := userReader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("could not get user input: %w", err)
	}

	return userMessage, nil

}

func sendMessages(conn net.Conn) error {
	for {
		message, err := getUserInput()
		if err != nil {
			return fmt.Errorf("could not get user input: %w", err)
		}

		_, err = fmt.Fprint(conn, message)
		if err != nil {
			return fmt.Errorf("could not write to connection: %w", err)
		}
	}
}

func printAboveLine(s string) {
	fmt.Print("\0337")
	fmt.Print("\033[A")
	fmt.Print("\033[999D")
	fmt.Print("\033[S")
	fmt.Print("\033[L")
	fmt.Printf("Message received: %q\n", s)
	fmt.Print("\0338")
}

func receiveMessages(conn net.Conn) error {
	for {
		response, err := bufio.NewReader(conn).ReadString(delimiter)
		if err != nil {
			return fmt.Errorf("could not receive message from connection: %w", err)
		}

		printAboveLine(strings.TrimSpace(response))
	}
}

func terminateIfError(err error) {
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func main() {

	conn, err := startConnection(serverAddress)
	terminateIfError(err)
	defer conn.Close()

	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		return sendMessages(conn)
	})

	g.Go(func() error {
		return receiveMessages(conn)
	})

	err = g.Wait()
	terminateIfError(err)

}
