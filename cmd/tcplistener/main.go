package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
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

		for line := range getLinesChannel(conn) {
			fmt.Printf("%s\n", line)
		}
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	strChan := make(chan string)

	go func() {

		defer func() {
			f.Close()
			close(strChan)
		}()

		str := ""
		for {
			data := make([]byte, 8)
			n, err := f.Read(data)

			if err == io.EOF || err != nil {
				break
			}

			data = data[:n]
			if i := bytes.IndexByte(data, '\n'); i != -1 {
				str += string(data[:i])
				data = data[i+1:]
				strChan <- str
				str = ""
			}

			str += string(data)

		}

		if len(str) != 0 {
			strChan <- str
		}

	}()
	return strChan
}
