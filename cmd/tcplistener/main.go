package main

import (
	"fmt"
	"log"
	"net"

	r "github.com/finfreezer/httpfromtcp/internal/request"
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

}

func listenerLoop(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func(c net.Conn) {
			fmt.Println("Connection accepted.")
			linechan, err := r.RequestFromReader(c)
			if err != nil {
				log.Fatal(err)
			}
			helperPrintRequest(linechan)

			fmt.Println("Connection closed.")
		}(conn)
	}
}

func helperPrintRequest(r *r.Request) {
	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
	fmt.Println("Headers:")
	for key, value := range r.Headers {
		fmt.Printf("- %s: %s\n", key, value)
	}
	fmt.Println("Body:")
	fmt.Println(string(r.Body))
}
