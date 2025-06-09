package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	address := "localhost:42069"
	udp, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("Failed to resolve udp at address %s. Error: %s\n", address, err.Error())
	}
	conn, err := net.DialUDP("udp", nil, udp)
	if err != nil {
		log.Fatalf("Failed to connect udp at address %s. Error: %s\n", address, err.Error())
	}
	defer conn.Close()
	fmt.Printf("Connected at address %s. Type your message.\n", address)
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		s, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}
		_, err = conn.Write([]byte(s))
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Message sent: %s\n", s)
	}
}
