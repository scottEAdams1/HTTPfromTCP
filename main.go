package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	filePath := "messages.txt"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open %s\n", filePath)
	}

	ch := getLinesChannel(file)
	for v := range ch {
		fmt.Printf("read: %s\n", v)
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
