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
	defer f.Close() // Make sure the file gets closed when we're done

	str := "" // This will store characters until we reach a full line

	for {
		// Create a byte slice to read up to 8 bytes at a time
		data := make([]byte, 8)

		// Read from the file into the data slice
		n, err := f.Read(data)

		if n > 0 {
			// Trim the data slice to only the bytes that were read
			data = data[:n]

			// Loop in case there are multiple newlines in one chunk
			for {
				// Look for the position of a newline character '\n'
				i := bytes.IndexByte(data, '\n')

				if i == -1 {
					// No newline found — just add everything to str
					str += string(data)
					break // exit inner loop, get next chunk
				}

				// Newline found — add up to it to str
				str += string(data[:i])

				// Print the full line we just built
				fmt.Printf("read: %s\n", str)

				// Reset str for the next line
				str = ""

				// Move data pointer to after the newline
				data = data[i+1:]

				// If no more data left after the newline, break the inner loop
				if len(data) == 0 {
					break
				}
			}
		}

		// If we've reached the end of the file, stop the loop
		if err == io.EOF {
			break
		}

		// If another error happened, stop the program
		if err != nil {
			log.Fatal(err)
		}
	}

	// If there’s leftover text (i.e. the last line didn’t end with a newline), print it
	if len(str) != 0 {
		fmt.Printf("read: %s\n", str)
	}
}
