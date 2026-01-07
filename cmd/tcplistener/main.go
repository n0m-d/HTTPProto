package main

import (
	"fmt"
	"log"
	"net"

	"example.com/HProtocol/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err) // Stop the program if the file can't be opened
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		r, err := request.RequestFromReader(conn)
		if err != nil {
			log.Printf("Error: %v", err)
		}

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for header, value := range r.Headers {
			fmt.Printf("- %s: %s\n", header, value)
		}

	}

}
