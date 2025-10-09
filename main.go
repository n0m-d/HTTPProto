package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	// Open the file for reading
	f, err := os.Open("messages.txt")
	if err != nil {
		log.Fatal(err) // Stop the program if the file can't be opened
	}

	lines := getLinesChannel(f)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	strChan := make(chan string)

	go func() {
		defer f.Close()
		defer close(strChan)

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
