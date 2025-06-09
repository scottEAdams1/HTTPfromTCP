package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	port := ":42069"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Failed to open listener on port %s. Error: %s\n", port, err.Error())
	}
	defer listener.Close()
	fmt.Printf("Listening on port %s\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error: %s\n", err.Error())
		}
		fmt.Printf("Connection accepted from %s\n", conn.RemoteAddr())
		ch := getLinesChannel(conn)
		for v := range ch {
			fmt.Printf("%s\n", v)
		}
		fmt.Printf("Connection to %s closed\n", conn.RemoteAddr())
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	currentLine := ""

	go func() {
		defer f.Close()
		defer close(ch)
		for {
			b := make([]byte, 8)
			_, err := f.Read(b)
			if err != nil {
				if currentLine != "" {
					ch <- currentLine
				}
				if errors.Is(err, io.EOF) {
					break
				} else {
					fmt.Printf("Error: %s\n", err.Error())
					break
				}
			}
			s := string(b)
			parts := strings.Split(s, "\n")
			for i, v := range parts {
				if i < len(parts)-1 {
					ch <- currentLine + v
					currentLine = ""
				}
			}
			currentLine += parts[len(parts)-1]
		}
	}()

	return ch
}
