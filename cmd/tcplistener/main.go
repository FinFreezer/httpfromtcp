package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	fmt.Println("I hope I get the job!")
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Printf("Error listening for traffic: %s.", err)
		return
	}
	defer listener.Close()
	listenerLoop(listener)
	/*lineChan := getLinesChannel(file)

	for line := range lineChan {
		fmt.Println(line)
	}*/

}

func listenerLoop(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func(c net.Conn) {
			fmt.Println("Connection accepted.")
			linechan := getLinesChannel(c)
			for line := range linechan {
				fmt.Println(line)
			}
			fmt.Println("Connection closed.")
		}(conn)
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	currentLine := ""
	lineChan := make(chan string)

	buf := make([]byte, 8)
	go func() {
		defer f.Close()
		for {
			n, err := f.Read(buf)

			parts := strings.Split(string(buf[:n]), "\n")
			if len(parts) == 1 {
				currentLine += parts[0]

			} else {
				for i := range len(parts) - 1 {
					currentLine += parts[i]
					lineChan <- currentLine
					currentLine = ""
					currentLine += parts[i+1]
				}
			}
			if err != nil {
				lineChan <- currentLine
				close(lineChan)
				return
			}
		}
	}()
	return lineChan
}
