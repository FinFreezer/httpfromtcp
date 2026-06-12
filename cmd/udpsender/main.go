package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	fmt.Println("Hello from UDP.")
	dst, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialUDP("udp", nil, dst)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	newReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">")
		line, err := newReader.ReadString(byte('\n'))
		lineBytes := []byte(line)
		if err != nil {
			log.Fatal(err)
		}
		_, err = conn.Write(lineBytes)
	}
}
